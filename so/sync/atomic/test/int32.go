package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
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
	p := conc.NewPool(mem.System, opts)
	for range n {
		p.Go(addOne32, &cnt)
	}
	p.Free()

	if cnt.Load() != n {
		t.Error("lost updates under atomic add")
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
