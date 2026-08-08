package sync_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
)

// onceVal is set by onceInit; onceRuns counts how many times onceInit ran.
var onceVal int
var onceRuns int

// onceInit is the one-time initialization run through sync.Once.
func onceInit() {
	onceVal = 42
	onceRuns++
}

// onceTask carries the shared Once and a slot for the value
// each worker observes right after its Do returns.
type onceTask struct {
	once *sync.Once
	seen *int
}

func callOnce(arg any) {
	task := arg.(*onceTask)
	task.once.Do(onceInit)
	*task.seen = onceVal
}

// Has many workers race on a single Once and checks that the
// initializer ran exactly once and that every Do returned only after
// it completed (each worker observes the initialized value).
func TestOnce(t *testing.T) {
	const n = 1000

	var once sync.Once
	once.Init()
	defer once.Free()

	onceVal = 0
	onceRuns = 0

	tasks := make([]onceTask, n)
	seen := make([]int, n)
	opts := conc.PoolOptions{NumThreads: 8}
	p := conc.NewPool(mem.System, opts)
	for i := range tasks {
		tasks[i].once = &once
		tasks[i].seen = &seen[i]
		p.Go(callOnce, &tasks[i])
	}
	p.Free()

	if onceRuns != 1 {
		t.Fatal("once ran the initializer more than once")
		return
	}
	for i := range seen {
		if seen[i] != 42 {
			t.Fatal("Do returned before the initializer completed")
			return
		}
	}
}
