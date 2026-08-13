package testing

import (
	"solod.dev/so/flag"
	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/strings"
)

// The C backing for the variadic Errorf/Fatalf methods, which cannot be
// expressed in So (a So variadic packs its args into a slice; a real C variadic
// is needed to forward them to fmt).
//
//so:embed testing.h
var testing_h string

//so:embed testing.c
var testing_c string

// T is the context passed to a test function. It records failure and skip
// state for a single test.
//
// The plain message methods (Log, Error, Fatal, Skip) take a preformatted
// string. For formatted messages use the variadic [T.Errorf] and [T.Fatalf]:
//
//	t.Errorf("Index = %d, want 6", got)
//
// So also has no recover, so T cannot unwind a running test. Fatal only marks
// the test failed and prints the message; by convention the test function must
// return right after calling it:
//
//	if err != nil {
//		t.Fatal("open failed")
//		return
//	}
//
// By design, a hard crash (panic or segfault) in a test aborts
// the entire test run, not just the current test.
type T struct {
	name    string
	w       io.Writer
	failed  bool
	skipped bool

	alloc mem.Tracker
}

// Name returns the name of the running test.
func (t *T) Name() string { return t.name }

// Allocator returns the memory allocator for the test. Allocations made
// through it are tracked, and after the test function returns the runner fails
// the test if any of them were not freed. Use it in place of [mem.System] to
// enable leak checking:
//
//	alloc := t.Allocator()
//	p := mem.Alloc[int](alloc)
//	defer mem.Free(alloc, p)
//
// Allocations made through any other allocator are not tracked.
func (t *T) Allocator() mem.Allocator { return &t.alloc }

// Failed reports whether the test has failed.
func (t *T) Failed() bool { return t.failed }

// Fail marks the test as failed but continues execution.
func (t *T) Fail() { t.failed = true }

// Log records msg in the test log.
func (t *T) Log(msg string) {
	fmt.Fprintf(t.w, "    %s\n", msg)
}

// Error is equivalent to Log followed by Fail.
func (t *T) Error(msg string) {
	t.Log(msg)
	t.Fail()
}

// msgSize is the size of the buffer that [T.Errorf] and [T.Fatalf] format
// into. A longer message is truncated.
//
//so:extern testing_msgSize
const msgSize = 1024

// Errorf formats its arguments like [fmt.Sprintf], then
// behaves like [T.Error] (Log followed by Fail).
//
//so:extern nodecay
func (t *T) Errorf(format string, args ...any) {
	buf := make([]byte, msgSize)
	t.Error(fmt.Sprintf(buf, format, args...))
}

// Fatal is equivalent to Log followed by Fail. The test function must return
// after calling it; see [T].
func (t *T) Fatal(msg string) {
	t.Log(msg)
	t.Fail()
}

// Fatalf formats its arguments like [fmt.Sprintf], then behaves like [T.Fatal].
// The test function must return after calling it; see [T].
//
//so:extern nodecay
func (t *T) Fatalf(format string, args ...any) {
	buf := make([]byte, msgSize)
	t.Fatal(fmt.Sprintf(buf, format, args...))
}

// Skip marks the test as skipped. Like Fatal, the test must return afterwards.
func (t *T) Skip(msg string) {
	t.Log(msg)
	t.skipped = true
}

// Test represents a single test to be run by the test runner.
type Test struct {
	Name string
	F    func(t *T)
}

// Suite holds the tests of a single package. The test runner of a program
// that tests many packages passes one Suite per package to [RunSuites].
type Suite struct {
	Pkg   string
	Tests []Test
}

// RunTests runs the given tests for package pkg, prints per-test results
// to stdout, and exits with a non-zero status if any test failed.
// args is the runner's os.Args; RunTests parses flags from it.
func RunTests(pkg string, args []string, tests []Test) {
	suite := Suite{Pkg: pkg, Tests: tests}
	suites := []Suite{suite}
	RunSuites(args, suites)
}

// RunSuites runs the tests of every suite, prints per-test results to stdout,
// and exits with a non-zero status if any test failed. It runs every suite
// before it reports the status, so one failed package does not hide the
// results of the packages after it.
//
// args is the runner's os.Args; RunSuites parses flags from it.
func RunSuites(args []string, suites []Suite) {
	var run string
	fs := flag.NewFlagSet("so test", flag.ContinueOnError)
	fs.StringVar(&run, "run", "", "run only tests whose names start with this prefix")
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(2)
	}

	ok := true
	for _, suite := range suites {
		if !runSuite(suite, run) {
			ok = false
		}
	}
	if !ok {
		os.Exit(1)
	}
}

// runSuite runs the tests of one suite whose names start with run,
// prints the results, and reports whether every test passed.
func runSuite(suite Suite, run string) bool {
	failed := 0
	skipped := 0
	total := 0
	fmt.Fprintf(os.Stdout, "%s\n", suite.Pkg)

	for _, tc := range suite.Tests {
		if !strings.HasPrefix(tc.Name, run) {
			continue
		}
		total++

		t := &T{name: tc.Name, w: os.Stdout}
		t.alloc.Allocator = mem.System
		fmt.Fprintf(t.w, "=== RUN   %s\n", t.name)
		// A test that crashes takes the whole program down. The name of the
		// test must reach the output before the test runs.
		os.Stdout.Sync()
		tc.F(t)

		// Fail a passing test that leaked memory allocated through t.Allocator().
		if !t.failed && !t.skipped {
			stats := t.alloc.Stats()
			if stats.Mallocs != stats.Frees {
				// %d takes an int, so the uint64 counts need a conversion.
				fmt.Fprintf(t.w, "    memory leak: %d unfreed allocation(s), %d byte(s)\n",
					int(stats.Mallocs-stats.Frees), int(stats.Alloc))
				t.failed = true
			}
		}

		if t.skipped {
			fmt.Fprintf(t.w, "--- SKIP: %s\n", t.name)
			skipped++
			continue
		}
		if t.failed {
			fmt.Fprintf(t.w, "--- FAIL: %s\n", t.name)
			failed++
			continue
		}
		fmt.Fprintf(t.w, "--- PASS: %s\n", t.name)
	}

	if total == 0 {
		fmt.Fprintf(os.Stdout, "ok\t%s\t%d tests [no tests to run]\n", suite.Pkg, total)
		return true
	}
	if failed > 0 {
		fmt.Fprintf(os.Stdout, "FAIL\t%s\t%d of %d failed\n", suite.Pkg, failed, total)
		return false
	}
	if skipped > 0 {
		fmt.Fprintf(os.Stdout, "ok\t%s\t%d tests (%d skipped)\n", suite.Pkg, total, skipped)
		return true
	}
	fmt.Fprintf(os.Stdout, "ok\t%s\t%d tests\n", suite.Pkg, total)
	return true
}
