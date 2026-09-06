package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// emitInterfaceTypeSpec emits a typedef struct with void* self and function pointers,
// followed by a dispatch function for every method.
func (g *Generator) emitInterfaceTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	if spec.TypeParams != nil {
		g.fail(spec, "generic interfaces are not supported")
	}
	if !g.state.atTopLevel() {
		// The dispatch functions go next to the typedef, and C does not
		// allow a function definition inside a function body.
		g.fail(spec, "interface declaration inside a function is not supported")
	}
	typ := types.Unalias(g.types.Defs[spec.Name].Type()).(*types.Named)
	iface := typ.Underlying().(*types.Interface)
	cName := g.mapObjName(g.types.Defs[spec.Name])
	fmt.Fprintf(w, "typedef struct %s {\n", cName)
	fmt.Fprint(w, "    void* self;\n")
	for m := range iface.Methods() {
		sig := m.Type().(*types.Signature)
		retType := g.returnType(spec, sig)
		// Parameter names are omitted, because they are not required in a C function
		// pointer declaration. Emitting them could cause a conflict with a C keyword.
		var params strings.Builder
		params.WriteString("void* self")
		for p := range sig.Params().Variables() {
			params.WriteString(", ")
			params.WriteString(g.mapParamType(spec, p.Type()))
		}
		fmt.Fprintf(w, "    %s (*%s)(%s);\n", retType, m.Name(), params.String())
	}
	fmt.Fprintf(w, "} %s;\n", cName)
	g.emitInterfaceMethods(w, spec, iface, cName)
}

// emitInterfaceMethods emits a dispatch function for every interface method.
func (g *Generator) emitInterfaceMethods(w io.Writer, spec *ast.TypeSpec, iface *types.Interface, cName string) {
	unused := ""
	if !ast.IsExported(spec.Name.Name) {
		// The typedef of an unexported interface goes to the .c file,
		// where the C compiler warns about an uncalled static function.
		unused = "so_unused "
	}

	astParams := interfaceMethodParams(spec)
	for m := range iface.Methods() {
		sig := m.Type().(*types.Signature)
		retType := g.returnType(spec, sig)

		params := []string{cName + " self"}
		args := []string{"self.self"}
		names := interfaceParamNames(astParams[m.Name()])
		if len(names) != sig.Params().Len() {
			// A mismatch means this method isn't declared in the interface,
			// like with an embedded interface, which is rejected elsewhere.
			g.fail(spec, "the declaration of method %s has %d parameters, but its type has %d",
				m.Name(), len(names), sig.Params().Len())
		}
		for i, name := range names {
			ct := g.mapTypeDecl(spec, sig.Params().At(i).Type())
			params = append(params, ct.Decl(name))
			args = append(args, name)
		}

		ret := "return "
		if retType == "void" {
			ret = ""
		}
		fmt.Fprintf(w, "\nstatic inline %s%s %s_%s(%s) {\n", unused, retType, cName, m.Name(), strings.Join(params, ", "))
		fmt.Fprintf(w, "    %sself.%s(%s);\n", ret, m.Name(), strings.Join(args, ", "))
		fmt.Fprint(w, "}\n")
	}
}

// interfaceMethodParams returns the parameter list of every method
// in an interface declaration, by method name.
func interfaceMethodParams(spec *ast.TypeSpec) map[string]*ast.FieldList {
	params := map[string]*ast.FieldList{}
	itype, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return params
	}
	for _, method := range itype.Methods.List {
		ftype, ok := method.Type.(*ast.FuncType)
		if !ok || len(method.Names) != 1 {
			// Not a method. An embedded interface is rejected elsewhere.
			continue
		}
		params[method.Names[0].Name] = ftype.Params
	}
	return params
}

// interfaceParamNames returns the C parameter names of an interface method.
// The declaration can omit a parameter name or use the blank identifier;
// such a parameter is named after its position (_0, _1, ...).
func interfaceParamNames(params *ast.FieldList) []string {
	if params == nil {
		return nil
	}
	var names []string
	for _, param := range params.List {
		if len(param.Names) == 0 {
			names = append(names, blankParamName(len(names)))
			continue
		}
		for _, n := range param.Names {
			if n.Name == "_" {
				names = append(names, blankParamName(len(names)))
				continue
			}
			names = append(names, n.Name)
		}
	}
	return names
}

// emitInterfaceLit emits a compound literal that wraps a concrete value as an interface.
// Example: (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim}
func (g *Generator) emitInterfaceLit(w io.Writer, ifaceType types.Type, expr ast.Expr) {
	named := types.Unalias(ifaceType).(*types.Named)
	iface := named.Underlying().(*types.Interface)

	// Get value type, dereferencing if it's a pointer.
	concreteType, isPtr := ptrElem(g.types.TypeOf(expr))
	concreteNamed := concreteType.(*types.Named)
	g.checkMethodReceivers(expr, iface, concreteNamed)

	cIface := g.mapTypeName(expr, named)
	cConcrete := g.mapTypeName(expr, concreteNamed)

	if !isPtr {
		// Unreachable: checkMethodReceivers rejects value receivers, and Go
		// rejects a value whose methods have pointer receivers.
		g.fail(expr, "cannot convert value of type %s to an interface", g.typeString(concreteNamed))
	}
	fmt.Fprintf(w, "(%s){.self = ", cIface)
	g.emitExpr(w, expr)
	for m := range iface.Methods() {
		fmt.Fprintf(w, ", .%s = %s_%s", m.Name(), cConcrete, m.Name())
	}
	fmt.Fprint(w, "}")
}

// emitTypeAssertion emits a comma-ok type assertion (e.g. _, ok := s.(Rect)).
// Uses function pointer comparison to identify the concrete type.
func (g *Generator) emitTypeAssertion(w io.Writer, stmt *ast.AssignStmt, ta *ast.TypeAssertExpr) {
	sourceType := g.types.TypeOf(ta.X)
	if isEmptyInterface(sourceType) {
		g.fail(ta, "comma-ok type assertion on any is not supported")
	}
	ifaceType := types.Unalias(sourceType).(*types.Named)
	iface := ifaceType.Underlying().(*types.Interface)
	firstMethod := iface.Method(0).Name()

	// Get value type, dereferencing if it's a pointer.
	assertedType, _ := ptrElem(g.types.TypeOf(ta.Type))
	concreteNamed := assertedType.(*types.Named)
	cConcrete := g.mapTypeName(ta, concreteNamed)

	okIdent := stmt.Lhs[1].(*ast.Ident)
	if stmt.Tok == token.DEFINE {
		fmt.Fprintf(w, "%sbool %s = (", g.indent(), okIdent.Name)
	} else {
		fmt.Fprintf(w, "%s%s = (", g.indent(), okIdent.Name)
	}
	g.emitExpr(w, ta.X)
	fmt.Fprintf(w, ".%s == %s_%s);\n", firstMethod, cConcrete, firstMethod)
}

// emitTypeAssertExpr emits a type assertion.
func (g *Generator) emitTypeAssertExpr(w io.Writer, n *ast.TypeAssertExpr) {
	sourceType := g.types.TypeOf(n.X)
	if isEmptyInterface(sourceType) {
		targetType := g.types.TypeOf(n.Type)
		cType := g.mapTypeName(n, targetType)
		if isPointerType(targetType) {
			// Pointer assertion: any.(*Type) -> (Type*)expr
			fmt.Fprintf(w, "(%s)", cType)
			g.emitExpr(w, n.X)
		} else {
			// Value assertion: any.(Type) -> (*(Type*)expr)
			fmt.Fprintf(w, "(*(%s*)", cType)
			g.emitExpr(w, n.X)
			fmt.Fprint(w, ")")
		}
		return
	}

	// Non-empty interface type assertion.
	targetType := g.types.TypeOf(n.Type)
	// A non-empty interface holds no runtime type information, so the
	// generator cannot emit a check against a second interface.
	if isInterfaceType(targetType) {
		g.fail(n, "type assertion from an interface to an interface is not supported")
	}
	elemType, isPtr := ptrElem(targetType)

	// Cast to a pointer or value type, depending on the request.
	concreteNamed := elemType.(*types.Named)
	cConcrete := g.mapTypeName(n, concreteNamed)
	if isPtr {
		// Pointer assertion: ival.(*Type) -> (Type*)ival.self
		fmt.Fprintf(w, "(%s*)", cConcrete)
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ".self")
	} else {
		// Value assertion: ival.(Type) -> *((Type*)ival.self)
		fmt.Fprintf(w, "*((%s*)", cConcrete)
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ".self)")
	}
}

// emitAnyValue emits an expression as a void* for empty interface storage.
func (g *Generator) emitAnyValue(w io.Writer, node ast.Node, expr ast.Expr) {
	valType := g.types.TypeOf(expr)
	if basic, ok := valType.(*types.Basic); ok && basic.Kind() == types.UntypedNil {
		// Nil values pass as NULL.
		fmt.Fprint(w, "NULL")
		return
	}

	if isPointerType(valType) || isEmptyInterface(valType) {
		// Pointer values pass through as-is (implicitly convertible to void*).
		// Empty interface (any) values pass through as-is (already void*).
		g.emitExpr(w, expr)
		return
	}

	// A non-empty interface is a fat struct, so it is boxed like a value type
	// below: its address is stored in the void*.

	if g.isLvalue(expr) {
		// An lvalue already has an address, so just take it: &val.
		if star, ok := ast.Unparen(expr).(*ast.StarExpr); ok {
			// &*p is p.
			g.emitExpr(w, star.X)
			return
		}
		fmt.Fprint(w, "&")
		g.emitExpr(w, expr)
		return
	}

	if !isScalarType(valType) {
		g.fail(expr, "cannot use a value of type %s as any; assign it to a variable first",
			g.typeString(valType))
	}

	// Every other scalar expression needs a composite literal: &(Type){val}.
	cType := g.mapTypeName(node, valType)
	fmt.Fprintf(w, "&(%s){", cType)
	g.emitExpr(w, expr)
	fmt.Fprint(w, "}")
}

// isLvalue reports whether an expression emits a C lvalue.
func (g *Generator) isLvalue(expr ast.Expr) bool {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		return g.isIdentLvalue(e)
	case *ast.BasicLit:
		// A string literal emits so_str, a composite literal.
		// Every other literal emits a C constant.
		return e.Kind == token.STRING
	case *ast.CompositeLit:
		// A map literal emits a macro call or an address.
		// C cannot take the address of either.
		return !isMapType(g.types.TypeOf(e))
	case *ast.StarExpr:
		// A dereference is an lvalue whatever the operand is.
		return true
	case *ast.SelectorExpr:
		return g.isSelectorLvalue(e)
	case *ast.IndexExpr:
		return g.isIndexLvalue(e)
	}
	return false
}

// isIdentLvalue reports whether an identifier emits a C lvalue.
func (g *Generator) isIdentLvalue(id *ast.Ident) bool {
	switch obj := g.types.ObjectOf(id).(type) {
	case *types.Var:
		return true
	case *types.Const:
		// A declared constant emits a C variable. A predeclared one
		// (true, false, iota) emits a literal or a macro name.
		return obj.Parent() != types.Universe
	}
	return false
}

// isSelectorLvalue reports whether a selector emits a C lvalue.
func (g *Generator) isSelectorLvalue(sel *ast.SelectorExpr) bool {
	info, ok := g.types.Selections[sel]
	if !ok {
		// A qualified identifier. A variable of another package emits a C
		// variable name; a constant can emit a macro name.
		_, isVar := g.types.ObjectOf(sel.Sel).(*types.Var)
		return isVar
	}
	if info.Kind() != types.FieldVal {
		// A method value is not a field read.
		return false
	}
	if info.Indirect() {
		// The selector emits x->f, an lvalue whatever x is.
		return true
	}
	return g.isLvalue(sel.X)
}

// isIndexLvalue reports whether an index expression emits a C lvalue.
func (g *Generator) isIndexLvalue(idx *ast.IndexExpr) bool {
	switch t := g.types.TypeOf(idx.X).Underlying().(type) {
	case *types.Array:
		// The index emits a[i], an lvalue when a is one.
		return g.isLvalue(idx.X)
	case *types.Pointer:
		// The index emits (*p)[i], an lvalue whatever p is.
		_, isArray := t.Elem().Underlying().(*types.Array)
		return isArray
	case *types.Slice:
		// The index emits so_at, which dereferences a pointer.
		return true
	case *types.Basic:
		// A string index also emits so_at.
		return t.Kind() == types.String || t.Kind() == types.UntypedString
	}
	// A map read emits so_map_get, which returns a value.
	// A generic instantiation is not a value at all.
	return false
}

// checkMethodReceivers rejects interface methods that are declared with
// a value receiver on the concrete type instead of a pointer receiver.
func (g *Generator) checkMethodReceivers(node ast.Node, iface *types.Interface, concrete *types.Named) {
	ptr := types.NewPointer(concrete)
	for m := range iface.Methods() {
		obj, _, _ := types.LookupFieldOrMethod(ptr, true, m.Pkg(), m.Name())
		fn, ok := obj.(*types.Func)
		if !ok {
			g.fail(node, "method %s not found on %s", m.Name(), g.typeString(concrete))
		}
		// Go allows a value receiver here, since the method set of *T includes
		// the methods of T. C does not: the vtable slot expects a function
		// taking void* self, and a value receiver compiles to a function
		// taking the struct itself.
		if _, isPtr := fn.Signature().Recv().Type().(*types.Pointer); !isPtr {
			g.fail(node, "method %s.%s has a value receiver; interface methods must use pointer receivers", g.typeString(concrete), m.Name())
		}
	}
}

// checkEmbeddedIfaces rejects an interface embedded in another interface.
func (g *Generator) checkEmbeddedIfaces(it *ast.InterfaceType) {
	if isConstraintInterface(g.types.TypeOf(it)) {
		// A constraint interface never reaches C,
		// so what it embeds does not matter.
		return
	}
	for _, elem := range it.Methods.List {
		if _, isMethod := elem.Type.(*ast.FuncType); isMethod {
			continue
		}
		typ := g.types.TypeOf(elem.Type)
		if typ == nil {
			continue
		}
		iface, ok := typ.Underlying().(*types.Interface)
		if !ok || iface.NumMethods() == 0 {
			continue
		}
		g.fail(elem, "embedded interface %s is not supported; declare its methods instead", g.typeString(typ))
	}
}

// comparableInC reports whether == and != can compare the C representations
// of x and y.
func comparableInC(x, y types.Type) bool {
	// A named interface is a struct with a pointer to the value, so
	// it only compares with another named interface. An empty interface is a
	// void*, so it compares with a pointer, but not with a value.
	if isNamedNonEmptyInterface(x) || isNamedNonEmptyInterface(y) {
		return isNamedNonEmptyInterface(x) && isNamedNonEmptyInterface(y)
	}
	if isEmptyInterface(x) || isEmptyInterface(y) {
		return isVoidPtrOperand(x) && isVoidPtrOperand(y)
	}
	return true
}

// isVoidPtrOperand reports whether a value of type t can be compared
// with a void*.
func isVoidPtrOperand(t types.Type) bool {
	return isPointerType(t) || isEmptyInterface(t)
}

// isNamedNonEmptyInterface reports whether t is a named non-empty interface.
func isNamedNonEmptyInterface(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok || iface.Empty() {
		return false
	}
	_, isNamed := types.Unalias(t).(*types.Named)
	return isNamed
}

// isConcreteNamedType reports whether t is a named type (or pointer to named type)
// that is not an interface. This is used to decide if a value can be wrapped
// as an interface literal (excludes nil, basic types, etc.).
func isConcreteNamedType(t types.Type) bool {
	if isInterfaceType(t) {
		return false
	}
	elem, _ := ptrElem(t)
	_, ok := elem.(*types.Named)
	return ok
}

// isInterfaceType reports whether t is an interface type.
func isInterfaceType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

// isConstraintInterface reports whether t is a constraint interface.
// Go allows such an interface only as a type parameter constraint,
// so it never becomes a C type.
func isConstraintInterface(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	return ok && !iface.IsMethodSet()
}

// isEmptyInterface reports whether t is an empty interface
// (interface{} or any), which maps to void*.
func isEmptyInterface(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	return ok && iface.Empty()
}
