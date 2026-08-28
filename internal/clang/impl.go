package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
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
			cName := g.declSymbolName(g.types.Defs[sym.typeSpec.Name])
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

// checkStaticInit fails when C cannot calculate the initializer of a
// package-level variable (C requires a constant expression at file scope).
func (g *Generator) checkStaticInit(expr ast.Expr) {
	if g.types.Types[expr].Value != nil {
		// A constant expression.
		return
	}

	switch e := expr.(type) {
	case *ast.ParenExpr:
		g.checkStaticInit(e.X)

	case *ast.Ident:
		// C reads a variable only at runtime.
		if _, ok := g.types.Uses[e].(*types.Var); ok {
			g.fail(e, "cannot read variable %s in a package-level initializer; use init()", e.Name)
		}

	case *ast.SelectorExpr:
		// A variable from another package.
		if ident, ok := e.X.(*ast.Ident); ok {
			if _, isPkg := g.types.Uses[ident].(*types.PkgName); isPkg {
				g.checkStaticInit(e.Sel)
				return
			}
		}
		// A struct field.
		g.checkStaticInit(e.X)

	case *ast.UnaryExpr:
		// C calculates the address of a package-level variable
		// and its fields at compile time.
		if e.Op == token.AND && isNameChain(e.X) {
			return
		}
		g.checkStaticInit(e.X)

	case *ast.BinaryExpr:
		g.checkStaticInit(e.X)
		g.checkStaticInit(e.Y)

	case *ast.CompositeLit:
		g.checkStaticLit(e)

	case *ast.CallExpr:
		g.checkStaticCall(e)

	case *ast.StarExpr:
		g.checkStaticInit(e.X)

	case *ast.IndexExpr:
		g.checkStaticInit(e.X)
		g.checkStaticInit(e.Index)
	}
}

// isNameChain reports whether an expression is a name, or a field selected
// from a name. Both "p" and "p.x.y" are name chains.
func isNameChain(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return isNameChain(e.X)
	}
	return false
}

// checkStaticLit checks the elements of a composite literal in the
// initializer of a package-level variable.
func (g *Generator) checkStaticLit(lit *ast.CompositeLit) {
	// A map literal expands to a statement expression,
	// which C does not allow at file scope.
	if _, ok := g.types.TypeOf(lit).Underlying().(*types.Map); ok {
		g.fail(lit, "cannot use a map literal in a package-level initializer")
	}
	for _, elt := range lit.Elts {
		// For a struct field, check the value.
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			g.checkStaticInit(kv.Value)
			continue
		}
		g.checkStaticInit(elt)
	}
}

// checkStaticCall checks a call in the initializer of a package-level
// variable. A conversion is a call node too.
func (g *Generator) checkStaticCall(call *ast.CallExpr) {
	// Function call.
	if !g.types.Types[call.Fun].IsType() {
		// An extern function can be a C macro that expands to a constant,
		// so leave the check to the C compiler. Fail for any other function.
		if _, ok := g.funcExtern(call); !ok {
			g.fail(call, "cannot call a function in a package-level initializer; use init()")
		}
		for _, arg := range call.Args {
			g.checkStaticInit(arg)
		}
		return
	}

	// A conversion between a string and a slice expands to a statement
	// expression, which C does not allow at file scope.
	target := g.types.TypeOf(call).Underlying()
	arg := g.types.TypeOf(call.Args[0]).Underlying()
	_, toSlice := target.(*types.Slice)
	_, fromSlice := arg.(*types.Slice)
	if (toSlice && isStringType(arg)) || (fromSlice && isStringType(target)) {
		g.fail(call, "cannot convert string<->slice in a package-level initializer")
	}
	g.checkStaticInit(call.Args[0])
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
