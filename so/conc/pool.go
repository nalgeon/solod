package conc

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/runtime"
	"solod.dev/so/sync"
)

// task is one unit of work: a type-erased body plus a pointer to its argument.
type task struct {
	fn  func(any)
	arg any
}

// Pool is a bounded pool of worker threads with a wait queue
// which execute tasks of the form func(any).
//
// A task is a named func(any), conventionally called with a pointer to its
// argument struct. The worker calls the function with that pointer, and the
// function reads inputs and writes results in place.
//
// If all workers are busy, submitting a task puts it in a wait queue until
// a worker picks it up. If the queue is full, submitting a task blocks until
// a slot frees.
//
// Tasks in a pool must be independent: a task must not block waiting on
// another task in the same pool, or the pool can deadlock. By convention,
// if a task can fail, it should report the error in an err field on its
// argument struct.
type Pool struct {
	alloc mem.Allocator

	mu       sync.Mutex
	notEmpty sync.Cond // signaled when a task is enqueued
	notFull  sync.Cond // signaled when a slot frees
	allDone  sync.Cond // broadcast when no task is in flight

	workers []Thread

	// queue is a ring buffer of tasks waiting for a worker.
	// qhead and qtail are indices into the queue,
	// and qcount is the number of tasks in the queue.
	queue  []task
	qhead  int
	qtail  int
	qcount int

	active  int  // number of tasks submitted but not yet finished
	stopped bool // true if the pool is shutting down
}

// PoolOptions holds the Pool settings. A zero PoolOptions selects defaults.
type PoolOptions struct {
	NumThreads int // number of worker threads; 0 = runtime.NumCPU
	QueueSize  int // task queue size; 0 = same as NumThreads
	StackSize  int // thread stack size in bytes; 0 = system default
}

// NewPool creates a pool with a given number of worker threads
// and starts them. A zero PoolOptions is valid and runs one worker per CPU.
// Call [Pool.Free] exactly once when done:
//
//	opts := conc.PoolOptions{NumThreads: 4}
//	p := conc.NewPool(mem.System, opts)
//	defer p.Free()
//	p.Go(work, &job1)
//	p.Go(work, &job2)
//	p.Wait()
func NewPool(alloc mem.Allocator, opts PoolOptions) *Pool {
	c.Assert(opts.NumThreads >= 0, "conc: NumThreads must be >= 0")
	c.Assert(opts.QueueSize >= 0, "conc: QueueSize must be >= 0")

	numThreads := opts.NumThreads
	if numThreads == 0 {
		numThreads = runtime.NumCPU()
	}

	queueSize := numThreads
	if opts.QueueSize > 0 {
		queueSize = opts.QueueSize
	}

	p := mem.Alloc[Pool](alloc)
	p.alloc = alloc
	p.queue = mem.AllocSlice[task](alloc, queueSize, queueSize)
	p.workers = mem.AllocSlice[Thread](alloc, numThreads, numThreads)
	p.qhead, p.qtail, p.qcount, p.active = 0, 0, 0, 0
	p.stopped = false

	p.mu.Init()
	p.notEmpty.Init(&p.mu)
	p.notFull.Init(&p.mu)
	p.allDone.Init(&p.mu)

	topts := ThreadOptions{StackSize: opts.StackSize}
	for i := range p.workers {
		p.workers[i] = GoWith(workerMain, any(p), topts)
	}
	return p
}

// Go submits a task for execution, blocking while the queue is full.
//
// The worker invokes fn with the given arg; arg must point to storage that
// outlives the task. If the queue is full, Go blocks until a slot frees.
//
// Unlike [Go], a pool task returns nothing: there is no join to retrieve a
// value. Write results into the arg struct in place, and report errors via
// an err field on it.
//
// Go is thread-safe.
func (p *Pool) Go(fn func(any), arg any) {
	p.mu.Lock()
	for p.qfull() {
		p.notFull.Wait()
	}
	p.qpush(task{fn: fn, arg: arg})
	p.active++
	p.notEmpty.Signal()
	p.mu.Unlock()
}

// Wait blocks until all submitted tasks finish. The pool stays usable
// afterward: submit more tasks and call Wait again. Call [Pool.Free]
// to release the pool's resources.
//
// Wait is thread-safe.
func (p *Pool) Wait() {
	p.mu.Lock()
	for p.active != 0 {
		p.allDone.Wait()
	}
	p.mu.Unlock()
}

// Free stops the pool and releases its resources. Any queued tasks are
// drained first: Free blocks until every submitted task finishes, then
// joins the workers. The pool is unusable afterward.
//
// Free should only be called once; it's not thread-safe.
func (p *Pool) Free() {
	p.mu.Lock()
	p.stopped = true
	p.notEmpty.Broadcast()
	p.mu.Unlock()

	for i := range p.workers {
		p.workers[i].Wait()
	}

	p.mu.Free()
	p.notEmpty.Free()
	p.notFull.Free()
	p.allDone.Free()
	mem.FreeSlice(p.alloc, p.queue)
	mem.FreeSlice(p.alloc, p.workers)
	mem.Free(p.alloc, p)
}

// qfull reports whether the task queue is at capacity.
func (p *Pool) qfull() bool { return p.qcount == len(p.queue) }

// qempty reports whether the task queue is empty.
func (p *Pool) qempty() bool { return p.qcount == 0 }

// qpush appends t to the tail of the task queue.
func (p *Pool) qpush(t task) {
	p.queue[p.qtail] = t
	p.qtail = (p.qtail + 1) % len(p.queue)
	p.qcount++
}

// qpop removes and returns the task at the head of the queue.
func (p *Pool) qpop() task {
	t := p.queue[p.qhead]
	p.qhead = (p.qhead + 1) % len(p.queue)
	p.qcount--
	return t
}

// workerMain is the pthread start routine: pop tasks
// and run them until the pool is stopped and drained.
func workerMain(arg any) any {
	p := arg.(*Pool)
	for {
		p.mu.Lock()
		for p.qempty() && !p.stopped {
			p.notEmpty.Wait()
		}
		if p.qempty() && p.stopped {
			p.mu.Unlock()
			break
		}
		t := p.qpop()
		p.notFull.Signal()
		p.mu.Unlock()

		t.fn(t.arg)

		p.mu.Lock()
		p.active--
		if p.active == 0 {
			p.allDone.Broadcast()
		}
		p.mu.Unlock()
	}
	return nil
}
