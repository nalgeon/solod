package testing

import (
	"os" // for testing

	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
)

//so:embed testing.h
var testing_h string

//so:embed testing.c
var testing_c string

// flushOut pushes the buffered standard output to the operating system.
// A freestanding host has no standard output, so flushOut does nothing there.
//
//so:extern fmt_flushOut
func flushOut() { os.Stdout.Sync() }

// exitFail ends the program with a failure status. A freestanding host has no
// exit status, so exitFail traps.
//
//so:extern testing_exitFail
func exitFail() { os.Exit(1) }

// heapMark returns the position of the next allocation in the heap in
// freestanding mode. Returns 0 in hosted mode.
//
//so:extern so_heap_mark
func heapMark() uintptr { return 0 }

// heapRelease returns the heap to the position mark in freestanding mode, which
// loses every allocation made after heapMark read the mark. No-op in hosted mode.
//
//so:extern so_heap_release
func heapRelease(mark uintptr) {}

// T is the context passed to a test function. It records failure and skip
// state for a single test.
//
// The plain message methods (Log, Error, Fatal, Skip) take a preformatted
// string. For formatted messages use the variadic [T.Errorf] and [T.Fatalf]:
//
//	t.Errorf("Index = %d, want 6", got)
//
// Solod also has no recover, so T cannot unwind a running test. Fatal only marks
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

// Options selects the tests or benchmarks of a run.
type Options struct {
	Run string // run only tests or benchmarks whose names start with this prefix
}

// RunTests runs the given tests for package pkg, prints per-test results,
// and exits with a non-zero status if any test failed.
func RunTests(pkg string, opts Options, tests []Test) {
	suite := Suite{Pkg: pkg, Tests: tests}
	suites := []Suite{suite}
	RunSuites(opts, suites)
}

// RunSuites runs the tests of every suite, prints per-test results, and exits
// with a non-zero status if any test failed. It runs every suite before it
// reports the status, so one failed package does not hide the results of
// the packages after it.
func RunSuites(opts Options, suites []Suite) {
	w := fmt.Output
	// The freestanding heap never reclaims memory, so every test returns it to
	// the position it holds now. The mark keeps the allocations made before the
	// first test, like the ones of a package level variable.
	mark := heapMark()
	ok := true
	for _, suite := range suites {
		if !runSuite(w, suite, opts.Run, mark) {
			ok = false
		}
	}
	if !ok {
		exitFail()
	}
}

// runSuite runs the tests of one suite whose names start with run,
// prints the results to w, and reports whether every test passed.
// Every test returns the heap to the position mark.
func runSuite(w io.Writer, suite Suite, run string, mark uintptr) bool {
	failed := 0
	skipped := 0
	total := 0
	fmt.Fprintf(w, "%s\n", suite.Pkg)

	for _, tc := range suite.Tests {
		if !strings.HasPrefix(tc.Name, run) {
			continue
		}
		total++

		t := &T{name: tc.Name, w: w}
		t.alloc.Allocator = mem.System
		fmt.Fprintf(t.w, "=== RUN   %s\n", t.name)
		// A test that crashes takes the whole program down. The name of the
		// test must reach the output before the test runs.
		flushOut()
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
		} else if t.failed {
			fmt.Fprintf(t.w, "--- FAIL: %s\n", t.name)
			failed++
		} else {
			fmt.Fprintf(t.w, "--- PASS: %s\n", t.name)
		}
		heapRelease(mark)
	}

	if total == 0 {
		fmt.Fprintf(w, "ok\t%s\t%d tests [no tests to run]\n", suite.Pkg, total)
		return true
	}
	if failed > 0 {
		fmt.Fprintf(w, "FAIL\t%s\t%d of %d failed\n", suite.Pkg, failed, total)
		return false
	}
	if skipped > 0 {
		fmt.Fprintf(w, "ok\t%s\t%d tests (%d skipped)\n", suite.Pkg, total, skipped)
		return true
	}
	fmt.Fprintf(w, "ok\t%s\t%d tests\n", suite.Pkg, total)
	return true
}
