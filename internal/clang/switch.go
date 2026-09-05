package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
)

// emitSwitchStmt emits a switch statement, wrapping in a scope block
// if there's an init statement or a tag temporary.
func (g *Generator) emitSwitchStmt(w io.Writer, stmt *ast.SwitchStmt) {
	cases, def := splitCases(stmt)
	if stmt.Tag != nil {
		g.checkSwitchTag(stmt.Tag)
	}
	g.checkBreak(stmt)

	// The if/else chain repeats the tag in every comparison, but Go evaluates
	// it exactly once, so the tag goes into a temporary.
	hoist := stmt.Tag != nil && len(cases) > 0
	if stmt.Init == nil && !hoist {
		g.emitSwitchBody(w, stmt.Tag, nil, cases, def)
		return
	}

	// Wrap the switch in a scope block for the init statement and the temporary.
	fmt.Fprintf(w, "%s{\n", g.indent())
	g.state.depth++
	if stmt.Init != nil {
		g.walkAST(w, stmt.Init)
	}
	var tagRef *ast.Ident
	if hoist {
		tagRef = g.emitSwitchTagTemp(w, stmt.Tag)
	}
	g.emitSwitchBody(w, stmt.Tag, tagRef, cases, def)
	g.state.depth--
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitSwitchTagTemp declares a temporary holding the switch tag value and returns
// a reference to it. The reference carries the tag's type and position, so
// comparisons and diagnostics treat it like the tag itself.
func (g *Generator) emitSwitchTagTemp(w io.Writer, tag ast.Expr) *ast.Ident {
	typ := types.Default(g.types.TypeOf(tag))
	ref := &ast.Ident{NamePos: tag.Pos(), Name: g.newTemp(tag, tempSwitch)}
	// The reference has no declaration to look up, so record its type directly.
	g.types.Types[ref] = types.TypeAndValue{Type: typ}
	fmt.Fprintf(w, "%s%s = ", g.indent(), g.mapTypeDecl(tag, typ).Decl(ref.Name))
	g.emitExpr(w, tag)
	fmt.Fprint(w, ";\n")
	return ref
}

// emitSwitchBody emits the if/else-if/else chain for a switch statement.
// tagRef refers to the temporary holding the tag value, if there is one.
func (g *Generator) emitSwitchBody(w io.Writer, tag ast.Expr, tagRef *ast.Ident, cases []*ast.CaseClause, def *ast.CaseClause) {
	// No comparisons to emit, but the tag still runs.
	if len(cases) == 0 {
		if tag != nil {
			g.emitDiscard(w, tag)
		}
		if def == nil {
			// Empty switch.
			fmt.Fprintf(w, "%sif (false) {\n%s}\n", g.indent(), g.indent())
			return
		}
		// Default-only. The body keeps its own scope block, so that its
		// declarations do not conflict with the surrounding ones.
		fmt.Fprintf(w, "%s{\n", g.indent())
		g.state.depth++
		g.walkStmts(w, def.Body)
		g.state.depth--
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	// Emit if/else-if chain.
	for i, cc := range cases {
		if i == 0 {
			fmt.Fprintf(w, "%sif (", g.indent())
		} else {
			fmt.Fprintf(w, "%s} else if (", g.indent())
		}
		for j, expr := range cc.List {
			if j > 0 {
				fmt.Fprint(w, " || ")
			}
			if tagRef == nil {
				g.emitExpr(w, expr)
				continue
			}
			cond := &ast.BinaryExpr{X: tagRef, OpPos: expr.Pos(), Op: token.EQL, Y: g.groupCase(expr)}
			g.emitExpr(w, cond)
		}
		fmt.Fprint(w, ") {\n")
		g.state.depth++
		g.walkStmts(w, cc.Body)
		g.state.depth--
	}

	if def != nil {
		fmt.Fprintf(w, "%s} else {\n", g.indent())
		g.state.depth++
		g.walkStmts(w, def.Body)
		g.state.depth--
	}
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// checkBreak rejects a break that leaves a case body. A switch becomes
// an if/else chain, so emitting a break would lead to incorrect behavior
// or a compilation error.
func (g *Generator) checkBreak(stmt *ast.SwitchStmt) {
	for _, clause := range stmt.Body.List {
		for _, s := range clause.(*ast.CaseClause).Body {
			ast.Inspect(s, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
					// A break inside these leaves them, not the outer switch.
					return false
				case *ast.BranchStmt:
					if n.Tok == token.BREAK && n.Label == nil {
						g.fail(n, "break in a switch case is not supported, use a labeled break")
					}
				}
				return true
			})
		}
	}
}

// checkSwitchTag rejects a switch tag whose type has no place to go.
func (g *Generator) checkSwitchTag(tag ast.Expr) {
	switch types.Default(g.types.TypeOf(tag)).Underlying().(type) {
	case *types.Array:
		// C has no array assignment, so the value has nowhere to go.
		g.fail(tag, "switch on an array is not supported")
	case *types.Struct:
		// Case comparisons follow the rules of ==, which rejects structs.
		g.fail(tag, "switch on a struct is not supported")
	}
}

// groupCase parenthesizes a case expression so that comparing it to the tag
// keeps the Go grouping. Go treats the expression as a whole, but C would
// regroup it: && and || bind looser than ==, and == and != chain left to right.
func (g *Generator) groupCase(expr ast.Expr) ast.Expr {
	if _, ok := expr.(*ast.BinaryExpr); !ok {
		return expr
	}
	paren := &ast.ParenExpr{Lparen: expr.Pos(), X: expr, Rparen: expr.End()}
	// Like any other node, the wrapper has to carry the type of its value,
	// because the comparison reads the types of both operands.
	g.types.Types[paren] = g.types.Types[expr]
	return paren
}

// splitCases separates the case clauses of a switch from its default clause.
func splitCases(stmt *ast.SwitchStmt) (cases []*ast.CaseClause, def *ast.CaseClause) {
	for _, s := range stmt.Body.List {
		cc := s.(*ast.CaseClause)
		if cc.List == nil {
			def = cc
		} else {
			cases = append(cases, cc)
		}
	}
	return cases, def
}
