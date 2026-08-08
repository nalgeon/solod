package atomic_test

import (
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
