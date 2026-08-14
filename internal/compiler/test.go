package compiler

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// testKind discovers TestXxx(t *testing.T) functions.
var testKind = kind{
	subdir:  "test",
	command: "so test",
	noun:    "test",
	prefix:  "Test",
	param:   "T",
	dir:     ".sotest",
}

// Selection limits a test run. PkgFile is the path of a file that lists the
// packages to run. Run is a name prefix: only the tests whose names start
// with it run. An empty field selects everything.
type Selection struct {
	PkgFile string
	Run     string
}

// Test discovers TestXxx functions in the "test" subdirectory of every package
// the pattern selects, then compiles and runs them as one program. sel limits
// the run to some of the packages and tests.
func Test(pattern string, sel Selection, opts Options) error {
	src, err := testSource(pattern, sel)
	if err != nil {
		return err
	}
	return run(src, nil, opts)
}

// TranslateTests is Test without the compile and run steps: it writes the C of
// the test program to outDir. It returns the C libraries the program must link
// against, deduplicated and sorted, without the -l prefix. sel applies when the
// runner is generated, so the translated program runs the selected tests alone.
func TranslateTests(pattern, outDir string, sel Selection, opts Options) ([]string, error) {
	src, err := testSource(pattern, sel)
	if err != nil {
		return nil, err
	}
	return translate(src, outDir, opts)
}

// testSource returns the entry package of the test program for the packages
// the pattern selects. A pattern that ends with "..." selects every package
// below its base directory.
func testSource(pattern string, sel Selection) (source, error) {
	dirs, err := testDirs(pattern)
	if err != nil {
		return source{}, err
	}

	root, modPath, err := findModule(dirs[0])
	if err != nil {
		return source{}, err
	}

	if sel.PkgFile != "" {
		dirs, err = selectPkgs(root, dirs, sel.PkgFile)
		if err != nil {
			return source{}, err
		}
	}

	suites, err := collectSuites(root, modPath, dirs)
	if err != nil {
		return source{}, err
	}

	return testKind.source(root, emitTestRunner(suites, sel.Run))
}

// testDirs returns the test directories the pattern selects, sorted.
// A pattern that ends with "..." selects every test directory below its base
// directory. Any other pattern names one package, which must have a test
// subdirectory.
func testDirs(pattern string) ([]string, error) {
	base, recursive := strings.CutSuffix(pattern, "...")
	if !recursive {
		dir := filepath.Join(pattern, testKind.subdir)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("no %s directory: %s", testKind.subdir, dir)
		}
		return []string{dir}, nil
	}

	if base == "" {
		base = "."
	}
	base = filepath.Clean(base)

	var dirs []string
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		name := d.Name()
		if path != base && (name == "testdata" ||
			strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
			return fs.SkipDir
		}
		// A directory with the right name but no Go files is not a package.
		// The transpiled output holds such directories.
		if name == testKind.subdir && hasGoFiles(path) {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find %s directories: %w", testKind.subdir, err)
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no %s directories under %s", testKind.subdir, base)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// selectPkgs collects the test directories of the packages that the file at path
// lists. Every listed package must have a test directory the pattern selects.
func selectPkgs(root string, dirs []string, path string) ([]string, error) {
	pkgs, err := readPkgFile(path)
	if err != nil {
		return nil, err
	}

	// Map the package under test to its test directory,
	// e.g. "so/sync" to "so/sync/test".
	byPkg := make(map[string]string, len(dirs))
	for _, dir := range dirs {
		rel, err := moduleRel(root, filepath.Dir(dir))
		if err != nil {
			return nil, err
		}
		byPkg[packageName(root, rel)] = dir
	}

	selected := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		dir, ok := byPkg[pkg]
		if !ok {
			return nil, fmt.Errorf("%s lists %s, but the pattern selects no %s directory for it",
				path, pkg, testKind.subdir)
		}
		selected = append(selected, dir)
	}
	sort.Strings(selected)
	return selected, nil
}

// readPkgFile returns the packages that the file at path lists, as paths
// relative to the module root (e.g. "so/sync"). A line holds one package.
// readPkgFile ignores a blank line and the text after a "#".
func readPkgFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkgs []string
	seen := make(map[string]bool)
	for line := range strings.Lines(string(data)) {
		pkg, _, _ := strings.Cut(line, "#")
		pkg = strings.TrimSpace(pkg)
		pkg = strings.TrimSuffix(strings.TrimPrefix(pkg, "./"), "/")
		if pkg == "" {
			continue
		}
		if seen[pkg] {
			return nil, fmt.Errorf("%s lists %s twice", path, pkg)
		}
		seen[pkg] = true
		pkgs = append(pkgs, pkg)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("%s lists no packages", path)
	}
	return pkgs, nil
}

// hasGoFiles reports whether dir holds a Go file that the So compiler reads.
// A Go test file (_test.go) is not one: `go test` runs it, `so test` does not.
func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true
		}
	}
	return false
}

// collectSuites discovers the tests of every directory and returns one suite
// per directory. The So compiler prefixes an exported C name with the package
// name, so two test packages with the same name give two identical C names.
// collectSuites rejects that.
func collectSuites(root, modPath string, dirs []string) ([]suite, error) {
	suites := make([]suite, 0, len(dirs))
	seen := make(map[string]string, len(dirs))

	for _, dir := range dirs {
		s, err := testKind.collect(root, modPath, dir)
		if err != nil {
			return nil, err
		}
		if other, ok := seen[s.pkg]; ok {
			return nil, fmt.Errorf("%s and %s both declare package %s: "+
				"the test packages of one run must have distinct names", other, dir, s.pkg)
		}
		seen[s.pkg] = dir
		suites = append(suites, s)
	}

	return suites, nil
}

// emitTestRunner returns the source of the runner program that dispatches
// every suite via testing.RunSuites.
func emitTestRunner(suites []suite, runPrefix string) []byte {
	var b strings.Builder
	b.WriteString(testKind.header())

	b.WriteString("import (\n")
	b.WriteString("\t\"solod.dev/so/testing\"\n\n")
	for _, s := range suites {
		fmt.Fprintf(&b, "\t%s %q\n", s.pkg, s.path)
	}
	b.WriteString(")\n\n")

	b.WriteString("func main() {\n")
	fmt.Fprintf(&b, "\topts := testing.Options{Run: %q}\n", runPrefix)
	b.WriteString("\ttesting.RunSuites(opts, []testing.Suite{\n")
	for _, s := range suites {
		fmt.Fprintf(&b, "\t\t{Pkg: %q, Tests: []testing.Test{\n", s.label)
		for _, name := range s.names {
			fmt.Fprintf(&b, "\t\t\t{Name: %q, F: %s.%s},\n", name, s.pkg, name)
		}
		b.WriteString("\t\t}},\n")
	}
	b.WriteString("\t})\n")
	b.WriteString("}\n")

	return []byte(b.String())
}
