package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// emitInterfaceTypeSpec emits a typedef struct with void* self and function pointers.
func (g *Generator) emitInterfaceTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	typ := types.Unalias(g.types.Defs[spec.Name].Type()).(*types.Named)
	iface := typ.Underlying().(*types.Interface)
	cName := g.declSymbolName(g.types.Defs[spec.Name])
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
			params.WriteString(g.mapTypeName(spec, p.Type()))
		}
		fmt.Fprintf(w, "    %s (*%s)(%s);\n", retType, m.Name(), params.String())
	}
	fmt.Fprintf(w, "} %s;\n", cName)
}

// emitInterfaceLit emits a compound literal that wraps a concrete value as an interface.
// Example: (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim}
func (g *Generator) emitInterfaceLit(w io.Writer, ifaceType types.Type, expr ast.Expr) {
	named := types.Unalias(ifaceType).(*types.Named)
	iface := named.Underlying().(*types.Interface)

	// Get value type, dereferencing if it's a pointer.
	concreteType := g.types.TypeOf(expr)
	isPtr := false
	if ptr, ok := concreteType.(*types.Pointer); ok {
		concreteType = ptr.Elem()
		isPtr = true
	}
	concreteNamed := types.Unalias(concreteType).(*types.Named)
	g.checkPointerReceivers(expr, iface, concreteNamed)

	cIface := g.mapTypeName(expr, named)
	cConcrete := g.mapTypeName(expr, concreteNamed)

	if !isPtr {
		// Unreachable: checkPointerReceivers rejects value receivers, and Go
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
	if iface, ok := sourceType.Underlying().(*types.Interface); ok && iface.Empty() {
		g.fail(ta, "comma-ok type assertion on any is not supported")
	}
	ifaceType := types.Unalias(sourceType).(*types.Named)
	iface := ifaceType.Underlying().(*types.Interface)
	firstMethod := iface.Method(0).Name()

	// Get value type, dereferencing if it's a pointer.
	assertedType := g.types.TypeOf(ta.Type)
	if ptr, ok := assertedType.(*types.Pointer); ok {
		assertedType = ptr.Elem()
	}
	concreteNamed := types.Unalias(assertedType).(*types.Named)
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
	if iface, ok := sourceType.Underlying().(*types.Interface); ok && iface.Empty() {
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
	isPtr := false
	if ptr, ok := targetType.(*types.Pointer); ok {
		targetType = ptr.Elem()
		isPtr = true
	}

	// Cast to a pointer or value type, depending on the request.
	concreteNamed := types.Unalias(targetType).(*types.Named)
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

	iface, isIface := valType.Underlying().(*types.Interface)
	if isPointerType(valType) || (isIface && iface.Empty()) {
		// Pointer values pass through as-is (implicitly convertible to void*).
		// Empty interface (any) values pass through as-is (already void*).
		g.emitExpr(w, expr)
		return
	}

	// A non-empty interface is a fat struct, so it is boxed like a value type
	// below: its address is stored in the void*.

	// Value types must be passed by reference for void* storage.
	// Identifiers, composite literals, and string literals emit as
	// addressable C expressions - just prepend &.
	// Other expressions need wrapping in a compound literal: &(Type){val}.
	addressable := false
	switch e := expr.(type) {
	case *ast.Ident:
		addressable = true
	case *ast.CompositeLit:
		addressable = true
	case *ast.BasicLit:
		addressable = e.Kind == token.STRING
	}

	if addressable {
		fmt.Fprint(w, "&")
		g.emitExpr(w, expr)
		return
	}

	cType := g.mapTypeName(node, valType)
	fmt.Fprintf(w, "&(%s){", cType)
	g.emitExpr(w, expr)
	fmt.Fprint(w, "}")
}

// checkPointerReceivers verifies that every interface method
// is declared with a pointer receiver on the concrete type.
func (g *Generator) checkPointerReceivers(node ast.Node, iface *types.Interface, concrete *types.Named) {
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
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	_, ok := types.Unalias(t).(*types.Named)
	return ok
}

// isInterfaceType reports whether t is an interface type.
func isInterfaceType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

// isEmptyInterface reports whether t is an empty interface
// (interface{} or any), which maps to void*.
func isEmptyInterface(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	return ok && iface.Empty()
}

// comparableInC reports whether == and != can compare the C representations
// of x and y. A named interface is a struct with a pointer to the value, so
// it only compares with another named interface. An empty interface is a
// void*, so it compares with a pointer, but not with a value.
func comparableInC(x, y types.Type) bool {
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
