package compiler

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/nalgeon/be"
)

func TestReadPkgFile(t *testing.T) {
	t.Run("packages", func(t *testing.T) {
		path := writePkgFile(t, "so/sync\nso/bytes\n")
		pkgs, err := readPkgFile(path)
		be.Err(t, err, nil)
		be.Equal(t, pkgs, []string{"so/sync", "so/bytes"})
	})
	t.Run("comments and blanks", func(t *testing.T) {
		path := writePkgFile(t, "# freestanding\n\nso/sync\n  so/bytes  # strings only\n")
		pkgs, err := readPkgFile(path)
		be.Err(t, err, nil)
		be.Equal(t, pkgs, []string{"so/sync", "so/bytes"})
	})
	t.Run("path forms", func(t *testing.T) {
		path := writePkgFile(t, "./so/sync\nso/bytes/\n")
		pkgs, err := readPkgFile(path)
		be.Err(t, err, nil)
		be.Equal(t, pkgs, []string{"so/sync", "so/bytes"})
	})
	t.Run("duplicate", func(t *testing.T) {
		path := writePkgFile(t, "so/sync\nso/sync\n")
		_, err := readPkgFile(path)
		be.Err(t, err, "lists so/sync twice")
	})
	t.Run("empty", func(t *testing.T) {
		path := writePkgFile(t, "# nothing\n")
		_, err := readPkgFile(path)
		be.Err(t, err, "lists no packages")
	})
	t.Run("missing file", func(t *testing.T) {
		_, err := readPkgFile(filepath.Join(t.TempDir(), "nope.txt"))
		be.Err(t, err, fs.ErrNotExist)
	})
}

func TestSelectPkgs(t *testing.T) {
	root := t.TempDir()
	dirs := []string{
		filepath.Join(root, "so", "bytes", "test"),
		filepath.Join(root, "so", "sync", "test"),
	}

	t.Run("subset", func(t *testing.T) {
		path := writePkgFile(t, "so/sync\n")
		got, err := selectPkgs(root, dirs, path)
		be.Err(t, err, nil)
		be.Equal(t, got, []string{dirs[1]})
	})
	t.Run("sorted", func(t *testing.T) {
		path := writePkgFile(t, "so/sync\nso/bytes\n")
		got, err := selectPkgs(root, dirs, path)
		be.Err(t, err, nil)
		be.Equal(t, got, dirs)
	})
	t.Run("unselected package", func(t *testing.T) {
		path := writePkgFile(t, "so/time\n")
		_, err := selectPkgs(root, dirs, path)
		be.Err(t, err, "lists so/time, but the pattern selects no test directory for it")
	})
}

// writePkgFile writes content to a package file in a temporary directory
// and returns the path of the file.
func writePkgFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "pkgs.txt")
	be.Err(t, os.WriteFile(path, []byte(content), 0o644), nil)
	return path
}
