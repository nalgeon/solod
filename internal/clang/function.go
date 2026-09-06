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
			exported = ast.IsExported(g.recvTypeName(decl.Recv.List[0]))
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
	if g.isGenericFunc(decl) {
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
	if g.isGenericFunc(decl) {
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
		_, typeParams := g.parseRecv(recv)
		params = append(params, typeParams...)
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
	// A generic function call emits as a macro, and
	// [Generator.emitGenericCall] writes its arguments.
	ext, isExtern := g.funcExtern(call)
	g.emitCallArgs(g.callArgs(w, call, false), call, sig, ext, isExtern)
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
	if isEmptyInterface(argType) {
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
	isSlice := isSliceType(g.types.TypeOf(arg))
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

// isGenericFunc reports whether a function declaration is generic
// (has type params on the function itself or on its receiver type).
func (g *Generator) isGenericFunc(decl *ast.FuncDecl) bool {
	if decl.Type.TypeParams != nil && len(decl.Type.TypeParams.List) > 0 {
		return true
	}
	if decl.Recv == nil {
		return false
	}
	_, typeParams := g.parseRecv(decl.Recv.List[0])
	return len(typeParams) > 0
}

// recvTypeName returns the Go type name from a method receiver field.
// Handles both pointer receivers (*Type) and value receivers (Type).
func (g *Generator) recvTypeName(recv *ast.Field) string {
	name, _ := g.parseRecv(recv)
	return name.Name
}

// recvTypeObj returns the types.Object for the receiver type of a method.
func (g *Generator) recvTypeObj(recv *ast.Field) types.Object {
	name, _ := g.parseRecv(recv)
	obj := g.types.Uses[name]
	// Resolve type aliases to the underlying named type.
	if named, ok := types.Unalias(obj.Type()).(*types.Named); ok {
		return named.Obj()
	}
	return obj
}

// parseRecv splits a method receiver into its type name and the names of its
// type parameters. A receiver is `Type`, `*Type`, `Type[T]` or `*Type[K, V]`.
func (g *Generator) parseRecv(recv *ast.Field) (name *ast.Ident, typeParams []string) {
	typ := recv.Type
	// Unwrap a pointer receiver.
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	// Unwrap the type parameters of a generic receiver.
	var indices []ast.Expr
	switch t := typ.(type) {
	case *ast.IndexExpr:
		typ, indices = t.X, []ast.Expr{t.Index}
	case *ast.IndexListExpr:
		typ, indices = t.X, t.Indices
	}
	name, ok := typ.(*ast.Ident)
	if !ok {
		g.fail(recv, "unsupported receiver type: %T", recv.Type)
	}
	// Go allows only a type parameter name as an index of a receiver type.
	for _, index := range indices {
		typeParams = append(typeParams, index.(*ast.Ident).Name)
	}
	return name, typeParams
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
