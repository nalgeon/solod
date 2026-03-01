package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"
)

// emitInterfaceTypeSpec emits a typedef struct with void* self and function pointers.
func (g *Generator) emitInterfaceTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	typ := g.types.Defs[spec.Name].Type().(*types.Named)
	iface := typ.Underlying().(*types.Interface)
	cName := g.symbolName(spec.Name.Name)
	fmt.Fprintf(w, "\ntypedef struct %s {\n", cName)
	fmt.Fprintf(w, "    void* self;\n")
	for m := range iface.Methods() {
		sig := m.Type().(*types.Signature)
		retType := g.returnType(spec, sig)
		var params strings.Builder
		params.WriteString("void* self")
		for p := range sig.Params().Variables() {
			params.WriteString(", ")
			params.WriteString(g.mapType(spec, p.Type()))
			params.WriteString(" ")
			params.WriteString(p.Name())
		}
		fmt.Fprintf(w, "    %s (*%s)(%s);\n", retType, m.Name(), params.String())
	}
	fmt.Fprintf(w, "} %s;\n", cName)
}

// emitInterfaceLit emits a compound literal that wraps a concrete value as an interface.
// Example: (main_Shape){.self = &r, .Area = main_Rect_Area, .Perim = main_Rect_Perim}
func (g *Generator) emitInterfaceLit(ifaceType types.Type, expr ast.Expr) {
	w := g.state.writer
	named := ifaceType.(*types.Named)
	iface := named.Underlying().(*types.Interface)

	// Get value type, dereferencing if it's a pointer.
	concreteType := g.types.TypeOf(expr)
	isPtr := false
	if ptr, ok := concreteType.(*types.Pointer); ok {
		concreteType = ptr.Elem()
		isPtr = true
	}
	concreteNamed := concreteType.(*types.Named)

	cIface := g.symbolName(named.Obj().Name())
	cConcrete := g.symbolName(concreteNamed.Obj().Name())

	if isPtr {
		fmt.Fprintf(w, "(%s){.self = ", cIface)
	} else {
		fmt.Fprintf(w, "(%s){.self = &", cIface)
	}
	g.emitExpr(expr)
	for m := range iface.Methods() {
		fmt.Fprintf(w, ", .%s = %s_%s", m.Name(), cConcrete, m.Name())
	}
	fmt.Fprintf(w, "}")
}

// emitTypeAssertion emits a comma-ok type assertion (e.g. _, ok := s.(Rect)).
// Uses function pointer comparison to identify the concrete type.
func (g *Generator) emitTypeAssertion(w io.Writer, stmt *ast.AssignStmt, ta *ast.TypeAssertExpr) {
	ifaceType := g.types.TypeOf(ta.X).(*types.Named)
	iface := ifaceType.Underlying().(*types.Interface)
	firstMethod := iface.Method(0).Name()

	// Get value type, dereferencing if it's a pointer.
	assertedType := g.types.TypeOf(ta.Type)
	if ptr, ok := assertedType.(*types.Pointer); ok {
		assertedType = ptr.Elem()
	}
	concreteNamed := assertedType.(*types.Named)
	cConcrete := g.symbolName(concreteNamed.Obj().Name())

	okIdent := stmt.Lhs[1].(*ast.Ident)
	fmt.Fprintf(w, "%sbool %s = (", g.indent(), okIdent.Name)
	g.emitExpr(ta.X)
	fmt.Fprintf(w, ".%s == %s_%s);\n", firstMethod, cConcrete, firstMethod)
}

// emitTypeAssertExpr emits a type assertion.
func (g *Generator) emitTypeAssertExpr(n *ast.TypeAssertExpr) {
	sourceType := g.types.TypeOf(n.X)
	if iface, ok := sourceType.Underlying().(*types.Interface); ok && iface.Empty() {
		// Empty interface, emit a simple cast: (void*)expr
		cType := g.mapType(n, g.types.TypeOf(n.Type))
		fmt.Fprintf(g.state.writer, "(%s)", cType)
		g.emitExpr(n.X)
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
	concreteNamed := targetType.(*types.Named)
	cConcrete := g.symbolName(concreteNamed.Obj().Name())
	if isPtr {
		// Pointer assertion: ival.(*Type) → (Type*)ival.self
		fmt.Fprintf(g.state.writer, "(%s*)", cConcrete)
		g.emitExpr(n.X)
		fmt.Fprintf(g.state.writer, ".self")
	} else {
		// Value assertion: ival.(Type) → *((Type*)ival.self)
		fmt.Fprintf(g.state.writer, "*((%s*)", cConcrete)
		g.emitExpr(n.X)
		fmt.Fprintf(g.state.writer, ".self)")
	}
}

// emitAnyValue emits an expression as a void* for empty interface storage.
// Pointers and interface values pass through as-is.
// Value types are wrapped in a compound literal: &(type){val}.
func (g *Generator) emitAnyValue(node ast.Node, expr ast.Expr) {
	valType := g.types.TypeOf(expr)
	if basic, ok := valType.(*types.Basic); ok && basic.Kind() == types.UntypedNil {
		fmt.Fprintf(g.state.writer, "NULL")
		return
	}
	_, isPtr := valType.Underlying().(*types.Pointer)
	_, isIface := valType.Underlying().(*types.Interface)
	if isPtr || isIface {
		g.emitExpr(expr)
		return
	}
	cType := g.mapType(node, valType)
	fmt.Fprintf(g.state.writer, "&(%s){", cType)
	g.emitExpr(expr)
	fmt.Fprintf(g.state.writer, "}")
}

// isInterfaceType reports whether t is an interface type.
func isInterfaceType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}
