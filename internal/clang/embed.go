package clang

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"
)

// embedFile represents a single embedded file.
type embedFile struct {
	name    string
	content string
}

// loadEmbedFile reads the contents of an embedded file from disk.
func loadEmbedFile(dir, filename string) (embedFile, error) {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return embedFile{}, err
	}
	ef := embedFile{name: filename, content: string(content)}
	return ef, nil
}

// Embeds holds the embedded .h and .c files.
type Embeds struct {
	header []embedFile     // .h contents to inline in header
	impl   []embedFile     // .c contents to inline in impl
	vars   map[string]bool // var names to skip during emission
}

// newEmbeds creates a new Embeds instance.
func newEmbeds() Embeds {
	return Embeds{
		vars: make(map[string]bool),
	}
}

// addFile adds an embedded file to the appropriate list
// based on its extension.
func (e *Embeds) addFile(ef embedFile) {
	switch filepath.Ext(ef.name) {
	case ".h":
		e.header = append(e.header, ef)
	case ".c":
		e.impl = append(e.impl, ef)
	}
}

// parseEmbed extracts the filename from a so:embed directive.
func parseEmbed(doc *ast.CommentGroup) (string, bool) {
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
