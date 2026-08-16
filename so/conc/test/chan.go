package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// point is a multi-field value the channel and engine tests hand off, so that
// the element size differs from the size of an int.
type point struct {
	x int
	y int
}

func TestChan_Buffered(t *testing.T) {
	// Fill a buffered channel without blocking
	// and checks that values come back in FIFO order.
	ch := conc.NewChan[int](t.Allocator(), 4)
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

func TestChan_ProducerConsumer(t *testing.T) {
	// Send 0..n-1 from the main thread through a small buffered channel
	// while a worker thread sums them, exercising back-pressure.
	const n = 1000
	task := sumTask{ch: conc.NewChan[int](t.Allocator(), 8), sum: 0}
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

func TestChan_Unbuffered(t *testing.T) {
	// Receive from an unbuffered channel fed by a worker thread
	// and check the handoff order.
	task := seqTask{ch: conc.NewChan[int](t.Allocator(), 0), n: 10}
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

// produceRange sends base..base+n-1 to the channel, as a pool task.
func produceRange(arg any) {
	task := arg.(*rangeTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(task.base + i)
	}
}

// produceRangeThread sends base..base+n-1 to the channel, as a thread entry.
func produceRangeThread(arg any) any {
	produceRange(arg)
	return nil
}

func TestChan_UnbufferedMultiProducer(t *testing.T) {
	// Run several producer threads sending on a single unbuffered channel while
	// the main thread receives. Each value 0..N-1 is sent exactly once across
	// producers; the receiver checks none is lost or duplicated. This exercises
	// the rendezvous handshake with concurrent senders.
	const producers = 4
	const perProducer = 250
	const total = producers * perProducer

	ch := conc.NewChan[int](t.Allocator(), 0)
	defer ch.Free()
	opts := conc.PoolOptions{NumThreads: producers}
	p := conc.NewPool(t.Allocator(), opts)

	tasks := make([]rangeTask, producers)
	for i := range tasks {
		tasks[i] = rangeTask{ch: ch, base: i * perProducer, n: perProducer}
		p.Go(produceRange, &tasks[i])
	}

	seen := make([]bool, total)
	ok := true
	var v int
	for range total {
		// Never leave this loop early. A producer blocked in Send keeps
		// p.Free below from returning, which hangs the test instead of
		// failing it.
		if !ch.Recv(&v) {
			ok = false
			continue
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

// countTask carries a channel, the shared table of received values, and the
// number of values one consumer took out of the channel.
type countTask struct {
	ch    conc.Chan[int]
	seen  []bool
	count int
}

// consumeSeen receives until the channel is closed and marks every value in
// the shared table. Each value is sent once, so the consumers write distinct
// elements of the table.
func consumeSeen(arg any) any {
	task := arg.(*countTask)
	var v int
	for task.ch.Recv(&v) {
		if v >= 0 && v < len(task.seen) {
			task.seen[v] = true
		}
		task.count++
	}
	return nil
}

func TestChan_BufferedMultiProducer(t *testing.T) {
	// Run several producers and several consumers on one buffered channel. Each
	// value 0..N-1 is sent exactly once, so the consumers together must take out
	// exactly those values, with none lost and none duplicated.
	const producers = 4
	const consumers = 3
	const perProducer = 250
	const total = producers * perProducer

	ch := conc.NewChan[int](t.Allocator(), 8)
	defer ch.Free()

	seen := make([]bool, total)
	cons := make([]countTask, consumers)
	consThreads := make([]conc.Thread, consumers)
	for i := range cons {
		cons[i] = countTask{ch: ch, seen: seen, count: 0}
		consThreads[i] = conc.Go(consumeSeen, &cons[i])
	}

	prod := make([]rangeTask, producers)
	prodThreads := make([]conc.Thread, producers)
	for i := range prod {
		prod[i] = rangeTask{ch: ch, base: i * perProducer, n: perProducer}
		prodThreads[i] = conc.Go(produceRangeThread, &prod[i])
	}

	for i := range prodThreads {
		prodThreads[i].Wait()
	}
	ch.Close()
	for i := range consThreads {
		consThreads[i].Wait()
	}

	got := 0
	for i := range cons {
		got += cons[i].count
	}
	if got != total {
		t.Errorf("consumers took %d values, want %d", got, total)
	}
	for i := range seen {
		if !seen[i] {
			t.Fatalf("value %d never arrived", i)
			return
		}
	}
}

func TestChan_CloseDrain(t *testing.T) {
	// Check that buffered values survive Close and are drained in order
	// before Recv reports the channel closed.
	ch := conc.NewChan[int](t.Allocator(), 4)
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

func TestChan_CloseUnbuffered(t *testing.T) {
	// Check that a receive on a closed unbuffered channel reports the channel
	// closed. An unbuffered channel holds nothing, so there is nothing to drain.
	ch := conc.NewChan[int](t.Allocator(), 0)
	defer ch.Free()
	ch.Close()

	var v int
	if ch.Recv(&v) {
		t.Error("Recv on a closed unbuffered chan = true, want false")
	}
}

// parkTask carries a channel, the signal that the receiver is about to park,
// and the result of the parked receive.
type parkTask struct {
	ch      conc.Chan[int]
	ready   latch
	gotOpen bool // the result of Recv: true means a value arrived
}

// recvParked signals that it is about to receive, then blocks in Recv.
func recvParked(arg any) any {
	task := arg.(*parkTask)
	task.ready.Done()
	var v int
	task.gotOpen = task.ch.Recv(&v)
	return nil
}

func TestChan_CloseWakesReceiver(t *testing.T) {
	// Check that Close wakes a receiver parked in Recv, for both engines.
	checkCloseWakesReceiver(t, 4)
	checkCloseWakesReceiver(t, 0)
}

func checkCloseWakesReceiver(t *testing.T, size int) {
	var task parkTask
	task.ch = conc.NewChan[int](t.Allocator(), size)
	defer task.ch.Free()
	task.ready.Init()
	defer task.ready.Free()

	thr := conc.Go(recvParked, &task)
	task.ready.Wait()
	// Give the receiver time to park. A close before it parks is correct too:
	// Recv on a closed empty channel reports the channel closed either way.
	time.Sleep(10 * time.Millisecond)
	task.ch.Close()
	thr.Wait()

	if task.gotOpen {
		t.Errorf("Recv after Close on a chan of size %d = true, want false", size)
	}
}

func TestChan_SendCopies(t *testing.T) {
	// Check that Send copies its argument: a write to the source after Send must
	// not change the value the receiver takes out.
	ch := conc.NewChan[point](t.Allocator(), 2)
	defer ch.Free()

	p := point{x: 1, y: 2}
	ch.Send(p)
	p.x = 30
	p.y = 40
	ch.Send(p)

	var got point
	if !ch.Recv(&got) || got.x != 1 || got.y != 2 {
		t.Errorf("first Recv = {%d, %d}, want {1, 2}", got.x, got.y)
	}
	if !ch.Recv(&got) || got.x != 30 || got.y != 40 {
		t.Errorf("second Recv = {%d, %d}, want {30, 40}", got.x, got.y)
	}
}

// pointTask for handing a point over an unbuffered channel.
type pointTask struct {
	ch conc.Chan[point]
	n  int
}

// producePoints sends n points to the channel and then closes it.
func producePoints(arg any) any {
	task := arg.(*pointTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(point{x: i, y: i * 2})
	}
	task.ch.Close()
	return nil
}

func TestChan_StructUnbuffered(t *testing.T) {
	task := pointTask{ch: conc.NewChan[point](t.Allocator(), 0), n: 10}
	defer task.ch.Free()

	thr := conc.Go(producePoints, &task)
	count := 0
	var got point
	for task.ch.Recv(&got) {
		if got.x != count || got.y != count*2 {
			t.Fatalf("Recv = {%d, %d}, want {%d, %d}", got.x, got.y, count, count*2)
			return
		}
		count++
	}
	thr.Wait()

	if count != 10 {
		t.Errorf("received %d points, want 10", count)
	}
}

func TestChan_Pointer(t *testing.T) {
	ch := conc.NewChan[*point](t.Allocator(), 2)
	defer ch.Free()

	p := point{x: 5, y: 6}
	ch.Send(&p)

	var got *point
	if !ch.Recv(&got) {
		t.Fatal("Recv on a chan of pointers = false, want true")
		return
	}
	if got != &p {
		t.Fatal("Recv on a chan of pointers returned another pointer")
		return
	}
	// The pointer refers to the sent value, so a later write is visible.
	p.x = 50
	if got.x != 50 {
		t.Error("the received pointer does not refer to the sent value")
	}
}

func TestChan_Byte(t *testing.T) {
	ch := conc.NewChan[byte](t.Allocator(), 4)
	defer ch.Free()

	for i := range 4 {
		ch.Send(byte(i + 1))
	}
	var v byte
	for i := range 4 {
		if !ch.Recv(&v) || v != byte(i+1) {
			t.Fatalf("Recv = %d, want %d", v, i+1)
			return
		}
	}
}
