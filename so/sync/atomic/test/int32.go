package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/math"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

func addOne32(arg any) {
	cnt := arg.(*atomic.Int32)
	cnt.Add(1)
}

func TestInt32_Concurrent(t *testing.T) {
	// Checks that no updates are lost when many workers concurrently
	// increment a shared atomic counter without a mutex.
	const n = 1000
	var cnt atomic.Int32
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(t.Allocator(), opts)
	for range n {
		p.Go(addOne32, &cnt)
	}
	p.Free()

	if cnt.Load() != n {
		t.Error("lost updates under atomic add")
	}
}

// swapJob is one worker's swap of a shared value.
// The worker writes id and reads the previous value into old.
type swapJob struct {
	v   *atomic.Int32
	id  int32
	old int32
}

func swapID(arg any) {
	j := arg.(*swapJob)
	j.old = j.v.Swap(j.id)
}

func TestInt32_ConcurrentSwap(t *testing.T) {
	// Checks that concurrent swaps neither lose nor duplicate a value.
	// The value starts at 0 and the ids are 1 to n, so the n old values
	// are the n+1 values 0 to n without the one left in v, each once.
	const n = 500
	var v atomic.Int32
	jobs := make([]swapJob, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(t.Allocator(), opts)
	for i := range jobs {
		jobs[i].v = &v
		jobs[i].id = int32(i + 1)
		p.Go(swapID, &jobs[i])
	}
	p.Free()

	seen := make([]bool, n+1)
	for i := range jobs {
		old := int(jobs[i].old)
		if old < 0 || old > n {
			t.Fatal("swap returned an unknown value")
			return
		}
		if seen[old] {
			t.Fatal("swap returned a duplicate value")
			return
		}
		seen[old] = true
	}
	if seen[int(v.Load())] {
		t.Error("swap returned the value left in v")
	}
}

func TestInt32_SwapCAS(t *testing.T) {
	// Checks single-threaded Load/Store/Add/Swap/CompareAndSwap semantics.
	var a atomic.Int32

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

func TestInt32_Negative(t *testing.T) {
	// Checks negative values and an add that crosses zero.
	var a atomic.Int32

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

func TestInt32_Limits(t *testing.T) {
	// Checks the values at the limits of the type. Add wraps like
	// the Go operator: past the maximum it continues at the minimum.
	var a atomic.Int32

	a.Store(math.MaxInt32)
	if a.Load() != math.MaxInt32 {
		t.Error("store max failed")
	}
	if a.Add(1) != math.MinInt32 {
		t.Error("add past the maximum must wrap to the minimum")
	}
	if a.Add(-1) != math.MaxInt32 {
		t.Error("add past the minimum must wrap to the maximum")
	}
	if !a.CompareAndSwap(math.MaxInt32, math.MinInt32) {
		t.Error("cas must succeed on max")
	}
	if a.Swap(0) != math.MinInt32 {
		t.Error("swap must return the minimum")
	}
	if a.Add(0) != 0 {
		t.Error("add of a zero delta must not change value")
	}
}
