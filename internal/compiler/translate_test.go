package compiler

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

func TestTranslate(t *testing.T) {
	testDirs, err := filepath.Glob("../../tests/*")
	be.Err(t, err, nil)

	for _, testDir := range testDirs {
		if !isDir(testDir) {
			continue
		}
		t.Run(filepath.Base(testDir), func(t *testing.T) {
			testPackage(t, testDir)
		})
	}
}

func testPackage(t *testing.T, testDir string) {
	srcDir := filepath.Join(testDir, "src")
	expectedDir := filepath.Join(testDir, "dst")

	// Create temp output dir
	tempOut, err := os.MkdirTemp("", "soan_out")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	err = Translate(srcDir, tempOut)
	be.Err(t, err, nil)

	// Compare output with expected (recursively)
	err = filepath.WalkDir(expectedDir, func(path string, d fs.DirEntry, err error) error {
		return assertFile(t, expectedDir, path, tempOut, d, err)
	})
	be.Err(t, err, nil)

	// Verify builtin files are copied to output
	for _, name := range []string{"solod.h", "solod.c"} {
		if _, err := os.Stat(filepath.Join(tempOut, name)); err != nil {
			t.Errorf("missing builtin file: %s", name)
		}
	}
}

func assertFile(t *testing.T, dir, path, tempOut string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}

	base := filepath.Base(path)
	if strings.HasSuffix(base, ".ext.c") || strings.HasSuffix(base, ".ext.h") {
		// Ignore externally-provided C files (e.g. from // #include comments).
		return nil
	}

	relPath, err := filepath.Rel(dir, path)
	be.Err(t, err, nil)
	actualPath := filepath.Join(tempOut, relPath)

	expectedContent, err := os.ReadFile(path)
	be.Err(t, err, nil)
	actualContent, err := os.ReadFile(actualPath)
	if err != nil {
		t.Errorf("missing output file: %s", relPath)
		return nil
	}

	got := strings.TrimSpace(string(actualContent))
	want := strings.TrimSpace(string(expectedContent))
	if got != want {
		t.Errorf("%s:\ngot:\n%s\nwant:\n%s", relPath, got, want)
	}
	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
