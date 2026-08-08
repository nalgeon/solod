package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/errors"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

// Task carries one task's input, output and error through a *Task.
type Task struct {
	in  int
	out int
	err error
}

func square(arg any) {
	task := arg.(*Task)
	task.out = task.in * task.in
}

func TestPool_ParallelMap(t *testing.T) {
	// Squares 0..99 in parallel and checks every result.
	const n = 100
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatal("wrong square result")
			return
		}
	}
}

func TestPool_DefaultOptions(t *testing.T) {
	// A zero PoolOptions must produce a working pool with a CPU-count default.
	const n = 100
	tasks := make([]Task, n)
	p := conc.NewPool(mem.System, conc.PoolOptions{})
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatal("wrong square result")
			return
		}
	}
}

func TestPool_BackPressure(t *testing.T) {
	// Submits far more tasks than workers, exercising the queue-full wait.
	const n = 1000
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 2}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	sum := 0
	for i := range tasks {
		sum += tasks[i].out
	}
	// Sum of i*i for i in 0..999.
	if sum != 332833500 {
		t.Error("wrong sum")
	}
}

func TestPool_QueueLarge(t *testing.T) {
	// Uses a queue far larger than the worker limit, so most submissions
	// enqueue without blocking. All results must still be correct.
	const n = 200
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 2, QueueSize: 128}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatal("wrong square result")
			return
		}
	}
}

func TestPool_QueueOne(t *testing.T) {
	// Uses the smallest possible queue, so each submission past the first must
	// wait for a worker to drain a slot. This stresses the queue-full
	// back-pressure path with an explicit queue size.
	const n = 50
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 4, QueueSize: 1}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatal("wrong square result")
			return
		}
	}
}

var errOddInput = errors.New("odd input")

func checkEven(arg any) {
	task := arg.(*Task)
	if task.in%2 != 0 {
		task.err = errOddInput
		return
	}
	task.out = task.in
}

func TestPool_Error(t *testing.T) {
	// Checks that a task can report an error through its argument struct.
	const n = 10
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 4}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(checkEven, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if i%2 != 0 && tasks[i].err != errOddInput {
			t.Fatal("expected error for odd input")
			return
		}
		if i%2 == 0 && tasks[i].err != nil {
			t.Fatal("unexpected error for even input")
			return
		}
	}
}
