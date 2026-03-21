package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// emitMapLit emits a map literal as a so_Map compound literal.
// Example: map[string]int{"a": 11, "b": 22} ->
//
//	(so_Map){(so_String[2]){so_str("a"), so_str("b")}, (so_int[2]){11, 22}, 2, 2}
func (g *Generator) emitMapLit(n *ast.CompositeLit) {
	w := g.state.writer
	mapType := g.types.TypeOf(n).Underlying().(*types.Map)
	g.validateMapValueType(n, mapType.Elem())
	keyType := g.mapType(n, mapType.Key())
	valType := g.mapType(n, mapType.Elem())
	size := len(n.Elts)

	if size == 0 {
		fmt.Fprintf(w, "&(so_Map){0}")
		return
	}

	fmt.Fprintf(w, "&(so_Map){(%s[%d]){", keyType, size)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExpr(elt.(*ast.KeyValueExpr).Key)
	}
	fmt.Fprintf(w, "}, (%s[%d]){", valType, size)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExpr(elt.(*ast.KeyValueExpr).Value)
	}
	fmt.Fprintf(w, "}, %d, %d}", size, size)
}

// emitMapIndexExpr emits a map index read as so_map_get(K, V, m, key).
func (g *Generator) emitMapIndexExpr(n *ast.IndexExpr) {
	w := g.state.writer
	mapType := g.types.TypeOf(n.X).Underlying().(*types.Map)
	keyType := g.mapType(n, mapType.Key())
	valType := g.mapType(n, mapType.Elem())

	fmt.Fprintf(w, "so_map_get(%s, %s, ", keyType, valType)
	g.emitExpr(n.X)
	fmt.Fprintf(w, ", ")
	g.emitExpr(n.Index)
	fmt.Fprintf(w, ")")
}

// emitMapIndexAssign emits a map index write as so_map_set(K, V, &m, key, val).
func (g *Generator) emitMapIndexAssign(node ast.Node, idx *ast.IndexExpr, rhs ast.Expr) {
	w := g.state.writer
	mapType := g.types.TypeOf(idx.X).Underlying().(*types.Map)
	keyType := g.mapType(node, mapType.Key())
	valType := g.mapType(node, mapType.Elem())

	fmt.Fprintf(w, "%sso_map_set(%s, %s, ", g.indent(), keyType, valType)
	g.emitExpr(idx.X)
	fmt.Fprintf(w, ", ")
	g.emitExpr(idx.Index)
	fmt.Fprintf(w, ", ")
	g.emitExpr(rhs)
	fmt.Fprintf(w, ");\n")
}

// emitMapCommaOk emits a comma-ok map access: v, ok := m[key] or v, ok = m[key].
// Emits two statements: a so_map_get for the value and a so_map_has for the bool.
func (g *Generator) emitMapCommaOk(stmt *ast.AssignStmt, idx *ast.IndexExpr, isDefine bool) {
	w := g.state.writer
	mapType := g.types.TypeOf(idx.X).Underlying().(*types.Map)
	keyType := g.mapType(stmt, mapType.Key())
	valType := g.mapType(stmt, mapType.Elem())

	vIdent := stmt.Lhs[0].(*ast.Ident)
	okIdent := stmt.Lhs[1].(*ast.Ident)

	// Emit: [type] v = so_map_get(K, V, &m, key);
	if vIdent.Name != "_" {
		vDecl := ""
		if isDefine && g.types.Defs[vIdent] != nil {
			vDecl = valType + " "
		}
		fmt.Fprintf(w, "%s%s%s = so_map_get(%s, %s, ", g.indent(), vDecl, vIdent.Name, keyType, valType)
		g.emitExpr(idx.X)
		fmt.Fprintf(w, ", ")
		g.emitExpr(idx.Index)
		fmt.Fprintf(w, ");\n")
	}

	// Emit: [bool] ok = so_map_has(K, &m, key);
	if okIdent.Name != "_" {
		okDecl := ""
		if isDefine && g.types.Defs[okIdent] != nil {
			okDecl = "bool "
		}
		fmt.Fprintf(w, "%s%s%s = so_map_has(%s, ", g.indent(), okDecl, okIdent.Name, keyType)
		g.emitExpr(idx.X)
		fmt.Fprintf(w, ", ")
		g.emitExpr(idx.Index)
		fmt.Fprintf(w, ");\n")
	}
}

// emitMapRange emits a for-range loop over a map.
// Uses a hidden _i variable to iterate the internal arrays.
func (g *Generator) emitMapRange(stmt *ast.RangeStmt) {
	w := g.state.writer
	mapType := g.types.TypeOf(stmt.X).Underlying().(*types.Map)
	keyType := g.mapType(stmt, mapType.Key())
	valType := g.mapType(stmt, mapType.Elem())

	fmt.Fprintf(w, "%sfor (so_int _i = 0; _i < (so_int)", g.indent())
	g.emitExpr(stmt.X)
	fmt.Fprintf(w, "->len; _i++) {\n")

	g.state.indent++

	// Emit key variable.
	if stmt.Key != nil {
		if keyIdent, ok := stmt.Key.(*ast.Ident); ok && keyIdent.Name != "_" {
			keyDecl := keyType + " "
			if stmt.Tok == token.ASSIGN {
				keyDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = ((%s*)", g.indent(), keyDecl, keyIdent.Name, keyType)
			g.emitExpr(stmt.X)
			fmt.Fprintf(w, "->keys)[_i];\n")
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
			g.emitExpr(stmt.X)
			fmt.Fprintf(w, "->vals)[_i];\n")
		}
	}

	g.state.indent--

	g.emitBlock(stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// validateMapValueType fails if the map value type is not supported in C.
func (g *Generator) validateMapValueType(node ast.Node, valType types.Type) {
	if _, ok := valType.Underlying().(*types.Array); ok {
		g.fail(node, "array as map value type is not supported")
	}
}
