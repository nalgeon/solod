package clang

import (
	"fmt"
	"go/ast"
	"go/types"
)

// mapType maps a Go type to its C equivalent.
func (g *Generator) mapType(node ast.Node, typ types.Type) string {
	typ = types.Unalias(typ)

	// Complex types (e.g. pointers, named types, structs).
	switch t := typ.(type) {
	case *types.Array, *types.Slice:
		return "so_Slice"

	case *types.Interface:
		// Special case: empty interface (any or interface{}) maps to void*.
		// Named interfaces are caught by the *types.Named case below.
		if t.Empty() {
			return "void*"
		}
		g.fail(node, "unsupported non-empty anonymous interface")

	case *types.Named:
		if isErrorType(typ) {
			return "so_Error"
		}
		return g.symbolName(t.Obj().Name())

	case *types.Pointer:
		return g.mapType(node, t.Elem()) + "*"

	case *types.Signature:
		// Look for a named type with the same
		// function signature to use as the C type name.
		scope := g.pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok {
				continue
			}
			if types.Identical(tn.Type().Underlying(), t) {
				return g.symbolName(tn.Name())
			}
		}
		g.fail(node, "no matching function type for signature")

	case *types.Struct:
		return "so_auto"
	}

	// Basic types (e.g. int, bool, string).
	basic := typ.Underlying().(*types.Basic)
	switch basic.Kind() {
	case types.Bool, types.UntypedBool:
		return "bool"
	case types.Float32:
		return "float"
	case types.Float64, types.UntypedFloat:
		return "double"
	case types.Int, types.UntypedInt:
		return "so_int"
	case types.Int8:
		return "int8_t"
	case types.Int16:
		return "int16_t"
	case types.Int32, types.UntypedRune:
		return "int32_t"
	case types.Int64:
		return "int64_t"
	case types.Uint:
		return "uint64_t"
	case types.Uint8:
		return "uint8_t"
	case types.Uint16:
		return "uint16_t"
	case types.Uint32:
		return "uint32_t"
	case types.Uint64:
		return "uint64_t"
	case types.Uintptr:
		return "uintptr_t"
	case types.String, types.UntypedString:
		return "so_String"
	default:
		g.fail(node, "unsupported type: %s", typ)
		panic("unreachable")
	}
}

// zeroValue returns the C zero value for a Go type.
func (g *Generator) zeroValue(node ast.Node, typ types.Type) string {
	// Arrays.
	if arr, ok := typ.Underlying().(*types.Array); ok {
		elemType := g.mapType(node, arr.Elem())
		size := arr.Len()
		return fmt.Sprintf("{(%s[%d]){0}, %d, %d}", elemType, size, size, size)
	}

	// Pointers.
	if _, ok := typ.Underlying().(*types.Pointer); ok {
		return "NULL"
	}

	// Slices.
	if _, ok := typ.Underlying().(*types.Slice); ok {
		return "{0}"
	}

	// Structs.
	if _, ok := typ.Underlying().(*types.Struct); ok {
		return "{0}"
	}

	// Interfaces (e.g. error).
	if iface, ok := typ.Underlying().(*types.Interface); ok {
		if iface.Empty() {
			return "NULL"
		}
		g.fail(node, "unsupported non-empty anonymous interface")
	}

	// Basic types (e.g. int, bool, string).
	basic := typ.Underlying().(*types.Basic)
	switch basic.Kind() {
	case types.Bool:
		return "false"
	case types.String:
		return `so_strlit("")`
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
		types.Float32, types.Float64:
		return "0"
	default:
		g.fail(node, "unsupported type for zero value: %s", typ)
		panic("unreachable")
	}
}

// symbolName returns the C symbol name for a Go identifier.
// Exported names are prefixed with the package name (e.g. RectArea → geom_RectArea).
// Extern symbols keep their original name (they come from C headers).
func (g *Generator) symbolName(name string) string {
	if g.externs[name] {
		return name
	}
	if ast.IsExported(name) {
		return g.pkg.Name + "_" + name
	}
	return name
}

// isErrorType checks if a type is the built-in error interface.
func isErrorType(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		return named.Obj().Name() == "error" && named.Obj().Parent() == types.Universe
	}
	return false
}
