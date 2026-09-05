package clang

import "go/types"

// sortTypes arranges type declarations so that each type follows the types it
// depends on. Types that don't depend on each other stay in their original order.
func (g *Generator) sortTypes(syms []symbol) []symbol {
	byObj := make(map[types.Object]symbol, len(syms))
	for _, sym := range syms {
		byObj[g.types.Defs[sym.typeSpec.Name]] = sym
	}

	const (
		visiting = 1
		done     = 2
	)
	state := make(map[types.Object]int, len(syms))
	sorted := make([]symbol, 0, len(syms))

	var visit func(sym symbol)
	visit = func(sym symbol) {
		obj := g.types.Defs[sym.typeSpec.Name]
		switch state[obj] {
		case done:
			return
		case visiting:
			// Reached for types Go accepts but C cannot express as a typedef,
			// such as `type StateFn func() StateFn` or `type Tree [2]*Tree`:
			// only a struct can close a cycle through a forward declaration.
			g.fail(sym.typeSpec.Name, "recursive type %s cannot be expressed in C", sym.typeSpec.Name.Name)
		}
		state[obj] = visiting
		for _, dep := range g.typeDeps(sym) {
			if depSym, ok := byObj[dep]; ok {
				visit(depSym)
			}
		}
		state[obj] = done
		sorted = append(sorted, sym)
	}

	for _, sym := range syms {
		visit(sym)
	}
	return sorted
}

// typeDeps returns the types that sym's declaration must follow.
//
// A reference only creates a dependency when the forward declarations from
// [Generator.emitForwardTypeDecls] do not already satisfy it. Those cover
// every type with a struct underlying type: a struct spec and a spec that
// names another struct type. An incomplete struct is enough behind a pointer
// or as a parameter or result of a function pointer. Everywhere else - a
// value field, an array element, the target of a typedef - the definition
// must come first.
//
// Objects from other packages may appear in the result; the caller ignores
// any that are not package-level types of its own.
func (g *Generator) typeDeps(sym symbol) []types.Object {
	var deps []types.Object

	// indirect reports whether an incomplete type is acceptable at this position.
	var walk func(typ types.Type, indirect bool)
	walk = func(typ types.Type, indirect bool) {
		switch t := types.Unalias(typ).(type) {
		case *types.Named:
			if _, isStruct := t.Underlying().(*types.Struct); isStruct && indirect {
				return
			}
			deps = append(deps, t.Obj())
		case *types.Pointer:
			walk(t.Elem(), true)
		case *types.Array:
			walk(t.Elem(), false)
		case *types.Struct:
			for f := range t.Fields() {
				walk(f.Type(), false)
			}
		case *types.Signature:
			for p := range t.Params().Variables() {
				walk(p.Type(), true)
			}
			for r := range t.Results().Variables() {
				walk(r.Type(), true)
			}
		case *types.Interface:
			for m := range t.Methods() {
				walk(m.Type(), true)
			}
		}
		// Basic types need no declaration. Slices, maps and channels map to
		// opaque builtins that do not mention their element type.
	}

	// Walking the spec rather than the underlying type keeps `type E P`
	// distinct from a copy of P's definition: the typedef needs P itself.
	walk(g.types.TypeOf(sym.typeSpec.Type), false)
	return deps
}
