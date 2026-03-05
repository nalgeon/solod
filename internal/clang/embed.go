package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// embedFile represents a single embedded file.
type embedFile struct {
	name    string
	content string
}

// Embeds holds the embedded .h and .c files.
type Embeds struct {
	header []embedFile     // .h contents to inline in header
	impl   []embedFile     // .c contents to inline in impl
	vars   map[string]bool // var names to skip during emission
}

// collectEmbeds scans package files for //so:embed directives, reads the
// referenced files, and categorizes them by extension (.h -> header, .c -> impl).
func collectEmbeds(pkg *packages.Package) (Embeds, error) {
	embeds := Embeds{vars: make(map[string]bool)}
	if len(pkg.GoFiles) == 0 {
		return embeds, nil
	}
	srcDir := filepath.Dir(pkg.GoFiles[0])

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}

			// Read the file specified in the //so:embed directive.
			filename, ok := embedDirective(gd.Doc)
			if !ok {
				continue
			}
			content, err := os.ReadFile(filepath.Join(srcDir, filename))
			if err != nil {
				return embeds, fmt.Errorf("read file %s: %w", filename, err)
			}

			// Register the embedded file.
			ef := embedFile{name: filename, content: string(content)}
			switch filepath.Ext(filename) {
			case ".h":
				embeds.header = append(embeds.header, ef)
			case ".c":
				embeds.impl = append(embeds.impl, ef)
			}
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				for _, name := range vs.Names {
					embeds.vars[name.Name] = true
				}
			}
		}
	}
	return embeds, nil
}

// embedDirective extracts the filename from a //so:embed comment directive.
func embedDirective(doc *ast.CommentGroup) (string, bool) {
	if doc == nil {
		return "", false
	}
	for _, c := range doc.List {
		if filename, ok := strings.CutPrefix(strings.TrimSpace(c.Text), "//so:embed "); ok {
			return strings.TrimSpace(filename), true
		}
	}
	return "", false
}
