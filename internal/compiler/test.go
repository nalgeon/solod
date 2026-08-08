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

// Test discovers TestXxx functions in the "test" subdirectory of every package
// the pattern selects, then compiles and runs them as one program.
func Test(pattern string, args []string, opts Options) error {
	src, err := testSource(pattern)
	if err != nil {
		return err
	}
	return run(src, args, opts)
}

// TranslateTests is Test without the compile and run steps: it writes the C of
// the test program to outDir. It returns the C libraries the program must link
// against, deduplicated and sorted, without the -l prefix.
func TranslateTests(pattern, outDir string, opts Options) ([]string, error) {
	src, err := testSource(pattern)
	if err != nil {
		return nil, err
	}
	return translate(src, outDir, opts)
}

// testSource returns the entry package of the test program for the packages
// the pattern selects. A pattern that ends with "..." selects every package
// below its base directory.
func testSource(pattern string) (source, error) {
	dirs, err := testDirs(pattern)
	if err != nil {
		return source{}, err
	}

	root, modPath, err := findModule(dirs[0])
	if err != nil {
		return source{}, err
	}

	suites, err := collectSuites(root, modPath, dirs)
	if err != nil {
		return source{}, err
	}

	return testKind.source(root, emitTestRunner(suites))
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
// every suite via testing.RunSuites. The runner imports os and forwards
// os.Args, so RunSuites can parse flags like -run.
func emitTestRunner(suites []suite) []byte {
	var b strings.Builder
	b.WriteString(testKind.header())

	b.WriteString("import (\n")
	b.WriteString("\t\"solod.dev/so/os\"\n")
	b.WriteString("\t\"solod.dev/so/testing\"\n\n")
	for _, s := range suites {
		fmt.Fprintf(&b, "\t%s %q\n", s.pkg, s.path)
	}
	b.WriteString(")\n\n")

	b.WriteString("func main() {\n")
	b.WriteString("\ttesting.RunSuites(os.Args, []testing.Suite{\n")
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
