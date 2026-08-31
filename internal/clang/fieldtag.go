package clang

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
)

// collectFieldTags records the C field name overrides declared with `c:"..."`
// struct tags. Overrides are honored only on the fields of a named struct type
// declaration; a tag on an anonymous or function-local struct is ignored.
func (g *Generator) collectFieldTags() {
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if st, ok := ts.Type.(*ast.StructType); ok {
					g.collectStructFieldTags(st)
				}
			}
		}
	}
	// Imports were validated when they were themselves compiled, so read their
	// overrides straight from type info.
	for _, imp := range g.pkg.Imports {
		g.collectImportFieldTags(imp.Types)
	}
}

// collectStructFieldTags validates and records the C field name overrides from
// the `c` tags on the fields of a named struct in the current package.
func (g *Generator) collectStructFieldTags(st *ast.StructType) {
	for _, field := range st.Fields.List {
		override := cFieldTag(field.Tag)
		if override == "" {
			continue
		}
		if len(field.Names) != 1 {
			g.fail(field, "c field tag requires exactly one field name")
		}
		if !isCIdent(override) || reservedC[override] {
			g.fail(field, "invalid C field name %q in c tag", override)
		}
		if v, ok := g.types.Defs[field.Names[0]].(*types.Var); ok {
			g.fieldNames[v] = override
		}
	}
}

// collectImportFieldTags records the C field name overrides from the named
// struct types of an imported package, read from type info.
func (g *Generator) collectImportFieldTags(pkg *types.Package) {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		st, ok := tn.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}
		for i := range st.NumFields() {
			if override := reflect.StructTag(st.Tag(i)).Get("c"); override != "" {
				g.fieldNames[st.Field(i).Origin()] = override
			}
		}
	}
}

// cFieldTag returns the value of the `c` key in a struct tag, or "" if absent.
func cFieldTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	value, err := strconv.Unquote(tag.Value)
	if err != nil {
		return ""
	}
	return reflect.StructTag(value).Get("c")
}

// isCIdent reports whether s is a valid C identifier.
func isCIdent(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		letter := c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		digit := c >= '0' && c <= '9'
		if !letter && !(i > 0 && digit) {
			return false
		}
	}
	return true
}
