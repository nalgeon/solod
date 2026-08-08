package sync_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
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

// Checks that no updates are lost when many workers
// concurrently increment a shared counter under a mutex.
func TestMutex_LockUnlock(t *testing.T) {
	const n = 1000
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	val := 0
	jobs := make([]counter, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(mem.System, opts)
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

// Checks that TryLock acquires a free mutex and refuses
// to acquire one that is already held.
func TestMutex_TryLock(t *testing.T) {
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	if !mu.TryLock() {
		t.Fatal("TryLock failed on free mutex")
		return
	}
	if mu.TryLock() {
		t.Fatal("TryLock succeeded on held mutex")
		return
	}
	mu.Unlock()

	if !mu.TryLock() {
		t.Fatal("TryLock failed after unlock")
		return
	}
	mu.Unlock()
}
