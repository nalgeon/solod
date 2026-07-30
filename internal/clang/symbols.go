package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
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
	g.embeds = Embeds{vars: make(map[string]bool)}
	g.handleReservedNames()

	srcDir := ""
	if len(g.pkg.GoFiles) > 0 {
		srcDir = filepath.Dir(g.pkg.GoFiles[0])
	}

	for _, file := range g.pkg.Syntax {
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

	// Collect externs from imported packages so that funcExtern/methodExtern
	// can identify cross-package extern calls (e.g. stdio.Printf).
	for _, imp := range g.pkg.Imports {
		for _, file := range imp.Syntax {
			g.collectFileExterns(imp.TypesInfo, file)
		}
	}

	g.collectFieldTags()
	g.collectPromoted()
	g.collectResultTypes()

	g.validateExported()
	g.validatePromoted()
	g.rejectEmbedding()
}

// validateExported rejects exported functions and methods that use
// unexported types, which would leak a .c-only type into the header.
func (g *Generator) validateExported() {
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

// validatePromoted rejects so:promote on an exported declaration (redundant)
// or combined with so:inline (which already emits the body in the header).
func (g *Generator) validatePromoted() {
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

// rejectEmbedding rejects embedded fields in structs
// and embedded interfaces in interfaces.
func (g *Generator) rejectEmbedding() {
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.StructType:
				g.rejectEmbeddedFields(t)
			case *ast.InterfaceType:
				g.rejectEmbeddedIfaces(t)
			}
			return true
		})
	}
}

// rejectEmbeddedFields rejects struct fields declared without a name.
func (g *Generator) rejectEmbeddedFields(st *ast.StructType) {
	for _, field := range st.Fields.List {
		if len(field.Names) > 0 {
			continue
		}
		typ := g.typeString(g.types.TypeOf(field.Type))
		g.fail(field, "embedded field %s is not supported; declare a named field instead", typ)
	}
}

// rejectEmbeddedIfaces rejects an interface embedded in another interface.
func (g *Generator) rejectEmbeddedIfaces(it *ast.InterfaceType) {
	if isConstraintInterface(g.types.TypeOf(it)) {
		// A constraint interface never reaches C,
		// so what it embeds does not matter.
		return
	}
	for _, elem := range it.Methods.List {
		if _, isMethod := elem.Type.(*ast.FuncType); isMethod {
			continue
		}
		typ := g.types.TypeOf(elem.Type)
		if typ == nil {
			continue
		}
		iface, ok := typ.Underlying().(*types.Interface)
		if !ok || iface.NumMethods() == 0 {
			continue
		}
		g.fail(elem, "embedded interface %s is not supported; declare its methods instead", g.typeString(typ))
	}
}

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
		if filename, ok := embedDirective(d.Doc); ok {
			content, err := os.ReadFile(filepath.Join(srcDir, filename))
			if err != nil {
				g.fail(d, "error reading file %s: %v", filename, err)
			}
			ef := embedFile{name: filename, content: string(content)}
			switch filepath.Ext(filename) {
			case ".h":
				g.embeds.header = append(g.embeds.header, ef)
			case ".c":
				g.embeds.impl = append(g.embeds.impl, ef)
			}
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
			exported = ast.IsExported(recvTypeName(d.Recv.List[0]))
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

// parseDirective matches a file-level directive comment against prefix and
// returns its trimmed argument. ok reports whether the prefix matched. It
// fails when the prefix matches but no argument follows.
func (g *Generator) parseDirective(c *ast.Comment, prefix string) (arg string, ok bool) {
	arg, ok = strings.CutPrefix(c.Text, prefix)
	if !ok {
		return "", false
	}
	arg = strings.TrimSpace(arg)
	if arg == "" {
		g.fail(c, "%s requires an argument", strings.TrimPrefix(prefix, "//"))
	}
	return arg, true
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
