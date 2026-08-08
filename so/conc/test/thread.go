package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
)

func increment(arg any) any {
	n := arg.(*int)
	*n = *n + 1
	return arg
}

// Starts a thread per element, waits for them all, and checks every result.
func TestThread_Wait(t *testing.T) {
	const n = 16
	nums := make([]int, n)
	threads := make([]conc.Thread, n)
	for i := range nums {
		nums[i] = i
		threads[i] = conc.Go(increment, &nums[i])
	}

	ok := true
	for i := range threads {
		res := threads[i].Wait()
		if *(res.(*int)) != i+1 {
			ok = false
		}
	}
	for i := range nums {
		if nums[i] != i+1 {
			ok = false
		}
	}
	if !ok {
		t.Error("wrong increment result")
	}
}

// latch lets a detached thread report completion, since it cannot be joined.
type latch struct {
	mu   sync.Mutex
	cond sync.Cond
	done bool
	out  int
}

// squareLatch squares l.out in place, then marks the latch done.
func squareLatch(arg any) any {
	l := arg.(*latch)
	l.mu.Lock()
	l.out = l.out * l.out
	l.done = true
	l.cond.Broadcast()
	l.mu.Unlock()
	return nil
}

// Runs a task on a detached thread and waits for it through a condition.
func TestThread_Detach(t *testing.T) {
	var l latch
	l.mu.Init()
	defer l.mu.Free()
	l.cond.Init(&l.mu)
	defer l.cond.Free()
	l.out = 9

	th := conc.Go(squareLatch, &l)
	th.Detach()

	l.mu.Lock()
	for !l.done {
		l.cond.Wait()
	}
	l.mu.Unlock()

	if l.out != 81 {
		t.Error("wrong detached result")
	}
}
