package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

func TestBool(t *testing.T) {
	var a atomic.Bool

	if a.Load() {
		t.Fatal("zero value must load false")
		return
	}
	a.Store(true)
	if !a.Load() {
		t.Error("store true failed")
	}
	if a.Swap(false) != true {
		t.Error("swap must return old value")
	}
	if a.Load() {
		t.Error("swap must set new value")
	}
	if !a.CompareAndSwap(false, true) {
		t.Error("cas must succeed on match")
	}
	if a.CompareAndSwap(false, false) {
		t.Error("cas must fail on mismatch")
	}
	if !a.Load() {
		t.Error("cas set wrong value")
	}
}

// claim is one worker's attempt to set the shared flag.
type claim struct {
	flag *atomic.Bool
	won  bool
}

func claimFlag(arg any) {
	c := arg.(*claim)
	c.won = c.flag.CompareAndSwap(false, true)
}

func TestBool_Concurrent(t *testing.T) {
	// Checks that exactly one of many workers wins the race
	// to set a shared flag with CompareAndSwap.
	const n = 1000
	var flag atomic.Bool
	claims := make([]claim, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(t.Allocator(), opts)
	for i := range claims {
		claims[i].flag = &flag
		p.Go(claimFlag, &claims[i])
	}
	p.Free()

	won := 0
	for i := range claims {
		if claims[i].won {
			won++
		}
	}
	if won != 1 {
		t.Error("cas must succeed for exactly one worker")
	}
	if !flag.Load() {
		t.Error("flag must be set after the race")
	}
}
