package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
)

// emitMapLit emits a map literal as a so_map_lit call.
// Example: map[string]int{"a": 11, "b": 22} ->
//
//	so_map_lit(so_String, so_int, 2,
//		((so_String[]){so_str("a"), so_str("b")}),
//		((so_int[]){11, 22})))
func (g *Generator) emitMapLit(w io.Writer, n *ast.CompositeLit) {
	mapType := g.types.TypeOf(n).Underlying().(*types.Map)
	g.checkMap(n, mapType.Key(), mapType.Elem())
	keyType := g.mapTypeName(n, mapType.Key())
	valType := g.mapTypeName(n, mapType.Elem())
	size := len(n.Elts)

	if size == 0 {
		fmt.Fprint(w, "&(so_Map){}")
		return
	}

	fmt.Fprintf(w, "so_map_lit(%s, %s, %d, ((%s[]){", keyType, valType, size, keyType)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExpr(w, elt.(*ast.KeyValueExpr).Key)
	}
	fmt.Fprintf(w, "}), ((%s[]){", valType)
	isAny := isEmptyInterface(mapType.Elem())
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		val := elt.(*ast.KeyValueExpr).Value
		if isAny {
			g.emitAnyMacroArg(w, n, val)
			continue
		}
		g.emitExprAsType(w, n, val, mapType.Elem())
	}
	fmt.Fprint(w, "}))")
}

// emitMapIndexExpr emits a map index read as so_map_get(K, V, m, key).
func (g *Generator) emitMapIndexExpr(w io.Writer, n *ast.IndexExpr) {
	mapType := g.types.TypeOf(n.X).Underlying().(*types.Map)
	keyType := g.mapTypeName(n, mapType.Key())
	valType := g.mapTypeName(n, mapType.Elem())

	fmt.Fprintf(w, "so_map_get(%s, %s, ", keyType, valType)
	g.emitExpr(w, n.X)
	fmt.Fprint(w, ", ")
	g.emitMacroArg(w, n.Index)
	fmt.Fprint(w, ")")
}

// emitMapIndexAssign emits a map index write as so_map_set(K, V, &m, key, val).
func (g *Generator) emitMapIndexAssign(w io.Writer, node ast.Node, idx *ast.IndexExpr, rhs ast.Expr) {
	mapType := g.types.TypeOf(idx.X).Underlying().(*types.Map)
	keyType := g.mapTypeName(node, mapType.Key())
	valType := g.mapTypeName(node, mapType.Elem())

	fmt.Fprintf(w, "%sso_map_set(%s, %s, ", g.indent(), keyType, valType)
	g.emitExpr(w, idx.X)
	fmt.Fprint(w, ", ")
	g.emitMacroArg(w, idx.Index)
	fmt.Fprint(w, ", ")
	g.emitMacroArgAsType(w, node, rhs, mapType.Elem())
	fmt.Fprint(w, ");\n")
}

// emitMapCommaOk emits a comma-ok map access: v, ok := m[key] or v, ok = m[key].
// Emits two statements: a so_map_get for the value and a so_map_has for the bool.
func (g *Generator) emitMapCommaOk(w io.Writer, stmt *ast.AssignStmt, idx *ast.IndexExpr, isDefine bool) {
	mapType := g.types.TypeOf(idx.X).Underlying().(*types.Map)
	keyType := g.mapTypeName(stmt, mapType.Key())
	valType := g.mapTypeName(stmt, mapType.Elem())

	vIdent := stmt.Lhs[0].(*ast.Ident)
	okIdent := stmt.Lhs[1].(*ast.Ident)

	// Emit: [type] v = so_map_get(K, V, &m, key);
	if vIdent.Name != "_" {
		vDecl := ""
		if isDefine && g.types.Defs[vIdent] != nil {
			vDecl = valType + " "
		}
		fmt.Fprintf(w, "%s%s%s = so_map_get(%s, %s, ", g.indent(), vDecl, vIdent.Name, keyType, valType)
		g.emitExpr(w, idx.X)
		fmt.Fprint(w, ", ")
		g.emitMacroArg(w, idx.Index)
		fmt.Fprint(w, ");\n")
	}

	// Emit: [bool] ok = so_map_has(K, &m, key);
	if okIdent.Name != "_" {
		okDecl := ""
		if isDefine && g.types.Defs[okIdent] != nil {
			okDecl = "bool "
		}
		fmt.Fprintf(w, "%s%s%s = so_map_has(%s, ", g.indent(), okDecl, okIdent.Name, keyType)
		g.emitExpr(w, idx.X)
		fmt.Fprint(w, ", ")
		g.emitMacroArg(w, idx.Index)
		fmt.Fprint(w, ");\n")
	}
}

// emitMapRange emits a for-range loop over a map.
func (g *Generator) emitMapRange(w io.Writer, stmt *ast.RangeStmt) {
	mapType := g.types.TypeOf(stmt.X).Underlying().(*types.Map)
	keyType := g.mapTypeName(stmt, mapType.Key())
	valType := g.mapTypeName(stmt, mapType.Elem())

	// An enclosing block scopes _m, so two range loops in one block do not collide.
	fmt.Fprintf(w, "%s{\n", g.indent())
	g.state.depth++

	// A hidden _m variable holds the map, so the map expression is evaluated once.
	fmt.Fprintf(w, "%sso_Map* _m = ", g.indent())
	g.emitExpr(w, stmt.X)
	fmt.Fprint(w, ";\n")

	// A hidden _i variable iterates the internal arrays.
	// The guard on NULL _m keeps the loop body unused for a nil map.
	fmt.Fprintf(w, "%sfor (so_int _i = 0; _m != NULL && _i < _m->cap; _i++) {\n", g.indent())

	g.state.depth++

	// Skip empty slots in hash table.
	fmt.Fprintf(w, "%sif (!_m->used[_i]) continue;\n", g.indent())

	// Emit key variable.
	if stmt.Key != nil {
		if keyIdent, ok := stmt.Key.(*ast.Ident); ok && keyIdent.Name != "_" {
			keyDecl := keyType + " "
			if stmt.Tok == token.ASSIGN {
				keyDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = ((%s*)_m->keys)[_i];\n", g.indent(), keyDecl, keyIdent.Name, keyType)
		}
	}

	// Emit value variable.
	if stmt.Value != nil {
		if valIdent, ok := stmt.Value.(*ast.Ident); ok && valIdent.Name != "_" {
			valDecl := valType + " "
			if stmt.Tok == token.ASSIGN {
				valDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = ((%s*)_m->vals)[_i];\n", g.indent(), valDecl, valIdent.Name, valType)
		}
	}

	g.state.depth--

	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())

	g.state.depth--
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// checkMap fails if the key or the value type of a map is not supported.
func (g *Generator) checkMap(node ast.Node, keyType, valType types.Type) {
	if isEmptyInterface(keyType) {
		g.fail(node, "any as map key type is not supported")
	}
	if _, ok := valType.Underlying().(*types.Array); ok {
		g.fail(node, "array as map value type is not supported")
	}
}
