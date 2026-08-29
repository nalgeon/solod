package clang

import (
	"go/ast"
	"go/token"
	"go/types"
)

// pkgValue is a package-level constant or variable declaration,
// together with its initializer.
type pkgValue struct {
	obj   types.Object
	name  *ast.Ident
	value ast.Expr // the emitted initializer, or nil
}

// checkValueOrder rejects a package-level constant or variable which references
// a constant or variable declared after it.
func (g *Generator) checkValueOrder() {
	header, impl := g.valueSeqs()
	// Nothing is available before the header.
	g.checkValueSeq(header, map[types.Object]bool{})
	// The .c file includes the header, so every header name is available in it.
	g.checkValueSeq(impl, valueSet(header))
}

// valueSeqs returns the package-level constants and variables
// in the order the header and the .c file emit them.
func (g *Generator) valueSeqs() (header, impl []pkgValue) {
	for _, sym := range g.symbols {
		if sym.kind != symbolVar && sym.kind != symbolConst {
			continue
		}
		isConst := sym.genDecl.Tok == token.CONST
		for _, spec := range sym.genDecl.Specs {
			vs := spec.(*ast.ValueSpec)
			for i, name := range vs.Names {
				v := pkgValue{obj: g.types.Defs[name], name: name}
				if !isIotaValue(vs, i) {
					v.value = vs.Values[i]
				}
				switch {
				case !ast.IsExported(name.Name) && !sym.dirs.promote:
					// The .c file holds the declaration and the value.
					impl = append(impl, v)
				case isConst:
					// The header holds the declaration and the value.
					header = append(header, v)
				default:
					// The header holds an extern declaration
					// and the .c file holds the value.
					header = append(header, pkgValue{obj: v.obj, name: name})
					impl = append(impl, v)
				}
			}
		}
	}
	return header, impl
}

// checkValueSeq walks the sequence and rejects any value which references
// a name that is not available yet. The avail set holds the names which are
// available before the sequence starts.
func (g *Generator) checkValueSeq(seq []pkgValue, avail map[types.Object]bool) {
	for _, v := range seq {
		if v.value != nil {
			g.checkValueRefs(v, avail)
		}
		avail[v.obj] = true
	}
}

// checkValueRefs rejects a value which references a name outside the avail set.
func (g *Generator) checkValueRefs(v pkgValue, avail map[types.Object]bool) {
	for _, ref := range g.valueRefs(v.value) {
		obj := g.types.Uses[ref]
		if avail[obj] {
			continue
		}
		kind := "variable"
		if _, ok := obj.(*types.Const); ok {
			kind = "constant"
		}
		g.fail(ref, "%s %s is declared after %s; move the declaration up",
			kind, ref.Name, v.name.Name)
	}
}

// valueRefs returns the identifiers of the package-level constants
// and variables that an expression references.
func (g *Generator) valueRefs(val ast.Expr) []*ast.Ident {
	var refs []*ast.Ident
	ast.Inspect(val, func(n ast.Node) bool {
		if arr, ok := n.(*ast.ArrayType); ok {
			// C gets an array length as a calculated value,
			// so the length needs no declaration.
			refs = append(refs, g.valueRefs(arr.Elt)...)
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		obj := g.types.Uses[ident]
		if obj == nil {
			return true
		}
		switch obj.(type) {
		case *types.Const, *types.Var:
		default:
			return true
		}
		// A struct field and an imported name are not package-level.
		if obj.Pkg() != g.pkg.Types || obj.Parent() != g.pkg.Types.Scope() {
			return true
		}
		// An extern name comes from a C header, so it needs no declaration.
		if g.hasExtern(obj) {
			return true
		}
		refs = append(refs, ident)
		return true
	})
	return refs
}

// valueSet returns the objects of a sequence.
func valueSet(seq []pkgValue) map[types.Object]bool {
	set := make(map[types.Object]bool, len(seq))
	for _, v := range seq {
		set[v.obj] = true
	}
	return set
}

// isIotaValue reports whether the generator emits the i-th name of a constant
// spec as the value resolved by the type checker instead of the source
// expression.
func isIotaValue(spec *ast.ValueSpec, i int) bool {
	return i >= len(spec.Values) || containsIota(spec.Values[i])
}
