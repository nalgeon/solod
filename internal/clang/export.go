package clang

import (
	"go/ast"
	"go/types"
)

// collectPromoted records the C-name objects of unexported symbols marked
// with so:promote. Such symbols are emitted in the header and get the package
// prefix so the header can reference them without colliding across packages.
func (g *Generator) collectPromoted() {
	g.promoted = make(map[types.Object]bool)
	for _, sym := range g.symbols {
		if !sym.dirs.promote {
			continue
		}
		switch sym.kind {
		case symbolType:
			g.promoted[g.types.Defs[sym.typeSpec.Name]] = true
		case symbolFunc:
			// Methods are named after their receiver type, so only a free
			// function needs its own object in the set (see [Generator.symbolName]).
			g.promoted[g.types.Defs[sym.funcDecl.Name]] = true
		case symbolVar, symbolConst:
			for _, spec := range sym.genDecl.Specs {
				vs := spec.(*ast.ValueSpec)
				for _, name := range vs.Names {
					g.promoted[g.types.Defs[name]] = true
				}
			}
		}
	}
}

// checkPromoted rejects so:promote on an exported declaration (redundant)
// or combined with so:inline (which already emits the body in the header).
func (g *Generator) checkPromoted() {
	for _, sym := range g.symbols {
		if !sym.dirs.promote {
			continue
		}
		var node ast.Node
		switch sym.kind {
		case symbolType:
			node = sym.typeSpec.Name
		case symbolFunc, symbolMethod:
			node = sym.funcDecl.Name
		default:
			node = sym.genDecl
		}
		if sym.dirs.inline {
			g.fail(node, "so:promote cannot be combined with so:inline")
		}
		if sym.exported {
			g.fail(node, "so:promote is forbidden on an exported declaration")
		}
		// A method's C name is built from its receiver type, so promoting the
		// method without promoting the receiver would emit an unprefixed name
		// in the header (see [Generator.symbolName]).
		if sym.kind == symbolMethod {
			recv := sym.funcDecl.Recv.List[0]
			recvObj := g.recvTypeObj(recv)
			if !ast.IsExported(recvObj.Name()) && !g.promoted[recvObj] {
				g.fail(node, "so:promote method %s needs so:promote on its receiver type %s",
					sym.funcDecl.Name.Name, recvTypeName(recv))
			}
		}
	}
}

// checkExportedFuncs rejects a header function or method declaration
// that names an unexported type.
func (g *Generator) checkExportedFuncs() {
	for _, sym := range g.symbols {
		if sym.kind != symbolFunc && sym.kind != symbolMethod {
			continue
		}
		// The header holds the prototype of an exported or so:promote function,
		// and the body of an so:inline function.
		if !sym.exported && !sym.dirs.inline && !sym.dirs.promote {
			continue
		}
		kind := "function"
		if sym.kind == symbolMethod {
			kind = "method"
		}
		// The receiver becomes a void* self parameter, so only the
		// signature can name an unexported type.
		sig := g.funcSig(sym.funcDecl)
		obj := g.unexportedRefSig(sig)
		if obj == nil {
			continue
		}
		g.fail(sym.funcDecl.Name, "%s %s %s uses unexported type %s",
			headerWord(sym.exported, sym.dirs), kind, sym.funcDecl.Name.Name, obj.Name())
	}
}

// checkExportedDecls rejects a header type, var or const declaration
// that names an unexported type.
func (g *Generator) checkExportedDecls() {
	for _, sym := range g.symbols {
		if !sym.exported && !sym.dirs.promote {
			continue
		}
		switch sym.kind {
		case symbolType:
			// A constraint interface is never emitted, so it can name any type.
			if isConstraintInterface(g.types.Defs[sym.typeSpec.Name].Type()) {
				continue
			}
			// Walk the declared type rather than the underlying type:
			// the typedef for `type E P` names P itself.
			obj := g.unexportedRef(g.types.TypeOf(sym.typeSpec.Type))
			if obj == nil {
				continue
			}
			g.fail(sym.typeSpec.Name, "%s type %s uses unexported type %s",
				headerWord(sym.exported, sym.dirs), sym.typeSpec.Name.Name, obj.Name())
		case symbolVar, symbolConst:
			g.checkExportedValues(sym)
		}
	}
}

// checkExportedValues rejects a header var or const declaration
// that names an unexported type.
func (g *Generator) checkExportedValues(sym symbol) {
	kind := "variable"
	if sym.kind == symbolConst {
		kind = "constant"
	}
	for _, spec := range sym.genDecl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		// A var or const group can mix exported and unexported names,
		// so select the names the header holds (see [Generator.emitHeaderGenDecl]).
		for _, name := range vs.Names {
			exported := ast.IsExported(name.Name)
			def := g.types.Defs[name]
			if def == nil || (!exported && !sym.dirs.promote) {
				continue
			}
			obj := g.unexportedRef(def.Type())
			if obj == nil {
				continue
			}
			g.fail(name, "%s %s %s uses unexported type %s",
				headerWord(exported, sym.dirs), kind, name.Name, obj.Name())
		}
	}
}

// headerWord names the reason a declaration goes in the header.
func headerWord(exported bool, dirs directives) string {
	switch {
	case exported:
		return "exported"
	case dirs.inline:
		return "inline"
	default:
		return "promoted"
	}
}

// unexportedRef returns the first unexported type that the C declaration of typ
// names, or nil. A slice, a map and a channel map to opaque builtins that do not
// name their element type, so the walk stops there. A named type also stops the
// walk: its own declaration carries the check.
func (g *Generator) unexportedRef(typ types.Type) types.Object {
	switch t := types.Unalias(typ).(type) {
	case *types.Named:
		if g.isUnexportedType(t) {
			return t.Obj()
		}
	case *types.Pointer:
		return g.unexportedRef(t.Elem())
	case *types.Array:
		return g.unexportedRef(t.Elem())
	case *types.Struct:
		for f := range t.Fields() {
			if obj := g.unexportedRef(f.Type()); obj != nil {
				return obj
			}
		}
	case *types.Signature:
		return g.unexportedRefSig(t)
	case *types.Interface:
		for m := range t.Methods() {
			if obj := g.unexportedRef(m.Type()); obj != nil {
				return obj
			}
		}
	}
	return nil
}

// unexportedRefSig returns the first unexported type that the parameters
// or the results of sig name, or nil.
func (g *Generator) unexportedRefSig(sig *types.Signature) types.Object {
	for p := range sig.Params().Variables() {
		if obj := g.unexportedRef(p.Type()); obj != nil {
			return obj
		}
	}
	for r := range sig.Results().Variables() {
		if obj := g.unexportedRef(r.Type()); obj != nil {
			return obj
		}
	}
	return nil
}

// isUnexportedType reports whether a type lives only in the current package's
// .c file. A promoted type is emitted in the header, so it does not count.
// An extern type comes from a C header, so it does not count either.
func (g *Generator) isUnexportedType(typ types.Type) bool {
	named, ok := types.Unalias(typ).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if g.hasExtern(obj) {
		return false
	}
	if obj.Pkg() != g.pkg.Types {
		return false
	}
	return !ast.IsExported(obj.Name()) && !g.promoted[obj]
}
