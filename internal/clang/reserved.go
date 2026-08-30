package clang

import (
	"go/ast"
	"go/types"
)

// reservedC contains C identifiers that would conflict with C grammar if
// used directly. This includes C keywords and some built-in macros.
var reservedC = map[string]bool{
	// C keywords that are valid Go identifiers.
	"auto": true, "char": true, "do": true, "double": true, "enum": true,
	"extern": true, "float": true, "inline": true, "int": true, "long": true,
	"register": true, "restrict": true, "short": true, "signed": true,
	"sizeof": true, "static": true, "typedef": true, "union": true,
	"unsigned": true, "void": true, "volatile": true, "while": true,
	// C23 keywords and predeclared Go identifiers that map to C keywords/macros.
	"bool": true, "true": true, "false": true, "nullptr": true,
	"alignas": true, "alignof": true, "constexpr": true, "typeof": true,
	"thread_local": true, "static_assert": true,
	// Macros from builtin.h headers. Best-effort: a macro from a system
	// header that is not listed here can still conflict.
	"assert": true, "offsetof": true,
}

// reservedSo contains identifiers that the generator emits itself.
var reservedSo = map[string]bool{
	// The method receiver parameter and the interface data pointer.
	"self": true,
}

// handleReservedNames rewrites identifiers that conflict with reserved C names
// and reject identifiers that conflict with reserved So names.
//
// Function-local variables, parameters, and constants are mangled by adding
// "_" at the end (Go "long" -> C "long_"). This works because every time a
// local variable's name is used, it's read directly from its AST identifier,
// not from the types object. So, changing the identifier updates both the
// declaration and all uses at once.
//
// Reserved names that cross the AST/types boundary - struct fields and
// package-level declarations - can't be renamed safely, so they are
// rejected at the declaration.
func (g *Generator) handleReservedNames() {
	pkgScope := g.pkg.Types.Scope()
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			ident, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			if reservedSo[ident.Name] {
				// Report the name once, at the declaration.
				if _, isDef := g.types.Defs[ident]; isDef {
					g.fail(ident, "identifier %q is reserved by the compiler; rename it", ident.Name)
				}
				return true
			}
			if !reservedC[ident.Name] {
				return true
			}
			obj := g.types.Uses[ident]
			if obj == nil {
				obj = g.types.Defs[ident]
			}
			if obj == nil {
				// Identifier that represents no object, like the package clause name.
				// It's never used as a C identifier, so there's nothing to mangle.
				return true
			}
			if isLocal(obj, pkgScope) {
				mangled := ident.Name + "_"
				if scopeDeclares(obj.Parent(), mangled) {
					g.fail(ident, "mangled name %q for reserved word %q collides with an existing identifier; rename one of them", mangled, ident.Name)
				}
				ident.Name = mangled
				return true
			}
			if isConcreteMethod(obj) {
				// A method's C name carries its receiver-type prefix
				// (Value.float -> slog_Value_float), never emitted bare.
				return true
			}
			// Reserved name we can't mangle. Report it once, at the declaration.
			if _, isDef := g.types.Defs[ident]; isDef {
				g.fail(ident, "identifier %q is a reserved C word; rename it", ident.Name)
			}
			return true
		})
	}
}

// scopeDeclares reports whether scope itself already declares name.
//
// Only the mangled variable's own scope matters. A C redefinition only happens
// within a single block, and go/types scopes match C blocks one-to-one.
// A name in an outer or inner scope is a valid C shadow, not a conflict,
// so those scopes aren't checked.
func scopeDeclares(scope *types.Scope, name string) bool {
	return scope.Lookup(name) != nil
}

// isLocal reports whether obj is a function-local variable, parameter, or
// constant, as opposed to a struct field or a package-level declaration.
// Locals are safe to mangle because their names never cross an API boundary.
func isLocal(obj types.Object, pkgScope *types.Scope) bool {
	switch obj := obj.(type) {
	case *types.Var:
		if obj.IsField() {
			return false
		}
	case *types.Const:
		// A function-local constant; mangled like a local variable.
	default:
		return false
	}
	parent := obj.Parent()
	return parent != nil && parent != pkgScope && parent != types.Universe
}

// isConcreteMethod reports whether obj is a method with a non-interface
// receiver. Interface methods have a receiver too, but they emit as bare
// struct fields, so they are excluded.
func isConcreteMethod(obj types.Object) bool {
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	recv := fn.Signature().Recv()
	return recv != nil && !types.IsInterface(recv.Type())
}
