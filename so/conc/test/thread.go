package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/testing"
)

func increment(arg any) any {
	n := arg.(*int)
	*n = *n + 1
	return arg
}

func TestThread_Wait(t *testing.T) {
	// Start a thread per element, wait for them all, and check every result.
	const n = 16
	nums := make([]int, n)
	threads := make([]conc.Thread, n)
	for i := range nums {
		nums[i] = i
		threads[i] = conc.Go(increment, &nums[i])
	}

	ok := true
	for i := range threads {
		res := threads[i].Wait()
		if *(res.(*int)) != i+1 {
			ok = false
		}
	}
	for i := range nums {
		if nums[i] != i+1 {
			ok = false
		}
	}
	if !ok {
		t.Error("wrong increment result")
	}
}

// squareArg carries the value a detached thread squares, plus the latch it
// reports completion through.
type squareArg struct {
	sig latch
	out int
}

// squareLatch squares a.out in place, then marks the latch done.
func squareLatch(arg any) any {
	a := arg.(*squareArg)
	a.out = a.out * a.out
	a.sig.Done()
	return nil
}

func TestThread_Detach(t *testing.T) {
	// Run a task on a detached thread and wait for it through a latch.
	var a squareArg
	a.sig.Init()
	defer a.sig.Free()
	a.out = 9

	th := conc.Go(squareLatch, &a)
	th.Detach()
	a.sig.Wait()

	if a.out != 81 {
		t.Error("wrong detached result")
	}
}

// depthArg carries the recursion depth and the resulting sum.
type depthArg struct {
	depth int
	sum   int
}

// sumDown adds n..1 by recursion, which needs one stack frame per level.
func sumDown(n int) int {
	if n <= 0 {
		return 0
	}
	return n + sumDown(n-1)
}

// runSumDown recurses on the thread's own stack.
func runSumDown(arg any) any {
	a := arg.(*depthArg)
	a.sum = sumDown(a.depth)
	return arg
}

func TestThread_StackSize(t *testing.T) {
	// Run a recursion on a thread with an explicit stack size. The recursion needs
	// much less stack than the request, so the thread must finish normally.
	a := depthArg{depth: 1000, sum: 0}
	opts := conc.ThreadOptions{StackSize: 1 << 20}
	th := conc.GoWith(runSumDown, &a, opts)
	th.Wait()

	// Sum of 1..1000.
	if a.sum != 500500 {
		t.Errorf("sumDown(1000) = %d, want 500500", a.sum)
	}
}
