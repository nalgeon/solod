package sync_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// gate coordinates waiters with the main thread through a condition
// variable and a shared ready flag.
type gate struct {
	mu    *sync.Mutex
	cond  *sync.Cond
	nwait *int  // number of threads blocked in Wait
	ready *bool // the condition the waiters wait for
	woke  *bool // set by a waiter after it wakes
}

// waiter blocks until the gate is ready, then records that it woke up.
func waiter(arg any) any {
	g := arg.(*gate)
	g.mu.Lock()
	*g.nwait++
	for !*g.ready {
		g.cond.Wait()
	}
	*g.woke = true
	g.mu.Unlock()
	return nil
}

// signaler makes the gate ready and wakes one waiter.
func signaler(arg any) any {
	g := arg.(*gate)
	g.mu.Lock()
	*g.ready = true
	g.cond.Signal()
	g.mu.Unlock()
	return nil
}

// blockUntilWaiting returns after n threads are blocked in Wait. The mutex is
// locked on return, so the caller can change the condition and signal it
// without a waiter missing the change.
func blockUntilWaiting(mu *sync.Mutex, nwait *int, n int) {
	for {
		mu.Lock()
		if *nwait == n {
			return
		}
		mu.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func TestCond_Signal(t *testing.T) {
	// Start a worker that waits on a condition variable, then check
	// that Signal wakes it. The worker unlocks the mutex after Wait returns,
	// so a free mutex afterward shows that Wait re-locked it.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	var cond sync.Cond
	cond.Init(&mu)
	defer cond.Free()

	nwait := 0
	ready := false
	woke := false

	g := gate{mu: &mu, cond: &cond, nwait: &nwait, ready: &ready, woke: &woke}
	thr := conc.Go(waiter, &g)

	blockUntilWaiting(&mu, &nwait, 1)
	ready = true
	cond.Signal()
	mu.Unlock()

	thr.Wait()

	if !woke {
		t.Fatal("waiter did not observe the signal")
		return
	}
	if !mu.TryLock() {
		t.Fatal("Wait did not re-lock the mutex")
		return
	}
	mu.Unlock()
}

func TestCond_Broadcast(t *testing.T) {
	// Start several workers that wait on one condition variable,
	// then check that Broadcast wakes every one of them.
	const n = 4
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	var cond sync.Cond
	cond.Init(&mu)
	defer cond.Free()

	nwait := 0
	ready := false
	woke := make([]bool, n)
	gates := make([]gate, n)
	threads := make([]conc.Thread, n)
	for i := range gates {
		gates[i].mu = &mu
		gates[i].cond = &cond
		gates[i].nwait = &nwait
		gates[i].ready = &ready
		gates[i].woke = &woke[i]
		threads[i] = conc.Go(waiter, &gates[i])
	}

	blockUntilWaiting(&mu, &nwait, n)
	ready = true
	cond.Broadcast()
	mu.Unlock()

	for i := range threads {
		threads[i].Wait()
	}
	for i := range woke {
		if !woke[i] {
			t.Fatalf("waiter %d did not observe the broadcast", i)
			return
		}
	}
}

func TestCond_WaitFor(t *testing.T) {
	// Check that WaitFor gives up when nothing signals the condition,
	// and that it waits for the requested time before it gives up.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	var cond sync.Cond
	cond.Init(&mu)
	defer cond.Free()

	const d = 20 * time.Millisecond
	start := time.Now()

	mu.Lock()
	// a spurious wakeup returns from WaitFor without a signal,
	// so wait again for the time that is left
	timedOut := false
	for !timedOut {
		timedOut = cond.WaitFor(int64(d - time.Since(start)))
	}
	mu.Unlock()

	elapsed := time.Since(start)
	if elapsed < d {
		t.Errorf("WaitFor returned after %d ns, want at least %d ns",
			elapsed.Nanoseconds(), int64(d))
	}
}

func TestCond_WaitForSignal(t *testing.T) {
	// Check that WaitFor returns without a timeout when
	// another thread signals the condition before the deadline.
	const timeout = 5 * time.Second
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	var cond sync.Cond
	cond.Init(&mu)
	defer cond.Free()

	nwait := 0
	ready := false
	woke := false

	g := gate{mu: &mu, cond: &cond, nwait: &nwait, ready: &ready, woke: &woke}

	// the worker blocks on the mutex until WaitFor releases it
	mu.Lock()
	thr := conc.Go(signaler, &g)

	start := time.Now()
	timedOut := false
	for !ready && !timedOut {
		timedOut = cond.WaitFor(int64(timeout - time.Since(start)))
	}
	mu.Unlock()
	thr.Wait()

	if timedOut {
		t.Fatal("WaitFor timed out before the signal")
		return
	}
	if !ready {
		t.Fatal("WaitFor returned before the condition was true")
	}
}

func TestCond_WaitForZero(t *testing.T) {
	// Check that a non-positive timeout makes WaitFor
	// report a timeout at once, without blocking.
	var mu sync.Mutex
	mu.Init()
	defer mu.Free()

	var cond sync.Cond
	cond.Init(&mu)
	defer cond.Free()

	mu.Lock()
	zero := cond.WaitFor(0)
	neg := cond.WaitFor(-int64(time.Second))
	mu.Unlock()

	if !zero {
		t.Error("WaitFor(0) did not time out")
	}
	if !neg {
		t.Error("WaitFor(-1s) did not time out")
	}
}
