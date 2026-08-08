package sync_bench

import (
	"sync"
	"testing"
)

func BenchmarkOnceUncontended_Go(b *testing.B) {
	var once sync.Once
	for b.Loop() {
		once.Do(noop)
	}
}

func BenchmarkOnceContended_Go(b *testing.B) {
	var once sync.Once
	once.Do(noop) // mark done, so the workers exercise the fast path

	task := func() {
		for range numLoops {
			once.Do(noop)
		}
	}

	p := newPool(numWorkers)
	defer p.Free()

	for b.Loop() {
		for range numWorkers {
			p.Go(task)
		}
		p.Wait()
	}
}
