package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"
)

// FuncDecl represents a function declaration.
type FuncDecl struct {
	gen *Generator

	decl *ast.FuncDecl
	spec string // specifier (e.g. "static")
	typ  *ast.FuncType
}

func newFuncDecl(gen *Generator, decl *ast.FuncDecl) FuncDecl {
	spec := ""
	if decl.Name.Name != "main" && !ast.IsExported(decl.Name.Name) {
		spec = "static "
	}
	return FuncDecl{
		gen:  gen,
		decl: decl,
		spec: spec,
		typ:  decl.Type,
	}
}

// name returns the C function name.
// For methods, this is structType_methodName (e.g. main_Rect_Area).
// For regular functions, this is the symbolName (e.g. main_RectArea).
func (f *FuncDecl) name() string {
	if f.decl.Recv != nil {
		recv := f.decl.Recv.List[0]
		return f.gen.symbolName(recvTypeName(recv)) + "_" + f.decl.Name.Name
	}
	return f.gen.symbolName(f.decl.Name.Name)
}

// params returns the C parameter list.
// For methods, prepends void* self.
// For functions with multiple returns, appends out-parameters.
func (f *FuncDecl) params() string {
	var parts []string

	// Prepend self parameter for methods.
	if f.decl.Recv != nil {
		parts = append(parts, "void* self")
	}

	// Append regular parameters.
	if f.typ.Params != nil {
		for _, field := range f.typ.Params.List {
			typ := f.gen.types.TypeOf(field.Type)
			cType := f.gen.mapType(f.decl, typ)
			for _, name := range field.Names {
				parts = append(parts, cType+" "+name.Name)
			}
		}
	}

	// Append out-parameters for multiple return values.
	if outNames := f.outParams(); len(outNames) > 0 {
		outIdx := 0
		first := true
		for _, field := range f.typ.Results.List {
			typ := f.gen.types.TypeOf(field.Type)
			cType := f.gen.mapType(f.decl, typ)
			count := len(field.Names)
			if count == 0 {
				count = 1
			}
			for j := 0; j < count; j++ {
				if first {
					first = false
					continue
				}
				parts = append(parts, cType+"* "+outNames[outIdx])
				outIdx++
			}
		}
	}

	if len(parts) == 0 {
		return "void"
	}
	return strings.Join(parts, ", ")
}

// outParams returns the names for out-parameters (all return values beyond the first).
// Named returns use their Go name; unnamed returns use _r1, _r2, etc.
func (f *FuncDecl) outParams() []string {
	if f.typ.Results == nil {
		return nil
	}
	// Flatten all result values.
	type nameInfo struct{ name string }
	var all []nameInfo
	for _, field := range f.typ.Results.List {
		if len(field.Names) > 0 {
			for _, n := range field.Names {
				all = append(all, nameInfo{n.Name})
			}
		} else {
			all = append(all, nameInfo{})
		}
	}

	if len(all) <= 1 {
		return nil
	}
	// Assign names for out-parameters, using
	// the Go name if available or _rN if unnamed.
	var names []string
	for i, info := range all[1:] {
		if info.name != "" {
			names = append(names, info.name)
		} else {
			names = append(names, fmt.Sprintf("_r%d", i+1))
		}
	}
	return names
}

// returnType returns the C return type.
func (f *FuncDecl) returnType() string {
	if f.decl.Name.Name == "main" {
		return "int"
	}
	if f.typ.Results == nil || len(f.typ.Results.List) == 0 {
		return "void"
	}
	typ := f.gen.types.TypeOf(f.typ.Results.List[0].Type)
	return f.gen.mapType(f.decl, typ)
}

// emitFuncTypeSpec emits a C function pointer typedef.
func (g *Generator) emitFuncTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	named := g.types.Defs[spec.Name].Type().(*types.Named)
	sig := named.Underlying().(*types.Signature)

	retType := "void"
	if sig.Results().Len() > 0 {
		retType = g.mapType(spec, sig.Results().At(0).Type())
	}

	var params []string
	for parVar := range sig.Params().Variables() {
		params = append(params, g.mapType(spec, parVar.Type()))
	}

	name := g.symbolName(spec.Name.Name)
	fmt.Fprintf(w, "\ntypedef %s (*%s)(%s);\n", retType, name, strings.Join(params, ", "))
}

// emitFuncDecl emits a function declaration.
func (g *Generator) emitFuncDecl(decl *ast.FuncDecl) {
	if decl.Body == nil {
		// Functions with no body are considered externs and ignored.
		return
	}
	if decl.Recv != nil {
		g.emitMethodDecl(decl)
		return
	}
	w := g.state.writer
	fn := newFuncDecl(g, decl)
	fmt.Fprintf(w, "\n%s%s %s(%s) {\n", fn.spec, fn.returnType(), fn.name(), fn.params())
	g.state.outParams = fn.outParams()
	g.emitBlock(decl.Body)
	g.state.outParams = nil
	fmt.Fprintf(w, "}\n")
}

// emitFuncCall emits a regular function call.
func (g *Generator) emitFuncCall(call *ast.CallExpr) {
	w := g.state.writer
	if ident, ok := call.Fun.(*ast.Ident); ok {
		// Simple function call (e.g. println("hello")).
		if bi, ok := g.types.Uses[ident].(*types.Builtin); ok {
			switch bi.Name() {
			case "append":
				g.emitAppendCall(call)
				return
			case "copy":
				g.emitCopyCall(call)
				return
			case "make":
				g.emitMakeCall(call)
				return
			case "panic":
				arg, ok := call.Args[0].(*ast.BasicLit)
				if !ok {
					g.fail(call, "panic() only supports string literals")
				}
				fmt.Fprintf(w, "so_panic(%s)", arg.Value)
				return
			case "print", "println":
				g.emitPrintCall(call, bi.Name())
				return
			}
			// Other builtins are emitted as regular calls
			// with a so_ prefix (e.g. so_len(slice)).
			fmt.Fprintf(w, "so_%s", ident.Name)
		} else {
			// Regular function call.
			g.emitExpr(call.Fun)
		}
	} else {
		// Complex function expression (e.g. selector or func literal).
		g.emitExpr(call.Fun)
	}

	// Emit arguments, wrapping as interfaces if needed.
	var sig *types.Signature
	if funType := g.types.TypeOf(call.Fun); funType != nil {
		// Get the function signature to wrap value arguments as interfaces if needed.
		sig, _ = funType.Underlying().(*types.Signature)
	}
	fmt.Fprintf(w, "(")

	if sig != nil && sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitFixedArgs(call, sig)
		g.emitVariadicArgs(call, sig)
	} else {
		// Regular call: emit all args as-is.
		for i, arg := range call.Args {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			if sig != nil && i < sig.Params().Len() &&
				isInterfaceType(sig.Params().At(i).Type()) &&
				!isInterfaceType(g.types.TypeOf(arg)) {
				// Argument needs to be wrapped as an interface.
				paramType := sig.Params().At(i).Type()
				if iface, ok := paramType.Underlying().(*types.Interface); ok && iface.Empty() {
					g.emitAnyValue(call, arg)
				} else {
					g.emitInterfaceLit(paramType, arg)
				}
			} else {
				// Regular argument.
				g.emitExpr(arg)
			}
		}
	}

	// Append out-parameter arguments for multi-return calls.
	for i, addr := range g.state.outArgs {
		if i > 0 || len(call.Args) > 0 {
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "%s", addr)
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
