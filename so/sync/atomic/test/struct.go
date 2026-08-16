package atomic_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync/atomic"
	"solod.dev/so/testing"
)

// bag mixes atomic and plain fields, and puts the wider atomic field
// after narrower fields. An atomic instruction needs its operand to be
// naturally aligned, so the layout of the struct must align big.
type bag struct {
	flag  bool
	small atomic.Int32
	big   atomic.Int64
	tail  byte
}

func bumpFields(arg any) {
	b := arg.(*bag)
	b.small.Add(1)
	b.big.Add(2)
}

func TestStructFields(t *testing.T) {
	// Checks atomic fields inside a struct: the updates are not lost,
	// and an update of one field leaves the other fields alone.
	const n = 1000
	var b bag
	b.flag = true
	b.tail = 7

	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(mem.System, opts)
	for range n {
		p.Go(bumpFields, &b)
	}
	p.Free()

	if b.small.Load() != n {
		t.Error("lost updates on the int32 field")
	}
	if b.big.Load() != 2*n {
		t.Error("lost updates on the int64 field")
	}
	if !b.flag {
		t.Error("atomic update changed the bool field")
	}
	if b.tail != 7 {
		t.Error("atomic update changed the byte field")
	}
}
