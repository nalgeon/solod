package conc_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
)

// randArg carries the values one thread draws
// from the top-level functions of math/rand.
type randArg struct {
	vals [4]uint64
}

// drawRand fills a.vals from the top-level functions of math/rand.
func drawRand(arg any) any {
	a := arg.(*randArg)
	for i := range a.vals {
		a.vals[i] = rand.Uint64()
	}
	return arg
}

func TestRandThread(t *testing.T) {
	// Each thread holds its own Rand, and the first top-level call seeds it.
	// A thread with no seed would read a nil Source and crash.
	const n = 4
	args := make([]randArg, n)
	threads := make([]conc.Thread, n)
	for i := range args {
		threads[i] = conc.Go(drawRand, &args[i])
	}
	for i := range threads {
		threads[i].Wait()
	}

	// The seeded source gives a different value every call, so two values of
	// one thread are almost never the same.
	for i := range args {
		vals := args[i].vals
		for j := 1; j < len(vals); j++ {
			if vals[j] == vals[j-1] {
				t.Errorf("thread %d drew %d twice in a row", i, vals[j])
			}
		}
	}

	// Each thread takes a seed of its own, so two threads almost never draw
	// the same first value.
	for i := 1; i < n; i++ {
		if args[i].vals[0] == args[i-1].vals[0] {
			t.Errorf("threads %d and %d drew the same first value", i-1, i)
		}
	}
}
