package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// emitImports emits #include directives for imports.
func (g *Generator) emitImports(w io.Writer) {
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.IMPORT {
				continue
			}
			for _, spec := range gd.Specs {
				g.emitImportSpec(w, spec.(*ast.ImportSpec))
			}
		}
	}
	fmt.Fprintln(w)
}

// emitImportSpec emits a #include directive for an import.
func (g *Generator) emitImportSpec(w io.Writer, spec *ast.ImportSpec) {
	path := strings.Trim(spec.Path.Value, `"`)
	if path == "embed" {
		// It's only a marker import for embedding files, not an actual dependency.
		return
	}
	// Strip the imported package's own module prefix.
	if imp, ok := g.pkg.Imports[path]; ok && imp.Module != nil {
		path = strings.TrimPrefix(path, imp.Module.Path+"/")
	}
	// Add the package.h file (e.g. package -> package/package.h).
	parts := strings.Split(path, "/")
	parts = append(parts, parts[len(parts)-1]+".h")
	cPath := strings.Join(parts, "/")
	// Emit the #include directive (e.g. #include "package/package.h").
	fmt.Fprintf(w, "#include \"%s\"\n", cPath)
}

// emitHeaderDecls writes declarations for exported package-level symbols.
// Types are emitted first so that const/var and function prototypes
// can reference them.
func (g *Generator) emitHeaderDecls(w io.Writer) {
	// Phase 1: exported types from collected symbols.
	var typeSyms []symbol
	for _, sym := range g.symbols {
		if !sym.exported || sym.kind != symbolType {
			continue
		}
		typeSyms = append(typeSyms, sym)
	}
	if len(typeSyms) > 0 {
		fmt.Fprintln(w, "// -- Types --")
		for _, sym := range typeSyms {
			g.emitTypeSpec(w, sym.typeSpec, sym.doc)
		}
		fmt.Fprintln(w)
	}

	// Phase 2: const/var declarations from the AST.
	var genDecls []*ast.GenDecl
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok {
				if gd.Tok != token.TYPE {
					// Types are already handled above.
					genDecls = append(genDecls, gd)
				}
			}
		}
	}
	if len(genDecls) > 0 {
		fmt.Fprintln(w, "// -- Variables and constants --")
		for _, decl := range genDecls {
			g.emitHeaderGenDecl(w, decl)
		}
		fmt.Fprintln(w)
	}

	// Phase 3: exported function/method prototypes from collected symbols.
	var funcSyms []symbol
	for _, sym := range g.symbols {
		if !sym.exported || sym.kind == symbolType {
			continue
		}
		funcSyms = append(funcSyms, sym)
	}
	if len(funcSyms) > 0 {
		fmt.Fprintln(w, "// -- Functions and methods --")
		for _, sym := range funcSyms {
			emitDocComment(w, sym.doc)
			fn := newFuncDecl(g, sym.funcDecl)
			fmt.Fprintf(w, "%s %s(%s);\n", fn.returnType(), fn.name(), fn.params())
		}
		fmt.Fprintln(w)
	}
}

// emitHeaderGenDecl emits extern const/var declarations.
// Type declarations are handled separately via collected symbols.
func (g *Generator) emitHeaderGenDecl(w io.Writer, decl *ast.GenDecl) {
	if hasExternDirective(decl.Doc) {
		return
	}
	if decl.Tok == token.TYPE {
		// Types are handled separately in [Generator.emitHeaderDecls].
		return
	}

	// Variable and constant declarations.
	emitted := false
	for _, spec := range decl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if !ast.IsExported(name.Name) {
				continue
			}
			if !emitted {
				// Emit the doc comment for the first
				// exported const/var in this declaration.
				emitDocComment(w, decl.Doc)
				emitted = true
			}
			typ := g.types.Defs[name].Type()
			cType := g.mapType(spec, typ)
			cName := g.symbolName(name.Name)
			switch decl.Tok {
			case token.CONST:
				fmt.Fprintf(w, "extern const %s %s;\n", cType, cName)
			case token.VAR:
				fmt.Fprintf(w, "extern %s %s;\n", cType, cName)
			}
		}
	}
}
