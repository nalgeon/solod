package clang

import (
	"go/ast"
	"strings"
)

// directives holds parsed so-directive annotations from a comment group.
type directives struct {
	inline      bool
	promote     bool
	volatile    bool
	threadLocal bool
	attrs       []string
}

// attrString returns a combined __attribute__((...)) string,
// or "" if no attrs are present.
func (d directives) attrString() string {
	if len(d.attrs) == 0 {
		return ""
	}
	return "__attribute__((" + strings.Join(d.attrs, ", ") + "))"
}

// parseDirectives scans a comment group for so: directives.
func parseDirectives(doc *ast.CommentGroup) directives {
	var d directives
	if doc == nil {
		return d
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(c.Text)
		switch {
		case text == "//so:inline":
			d.inline = true
		case text == "//so:promote":
			d.promote = true
		case text == "//so:volatile":
			d.volatile = true
		case text == "//so:thread_local":
			d.threadLocal = true
		case strings.HasPrefix(text, "//so:attr "):
			attr := strings.TrimSpace(strings.TrimPrefix(text, "//so:attr "))
			if attr != "" {
				d.attrs = append(d.attrs, attr)
			}
		}
	}
	return d
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
