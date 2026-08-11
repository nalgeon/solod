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
	g.checkMapValueType(n, mapType.Elem())
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
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExpr(w, elt.(*ast.KeyValueExpr).Value)
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
	g.emitMacroArg(w, rhs)
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
// Uses a hidden _i variable to iterate the internal arrays.
func (g *Generator) emitMapRange(w io.Writer, stmt *ast.RangeStmt) {
	mapType := g.types.TypeOf(stmt.X).Underlying().(*types.Map)
	keyType := g.mapTypeName(stmt, mapType.Key())
	valType := g.mapTypeName(stmt, mapType.Elem())

	fmt.Fprintf(w, "%sfor (so_int _i = 0; _i < ", g.indent())
	g.emitExpr(w, stmt.X)
	fmt.Fprint(w, "->cap; _i++) {\n")

	g.state.depth++

	// Skip empty slots in hash table.
	fmt.Fprintf(w, "%sif (!", g.indent())
	g.emitExpr(w, stmt.X)
	fmt.Fprint(w, "->used[_i]) continue;\n")

	// Emit key variable.
	if stmt.Key != nil {
		if keyIdent, ok := stmt.Key.(*ast.Ident); ok && keyIdent.Name != "_" {
			keyDecl := keyType + " "
			if stmt.Tok == token.ASSIGN {
				keyDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = ((%s*)", g.indent(), keyDecl, keyIdent.Name, keyType)
			g.emitExpr(w, stmt.X)
			fmt.Fprint(w, "->keys)[_i];\n")
		}
	}

	// Emit value variable.
	if stmt.Value != nil {
		if valIdent, ok := stmt.Value.(*ast.Ident); ok && valIdent.Name != "_" {
			valDecl := valType + " "
			if stmt.Tok == token.ASSIGN {
				valDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = ((%s*)", g.indent(), valDecl, valIdent.Name, valType)
			g.emitExpr(w, stmt.X)
			fmt.Fprint(w, "->vals)[_i];\n")
		}
	}

	g.state.depth--

	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// checkMapValueType fails if the map value type is not supported in C.
// mapTypeName rejects this too, but only when it sees the whole map type.
// Map literals and make() pass the value type alone, so check it here.
func (g *Generator) checkMapValueType(node ast.Node, valType types.Type) {
	if _, ok := valType.Underlying().(*types.Array); ok {
		g.fail(node, "array as map value type is not supported")
	}
}
