package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// emitMethodDecl emits a method as a C function.
// Pointer receivers use void* self with a cast; value receivers pass the struct by value.
func (g *Generator) emitMethodDecl(w io.Writer, decl *ast.FuncDecl) {
	sig := g.funcSig(decl)
	g.checkNamedReturns(decl, sig)

	// Init emission state.
	recv := decl.Recv.List[0]
	cStructType := g.mapObjName(g.recvTypeObj(recv))
	recvName, named := recvVarName(recv)
	_, isValueRecv := recv.Type.(*ast.Ident)

	g.state.enterFunc(decl, sig)
	defer g.state.leaveFunc()

	// Emit comments and function prototype.
	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	g.emitFuncProto(w, decl)
	fmt.Fprintln(w, " {")
	g.state.depth++

	// Emit receiver preamble.
	if isValueRecv {
		// Value receivers are passed by value - no cast needed.
		// Unnamed value receivers need (void)self to suppress unused warnings.
		if !named {
			fmt.Fprintf(w, "%s(void)self;\n", g.indent())
		}
	} else {
		// Pointer receivers: cast void* self to the concrete type.
		if named {
			fmt.Fprintf(w, "%s%s* %s = self;\n", g.indent(), cStructType, recvName)
		} else {
			fmt.Fprintf(w, "%s(void)self;\n", g.indent())
		}
	}

	// Emit method body, handling deferred calls if needed.
	g.walkStmts(w, decl.Body.List)
	if !endsWithReturn(decl.Body.List) {
		g.emitDeferredCalls(w)
	}
	g.state.depth--
	fmt.Fprint(w, "}\n")
}

// emitMethodCall emits a method call.
func (g *Generator) emitMethodCall(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr) {
	g.checkLocalCall(call)
	selection := g.types.Selections[sel]
	sig := selection.Type().(*types.Signature)

	// Get the struct type name.
	recv, _ := ptrElem(selection.Recv())
	named := recv.(*types.Named)

	// Method call: r.Area() → main_Rect_Area(&r).
	// An interface method call looks the same because the generator emits
	// a dispatch function for each interface method: s.Area() → main_Shape_Area(s).
	cStructType := g.mapTypeName(sel, named)
	cName := cStructType + "_" + sel.Sel.Name
	fmt.Fprintf(w, "%s(", cName)

	typeArgs := named.TypeArgs()
	args := g.callArgs(w, sel, typeArgs.Len() > 0)
	for t := range typeArgs.Types() {
		args.emitType(g.mapTypeName(sel, t))
	}

	// Pass receiver based on method's declared receiver type and call-site type.
	declSig := selection.Obj().Type().(*types.Signature)
	_, isMethodPtrRecv := declSig.Recv().Type().(*types.Pointer)
	xType := g.types.TypeOf(sel.X)
	_, isCallSitePtr := xType.Underlying().(*types.Pointer)

	// An array literal receiver emits a brace initializer,
	// which C rejects as a call argument.
	if _, ok := ast.Unparen(sel.X).(*ast.CompositeLit); ok {
		if _, ok := xType.Underlying().(*types.Array); ok {
			g.fail(sel, "cannot call a method on an array literal; assign it to a variable first")
		}
	}

	args.emit(func() {
		switch {
		case isMethodPtrRecv && !isCallSitePtr:
			// Pointer receiver on a value: pass the address of the value.
			fmt.Fprint(w, "&")
		case !isMethodPtrRecv && isCallSitePtr:
			// Value receiver on a pointer: pass the pointed-to value.
			fmt.Fprint(w, "*")
		}
		g.emitExpr(w, sel.X)
	})

	// Pass method arguments.
	ext, isExtern := g.methodExtern(sel)
	g.emitCallArgs(args, call, sig, ext, isExtern)
	fmt.Fprint(w, ")")
}
