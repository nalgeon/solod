package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// callArgs writes the arguments of a call. It holds the separator between
// the arguments, so a caller emits an argument without knowing its position.
type callArgs struct {
	g     *Generator
	w     io.Writer
	node  ast.Node // the position for a diagnostic
	macro bool     // the call emits as a macro
	count int      // the arguments written so far
}

// callArgs returns a writer for the arguments of a call.
// node gives the position of a diagnostic about an argument.
func (g *Generator) callArgs(w io.Writer, node ast.Node, macro bool) *callArgs {
	return &callArgs{g: g, w: w, node: node, macro: macro}
}

// emit writes one value argument. The write function emits the value itself.
func (a *callArgs) emit(write func()) {
	a.separate()
	if a.macro {
		fmt.Fprint(a.w, "(")
	}
	write()
	if a.macro {
		fmt.Fprint(a.w, ")")
	}
}

// emitType writes one type argument of a macro call.
func (a *callArgs) emitType(name string) {
	a.separate()
	fmt.Fprint(a.w, name)
}

// emitExpr writes one value argument coerced to the parameter type.
func (a *callArgs) emitExpr(arg ast.Expr, paramType types.Type) {
	a.emit(func() { a.g.emitCallArg(a.w, a.node, arg, paramType) })
}

// separate writes the comma before an argument that follows another one.
func (a *callArgs) separate() {
	if a.count > 0 {
		fmt.Fprint(a.w, ", ")
	}
	a.count++
}

// emitCallArgs writes the arguments of a function or a method call. ext holds
// the extern metadata of the callee, and isExtern reports whether the callee is
// extern. sig is nil for a call through a value with no signature.
func (g *Generator) emitCallArgs(args *callArgs, call *ast.CallExpr, sig *types.Signature, ext externInfo, isExtern bool) {
	if isExtern {
		if !ext.nodecay {
			// Extern C function: decay args to C-compatible types.
			g.emitDecayArgs(args, call, sig)
			return
		}
		if sig != nil && sig.Variadic() {
			// Extern nodecay function: emit the variadic args flat.
			g.emitExternVarArgs(args, call, sig)
			return
		}
	}
	if sig != nil && sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitVarArgs(args, call, sig)
		return
	}
	// Non-variadic call or variadic call with ellipsis: emit all args directly.
	for i, arg := range call.Args {
		if sig != nil && i < sig.Params().Len() {
			args.emitExpr(arg, sig.Params().At(i).Type())
		} else {
			// No signature available (e.g. func literal), emit arg as-is.
			args.emit(func() { g.emitExpr(args.w, arg) })
		}
	}
}

// emitVarArgs packs the trailing arguments into an inline so_Slice literal.
func (g *Generator) emitVarArgs(args *callArgs, call *ast.CallExpr, sig *types.Signature) {
	// Emit fixed args first.
	fixedCount := sig.Params().Len() - 1
	for i := 0; i < fixedCount && i < len(call.Args); i++ {
		args.emitExpr(call.Args[i], sig.Params().At(i).Type())
	}

	// Emit variadic args as a so_Slice literal.
	variadicArgs := call.Args[fixedCount:]
	elemType := sig.Params().At(fixedCount).Type().(*types.Slice).Elem()
	cElemType := g.mapTypeName(args.node, elemType)
	count := len(variadicArgs)

	if count == 0 {
		// No variadic args: emit a nil slice.
		args.emit(func() { fmt.Fprint(args.w, "(so_Slice){}") })
		return
	}

	args.emit(func() {
		fmt.Fprintf(args.w, "(so_Slice){(%s[%d]){", cElemType, count)
		// The slice literal already protects its elements from the
		// preprocessor, so an element needs no parentheses of its own.
		elems := g.callArgs(args.w, args.node, false)
		for _, arg := range variadicArgs {
			elems.emit(func() { g.emitExprAsType(args.w, args.node, arg, elemType) })
		}
		fmt.Fprintf(args.w, "}, %d, %d}", count, count)
	})
}

// emitDecayArgs writes the arguments of a call to an extern C function,
// decayed to their C-compatible types.
func (g *Generator) emitDecayArgs(args *callArgs, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i, arg := range call.Args {
		// Interface-typed parameters (e.g. Allocator) need emitExprAsType
		// to convert nil to a zero-initialized struct instead of NULL.
		if sig != nil && i < sig.Params().Len() && isNamedNonEmptyInterface(sig.Params().At(i).Type()) {
			paramType := sig.Params().At(i).Type()
			args.emit(func() { g.emitExprAsType(args.w, args.node, arg, paramType) })
		} else {
			args.emit(func() { g.emitCArg(args.w, arg) })
		}
	}
}

// emitExternVarArgs writes the arguments of a call to a variadic extern nodecay function.
func (g *Generator) emitExternVarArgs(args *callArgs, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i := range call.Args {
		args.emit(func() { g.emitExternVarArg(args.w, args.node, call, sig, i) })
	}
}
