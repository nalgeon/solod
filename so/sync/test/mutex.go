package sync_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// counter is a shared count guarded by a mutex.
type counter struct {
	mu  *sync.Mutex
	val *int
}

func bump(arg any) {
	c := arg.(*counter)
	c.mu.Lock()
	*c.val = *c.val + 1
	c.mu.Unlock()
}

func TestMutex_LockUnlock(t *testing.T) {
	// Check that no updates are lost when many workers
	// concurrently increment a shared counter under a mutex.
	const n = 1000
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	val := 0
	jobs := make([]counter, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(t.Allocator(), opts)
	for i := range jobs {
		jobs[i].mu = &mu
		jobs[i].val = &val
		p.Go(bump, &jobs[i])
	}
	p.Free()

	if val != n {
		t.Fatal("lost updates under mutex")
	}
}

// tryLockJob carries the mutex to try and a slot for the result.
type tryLockJob struct {
	mu *sync.Mutex
	ok *bool
}

// tryLock calls TryLock and records the result. It unlocks the
// mutex again if it acquired the lock.
func tryLock(arg any) any {
	job := arg.(*tryLockJob)
	*job.ok = job.mu.TryLock()
	if *job.ok {
		job.mu.Unlock()
	}
	return nil
}

func TestMutex_TryLock(t *testing.T) {
	// Check that TryLock acquires a free mutex and refuses to acquire one that is
	// already held. A second thread does the check, because POSIX leaves a relock
	// from the owning thread undefined.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	ok := false
	job := tryLockJob{mu: &mu, ok: &ok}

	mu.Lock()
	thr := conc.Go(tryLock, &job)
	thr.Wait()
	if ok {
		t.Fatal("TryLock succeeded on held mutex")
		return
	}
	mu.Unlock()

	thr = conc.Go(tryLock, &job)
	thr.Wait()
	if !ok {
		t.Fatal("TryLock failed on free mutex")
		return
	}

	// the worker unlocked the mutex, so the main thread can lock it again
	if !mu.TryLock() {
		t.Fatal("TryLock failed after the worker unlocked")
		return
	}
	mu.Unlock()
}

// lockJob carries the mutex to lock, the guarded flag,
// and a slot for the flag as the worker saw it.
type lockJob struct {
	mu    *sync.Mutex
	ready *bool
	seen  *bool
}

// lockAndRead locks the mutex and records the guarded flag.
func lockAndRead(arg any) any {
	job := arg.(*lockJob)
	job.mu.Lock()
	*job.seen = *job.ready
	job.mu.Unlock()
	return nil
}

func TestMutex_LockBlocks(t *testing.T) {
	// Check that Lock blocks while another thread holds the mutex. The worker
	// must not enter the critical section before the main thread leaves it, so
	// the worker must see the flag the main thread sets there.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	ready := false
	seen := false
	job := lockJob{mu: &mu, ready: &ready, seen: &seen}

	mu.Lock()
	thr := conc.Go(lockAndRead, &job)
	time.Sleep(10 * time.Millisecond)
	ready = true
	mu.Unlock()
	thr.Wait()

	if !seen {
		t.Fatal("Lock did not block while the mutex was held")
	}
}
