package conc_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

// chanCap is the buffer size of the uncontended channel.
// It holds a full batch so the fill loop never blocks.
const chanCap = 100

// chanBatch is the number of values pushed through the channel per timed
// iteration in the producer/consumer benchmarks. It is large enough to amortize
// the one-time thread setup so the measurement reflects steady-state handoff
// throughput.
const chanBatch = 1000

//so:volatile
var chanSink int

func BenchmarkChanUncontended_So(b *testing.B) {
	// Fills then drains a buffered channel from a single thread. Nobody ever
	// blocks, so this measures the bare send/recv cost (lock plus copy) with
	// no contention or wakeup.
	ch := conc.NewChan[int](mem.System, chanCap)
	defer ch.Free()

	var v int
	for b.Loop() {
		for range chanCap {
			ch.Send(0)
		}
		for range chanCap {
			ch.Recv(&v)
		}
		testing.Keep(&v)
	}
}

func BenchmarkChanProdCons0_So(b *testing.B) {
	benchChanProdCons_So(b, 0)
}

func BenchmarkChanProdCons10_So(b *testing.B) {
	benchChanProdCons_So(b, 10)
}

func BenchmarkChanProdCons100_So(b *testing.B) {
	benchChanProdCons_So(b, 100)
}

// benchChanProdCons measures single-producer/single-consumer handoff through a
// channel of the given buffer size. size == 0 exercises the unbuffered
// rendezvous engine (every send waits for the receiver); size > 0 exercises the
// buffered engine, with a smaller buffer forcing more blocking.
func benchChanProdCons_So(b *testing.B, size int) {
	// One consumer thread is started once and drains the channel for the whole run;
	// each timed iteration pushes chanBatch values through it.
	task := chanTask{ch: conc.NewChan[int](mem.System, size)}
	thr := conc.Go(chanDrain, &task)

	for b.Loop() {
		for i := range chanBatch {
			task.ch.Send(i)
		}
	}

	task.ch.Close()
	thr.Wait()
	task.ch.Free()
	chanSink = task.sum
}

// chanTask carries the channel and the consumer's running sum between
// the producer (main thread) and the consumer thread.
type chanTask struct {
	ch  conc.Chan[int]
	sum int
}

// chanDrain receives values until the channel is closed, accumulating them.
// The sum is only there to keep the received values live.
func chanDrain(arg any) any {
	task := arg.(*chanTask)
	var v int
	for task.ch.Recv(&v) {
		task.sum += v
	}
	return nil
}
