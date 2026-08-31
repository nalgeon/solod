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
	cStructType := g.symbolName(g.recvTypeObj(recv))
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
	selection := g.types.Selections[sel]
	// The receiver type can be an alias to a pointer, as in "type P = *T",
	// so it is unaliased before the pointer check.
	recv := types.Unalias(selection.Recv())
	sig := selection.Type().(*types.Signature)

	// Get the struct type name.
	var named *types.Named
	if ptr, ok := recv.(*types.Pointer); ok {
		named = types.Unalias(ptr.Elem()).(*types.Named)
	} else {
		named = types.Unalias(recv).(*types.Named)
	}

	// Method call: r.Area() → main_Rect_Area(&r).
	// An interface method call looks the same because the generator emits
	// a dispatch function for each interface method: s.Area() → main_Shape_Area(s).
	cStructType := g.mapTypeName(sel, named)
	cName := cStructType + "_" + sel.Sel.Name
	fmt.Fprintf(w, "%s(", cName)

	typeArgs := named.TypeArgs()
	isGeneric := typeArgs.Len() > 0
	// Prepend type arguments for generic method calls.
	if isGeneric {
		for i := 0; i < typeArgs.Len(); i++ {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, g.mapTypeName(sel, typeArgs.At(i)))
		}
		fmt.Fprint(w, ", ")
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

	// For generic (= macro) calls, wrap non-type args in parens to protect
	// against the preprocessor misinterpreting commas.
	lparen, rparen := "", ""
	if isGeneric {
		lparen, rparen = "(", ")"
	}

	fmt.Fprint(w, lparen)
	if isMethodPtrRecv {
		// Pointer receiver: pass address of value, or pointer directly.
		if isCallSitePtr {
			g.emitExpr(w, sel.X)
		} else {
			fmt.Fprint(w, "&")
			g.emitExpr(w, sel.X)
		}
	} else {
		// Value receiver: pass value directly, or dereference pointer.
		if isCallSitePtr {
			fmt.Fprint(w, "*")
			g.emitExpr(w, sel.X)
		} else {
			g.emitExpr(w, sel.X)
		}
	}
	fmt.Fprint(w, rparen)

	// Pass method arguments.
	g.emitMethodCallArgs(w, sel, call, sig, lparen, rparen)
	fmt.Fprint(w, ")")
}

// emitMethodCallArgs emits method arguments, handling variadic arg packing.
func (g *Generator) emitMethodCallArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature, lparen, rparen string) {
	args := call.Args

	if ext, ok := g.methodExtern(sel); ok {
		if !ext.nodecay {
			// Extern C method: decay args to C-compatible types.
			g.emitMethodExternArgs(w, sel, call, sig)
			return
		}
		if sig.Variadic() {
			// Extern nodecay method: emit the variadic args flat.
			g.emitMethodExternVarArgs(w, sel, call, sig)
			return
		}
	}

	if sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitMethodVarArgs(w, sel, call, sig, lparen, rparen)
		return
	}

	// Non-variadic call or variadic call with ellipsis: emit all args directly.
	for i, arg := range args {
		fmt.Fprintf(w, ", %s", lparen)
		g.emitCallArg(w, sel, arg, sig.Params().At(i).Type())
		fmt.Fprint(w, rparen)
	}
}

// emitMethodExternArgs emits method arguments for an extern C method, handling type decay.
func (g *Generator) emitMethodExternArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i, arg := range call.Args {
		fmt.Fprint(w, ", ")
		if i < sig.Params().Len() && isNamedNonEmptyInterface(sig.Params().At(i).Type()) {
			g.emitExprAsType(w, sel, arg, sig.Params().At(i).Type())
		} else {
			g.emitCArg(w, arg)
		}
	}
}

// emitMethodExternVarArgs emits the arguments of a call to a variadic extern
// nodecay method. See [Generator.emitExternVarArg] for the argument rules.
func (g *Generator) emitMethodExternVarArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i := range call.Args {
		fmt.Fprint(w, ", ")
		g.emitExternVarArg(w, sel, call, sig, i)
	}
}

func (g *Generator) emitMethodVarArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature, lparen, rparen string) {
	// Emit fixed args first.
	fixedCount := sig.Params().Len() - 1
	args := call.Args
	for i := 0; i < fixedCount && i < len(args); i++ {
		fmt.Fprintf(w, ", %s", lparen)
		g.emitCallArg(w, sel, args[i], sig.Params().At(i).Type())
		fmt.Fprint(w, rparen)
	}

	// Emit variadic args as a so_Slice literal.
	variadicArgs := args[fixedCount:]
	variadicParam := sig.Params().At(sig.Params().Len() - 1)
	elemType := g.mapTypeName(sel, variadicParam.Type().(*types.Slice).Elem())
	count := len(variadicArgs)
	targetType := variadicParam.Type().(*types.Slice).Elem()
	if count == 0 {
		// No variadic args: emit a nil slice.
		fmt.Fprintf(w, ", %s(so_Slice){}%s", lparen, rparen)
		return
	}

	fmt.Fprintf(w, ", %s(so_Slice){(%s[%d]){", lparen, elemType, count)
	for i, arg := range variadicArgs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExprAsType(w, sel, arg, targetType)
	}
	fmt.Fprintf(w, "}, %d, %d}%s", count, count, rparen)
}
