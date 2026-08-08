package sync_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
)

// noop is the do-nothing function passed to Once.Do: the benchmarks
// measure the Do machinery itself, not the work it guards.
func noop() {}

func BenchmarkOnceUncontended_So(b *testing.B) {
	// Measures the Once.Do hot path on a single thread.
	var once sync.Once
	once.Init()
	defer once.Free()
	for b.Loop() {
		once.Do(noop)
	}
}

func BenchmarkOnceContended_So(b *testing.B) {
	// Measures Once.Do under contention: numWorkers threads
	// hammer the same already-done Once.
	var once sync.Once
	once.Init()
	defer once.Free()
	once.Do(noop) // mark done, so the workers exercise the fast path

	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	for b.Loop() {
		for range numWorkers {
			p.Go(hammerOnce, &once)
		}
		p.Wait()
	}
}

// hammerOnce calls Do numLoops times on the shared, already-done Once.
func hammerOnce(arg any) {
	once := arg.(*sync.Once)
	for range numLoops {
		once.Do(noop)
	}
}
