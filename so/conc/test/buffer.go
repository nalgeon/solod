package conc_test

import (
	"solod.dev/so/c"
	"solod.dev/so/conc"
	"solod.dev/so/testing"
)

func TestBuffer_SendRecv(t *testing.T) {
	// Fill and drain the ring buffer three times, which wraps the head and the
	// tail past the end of the storage, and checks the FIFO order every round.
	const size = 3
	buf := conc.NewBuffer(t.Allocator(), c.Sizeof[point](), size)
	defer buf.Free()

	for round := range 3 {
		for i := range size {
			p := point{x: round, y: i}
			buf.Send(&p)
		}
		for i := range size {
			var got point
			if !buf.Recv(&got) {
				t.Fatalf("round %d: Recv = false, want true", round)
				return
			}
			if got.x != round || got.y != i {
				t.Fatalf("round %d: Recv = {%d, %d}, want {%d, %d}",
					round, got.x, got.y, round, i)
				return
			}
		}
	}
}

func TestBuffer_CloseDrain(t *testing.T) {
	// Check that buffered values survive Close and are drained in order before
	// Recv reports the channel closed.
	buf := conc.NewBuffer(t.Allocator(), c.Sizeof[point](), 4)
	defer buf.Free()

	for i := range 3 {
		p := point{x: i, y: i * 2}
		buf.Send(&p)
	}
	buf.Close()

	count := 0
	var got point
	for buf.Recv(&got) {
		if got.x != count || got.y != count*2 {
			t.Fatalf("Recv = {%d, %d}, want {%d, %d}",
				got.x, got.y, count, count*2)
			return
		}
		count++
	}
	if count != 3 {
		t.Errorf("drained %d values, want 3", count)
	}
}

func TestBuffer_Timeout(t *testing.T) {
	// Exercise the non-blocking timed operations from a single thread, where the
	// outcomes are fully deterministic: sends fail once full, receives fail once
	// empty, and a drained closed channel reports Closed.
	buf := conc.NewBuffer(t.Allocator(), c.Sizeof[point](), 2)
	defer buf.Free()

	p := point{x: 1, y: 2}
	if buf.SendTimeout(&p, 0) != conc.Ok {
		t.Fatal("SendTimeout with room, want Ok")
		return
	}
	p.x, p.y = 3, 4
	if buf.SendTimeout(&p, 0) != conc.Ok {
		t.Fatal("SendTimeout with room, want Ok")
		return
	}
	if buf.SendTimeout(&p, 0) != conc.Timeout {
		t.Error("SendTimeout on a full buffer, want Timeout")
	}

	var got point
	if buf.RecvTimeout(&got, 0) != conc.Ok || got.x != 1 || got.y != 2 {
		t.Errorf("first RecvTimeout = {%d, %d}, want {1, 2}", got.x, got.y)
	}
	if buf.RecvTimeout(&got, 0) != conc.Ok || got.x != 3 || got.y != 4 {
		t.Errorf("second RecvTimeout = {%d, %d}, want {3, 4}", got.x, got.y)
	}
	if buf.RecvTimeout(&got, 0) != conc.Timeout {
		t.Error("RecvTimeout on an empty buffer, want Timeout")
	}

	buf.Close()
	if buf.RecvTimeout(&got, 0) != conc.Closed {
		t.Error("RecvTimeout on a closed buffer, want Closed")
	}
}
