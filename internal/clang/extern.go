package clang

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// externInfo holds metadata parsed from a so:extern directive.
type externInfo struct {
	name    string // C name override (empty = use default)
	nodecay bool   // skip decay for call args
}

// collectImportExterns gathers extern symbols from all imported packages,
// both direct and indirect.
//
// An indirect import can provide a type name to the generated C code:
// for example, package A calls a function in B, and that function's signature
// uses an extern type from C (like c.Int), but A doesn't import C. Without
// the extern name for that type, the generator writes the package-qualified
// name (c_Int), which isn't declared in any C header.
//
//	func Calc(n c.Int) // in package B
//	const n = 21       // in package A
//	Calc(n)            // n is implicitly cast to c.Int
func (g *Generator) collectImportExterns() {
	var collect func(pkg *packages.Package)
	collect = func(pkg *packages.Package) {
		for path, imp := range pkg.Imports {
			// Skip the Go stdlib and the already-seen packages.
			if g.modulePkgs[path] || imp.Module == nil {
				continue
			}
			g.modulePkgs[path] = true
			for _, file := range imp.Syntax {
				g.collectFileExterns(imp.TypesInfo, file)
			}
			collect(imp)
		}
	}
	collect(g.pkg)
}

// collectFileExterns collects extern symbols from a single file's declarations.
func (g *Generator) collectFileExterns(typesInfo *types.Info, file *ast.File) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			found, info := parseExtern(d.Doc)
			if !found {
				continue
			}
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					g.markExtern(typesInfo.Defs[s.Name], info)
					g.markExternFields(typesInfo, s, info)
				case *ast.ValueSpec:
					for _, name := range s.Names {
						g.markExtern(typesInfo.Defs[name], info)
					}
				}
			}
		case *ast.FuncDecl:
			found, info := parseExtern(d.Doc)
			if d.Body == nil || found {
				g.markExtern(typesInfo.Defs[d.Name], info)
			}
		}
	}
}

// parseExtern checks if a comment group contains the so:extern
// directive and parses its options (name override and nodecay flag).
func parseExtern(doc *ast.CommentGroup) (bool, externInfo) {
	if doc == nil {
		return false, externInfo{}
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(c.Text)
		rest, ok := strings.CutPrefix(text, "//so:extern")
		if !ok {
			continue
		}
		var info externInfo
		fields := strings.Fields(rest)
		if len(fields) > 0 && fields[len(fields)-1] == "nodecay" {
			// nodecay can only be the last field.
			info.nodecay = true
			fields = fields[:len(fields)-1]
		}
		if len(fields) > 0 {
			// Use the remaining fields as the C name override.
			info.name = strings.Join(fields, " ")
		}
		return true, info
	}
	return false, externInfo{}
}

// funcExtern returns the extern metadata for a function call
// if the function is marked as extern.
func (g *Generator) funcExtern(call *ast.CallExpr) (externInfo, bool) {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Local package call.
		return g.getExtern(g.types.Uses[fun])
	case *ast.SelectorExpr:
		// Package-qualified call (e.g. stdio.Printf).
		if ident, ok := fun.X.(*ast.Ident); ok {
			if _, ok := g.types.Uses[ident].(*types.PkgName); ok {
				return g.getExtern(g.types.Uses[fun.Sel])
			}
		}
		// Function pointer field on an extern struct (e.g. acc.write(...)).
		return g.callExternField(fun)
	}
	return externInfo{}, false
}

// methodExtern returns the extern metadata for a method-value selector
// if the method is marked as extern.
func (g *Generator) methodExtern(sel *ast.SelectorExpr) (externInfo, bool) {
	selection, ok := g.types.Selections[sel]
	if !ok || selection.Kind() != types.MethodVal {
		return externInfo{}, false
	}
	return g.getExtern(selection.Obj())
}

// callExternField checks whether a selector targets a function pointer field
// on an extern struct (e.g. acc.write).
func (g *Generator) callExternField(sel *ast.SelectorExpr) (externInfo, bool) {
	selection, ok := g.types.Selections[sel]
	if !ok || selection.Kind() != types.FieldVal {
		return externInfo{}, false
	}
	info, ok := g.externs[selection.Obj()]
	return info, ok
}

// markExternFields registers function pointer fields of an extern struct type,
// so that calls like acc.write(...) can be resolved via a map lookup.
func (g *Generator) markExternFields(typesInfo *types.Info, spec *ast.TypeSpec, info externInfo) {
	obj := typesInfo.Defs[spec.Name]
	if obj == nil {
		return
	}
	st, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}
	fieldInfo := externInfo{nodecay: info.nodecay}
	for field := range st.Fields() {
		if _, ok := field.Type().Underlying().(*types.Signature); ok {
			g.externs[field] = fieldInfo
		}
	}
}

// markExtern marks a types.Object as extern.
func (g *Generator) markExtern(obj types.Object, info externInfo) {
	g.externs[obj] = info
}

// hasExtern reports whether a types.Object is marked as extern.
func (g *Generator) hasExtern(obj types.Object) bool {
	_, ok := g.externs[obj]
	return ok
}

// readsExtern reports whether an expression reads an extern symbol.
func (g *Generator) readsExtern(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if ident, ok := n.(*ast.Ident); ok && g.hasExtern(g.types.Uses[ident]) {
			found = true
		}
		return !found
	})
	return found
}

// getExtern returns the extern metadata for a types.Object.
func (g *Generator) getExtern(obj types.Object) (externInfo, bool) {
	info, ok := g.externs[obj]
	return info, ok
}

// cIntrinsic checks whether an expression is a c.Raw or c.Val call
// and returns the raw string content. The argument must be a string literal.
func (g *Generator) cIntrinsic(expr ast.Expr) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	// Unwrap IndexExpr for generic calls like c.Val[T]("...").
	fun := call.Fun
	if idx, ok := fun.(*ast.IndexExpr); ok {
		fun = idx.X
	}
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || (sel.Sel.Name != "Raw" && sel.Sel.Name != "Val") {
		return "", false
	}
	obj := g.types.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "solod.dev/so/c" {
		return "", false
	}
	if len(call.Args) != 1 {
		g.fail(call, "Raw C call requires exactly one argument")
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		g.fail(call, "Raw C call argument must be a string literal")
	}

	// Extract raw content: strip backticks for raw strings, unquote interpreted strings.
	var s string
	if strings.HasPrefix(lit.Value, "`") {
		s = lit.Value[1 : len(lit.Value)-1]
	} else {
		var err error
		s, err = strconv.Unquote(lit.Value)
		if err != nil {
			g.fail(lit, "Raw C call: invalid string literal")
		}
	}
	return dedent(s), true
}

// dedent removes common leading whitespace from a multi-line string.
// Also trims leading and trailing blank lines.
func dedent(s string) string {
	lines := strings.Split(s, "\n")

	// Trim leading and trailing blank lines.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}

	// Find the shortest whitespace prefix among non-empty lines.
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return strings.Join(lines, "\n")
	}

	// Strip the common prefix.
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}
