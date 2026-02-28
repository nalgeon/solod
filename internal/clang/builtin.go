package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// emitMakeCall emits a make() builtin call as so_make_slice(T, len, cap).
func (g *Generator) emitMakeCall(call *ast.CallExpr) {
	w := g.state.writer
	sliceType := g.types.Types[call.Args[0]].Type.Underlying().(*types.Slice)
	elemType := g.mapType(call, sliceType.Elem())
	fmt.Fprintf(w, "so_make_slice(%s, ", elemType)
	g.emitExpr(call.Args[1])
	fmt.Fprintf(w, ", ")
	if len(call.Args) >= 3 {
		g.emitExpr(call.Args[2])
	} else {
		g.emitExpr(call.Args[1])
	}
	fmt.Fprintf(w, ")")
}

// emitAppendCall emits an append() builtin call as so_append(slice, T, vals...).
func (g *Generator) emitAppendCall(call *ast.CallExpr) {
	w := g.state.writer
	sliceType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapType(call, sliceType.Elem())
	fmt.Fprintf(w, "so_append(")
	g.emitExpr(call.Args[0])
	fmt.Fprintf(w, ", %s", elemType)
	for _, arg := range call.Args[1:] {
		fmt.Fprintf(w, ", ")
		g.emitExpr(arg)
	}
	fmt.Fprintf(w, ")")
}

// emitCopyCall emits a copy() builtin call as so_copy(dst, src, T).
func (g *Generator) emitCopyCall(call *ast.CallExpr) {
	w := g.state.writer
	dstType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapType(call, dstType.Elem())
	fmt.Fprintf(w, "so_copy(")
	g.emitExpr(call.Args[0])
	fmt.Fprintf(w, ", ")
	g.emitExpr(call.Args[1])
	fmt.Fprintf(w, ", %s)", elemType)
}

// emitPrintCall emits a print/println call with an auto-generated format string.
func (g *Generator) emitPrintCall(call *ast.CallExpr, name string) {
	w := g.state.writer

	var specs []string
	for _, arg := range call.Args {
		typ := g.types.TypeOf(arg)
		specs = append(specs, g.formatSpec(call, typ))
	}

	format := strings.Join(specs, " ")
	fmt.Fprintf(w, "so_%s(\"%s\"", name, format)
	for _, arg := range call.Args {
		fmt.Fprintf(w, ", ")
		g.emitPrintArg(arg)
	}
	fmt.Fprintf(w, ")")
}

// emitPrintArg emits a single argument for a print/println call.
func (g *Generator) emitPrintArg(arg ast.Expr) {
	w := g.state.writer
	typ := g.types.TypeOf(arg)
	if basic, ok := typ.Underlying().(*types.Basic); ok && basic.Kind() == types.String {
		// Special handling for strings.
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// String literals are emitted as raw C strings.
			fmt.Fprintf(w, "%s", lit.Value)
		} else {
			// String variables are emitted as .ptr to get the C string pointer.
			g.emitExpr(arg)
			fmt.Fprintf(w, ".ptr")
		}
		return
	}
	// All other types are emitted normally.
	g.emitExpr(arg)
}

// formatSpec returns the C printf format specifier for a Go type.
func (g *Generator) formatSpec(node ast.Node, typ types.Type) string {
	if _, ok := typ.(*types.Pointer); ok {
		return "%p"
	}
	if isErrorType(typ) {
		return "%p"
	}
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported type for print: %s", typ)
		panic("unreachable")
	}
	switch basic.Kind() {
	case types.Bool:
		return "%d"
	case types.Float32, types.Float64, types.UntypedFloat:
		return "%f"
	case types.Int, types.UntypedInt:
		return "%lld"
	case types.Int8, types.Int16, types.Int32:
		return "%d"
	case types.Int64:
		return "%lld"
	case types.Uint8, types.Uint16, types.Uint32:
		return "%u"
	case types.Uint, types.Uint64, types.Uintptr:
		return "%llu"
	case types.String:
		return "%s"
	default:
		g.fail(node, "unsupported type for print: %s", typ)
		panic("unreachable")
	}
}
