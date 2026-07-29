package clang

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"
)

// failure is a diagnostic produced by fail.
type failure struct {
	pos token.Position
	msg string
}

func (f *failure) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s", f.pos, f.msg)
	if srcLine, err := readSourceLine(f.pos.Filename, f.pos.Line); err == nil {
		fmt.Fprintf(&b, "\n%s\n%s", srcLine, errorMarker(srcLine, f.pos))
	}
	return b.String()
}

// fail aborts code generation with a diagnostic anchored at node. It does not return.
func (g *Generator) fail(node ast.Node, format string, args ...any) {
	pos := g.pkg.Fset.Position(node.Pos())
	err := &failure{pos: pos, msg: fmt.Sprintf(format, args...)}
	panic(err)
}

// typeString returns a type name for diagnostics.
func (g *Generator) typeString(typ types.Type) string {
	return types.TypeString(typ, types.RelativeTo(g.pkg.Types))
}

// readSourceLine reads a single line from a source file (1-indexed).
func readSourceLine(filename string, line int) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for i := 1; scanner.Scan(); i++ {
		if i == line {
			return scanner.Text(), nil
		}
	}
	return "", fmt.Errorf("line %d not found in %s", line, filename)
}

// errorMarker return a string with a caret pointing to the error column.
func errorMarker(srcLine string, pos token.Position) string {
	col := min(pos.Column-1, len(srcLine))
	pad := make([]byte, col)
	for i := range col {
		if srcLine[i] == '\t' {
			pad[i] = '\t'
		} else {
			pad[i] = ' '
		}
	}
	return fmt.Sprintf("%s^here", pad)
}
