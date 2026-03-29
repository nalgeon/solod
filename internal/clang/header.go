package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"slices"
	"strings"
)

// emitImports emits deduplicated, sorted #include directives for imports.
func (g *Generator) emitImports(w io.Writer) {
	seen := make(map[string]bool)
	var paths []string
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.IMPORT {
				continue
			}
			for _, spec := range gd.Specs {
				cPath := g.resolveIncludePath(spec.(*ast.ImportSpec))
				if cPath == "" || seen[cPath] {
					continue
				}
				seen[cPath] = true
				paths = append(paths, cPath)
			}
		}
	}
	slices.Sort(paths)
	for _, p := range paths {
		fmt.Fprintf(w, "#include \"%s\"\n", p)
	}
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
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Types --")
		for _, sym := range typeSyms {
			// The CommentMap might attach the doc comment to either decl
			// or type spec, depending on whether it's a standalone or
			// grouped declaration, so check both.
			hasDocs := g.emitComments(w, sym.genDecl, sym.typeSpec)
			if !hasDocs && isBlockTypeSpec(sym.typeSpec) {
				fmt.Fprintln(w)
			}
			g.emitTypeSpec(w, sym.typeSpec)
		}
	}

	// Phase 2: exported const/var declarations from collected symbols.
	var varDecls []*ast.GenDecl
	for _, sym := range g.symbols {
		if !sym.exported || (sym.kind != symbolVar && sym.kind != symbolConst) {
			continue
		}
		varDecls = append(varDecls, sym.genDecl)
	}
	if len(varDecls) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Variables and constants --")
		for _, decl := range varDecls {
			g.emitHeaderGenDecl(w, decl)
		}
	}

	// Phase 3: exported function/method prototypes from collected symbols.
	var funcSyms []symbol
	for _, sym := range g.symbols {
		if !sym.exported || (sym.kind != symbolFunc && sym.kind != symbolMethod) {
			continue
		}
		funcSyms = append(funcSyms, sym)
	}
	if len(funcSyms) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Functions and methods --")
		for _, sym := range funcSyms {
			g.emitComments(w, sym.funcDecl)
			g.emitFuncProto(w, sym.funcDecl)
			fmt.Fprintln(w, ";")
		}
	}
}

// emitHeaderGenDecl emits extern const/var declarations.
// Type declarations are handled separately via collected symbols.
func (g *Generator) emitHeaderGenDecl(w io.Writer, decl *ast.GenDecl) {
	if found, _ := parseExternDirective(decl.Doc); found {
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
				g.emitComments(w, decl)
				emitted = true
			}
			typ := g.types.Defs[name].Type()
			ct := g.mapCType(spec, typ)
			cName := g.symbolName(name.Name)
			switch decl.Tok {
			case token.CONST:
				fmt.Fprintf(w, "extern const %s;\n", ct.Decl(cName))
			case token.VAR:
				fmt.Fprintf(w, "extern %s;\n", ct.Decl(cName))
			}
		}
	}
}

// resolveIncludePath returns the C include path for an import spec,
// or an empty string if the import should be ignored.
func (g *Generator) resolveIncludePath(spec *ast.ImportSpec) string {
	path := strings.Trim(spec.Path.Value, `"`)
	if isIgnoredPackage(path) {
		return ""
	}
	// Strip the imported package's own module prefix.
	if imp, ok := g.pkg.Imports[path]; ok && imp.Module != nil {
		path = strings.TrimPrefix(path, imp.Module.Path+"/")
	}
	// Add the package.h file (e.g. package -> package/package.h).
	parts := strings.Split(path, "/")
	parts = append(parts, parts[len(parts)-1]+".h")
	return strings.Join(parts, "/")
}

// isIgnoredPackage returns true if the import path is for
// a package that should not be emitted as a #include directive.
func isIgnoredPackage(path string) bool {
	// embed is only a marker import for embedding files,
	// unsafe is implemented in builtin.h,
	// so they neither requires a #include directive.
	// Other ignored packages are only used for extern functions in tests.
	return path == "embed" || path == "fmt" || path == "math" || path == "unsafe"
}
