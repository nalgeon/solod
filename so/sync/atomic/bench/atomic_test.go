package atomic_bench

import (
	"sync/atomic"
	"testing"
)

func BenchmarkAtomicLoad64_Go(b *testing.B) {
	var x atomic.Uint64
	x.Store(42)
	for b.Loop() {
		sinkUint = x.Load()
	}
}

func BenchmarkAtomicStore64_Go(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		x.Store(1)
	}
}

func BenchmarkAtomicAdd64_Go(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		x.Add(1)
	}
}

func BenchmarkAtomicSwap64_Go(b *testing.B) {
	var x atomic.Uint64
	for b.Loop() {
		sinkUint = x.Swap(1)
	}
}

func BenchmarkAtomicCAS64_Go(b *testing.B) {
	var x atomic.Uint64
	x.Store(1)
	for b.Loop() {
		x.CompareAndSwap(1, 0)
		x.CompareAndSwap(0, 1)
	}
}

func BenchmarkAtomicAddContended_Go(b *testing.B) {
	var x atomic.Uint64

	task := func() {
		for range numLoops {
			x.Add(1)
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

// pool is a fixed set of persistent worker goroutines. It mirrors Solod's conc.Pool
// so the contended benchmark is structurally equivalent on both sides: numWorkers
// goroutines stay alive for the whole benchmark and pick up tasks each iteration,
// instead of the benchmark spawning fresh goroutines every iteration (which would
// measure goroutine startup, not only the atomic contention).
type pool struct {
	tasks chan func()
	done  chan struct{}
	n     int // tasks submitted since the last Wait
}

// newPool starts n worker goroutines that run submitted tasks until Free.
func newPool(n int) *pool {
	p := &pool{tasks: make(chan func()), done: make(chan struct{})}
	for range n {
		go func() {
			for task := range p.tasks {
				task()
				p.done <- struct{}{}
			}
		}()
	}
	return p
}

// Go submits a task to the pool.
func (p *pool) Go(task func()) {
	p.n++
	p.tasks <- task
}

// Wait blocks until all tasks submitted since the last Wait finish.
func (p *pool) Wait() {
	for range p.n {
		<-p.done
	}
	p.n = 0
}

// Free stops the pool's workers.
func (p *pool) Free() { close(p.tasks) }
