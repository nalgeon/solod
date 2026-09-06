package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// mapTypes returns the map type of x and the C names of its key and value types.
func (g *Generator) mapTypes(node ast.Node, x ast.Expr) (*types.Map, string, string) {
	mapType := g.types.TypeOf(x).Underlying().(*types.Map)
	return mapType, g.mapTypeName(node, mapType.Key()), g.mapTypeName(node, mapType.Elem())
}

// emitMapLit emits a map literal as a so_map_lit call.
// Example: map[string]int{"a": 11, "b": 22} ->
//
//	so_map_lit(so_String, so_int, 2,
//		((so_String[]){so_str("a"), so_str("b")}),
//		((so_int[]){11, 22})))
func (g *Generator) emitMapLit(w io.Writer, n *ast.CompositeLit) {
	mapType, keyType, valType := g.mapTypes(n, n)
	g.checkMap(n, mapType.Key(), mapType.Elem())
	size := len(n.Elts)

	g.checkLocalScope(n)
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
		g.checkMacroArg(val)
		g.emitExprAsType(w, n, val, mapType.Elem())
	}
	fmt.Fprint(w, "}))")
}

// emitMapIndexExpr emits a map index read as so_map_get(K, V, m, key).
func (g *Generator) emitMapIndexExpr(w io.Writer, n *ast.IndexExpr) {
	_, keyType, valType := g.mapTypes(n, n.X)

	g.checkLocalScope(n)
	fmt.Fprintf(w, "so_map_get(%s, %s, ", keyType, valType)
	g.emitExpr(w, n.X)
	fmt.Fprint(w, ", ")
	g.emitMacroArg(w, n.Index)
	fmt.Fprint(w, ")")
}

// emitMapIndexAssign emits a map index write as so_map_set(K, V, &m, key, val).
func (g *Generator) emitMapIndexAssign(w io.Writer, node ast.Node, idx *ast.IndexExpr, rhs ast.Expr) {
	mapType, keyType, valType := g.mapTypes(node, idx.X)

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
	_, keyType, valType := g.mapTypes(stmt, idx.X)

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
	_, keyType, valType := g.mapTypes(stmt, stmt.X)

	// An enclosing block scopes _m, so two range loops in one block do not conflict.
	fmt.Fprintf(w, "%s{\n", g.indent())
	g.state.depth++

	// A hidden _m variable holds the map, so the map expression is evaluated once.
	fmt.Fprintf(w, "%sso_Map* _m = ", g.indent())
	g.emitExpr(w, stmt.X)
	fmt.Fprint(w, ";\n")

	// A hidden _i variable iterates the internal arrays.
	// The guard on NULL _m keeps the loop body unused for a nil map.
	fmt.Fprintf(w, "%sfor (so_int _i = 0; _m != NULL && _i < _m->cap; _i++) {\n", g.indent())

	// The map key is not a loop counter, so the loop does not need rangeKey.
	g.emitRangeBody(w, stmt, rangeKey{}, func() {
		// Skip empty slots in hash table.
		fmt.Fprintf(w, "%sif (!_m->used[_i]) continue;\n", g.indent())
		if name, declare := rangeVar(stmt, stmt.Key); name != "" {
			fmt.Fprintf(w, "%s%s%s = ((%s*)_m->keys)[_i];\n",
				g.indent(), rangeDecl(keyType, declare), name, keyType)
		}
		if name, declare := rangeVar(stmt, stmt.Value); name != "" {
			fmt.Fprintf(w, "%s%s%s = ((%s*)_m->vals)[_i];\n",
				g.indent(), rangeDecl(valType, declare), name, valType)
		}
	})

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

// checkMapIndex fails if target updates a map index in place.
func (g *Generator) checkMapIndex(target ast.Expr) {
	idx, ok := ast.Unparen(target).(*ast.IndexExpr)
	if !ok || !isMapType(g.types.TypeOf(idx.X)) {
		return
	}
	g.fail(target, "operation on map item is not supported")
}
