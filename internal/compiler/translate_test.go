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
	testDirs, err := filepath.Glob("../../testdata/*")
	be.Err(t, err, nil)

	for _, testDir := range testDirs {
		if !isDir(testDir) {
			continue
		}
		parts := strings.Split(testDir, string(filepath.Separator))
		name := parts[len(parts)-1]
		t.Run(name, func(t *testing.T) {
			testPackage(t, testDir)
		})
	}
}

func testPackage(t *testing.T, testDir string) {
	srcDir := filepath.Join(testDir, "src")
	expectedDir := filepath.Join(testDir, "dst")

	// Create temp output dir
	tempOut, err := os.MkdirTemp("", "solod_out")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	_, err = Translate(srcDir, tempOut, Options{})
	be.Err(t, err, nil)

	// Compare output with expected (recursively)
	err = filepath.WalkDir(expectedDir, func(path string, d fs.DirEntry, err error) error {
		return assertFile(t, expectedDir, path, tempOut, d, err)
	})
	be.Err(t, err, nil)

	// Verify builtin files are copied to output
	for _, name := range []string{"so/builtin/builtin.h", "so/builtin/builtin.c"} {
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

func TestTrackSource(t *testing.T) {
	srcDir := "../../testdata/panic/src"
	tempOut, err := os.MkdirTemp("", "so_tracksource")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	_, err = Translate(srcDir, tempOut, Options{TrackSource: true})
	be.Err(t, err, nil)

	content, err := os.ReadFile(filepath.Join(tempOut, "main.c"))
	be.Err(t, err, nil)

	// Verify #line directives: format "#line N "filename""
	found := false
	for line := range strings.SplitSeq(string(content), "\n") {
		if strings.HasPrefix(line, "#line ") {
			found = true
			parts := strings.SplitN(line, " ", 3)
			if len(parts) != 3 {
				t.Errorf("malformed #line directive: %s", line)
			}
			if !strings.HasSuffix(parts[2], `main.go"`) {
				t.Errorf("expected #line to reference main.go: %s", line)
			}
		}
	}
	if !found {
		t.Fatal("no #line directives found")
	}
}

func TestTranslateLinks(t *testing.T) {
	// The fixture imports so/math, which declares //so:link m.
	srcDir := "testdata/link"
	tempOut, err := os.MkdirTemp("", "so_link")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	libs, err := Translate(srcDir, tempOut, Options{})
	be.Err(t, err, nil)
	be.Equal(t, libs, []string{"m"})
}

func TestTranslateLinkEmpty(t *testing.T) {
	// A so:link directive without a library name must be rejected.
	srcDir := "testdata/link_empty"
	tempOut, err := os.MkdirTemp("", "so_link_empty")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	_, err = Translate(srcDir, tempOut, Options{})
	be.True(t, err != nil)
}

func TestTranslateGenericField(t *testing.T) {
	// A type parameter in a struct field has no C representation.
	srcDir := "testdata/generic_field"
	tempOut, err := os.MkdirTemp("", "so_generic_field")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	_, err = Translate(srcDir, tempOut, Options{})
	be.True(t, err != nil)
	be.True(t, strings.Contains(err.Error(), "type parameter T"))
}

func TestTranslateGenericMethod(t *testing.T) {
	// A generic method must be so:inline or so:extern, because call sites
	// pass the receiver's type arguments that a regular C function cannot take.
	srcDir := "testdata/generic_method"
	tempOut, err := os.MkdirTemp("", "so_generic_method")
	be.Err(t, err, nil)
	defer os.RemoveAll(tempOut)

	_, err = Translate(srcDir, tempOut, Options{})
	be.True(t, err != nil)
	be.True(t, strings.Contains(err.Error(), "generic method Get"))
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
