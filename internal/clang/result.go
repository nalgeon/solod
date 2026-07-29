package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// resultTypeInfo describes an auto-generated result struct for (T, error) returns.
type resultTypeInfo struct {
	cName    string // e.g. "main_FileResult"
	valType  string // e.g. "main_File"
	exported bool   // T's export status
}

// collectResultTypes collects result types for (T, error) returns.
func (g *Generator) collectResultTypes() {
	seen := make(map[string]bool)
	for _, sym := range g.symbols {
		if sym.kind != symbolFunc && sym.kind != symbolMethod {
			continue
		}
		sig := g.types.ObjectOf(sym.funcDecl.Name).Type().(*types.Signature)
		if sig.Results().Len() != 2 {
			continue
		}
		first := sig.Results().At(0).Type()
		second := sig.Results().At(1).Type()
		if !isErrorType(second) {
			continue
		}
		named, ok := types.Unalias(first).(*types.Named)
		if !ok {
			continue
		}
		if _, ok := named.Underlying().(*types.Struct); !ok {
			continue
		}
		// Only handle types from the current package.
		if named.Obj().Pkg() != g.pkg.Types {
			continue
		}
		valType := g.mapTypeName(sym.funcDecl, named)
		cName := valType + "Result"
		if seen[cName] {
			continue
		}
		seen[cName] = true
		g.resultTypes = append(g.resultTypes, resultTypeInfo{
			cName:    cName,
			valType:  valType,
			exported: ast.IsExported(named.Obj().Name()),
		})
	}
}

// emitResultTypes writes auto-generated result type structs filtered by export status.
func (g *Generator) emitResultTypes(w io.Writer, exported bool) {
	var resTypes []resultTypeInfo
	for _, typ := range g.resultTypes {
		if typ.exported == exported {
			resTypes = append(resTypes, typ)
		}
	}
	if len(resTypes) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// -- Result types --")
	for _, typ := range resTypes {
		const s = "\ntypedef struct %s {\n    %s val;\n    so_Error err;\n} %s;\n"
		fmt.Fprintf(w, s, typ.cName, typ.valType, typ.cName)
	}
}

// returnType returns the C return type for a function signature.
// For multi-return (T, error) or (T, T), returns the per-signature result type.
// For single return, maps the Go type to C. For no return, returns "void".
func (g *Generator) returnType(node ast.Node, sig *types.Signature) string {
	if sig.Results().Len() > 1 {
		info := g.multiReturnFields(node, sig)
		return info.typeName()
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

// emitMultiReturnDefine emits a multi-return define: x, y := f()
// Produces:
//
//	so_R_int_err _res1 = f();
//	so_int x = _res1.val;
//	so_Error y = _res1.err;           // (T, error)
//	so_int y = _res1.val2;            // (T, T)
func (g *Generator) emitMultiReturnDefine(w io.Writer, stmt *ast.AssignStmt, call *ast.CallExpr) {
	sig := g.callSig(call)
	multi := g.multiReturnFields(stmt, sig)

	// Emit temp variable with result of the call.
	tmp := g.newTemp(stmt, tempResult)
	fmt.Fprintf(w, "%s%s %s = ", g.indent(), multi.typeName(), tmp)
	g.emitExpr(w, call)
	fmt.Fprint(w, ";\n")

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
		cType := g.mapTypeName(stmt, def.Type())
		fmt.Fprintf(w, "%s%s %s = %s;\n", g.indent(), cType, ident.Name, accessor)
	}
}

// emitMultiReturnAssign emits a multi-return assign: x, y = f()
// Produces:
//
//	so_R_int_err _res1 = f();
//	x = _res1.val;
//	y = _res1.err;                    // (T, error)
//	y = _res1.val2;                   // (T, T)
func (g *Generator) emitMultiReturnAssign(w io.Writer, stmt *ast.AssignStmt, call *ast.CallExpr) {
	sig := g.callSig(call)
	multi := g.multiReturnFields(stmt, sig)

	// Emit temp variable with result of the call.
	tmp := g.newTemp(stmt, tempResult)
	fmt.Fprintf(w, "%s%s %s = ", g.indent(), multi.typeName(), tmp)
	g.emitExpr(w, call)
	fmt.Fprint(w, ";\n")

	// Emit assignments from result fields.
	for i, lhs := range stmt.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}
		accessor := multi.accessor(tmp, i)
		fmt.Fprint(w, g.indent())
		g.emitExpr(w, lhs)
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

	// Check for named struct result type: (T, error).
	if isErrorType(second) {
		if named, ok := types.Unalias(first).(*types.Named); ok {
			if _, ok := named.Underlying().(*types.Struct); ok {
				resultType := g.mapTypeName(node, named) + "Result"
				return multiReturn{resultType: resultType, hasError: true}
			}
		}
	}

	s1 := resultTypeSuffix(g, node, first)
	if isErrorType(second) {
		return multiReturn{suffix1: s1, hasError: true}
	}
	s2 := resultTypeSuffix(g, node, second)
	return multiReturn{suffix1: s1, suffix2: s2}
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

// resultTypeSuffix maps a Go type to the corresponding result type suffix.
func resultTypeSuffix(g *Generator, node ast.Node, typ types.Type) string {
	typ = types.Unalias(typ)
	switch t := typ.(type) {
	case *types.Array:
		g.fail(node, "arrays in multi-return are not supported")
	case *types.Slice:
		return "slice"
	case *types.Pointer:
		return "ptr"
	case *types.Interface:
		if t.Empty() {
			return "ptr"
		}
	}
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported multi-return type: %s", g.typeString(typ))
	}
	switch basic.Kind() {
	case types.Bool, types.UntypedBool:
		return "bool"
	case types.Float32:
		return "f32"
	case types.Float64, types.UntypedFloat:
		return "f64"
	case types.Int, types.UntypedInt:
		return "int"
	case types.Int8:
		return "i8"
	case types.Int16:
		return "i16"
	case types.Int32:
		if basic.Name() == "rune" {
			return "rune"
		}
		return "i32"
	case types.Int64:
		return "i64"
	case types.Uint:
		return "uint"
	case types.Uint16:
		return "u16"
	case types.Uint32:
		return "u32"
	case types.Uint64:
		return "u64"
	case types.Uint8:
		return "byte"
	case types.UntypedRune:
		return "rune"
	case types.String, types.UntypedString:
		return "str"
	default:
		g.fail(node, "unsupported multi-return type: %s", g.typeString(typ))
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
	suffix1    string // type suffix for first value (e.g. "int", "str")
	suffix2    string // type suffix for second value (e.g. "int", "bool"), empty if hasError
	hasError   bool   // true when second return is error
	resultType string // C type name when using custom result struct (e.g. "main_FileResult")
}

// typeName returns the C type name for this multi-return.
func (mr multiReturn) typeName() string {
	if mr.resultType != "" {
		return mr.resultType
	}
	if mr.hasError {
		return "so_R_" + mr.suffix1 + "_err"
	}
	return "so_R_" + mr.suffix1 + "_" + mr.suffix2
}

// accessor returns the C accessor for position i of a multi-return.
// Position 0 -> tmp.val
// Position 1 -> tmp.err (T, error) or tmp.val2 (T, T)
func (mr multiReturn) accessor(tmp string, i int) string {
	if mr.resultType != "" {
		if i == 0 {
			return tmp + ".val"
		}
		return tmp + ".err"
	}
	if i == 0 {
		return tmp + ".val"
	}
	if mr.hasError {
		return tmp + ".err"
	}
	return tmp + ".val2"
}
