// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/runtime"
	"solod.dev/so/strings"
	"solod.dev/so/time"
)

// This is a simple benchmarking framework based on Go's standard library.
// https://github.com/golang/go/blob/go1.26.1/src/testing/benchmark.go

// BenchTime specifies the amount of time to run a benchmark for,
// or the number of iterations to run.
type BenchTime struct {
	d time.Duration
	n int
}

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	N         int           // The number of iterations.
	T         time.Duration // The total time taken.
	Bytes     int64         // Bytes processed in one iteration.
	MemAllocs uint64        // The total number of memory allocations.
	MemBytes  uint64        // The total number of bytes allocated.
}

// B is used to manage benchmark timing and control the number of iterations.
type B struct {
	name     string    // Name of test or benchmark.
	start    time.Time // Time test or benchmark started
	duration time.Duration

	w        io.Writer // Output writer for test or benchmark.
	failed   bool      // Test or benchmark has failed.
	skipped  bool      // Test or benchmark has been skipped.
	finished bool      // Test function has completed.

	N         int
	benchFunc func(b *B)
	benchTime BenchTime
	bytes     int64
	timerOn   bool
	result    BenchmarkResult

	a mem.Tracker
	// The initial states of memStats.Mallocs and memStats.TotalAlloc.
	startAllocs uint64
	startBytes  uint64
	// The net total of this test after being run.
	netAllocs uint64
	netBytes  uint64

	// loop tracks the state of B.Loop
	loop struct {
		// n is the target number of iterations. It gets bumped up as we go.
		// When the benchmark loop is done, we commit this to b.N so users can
		// do reporting based on it, but we avoid exposing it until then.
		n uint64
		// i is the current Loop iteration. It's strictly monotonically
		// increasing toward n.
		i uint64

		done bool // set when B.Loop return false
	}
}

// Allocator returns the allocator used by the benchmark.
// The benchmarking function should use this allocator if it wants
// to track memory allocations and report them in the benchmark results.
func (b *B) Allocator() mem.Allocator {
	return &b.a
}

// StartTimer starts timing a test. This function is called automatically
// before a benchmark starts, but it can also be used to resume timing after
// a call to [B.StopTimer].
func (b *B) StartTimer() {
	if !b.timerOn {
		stats := b.a.Stats()
		b.startAllocs = stats.Mallocs
		b.startBytes = stats.TotalAlloc
		b.start = time.Now()
		b.timerOn = true
	}
}

// StopTimer stops timing a test. This can be used to pause the timer
// while performing steps that you don't want to measure.
func (b *B) StopTimer() {
	if b.timerOn {
		b.duration += time.Since(b.start)
		stats := b.a.Stats()
		b.netAllocs += stats.Mallocs - b.startAllocs
		b.netBytes += stats.TotalAlloc - b.startBytes
		b.timerOn = false
	}
}

// ResetTimer zeroes the elapsed benchmark time and memory allocation counters
// and deletes user-reported metrics.
// It does not affect whether the timer is running.
func (b *B) ResetTimer() {
	if b.timerOn {
		stats := b.a.Stats()
		b.startAllocs = stats.Mallocs
		b.startBytes = stats.TotalAlloc
		b.start = time.Now()
	}
	b.duration = 0
	b.netAllocs = 0
	b.netBytes = 0
}

// SetBytes records the number of bytes processed in a single operation.
// If this is called, the benchmark will report ns/op and MB/s.
func (b *B) SetBytes(n int64) { b.bytes = n }

// runN runs a single benchmark for the specified number of iterations.
func (b *B) runN(n int) {
	b.N = n
	b.loop.n = 0
	b.loop.i = 0
	b.loop.done = false

	b.ResetTimer()
	b.StartTimer()
	b.benchFunc(b)
	b.StopTimer()

	if b.loop.n > 0 && !b.loop.done && !b.failed {
		panic("benchmark function returned without B.Loop() == false (break or return in loop?)")
	}
}

// run1 runs the first iteration of benchFunc. It reports whether more
// iterations of this benchmarks should be run.
func (b *B) run1() bool {
	b.runN(1)
	if b.failed {
		fmt.Fprintf(b.w, "--- FAIL: %s\n", b.name)
		return false
	}

	if b.finished {
		tag := "BENCH"
		if b.skipped {
			tag = "SKIP"
		}
		if b.finished {
			fmt.Fprintf(b.w, "--- %s: %s\n", tag, b.name)
		}
		return false
	}
	return true
}

// run executes the benchmark. It gradually increases the number
// of benchmark iterations until the benchmark runs for the requested benchtime.
// run1 must have been called on b before calling run.
func (b *B) run() {
	// b.Loop does its own ramp-up logic so we just need to run it once.
	// If b.loop.n is non zero, it means b.Loop has already run.
	if b.loop.n == 0 {
		// Run the benchmark for at least the specified amount of time.
		if b.benchTime.n > 0 {
			// We already ran a single iteration in run1.
			// If -benchtime=1x was requested, use that result.
			// See https://golang.org/issue/32051.
			if b.benchTime.n > 1 {
				b.runN(b.benchTime.n)
			}
		} else {
			d := b.benchTime.d
			for n := int64(1); !b.failed && b.duration < d && n < 1e9; {
				last := n
				// Predict required iterations.
				goalns := d.Nanoseconds()
				prevIters := int64(b.N)
				n = int64(predictN(goalns, prevIters, b.duration.Nanoseconds(), last))
				b.runN(int(n))
			}
		}
	}
	b.result = BenchmarkResult{b.N, b.duration, b.bytes, b.netAllocs, b.netBytes}
}

// Don't run more than 1e9 times.
const maxBenchPredictIters = 1_000_000_000

func predictN(goalns int64, prevIters int64, prevns int64, last int64) int {
	if prevns == 0 {
		// Round up to dodge divide by zero. See https://go.dev/issue/70709.
		prevns = 1
	}

	// Order of operations matters.
	// For very fast benchmarks, prevIters ~= prevns.
	// If you divide first, you get 0 or 1,
	// which can hide an order of magnitude in execution time.
	// So multiply first, then divide.
	n := goalns * prevIters / prevns
	// Run more iterations than we think we'll need (1.2x).
	n += n / 5
	// Don't grow too fast in case we had timing errors previously.
	n = min(n, 100*last)
	// Be sure to run at least one more than last time.
	n = max(n, last+1)
	// Don't run more than 1e9 times. (This also keeps n in int range on 32 bit platforms.)
	n = min(n, maxBenchPredictIters)
	return int(n)
}

// Elapsed returns the measured elapsed time of the benchmark.
// The duration reported by Elapsed matches the one measured by
// [B.StartTimer], [B.StopTimer], and [B.ResetTimer].
func (b *B) Elapsed() time.Duration {
	d := b.duration
	if b.timerOn {
		d += time.Since(b.start)
	}
	return d
}

func (b *B) stopOrScaleBLoop() bool {
	t := b.Elapsed()
	if t >= b.benchTime.d {
		// We've reached the target
		return false
	}
	// Loop scaling
	goalns := b.benchTime.d.Nanoseconds()
	prevIters := int64(b.loop.n)
	b.loop.n = uint64(predictN(goalns, prevIters, t.Nanoseconds(), prevIters))
	// predictN may have capped the number of iterations; make sure to
	// terminate if we've already hit that cap.
	return uint64(prevIters) < b.loop.n
}

func (b *B) loopSlowPath() bool {
	// Consistency checks
	if !b.timerOn {
		panic("B.Loop called with timer stopped")
	}

	if b.loop.n == 0 {
		// It's the first call to b.Loop() in the benchmark function.
		if b.benchTime.n > 0 {
			// Fixed iteration count.
			b.loop.n = uint64(b.benchTime.n)
		} else {
			// Initialize target to 1 to kick start loop scaling.
			b.loop.n = 1
		}
		// Within a b.Loop loop, we don't use b.N (to avoid confusion).
		b.N = 0
		b.ResetTimer()

		// Start the next iteration.
		b.loop.i++
		return true
	}

	// Should we keep iterating?
	var more bool
	if b.benchTime.n > 0 {
		// The iteration count is fixed, so we should have run this many and now
		// be done.
		if b.loop.i != uint64(b.benchTime.n) {
			// We shouldn't be able to reach the slow path in this case.
			panic("iteration count < fixed target")
		}
		more = false
	} else {
		// Handle fixed time case
		more = b.stopOrScaleBLoop()
	}
	if !more {
		b.StopTimer()
		// Commit iteration count
		b.N = int(b.loop.n)
		b.loop.done = true
		return false
	}

	// Start the next iteration.
	b.loop.i++
	return true
}

// Loop returns true as long as the benchmark should continue running.
//
// A typical benchmark is structured like:
//
//	func Benchmark(b *testing.B) {
//		... setup ...
//		for b.Loop() {
//			... code to measure ...
//		}
//		... cleanup ...
//	}
//
// Loop resets the benchmark timer the first time it is called in a benchmark,
// so any setup performed prior to starting the benchmark loop does not count
// toward the benchmark measurement. Likewise, when it returns false, it stops
// the timer so cleanup code is not measured.
//
// Unlike Go, Solod does not automatically keep the loop body's values alive, so
// the C optimizer may delete measured work whose result is unused (for example
// a void method, or a computation whose return value is discarded). Assign the
// result to a package-level //so:volatile sink, or pass the value's address to
// [Keep], to force the work to run:
//
//	for b.Loop() {
//		x.Store(1)
//		testing.Keep(&x)
//	}
//
// After Loop returns false, b.N contains the total number of iterations that
// ran, so the benchmark may use b.N to compute other average metrics.
//
// Prior to the introduction of Loop, benchmarks were expected to contain an
// explicit loop from 0 to b.N. Benchmarks should either use Loop or contain a
// loop to b.N, but not both. Loop offers more automatic management of the
// benchmark timer, and runs each benchmark function only once per measurement,
// whereas b.N-based benchmarks must run the benchmark function (and any
// associated setup and cleanup) several times.
func (b *B) Loop() bool {
	// This is written such that the fast path is as fast as possible and can be
	// inlined.
	//
	// There are three cases where we'll fall out of the fast path:
	//
	// - On the first call, both i and n are 0.
	//
	// - If the loop reaches the n'th iteration, then i == n and we need
	//   to figure out the new target iteration count or if we're done.
	//
	// - If the timer is stopped, it poisons the top bit of i so the slow
	//   path can do consistency checks and fail.
	if b.loop.i < b.loop.n {
		b.loop.i++
		return true
	}
	return b.loopSlowPath()
}

// NsPerOp returns the "ns/op" metric.
func (r BenchmarkResult) NsPerOp() int64 {
	if r.N <= 0 {
		return 0
	}
	return r.T.Nanoseconds() / int64(r.N)
}

// mbPerSec returns the "MB/s" metric.
func (r BenchmarkResult) mbPerSec() float64 {
	if r.Bytes <= 0 || r.T <= 0 || r.N <= 0 {
		return 0
	}
	return (float64(r.Bytes) * float64(r.N) / 1e6) / r.T.Seconds()
}

// AllocsPerOp returns the "allocs/op" metric,
// which is calculated as r.MemAllocs / r.N.
func (r BenchmarkResult) AllocsPerOp() int64 {
	if r.N <= 0 {
		return 0
	}
	return int64(r.MemAllocs) / int64(r.N)
}

// AllocedBytesPerOp returns the "B/op" metric,
// which is calculated as r.MemBytes / r.N.
func (r BenchmarkResult) AllocedBytesPerOp() int64 {
	if r.N <= 0 {
		return 0
	}
	return int64(r.MemBytes) / int64(r.N)
}

// String returns a summary of the benchmark results.
// It follows the benchmark result line format from
// https://golang.org/design/14313-benchmark-format, not including the
// benchmark name.
// Extra metrics override built-in metrics of the same name.
// String does not include allocs/op or B/op, since those are reported
// by [BenchmarkResult.MemString].
func (r BenchmarkResult) String(buf []byte) string {
	sb := strings.FixedBuilder(buf)
	fmt.Fprintf(&sb, "%8d", r.N)

	// Get ns/op as a float.
	ns := float64(r.T.Nanoseconds()) / float64(r.N)
	if ns != 0 {
		sb.WriteString("  ")
		prettyPrint(&sb, ns, "ns/op")
	}

	if mbs := r.mbPerSec(); mbs != 0 {
		fmt.Fprintf(&sb, "  %7.2f MB/s", mbs)
	}

	return sb.String()
}

func prettyPrint(w io.Writer, x float64, unit string) {
	// Print all numbers with 10 places before the decimal point
	// and small numbers with four sig figs. Field widths are
	// chosen to fit the whole part in 10 places while aligning
	// the decimal point of all fractional formats.
	var format string
	switch y := floatAbs(x); {
	case y == 0 || y >= 999.95:
		format = "%10.0f %s"
	case y >= 99.995:
		format = "%12.1f %s"
	case y >= 9.9995:
		format = "%13.2f %s"
	case y >= 0.99995:
		format = "%14.3f %s"
	case y >= 0.099995:
		format = "%15.4f %s"
	case y >= 0.0099995:
		format = "%16.5f %s"
	case y >= 0.00099995:
		format = "%17.6f %s"
	default:
		format = "%18.7f %s"
	}
	fmt.Fprintf(w, format, x, unit)
}

// MemString returns r.AllocedBytesPerOp and r.AllocsPerOp in the same format as 'go test'.
func (r BenchmarkResult) MemString(buf []byte) string {
	return fmt.Sprintf(buf, "%8d B/op  %8d allocs/op",
		r.AllocedBytesPerOp(), r.AllocsPerOp())
}

// Benchmark represents a single benchmark to be run by the benchmark runner.
type Benchmark struct {
	Name string
	F    func(b *B)
}

// BenchmarkFunc is a function that benchmarks a piece of code.
type BenchmarkFunc func(b *B)

// RunBenchmarks runs the given benchmarks for package pkg with allocator a,
// prints the results, and exits with a non-zero status if any benchmark failed.
func RunBenchmarks(a mem.Allocator, pkg string, opts Options, benchmarks []Benchmark) {
	w := fmt.Output
	fmt.Fprintf(w, "goos: %s\n", runtime.GOOS)
	fmt.Fprintf(w, "goarch: %s\n", runtime.GOARCH)
	fmt.Fprintf(w, "pkg: %s\n", pkg)

	failed := 0
	total := 0
	for _, bench := range benchmarks {
		if !strings.HasPrefix(bench.Name, opts.Run) {
			continue
		}
		total++

		b := &B{
			name:      bench.Name,
			a:         mem.Tracker{Allocator: a},
			w:         w,
			benchFunc: bench.F,
			benchTime: BenchTime{d: 1 * time.Second},
		}
		if b.run1() {
			b.run()
		}
		if b.failed {
			fmt.Fprintf(b.w, "--- FAIL: %s\n", b.name)
			failed++
			continue
		}
		var buf [1024]byte
		resStr := b.result.String(buf[:])
		memStr := b.result.MemString(buf[len(resStr):])
		fmt.Fprintf(b.w, "%s  %s  %s\n", b.name, resStr, memStr)
	}

	if total == 0 {
		fmt.Fprintf(w, "ok\t%s\t%d benchmarks [no benchmarks to run]\n", pkg, total)
		return
	}
	if failed > 0 {
		fmt.Fprintf(w, "FAIL\t%s\t%d of %d failed\n", pkg, failed, total)
		exitFail()
	}
	fmt.Fprintf(w, "ok\t%s\t%d benchmarks\n", pkg, total)
}

// RunBenchmark benchmarks a single function and returns the results.
func RunBenchmark(a mem.Allocator, f func(b *B)) BenchmarkResult {
	benchTime := BenchTime{d: 1 * time.Second}
	b := &B{
		a:         mem.Tracker{Allocator: a},
		w:         io.Discard,
		benchFunc: f,
		benchTime: benchTime,
	}
	if b.run1() {
		b.run()
	}
	return b.result
}
