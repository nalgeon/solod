package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/errors"
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

// runSquares submits n square tasks to a pool with the given options, waits for
// them, and checks every result. Only the options and the task count change
// between the pool tests below.
func runSquares(t *testing.T, opts conc.PoolOptions, n int) {
	tasks := make([]Task, n)
	p := conc.NewPool(t.Allocator(), opts)
	defer p.Free()
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatalf("task %d = %d, want %d", i, tasks[i].out, i*i)
			return
		}
	}
}

func TestPool_ParallelMap(t *testing.T) {
	// Square 0..99 in parallel and checks every result.
	runSquares(t, conc.PoolOptions{NumThreads: 8}, 100)
}

func TestPool_DefaultOptions(t *testing.T) {
	// A zero PoolOptions must produce a working pool with a CPU-count default.
	runSquares(t, conc.PoolOptions{}, 100)
}

func TestPool_BackPressure(t *testing.T) {
	// Submit far more tasks than workers, exercising the queue-full wait.
	runSquares(t, conc.PoolOptions{NumThreads: 2}, 1000)
}

func TestPool_QueueLarge(t *testing.T) {
	// Use a queue far larger than the worker limit, so most submissions
	// enqueue without blocking. All results must still be correct.
	runSquares(t, conc.PoolOptions{NumThreads: 2, QueueSize: 128}, 200)
}

func TestPool_QueueOne(t *testing.T) {
	// Use the smallest possible queue, so each submission past the first must
	// wait for a worker to drain a slot. This stresses the queue-full
	// back-pressure path with an explicit queue size.
	runSquares(t, conc.PoolOptions{NumThreads: 4, QueueSize: 1}, 50)
}

func TestPool_StackSize(t *testing.T) {
	// An explicit worker stack size must produce a working pool.
	runSquares(t, conc.PoolOptions{NumThreads: 4, StackSize: 1 << 20}, 50)
}

func TestPool_WaitReuse(t *testing.T) {
	// The pool stays usable after Wait, so a second batch must run just as
	// well as the first.
	const n = 50
	tasks := make([]Task, n)
	p := conc.NewPool(t.Allocator(), conc.PoolOptions{NumThreads: 4})
	defer p.Free()

	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Wait()
	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatalf("first batch: task %d = %d, want %d", i, tasks[i].out, i*i)
			return
		}
	}

	// The second batch squares the results of the first.
	for i := range tasks {
		tasks[i].in = tasks[i].out
		p.Go(square, &tasks[i])
	}
	p.Wait()
	for i := range tasks {
		want := i * i * i * i
		if tasks[i].out != want {
			t.Fatalf("second batch: task %d = %d, want %d", i, tasks[i].out, want)
			return
		}
	}
}

func TestPool_WaitIdle(t *testing.T) {
	// Wait on a pool with no submitted task must return at once. The test
	// hangs if it does not, so there is nothing else to assert.
	p := conc.NewPool(t.Allocator(), conc.PoolOptions{NumThreads: 2})
	defer p.Free()
	p.Wait()
	p.Wait()
}

func TestPool_FreeDrains(t *testing.T) {
	// Free drains the queue, so every submitted task must be complete once
	// Free returns, even with no Wait before it.
	const n = 200
	tasks := make([]Task, n)
	p := conc.NewPool(t.Allocator(), conc.PoolOptions{NumThreads: 2, QueueSize: 64})
	for i := range tasks {
		tasks[i].in = i
		p.Go(square, &tasks[i])
	}
	p.Free()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatalf("task %d = %d, want %d", i, tasks[i].out, i*i)
			return
		}
	}
}

// submitArg carries the pool and the tasks one submitter thread pushes into it.
type submitArg struct {
	pool  *conc.Pool
	tasks []Task
}

// submitSquares submits a square task for every task of the submitter.
func submitSquares(arg any) any {
	a := arg.(*submitArg)
	for i := range a.tasks {
		a.pool.Go(square, &a.tasks[i])
	}
	return nil
}

func TestPool_ConcurrentSubmit(t *testing.T) {
	// Several threads submit to one pool at the same time, which exercises the
	// thread safety of Go and the queue-full wait under contention. The queue
	// is much smaller than the batch, so the submitters block often.
	const submitters = 4
	const perSubmitter = 250
	const n = submitters * perSubmitter

	tasks := make([]Task, n)
	for i := range tasks {
		tasks[i].in = i
	}
	opts := conc.PoolOptions{NumThreads: 4, QueueSize: 8}
	p := conc.NewPool(t.Allocator(), opts)
	defer p.Free()

	args := make([]submitArg, submitters)
	threads := make([]conc.Thread, submitters)
	for i := range args {
		base := i * perSubmitter
		args[i] = submitArg{pool: p, tasks: tasks[base : base+perSubmitter]}
		threads[i] = conc.Go(submitSquares, &args[i])
	}
	// Join the submitters first, so every task is in flight before Wait.
	for i := range threads {
		threads[i].Wait()
	}
	p.Wait()

	for i := range tasks {
		if tasks[i].out != i*i {
			t.Fatalf("task %d = %d, want %d", i, tasks[i].out, i*i)
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
	// Check that a task can report an error through its argument struct.
	const n = 10
	tasks := make([]Task, n)
	opts := conc.PoolOptions{NumThreads: 4}
	p := conc.NewPool(t.Allocator(), opts)
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
