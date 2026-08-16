package atomic_test

import (
	"solod.dev/so/conc"
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
	n3 := node{val: 3}
	if a.CompareAndSwap(&n2, &n3) {
		t.Error("cas must fail on mismatch")
	}
	if a.Load() != &n1 {
		t.Error("failed cas must not change the pointer")
	}
}

func TestPointer_Nil(t *testing.T) {
	// Checks that nil is an ordinary value: it is stored, swapped
	// and compared like any other pointer.
	var a atomic.Pointer[node]
	n1 := node{val: 1}

	if !a.CompareAndSwap(nil, &n1) {
		t.Error("cas from nil must succeed")
	}
	if a.Load() != &n1 {
		t.Error("cas from nil set wrong pointer")
	}
	if a.CompareAndSwap(nil, nil) {
		t.Error("cas from nil must fail on a non-nil value")
	}
	if a.Load() != &n1 {
		t.Error("failed cas must not change the pointer")
	}
	if !a.CompareAndSwap(&n1, nil) {
		t.Error("cas to nil must succeed")
	}
	if a.Load() != nil {
		t.Error("cas to nil set wrong pointer")
	}

	a.Store(&n1)
	a.Store(nil)
	if a.Load() != nil {
		t.Error("store nil failed")
	}
	a.Store(&n1)
	if a.Swap(nil) != &n1 {
		t.Error("swap to nil must return old pointer")
	}
	if a.Load() != nil {
		t.Error("swap to nil set wrong pointer")
	}
	if a.Swap(&n1) != nil {
		t.Error("swap from nil must return nil")
	}
}

func TestPointer_Types(t *testing.T) {
	// Checks two instantiations of Pointer in one program.
	// C erases the element type to void*, so each instantiation
	// must cast the loaded pointer back to its own type.
	var pn atomic.Pointer[node]
	var pi atomic.Pointer[int]

	n := node{val: 1}
	i := 42
	pn.Store(&n)
	pi.Store(&i)

	if pn.Load() != &n {
		t.Error("node pointer store failed")
	}
	if pi.Load() != &i {
		t.Error("int pointer store failed")
	}
	if pn.Load().val != 1 {
		t.Error("node pointer loaded wrong value")
	}
	if *pi.Load() != 42 {
		t.Error("int pointer loaded wrong value")
	}
	if pi.Swap(nil) != &i {
		t.Error("int pointer swap must return old pointer")
	}
	if pn.Load() != &n {
		t.Error("instantiations must not share storage")
	}
}

// publisher is one worker that repeatedly publishes its own node
// into a shared pointer and reads the pointer back.
type publisher struct {
	p     *atomic.Pointer[node]
	nodes []node
	idx   int
	bad   bool
}

func publishNode(arg any) {
	pb := arg.(*publisher)
	for range 100 {
		pb.p.Store(&pb.nodes[pb.idx])
		got := pb.p.Load()
		if got == nil {
			pb.bad = true
			return
		}
		// Every published node holds its own index, so a value
		// read back must point at the node its index names.
		if got.val < 0 || got.val >= len(pb.nodes) || got != &pb.nodes[got.val] {
			pb.bad = true
			return
		}
	}
}

func TestPointer_Concurrent(t *testing.T) {
	// Checks that a concurrently updated pointer always reads back
	// as one of the published pointers, never as a mix of two.
	const n = 64
	var p atomic.Pointer[node]
	nodes := make([]node, n)
	for i := range nodes {
		nodes[i].val = i
	}
	p.Store(&nodes[0])

	jobs := make([]publisher, n)
	opts := conc.PoolOptions{NumThreads: 8}
	pool := conc.NewPool(t.Allocator(), opts)
	for i := range jobs {
		jobs[i].p = &p
		jobs[i].nodes = nodes
		jobs[i].idx = i
		pool.Go(publishNode, &jobs[i])
	}
	pool.Free()

	for i := range jobs {
		if jobs[i].bad {
			t.Fatal("load returned a pointer that was never stored")
			return
		}
	}
}
