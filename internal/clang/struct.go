package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
)

// emitStructTypeSpec emits a typedef struct for a struct type declaration.
// dirs provides parsed so: directives for package-level declarations.
func (g *Generator) emitStructTypeSpec(w io.Writer, spec *ast.TypeSpec, dirs directives) {
	st := spec.Type.(*ast.StructType)
	cName := g.declSymbolName(g.types.Defs[spec.Name])
	attr := dirs.attrString()
	if attr != "" {
		fmt.Fprintf(w, "%stypedef struct %s %s {\n", g.indent(), attr, cName)
	} else {
		fmt.Fprintf(w, "%stypedef struct %s {\n", g.indent(), cName)
	}
	g.state.indent++
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		for _, name := range field.Names {
			if innerSt, ok := field.Type.(*ast.StructType); ok {
				g.emitInlineStructField(w, innerSt, name.Name)
			} else if sig, ok := typ.(*types.Signature); ok {
				g.emitFuncPtrField(w, spec, name.Name, sig, cName)
			} else {
				// Regular struct field (arrays get dimension suffix).
				ct := g.mapTypeDecl(field, typ)
				fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(name.Name))
			}
		}
	}
	g.state.indent--
	fmt.Fprintf(w, "%s} %s;\n", g.indent(), cName)
}

// emitFuncPtrField emits a function pointer field in a struct typedef.
// Example: so_int (*ratingFn)(struct main_Movie);
//
// Parameter names are omitted, because they are not required in a C function
// pointer declaration. Emitting them could cause a conflict with a C keyword.
func (g *Generator) emitFuncPtrField(w io.Writer, node ast.Node, fieldName string, sig *types.Signature, enclosingStruct string) {
	retType := g.returnType(node, sig)
	var params []string
	for p := range sig.Params().Variables() {
		cType := g.mapTypeName(node, p.Type())
		if cType == enclosingStruct || cType == enclosingStruct+"*" {
			cType = "struct " + cType
		}
		params = append(params, cType)
	}
	fmt.Fprintf(w, "%s%s (*%s)(%s);\n", g.indent(), retType, fieldName, funcParams(params))
}

// emitInlineStructField emits an anonymous struct field inline within a parent struct.
// Example: struct { so_int n; so_int i; } loop;
// Does not support function pointer fields within the inline struct.
func (g *Generator) emitInlineStructField(w io.Writer, st *ast.StructType, fieldName string) {
	fmt.Fprintf(w, "%sstruct {\n", g.indent())
	g.state.indent++
	for _, f := range st.Fields.List {
		typ := g.types.TypeOf(f.Type)
		ct := g.mapTypeDecl(f, typ)
		for _, name := range f.Names {
			fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(name.Name))
		}
	}
	g.state.indent--
	fmt.Fprintf(w, "%s} %s;\n", g.indent(), fieldName)
}

// emitMethodDecl emits a method as a C function.
// Pointer receivers use void* self with a cast; value receivers pass the struct by value.
func (g *Generator) emitMethodDecl(w io.Writer, decl *ast.FuncDecl) {
	sig := g.funcSig(decl)
	g.rejectNamedReturns(decl, sig)

	// Init emission state.
	recv := decl.Recv.List[0]
	cStructType := g.symbolName(g.recvTypeObj(recv))
	named := len(recv.Names) > 0
	_, isValueRecv := recv.Type.(*ast.Ident)

	g.state.funcSig = sig
	g.state.tempCount = 0

	// Emit comments and function prototype.
	if !g.emitComments(w, decl) {
		fmt.Fprintln(w)
	}
	g.emitFuncProto(w, decl)
	fmt.Fprintln(w, " {")
	g.state.indent++

	// Emit receiver preamble.
	if isValueRecv {
		// Value receivers are passed by value - no cast needed.
		// Unnamed value receivers need (void)self to suppress unused warnings.
		if !named {
			fmt.Fprintf(w, "%s(void)self;\n", g.indent())
		}
	} else {
		// Pointer receivers: cast void* self to the concrete type.
		if named {
			recvName := recv.Names[0].Name
			fmt.Fprintf(w, "%s%s* %s = self;\n", g.indent(), cStructType, recvName)
		} else {
			fmt.Fprintf(w, "%s(void)self;\n", g.indent())
		}
	}

	// Emit method body, handling deferred calls if needed.
	g.walkStmts(w, decl.Body.List)
	if !endsWithReturn(decl.Body.List) {
		g.emitDeferredCalls(w)
	}
	g.state.indent--
	fmt.Fprint(w, "}\n")

	// Reset state.
	g.state.defers = nil
	g.state.funcSig = nil
}

// emitAnonStructLit emits an anonymous struct literal.
// (e.g. struct{ x, y int }{1, 2} or struct{ x, y int }{ x: 1, y: 2 })
func (g *Generator) emitAnonStructLit(w io.Writer, n *ast.CompositeLit, st *ast.StructType) {
	// Struct fields declaration.
	fmt.Fprint(w, "(struct {\n")
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		cType := g.mapTypeName(field, typ)
		for _, name := range field.Names {
			fmt.Fprintf(w, "%s    %s %s;\n", g.indent(), cType, name.Name)
		}
	}
	fmt.Fprintf(w, "%s})", g.indent())

	// Struct fields initialization.
	fmt.Fprint(w, "{\n")
	fields := collectFieldNames(st)
	struc := g.types.TypeOf(n).Underlying().(*types.Struct)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ",\n")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fieldName := kv.Key.(*ast.Ident).Name
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), fieldName)
			g.emitFieldValue(w, n, kv.Value, structFieldType(struc, fieldName))
		} else {
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), fields[i])
			g.emitFieldValue(w, n, elt, struc.Field(i).Type())
		}
	}
	fmt.Fprint(w, ",\n")
	fmt.Fprintf(w, "%s}", g.indent())
}

// emitStructLit emits a struct literal (e.g. Point{1, 2} or Point{x: 1, y: 2}).
func (g *Generator) emitStructLit(w io.Writer, n *ast.CompositeLit) {
	var typ types.Type
	if n.Type != nil {
		typ = g.types.TypeOf(n.Type)
	} else {
		typ = g.types.TypeOf(n)
	}
	cType := g.mapTypeName(n, typ)
	fmt.Fprintf(w, "(%s)", cType)
	g.emitBareStructInit(w, n)
}

// emitBareStructInit emits a struct literal as a bare initializer
// (e.g. {.n = 200, .i = 10}) without a compound literal cast prefix.
func (g *Generator) emitBareStructInit(w io.Writer, n *ast.CompositeLit) {
	struc := g.types.TypeOf(n).Underlying().(*types.Struct)
	fmt.Fprint(w, "{")
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fieldName := kv.Key.(*ast.Ident).Name
			fmt.Fprintf(w, ".%s = ", fieldName)
			if lit, ok := isAnonStructLit(kv.Value); ok {
				g.emitBareStructInit(w, lit)
			} else {
				g.emitFieldValue(w, n, kv.Value, structFieldType(struc, fieldName))
			}
		} else {
			if lit, ok := isAnonStructLit(elt); ok {
				g.emitBareStructInit(w, lit)
			} else {
				g.emitFieldValue(w, n, elt, struc.Field(i).Type())
			}
		}
	}
	fmt.Fprint(w, "}")
}

// emitFieldValue emits a struct literal field value.
// Unlike a plain assignment, an array field can only take an array literal.
func (g *Generator) emitFieldValue(w io.Writer, n ast.Node, expr ast.Expr, fieldType types.Type) {
	g.checkArrayValue(expr, fieldType)
	g.emitExprAsType(w, n, expr, fieldType)
}

// emitMethodCall emits a method call.
func (g *Generator) emitMethodCall(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr) {
	selection := g.types.Selections[sel]
	recv := selection.Recv()
	sig := selection.Type().(*types.Signature)

	// Get the struct type name.
	var named *types.Named
	if ptr, ok := recv.(*types.Pointer); ok {
		named = types.Unalias(ptr.Elem()).(*types.Named)
	} else {
		named = types.Unalias(recv).(*types.Named)
	}

	// Interface method dispatch: s.Perim(2) → s.Perim(s.self, 2)
	if isInterfaceType(named) {
		g.emitExpr(w, sel.X)
		fmt.Fprintf(w, ".%s(", sel.Sel.Name)
		g.emitExpr(w, sel.X)
		fmt.Fprint(w, ".self")
		g.emitMethodCallArgs(w, sel, call, sig, "", "")
		fmt.Fprint(w, ")")
		return
	}

	// Regular method call: r.Area() → main_Rect_Area(&r)
	cStructType := g.mapTypeName(sel, named)
	cName := cStructType + "_" + sel.Sel.Name
	fmt.Fprintf(w, "%s(", cName)

	typeArgs := named.TypeArgs()
	isGeneric := typeArgs.Len() > 0
	// Prepend type arguments for generic method calls.
	if isGeneric {
		for i := 0; i < typeArgs.Len(); i++ {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, g.mapTypeName(sel, typeArgs.At(i)))
		}
		fmt.Fprint(w, ", ")
	}

	// Pass receiver based on method's declared receiver type and call-site type.
	declSig := selection.Obj().Type().(*types.Signature)
	_, isMethodPtrRecv := declSig.Recv().Type().(*types.Pointer)
	xType := g.types.TypeOf(sel.X)
	_, isCallSitePtr := xType.Underlying().(*types.Pointer)

	// For generic (= macro) calls, wrap non-type args in parens to protect
	// against the preprocessor misinterpreting commas.
	lparen, rparen := "", ""
	if isGeneric {
		lparen, rparen = "(", ")"
	}

	fmt.Fprint(w, lparen)
	if isMethodPtrRecv {
		// Pointer receiver: pass address of value, or pointer directly.
		if isCallSitePtr {
			g.emitExpr(w, sel.X)
		} else {
			fmt.Fprint(w, "&")
			g.emitExpr(w, sel.X)
		}
	} else {
		// Value receiver: pass value directly, or dereference pointer.
		if isCallSitePtr {
			fmt.Fprint(w, "*")
			g.emitExpr(w, sel.X)
		} else {
			g.emitExpr(w, sel.X)
		}
	}
	fmt.Fprint(w, rparen)

	// Pass method arguments.
	g.emitMethodCallArgs(w, sel, call, sig, lparen, rparen)
	fmt.Fprint(w, ")")
}

// emitMethodCallArgs emits method arguments, handling variadic arg packing.
func (g *Generator) emitMethodCallArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature, lparen, rparen string) {
	args := call.Args

	if ext, ok := g.methodExtern(sel); ok && !ext.nodecay {
		// Extern C method: decay args to C-compatible types.
		g.emitMethodExternArgs(w, sel, call, sig)
		return
	}

	if sig.Variadic() && !call.Ellipsis.IsValid() {
		// Variadic call with individual args: pack trailing args into a slice literal.
		g.emitMethodVarArgs(w, sel, call, sig, lparen, rparen)
		return
	}

	// Non-variadic call or variadic call with ellipsis: emit all args directly.
	for i, arg := range args {
		fmt.Fprintf(w, ", %s", lparen)
		g.emitCallArg(w, sel, arg, sig.Params().At(i).Type())
		fmt.Fprint(w, rparen)
	}
}

// emitMethodExternArgs emits method arguments for an extern C method, handling type decay.
func (g *Generator) emitMethodExternArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature) {
	if call.Ellipsis.IsValid() {
		g.fail(call, "spreading variadic arguments to an extern function is not supported")
	}
	for i, arg := range call.Args {
		fmt.Fprint(w, ", ")
		if i < sig.Params().Len() && isNamedNonEmptyInterface(sig.Params().At(i).Type()) {
			g.emitExprAsType(w, sel, arg, sig.Params().At(i).Type())
		} else {
			g.emitCArg(w, arg)
		}
	}
}

func (g *Generator) emitMethodVarArgs(w io.Writer, sel *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature, lparen, rparen string) {
	// Emit fixed args first.
	fixedCount := sig.Params().Len() - 1
	args := call.Args
	for i := 0; i < fixedCount && i < len(args); i++ {
		fmt.Fprintf(w, ", %s", lparen)
		g.emitCallArg(w, sel, args[i], sig.Params().At(i).Type())
		fmt.Fprint(w, rparen)
	}

	// Emit variadic args as a so_Slice literal.
	variadicArgs := args[fixedCount:]
	variadicParam := sig.Params().At(sig.Params().Len() - 1)
	elemType := g.mapTypeName(sel, variadicParam.Type().(*types.Slice).Elem())
	count := len(variadicArgs)
	targetType := variadicParam.Type().(*types.Slice).Elem()
	if count == 0 {
		// No variadic args: emit a nil slice.
		fmt.Fprintf(w, ", %s(so_Slice){0}%s", lparen, rparen)
		return
	}

	fmt.Fprintf(w, ", %s(so_Slice){(%s[%d]){", lparen, elemType, count)
	for i, arg := range variadicArgs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExprAsType(w, sel, arg, targetType)
	}
	fmt.Fprintf(w, "}, %d, %d}%s", count, count, rparen)
}

// structFieldType returns the type of a struct field by name.
func structFieldType(st *types.Struct, name string) types.Type {
	for field := range st.Fields() {
		if field.Name() == name {
			return field.Type()
		}
	}
	panic("structFieldType: field not found: " + name)
}

// collectFieldNames returns the field names from a struct type in order.
func collectFieldNames(st *ast.StructType) []string {
	var names []string
	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

// isAnonStructLit checks if an expression is an anonymous struct composite literal.
func isAnonStructLit(expr ast.Expr) (*ast.CompositeLit, bool) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}
	_, ok = lit.Type.(*ast.StructType)
	return lit, ok
}
