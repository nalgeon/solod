package clang

import (
	"go/ast"
	"go/token"
	"maps"
	"path/filepath"
)

type symbolKind int

const (
	symbolFunc symbolKind = iota
	symbolMethod
	symbolType
	symbolVar
	symbolConst
)

type symbol struct {
	kind     symbolKind
	dirs     directives   // parsed so: directives
	genDecl  *ast.GenDecl // parent GenDecl (for type symbols, enables comment lookup)
	typeSpec *ast.TypeSpec
	funcDecl *ast.FuncDecl

	// var and const groups can contain both exported and unexported names,
	// so they can have both exported == true and unexported == true.
	// Other symbol kinds are either exported or unexported, never both.
	exported   bool
	unexported bool
}

// collect performs a single pass over all package files, collecting:
// - Comment map (for doc comment emission)
// - Include directives (so:include, so:include.c)
// - Link directives (so:link)
// - Embed directives (so:embed) with file reads
// - Extern symbols (so:extern, body-less functions)
// - Promoted unexported symbols (so:promote) to be emitted in the header.
// - Symbol list (types, vars, consts, functions) with parsed directives
func (g *Generator) collect() {
	g.comments = ast.CommentMap{}
	g.embeds = newEmbeds()

	srcDir := ""
	if len(g.pkg.GoFiles) > 0 {
		srcDir = filepath.Dir(g.pkg.GoFiles[0])
	}

	for _, file := range g.pkg.Syntax {
		g.checkImports(file)

		// Merge file comments into the global comment map.
		fileComments := ast.NewCommentMap(g.pkg.Fset, file, file.Comments)
		maps.Copy(g.comments, fileComments)

		// Collect include and link directives from file-level comments.
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if path, ok := g.parseDirective(c, "//so:include.c"); ok {
					g.includes.impl = append(g.includes.impl, path)
				} else if path, ok := g.parseDirective(c, "//so:include"); ok {
					g.includes.header = append(g.includes.header, path)
				} else if lib, ok := g.parseDirective(c, "//so:link"); ok {
					g.links = append(g.links, lib)
				}
			}
		}

		// Collect extern symbols and build the symbol list.
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				g.collectGenDecl(srcDir, d)
			case *ast.FuncDecl:
				g.collectFuncDecl(d)
			}
		}
	}

	g.collectImportExterns()
	g.collectFieldTags()
	g.collectPromoted()
	g.collectImplObjs()
	g.collectResultTypes()

	g.resolveReservedNames()
	g.checkExportedFuncs()
	g.checkExportedDecls()
	g.checkPromoted()
	g.checkEmbeddedTypes()
	g.checkFieldNames()
	g.checkAnonStructAliases()
	g.checkValueOrder()
	g.checkFrameValues()
}

// checkImports rejects a dot import.
func (g *Generator) checkImports(file *ast.File) {
	for _, spec := range file.Imports {
		if spec.Name != nil && spec.Name.Name == "." {
			g.fail(spec, "dot import is not supported")
		}
	}
}

// collectGenDecl processes a GenDecl for externs, embeds, and symbol collection.
func (g *Generator) collectGenDecl(srcDir string, d *ast.GenDecl) {
	// Handle so:extern declarations.
	foundExtern, externInf := parseExtern(d.Doc)
	if foundExtern {
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				g.markExtern(g.types.Defs[s.Name], externInf)
				g.markExternFields(g.types, s, externInf)
			case *ast.ValueSpec:
				for _, name := range s.Names {
					g.markExtern(g.types.Defs[name], externInf)
				}
			}
		}
		return
	}

	// Handle so:embed on variable declarations.
	if d.Tok == token.VAR {
		if filename, ok := parseEmbed(d.Doc); ok {
			ef, err := loadEmbedFile(srcDir, filename)
			if err != nil {
				g.fail(d, "error reading file %s: %v", filename, err)
			}
			g.embeds.addFile(ef)
			for _, spec := range d.Specs {
				vs := spec.(*ast.ValueSpec)
				for _, name := range vs.Names {
					g.embeds.vars[name.Name] = true
				}
			}
			return
		}
	}

	// Parse directives for non-extern, non-embed GenDecls.
	dirs := parseDirectives(d.Doc)

	// Validate directive/declaration-kind compatibility.
	if dirs.inline {
		g.fail(d, "so:inline is only allowed on functions")
	}
	switch d.Tok {
	case token.TYPE:
		if dirs.volatile {
			g.fail(d, "so:volatile is not allowed on type declarations")
		}
		if dirs.threadLocal {
			g.fail(d, "so:thread_local is not allowed on type declarations")
		}
		for _, spec := range d.Specs {
			ts := spec.(*ast.TypeSpec)
			if g.hasExtern(g.types.Defs[ts.Name]) {
				continue
			}
			exported := ast.IsExported(ts.Name.Name)
			g.symbols = append(g.symbols, symbol{
				kind:       symbolType,
				exported:   exported,
				unexported: !exported,
				dirs:       dirs,
				genDecl:    d,
				typeSpec:   ts,
			})
		}
	case token.VAR:
		exported, unexported := detectExported(d)
		g.symbols = append(g.symbols, symbol{
			kind:       symbolVar,
			exported:   exported,
			unexported: unexported,
			dirs:       dirs,
			genDecl:    d,
		})
	case token.CONST:
		if dirs.volatile {
			g.fail(d, "so:volatile is not allowed on const declarations")
		}
		if dirs.threadLocal {
			g.fail(d, "so:thread_local is not allowed on const declarations")
		}
		exported, unexported := detectExported(d)
		g.symbols = append(g.symbols, symbol{
			kind:       symbolConst,
			exported:   exported,
			unexported: unexported,
			dirs:       dirs,
			genDecl:    d,
		})
	}
}

// collectFuncDecl processes a FuncDecl for externs and symbol collection.
func (g *Generator) collectFuncDecl(d *ast.FuncDecl) {
	// Handle extern functions (body-less or so:extern).
	foundExtern, externInf := parseExtern(d.Doc)
	if d.Body == nil || foundExtern {
		g.markExtern(g.types.Defs[d.Name], externInf)
		return
	}

	if isMainFunc(d) {
		return
	}
	if isInitFunc(d) {
		if g.initFunc != nil {
			g.fail(d.Name, "multiple init functions in package %s", g.pkg.Name)
		}
		g.initFunc = d
		return
	}
	if g.hasExtern(g.types.Defs[d.Name]) {
		return
	}

	dirs := parseDirectives(d.Doc)

	// Validate directive/declaration-kind compatibility.
	if dirs.volatile {
		g.fail(d, "so:volatile is not allowed on functions")
	}
	if dirs.threadLocal {
		g.fail(d, "so:thread_local is not allowed on functions")
	}

	g.funcDirs[d] = dirs

	kind := symbolFunc
	exported := ast.IsExported(d.Name.Name)
	if d.Recv != nil {
		kind = symbolMethod
		if exported {
			exported = ast.IsExported(g.recvTypeName(d.Recv.List[0]))
		}
	}
	g.symbols = append(g.symbols, symbol{
		kind:       kind,
		exported:   exported,
		unexported: !exported,
		dirs:       dirs,
		funcDecl:   d,
	})
}

// checkEmbeddedTypes rejects embedded fields in structs
// and embedded interfaces in interfaces.
func (g *Generator) checkEmbeddedTypes() {
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.StructType:
				g.checkEmbeddedFields(t)
			case *ast.InterfaceType:
				g.checkEmbeddedIfaces(t)
			}
			return true
		})
	}
}

// typeSymbols returns the type symbols emitted in the header (header == true)
// or in the .c file (header == false), ordered so that each type follows the
// types it depends on.
func (g *Generator) typeSymbols(header bool) []symbol {
	var syms []symbol
	inHeader := func(sym symbol) bool {
		return sym.exported || (sym.unexported && sym.dirs.promote)
	}
	for _, sym := range g.symbols {
		if sym.kind != symbolType || inHeader(sym) != header {
			continue
		}
		if isConstraintInterface(g.types.Defs[sym.typeSpec.Name].Type()) {
			continue
		}
		syms = append(syms, sym)
	}
	return g.sortTypes(syms)
}

// detectExported reports whether a GenDecl contains at least one
// exported and at least one unexported name in its value specs.
func detectExported(d *ast.GenDecl) (exported bool, unexported bool) {
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if ast.IsExported(name.Name) {
				exported = true
			} else {
				unexported = true
			}
		}
	}
	return exported, unexported
}
