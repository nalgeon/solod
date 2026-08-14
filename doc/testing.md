# Testing

So programs are tested with the `so test` command and the [so/testing](https://pkg.go.dev/solod.dev/so/testing) package. The model mirrors Go's `testing`, but stays deliberately small: no subtests and no parallelism.

## Layout

Tests for a package live in a `test` subdirectory of that package:

```
so/bytes/
    bytes.go
    ...
    test/
        bytes.go     # your tests
```

Test files are plain `.go` files (no `_test.go` suffix). `go test` ignores them, so you can still keep ordinary Go-level tests elsewhere and run them with `go test`.

The test directory declares a package of its own, not `package main`. `so test` can join the tests of many packages into one program, so it must be able to import each test package. Two test packages of one run must also have distinct names: the So compiler prefixes an exported C name with the package name, so equal package names give equal C names.

A common convention is to name a test package after the package under test, with a `_test` suffix: `so/bytes/test` declares `package bytes_test`, `so/sync/test` declares `package sync_test`, etc.

Since tests are in a separate package, they see only the exported API of the package under test (black-box testing).

## Writing tests

A test is a function named `TestXxx` taking a single `*testing.T`:

```go
package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/testing"
)

func TestEqual(t *testing.T) {
	if !bytes.Equal([]byte("abc"), []byte("abc")) {
		t.Error("Equal(abc, abc) = false, want true")
	}
}
```

The `T` type records failure and skip state for one test:

| Method             | Description                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `Name() string`    | Name of the running test.                                          |
| `Allocator()`      | Tracking allocator; see [Leak checking](#leak-checking).           |
| `Fail()`           | Mark the test failed, keep running.                                |
| `Failed() bool`    | Whether the test has failed.                                       |
| `Log(msg)`         | Record a log line.                                                 |
| `Error(msg)`       | `Log` + `Fail`.                                                    |
| `Errorf(msg, ...)` | `fmt.Sprintf` + `Log` + `Fail`.                                    |
| `Fatal(msg)`       | `Log` + `Fail`. The test must `return` afterwards (see below).     |
| `Fatalf(msg, ...)` | `fmt.Sprintf` + `Log` + `Fail`. The test must `return` afterwards. |
| `Skip(msg)`        | Mark the test skipped. The test must `return` afterwards.          |

### Leak checking

Allocate through `t.Allocator()` instead of `mem.System` to have the test fail if any of those allocations are not freed by the time the test function returns:

```go
func TestAlloc(t *testing.T) {
	alloc := t.Allocator()
	p := mem.Alloc[int](alloc)
	defer mem.Free(alloc, p)
	// ...
}
```

`t.Allocator()` wraps the system allocator in a `mem.Tracker` and compares its allocation and free counts after the test:

```
=== RUN   TestAlloc
    memory leak: 1 unfreed allocation(s), 16 byte(s)
--- FAIL: TestAlloc
```

Only allocations made through `t.Allocator()` are tracked; anything allocated through `mem.System` or another allocator is ignored. The check runs only for a test that would otherwise pass (not for one that already failed or skipped).

## Running

```
so test ./so/sync
```

```
=== RUN   TestCond
--- PASS: TestCond
=== RUN   TestMutex_LockUnlock
--- PASS: TestMutex_LockUnlock
=== RUN   TestMutex_TryLock
--- PASS: TestMutex_TryLock
=== RUN   TestOnce
--- PASS: TestOnce
ok	so/sync	4 tests
```

`so test` exits non-zero if any test fails. The `=== RUN` line is printed before each test, so if a test hard-crashes (a `panic` or a segfault, which cannot be recovered), the output still identifies the culprit.

The report goes to `fmt.Output`, which is the standard output of the host by default. The runner reads `fmt.Output` once, before the first test, so a test that assigns another writer changes its own output only.

### Running many packages

A pattern that ends with `...` selects every package with a `test` subdirectory below its base directory:

```
so test ./so/...
```

The selected packages become one program, so the whole run costs one translate, one compile and one execution. This is much faster than one run per package, but the packages share a process: a hard crash in one package stops the packages after it.

The report keeps the packages apart:

```
so/sync
=== RUN   TestCond
--- PASS: TestCond
...
ok	so/sync	4 tests
so/time
=== RUN   TestNow
--- PASS: TestNow
...
ok	so/time	11 tests
```

### Running a subset

The `-run` flag limits the run to tests whose names start with a given prefix. Unlike Go's `-run`, it is a plain prefix match, not a regexp:

```
so test -run=TestMutex ./so/sync
```

```
=== RUN   TestMutex_LockUnlock
--- PASS: TestMutex_LockUnlock
=== RUN   TestMutex_TryLock
--- PASS: TestMutex_TryLock
ok	so/sync	2 tests
```

Here, `-run=TestMutex` runs all tests that start with `TestMutex`, while `-run=TestMutex_TryLock` runs only that specific test. If a prefix doesn't match any test, no tests will run.

To limit the run to a group of packages, narrow the pattern. `so test ./so/net/...` runs the tests of `so/net` and `so/net/netip`, and nothing else.

For a set of packages a pattern cannot describe, list them in a file and pass it with `-pkg-file`:

```
# freestanding.txt
so/bytes
so/io
so/strconv
```

```
so test -pkg-file=freestanding.txt ./so/...
```

The file holds one package per line, as a path relative to the module root. Blank lines and the text after a `#` are ignored. The list is a filter over the packages the pattern selects, so a listed package that the pattern does not select is an error. `-pkg-file` and `-run` combine: the file selects the packages, the prefix selects the tests.

## How it works

`so test`:

1. Finds the `test` directories the pattern selects, and scans each one for `TestXxx(t *testing.T)` functions.
2. Generates a runner that dispatches every package through `testing.RunSuites`. No file is written: the runner goes to the Go loader as an in-memory overlay, in a `.sotest` directory of the module root that never exists on disk.
3. Compiles and runs the generated program with the equivalent of `so run`.

To keep the C instead of running it, use `so translate-test`:

```
so translate-test -o out ./so/...
```

This is what a target the host cannot run needs: translate the test program here, then compile and run the C there. See [freestanding mode](freestanding.md).

`so translate-test` takes the same `-pkg-file` and `-run` flags as `so test`:

```
so translate-test -pkg-file=freestanding.txt -run=TestBuffer -o out ./so/...
```

Both flags apply when the runner is generated, so the translated program needs no argument of its own. `-pkg-file` keeps the listed packages out of the program. `-run` writes the prefix into the runner, which still links every test function but runs the matching ones alone.

## Benchmarks

Benchmarks work just like tests, but live in a `bench` subdirectory and are run
with `so bench`:

```
so/bytes/
    bytes.go
    ...
    bench/
        bytes.go       # your benchmarks
        bytes_test.go  # go benchmarks (optional)
```

Like a test directory, a bench directory declares a package of its own, not
`package main`. A common convention is to name it after the package under test,
with a `_bench` suffix: `so/bytes/bench` declares `package bytes_bench`,
`so/sync/bench` declares `package sync_bench`, etc.

A benchmark is a function named `BenchmarkXxx` taking a single `*testing.B`.
The measured code goes in a `for b.Loop()` loop; setup before the loop and
cleanup after it are not timed:

```go
package bytes_bench

import (
	"solod.dev/so/bytes"
	"solod.dev/so/testing"
)

func BenchmarkEqual(b *testing.B) {
	x := []byte("hello world")
	for b.Loop() {
		bytes.Equal(x, x)
	}
}
```

Run them with:

```
so bench ./so/bytes
```

```
BenchmarkEqual  482547852           2.215 ns/op         0 B/op         0 allocs/op
```

To measure allocations, allocate through `b.Allocator()`; the benchmark tracks
memory routed through it and reports `B/op` and `allocs/op`.

Like `so test`, `so bench` takes a `-run` flag that limits the run to benchmarks
whose names start with a given prefix (plain prefix match, not a regexp):

```
so bench -run=BenchmarkEqual ./so/bytes
```

`so bench` mirrors `so test`: it scans the `bench` directory for
`BenchmarkXxx(b *testing.B)` functions, generates a runner that dispatches them
through `testing.RunBenchmarks`, then compiles and runs it. The runner is an
in-memory overlay like the test one, so no file is written. One run measures
one package: `so bench` takes no `...` pattern.

The generated runner always uses the system allocator (`mem.System`). If a
benchmark needs a different allocator, write a main package of your own that
imports the bench package and calls `testing.RunBenchmarks`, then run it with
`so run` instead of `so bench`:

```go
func main() {
	opts := testing.Options{}
	testing.RunBenchmarks(alloc, "mypkg", opts, []testing.Benchmark{
		{Name: "BenchmarkEqual", F: mypkg_bench.BenchmarkEqual},
	})
}
```

`Options` selects the benchmarks to run, the same way the `-run` flag of
`so bench` does.

### Keeping the measured code alive

Unlike Go, So's `b.Loop` doesn't automatically keep the loop body alive. If you
compile the benchmarks to C with aggressive optimization (`-Ofast -flto`), the
compiler can remove any work whose result isn't used. A clear sign of this is
seeing an unrealistically low time, like `0.3 ns/op`, for something that should
take longer.

If the code returns a value, assign it to a package-level `//so:volatile` sink:

```go
//so:volatile
var sink int

func BenchmarkGet(b *testing.B) {
	m := buildMap()
	for b.Loop() {
		sink = m.Get(42)
	}
}
```

For work with no result to consume, such as a method with no return value, pass
the object's address to `testing.Keep`. It is a compiler barrier that emits no
instructions:

```go
func BenchmarkStore(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		x.Store(1)
		testing.Keep(&x)
	}
}
```

### Comparing against Go

`so bench` (and `so test`) ignore `_test.go` files. This lets you drop native
Go benchmarks of the same code into the `bench` directory and compare the two
side by side: `so bench ./so/bufio` runs the So versions, while
`go test -bench=. ./so/bufio/bench` runs the Go ones. Give the Go functions
distinct names (e.g. a `_Go` suffix) so both sets can share the directory
without colliding.

## Caveats

### Fatal and Skip need an explicit return

So has no `recover`, so a test cannot be unwound from the outside. `Fatal` and `Skip` only set state and print the message; they do **not** stop the function. Always `return` right after:

```go
if err != nil {
	t.Fatal("open failed")
	return
}
```

Use `Fatal` when continuing makes no sense (a precondition failed), and `Error` when you want to report several problems from one test.

### Freestanding runs report differently

The `testing` package works in [freestanding mode](freestanding.md), but the environment there gives it less to work with. The report goes to `fmt.Output`, which drops the bytes unless the target assigns a writer, and a failed run traps instead of exiting with a non-zero status. The runner also returns the heap to the position it holds before the first test, after every test, because the freestanding allocator never reclaims memory. The allocations made before the first test stay, so a package-level variable that allocates is safe, but a test cannot pass an allocation to the test after it. A test of your own can still need a hosted environment, even though the framework does not.

### A hard crash aborts the whole run

Tests run in a single process, and So has no `recover`. So a hard crash — a `panic` or a segfault — in one test aborts the entire run, not just that test: the tests after it do not execute. With a `...` pattern the run covers many packages, so a crash also stops the packages after the crashing one. Reported failures (`Error`, `Fatal`) are unaffected; only an actual crash stops the run. The `=== RUN` line printed before each test tells you which one crashed.
