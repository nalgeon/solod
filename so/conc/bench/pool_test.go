package conc_bench

import (
	"sync"
	"testing"
	"time"
)

func BenchmarkPoolCPU_Go(b *testing.B) {
	p := newPool(numWorkers, numWorkers)
	defer p.Free()

	args := make([]workArg, numCPUTasks)
	for i := range args {
		args[i].n = numCPUIter
	}

	for b.Loop() {
		for i := range numCPUTasks {
			p.Go(cpuWork, &args[i])
		}
		p.Wait()
	}
}

func BenchmarkPoolIO_Go(b *testing.B) {
	p := newPool(numWorkers, numWorkers)
	defer p.Free()

	ioWork := func(arg any) {
		time.Sleep(time.Millisecond) // matches ioLatency
	}

	for b.Loop() {
		for range numIOTasks {
			p.Go(ioWork, &ioSink)
		}
		p.Wait()
	}
}

// task is one unit of work: a type-erased body plus a pointer to its argument.
type task struct {
	fn  func(any)
	arg any
}

// pool is a hand-written Go worker pool that mirrors conc.Pool: a fixed set of
// persistent worker goroutines draining a buffered task channel, with the same
// func(any)+arg task shape and a batch Wait. It keeps the comparison focused on
// the dispatch machinery (buffered channel vs mutex+cond queue) rather than on
// API differences.
type pool struct {
	tasks chan task
	wg    sync.WaitGroup
}

// newPool starts numThreads workers draining a queue of queueSize.
func newPool(numThreads, queueSize int) *pool {
	p := &pool{tasks: make(chan task, queueSize)}
	for range numThreads {
		go func() {
			for t := range p.tasks {
				t.fn(t.arg)
				p.wg.Done()
			}
		}()
	}
	return p
}

// Go submits a task, blocking while the queue is full.
func (p *pool) Go(fn func(any), arg any) {
	p.wg.Add(1)
	p.tasks <- task{fn, arg}
}

// Wait blocks until all submitted tasks finish. The pool stays usable after.
func (p *pool) Wait() { p.wg.Wait() }

// Free stops the workers.
func (p *pool) Free() { close(p.tasks) }
