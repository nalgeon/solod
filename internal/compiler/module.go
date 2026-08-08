package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// findModule returns the root directory and the module path of the module
// that holds dir.
func findModule(dir string) (root, path string, err error) {
	cur, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	for {
		gomod := filepath.Join(cur, "go.mod")
		data, err := os.ReadFile(gomod)
		if err == nil {
			path := modfile.ModulePath(data)
			if path == "" {
				return "", "", fmt.Errorf("no module path in %s", gomod)
			}
			return cur, path, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", "", fmt.Errorf("no go.mod in any directory above %s", dir)
		}
		cur = parent
	}
}

// moduleRel returns the path of dir relative to the module root, in slash form.
// The import path of the package in dir is the module path joined with it. A
// dir outside the module has no import path, so moduleRel rejects it.
func moduleRel(root, dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("%s is outside the module in %s: "+
			"a package of another module needs a separate run", dir, root)
	}
	return rel, nil
}
