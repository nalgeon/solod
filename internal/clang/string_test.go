package clang

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/nalgeon/be"
	"golang.org/x/tools/go/packages"
)

func TestCStringLit(t *testing.T) {
	tests := []struct {
		name string
		in   string // Go string literal, as written in the source
		want string // C string literal
	}{
		{"plain", `"abc"`, `"abc"`},
		{"quotes and backslash", `"say \"hi\" \\"`, `"say \"hi\" \\"`},
		{"known escapes", `"a\n\t\rb"`, `"a\n\t\rb"`},
		{"utf8 passes through", `"日本語"`, `"日本語"`},
		// A C hex escape absorbs every hex digit that follows, so these must
		// not come out as \x.
		{"hex escape before letter", `"a\xffb"`, `"a\377b"`},
		{"hex escape before digit", `"\x0aa"`, `"\na"`},
		{"stray high byte", `"\xff"`, `"\377"`},
		{"nul byte", `"a\x00b"`, `"a\000b"`},
		{"delete byte", `"\x7f"`, `"\177"`},
		// C rejects a universal character name for a basic character.
		{"escaped basic character", "\"\\u0041\"", `"A"`},
		{"escaped non-basic character", "\"\\u00e9\"", `"é"`},
		{"raw string", "`a\\x41`", `"a\\x41"`},
		{"raw string keeps quotes", "`say \"hi\"`", `"say \"hi\""`},
		{"raw string drops cr", "`a\rb`", `"ab"`},
		// Two ? next to each other can start a C trigraph, so the second one
		// is escaped. A lone ? is left alone.
		{"lone question mark", `"why?"`, `"why?"`},
		{"trigraph", `"what??!"`, `"what?\?!"`},
		{"question marks in a row", `"????"`, `"?\?\?\?"`},
	}
	g := &Generator{pkg: &packages.Package{Fset: token.NewFileSet()}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{Kind: token.STRING, Value: tt.in}
			be.Equal(t, g.cStringLit(lit), tt.want)
		})
	}
}

func TestEmitCharLit(t *testing.T) {
	tests := []struct {
		name string
		decl string // variable declaration with a character literal
		want string // C expression
	}{
		{"ascii rune", "var v rune = 'a'", "'a'"},
		{"ascii byte", "var v byte = 'a'", "'a'"},
		{"space", "var v rune = ' '", "' '"},
		{"newline", `var v rune = '\n'`, `'\n'`},
		{"quote", `var v rune = '\''`, `'\''`},
		{"backslash", `var v rune = '\\'`, `'\\'`},
		{"escape char", `var v rune = '\x1b'`, "0x1b"},
		{"delete char", `var v rune = '\x7f'`, "0x7f"},
		{"high byte", `var v byte = '\xe9'`, "0xe9"},
		{"latin rune", "var v rune = 'é'", "0xe9"},
		{"cjk rune", "var v rune = '世'", "0x4e16"},
		{"emoji rune", "var v rune = '😀'", "0x1f600"},
		{"named type", "type R rune; var v R = 'a'", "'a'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be.Equal(t, emitCharLitOf(t, tt.decl), tt.want)
		})
	}
}

// emitCharLitOf type-checks a declaration and returns
// the C emitted for the character literal in it.
func emitCharLitOf(t *testing.T, decl string) string {
	t.Helper()
	info, f := checkSnippet(t, "package x\n"+decl)
	var lit *ast.BasicLit
	ast.Inspect(f, func(n ast.Node) bool {
		if bl, ok := n.(*ast.BasicLit); ok && bl.Kind == token.CHAR {
			lit = bl
		}
		return true
	})
	if lit == nil {
		t.Fatalf("no character literal in %s", decl)
	}
	g := &Generator{pkg: &packages.Package{Syntax: []*ast.File{f}}, types: info}
	var b strings.Builder
	g.emitCharLit(&b, lit)
	return b.String()
}
