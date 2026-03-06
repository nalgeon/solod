package clang

import (
	"fmt"
	"go/ast"
	"go/types"
)

// returnType returns the C return type for a function signature.
// For multi-return (T, error), validates the pattern and returns "so_Result".
// For single return, maps the Go type to C. For no return, returns "void".
func (g *Generator) returnType(node ast.Node, sig *types.Signature) string {
	if sig.Results().Len() > 1 {
		g.resultField(node, sig)
		return "so_Result"
	}
	if sig.Results().Len() == 1 {
		ret := sig.Results().At(0).Type()
		if _, ok := ret.Underlying().(*types.Array); ok {
			g.fail(node, "returning arrays from functions is not supported")
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

// resultField validates the (T, error) pattern and returns the union field name
// for the first return type (e.g. "as_int" for int).
func (g *Generator) resultField(node ast.Node, sig *types.Signature) string {
	if sig.Results().Len() != 2 {
		g.fail(node, "multi-return must be (T, error)")
	}
	// Verify second return is error.
	second := sig.Results().At(1).Type()
	if !isErrorType(second) {
		g.fail(node, "multi-return second value must be error, got %s", second)
	}
	// Map first return type to union field.
	first := sig.Results().At(0).Type()
	return resultFieldName(g, node, first)
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
	case types.Int, types.Int64, types.UntypedInt:
		return "as_int"
	case types.Int32, types.UntypedRune:
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

// emitMultiReturnDefine emits a multi-return define: x, err := f()
// Produces:
//
//	so_Result _res1 = f();
//	so_int x = _res1.val.as_int;
//	so_Error err = _res1.err;
func (g *Generator) emitMultiReturnDefine(stmt *ast.AssignStmt, call *ast.CallExpr) {
	w := g.state.writer
	sig := g.callSig(call)
	field := g.resultField(stmt, sig)

	// Emit temp variable with result of the call.
	g.state.tempCount++
	tmp := fmt.Sprintf("_res%d", g.state.tempCount)
	fmt.Fprintf(w, "%sso_Result %s = ", g.indent(), tmp)
	g.emitExpr(call)
	fmt.Fprintf(w, ";\n")

	// Emit individual variable declarations from result fields.
	for i, lhs := range stmt.Lhs {
		ident := lhs.(*ast.Ident)
		if ident.Name == "_" {
			continue
		}
		accessor := fmt.Sprintf("%s.val.%s", tmp, field)
		if i == 1 {
			accessor = fmt.Sprintf("%s.err", tmp)
		}
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

// emitMultiReturnAssign emits a multi-return assign: x, err = f()
// Produces:
//
//	so_Result _res1 = f();
//	x = _res1.val.as_int;
//	err = _res1.err;
func (g *Generator) emitMultiReturnAssign(stmt *ast.AssignStmt, call *ast.CallExpr) {
	w := g.state.writer
	sig := g.callSig(call)
	field := g.resultField(stmt, sig)

	// Emit temp variable with result of the call.
	g.state.tempCount++
	tmp := fmt.Sprintf("_res%d", g.state.tempCount)
	fmt.Fprintf(w, "%sso_Result %s = ", g.indent(), tmp)
	g.emitExpr(call)
	fmt.Fprintf(w, ";\n")

	// Emit assignments from result fields.
	for i, lhs := range stmt.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}
		accessor := fmt.Sprintf("%s.val.%s", tmp, field)
		if i == 1 {
			accessor = fmt.Sprintf("%s.err", tmp)
		}
		fmt.Fprintf(w, "%s", g.indent())
		g.emitExpr(lhs)
		fmt.Fprintf(w, " = %s;\n", accessor)
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
