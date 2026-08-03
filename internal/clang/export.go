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

// checkExported rejects exported functions and methods that use
// unexported types, which would leak a .c-only type into the header.
func (g *Generator) checkExported() {
	for _, sym := range g.symbols {
		if (sym.kind != symbolFunc && sym.kind != symbolMethod) || !sym.exported {
			continue
		}
		decl := sym.funcDecl
		if g.hasUnexportedTypes(decl) {
			g.fail(decl.Name, "exported function %s uses unexported types", decl.Name.Name)
		}
	}
}

// hasUnexportedTypes reports whether a function declaration
// references any unexported types from the current package.
func (g *Generator) hasUnexportedTypes(decl *ast.FuncDecl) bool {
	sig := g.funcSig(decl)
	for p := range sig.Params().Variables() {
		if g.isUnexportedType(p.Type()) {
			return true
		}
	}
	for r := range sig.Results().Variables() {
		if g.isUnexportedType(r.Type()) {
			return true
		}
	}
	return false
}

// isUnexportedType reports whether a type lives only in the current package's
// .c file. A promoted type is emitted in the header, so it does not count.
func (g *Generator) isUnexportedType(typ types.Type) bool {
	named, ok := types.Unalias(typ).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() != g.pkg.Types {
		return false
	}
	return !ast.IsExported(obj.Name()) && !g.promoted[obj]
}
