package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strconv"
	"strings"
)

// emitFuncProto writes a full C function prototype (e.g. "static void main_foo(int x)")
// without a terminator. Returns the function's type signature for callers that need it.
func (g *Generator) emitFuncProto(w io.Writer, decl *ast.FuncDecl) *types.Signature {
	// Specifier: static inline for so:inline; static for unexported;
	// empty for exported, so:promote, and main.
	dirs := g.funcDirs[decl]
	spec := ""
	if dirs.inline {
		spec = "static inline "
	} else if decl.Name.Name != "main" {
		exported := ast.IsExported(decl.Name.Name)
		if exported && decl.Recv != nil {
			exported = ast.IsExported(recvTypeName(decl.Recv.List[0]))
		}
		if !exported && !dirs.promote {
			spec = "static "
		}
	}
	attr := dirs.attrString()
	if attr != "" {
		spec = spec + attr + " "
	}

	sig := g.funcSig(decl)

	// Return type.
	retType := "void"
	if isMainFunc(decl) {
		retType = "int"
	} else if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		retType = g.returnType(decl, sig)
	}

	name := g.mapObjName(g.types.Defs[decl.Name])

	// Parameters: methods prepend receiver
	// (void* self for pointer, T name for value).
	names := g.paramNames(decl)
	g.checkParamNames(decl, names)
	var parts []string
	if decl.Recv != nil {
		recv := decl.Recv.List[0]
		if _, ok := recv.Type.(*ast.Ident); ok {
			// Value receiver: pass struct by value.
			cStructType := g.mapObjName(g.recvTypeObj(recv))
			parts = append(parts, cStructType+" "+names[0])
		} else {
			parts = append(parts, "void* self")
		}
	}
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			typ := g.types.TypeOf(field.Type)
			ct := g.mapTypeDecl(field.Type, typ)
			if len(field.Names) == 0 {
				parts = append(parts, ct.Decl(names[len(parts)])+" so_unused")
				continue
			}
			for _, n := range field.Names {
				part := ct.Decl(names[len(parts)])
				if n.Name == "_" {
					part += " so_unused"
				}
				parts = append(parts, part)
			}
		}
	}
	params := "void"
	if isMainFunc(decl) && g.opts.InitArgs {
		params = "int argc, char* argv[]"
	} else if len(parts) > 0 {
		params = strings.Join(parts, ", ")
	}

	fmt.Fprintf(w, "%s%s %s(%s)", spec, retType, name, params)
	return sig
}

// emitFuncTypeSpec emits a C function pointer typedef.
func (g *Generator) emitFuncTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	// A type alias to a function type has no named type, so the signature
	// comes from the underlying type in both cases.
	typ := types.Unalias(g.types.Defs[spec.Name].Type())
	sig := typ.Underlying().(*types.Signature)

	retType := g.returnType(spec, sig)

	var params []string
	for parVar := range sig.Params().Variables() {
		params = append(params, g.mapParamType(spec, parVar.Type()))
	}

	name := g.mapObjName(g.types.Defs[spec.Name])
	fmt.Fprintf(w, "%stypedef %s (*%s)(%s);\n", g.indent(), retType, name, funcParams(params))
}

// emitFuncDecl emits a function declaration into the .c file.
// Inline functions are skipped here - they are emitted into the header
// by [Generator.emitInlineFuncDecl].
func (g *Generator) emitFuncDecl(w io.Writer, decl *ast.FuncDecl) {
	if decl.Body == nil || g.hasExtern(g.types.Defs[decl.Name]) {
		return
	}
	if isInitFunc(decl) {
		return
	}
	if g.funcDirs[decl].inline {
		return
	}
	if isGenericFunc(decl) {
		// Type parameters only exist as macro arguments. A regular function
		// can't handle them, yet call sites still pass them.
		kind := "function"
		if decl.Recv != nil {
			kind = "method"
		}
		g.fail(decl, "generic %s %s must be so:inline or so:extern", kind, decl.Name.Name)
	}
	g.emitFuncBody(w, decl)
}

// emitInlineFuncDecl emits a so:inline function declaration into the header.
// Generic functions are emitted as #define macros; non-generic as static inline.
func (g *Generator) emitInlineFuncDecl(w io.Writer, decl *ast.FuncDecl) {
	if isGenericFunc(decl) {
		g.emitMacroFuncDecl(w, decl)
		return
	}
	g.emitFuncBody(w, decl)
}

// emitMacroFuncDecl emits a generic so:inline function as a #define macro.
func (g *Generator) emitMacroFuncDecl(w io.Writer, decl *ast.FuncDecl) {
	sig := g.funcSig(decl)
	g.checkNamedReturns(decl, sig)

	name := g.mapObjName(g.types.Defs[decl.Name])

	// Build param list: type params, then receiver (for methods), then regular params.
	// Non-type params are suffixed with _ to avoid name collisions (b->val = val).
	// References are wrapped in parens to avoid syntax errors (&b->val).
	var params []string
	macroParams := make(map[string]bool)
	if decl.Type.TypeParams != nil {
		for _, field := range decl.Type.TypeParams.List {
			for _, n := range field.Names {
				params = append(params, n.Name)
			}
		}
	}
	if decl.Recv != nil {
		recv := decl.Recv.List[0]
		// Add receiver type params (no suffix - these are type names).
		params = append(params, recvTypeParams(recv)...)
		// Add receiver as parameter (suffixed).
		recvName, named := recvVarName(recv)
		if named {
			macroParams[recvName] = true
		}
		params = append(params, recvName+"_")
	}
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			if len(field.Names) == 0 {
				params = append(params, blankParamName(len(params))+"_")
				continue
			}
			for _, n := range field.Names {
				if n.Name == "_" {
					params = append(params, blankParamName(len(params))+"_")
					continue
				}
				macroParams[n.Name] = true
				params = append(params, n.Name+"_")
			}
		}
	}
	g.checkParamNames(decl, params)

	// Capture body output.
	var buf strings.Builder
	g.state.enterMacro(decl, sig, macroParams)
	g.walkStmts(&buf, decl.Body.List)
	g.state.leaveFunc()

	// Determine if returning or void.
	hasReturn := sig.Results() != nil && sig.Results().Len() > 0

	// Emit #define with line continuations.
	body := buf.String()
	// Trim trailing newline.
	body = strings.TrimRight(body, "\n")
	lines := strings.Split(body, "\n")

	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	if hasReturn {
		fmt.Fprintf(w, "#define %s(%s) ({", name, strings.Join(params, ", "))
	} else {
		fmt.Fprintf(w, "#define %s(%s) do {", name, strings.Join(params, ", "))
	}
	for _, line := range lines {
		fmt.Fprintf(w, " \\\n%s", line)
	}
	if hasReturn {
		fmt.Fprintln(w, " \\")
		fmt.Fprintln(w, "})")
	} else {
		fmt.Fprintln(w, " \\")
		fmt.Fprintln(w, "} while (0)")
	}
}

// emitFuncBody emits a function or method body. Shared by [Generator.emitFuncDecl]
// and [Generator.emitInlineFuncDecl].
func (g *Generator) emitFuncBody(w io.Writer, decl *ast.FuncDecl) {
	if decl.Recv != nil {
		g.emitMethodDecl(w, decl)
		return
	}

	// Init emission state.
	sig := g.funcSig(decl)
	g.checkNamedReturns(decl, sig)
	g.state.enterFunc(decl, sig)
	defer g.state.leaveFunc()

	// Emit comments and function prototype.
	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	g.emitFuncProto(w, decl)
	fmt.Fprintln(w, " {")

	// Emit function body, handling deferred calls if needed.
	g.state.depth++
	if isMainFunc(decl) && g.opts.InitArgs {
		fmt.Fprintf(w, "%sso_String _so_argv[argc];\n", g.indent())
		fmt.Fprintf(w, "%sso_args_init(argc, argv, _so_argv);\n", g.indent())
	}
	g.walkStmts(w, decl.Body.List)
	if !endsWithReturn(decl.Body.List) {
		g.emitDeferredCalls(w)
		if isMainFunc(decl) {
			fmt.Fprintf(w, "%sreturn 0;\n", g.indent())
		}
	}
	g.state.depth--
	fmt.Fprint(w, "}\n")
}

// emitFuncCall emits a regular function call.
func (g *Generator) emitFuncCall(w io.Writer, call *ast.CallExpr) {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if bi, ok := g.types.Uses[ident].(*types.Builtin); ok {
			if g.emitBuiltin(w, call, ident, bi) {
				return
			}
		} else {
			g.checkLocalCall(call)
			g.emitExpr(w, call.Fun)
		}
	} else {
		g.checkLocalCall(call)
		g.emitExpr(w, call.Fun)
	}

	fmt.Fprint(w, "(")
	g.emitFuncCallArgs(w, call)
	fmt.Fprint(w, ")")
}

// emitFuncCallArgs emits the arguments for a function call.
func (g *Generator) emitFuncCallArgs(w io.Writer, call *ast.CallExpr) {
	var sig *types.Signature
	if funType := g.types.TypeOf(call.Fun); funType != nil {
		// Get the function signature to wrap value arguments as interfaces if needed.
		sig, _ = funType.Underlying().(*types.Signature)
	}

	if ext, ok := g.funcExtern(call); ok {
		if !ext.nodecay {
			// Extern C function: decay args to C-compatible types.
			g.emitFuncExternArgs(w, call)
			return
		}
		if sig != nil && sig.Variadic() {
			// Extern nodecay function: emit the variadic args flat.
			g.emitFuncExternVarArgs(w, call, sig)
			return
		}
	}

	if sig != nil && sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitFuncVarArgs(w, call, sig)
		return
	}

	// Non-variadic call or variadic call with ellipsis: emit all args directly.
	for i, arg := range call.Args {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		if sig != nil && i < sig.Params().Len() {
			g.emitCallArg(w, call, arg, sig.Params().At(i).Type())
		} else {
			// No signature available (e.g. func literal), emit arg as-is.
			g.emitExpr(w, arg)
		}
	}
}

// emitFuncVarArgs packs trailing arguments into an inline so_Slice literal.
func (g *Generator) emitFuncVarArgs(w io.Writer, call *ast.CallExpr, sig *types.Signature) {
	// Emit fixed args first.
	fixedCount := sig.Params().Len() - 1
	for i := 0; i < fixedCount && i < len(call.Args); i++ {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitCallArg(w, call, call.Args[i], sig.Params().At(i).Type())
	}

	// Emit variadic args as a so_Slice literal.
	variadicArgs := call.Args[fixedCount:]
	if fixedCount > 0 {
		fmt.Fprint(w, ", ")
	}

	variadicParam := sig.Params().At(sig.Params().Len() - 1)
	elemType := g.mapTypeName(call, variadicParam.Type().(*types.Slice).Elem())
	count := len(variadicArgs)

	if count == 0 {
		// No variadic args: emit a nil slice.
		fmt.Fprint(w, "(so_Slice){}")
		return
	}

	fmt.Fprintf(w, "(so_Slice){(%s[%d]){", elemType, count)
	targetType := variadicParam.Type().(*types.Slice).Elem()
	for i, arg := range variadicArgs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExprAsType(w, call, arg, targetType)
	}
	fmt.Fprintf(w, "}, %d, %d}", count, count)
}

// emitFuncExternArgs emits arguments for an extern C function call, handling type decay.
func (g *Generator) emitFuncExternArgs(w io.Writer, call *ast.CallExpr) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}

	var sig *types.Signature
	if funType := g.types.TypeOf(call.Fun); funType != nil {
		sig, _ = funType.Underlying().(*types.Signature)
	}

	for i, arg := range call.Args {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		// Interface-typed parameters (e.g. Allocator) need emitExprAsType
		// to convert nil to a zero-initialized struct instead of NULL.
		if sig != nil && i < sig.Params().Len() && isNamedNonEmptyInterface(sig.Params().At(i).Type()) {
			g.emitExprAsType(w, call, arg, sig.Params().At(i).Type())
		} else {
			g.emitCArg(w, arg)
		}
	}
}

// emitFuncExternVarArgs emits the arguments of a call to a variadic extern
// nodecay function.
func (g *Generator) emitFuncExternVarArgs(w io.Writer, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i := range call.Args {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExternVarArg(w, call, call, sig, i)
	}
}

// emitExternVarArg emits argument i of a call to a variadic extern nodecay
// function. A fixed argument takes the declared parameter type. A variadic
// argument goes flat, at its own type, because the callee reads one value per
// va_arg call. A so_Slice literal cannot cross a C variadic.
func (g *Generator) emitExternVarArg(w io.Writer, node ast.Node, call *ast.CallExpr, sig *types.Signature, i int) {
	arg := call.Args[i]
	if i < sig.Params().Len()-1 {
		g.emitCallArg(w, node, arg, sig.Params().At(i).Type())
		return
	}
	argType := g.types.TypeOf(arg)
	if iface, ok := argType.Underlying().(*types.Interface); ok && iface.Empty() {
		g.fail(arg, "cannot pass an any value to a variadic extern function")
	}
	// The generator widens the scalar types to avoid C promotion issues.
	if cType, ok := varArgType(argType); ok {
		fmt.Fprintf(w, "(%s)", cType)
		g.emitParenExpr(w, arg)
		return
	}
	g.emitExprAsType(w, node, arg, argType)
}

// emitCallArg emits a call argument coerced to the parameter type.
func (g *Generator) emitCallArg(w io.Writer, node ast.Node, arg ast.Expr, paramType types.Type) {
	if arr, ok := arrayType(paramType); ok {
		g.emitArrayValue(w, node, arg, arr)
		return
	}
	g.emitExprAsType(w, node, arg, paramType)
}

// emitCArg emits an expression decayed to its C-compatible type:
// string literals to raw C strings, strings to char*, slices to void*.
func (g *Generator) emitCArg(w io.Writer, arg ast.Expr) {
	if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		fmt.Fprint(w, g.cStringLit(lit))
		return
	}
	_, isSlice := g.types.TypeOf(arg).Underlying().(*types.Slice)
	var macro string
	switch {
	case g.hasStringType(arg):
		macro = "so_cstr"
	case isSlice:
		macro = "so_decay"
	case isErrorType(g.types.TypeOf(arg)):
		macro = "so_error_cstr"
	default:
		g.emitExpr(w, arg)
		return
	}
	g.checkLocalScope(arg)
	fmt.Fprintf(w, "%s(", macro)
	g.emitMacroArg(w, arg)
	fmt.Fprint(w, ")")
}

// funcSig extracts the function signature from a function or method declaration.
func (g *Generator) funcSig(decl *ast.FuncDecl) *types.Signature {
	if decl.Recv != nil {
		return g.types.ObjectOf(decl.Name).Type().(*types.Signature)
	}
	return g.types.Defs[decl.Name].Type().(*types.Signature)
}

// callSig extracts the function signature from a call expression.
func (g *Generator) callSig(call *ast.CallExpr) *types.Signature {
	return g.types.TypeOf(call.Fun).Underlying().(*types.Signature)
}

// paramNames returns the C name of every parameter of a non-generic function,
// in order: the receiver first, then the declared parameters. A blank ("_")
// parameter has no Go name, so it gets a generated name.
func (g *Generator) paramNames(decl *ast.FuncDecl) []string {
	var names []string
	if decl.Recv != nil {
		recv := decl.Recv.List[0]
		if _, ok := recv.Type.(*ast.Ident); ok {
			name, _ := recvVarName(recv)
			names = append(names, name)
		} else {
			names = append(names, "self")
		}
	}
	if decl.Type.Params == nil {
		return names
	}
	for _, field := range decl.Type.Params.List {
		if len(field.Names) == 0 {
			names = append(names, blankParamName(len(names)))
			continue
		}
		for _, n := range field.Names {
			if n.Name == "_" {
				names = append(names, blankParamName(len(names)))
				continue
			}
			names = append(names, n.Name)
		}
	}
	return names
}

// checkParamNames rejects a function with two parameters of the same C name.
func (g *Generator) checkParamNames(decl *ast.FuncDecl, names []string) {
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			g.fail(decl, "duplicate C parameter name %q in func %s", name, decl.Name.Name)
		}
		seen[name] = true
	}
}

// recvTypeName returns the Go type name from a method receiver field.
// Handles both pointer receivers (*Rect) and value receivers (Rect).
func recvTypeName(recv *ast.Field) string {
	typ := recv.Type
	// Unwrap pointer receiver.
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	// Unwrap generic type parameters.
	switch t := typ.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return t.X.(*ast.Ident).Name
	case *ast.IndexListExpr:
		return t.X.(*ast.Ident).Name
	}
	panic(fmt.Sprintf("unsupported receiver type: %T", recv.Type))
}

// recvTypeObj returns the types.Object for the receiver type of a method.
func (g *Generator) recvTypeObj(recv *ast.Field) types.Object {
	typ := recv.Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	var obj types.Object
	switch t := typ.(type) {
	case *ast.Ident:
		obj = g.types.Uses[t]
	case *ast.IndexExpr:
		obj = g.types.Uses[t.X.(*ast.Ident)]
	case *ast.IndexListExpr:
		obj = g.types.Uses[t.X.(*ast.Ident)]
	default:
		g.fail(recv, "unsupported receiver type: %T", recv.Type)
	}
	// Resolve type aliases to the underlying named type.
	if named, ok := types.Unalias(obj.Type()).(*types.Named); ok {
		return named.Obj()
	}
	return obj
}

// recvVarName returns the C name of the method receiver. It also reports
// whether the method body can use the name. A method with no receiver name or
// with a blank one gets the name "self", which the body never uses.
func recvVarName(recv *ast.Field) (name string, named bool) {
	if len(recv.Names) == 0 || recv.Names[0].Name == "_" {
		return "self", false
	}
	return recv.Names[0].Name, true
}

// recvTypeParams extracts type parameter names from a generic receiver field.
func recvTypeParams(recv *ast.Field) []string {
	typ := recv.Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	switch t := typ.(type) {
	case *ast.IndexExpr:
		if ident, ok := t.Index.(*ast.Ident); ok {
			return []string{ident.Name}
		}
	case *ast.IndexListExpr:
		var names []string
		for _, idx := range t.Indices {
			if ident, ok := idx.(*ast.Ident); ok {
				names = append(names, ident.Name)
			}
		}
		return names
	}
	return nil
}

// varArgType returns the C type a scalar widens to in the variadic position of
// an extern nodecay call. The second result reports whether the type widens.
func varArgType(typ types.Type) (string, bool) {
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return "", false
	}
	switch info := basic.Info(); {
	case info&types.IsBoolean != 0:
		return "so_int", true
	case info&types.IsUnsigned != 0:
		return "so_uint", true
	case info&types.IsInteger != 0:
		return "so_int", true
	case info&types.IsFloat != 0:
		return "double", true
	}
	return "", false
}

// isGenericFunc reports whether a function declaration is generic
// (has type params on the function itself or on its receiver type).
func isGenericFunc(decl *ast.FuncDecl) bool {
	if decl.Type.TypeParams != nil && len(decl.Type.TypeParams.List) > 0 {
		return true
	}
	if decl.Recv != nil {
		recv := decl.Recv.List[0]
		typ := recv.Type
		if star, ok := typ.(*ast.StarExpr); ok {
			typ = star.X
		}
		switch typ.(type) {
		case *ast.IndexExpr, *ast.IndexListExpr:
			return true
		}
	}
	return false
}

// isMainFunc reports whether a function declaration is the main function.
func isMainFunc(decl *ast.FuncDecl) bool {
	return decl.Name.Name == "main" && decl.Recv == nil
}

// isInitFunc reports whether a function declaration is the init function.
func isInitFunc(decl *ast.FuncDecl) bool {
	return decl.Name.Name == "init" && decl.Recv == nil
}

// funcParams formats a C function pointer parameter list.
func funcParams(params []string) string {
	// An empty list becomes "void", because C reads "()" as
	// unspecified parameters rather than none.
	if len(params) == 0 {
		return "void"
	}
	return strings.Join(params, ", ")
}

// blankParamName returns the C name of the blank ("_") parameter at
// position i. C needs a distinct name for every parameter.
func blankParamName(i int) string {
	return "_" + strconv.Itoa(i)
}

// endsWithReturn reports whether a statement list ends with a return statement.
func endsWithReturn(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	_, ok := stmts[len(stmts)-1].(*ast.ReturnStmt)
	return ok
}
