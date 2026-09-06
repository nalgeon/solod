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
			// function needs its own object in the promoted set.
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

// collectImplObjs records the package-level objects
// the .c file declares with static linkage.
func (g *Generator) collectImplObjs() {
	g.implObjs = make(map[types.Object]bool)
	for _, sym := range g.symbols {
		switch sym.kind {
		case symbolType:
			// A constraint interface is not emitted at all.
			obj := g.types.Defs[sym.typeSpec.Name]
			if sym.inHeader() || isConstraintInterface(obj.Type()) {
				continue
			}
			g.implObjs[obj] = true
		case symbolFunc, symbolMethod:
			// The header holds the body of an so:inline function,
			// so an unexported one is available there too.
			if sym.inHeader() {
				continue
			}
			g.implObjs[g.types.Defs[sym.funcDecl.Name]] = true
		case symbolVar, symbolConst:
			for _, spec := range sym.genDecl.Specs {
				vs := spec.(*ast.ValueSpec)
				for _, name := range vs.Names {
					if nameInHeader(name, sym.dirs) {
						continue
					}
					g.implObjs[g.types.Defs[name]] = true
				}
			}
		}
	}
}

// implRef returns the first identifier of a header declaration
// that references an object of the .c file, or nil.
func (g *Generator) implRef(n ast.Node) *ast.Ident {
	var found *ast.Ident
	ast.Inspect(n, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		if arr, ok := node.(*ast.ArrayType); ok {
			// C gets an array length as a calculated value,
			// so the length needs no declaration.
			found = g.implRef(arr.Elt)
			return false
		}
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		if !g.implObjs[g.types.Uses[ident]] {
			return true
		}
		found = ident
		return false
	})
	return found
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
		// in the header.
		if sym.kind == symbolMethod {
			recv := sym.funcDecl.Recv.List[0]
			recvObj := g.recvTypeObj(recv)
			if !ast.IsExported(recvObj.Name()) && !g.promoted[recvObj] {
				g.fail(node, "so:promote method %s needs so:promote on its receiver type %s",
					sym.funcDecl.Name.Name, g.recvTypeName(recv))
			}
		}
	}
}

// checkExportedFuncs rejects a header function or method declaration
// that references an unexported type.
func (g *Generator) checkExportedFuncs() {
	for _, sym := range g.symbols {
		if sym.kind != symbolFunc && sym.kind != symbolMethod {
			continue
		}
		// The header holds the prototype of an exported or so:promote function,
		// and the body of an so:inline function.
		if !sym.inHeader() {
			continue
		}
		kind := "function"
		if sym.kind == symbolMethod {
			kind = "method"
		}
		// The receiver becomes a void* self parameter, so only the
		// signature might reference an unexported type.
		sig := g.funcSig(sym.funcDecl)
		if obj := g.unexportedRefSig(sig); obj != nil {
			g.fail(sym.funcDecl.Name, "%s %s %s uses unexported type %s",
				headerWord(sym.exported, sym.dirs), kind, sym.funcDecl.Name.Name, obj.Name())
		}
		// The header holds the body of an so:inline function,
		// so the body might reference an object outside the header.
		if sym.dirs.inline {
			g.checkImplRef(sym.funcDecl.Body, kind, sym.funcDecl.Name.Name, sym.exported, sym.dirs)
		}
	}
}

// checkImplRef rejects a header declaration that references
// an object which only the .c file declares.
func (g *Generator) checkImplRef(n ast.Node, kind, name string, exported bool, dirs directives) {
	ref := g.implRef(n)
	if ref == nil {
		return
	}
	obj := g.types.Uses[ref]
	g.fail(ref, "%s %s %s uses unexported %s %s",
		headerWord(exported, dirs), kind, name, objWord(obj), obj.Name())
}

// checkExportedDecls rejects a header type, var or const declaration
// that references an unexported type.
func (g *Generator) checkExportedDecls() {
	for _, sym := range g.symbols {
		if !sym.inHeader() {
			continue
		}
		switch sym.kind {
		case symbolType:
			// A constraint interface is never emitted, so it can reference any type.
			if isConstraintInterface(g.types.Defs[sym.typeSpec.Name].Type()) {
				continue
			}
			// Walk the declared type rather than the underlying type:
			// the typedef for `type E P` references P itself.
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
// that references an unexported type.
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
		for i, name := range vs.Names {
			exported := ast.IsExported(name.Name)
			def := g.types.Defs[name]
			if def == nil || !nameInHeader(name, sym.dirs) {
				continue
			}
			if obj := g.unexportedRef(def.Type()); obj != nil {
				g.fail(name, "%s %s %s uses unexported type %s",
					headerWord(exported, sym.dirs), kind, name.Name, obj.Name())
			}
			// The header holds the value of a constant, and the .c file holds
			// the value of a variable. An iota value becomes a number, so only
			// a source expression can reference an object outside the header.
			if sym.kind != symbolConst || isIotaValue(vs, i) {
				continue
			}
			g.checkImplRef(vs.Values[i], kind, name.Name, exported, sym.dirs)
		}
	}
}

// unexportedRef returns the first unexported type that the C declaration of typ
// references, or nil. A slice, a map and a channel map to opaque builtins that
// do not reference their element type, so the walk stops there. A named type
// also stops the walk: its own declaration carries the check.
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
// or the results of sig reference, or nil.
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

// objWord names the kind of a declaration.
func objWord(obj types.Object) string {
	switch o := obj.(type) {
	case *types.Const:
		return "constant"
	case *types.TypeName:
		return "type"
	case *types.Func:
		if o.Signature().Recv() != nil {
			return "method"
		}
		return "function"
	default:
		return "variable"
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
