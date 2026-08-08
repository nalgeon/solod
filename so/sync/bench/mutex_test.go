package sync_bench

import (
	"sync"
	"testing"
)

func BenchmarkMutexUncontended_Go(b *testing.B) {
	var mu sync.Mutex
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkMutexTryLock_Go(b *testing.B) {
	var mu sync.Mutex
	for b.Loop() {
		if mu.TryLock() {
			mu.Unlock()
		}
	}
}

func BenchmarkMutexContendedSpin_Go(b *testing.B) {
	var mu sync.Mutex
	task := func() {
		for range numLoops {
			mu.Lock()
			mu.Unlock()
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

func BenchmarkMutexContendedWork_Go(b *testing.B) {
	var mu sync.Mutex
	task := func() {
		for range numLoops {
			mu.Lock()
			sink = busy(sink, workIters)
			mu.Unlock()
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

// pool is a fixed set of persistent worker goroutines. It mirrors So's conc.Pool
// so the contended benchmark is structurally equivalent on both sides: numWorkers
// goroutines stay alive for the whole benchmark and pick up tasks each iteration,
// instead of the benchmark spawning fresh goroutines every iteration (which would
// measure goroutine startup, not only the lock contention).
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
