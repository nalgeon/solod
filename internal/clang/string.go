package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// emitStringLit emits a string literal, handling both interpreted and raw strings.
func (g *Generator) emitStringLit(w io.Writer, n *ast.BasicLit) {
	fmt.Fprintf(w, "so_str(%s)", g.cStringLit(n))
}

// emitStringLitConcat emits a chain of string literal additions as adjacent C string literals.
func (g *Generator) emitStringLitConcat(w io.Writer, expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		fmt.Fprint(w, g.cStringLit(e))
	case *ast.BinaryExpr:
		g.emitStringLitConcat(w, e.X)
		fmt.Fprint(w, " ")
		g.emitStringLitConcat(w, e.Y)
	}
}

// emitCharLit emits a byte or rune literal. Like cStringLit, it works from the
// value rather than the source text, since Go and C disagree on some escapes.
func (g *Generator) emitCharLit(w io.Writer, n *ast.BasicLit) {
	tv, ok := g.types.Types[n]
	if !ok || tv.Value == nil {
		g.fail(n, "unresolved character literal: %s", n.Value)
	}
	code, exact := constant.Int64Val(constant.ToInt(tv.Value))
	if !exact {
		g.fail(n, "character literal out of range: %s", n.Value)
	}

	switch {
	case code == '\n':
		fmt.Fprint(w, `'\n'`)
	case code == '\t':
		fmt.Fprint(w, `'\t'`)
	case code == '\r':
		fmt.Fprint(w, `'\r'`)
	case code == '\\' || code == '\'':
		fmt.Fprintf(w, `'\%c'`, code)
	case code >= 0x20 && code < 0x7f:
		// Only printable ASCII becomes a C character literal.
		fmt.Fprintf(w, "'%c'", code)
	default:
		// Anything else becomes a number. Not U'...': that is a char32_t,
		// which is unsigned and would make rune arithmetic unsigned too.
		fmt.Fprintf(w, "0x%02x", code)
	}
}

// hasStringType reports whether the given expression has string type.
func (g *Generator) hasStringType(expr ast.Expr) bool {
	typ := g.types.TypeOf(expr)
	basic, ok := typ.Underlying().(*types.Basic)
	return ok && (basic.Kind() == types.String || basic.Kind() == types.UntypedString)
}

// isStringLit reports whether an expression is a string literal
// or a chain of string literal additions.
func isStringLit(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind == token.STRING
	case *ast.BinaryExpr:
		return e.Op == token.ADD && isStringLit(e.X) && isStringLit(e.Y)
	}
	return false
}

// cStringLit returns the C string literal for a Go string literal, interpreted or
// raw. The literal is decoded to its bytes and re-encoded, because Go escape rules
// differ from C. Does not include the so_str() wrapper.
func (g *Generator) cStringLit(n *ast.BasicLit) string {
	s, err := strconv.Unquote(n.Value)
	if err != nil {
		g.fail(n, "invalid string literal: %s", n.Value)
	}

	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		ch := s[i]
		// Multi-byte UTF-8 passes through, so the C stays readable.
		if ch >= utf8.RuneSelf {
			if _, size := utf8.DecodeRuneInString(s[i:]); size > 1 {
				b.WriteString(s[i : i+size])
				i += size - 1
				continue
			}
		}
		switch {
		case ch == '\\' || ch == '"':
			b.WriteByte('\\')
			b.WriteByte(ch)
		case ch == '\n':
			b.WriteString(`\n`)
		case ch == '\t':
			b.WriteString(`\t`)
		case ch == '\r':
			b.WriteString(`\r`)
		case ch < 0x20 || ch >= 0x7f:
			// A C octal escape ends after three digits,
			// while \x absorbs every hex digit that follows it.
			fmt.Fprintf(&b, `\%03o`, ch)
		case ch == '?' && i > 0 && s[i-1] == '?':
			// Escape the second ? of a pair, so no two ? end up next to each
			// other and the bytes cannot form a C trigraph like ??!.
			b.WriteString(`\?`)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func stringCompareFunc(op token.Token) string {
	switch op {
	case token.EQL:
		return "so_string_eq"
	case token.NEQ:
		return "so_string_ne"
	case token.LSS:
		return "so_string_lt"
	case token.LEQ:
		return "so_string_lte"
	case token.GTR:
		return "so_string_gt"
	case token.GEQ:
		return "so_string_gte"
	}
	panic("unreachable")
}
