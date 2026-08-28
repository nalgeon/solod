package clang

import (
	"go/ast"
	"go/types"
	"strings"
)

// CType represents a C type with all information
// needed to declare a variable of that type.
type CType struct {
	Base       string   // e.g. "int", "so_int"
	Dims       string   // e.g. "[3]", "[2][3]", ""
	PtrToArray bool     // pointer-to-array: so_int (*name)[3]
	FuncPtr    bool     // function pointer: ret (*name)(params)
	FuncRet    string   // return type for a FuncPtr
	FuncParams []string // parameter types for a FuncPtr
}

// Decl formats a C declaration:
//   - regular variable: so_int count
//   - array variable: so_int arr[3]
//   - array pointer: so_int (*arr)[3]
//   - func pointer: bool (*fn)(double)
func (t CType) Decl(name string) string {
	if t.FuncPtr {
		return t.FuncRet + " (*" + name + ")(" + funcParams(t.FuncParams) + ")"
	}
	if t.PtrToArray {
		return t.Base + " (*" + name + ")" + t.Dims
	}
	return t.Base + " " + name + t.Dims
}

// Cast formats a C type token for a cast expression:
//   - regular type: so_int
//   - array pointer: so_int (*)[3]
//   - func pointer: bool (*)(double)
func (t CType) Cast() string {
	return strings.TrimSpace(t.Decl(""))
}

// IsArray reports whether this is an array type.
func (t CType) IsArray() bool {
	return t.Dims != "" && !t.PtrToArray
}

// IsPointer reports whether the declaration uses a `*` to indicate a pointer.
// A named pointer type is a typedef, so it hides the `*` and reports false.
func (t CType) IsPointer() bool {
	return t.PtrToArray || strings.HasSuffix(t.Base, "*")
}

// mapTypeDecl maps a Go type to a [CType].
//
// Use it when the type introduces a name (locals, params, struct fields,
// typedefs), because C declaration syntax is name-embedded: arrays `T x[N]`,
// function pointers `R (*x)(A)`, pointer-to-array `T (*x)[N]`.
//
// For a bare inline type token (casts, macro args, compound-literal
// element types), use [Generator.mapTypeName] instead.
func (g *Generator) mapTypeDecl(node ast.Node, typ types.Type) CType {
	// Pointer to array.
	if ptr, ok := types.Unalias(typ).(*types.Pointer); ok {
		if _, ok := types.Unalias(ptr.Elem()).(*types.Array); ok {
			return CType{
				Base:       g.mapTypeName(node, ptr.Elem()),
				Dims:       arrayDims(ptr.Elem()),
				PtrToArray: true,
			}
		}
	}
	// Anonymous function.
	if sig, ok := types.Unalias(typ).(*types.Signature); ok {
		var params []string
		for p := range sig.Params().Variables() {
			params = append(params, g.mapTypeName(node, p.Type()))
		}
		return CType{
			FuncPtr:    true,
			FuncRet:    g.returnType(node, sig),
			FuncParams: params,
		}
	}
	// Anonymous struct is only valid as a local variable with an initializer:
	// `s := struct{ v int }{42}` translates to `so_auto s = (struct{ so_int v; }){42}`.
	// Value contexts (slice/array elements, params, returns) go through
	// mapTypeName instead, which fails because there is no C type name.
	if _, ok := types.Unalias(typ).(*types.Struct); ok {
		return CType{Base: "so_auto"}
	}
	// Regular type (including arrays and named function types).
	return CType{
		Base: g.mapTypeName(node, typ),
		Dims: arrayDims(typ),
	}
}

// mapTypeName maps a Go type to a bare C type token.
//
// Use it when the type stands on its own: casts, macro arguments like
// `so_at(T, ...)`, compound-literal element types like `(T[N]){...}`,
// and return types.
//
// For a named declaration, use [Generator.mapTypeDecl] instead.
//
// For an array type mapTypeName returns the innermost element token and drops
// the dimensions. If you need them, pair it with [arrayDims] or use mapTypeDecl.
func (g *Generator) mapTypeName(node ast.Node, typ types.Type) string {
	typ = types.Unalias(typ)

	// Complex types (e.g. pointers, named types, structs).
	switch t := typ.(type) {
	case *types.Array:
		// Return the innermost non-array element type.
		elem := t.Elem()
		for inner, ok := elem.(*types.Array); ok; inner, ok = elem.(*types.Array) {
			elem = inner.Elem()
		}
		return g.mapTypeName(node, elem)

	case *types.Slice:
		g.checkSliceElemType(node, t.Elem())
		return "so_Slice"

	case *types.Map:
		if _, ok := t.Elem().Underlying().(*types.Array); ok {
			g.fail(node, "array as map value type is not supported")
		}
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
			if info, ok := g.getExtern(obj); ok && info.name != "" {
				return info.name
			}
			return obj.Pkg().Name() + "_" + obj.Name()
		}
		if obj.Parent() == g.pkg.Types.Scope() {
			return g.symbolName(obj)
		}
		return obj.Name()

	case *types.Pointer:
		elem := t.Elem()
		if _, ok := types.Unalias(elem).(*types.Array); ok {
			return g.mapTypeName(node, elem) + "(*)" + arrayDims(elem)
		}
		return g.mapTypeName(node, elem) + "*"

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
				return g.symbolName(tn)
			}
		}
		g.fail(node, "no matching function type for signature")

	case *types.Struct:
		// Anonymous structs have no C type name, so they are unusable
		// in value contexts (slice/array elements, params, returns).
		g.fail(node, "use a named struct type instead of an anonymous struct")

	case *types.TypeParam:
		// A type parameter only exists as a macro argument. Anywhere else it
		// would be emitted as a literal C type name, which does not exist.
		if !g.state.inMacro {
			g.fail(node, "type parameter %s is only allowed in a so:inline macro", t.Obj().Name())
		}
		return t.Obj().Name()
	}

	// Non-basic types that reach this point are not supported.
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported type: %s", g.typeString(typ))
	}

	// Basic types (e.g. int, bool, string).
	switch basic.Kind() {
	case types.Bool, types.UntypedBool:
		return "bool"
	case types.Float32:
		return "float"
	case types.Float64, types.UntypedFloat:
		return "double"
	case types.Int:
		return "so_int"
	case types.UntypedInt:
		return "int64_t"
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
		return "so_uint"
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
		g.fail(node, "unsupported type: %s", g.typeString(typ))
		panic("unreachable")
	}
}

// zeroValue returns the C zero value for a Go type.
//
// Aggregates use the empty initializer {}, defined in C23 and accepted
// by GCC and Clang with -std=gnu11. The generator does not use {0}, because
// a struct{} type has no member to set, so {0} is not valid for struct{},
// for an array of struct{}, or for a struct with a struct{} first field.
func (g *Generator) zeroValue(node ast.Node, typ types.Type) string {
	// Arrays.
	if _, ok := typ.Underlying().(*types.Array); ok {
		return "{}"
	}

	// Pointers.
	if _, ok := typ.Underlying().(*types.Pointer); ok {
		return "NULL"
	}

	// Slices.
	if _, ok := typ.Underlying().(*types.Slice); ok {
		return "{}"
	}

	// Maps.
	if _, ok := typ.Underlying().(*types.Map); ok {
		return "NULL"
	}

	// Functions (map to C function pointers).
	if _, ok := typ.Underlying().(*types.Signature); ok {
		return "NULL"
	}

	// Structs.
	if _, ok := typ.Underlying().(*types.Struct); ok {
		return "{}"
	}

	// Error type.
	if isErrorType(typ) {
		return "{}"
	}

	// Interfaces.
	if iface, ok := typ.Underlying().(*types.Interface); ok {
		if iface.Empty() {
			// any (interface{}) maps to void*, so zero value is NULL.
			return "NULL"
		}
		if _, ok := types.Unalias(typ).(*types.Named); ok {
			// Named interfaces map to structs, so zero value is {}.
			return "{}"
		}
		g.fail(node, "unsupported non-empty anonymous interface")
	}

	// Non-basic types that reach this point are not supported.
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		g.fail(node, "unsupported type: %s", g.typeString(typ))
	}

	// Basic types (e.g. int, bool, string).
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
		g.fail(node, "unsupported type for zero value: %s", g.typeString(typ))
		panic("unreachable")
	}
}

// declSymbolName returns the C name for a declaration that could be
// either package-level or function-local.
func (g *Generator) declSymbolName(obj types.Object) string {
	if g.state.atTopLevel() {
		return g.symbolName(obj)
	}
	return obj.Name()
}

// symbolName returns the C symbol name for a Go identifier.
// Exported names are prefixed with the package name (e.g. RectArea -> geom_RectArea).
// Unexported names marked with so:promote get the same prefix.
// Extern symbols with a name override use the specified C name instead.
func (g *Generator) symbolName(obj types.Object) string {
	if info, ok := g.getExtern(obj); ok && info.name != "" {
		return info.name
	}
	name := obj.Name()
	if ast.IsExported(name) || g.promoted[obj] {
		return g.pkg.Name + "_" + name
	}
	return name
}

// constType returns the type to declare a constant with. Untyped int normally
// maps to int64_t; values larger than math.MaxInt64 map to uint64_t.
func (g *Generator) constType(node ast.Node, obj types.Object) types.Type {
	typ := obj.Type()
	c, ok := obj.(*types.Const)
	if !ok || !emitsAsUint64(typ, c.Val()) {
		return typ
	}
	if exceedsUint64(c.Val()) {
		g.fail(node, "constant %s does not fit in int64 or uint64", obj.Name())
	}
	return types.Typ[types.Uint64]
}

// isErrorType checks if a type is the built-in error interface.
func isErrorType(typ types.Type) bool {
	if named, ok := types.Unalias(typ).(*types.Named); ok {
		return named.Obj().Name() == "error" && named.Obj().Parent() == types.Universe
	}
	return false
}

// isNilType checks if a type is the untyped nil.
func isNilType(t types.Type) bool {
	basic, ok := t.(*types.Basic)
	return ok && basic.Kind() == types.UntypedNil
}

// isPointerType reports whether t is a pointer type. unsafe.Pointer counts:
// it maps to void*, so it behaves like a pointer everywhere in the generated C.
func isPointerType(t types.Type) bool {
	switch u := t.Underlying().(type) {
	case *types.Pointer:
		return true
	case *types.Basic:
		return u.Kind() == types.UnsafePointer
	}
	return false
}
