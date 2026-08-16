package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/math"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

func TestUint64(t *testing.T) {
	var a atomic.Uint64

	if a.Load() != 0 {
		t.Error("zero value must load 0")
	}
	a.Store(100)
	if a.Add(23) != 123 {
		t.Error("add must return new value")
	}
	if a.Sub(23) != 100 {
		t.Error("sub must return new value")
	}
	if a.Swap(7) != 100 {
		t.Error("swap must return old value")
	}
	if !a.CompareAndSwap(7, 9) {
		t.Error("cas must succeed on match")
	}
	if a.Load() != 9 {
		t.Error("cas set wrong value")
	}
	if a.CompareAndSwap(7, 11) {
		t.Error("cas must fail on mismatch")
	}
	if a.Load() != 9 {
		t.Error("failed cas must not change value")
	}
}

func TestUint64_Limits(t *testing.T) {
	// Checks the values at the limits of the type. Add and Sub wrap
	// like the Go operators, and a zero delta changes nothing.
	var a atomic.Uint64

	if a.Sub(0) != 0 {
		t.Error("sub of a zero delta must not change value")
	}
	if a.Sub(1) != math.MaxUint64 {
		t.Error("sub below zero must wrap to the maximum")
	}
	if a.Sub(0) != math.MaxUint64 {
		t.Error("sub of a zero delta must not change the maximum")
	}
	if a.Add(1) != 0 {
		t.Error("add past the maximum must wrap to zero")
	}
	a.Store(math.MaxUint64)
	if a.Load() != math.MaxUint64 {
		t.Error("store max failed")
	}
	if !a.CompareAndSwap(math.MaxUint64, 1) {
		t.Error("cas must succeed on max")
	}
	if a.Sub(2) != math.MaxUint64 {
		t.Error("sub across zero must wrap to the maximum")
	}
}

// addSub64 is one worker's update of a shared counter.
// The worker adds one if add is set, and subtracts one if it is not.
type addSub64 struct {
	cnt *atomic.Uint64
	add bool
}

func addSubOne64(arg any) {
	j := arg.(*addSub64)
	if j.add {
		j.cnt.Add(1)
	} else {
		j.cnt.Sub(1)
	}
}

func TestUint64_Concurrent(t *testing.T) {
	// Checks that an equal number of concurrent adds and subs
	// leaves the counter at its starting value.
	const n = 1000
	var cnt atomic.Uint64
	cnt.Store(n)
	jobs := make([]addSub64, 2*n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(t.Allocator(), opts)
	for i := range jobs {
		jobs[i].cnt = &cnt
		jobs[i].add = i%2 == 0
		p.Go(addSubOne64, &jobs[i])
	}
	p.Free()

	if cnt.Load() != n {
		t.Error("lost updates under atomic add and sub")
	}
}
