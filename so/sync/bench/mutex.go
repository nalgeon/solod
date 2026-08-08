package sync_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
)

// numWorkers is the number of threads contending for the mutex
// in the contended benchmark.
const numWorkers = 8

// numLoops is the number of Lock/Unlock rounds each worker performs per
// benchmark iteration. It is large enough to amortize the pool submission
// and thread-wakeup overhead so the measurement reflects lock contention.
const numLoops = 1000

// workIters is the number of xorshift rounds done while holding the lock in
// the ContendedWork benchmark - roughly a microsecond of work, standing in for
// a small but real critical section. That is long enough that the hold time
// exceeds the mutex's adaptive-spin window, so contending threads stop spinning
// and park in the kernel.
const workIters = 500

// sink absorbs the result of busy so the compiler cannot elide the work.
// It is only ever read and written while the benchmark mutex is held,
// so the concurrent accesses are serialized.
var sink uint64 = 1

func BenchmarkMutexUncontended_So(b *testing.B) {
	// Measures Lock/Unlock on a mutex that is never contended,
	// i.e. the fast path of the primitive.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkMutexTryLock_So(b *testing.B) {
	// Measures TryLock/Unlock on an uncontended mutex.
	// TryLock always succeeds here since nothing else holds the lock.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()
	for b.Loop() {
		if mu.TryLock() {
			mu.Unlock()
		}
	}
}

func BenchmarkMutexContendedSpin_So(b *testing.B) {
	// Measures Lock/Unlock under contention with an empty critical section:
	// numWorkers threads each hammer the same mutex. The hold time is near
	// zero, so a contending thread reacquires the lock while still spinning
	// and rarely parks in the kernel. This is the spin-friendly regime.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	for b.Loop() {
		for range numWorkers {
			p.Go(hammerMutexSpin, &mu)
		}
		p.Wait()
	}
}

func BenchmarkMutexContendedWork_So(b *testing.B) {
	// Measures Lock/work/Unlock under contention: numWorkers threads each hold
	// the mutex long enough (busy(workIters)) that contending threads exhaust
	// their spin budget and park in the kernel. Every handoff then costs a
	// wakeup syscall, unlike BenchmarkMutexContendedSpin.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	for b.Loop() {
		for range numWorkers {
			p.Go(hammerMutexWork, &mu)
		}
		p.Wait()
	}
}

// hammerMutexSpin locks and unlocks the shared mutex numLoops times.
func hammerMutexSpin(arg any) {
	mu := arg.(*sync.Mutex)
	for range numLoops {
		mu.Lock()
		mu.Unlock()
	}
}

// hammerMutexWork locks the shared mutex, does a fixed chunk of work while
// holding it, and unlocks - numLoops times.
func hammerMutexWork(arg any) {
	mu := arg.(*sync.Mutex)
	for range numLoops {
		mu.Lock()
		sink = busy(sink, workIters)
		mu.Unlock()
	}
}

// busy runs n xorshift rounds over x and returns the result. The dependency
// chain resists closed-form folding by the optimizer, so it is a genuine chunk
// of work standing in for a real critical section.
func busy(x uint64, n int) uint64 {
	for range n {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}
