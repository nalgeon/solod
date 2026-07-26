package clang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestIsCIdent(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"type", true},
		{"etype", true},
		{"_x", true},
		{"x1", true},
		{"X_Y_9", true},
		{"", false},
		{"1x", false},   // leading digit
		{"a b", false},  // space
		{"a-b", false},  // hyphen
		{"a.b", false},  // dot
		{"ünïx", false}, // non-ASCII
	}
	for _, tt := range tests {
		if got := isCIdent(tt.in); got != tt.want {
			t.Errorf("isCIdent(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestFieldTagValidation(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string // substring of the expected failure, "" means must succeed
	}{
		{
			name: "valid override",
			src:  "package x\ntype T struct{ etype uint32 `c:\"type\"` }",
			want: "",
		},
		{
			name: "invalid C identifier",
			src:  "package x\ntype T struct{ x int `c:\"1bad\"` }",
			want: "invalid C field name",
		},
		{
			name: "reserved C word",
			src:  "package x\ntype T struct{ x int `c:\"int\"` }",
			want: "invalid C field name",
		},
		{
			name: "override collides with plain field",
			src:  "package x\ntype T struct{ a int `c:\"b\"`; b int }",
			want: "duplicate C field name",
		},
		{
			name: "two overrides collide",
			src:  "package x\ntype T struct{ a int `c:\"x\"`; b int `c:\"x\"` }",
			want: "duplicate C field name",
		},
		{
			name: "grouped names with override",
			src:  "package x\ntype T struct{ a, b int `c:\"x\"` }",
			want: "exactly one field name",
		},
		{
			name: "unrelated tag ignored",
			src:  "package x\ntype T struct{ a int `json:\"a\"` }",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectFieldTagsFail(t, tt.src)
			switch {
			case tt.want == "" && got != "":
				t.Errorf("unexpected failure: %s", got)
			case tt.want != "" && !strings.Contains(got, tt.want):
				t.Errorf("failure = %q, want substring %q", got, tt.want)
			}
		})
	}
}

// collectFieldTagsFail runs collectFieldTags over src and returns the failure
// message, or "" if collection succeeds.
func collectFieldTagsFail(t *testing.T, src string) (msg string) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	conf := types.Config{Error: func(error) {}}
	conf.Check("x", fset, []*ast.File{f}, info)

	g := &Generator{
		pkg:        &packages.Package{Fset: fset, Syntax: []*ast.File{f}},
		types:      info,
		fieldNames: map[*types.Var]string{},
	}
	defer func() {
		if r := recover(); r != nil {
			if fail, ok := r.(*failure); ok {
				msg = fail.msg
				return
			}
			panic(r)
		}
	}()
	g.collectFieldTags()
	return ""
}
