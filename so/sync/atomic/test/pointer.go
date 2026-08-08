package atomic_test

import (
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

type node struct {
	val int
}

func TestPointer(t *testing.T) {
	var a atomic.Pointer[node]

	if a.Load() != nil {
		t.Fatal("zero value must load nil")
		return
	}
	n1 := node{val: 1}
	a.Store(&n1)
	if a.Load().val != 1 {
		t.Error("store failed")
	}
	n2 := node{val: 2}
	old := a.Swap(&n2)
	if old.val != 1 {
		t.Error("swap must return old pointer")
	}
	if a.Load().val != 2 {
		t.Error("swap must set new pointer")
	}
	if !a.CompareAndSwap(&n2, &n1) {
		t.Error("cas must succeed on match")
	}
	if a.CompareAndSwap(&n2, &n1) {
		t.Error("cas must fail on mismatch")
	}
	if a.Load().val != 1 {
		t.Error("cas set wrong pointer")
	}
}
