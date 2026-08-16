package conc_test

import "solod.dev/so/sync"

// latch is a one-shot signal between threads. One thread calls [latch.Done],
// another blocks in [latch.Wait] until it does. A detached thread needs it to
// report completion, because it cannot be joined.
type latch struct {
	mu   sync.Mutex
	cond sync.Cond
	done bool
}

// Init prepares the latch. Call [latch.Free] exactly once when done.
func (l *latch) Init() {
	l.mu.Init()
	l.cond.Init(&l.mu)
	l.done = false
}

// Free releases the latch resources.
func (l *latch) Free() {
	l.cond.Free()
	l.mu.Free()
}

// Done marks the latch signaled and wakes every waiter.
func (l *latch) Done() {
	l.mu.Lock()
	l.done = true
	l.cond.Broadcast()
	l.mu.Unlock()
}

// Wait blocks until another thread calls Done.
func (l *latch) Wait() {
	l.mu.Lock()
	for !l.done {
		l.cond.Wait()
	}
	l.mu.Unlock()
}
