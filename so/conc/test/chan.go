package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// Fills a buffered channel without blocking
// and checks that values come back in FIFO order.
func TestChan_Buffered(t *testing.T) {
	ch := conc.NewChan[int](mem.System, 4)
	defer ch.Free()

	for i := range 4 {
		ch.Send(i * 10)
	}
	var v int
	for i := range 4 {
		if !ch.Recv(&v) || v != i*10 {
			t.Fatal("wrong buffered value")
			return
		}
	}
}

// sumTask carries a channel and the resulting sum between threads.
type sumTask struct {
	ch  conc.Chan[int]
	sum int
}

// consume receives values until the channel is closed and accumulates them.
func consume(arg any) any {
	task := arg.(*sumTask)
	var v int
	for task.ch.Recv(&v) {
		task.sum += v
	}
	return nil
}

// Sends 0..n-1 from the main thread through a small buffered channel
// while a worker thread sums them, exercising back-pressure.
func TestChan_ProducerConsumer(t *testing.T) {
	const n = 1000
	task := sumTask{ch: conc.NewChan[int](mem.System, 8), sum: 0}
	defer task.ch.Free()

	thr := conc.Go(consume, &task)
	for i := range n {
		task.ch.Send(i)
	}
	task.ch.Close()
	thr.Wait()

	// Sum of 0..999.
	if task.sum != 499500 {
		t.Error("wrong producer/consumer sum")
	}
}

// seqTask for sending a sequence of values to a channel.
type seqTask struct {
	ch conc.Chan[int]
	n  int
}

// produceSeq sends 0..n-1 to the channel and then closes it.
func produceSeq(arg any) any {
	task := arg.(*seqTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(i)
	}
	task.ch.Close()
	return nil
}

// Receives from an unbuffered channel fed by a worker thread
// and checks the handoff order.
func TestChan_Unbuffered(t *testing.T) {
	task := seqTask{ch: conc.NewChan[int](mem.System, 0), n: 10}
	defer task.ch.Free()

	thr := conc.Go(produceSeq, &task)
	want := 0
	ordered := true
	var v int
	for task.ch.Recv(&v) {
		if v != want {
			ordered = false
		}
		want++
	}
	thr.Wait()

	if !ordered {
		t.Error("wrong unbuffered handoff order")
	}
	if want != 10 {
		t.Error("missing unbuffered values")
	}
}

// rangeTask for sending a range of values to a channel.
type rangeTask struct {
	ch   conc.Chan[int]
	base int
	n    int
}

// produceRange sends base..base+n-1 to the channel.
func produceRange(arg any) {
	task := arg.(*rangeTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(task.base + i)
	}
}

// Runs several producer threads sending on a single unbuffered channel while
// the main thread receives. Each value 0..N-1 is sent exactly once across
// producers; the receiver checks none is lost or duplicated. This exercises
// the rendezvous handshake with concurrent senders.
func TestChan_UnbufferedMultiProducer(t *testing.T) {
	const producers = 4
	const perProducer = 250
	const total = producers * perProducer

	ch := conc.NewChan[int](mem.System, 0)
	defer ch.Free()
	opts := conc.PoolOptions{NumThreads: producers}
	p := conc.NewPool(mem.System, opts)

	tasks := make([]rangeTask, producers)
	for i := range tasks {
		tasks[i] = rangeTask{ch: ch, base: i * perProducer, n: perProducer}
		p.Go(produceRange, &tasks[i])
	}

	seen := make([]bool, total)
	ok := true
	var v int
	for range total {
		if !ch.Recv(&v) {
			ok = false
			break
		}
		if v < 0 || v >= total || seen[v] {
			ok = false
			continue
		}
		seen[v] = true
	}
	p.Free()

	if !ok {
		t.Error("lost or duplicated unbuffered value")
	}
}

// Checks that buffered values survive Close and are drained in order
// before Recv reports the channel closed.
func TestChan_CloseDrain(t *testing.T) {
	ch := conc.NewChan[int](mem.System, 4)
	defer ch.Free()

	for i := 1; i <= 3; i++ {
		ch.Send(i)
	}
	ch.Close()

	seen := 0
	want := 1
	var v int
	for ch.Recv(&v) {
		if v != want {
			t.Fatal("wrong drained value")
			return
		}
		want++
		seen++
	}
	if seen != 3 {
		t.Error("did not drain all buffered values")
	}
}

// Exercises non-blocking SendTimeout/RecvTimeout (d == 0) on a buffered channel
// from a single thread, where the outcomes are fully deterministic: sends fail
// once full, receives fail once empty, and a drained closed channel reports
// Closed.
func TestChan_TimeoutBuffered(t *testing.T) {
	ch := conc.NewChan[int](mem.System, 2)
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

// Checks that timed operations actually give up at the deadline when no peer
// ever appears: both a send and a receive on an idle unbuffered channel must
// return Timeout rather than block forever.
func TestChan_TimeoutExpires(t *testing.T) {
	ch := conc.NewChan[int](mem.System, 0)
	defer ch.Free()

	if ch.SendTimeout(1, 10*time.Millisecond) != conc.Timeout {
		t.Error("SendTimeout should time out with no receiver")
	}
	var v int
	if ch.RecvTimeout(&v, 10*time.Millisecond) != conc.Timeout {
		t.Error("RecvTimeout should time out with no sender")
	}
}

// Receives from an unbuffered channel with a deadline while a worker thread
// feeds it with blocking sends. The loop tolerates timeouts and stops on
// Closed, checking the handoff order.
func TestChan_TimeoutHandoff(t *testing.T) {
	task := seqTask{ch: conc.NewChan[int](mem.System, 0), n: 10}
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

// Sends on an unbuffered channel with a deadline while a worker thread drains
// it with blocking receives. Each send retries until a receiver takes it.
func TestChan_TimeoutSend(t *testing.T) {
	const n = 100
	task := sumTask{ch: conc.NewChan[int](mem.System, 0), sum: 0}
	defer task.ch.Free()

	thr := conc.Go(consume, &task)
	for i := range n {
		for task.ch.SendTimeout(i, 50*time.Millisecond) != conc.Ok {
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
