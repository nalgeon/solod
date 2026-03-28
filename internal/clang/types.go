package clang

import (
	"go/ast"
	"go/types"
)

// CType represents a C type with optional array dimensions.
type CType struct {
	Base       string // e.g. "int", "so_int"
	Dims       string // e.g. "[3]", "[2][3]", ""
	PtrToArray bool   // pointer-to-array: so_int (*name)[3]
}

// Decl formats a C declaration: "int name[3]".
func (t CType) Decl(name string) string {
	if t.PtrToArray {
		return t.Base + " (*" + name + ")" + t.Dims
	}
	return t.Base + " " + name + t.Dims
}

// IsArray reports whether this is an array type.
func (t CType) IsArray() bool {
	return t.Dims != "" && !t.PtrToArray
}

// mapCType maps a Go type to a CType (base + array dims).
func (g *Generator) mapCType(node ast.Node, typ types.Type) CType {
	if ptr, ok := types.Unalias(typ).(*types.Pointer); ok {
		if _, ok := types.Unalias(ptr.Elem()).(*types.Array); ok {
			return CType{
				Base:       g.mapType(node, ptr.Elem()),
				Dims:       arrayDims(ptr.Elem()),
				PtrToArray: true,
			}
		}
	}
	return CType{
		Base: g.mapType(node, typ),
		Dims: arrayDims(typ),
	}
}

// mapType maps a Go type to its C equivalent.
func (g *Generator) mapType(node ast.Node, typ types.Type) string {
	typ = types.Unalias(typ)

	// Complex types (e.g. pointers, named types, structs).
	switch t := typ.(type) {
	case *types.Array:
		// Return the innermost non-array element type.
		elem := t.Elem()
		for inner, ok := elem.(*types.Array); ok; inner, ok = elem.(*types.Array) {
			elem = inner.Elem()
		}
		return g.mapType(node, elem)

	case *types.Slice:
		return "so_Slice"

	case *types.Map:
		return "so_Map*"

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
		obj := t.Obj()
		if obj.Pkg() != nil && obj.Pkg() != g.pkg.Types {
			// This is a named type from another package.
			return obj.Pkg().Name() + "_" + obj.Name()
		}
		if obj.Parent() == g.pkg.Types.Scope() {
			return g.symbolName(obj.Name())
		}
		return obj.Name()

	case *types.Pointer:
		elem := t.Elem()
		if _, ok := types.Unalias(elem).(*types.Array); ok {
			return g.mapType(node, elem) + "(*)" + arrayDims(elem)
		}
		return g.mapType(node, elem) + "*"

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
	case types.Int32:
		if basic.Name() == "rune" {
			return "so_rune"
		}
		return "int32_t"
	case types.UntypedRune:
		return "so_rune"
	case types.Int64:
		return "int64_t"
	case types.Uint:
		return "uint64_t"
	case types.Uint8:
		if basic.Name() == "byte" {
			return "so_byte"
		}
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
	case types.UnsafePointer:
		return "void*"
	default:
		g.fail(node, "unsupported type: %s", typ)
		panic("unreachable")
	}
}

// zeroValue returns the C zero value for a Go type.
func (g *Generator) zeroValue(node ast.Node, typ types.Type) string {
	// Arrays.
	if _, ok := typ.Underlying().(*types.Array); ok {
		return "{0}"
	}

	// Pointers.
	if _, ok := typ.Underlying().(*types.Pointer); ok {
		return "NULL"
	}

	// Slices.
	if _, ok := typ.Underlying().(*types.Slice); ok {
		return "{0}"
	}

	// Maps.
	if _, ok := typ.Underlying().(*types.Map); ok {
		return "NULL"
	}

	// Structs.
	if _, ok := typ.Underlying().(*types.Struct); ok {
		return "{0}"
	}

	// Error type.
	if isErrorType(typ) {
		return "NULL"
	}

	// Interfaces.
	if iface, ok := typ.Underlying().(*types.Interface); ok {
		if iface.Empty() {
			// any (interface{}) maps to void*, so zero value is NULL.
			return "NULL"
		}
		if _, ok := typ.(*types.Named); ok {
			// Named interfaces map to structs, so zero value is {0}.
			return "{0}"
		}
		g.fail(node, "unsupported non-empty anonymous interface")
	}

	// Basic types (e.g. int, bool, string).
	basic := typ.Underlying().(*types.Basic)
	switch basic.Kind() {
	case types.Bool:
		return "false"
	case types.String, types.UntypedString:
		return `so_str("")`
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
		types.Float32, types.Float64:
		return "0"
	default:
		g.fail(node, "unsupported type for zero value: %s", typ)
		panic("unreachable")
	}
}

// declSymbolName returns the C name for a declaration that could be
// either package-level or function-local.
func (g *Generator) declSymbolName(name string) string {
	if g.state.indent == 0 {
		return g.symbolName(name)
	}
	return name
}

// symbolName returns the C symbol name for a Go identifier.
// Exported names are prefixed with the package name (e.g. RectArea -> geom_RectArea).
// Extern symbols with a name override use the specified C name instead.
func (g *Generator) symbolName(name string) string {
	if info, ok := g.getExtern("", name); ok && info.name != "" {
		return info.name
	}
	if ast.IsExported(name) {
		return g.pkg.Name + "_" + name
	}
	return name
}

// isUnexportedType reports whether a type is unexported for the current package.
func (g *Generator) isUnexportedType(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() != g.pkg.Types {
		return false
	}
	return !ast.IsExported(obj.Name())
}

// isErrorType checks if a type is the built-in error interface.
func isErrorType(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		return named.Obj().Name() == "error" && named.Obj().Parent() == types.Universe
	}
	return false
}

// isNilType checks if a type is the untyped nil.
func isNilType(t types.Type) bool {
	basic, ok := t.(*types.Basic)
	return ok && basic.Kind() == types.UntypedNil
}
