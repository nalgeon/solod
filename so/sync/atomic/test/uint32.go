package atomic_test

import (
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

func TestUint32(t *testing.T) {
	var a atomic.Uint32

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
}
