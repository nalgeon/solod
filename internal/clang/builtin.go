package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// emitBuiltin handles a builtin function call. Returns true if the call
// was fully emitted (including arguments), false if only the function
// name was emitted and the caller must still emit arguments.
func (g *Generator) emitBuiltin(w io.Writer, call *ast.CallExpr, ident *ast.Ident, bi *types.Builtin) bool {
	switch bi.Name() {
	case "append":
		g.emitAppendCall(w, call)
		return true
	case "clear":
		g.emitClearCall(w, call)
		return true
	case "close", "complex", "delete", "imag", "real", "recover":
		g.fail(call, "%s() is not supported", bi.Name())
		return true
	case "copy":
		g.emitCopyCall(w, call)
		return true
	case "make":
		g.emitMakeCall(w, call)
		return true
	case "min", "max":
		g.emitMinMaxCall(w, call, bi.Name())
		return true
	case "new":
		g.emitNewCall(w, call)
		return true
	case "panic":
		g.emitPanicCall(w, call)
		return true
	case "print", "println":
		g.emitPrintCall(w, call, bi.Name())
		return true
	}

	// len/cap on arrays emit the compile-time size.
	if (bi.Name() == "len" || bi.Name() == "cap") && len(call.Args) == 1 {
		if size := arraySize(g.types.TypeOf(call.Args[0])); size >= 0 {
			fmt.Fprintf(w, "%d", size)
			return true
		}
	}

	// len on maps emits m->len since maps are pointers.
	if bi.Name() == "len" && len(call.Args) == 1 {
		if _, ok := g.types.TypeOf(call.Args[0]).Underlying().(*types.Map); ok {
			g.emitExpr(w, call.Args[0])
			fmt.Fprint(w, "->len")
			return true
		}
	}

	// len/cap on slices (and len on strings) emit so_len/so_cap macros.
	if bi.Name() == "len" || bi.Name() == "cap" {
		fmt.Fprintf(w, "so_%s(", ident.Name)
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ")")
		return true
	}

	// Other builtins are emitted as regular calls with a so_ prefix.
	fmt.Fprintf(w, "so_%s", ident.Name)
	return false
}

// emitAppendCall emits an append() builtin call.
func (g *Generator) emitAppendCall(w io.Writer, call *ast.CallExpr) {
	sliceType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapTypeName(call, sliceType.Elem())
	if call.Ellipsis.IsValid() {
		srcType := g.types.TypeOf(call.Args[1]).Underlying()
		if basic, ok := srcType.(*types.Basic); ok && basic.Info()&types.IsString != 0 {
			// Appending a string to a byte slice (e.g. append(dst, str...)).
			fmt.Fprintf(w, "so_extend(%s, ", elemType)
			g.emitMacroArg(w, call.Args[0])
			fmt.Fprint(w, ", so_string_bytes(")
			g.emitExpr(w, call.Args[1])
			fmt.Fprint(w, "))")
		} else {
			// Appending a slice (e.g. append(dst, src...)).
			fmt.Fprintf(w, "so_extend(%s, ", elemType)
			g.emitMacroArg(w, call.Args[0])
			fmt.Fprint(w, ", (")
			g.emitExpr(w, call.Args[1])
			fmt.Fprint(w, "))")
		}
	} else {
		// Appending individual values (e.g. append(dst, v1, v2, v3)).
		fmt.Fprintf(w, "so_append(%s, ", elemType)
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprintf(w, ", %d", len(call.Args)-1)
		for _, arg := range call.Args[1:] {
			fmt.Fprint(w, ", ")
			g.emitMacroArg(w, arg)
		}
		fmt.Fprint(w, ")")
	}
}

// emitClearCall emits a clear() builtin call as so_clear(T, s).
func (g *Generator) emitClearCall(w io.Writer, call *ast.CallExpr) {
	typ := g.types.TypeOf(call.Args[0]).Underlying()
	if _, ok := typ.(*types.Map); ok {
		g.fail(call, "clear() is not supported for maps")
	}
	sliceType := typ.(*types.Slice)
	elemType := g.mapTypeName(call, sliceType.Elem())
	fmt.Fprintf(w, "so_clear(%s, ", elemType)
	g.emitMacroArg(w, call.Args[0])
	fmt.Fprint(w, ")")
}

// emitCopyCall emits a copy() builtin call as so_copy(T, dst, src).
func (g *Generator) emitCopyCall(w io.Writer, call *ast.CallExpr) {
	srcType := g.types.TypeOf(call.Args[1]).Underlying()
	if basic, ok := srcType.(*types.Basic); ok && basic.Info()&types.IsString != 0 {
		// copy([]byte, string) - copy bytes directly from string.
		fmt.Fprint(w, "so_copy_string(")
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ", ")
		g.emitMacroArg(w, call.Args[1])
		fmt.Fprint(w, ")")
		return
	}
	// copy([]T, []T) - copy elements of any slice type.
	dstType := g.types.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	elemType := g.mapTypeName(call, dstType.Elem())
	fmt.Fprintf(w, "so_copy(%s, ", elemType)
	g.emitMacroArg(w, call.Args[0])
	fmt.Fprint(w, ", ")
	g.emitMacroArg(w, call.Args[1])
	fmt.Fprint(w, ")")
}

// emitMakeCall emits a make() builtin call for slices or maps.
func (g *Generator) emitMakeCall(w io.Writer, call *ast.CallExpr) {
	typ := g.types.Types[call.Args[0]].Type.Underlying()

	switch t := typ.(type) {
	case *types.Slice:
		g.checkSliceElemType(call, t.Elem())
		elemType := g.mapTypeName(call, t.Elem())
		fmt.Fprintf(w, "so_make_slice(%s, ", elemType)
		g.emitExpr(w, call.Args[1])
		fmt.Fprint(w, ", ")
		if len(call.Args) >= 3 {
			g.emitExpr(w, call.Args[2])
		} else {
			g.emitExpr(w, call.Args[1])
		}
		fmt.Fprint(w, ")")

	case *types.Map:
		if len(call.Args) < 2 {
			g.fail(call, "make(map) requires a capacity argument")
		}
		g.checkMapValueType(call, t.Elem())
		keyType := g.mapTypeName(call, t.Key())
		valType := g.mapTypeName(call, t.Elem())
		fmt.Fprintf(w, "so_make_map(%s, %s, ", keyType, valType)
		g.emitExpr(w, call.Args[1])
		fmt.Fprint(w, ")")

	default:
		g.fail(call, "make() unsupported type: %s", g.typeString(typ))
	}
}

// emitMinMaxCall emits a min() or max() builtin call.
// For numeric types: so_min(a, b) / so_max(a, b)
// For string types: so_string_min(a, b) / so_string_max(a, b)
// For 3+ args, nests calls: min(a, b, c) -> so_min(so_min(a, b), c)
func (g *Generator) emitMinMaxCall(w io.Writer, call *ast.CallExpr, name string) {
	typ := g.types.TypeOf(call.Args[0])
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(call, "%s() requires a basic type, got %s", name, g.typeString(typ))
	}

	var fn string
	switch basic.Kind() {
	case types.String, types.UntypedString:
		fn = "so_string_" + name
	default:
		if basic.Info()&types.IsNumeric == 0 {
			g.fail(call, "%s() unsupported type: %s", name, g.typeString(typ))
		}
		fn = "so_" + name
	}

	// Emit nested calls for 2+ args: so_min(so_min(a, b), c)
	for i := 0; i < len(call.Args)-1; i++ {
		fmt.Fprintf(w, "%s(", fn)
	}
	g.emitExpr(w, call.Args[0])
	for _, arg := range call.Args[1:] {
		fmt.Fprint(w, ", ")
		g.emitExpr(w, arg)
		fmt.Fprint(w, ")")
	}
}

// emitNewCall emits a new() builtin call as a compound literal address.
func (g *Generator) emitNewCall(w io.Writer, call *ast.CallExpr) {
	tv := g.types.Types[call.Args[0]]
	if tv.IsType() {
		// new(T) - zero-initialized compound literal. An array type keeps its
		// dimensions, so the address is a pointer to the array and not to the
		// first element.
		cType := g.mapTypeName(call, tv.Type)
		fmt.Fprintf(w, "&(%s%s){}", cType, arrayDims(tv.Type))
		return
	}
	if _, ok := call.Args[0].(*ast.CompositeLit); ok {
		// new(T{...}) - addressed composite literal.
		fmt.Fprint(w, "&")
		if arr, ok := g.types.TypeOf(call.Args[0]).Underlying().(*types.Array); ok {
			// An array literal carries no type token of its own,
			// so the generator adds one.
			g.emitArrayArg(w, call, call.Args[0], arr)
			return
		}
		g.emitExpr(w, call.Args[0])
		return
	}
	if _, ok := call.Args[0].(*ast.CallExpr); ok {
		g.fail(call, "new() with function call argument is not supported")
		return
	}
	// new(expr) - take address of the expression.
	elemType := g.types.TypeOf(call).(*types.Pointer).Elem()
	switch elemType.Underlying().(type) {
	case *types.Struct, *types.Array:
		// C cannot copy a struct or an array in a compound literal,
		// so the generator takes the address of the expression.
		fmt.Fprint(w, "&")
		g.emitExpr(w, call.Args[0])
		return
	}
	// Scalar: wrap in compound literal.
	cType := g.mapTypeName(call, elemType)
	fmt.Fprintf(w, "&(%s){", cType)
	g.emitExpr(w, call.Args[0])
	fmt.Fprint(w, "}")
}

// isPanicCall reports whether an expression is a call to the panic builtin.
func (g *Generator) isPanicCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	bi, ok := g.types.Uses[ident].(*types.Builtin)
	return ok && bi.Name() == "panic"
}

// emitPanicCall emits a panic() builtin call as so_panic(arg).
func (g *Generator) emitPanicCall(w io.Writer, call *ast.CallExpr) {
	arg := call.Args[0]
	typ := g.types.TypeOf(arg)
	if !g.hasStringType(arg) && !isErrorType(typ) {
		g.fail(call, "panic() only supports string and error arguments, got %s", g.typeString(typ))
	}
	fmt.Fprint(w, "so_panic(")
	g.emitCArg(w, arg)
	fmt.Fprint(w, ")")
}

// emitPrintCall emits a print/println call with an auto-generated format string.
func (g *Generator) emitPrintCall(w io.Writer, call *ast.CallExpr, name string) {
	if len(call.Args) == 0 {
		fmt.Fprintf(w, "so_%s(\"\")", name)
		return
	}
	format := g.buildFormatString(call)
	fmt.Fprintf(w, "so_%s(%s", name, format)
	for _, arg := range call.Args {
		fmt.Fprint(w, ", ")
		if g.hasStringType(arg) {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				// String literal: emit as a simple C string.
				fmt.Fprint(w, g.cStringLit(lit))
			} else {
				// String expression: emit expr.len, expr.ptr
				g.emitPostfixOperand(w, arg)
				fmt.Fprint(w, ".len, ")
				g.emitPostfixOperand(w, arg)
				fmt.Fprint(w, ".ptr")
			}
		} else {
			g.emitCArg(w, arg)
		}
	}
	fmt.Fprint(w, ")")
}

// emitOffsetof emits an unsafe.Offsetof call as offsetof(type, field).
// Returns false if the call is not unsafe.Offsetof.
func (g *Generator) emitOffsetof(w io.Writer, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Offsetof" {
		return false
	}
	if obj, ok := g.types.Uses[sel.Sel].(*types.Builtin); !ok || obj.Pkg() != types.Unsafe {
		return false
	}
	field, ok := ast.Unparen(call.Args[0]).(*ast.SelectorExpr)
	if !ok {
		g.fail(call, "unsafe.Offsetof requires a struct field")
	}
	structType := g.types.TypeOf(field.X)
	if ptr, ok := structType.Underlying().(*types.Pointer); ok {
		structType = ptr.Elem()
	}
	fmt.Fprintf(w, "unsafe_Offsetof(%s, %s)",
		g.mapTypeName(call, structType), g.fieldNameOf(field.Sel))
	return true
}

// buildFormatString constructs a C format string for the given print/println call,
// using the types of the arguments. It breaks out of string literals when macros
// are needed (e.g. "Value: %" PRId64) to avoid issues with macro expansion.
func (g *Generator) buildFormatString(call *ast.CallExpr) string {
	var format strings.Builder
	inStr := false
	for i, arg := range call.Args {
		typ := g.types.TypeOf(arg)
		spec, macro := g.formatSpec(arg, typ)
		if spec == "" {
			g.fail(call, "unsupported type for print/println: %s", g.typeString(typ))
			panic("unreachable")
		}
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
func (g *Generator) formatSpec(arg ast.Expr, typ types.Type) (spec, macro string) {
	if _, ok := typ.(*types.Pointer); ok {
		return "%p", ""
	}
	if iface, ok := typ.Underlying().(*types.Interface); ok && iface.Empty() {
		return "%p", ""
	}
	if isErrorType(typ) {
		return "%s", ""
	}
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return "", ""
	}
	switch basic.Kind() {
	case types.Bool:
		return "%d", ""
	case types.Float32, types.Float64, types.UntypedFloat:
		return "%f", ""
	case types.Int:
		return "%", "PRIdINT"
	case types.UntypedInt:
		return "%", "PRId64"
	case types.Int8, types.Int16, types.Int32:
		return "%d", ""
	case types.Int64:
		return "%", "PRId64"
	case types.Uint8, types.Uint16, types.Uint32:
		return "%u", ""
	case types.Uint:
		return "%", "PRIuINT"
	case types.Uint64:
		return "%", "PRIu64"
	case types.Uintptr:
		return "%", "PRIuPTR"
	case types.String, types.UntypedString:
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return "%s", ""
		}
		return "%.*s", ""
	default:
		return "", ""
	}
}
