package clang

import (
	"fmt"
	"go/ast"
	"go/token"
)

// emitAssignStmt emits an assignment statement.
func (g *Generator) emitAssignStmt(stmt *ast.AssignStmt) {
	switch stmt.Tok {
	case token.DEFINE:
		w := g.state.writer
		// Detect: _, ok := s.(Rect)
		if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
			if ta, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok {
				g.emitTypeAssertion(w, stmt, ta)
				return
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
			cType := g.mapType(stmt, typ)
			fmt.Fprintf(w, "%s%s %s = ", g.indent(), cType, ident.Name)
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
				if nextCType != cType {
					break
				}
				fmt.Fprintf(w, ", %s = ", nextIdent.Name)
				g.emitExpr(stmt.Rhs[i])
				i++
			}
			fmt.Fprintf(w, ";\n")
		}

	case token.ASSIGN:
		w := g.state.writer
		// Multi-return destructuring: x, y = f()
		if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
			if call, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
				g.emitMultiReturnAssign(stmt, call)
				return
			}
		}
		// Regular assignment.
		for i, lhs := range stmt.Lhs {
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
			fmt.Fprintf(w, "%s", g.indent())
			g.emitExpr(lhs)
			fmt.Fprintf(w, " = ")
			g.emitExpr(stmt.Rhs[i])
			fmt.Fprintf(w, ";\n")
		}

	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.OR_ASSIGN, token.AND_ASSIGN, token.XOR_ASSIGN,
		token.SHL_ASSIGN, token.SHR_ASSIGN:
		w := g.state.writer
		fmt.Fprintf(w, "%s", g.indent())
		g.emitExpr(stmt.Lhs[0])
		fmt.Fprintf(w, " %s ", stmt.Tok)
		g.emitExpr(stmt.Rhs[0])
		fmt.Fprintf(w, ";\n")

	default:
		g.fail(stmt, "unsupported AssignStmt token: %s", stmt.Tok)
	}
}

