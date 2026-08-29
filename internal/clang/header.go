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
	typeSyms := g.typeSymbols(true)
	if len(typeSyms) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Types --")
		g.emitForwardTypeDecls(w, typeSyms)
		for _, sym := range typeSyms {
			// The CommentMap might attach the doc comment to either decl
			// or type spec, depending on whether it's a standalone or
			// grouped declaration, so check both.
			hasDocs := g.emitComments(w, sym.genDecl, sym.typeSpec)
			if !hasDocs && isBlockTypeSpec(sym.typeSpec) {
				fmt.Fprintln(w)
			}
			g.emitTypeSpec(w, sym.typeSpec, sym.dirs)
		}
	}

	g.emitResultTypes(w, true)

	// Phase 2: exported const/var declarations from collected symbols.
	var varSyms []symbol
	for _, sym := range g.symbols {
		if sym.kind != symbolVar && sym.kind != symbolConst {
			continue
		}
		if !sym.exported && !sym.dirs.promote {
			continue
		}
		varSyms = append(varSyms, sym)
	}
	if len(varSyms) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Variables and constants --")
		for _, sym := range varSyms {
			g.emitHeaderGenDecl(w, sym.genDecl, sym.dirs)
		}
	}

	// Phase 3: exported function/method prototypes and inline function bodies.
	var funcSyms []symbol
	for _, sym := range g.symbols {
		if sym.kind != symbolFunc && sym.kind != symbolMethod {
			continue
		}
		if !sym.exported && !sym.dirs.inline && !sym.dirs.promote {
			continue
		}
		funcSyms = append(funcSyms, sym)
	}
	if len(funcSyms) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Functions and methods --")
		for _, sym := range funcSyms {
			if sym.dirs.inline {
				g.emitInlineFuncDecl(w, sym.funcDecl)
			} else {
				g.emitComments(w, sym.funcDecl)
				g.emitFuncProto(w, sym.funcDecl)
				fmt.Fprintln(w, ";")
			}
		}
	}
}

// emitHeaderGenDecl emits extern const/var declarations.
// Type declarations are handled separately via collected symbols.
func (g *Generator) emitHeaderGenDecl(w io.Writer, decl *ast.GenDecl, dirs directives) {
	if found, _ := parseExtern(decl.Doc); found {
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
		for i, name := range vs.Names {
			if !ast.IsExported(name.Name) && !dirs.promote {
				continue
			}
			if !emitted {
				// Emit the doc comment for the first
				// exported const/var in this declaration.
				g.emitComments(w, decl)
				emitted = true
			}
			typ := g.constType(spec, g.types.Defs[name])
			ct := g.mapTypeDecl(spec, typ)
			cName := g.symbolName(g.types.Defs[name])

			switch decl.Tok {
			case token.CONST:
				// Emit the full definition with static linkage instead of using
				// extern (declaration in .h, definition in .c):
				// 	static const int x = 42;
				//
				// It's slightly "wasteful" since the constant value is emitted
				// in every translation unit that includes the header. But it allows
				// GCC/Clang to recognize the package-level constants from 3rd-party
				// packages in definitions, which is not possible with externs:
				// 	var PointZero = Point{X: sub.Zero, Y: sub.Zero}
				isIota := isIotaValue(vs, i)
				fmt.Fprintf(w, "static const %s = ", ct.Decl(cName))
				if isIota {
					g.emitConstVal(w, vs, name)
				} else {
					g.emitExpr(w, vs.Values[i])
				}
				fmt.Fprint(w, ";\n")
			case token.VAR:
				// Build qualifier prefix for extern declarations.
				qualifier := ""
				if dirs.threadLocal {
					qualifier += "_Thread_local "
				}
				if dirs.volatile {
					qualifier += "volatile "
				}
				fmt.Fprintf(w, "extern %s%s;\n", qualifier, ct.Decl(cName))
			}
		}
	}
}

// resolveIncludePath returns the C include path for an import spec,
// or an empty string if the import should be ignored.
func (g *Generator) resolveIncludePath(spec *ast.ImportSpec) string {
	path := strings.Trim(spec.Path.Value, `"`)
	imp, ok := g.pkg.Imports[path]
	if !ok {
		g.fail(spec, "import not found: %s", path)
	}
	if imp.Module == nil {
		// Ignore all Go stdlib imports (the code might
		// only reference stdlib packages for testing).
		return ""
	}
	// Strip the imported package's own module prefix.
	path = strings.TrimPrefix(path, imp.Module.Path+"/")
	// Add the header of the package. The header is named after the package,
	// which is not always the last element of the import path
	// (e.g. sync/test -> sync/test/sync_test.h).
	parts := strings.Split(path, "/")
	parts = append(parts, imp.Name+".h")
	return strings.Join(parts, "/")
}
