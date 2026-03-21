package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// emitAssignStmt emits an assignment statement.
func (g *Generator) emitAssignStmt(stmt *ast.AssignStmt) {
	switch stmt.Tok {
	case token.DEFINE:
		g.emitDefine(stmt)

	case token.ASSIGN:
		g.emitAssign(stmt)

	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.OR_ASSIGN, token.AND_ASSIGN, token.XOR_ASSIGN,
		token.SHL_ASSIGN, token.SHR_ASSIGN:
		if idx, ok := stmt.Lhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.fail(stmt, "compound assignment on map index is not supported")
			}
		}
		w := g.state.writer
		// String += uses so_string_add.
		if stmt.Tok == token.ADD_ASSIGN && g.hasStringType(stmt.Lhs[0]) {
			fmt.Fprintf(w, "%s", g.indent())
			g.emitExpr(stmt.Lhs[0])
			fmt.Fprintf(w, " = so_string_add(")
			g.emitExpr(stmt.Lhs[0])
			fmt.Fprintf(w, ", ")
			g.emitExpr(stmt.Rhs[0])
			fmt.Fprintf(w, ");\n")
			return
		}
		fmt.Fprintf(w, "%s", g.indent())
		g.emitExpr(stmt.Lhs[0])
		fmt.Fprintf(w, " %s ", stmt.Tok)
		g.emitExpr(stmt.Rhs[0])
		fmt.Fprintf(w, ";\n")

	default:
		g.fail(stmt, "unsupported AssignStmt token: %s", stmt.Tok)
	}
}

// emitDefine emits a short variable declaration (:=).
func (g *Generator) emitDefine(stmt *ast.AssignStmt) {
	w := g.state.writer
	// Detect: _, ok := s.(Rect)
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if ta, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok {
			g.emitTypeAssertion(w, stmt, ta)
			return
		}
	}
	// Map comma-ok: v, ok := m[key]
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if idx, ok := stmt.Rhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapCommaOk(stmt, idx, true)
				return
			}
		}
	}
	// Multi-return destructuring: x, y := f()
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
		if call, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			g.emitMultiReturnDefine(stmt, call)
			return
		}
	}
	// Regular define: group consecutive variables by type.
	i := 0
	for i < len(stmt.Lhs) {
		ident := stmt.Lhs[i].(*ast.Ident)
		if ident.Name == "_" {
			i++
			continue
		}

		def := g.types.Defs[ident]
		if def == nil {
			// Redeclared variable - emit plain assignment.
			fmt.Fprintf(w, "%s%s = ", g.indent(), ident.Name)
			g.emitExpr(stmt.Rhs[i])
			fmt.Fprintf(w, ";\n")
			i++
			continue
		}

		typ := def.Type()
		ct := g.mapCType(stmt, typ)

		if ct.IsArray() {
			// Arrays can't be grouped with other variables.
			if _, isLit := stmt.Rhs[i].(*ast.CompositeLit); isLit {
				// Composite literal: so_int d[3] = {1, 2, 3};
				fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ident.Name))
				g.emitExpr(stmt.Rhs[i])
				fmt.Fprintf(w, ";\n")
			} else {
				// Variable: declaration + memcpy.
				fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(ident.Name))
				fmt.Fprintf(w, "%smemcpy(%s, ", g.indent(), ident.Name)
				g.emitExpr(stmt.Rhs[i])
				fmt.Fprintf(w, ", sizeof(%s));\n", ident.Name)
			}
			i++
			continue
		}

		// Emit a variable declaration for this variable
		// (grouped with subsequent variables of the same type).
		fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ident.Name))
		g.emitExpr(stmt.Rhs[i])
		i++
		for i < len(stmt.Lhs) {
			nextIdent := stmt.Lhs[i].(*ast.Ident)
			if nextIdent.Name == "_" {
				break
			}
			nextDef := g.types.Defs[nextIdent]
			if nextDef == nil {
				break
			}
			nextCType := g.mapType(stmt, nextDef.Type())
			if nextCType != ct.Base {
				break
			}
			if isArrayType(nextDef.Type()) {
				break
			}
			fmt.Fprintf(w, ", %s = ", nextIdent.Name)
			g.emitExpr(stmt.Rhs[i])
			i++
		}
		fmt.Fprintf(w, ";\n")
	}
}

// emitAssign emits a regular assignment (=).
func (g *Generator) emitAssign(stmt *ast.AssignStmt) {
	w := g.state.writer
	// Detect: _, ok = s.(Rect)
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if ta, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok {
			g.emitTypeAssertion(w, stmt, ta)
			return
		}
	}
	// Map comma-ok: v, ok = m[key]
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if idx, ok := stmt.Rhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapCommaOk(stmt, idx, false)
				return
			}
		}
	}
	// Multi-return destructuring: x, y = f()
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
		if call, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			g.emitMultiReturnAssign(stmt, call)
			return
		}
	}
	// Regular assignment.
	for i, lhs := range stmt.Lhs {
		// Blank identifier - emit a void expression.
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			fmt.Fprintf(w, "%s(void)", g.indent())
			if g.needsVoidParens(stmt.Rhs[i]) {
				fmt.Fprintf(w, "(")
				g.emitExpr(stmt.Rhs[i])
				fmt.Fprintf(w, ")")
			} else {
				g.emitExpr(stmt.Rhs[i])
			}
			fmt.Fprintf(w, ";\n")
			continue
		}

		// Map index assignment uses so_map_set.
		if idx, ok := lhs.(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapIndexAssign(stmt, idx, stmt.Rhs[i])
				continue
			}
		}

		// Array assignment uses memcpy.
		lhsType := g.types.TypeOf(lhs)
		if arr, ok := lhsType.Underlying().(*types.Array); ok {
			fmt.Fprintf(w, "%smemcpy(", g.indent())
			g.emitExpr(lhs)
			fmt.Fprintf(w, ", ")
			if _, isLit := stmt.Rhs[i].(*ast.CompositeLit); isLit {
				// Compound literal: (int[3]){1, 2, 3}
				elemType := g.mapType(stmt, arr.Elem())
				fmt.Fprintf(w, "(%s%s)", elemType, arrayDims(arr))
			}
			g.emitExpr(stmt.Rhs[i])
			fmt.Fprintf(w, ", sizeof(")
			g.emitExpr(lhs)
			fmt.Fprintf(w, "));\n")
			continue
		}

		// Non-array assignment.
		fmt.Fprintf(w, "%s", g.indent())
		g.emitExpr(lhs)
		fmt.Fprintf(w, " = ")
		g.emitExprAsType(stmt, stmt.Rhs[i], lhsType)
		fmt.Fprintf(w, ";\n")
	}
}
