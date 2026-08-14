package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// benchKind discovers BenchmarkXxx(b *testing.B) functions
// and runs them via testing.RunBenchmarks.
var benchKind = kind{
	subdir:  "bench",
	command: "so bench",
	noun:    "benchmark",
	prefix:  "Benchmark",
	param:   "B",
	dir:     ".sobench",
}

// Bench discovers BenchmarkXxx functions in the "bench" subdirectory of srcDir,
// then compiles and runs them. runPrefix limits the run to the benchmarks whose
// names start with it.
func Bench(srcDir, runPrefix string, opts Options) error {
	src, err := benchSource(srcDir, runPrefix)
	if err != nil {
		return err
	}
	return run(src, nil, opts)
}

// benchSource returns the entry package of the benchmark program for srcDir.
// One run measures one package, so the pattern selects no other package.
func benchSource(srcDir, runPrefix string) (source, error) {
	dir := filepath.Join(srcDir, benchKind.subdir)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return source{}, fmt.Errorf("no %s directory: %s", benchKind.subdir, dir)
	}

	root, modPath, err := findModule(dir)
	if err != nil {
		return source{}, err
	}
	s, err := benchKind.collect(root, modPath, dir)
	if err != nil {
		return source{}, err
	}

	return benchKind.source(root, emitBenchRunner(s, runPrefix))
}

// emitBenchRunner returns the source of the runner program that dispatches the
// benchmarks via testing.RunBenchmarks. Benchmarks always use the system
// allocator; a package that needs a different one can write a main package
// of its own and use `so run`.
func emitBenchRunner(s suite, runPrefix string) []byte {
	var b strings.Builder
	b.WriteString(benchKind.header())

	b.WriteString("import (\n")
	b.WriteString("\t\"solod.dev/so/mem\"\n")
	b.WriteString("\t\"solod.dev/so/testing\"\n\n")
	fmt.Fprintf(&b, "\t%s %q\n", s.pkg, s.path)
	b.WriteString(")\n\n")

	b.WriteString("func main() {\n")
	fmt.Fprintf(&b, "\topts := testing.Options{Run: %q}\n", runPrefix)
	fmt.Fprintf(&b, "\ttesting.RunBenchmarks(mem.System, %q, opts, []testing.Benchmark{\n", s.label)
	for _, name := range s.names {
		fmt.Fprintf(&b, "\t\t{Name: %q, F: %s.%s},\n", name, s.pkg, name)
	}
	b.WriteString("\t})\n")
	b.WriteString("}\n")

	return []byte(b.String())
}
