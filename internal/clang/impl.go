package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
)

// emitPackageVars writes all package-level variable and constant
// declarations at the top of the .c file, after forward func declarations
// (so initializers can reference functions) but before any function body
// (so functions can reference these vars).
func (g *Generator) emitPackageVars(w io.Writer) {
	var symbols []symbol
	for _, sym := range g.symbols {
		if sym.kind != symbolVar && sym.kind != symbolConst {
			continue
		}
		if sym.kind == symbolConst && (!sym.unexported || sym.dirs.promote) {
			// All constants in the group are exported, or the group is
			// so:promote. Either way it is emitted in the header, not here.
			continue
		}
		symbols = append(symbols, sym)
	}
	if len(symbols) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// -- Variables and constants --")
	for _, sym := range symbols {
		g.emitComments(w, sym.genDecl)
		switch sym.genDecl.Tok {
		case token.CONST:
			for _, spec := range sym.genDecl.Specs {
				g.emitConstSpec(w, spec.(*ast.ValueSpec))
			}
		case token.VAR:
			for _, spec := range sym.genDecl.Specs {
				vs := spec.(*ast.ValueSpec)
				if len(vs.Names) > 0 && g.embeds.vars[vs.Names[0].Name] {
					continue
				}
				g.emitVarSpec(w, vs, sym.dirs)
			}
		}
	}
}

// emitUnexportedTypes writes full type definitions for unexported types that
// stay in the .c file, i.e. those not marked with so:promote.
// Emitted before package vars so that compound literals can reference them.
func (g *Generator) emitUnexportedTypes(w io.Writer) {
	typeSyms := g.typeSymbols(false)
	if len(typeSyms) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// -- Types --")
	g.emitForwardTypeDecls(w, typeSyms)
	for _, sym := range typeSyms {
		hasDocs := g.emitComments(w, sym.genDecl, sym.typeSpec)
		if !hasDocs && isBlockTypeSpec(sym.typeSpec) {
			fmt.Fprintln(w)
		}
		g.emitTypeSpec(w, sym.typeSpec, sym.dirs)
	}
}

// emitForwardTypeDecls writes forward declarations for struct types
// so that self-referencing and out-of-order references resolve.
func (g *Generator) emitForwardTypeDecls(w io.Writer, typeSyms []symbol) {
	hasDecls := false
	for _, sym := range typeSyms {
		if _, ok := sym.typeSpec.Type.(*ast.StructType); ok {
			cName := g.mapObjName(g.types.Defs[sym.typeSpec.Name])
			fmt.Fprintf(w, "\ntypedef struct %s %s;", cName, cName)
			hasDecls = true
		}
	}
	if hasDecls {
		fmt.Fprintln(w)
	}
}

// emitForwardFuncDecls writes forward declarations for unexported functions
// and methods so that they can be called before their definition.
func (g *Generator) emitForwardFuncDecls(w io.Writer) {
	var funcDecls []*ast.FuncDecl
	for _, sym := range g.symbols {
		if sym.kind != symbolFunc && sym.kind != symbolMethod {
			continue
		}
		if sym.exported || sym.dirs.inline || sym.dirs.promote {
			continue
		}
		funcDecls = append(funcDecls, sym.funcDecl)
	}
	if len(funcDecls) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// -- Forward declarations --")
	for _, decl := range funcDecls {
		g.emitFuncProto(w, decl)
		fmt.Fprintln(w, ";")
	}
}

// isBlockTypeSpec returns true for type specs that emit multi-line blocks
// (structs, non-empty interfaces, func types) and need a blank line separator.
func isBlockTypeSpec(spec *ast.TypeSpec) bool {
	switch typ := spec.Type.(type) {
	case *ast.StructType, *ast.FuncType:
		return true
	case *ast.InterfaceType:
		// Non-empty interfaces are block types; empty ones are single-line typedefs.
		return len(typ.Methods.List) > 0
	}
	return false
}
