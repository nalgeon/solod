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

	// Parameters: methods prepend void* self.
	var parts []string
	if decl.Recv != nil {
		parts = append(parts, "void* self")
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

	name := g.symbolName(spec.Name.Name)
	fmt.Fprintf(w, "typedef %s (*%s)(%s);\n", retType, name, strings.Join(params, ", "))
}

// emitFuncDecl emits a function declaration.
func (g *Generator) emitFuncDecl(decl *ast.FuncDecl) {
	if decl.Body == nil || g.externs[externFuncKey(decl)] {
		return
	}
	if decl.Recv != nil {
		g.emitMethodDecl(decl)
		return
	}
	w := g.state.writer
	sig := g.funcSig(decl)
	g.rejectNamedReturns(decl, sig)
	g.state.funcSig = sig
	g.state.tempCount = 0
	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	g.emitFuncProto(w, decl)
	fmt.Fprintln(w, " {")
	g.state.indent++
	g.walkStmts(decl.Body.List)
	g.state.indent--
	fmt.Fprintf(w, "}\n")
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

	if g.isExternCall(call) {
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
		if isInterfaceType(sig.Params().At(i).Type()) &&
			!isInterfaceType(g.types.TypeOf(call.Args[i])) {
			paramType := sig.Params().At(i).Type()
			if iface, ok := paramType.Underlying().(*types.Interface); ok && iface.Empty() {
				g.emitAnyValue(call, call.Args[i])
			} else {
				g.emitInterfaceLit(paramType, call.Args[i])
			}
		} else {
			g.emitExpr(call.Args[i])
		}
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
	for i, arg := range variadicArgs {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExpr(arg)
	}
	fmt.Fprintf(w, "}, %d, %d}", count, count)
}

// emitCArgs emits arguments for an extern C function call.
func (g *Generator) emitCArgs(call *ast.CallExpr) {
	w := g.state.writer
	for i, arg := range call.Args {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitCArg(arg)
	}
}

// emitCArg emits an expression decayed to its C-compatible type:
// string literals to raw C strings, strings to char*, slices to void*.
func (g *Generator) emitCArg(arg ast.Expr) {
	w := g.state.writer
	if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		fmt.Fprintf(w, "%s", rawStringValue(lit))
	} else if basic, ok := g.types.TypeOf(arg).Underlying().(*types.Basic); ok && basic.Kind() == types.String {
		g.emitExpr(arg)
		fmt.Fprintf(w, ".ptr")
	} else if _, ok := g.types.TypeOf(arg).Underlying().(*types.Slice); ok {
		g.emitExpr(arg)
		fmt.Fprintf(w, ".ptr")
	} else {
		g.emitExpr(arg)
	}
}

// funcSig returns the types.Signature for a function or method declaration.
func (g *Generator) funcSig(decl *ast.FuncDecl) *types.Signature {
	if decl.Recv != nil {
		return g.types.ObjectOf(decl.Name).Type().(*types.Signature)
	}
	return g.types.Defs[decl.Name].Type().(*types.Signature)
}

// isExternCall reports whether a call expression targets an extern C function.
func (g *Generator) isExternCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return g.externs[fun.Name]
	case *ast.SelectorExpr:
		// Package-qualified call (e.g. stdio.Printf).
		if ident, ok := fun.X.(*ast.Ident); ok {
			if _, ok := g.types.Uses[ident].(*types.PkgName); ok {
				return g.externs[fun.Sel.Name]
			}
		}
	}
	return false
}

// recvTypeName returns the Go type name from a method receiver field.
// Handles both pointer receivers (*Rect) and value receivers (Rect).
func recvTypeName(recv *ast.Field) string {
	switch t := recv.Type.(type) {
	case *ast.StarExpr:
		return t.X.(*ast.Ident).Name
	case *ast.Ident:
		return t.Name
	}
	panic(fmt.Sprintf("unsupported receiver type: %T", recv.Type))
}
