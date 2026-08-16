package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/math"
	"solod.dev/so/mem"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

func addOne(arg any) {
	cnt := arg.(*atomic.Int64)
	cnt.Add(1)
}

func TestInt64_Concurrent(t *testing.T) {
	// Checks that no updates are lost when many workers concurrently
	// increment a shared atomic counter without a mutex.
	const n = 1000
	var cnt atomic.Int64
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(mem.System, opts)
	for range n {
		p.Go(addOne, &cnt)
	}
	p.Free()

	if cnt.Load() != n {
		t.Error("lost updates under atomic add")
	}
}

func TestInt64_SwapCAS(t *testing.T) {
	// Checks single-threaded Load/Store/Add/Swap/CompareAndSwap semantics.
	var a atomic.Int64

	if a.Load() != 0 {
		t.Error("zero value must load 0")
	}
	a.Store(10)
	if a.Load() != 10 {
		t.Error("store failed")
	}
	if a.Add(5) != 15 {
		t.Error("add must return new value")
	}
	if a.Swap(20) != 15 {
		t.Error("swap must return old value")
	}
	if a.Load() != 20 {
		t.Error("swap must set new value")
	}
	if !a.CompareAndSwap(20, 30) {
		t.Error("cas must succeed on match")
	}
	if a.CompareAndSwap(20, 40) {
		t.Error("cas must fail on mismatch")
	}
	if a.Load() != 30 {
		t.Error("cas set wrong value")
	}
}

func TestInt64_Negative(t *testing.T) {
	// Checks negative values and an add that crosses zero.
	var a atomic.Int64

	a.Store(-1)
	if a.Load() != -1 {
		t.Error("store negative failed")
	}
	if a.Add(-4) != -5 {
		t.Error("add of a negative delta must return new value")
	}
	if a.Add(10) != 5 {
		t.Error("add across zero must return new value")
	}
	if a.Swap(-7) != 5 {
		t.Error("swap must return old value")
	}
	if a.Load() != -7 {
		t.Error("swap must set new value")
	}
	if !a.CompareAndSwap(-7, -8) {
		t.Error("cas must succeed on negative match")
	}
	if a.CompareAndSwap(-7, 0) {
		t.Error("cas must fail on negative mismatch")
	}
	if a.Load() != -8 {
		t.Error("failed cas must not change value")
	}
}

func TestInt64_Limits(t *testing.T) {
	// Checks the values at the limits of the type. Add wraps like
	// the Go operator: past the maximum it continues at the minimum.
	var a atomic.Int64

	a.Store(math.MaxInt64)
	if a.Load() != math.MaxInt64 {
		t.Error("store max failed")
	}
	if a.Add(1) != math.MinInt64 {
		t.Error("add past the maximum must wrap to the minimum")
	}
	if a.Add(-1) != math.MaxInt64 {
		t.Error("add past the minimum must wrap to the maximum")
	}
	if !a.CompareAndSwap(math.MaxInt64, math.MinInt64) {
		t.Error("cas must succeed on max")
	}
	if a.Swap(0) != math.MinInt64 {
		t.Error("swap must return the minimum")
	}
	if a.Add(0) != 0 {
		t.Error("add of a zero delta must not change value")
	}
}
