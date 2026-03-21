package clang

import (
	"fmt"
	"go/ast"
	"go/types"
)

// returnType returns the C return type for a function signature.
// For multi-return (T, error) or (T, T), validates the pattern and returns "so_Result".
// For single return, maps the Go type to C. For no return, returns "void".
func (g *Generator) returnType(node ast.Node, sig *types.Signature) string {
	if sig.Results().Len() > 1 {
		info := g.multiReturnFields(node, sig)
		if info.resultType != "" {
			return info.resultType
		}
		return "so_Result"
	}
	if sig.Results().Len() == 1 {
		ret := sig.Results().At(0).Type()
		if _, ok := ret.Underlying().(*types.Array); ok {
			g.fail(node, "returning arrays from functions is not supported")
		}
		if _, ok := ret.Underlying().(*types.Map); ok {
			g.fail(node, "returning maps from functions is not supported")
		}
		if ptr, ok := ret.Underlying().(*types.Pointer); ok {
			if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
				g.fail(node, "returning pointer-to-array from functions is not supported")
			}
		}
		return g.mapType(node, ret)
	}
	return "void"
}

// emitMultiReturnDefine emits a multi-return define: x, y := f()
// Produces:
//
//	so_Result _res1 = f();
//	so_int x = _res1.val.as_int;
//	so_Error y = _res1.err;           // (T, error)
//	so_int y = _res1.val2.as_int;     // (T, T)
func (g *Generator) emitMultiReturnDefine(stmt *ast.AssignStmt, call *ast.CallExpr) {
	w := g.state.writer
	sig := g.callSig(call)
	multi := g.multiReturnFields(stmt, sig)

	// Emit temp variable with result of the call.
	g.state.tempCount++
	tmp := fmt.Sprintf("_res%d", g.state.tempCount)
	fmt.Fprintf(w, "%s%s %s = ", g.indent(), multi.typeName(), tmp)
	g.emitExpr(call)
	fmt.Fprintf(w, ";\n")

	// Emit individual variable declarations from result fields.
	for i, lhs := range stmt.Lhs {
		ident := lhs.(*ast.Ident)
		if ident.Name == "_" {
			continue
		}
		accessor := multi.accessor(tmp, i)
		def := g.types.Defs[ident]
		if def == nil {
			// Redeclared variable - plain assignment.
			fmt.Fprintf(w, "%s%s = %s;\n", g.indent(), ident.Name, accessor)
			continue
		}
		cType := g.mapType(stmt, def.Type())
		fmt.Fprintf(w, "%s%s %s = %s;\n", g.indent(), cType, ident.Name, accessor)
	}
}

// emitMultiReturnAssign emits a multi-return assign: x, y = f()
// Produces:
//
//	so_Result _res1 = f();
//	x = _res1.val.as_int;
//	y = _res1.err;                    // (T, error)
//	y = _res1.val2.as_int;            // (T, T)
func (g *Generator) emitMultiReturnAssign(stmt *ast.AssignStmt, call *ast.CallExpr) {
	w := g.state.writer
	sig := g.callSig(call)
	multi := g.multiReturnFields(stmt, sig)

	// Emit temp variable with result of the call.
	g.state.tempCount++
	tmp := fmt.Sprintf("_res%d", g.state.tempCount)
	fmt.Fprintf(w, "%s%s %s = ", g.indent(), multi.typeName(), tmp)
	g.emitExpr(call)
	fmt.Fprintf(w, ";\n")

	// Emit assignments from result fields.
	for i, lhs := range stmt.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}
		accessor := multi.accessor(tmp, i)
		fmt.Fprintf(w, "%s", g.indent())
		g.emitExpr(lhs)
		fmt.Fprintf(w, " = %s;\n", accessor)
	}
}

// multiReturnFields validates a multi-return signature and returns info
// about both positions. The second type is either error or a supported type.
func (g *Generator) multiReturnFields(node ast.Node, sig *types.Signature) multiReturn {
	if sig.Results().Len() != 2 {
		g.fail(node, "multi-return must have exactly 2 values")
	}
	first := sig.Results().At(0).Type()
	second := sig.Results().At(1).Type()
	if isErrorType(first) {
		g.fail(node, "error must be the second return value")
	}

	// Check for custom result type: (NamedStruct, error).
	if isErrorType(second) {
		if named, ok := types.Unalias(first).(*types.Named); ok {
			if _, isStruct := named.Underlying().(*types.Struct); isStruct {
				resultType := g.findResultType(node, named)
				return multiReturn{resultType: resultType, hasError: true}
			}
		}
	}

	f1 := resultFieldName(g, node, first)
	if isErrorType(second) {
		return multiReturn{field1: f1, hasError: true}
	}
	f2 := resultFieldName(g, node, second)
	return multiReturn{field1: f1, field2: f2}
}

// findResultType looks up the {TypeName}Result type in the package scope.
func (g *Generator) findResultType(node ast.Node, named *types.Named) string {
	resultName := named.Obj().Name() + "Result"
	obj := named.Obj().Pkg().Scope().Lookup(resultName)
	if obj == nil {
		g.fail(node, "returning struct %s requires a %s type declaration", named.Obj().Name(), resultName)
	}
	return g.mapType(node, obj.Type())
}

// resultFieldName maps a Go type to the corresponding so_Result union field name.
func resultFieldName(g *Generator, node ast.Node, typ types.Type) string {
	typ = types.Unalias(typ)
	switch t := typ.(type) {
	case *types.Array:
		g.fail(node, "arrays in multi-return are not supported")
	case *types.Slice:
		return "as_slice"
	case *types.Pointer:
		return "as_ptr"
	case *types.Interface:
		if t.Empty() {
			return "as_ptr"
		}
	}
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported result type for so_Result: %s", typ)
	}
	switch basic.Kind() {
	case types.Bool, types.UntypedBool:
		return "as_bool"
	case types.Float32, types.Float64, types.UntypedFloat:
		return "as_double"
	case types.Int, types.UntypedInt:
		return "as_int"
	case types.Int32:
		if basic.Name() == "rune" {
			return "as_rune"
		}
		return "as_i32"
	case types.Int64:
		return "as_i64"
	case types.Uint:
		return "as_uint"
	case types.Uint32:
		return "as_u32"
	case types.Uint64:
		return "as_u64"
	case types.UntypedRune:
		return "as_rune"
	case types.String, types.UntypedString:
		return "as_string"
	case types.Uint8:
		return "as_byte"
	default:
		g.fail(node, "unsupported result type for so_Result: %s", typ)
		panic("unreachable")
	}
}

// rejectNamedReturns fails if any return value has a name.
func (g *Generator) rejectNamedReturns(node ast.Node, sig *types.Signature) {
	for v := range sig.Results().Variables() {
		if v.Name() != "" {
			g.fail(node, "named return values are not supported")
		}
	}
}

// callSig extracts the function signature from a call expression.
func (g *Generator) callSig(call *ast.CallExpr) *types.Signature {
	return g.types.TypeOf(call.Fun).Underlying().(*types.Signature)
}

// multiReturn describes a two-value return: (T, error) or (T, T).
type multiReturn struct {
	field1     string // union field for first value (e.g. "as_int")
	field2     string // union field for second value (e.g. "as_int"), empty if hasError
	hasError   bool   // true when second return is error
	resultType string // C type name when using custom result struct (e.g. "main_FileResult")
}

// typeName returns the C type name for this multi-return - either the custom
// result type or "so_Result".
func (mr multiReturn) typeName() string {
	if mr.resultType != "" {
		return mr.resultType
	}
	return "so_Result"
}

// accessor returns the C accessor for position i of a multi-return.
// Position 0 -> tmp.val.<field1>  (or tmp.val for custom result types)
// Position 1 -> tmp.err (if hasError) or tmp.val2.<field2>
func (mr multiReturn) accessor(tmp string, i int) string {
	if mr.resultType != "" {
		// Custom result type: access fields directly (e.g. tmp.val and tmp.err).
		if i == 0 {
			return fmt.Sprintf("%s.val", tmp)
		}
		return fmt.Sprintf("%s.err", tmp)
	}
	// Standard so_Result union: access via field1 and field2.
	if i == 0 {
		return fmt.Sprintf("%s.val.%s", tmp, mr.field1)
	}
	if mr.hasError {
		return fmt.Sprintf("%s.err", tmp)
	}
	return fmt.Sprintf("%s.val2.%s", tmp, mr.field2)
}
