package atomic_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

//so:volatile
var sinkUint uint64

func BenchmarkAtomicLoad64_So(b *testing.B) {
	var x atomic.Uint64
	x.Store(42)
	for b.Loop() {
		sinkUint = x.Load()
	}
}

func BenchmarkAtomicStore64_So(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		x.Store(1)
		testing.Keep(&x)
	}
}

func BenchmarkAtomicAdd64_So(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		x.Add(1)
		testing.Keep(&x)
	}
}

func BenchmarkAtomicSwap64_So(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		sinkUint = x.Swap(1)
	}
}

func BenchmarkAtomicCAS64_So(b *testing.B) {
	var x atomic.Uint64
	x.Store(1)
	for b.Loop() {
		x.CompareAndSwap(1, 0)
		x.CompareAndSwap(0, 1)
		testing.Keep(&x)
	}
}

// numWorkers is the number of threads contending
// for the counter in the contended benchmark.
const numWorkers = 8

// numLoops is the number of Add rounds each worker performs per benchmark
// iteration. It is large enough to amortize the pool submission and
// thread-wakeup overhead so the measurement reflects the atomic contention.
const numLoops = 1000

func BenchmarkAtomicAddContended_So(b *testing.B) {
	// Measures Add under contention: numWorkers threads hammer the same
	// counter, the canonical lock-free-counter workload.
	var x atomic.Uint64

	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(b.Allocator(), opts)
	defer p.Free()

	for b.Loop() {
		for range numWorkers {
			p.Go(hammerAdd, &x)
		}
		p.Wait()
	}
}

// hammerAdd adds to the shared counter numLoops times.
func hammerAdd(arg any) {
	x := arg.(*atomic.Uint64)
	for range numLoops {
		x.Add(1)
	}
}
