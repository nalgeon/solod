package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// emitReturnStmt emits a return statement, preceded by any deferred generic calls.
func (g *Generator) emitReturnStmt(w io.Writer, stmt *ast.ReturnStmt) {
	if g.state.inMacro {
		// In macro mode: "return X" becomes just "X;", void return is a no-op.
		if len(stmt.Results) > 0 {
			fmt.Fprint(w, g.indent())
			g.emitReturnExpr(w, stmt)
			fmt.Fprint(w, ";\n")
		}
		return
	}

	// When defers are active and the return value is non-constant, evaluate it
	// into a temp before running the deferred calls, so the value is captured
	// before the defers (matching Go, which evaluates the return value first).
	if len(stmt.Results) > 0 && len(g.state.defers) > 0 && g.returnIsNotConst(stmt) {
		tmp := g.newTemp(stmt, tempResult)
		retType := g.returnType(stmt, g.state.funcSig)
		fmt.Fprintf(w, "%s%s %s = ", g.indent(), retType, tmp)
		g.emitReturnExpr(w, stmt)
		fmt.Fprint(w, ";\n")
		g.emitDeferredCalls(w)
		fmt.Fprintf(w, "%sreturn %s;\n", g.indent(), tmp)
		return
	}

	g.emitDeferredCalls(w)

	if len(stmt.Results) == 0 {
		fmt.Fprintf(w, "%sreturn;\n", g.indent())
		return
	}

	fmt.Fprintf(w, "%sreturn ", g.indent())
	g.emitReturnExpr(w, stmt)
	fmt.Fprint(w, ";\n")
}

// emitReturnExpr emits the return value expression (without "return" keyword or ";").
// Handles single-return and multi-return compound literals.
func (g *Generator) emitReturnExpr(w io.Writer, stmt *ast.ReturnStmt) {
	// Single return value: emit directly.
	if len(stmt.Results) == 1 {
		// Forwarding a multi-value call ("return f()" where f returns a tuple):
		// the call already yields the whole result struct, so emit it as-is
		// without per-result type conversion.
		if _, ok := g.types.TypeOf(stmt.Results[0]).(*types.Tuple); ok {
			g.emitExpr(w, stmt.Results[0])
			return
		}
		retType := g.state.funcSig.Results().At(0).Type()
		g.emitExprAsType(w, stmt, stmt.Results[0], retType)
		return
	}

	// Multi-return: emit compound literal with per-signature result fields.
	// Each value converts to the result type of its position.
	multi := g.makeMultiReturn(stmt, g.state.funcSig)
	fmt.Fprintf(w, "(%s){", multi.typeName())
	for i, res := range stmt.Results {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, ".%s = ", multi.field(i))
		g.emitExprAsType(w, stmt, res, g.state.funcSig.Results().At(i).Type())
	}
	fmt.Fprint(w, "}")
}

// checkNamedReturns rejects any return value that has a name.
func (g *Generator) checkNamedReturns(node ast.Node, sig *types.Signature) {
	for v := range sig.Results().Variables() {
		if v.Name() != "" {
			g.fail(node, "named return values are not supported")
		}
	}
}

// returnType returns the C return type for a function signature.
// For multi-return (T, error) or (T, T), returns the per-signature result type.
// For single return, maps the Go type to C. For no return, returns "void".
func (g *Generator) returnType(node ast.Node, sig *types.Signature) string {
	if sig.Results().Len() > 1 {
		multi := g.makeMultiReturn(node, sig)
		return multi.typeName()
	}
	if sig.Results().Len() == 1 {
		ret := sig.Results().At(0).Type()
		if arr, ok := ret.Underlying().(*types.Array); ok {
			if _, ok := arr.Elem().(*types.Array); ok {
				g.fail(node, "returning multi-dimensional arrays is not supported")
			}
			return g.mapTypeName(node, arr) + "*"
		}
		return g.mapTypeName(node, ret)
	}
	return "void"
}

// returnIsNotConst reports whether any return value
// in the statement is not a compile-time constant.
func (g *Generator) returnIsNotConst(stmt *ast.ReturnStmt) bool {
	for _, r := range stmt.Results {
		if tv, ok := g.types.Types[r]; ok && tv.Value != nil {
			continue // compile-time constant
		}
		return true // non-constant
	}
	return false
}
