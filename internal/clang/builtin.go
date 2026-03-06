package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

// emitAppendCall emits an append() builtin call.
func (g *Generator) emitAppendCall(call *ast.CallExpr) {
	w := g.state.writer
	sliceType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapType(call, sliceType.Elem())
	if call.Ellipsis.IsValid() {
		// Appending a slice (e.g. append(dst, src...)).
		fmt.Fprintf(w, "so_extend(%s, ", elemType)
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, ", (")
		g.emitExpr(call.Args[1])
		fmt.Fprintf(w, "))")
	} else {
		// Appending individual values (e.g. append(dst, v1, v2, v3)).
		fmt.Fprintf(w, "so_append(%s, ", elemType)
		g.emitExpr(call.Args[0])
		for _, arg := range call.Args[1:] {
			fmt.Fprintf(w, ", ")
			g.emitExpr(arg)
		}
		fmt.Fprintf(w, ")")
	}
}

// emitCopyCall emits a copy() builtin call as so_copy(T, dst, src).
func (g *Generator) emitCopyCall(call *ast.CallExpr) {
	w := g.state.writer
	dstType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapType(call, dstType.Elem())
	fmt.Fprintf(w, "so_copy(%s, ", elemType)
	g.emitExpr(call.Args[0])
	fmt.Fprintf(w, ", ")
	g.emitExpr(call.Args[1])
	fmt.Fprintf(w, ")")
}

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

// emitMinMaxCall emits a min() or max() builtin call.
// For numeric types: so_min(a, b) / so_max(a, b)
// For string types: so_string_min(a, b) / so_string_max(a, b)
// For 3+ args, nests calls: min(a, b, c) -> so_min(so_min(a, b), c)
func (g *Generator) emitMinMaxCall(call *ast.CallExpr, name string) {
	w := g.state.writer
	typ := g.types.TypeOf(call.Args[0])
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(call, "%s() requires a basic type, got %s", name, typ)
	}

	var fn string
	switch basic.Kind() {
	case types.String, types.UntypedString:
		fn = "so_string_" + name
	default:
		if basic.Info()&types.IsNumeric == 0 {
			g.fail(call, "%s() unsupported type: %s", name, typ)
		}
		fn = "so_" + name
	}

	// Emit nested calls for 2+ args: so_min(so_min(a, b), c)
	for i := 0; i < len(call.Args)-1; i++ {
		fmt.Fprintf(w, "%s(", fn)
	}
	g.emitExpr(call.Args[0])
	for _, arg := range call.Args[1:] {
		fmt.Fprintf(w, ", ")
		g.emitExpr(arg)
		fmt.Fprintf(w, ")")
	}
}

// emitNewCall emits a new() builtin call as a compound literal address.
func (g *Generator) emitNewCall(call *ast.CallExpr) {
	w := g.state.writer
	tv := g.types.Types[call.Args[0]]
	if tv.IsType() {
		// new(T) - zero-initialized compound literal.
		cType := g.mapType(call, tv.Type)
		fmt.Fprintf(w, "&(%s){0}", cType)
		return
	}
	if _, ok := call.Args[0].(*ast.CompositeLit); ok {
		// new(T{...}) - addressed composite literal.
		fmt.Fprintf(w, "&")
		g.emitExpr(call.Args[0])
		return
	}
	if _, ok := call.Args[0].(*ast.CallExpr); ok {
		g.fail(call, "new() with function call argument is not supported")
		return
	}
	// new(expr) - take address of the expression.
	elemType := g.types.TypeOf(call).(*types.Pointer).Elem()
	if _, ok := elemType.Underlying().(*types.Struct); ok {
		// Struct: take address directly.
		fmt.Fprintf(w, "&")
		g.emitExpr(call.Args[0])
		return
	}
	// Scalar: wrap in compound literal.
	cType := g.mapType(call, elemType)
	fmt.Fprintf(w, "&(%s){", cType)
	g.emitExpr(call.Args[0])
	fmt.Fprintf(w, "}")
}

// emitPrintCall emits a print/println call with an auto-generated format string.
func (g *Generator) emitPrintCall(call *ast.CallExpr, name string) {
	w := g.state.writer
	format := g.buildFormatString(call)
	fmt.Fprintf(w, "so_%s(%s", name, format)
	for _, arg := range call.Args {
		fmt.Fprintf(w, ", ")
		g.emitCArg(arg)
	}
	fmt.Fprintf(w, ")")
}

// buildFormatString constructs a C format string for the given print/println call,
// using the types of the arguments. It breaks out of string literals when macros
// are needed (e.g. "Value: %" PRId64) to avoid issues with macro expansion.
func (g *Generator) buildFormatString(call *ast.CallExpr) string {
	var format strings.Builder
	inStr := false
	for i, arg := range call.Args {
		spec, macro := g.formatSpec(call, g.types.TypeOf(arg))
		if !inStr {
			if format.Len() > 0 {
				format.WriteByte(' ')
			}
			format.WriteByte('"')
			inStr = true
		}
		if i > 0 {
			format.WriteByte(' ')
		}
		format.WriteString(spec)
		if macro != "" {
			format.WriteString(`" `)
			format.WriteString(macro)
			inStr = false
		}
	}
	if inStr {
		format.WriteByte('"')
	}
	return format.String()
}

// formatSpec returns the C printf format specifier and optional macro name
// for a Go type. When macro is non-empty (e.g. "PRId64"), the specifier
// ends with "%" and the macro must follow outside the string literal.
func (g *Generator) formatSpec(node ast.Node, typ types.Type) (spec, macro string) {
	if _, ok := typ.(*types.Pointer); ok {
		return "%p", ""
	}
	if isErrorType(typ) {
		return "%p", ""
	}
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported type for print: %s", typ)
		panic("unreachable")
	}
	switch basic.Kind() {
	case types.Bool:
		return "%d", ""
	case types.Float32, types.Float64, types.UntypedFloat:
		return "%f", ""
	case types.Int, types.UntypedInt:
		return "%", "PRId64"
	case types.Int8, types.Int16, types.Int32:
		return "%d", ""
	case types.Int64:
		return "%", "PRId64"
	case types.Uint8, types.Uint16, types.Uint32:
		return "%u", ""
	case types.Uint, types.Uint64, types.Uintptr:
		return "%", "PRIu64"
	case types.String:
		return "%s", ""
	default:
		g.fail(node, "unsupported type for print: %s", typ)
		panic("unreachable")
	}
}
