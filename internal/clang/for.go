package clang

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// emitForStmt emits a for statement.
func (g *Generator) emitForStmt(w io.Writer, stmt *ast.ForStmt) {
	if stmt.Post != nil && !g.fitsForClause(stmt.Post, false) {
		g.fail(stmt.Post, "a for post statement must be an increment, a call, or a scalar assignment")
	}

	// An init statement that needs more than one C statement
	// goes before the loop. A block statement keeps the scope.
	if stmt.Init != nil && !g.fitsForClause(stmt.Init, true) {
		inner := *stmt
		inner.Init = nil
		fmt.Fprintf(w, "%s{\n", g.indent())
		g.state.depth++
		g.walkAST(w, stmt.Init)
		g.emitForStmt(w, &inner)
		g.state.depth--
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	fmt.Fprintf(w, "%sfor (", g.indent())

	if stmt.Init != nil {
		g.emitForClause(w, stmt.Init)
	}
	fmt.Fprint(w, ";")

	if stmt.Cond != nil {
		fmt.Fprint(w, " ")
		g.emitExpr(w, stmt.Cond)
	}
	fmt.Fprint(w, ";")

	if stmt.Post != nil {
		fmt.Fprint(w, " ")
		g.emitForClause(w, stmt.Post)
	}

	fmt.Fprint(w, ") {\n")
	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitForClause emits a simple statement inside a for clause.
func (g *Generator) emitForClause(w io.Writer, stmt ast.Stmt) {
	var buf bytes.Buffer
	g.walkAST(&buf, stmt)
	text := strings.TrimSpace(buf.String())
	fmt.Fprint(w, strings.TrimSuffix(text, ";"))
}

// fitsForClause reports whether the emitted C for stmt fits in a for clause.
func (g *Generator) fitsForClause(stmt ast.Stmt, isInit bool) bool {
	switch s := stmt.(type) {
	case *ast.IncDecStmt:
		return true
	case *ast.ExprStmt:
		if _, ok := g.cIntrinsic(s.X); ok {
			return false
		}
		return !g.isPanicCall(s.X)
	case *ast.AssignStmt:
		return g.fitsForAssign(s, isInit)
	}
	return false
}

// fitsForAssign reports whether the emitted C for an assignment fits in a for clause.
func (g *Generator) fitsForAssign(stmt *ast.AssignStmt, isInit bool) bool {
	if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return false
	}
	lhs := ast.Unparen(stmt.Lhs[0])

	if stmt.Tok == token.DEFINE {
		if !isInit {
			return false
		}
		// No arrays.
		ident, ok := lhs.(*ast.Ident)
		if !ok || g.types.Defs[ident] == nil {
			return false
		}
		return !isArrayType(g.types.Defs[ident].Type())
	}

	// No blank identifiers.
	if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
		return true
	}
	// No map items.
	if idx, ok := lhs.(*ast.IndexExpr); ok && isMapType(g.types.TypeOf(idx.X)) {
		return false
	}
	// Only scalar types.
	return isScalarType(g.types.TypeOf(lhs))
}
