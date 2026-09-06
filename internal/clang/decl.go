package clang

import (
	"go/ast"
	"go/types"
)

// A multi-variable declaration groups consecutive names of the same C type into
// one declaration, like `so_int a = 1, b = 2;`.

// groupable reports whether a declaration of ctyp can hold more than one declarator.
func groupable(ctyp CType, typ types.Type) bool {
	return (
	// an array needs a dimension, like `so_byte a[8]`
	!ctyp.IsArray() &&
		// `T* a, b` declares a as T* but b as T
		!ctyp.IsPointer() &&
		// a function pointer has the name inside the declarator
		!ctyp.FuncPtr &&
		// __auto_type allows only one declarator per statement
		!isAnonStruct(typ) &&
		// a macro takes a pointer type for a type parameter, which gives `T* a, b`
		!isTypeParam(typ))
}

// groupType reports whether name can join a declaration of ctyp. It also returns
// the type of the name, which the caller needs to emit the initializer.
// hasValue reports whether the declaration gives name an initializer.
func (g *Generator) groupType(node ast.Node, name *ast.Ident, ctyp CType, hasValue bool) (types.Type, bool) {
	if name.Name == "_" {
		return nil, false
	}
	def := g.types.Defs[name]
	if def == nil {
		// Redeclared variable, which needs an assignment instead.
		return nil, false
	}
	typ := def.Type()
	nodeCtyp := g.mapVarType(node, typ, hasValue)
	if nodeCtyp.Base != ctyp.Base || !groupable(nodeCtyp, typ) {
		return nil, false
	}
	return typ, true
}
