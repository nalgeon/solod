package conc_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// numWorkers is the number of worker threads in the benchmarked pool.
const numWorkers = 8

// numCPUTasks is the number of CPU-bound tasks submitted per iteration.
// It is large enough to amortize pool setup so the measurement reflects
// steady-state throughput rather than startup.
const numCPUTasks = 1000

// numCPUIter is the number of iterations in a single compute task. It is
// large enough that each task is a realistically coarse unit of work whose
// runtime dominates the pool's per-task dispatch overhead, exposing how well
// the pool parallelizes real work rather than just its dispatch cost.
const numCPUIter = 10000

// numIOTasks is the number of IO-bound tasks submitted per iteration.
// With numWorkers workers the tasks run in numIOTasks/numWorkers waves.
const numIOTasks = 64

// ioLatency is the blocking wait each IO-bound task simulates with a sleep,
// standing in for a network or disk round-trip.
const ioLatency = time.Millisecond

// ioSink absorbs the argument of IO-bound tasks.
var ioSink int

func BenchmarkPoolCPU_So(b *testing.B) {
	// Measures pool throughput on CPU-bound work: numCPUTasks
	// compute tasks run across numWorkers workers.
	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	args := make([]workArg, numCPUTasks)
	for i := range args {
		args[i].n = numCPUIter
	}

	for b.Loop() {
		for i := range numCPUTasks {
			p.Go(cpuWork, &args[i])
		}
		p.Wait()
	}
}

func BenchmarkPoolIO_So(b *testing.B) {
	// Measures pool throughput on IO-bound work: numIOTasks tasks
	// that each block for ioLatency run across numWorkers workers.
	opts := conc.PoolOptions{NumThreads: numWorkers}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	for b.Loop() {
		for range numIOTasks {
			p.Go(ioWork, &ioSink)
		}
		p.Wait()
	}
}

// workArg is the per-task state for the cpuWork benchmark:
// the task reads n and writes its result into sum.
type workArg struct {
	n   int
	sum int
}

// cpuWork runs a small compute task that resists constant folding (the
// running modulo makes each step depend on the last), then stores the result.
func cpuWork(arg any) {
	w := arg.(*workArg)
	acc := 0
	for i := range w.n {
		acc = (acc + i*i) % 1000000
	}
	w.sum = acc
}

// ioWork blocks for ioLatency, standing in for a blocking IO call.
func ioWork(arg any) {
	time.Sleep(ioLatency)
}
