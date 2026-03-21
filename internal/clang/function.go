package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// emitFuncProto writes a full C function prototype (e.g. "static void main_foo(int x)")
// without a terminator. Returns the function's type signature for callers that need it.
func (g *Generator) emitFuncProto(w io.Writer, decl *ast.FuncDecl) *types.Signature {
	// Specifier: static for unexported, empty for exported and main.
	spec := ""
	if decl.Name.Name != "main" {
		exported := ast.IsExported(decl.Name.Name)
		if exported && decl.Recv != nil {
			exported = ast.IsExported(recvTypeName(decl.Recv.List[0]))
		}
		if !exported {
			spec = "static "
		}
	}

	sig := g.funcSig(decl)

	// Return type.
	retType := "void"
	if decl.Name.Name == "main" {
		retType = "int"
	} else if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		retType = g.returnType(decl, sig)
	}

	// Name: methods use RecvType_Method, functions use symbolName.
	name := g.symbolName(decl.Name.Name)
	if decl.Recv != nil {
		name = g.symbolName(recvTypeName(decl.Recv.List[0])) + "_" + decl.Name.Name
	}

	// Parameters: methods prepend receiver
	// (void* self for pointer, T name for value).
	var parts []string
	if decl.Recv != nil {
		recv := decl.Recv.List[0]
		if _, ok := recv.Type.(*ast.Ident); ok {
			// Value receiver: pass struct by value.
			cStructType := g.symbolName(recvTypeName(recv))
			recvName := "self"
			if len(recv.Names) > 0 {
				recvName = recv.Names[0].Name
			}
			parts = append(parts, cStructType+" "+recvName)
		} else {
			parts = append(parts, "void* self")
		}
	}
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			typ := g.types.TypeOf(field.Type)
			ct := g.mapCType(decl, typ)
			for _, n := range field.Names {
				parts = append(parts, ct.Decl(n.Name))
			}
		}
	}
	params := "void"
	if len(parts) > 0 {
		params = strings.Join(parts, ", ")
	}

	fmt.Fprintf(w, "%s%s %s(%s)", spec, retType, name, params)
	return sig
}

// emitFuncTypeSpec emits a C function pointer typedef.
func (g *Generator) emitFuncTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	named := g.types.Defs[spec.Name].Type().(*types.Named)
	sig := named.Underlying().(*types.Signature)

	retType := g.returnType(spec, sig)

	var params []string
	for parVar := range sig.Params().Variables() {
		params = append(params, g.mapType(spec, parVar.Type()))
	}

	name := g.declSymbolName(spec.Name.Name)
	fmt.Fprintf(w, "%stypedef %s (*%s)(%s);\n", g.indent(), retType, name, strings.Join(params, ", "))
}

// emitFuncDecl emits a function declaration.
func (g *Generator) emitFuncDecl(decl *ast.FuncDecl) {
	if decl.Body == nil || g.hasExtern("", externFuncKey(decl)) {
		return
	}
	if decl.Recv != nil {
		g.emitMethodDecl(decl)
		return
	}

	// Init emission state.
	w := g.state.writer
	sig := g.funcSig(decl)
	g.rejectNamedReturns(decl, sig)
	g.state.funcSig = sig
	g.state.tempCount = 0

	// Emit comments and function prototype.
	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	g.emitFuncProto(w, decl)
	fmt.Fprintln(w, " {")

	// Emit function body, handling deferred calls if needed.
	g.state.indent++
	g.walkStmts(decl.Body.List)
	if !endsWithReturn(decl.Body.List) {
		g.emitDeferredCalls()
	}
	g.state.indent--
	fmt.Fprintf(w, "}\n")

	// Reset state.
	g.state.defers = nil
	g.state.funcSig = nil
}

// emitFuncCall emits a regular function call.
func (g *Generator) emitFuncCall(call *ast.CallExpr) {
	w := g.state.writer
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if bi, ok := g.types.Uses[ident].(*types.Builtin); ok {
			if g.emitBuiltin(call, ident, bi) {
				return
			}
		} else {
			g.emitExpr(call.Fun)
		}
	} else {
		g.emitExpr(call.Fun)
	}

	// Emit arguments, wrapping as interfaces if needed.
	var sig *types.Signature
	if funType := g.types.TypeOf(call.Fun); funType != nil {
		// Get the function signature to wrap value arguments as interfaces if needed.
		sig, _ = funType.Underlying().(*types.Signature)
	}
	fmt.Fprintf(w, "(")

	if ext, ok := g.callExtern(call); ok && !ext.nodecay {
		// Extern C call: decay all args to C-compatible types.
		// So wrapper types (so_String, so_Slice) must be unwrapped to their
		// underlying C representations for C function macros.
		g.emitCArgs(call)
	} else if sig != nil && sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitFixedArgs(call, sig)
		g.emitVariadicArgs(call, sig)
	} else {
		// Regular call: emit all args as-is.
		for i, arg := range call.Args {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			if sig != nil && i < sig.Params().Len() {
				// Emit arg, wrapping as interface if needed based on parameter type.
				g.emitExprAsType(call, arg, sig.Params().At(i).Type())
			} else {
				// No signature available (e.g. func literal), emit arg as-is.
				g.emitExpr(arg)
			}
		}
	}

	fmt.Fprintf(w, ")")
}

// emitFixedArgs emits the non-variadic arguments for a variadic call.
func (g *Generator) emitFixedArgs(call *ast.CallExpr, sig *types.Signature) {
	w := g.state.writer
	fixedCount := sig.Params().Len() - 1
	for i := 0; i < fixedCount && i < len(call.Args); i++ {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExprAsType(call, call.Args[i], sig.Params().At(i).Type())
	}
}

// emitVariadicArgs packs trailing arguments into an inline so_Slice literal.
func (g *Generator) emitVariadicArgs(call *ast.CallExpr, sig *types.Signature) {
	w := g.state.writer
	fixedCount := sig.Params().Len() - 1
	variadicArgs := call.Args[fixedCount:]

	if fixedCount > 0 {
		fmt.Fprintf(w, ", ")
	}

	variadicParam := sig.Params().At(sig.Params().Len() - 1)
	elemType := g.mapType(call, variadicParam.Type().(*types.Slice).Elem())
	count := len(variadicArgs)

	fmt.Fprintf(w, "(so_Slice){(%s[%d]){", elemType, count)
	targetType := variadicParam.Type().(*types.Slice).Elem()
	for i, arg := range variadicArgs {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExprAsType(call, arg, targetType)
	}
	fmt.Fprintf(w, "}, %d, %d}", count, count)
}

// emitCArgs emits arguments for an extern C function call.
func (g *Generator) emitCArgs(call *ast.CallExpr) {
	w := g.state.writer
	var sig *types.Signature
	if funType := g.types.TypeOf(call.Fun); funType != nil {
		sig, _ = funType.Underlying().(*types.Signature)
	}
	for i, arg := range call.Args {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		// Interface-typed parameters (e.g. Allocator) need emitExprAsType
		// to convert nil to a zero-initialized struct instead of NULL.
		if sig != nil && i < sig.Params().Len() && isNamedNonEmptyInterface(sig.Params().At(i).Type()) {
			g.emitExprAsType(call, arg, sig.Params().At(i).Type())
		} else {
			g.emitCArg(arg)
		}
	}
}

// emitCArg emits an expression decayed to its C-compatible type:
// string literals to raw C strings, strings to char*, slices to void*.
func (g *Generator) emitCArg(arg ast.Expr) {
	w := g.state.writer
	if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		fmt.Fprintf(w, "%s", rawStringValue(lit))
	} else if g.hasStringType(arg) {
		fmt.Fprintf(w, "so_cstr(")
		g.emitExpr(arg)
		fmt.Fprintf(w, ")")
	} else if _, ok := g.types.TypeOf(arg).Underlying().(*types.Slice); ok {
		g.emitExpr(arg)
		fmt.Fprintf(w, ".ptr")
	} else if isErrorType(g.types.TypeOf(arg)) {
		g.emitExpr(arg)
		fmt.Fprintf(w, "->msg")
	} else {
		g.emitExpr(arg)
	}
}

// hasUnexportedTypes reports whether a function declaration
// references any unexported types from the current package.
func (g *Generator) hasUnexportedTypes(decl *ast.FuncDecl) bool {
	sig := g.funcSig(decl)
	for p := range sig.Params().Variables() {
		if g.isUnexportedType(p.Type()) {
			return true
		}
	}
	for r := range sig.Results().Variables() {
		if g.isUnexportedType(r.Type()) {
			return true
		}
	}
	return false
}

// funcSig returns the types.Signature for a function or method declaration.
func (g *Generator) funcSig(decl *ast.FuncDecl) *types.Signature {
	if decl.Recv != nil {
		return g.types.ObjectOf(decl.Name).Type().(*types.Signature)
	}
	return g.types.Defs[decl.Name].Type().(*types.Signature)
}

// endsWithReturn reports whether a statement list ends with a return statement.
func endsWithReturn(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	_, ok := stmts[len(stmts)-1].(*ast.ReturnStmt)
	return ok
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
