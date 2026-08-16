package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

func TestChan_TimeoutBuffered(t *testing.T) {
	// Exercise non-blocking SendTimeout/RecvTimeout (d == 0) on a buffered channel
	// from a single thread, where the outcomes are fully deterministic: sends fail
	// once full, receives fail once empty, and a drained closed channel reports
	// Closed.
	ch := conc.NewChan[int](t.Allocator(), 2)
	defer ch.Free()

	// The buffer holds 2; the third non-blocking send must time out.
	if ch.SendTimeout(10, 0) != conc.Ok || ch.SendTimeout(20, 0) != conc.Ok {
		t.Fatal("SendTimeout should succeed with room")
		return
	}
	if ch.SendTimeout(30, 0) != conc.Timeout {
		t.Error("SendTimeout should time out when full")
	}

	// Drain in FIFO order, then a non-blocking receive must time out.
	var v int
	if ch.RecvTimeout(&v, 0) != conc.Ok || v != 10 {
		t.Fatal("wrong first RecvTimeout value")
		return
	}
	if ch.RecvTimeout(&v, 0) != conc.Ok || v != 20 {
		t.Fatal("wrong second RecvTimeout value")
		return
	}
	if ch.RecvTimeout(&v, 0) != conc.Timeout {
		t.Error("RecvTimeout should time out when empty")
	}

	// After close with no buffered values, a receive reports Closed.
	ch.Close()
	if ch.RecvTimeout(&v, 0) != conc.Closed {
		t.Error("RecvTimeout should report Closed")
	}
}

func TestChan_TimeoutExpires(t *testing.T) {
	// Check that timed operations actually give up at the deadline when no peer
	// ever appears: both a send and a receive on an idle unbuffered channel must
	// return Timeout rather than block forever.
	ch := conc.NewChan[int](t.Allocator(), 0)
	defer ch.Free()

	if ch.SendTimeout(1, 10*time.Millisecond) != conc.Timeout {
		t.Error("SendTimeout should time out with no receiver")
	}
	var v int
	if ch.RecvTimeout(&v, 10*time.Millisecond) != conc.Timeout {
		t.Error("RecvTimeout should time out with no sender")
	}
}

func TestChan_TimeoutSendClosed(t *testing.T) {
	// Check that SendTimeout on a closed channel reports Closed for both engines.
	// Send panics in the same situation, so only the timed send is testable.
	checkSendTimeoutClosed(t, 2)
	checkSendTimeoutClosed(t, 0)
}

func checkSendTimeoutClosed(t *testing.T, size int) {
	ch := conc.NewChan[int](t.Allocator(), size)
	defer ch.Free()
	ch.Close()

	if st := ch.SendTimeout(1, 10*time.Millisecond); st != conc.Closed {
		t.Errorf("SendTimeout on a closed chan of size %d = %d, want Closed",
			size, int(st))
	}
}

func TestChan_TimeoutNegative(t *testing.T) {
	// Check that a negative duration is non-blocking, exactly like a zero one.
	const d = -1 * time.Millisecond

	ch := conc.NewChan[int](t.Allocator(), 1)
	defer ch.Free()
	var v int
	if ch.RecvTimeout(&v, d) != conc.Timeout {
		t.Error("RecvTimeout(negative) on an empty buffer, want Timeout")
	}
	if ch.SendTimeout(1, d) != conc.Ok {
		t.Error("SendTimeout(negative) with room, want Ok")
	}
	if ch.SendTimeout(2, d) != conc.Timeout {
		t.Error("SendTimeout(negative) on a full buffer, want Timeout")
	}
	if ch.RecvTimeout(&v, d) != conc.Ok || v != 1 {
		t.Error("RecvTimeout(negative) with a buffered value, want Ok")
	}

	rdv := conc.NewChan[int](t.Allocator(), 0)
	defer rdv.Free()
	if rdv.SendTimeout(1, d) != conc.Timeout {
		t.Error("SendTimeout(negative) with no receiver, want Timeout")
	}
	if rdv.RecvTimeout(&v, d) != conc.Timeout {
		t.Error("RecvTimeout(negative) with no sender, want Timeout")
	}
}

func TestChan_TimeoutHandoff(t *testing.T) {
	// Receive from an unbuffered channel with a deadline while a worker thread
	// feeds it with blocking sends. The loop tolerates timeouts and stops on
	// Closed, checking the handoff order.
	task := seqTask{ch: conc.NewChan[int](t.Allocator(), 0), n: 10}
	defer task.ch.Free()

	thr := conc.Go(produceSeq, &task)
	want := 0
	ordered := true
	var v int
	for {
		st := task.ch.RecvTimeout(&v, 50*time.Millisecond)
		if st == conc.Closed {
			break
		}
		if st == conc.Timeout {
			continue // no sender ready yet; keep polling
		}
		if v != want {
			ordered = false
		}
		want++
	}
	thr.Wait()

	if !ordered {
		t.Error("wrong timeout handoff order")
	}
	if want != 10 {
		t.Error("missing timeout handoff values")
	}
}

func TestChan_TimeoutSend(t *testing.T) {
	// Send on an unbuffered channel with a deadline while a worker thread drains
	// it with blocking receives. The deadline is short, so most sends give up
	// before a receiver arrives and the loop retries them.
	const n = 100
	task := sumTask{ch: conc.NewChan[int](t.Allocator(), 0), sum: 0}
	defer task.ch.Free()

	thr := conc.Go(consume, &task)
	for i := range n {
		for task.ch.SendTimeout(i, 100*time.Microsecond) != conc.Ok {
			// No receiver ready yet; keep retrying.
		}
	}
	task.ch.Close()
	thr.Wait()

	// Sum of 0..99.
	if task.sum != 4950 {
		t.Error("wrong timeout send sum")
	}
}
