package compiler

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

// badDir holds the cases that must fail to translate.
const badDir = "bad"

func TestTranslate(t *testing.T) {
	for _, testDir := range caseDirs(t, "../../testdata/*") {
		name := filepath.Base(testDir)
		if name == badDir {
			continue
		}
		t.Run(name, func(t *testing.T) {
			testPackage(t, testDir)
		})
	}
}

func TestTranslateBad(t *testing.T) {
	for _, testDir := range caseDirs(t, filepath.Join("../../testdata", badDir, "*")) {
		t.Run(filepath.Base(testDir), func(t *testing.T) {
			testBadPackage(t, testDir)
		})
	}
}

func caseDirs(t *testing.T, pattern string) []string {
	matches, err := filepath.Glob(pattern)
	be.Err(t, err, nil)

	dirs := make([]string, 0, len(matches))
	for _, match := range matches {
		if isDir(match) {
			dirs = append(dirs, match)
		}
	}
	return dirs
}

// testPackage asserts that the case translates and matches its dst files.
func testPackage(t *testing.T, testDir string) {
	srcDir := filepath.Join(testDir, "src")
	expectedDir := filepath.Join(testDir, "dst")
	tempOut := t.TempDir()

	libs, err := Translate(srcDir, tempOut, Options{})
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

	// Compare linked libraries with expected, if the case declares any.
	if want, ok := readGolden(t, filepath.Join(testDir, "links.txt")); ok {
		be.Equal(t, strings.Join(libs, "\n"), want)
	}
}

// testBadPackage asserts that the case fails to translate
// with the error recorded in error.txt.
func testBadPackage(t *testing.T, testDir string) {
	srcDir := filepath.Join(testDir, "src")

	_, err := Translate(srcDir, t.TempDir(), Options{})
	if err == nil {
		t.Fatal("expected translation to fail")
	}

	got := cleanError(err, srcDir)
	want, ok := readGolden(t, filepath.Join(testDir, "error.txt"))
	if !ok {
		t.Fatalf("missing error.txt:\ngot:\n%s", got)
	}
	if got != want {
		t.Errorf("error.txt:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// readGolden reads a golden file, reporting whether it exists.
func readGolden(t *testing.T, path string) (string, bool) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false
	}
	be.Err(t, err, nil)
	return strings.TrimSpace(string(content)), true
}

// cleanError strips the absolute source path from the error message,
// so that golden errors do not depend on the checkout location.
func cleanError(err error, srcDir string) string {
	msg := err.Error()
	if absSrc, absErr := filepath.Abs(srcDir); absErr == nil {
		msg = strings.ReplaceAll(msg, absSrc+string(filepath.Separator), "")
	}
	return strings.TrimSpace(msg)
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
	tempOut := t.TempDir()

	_, err := Translate(srcDir, tempOut, Options{TrackSource: true})
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

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
