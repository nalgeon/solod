package rand_test

import (
	"solod.dev/so/math"
	"solod.dev/so/math/rand"
	"solod.dev/so/runtime"
	"solod.dev/so/testing"
)

// normDraws is the number of values that a distribution test takes.
const normDraws = 100000

// tail is the start of the base strip of the ziggurat. A value past the tail
// takes the slow path of NormFloat64.
const tail = 3.442619855899

func TestNormFloat64(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("NormFloat64 needs a hosted environment")
		return
	}
	// The results follow the standard normal distribution, so the mean is near
	// 0 and the standard deviation is near 1. The bounds are wide, because the
	// test checks the shape of the distribution, not the stream.
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	sum, sumSq := 0.0, 0.0
	for i := 0; i < normDraws; i++ {
		x := r.NormFloat64()
		if math.IsNaN(x) || math.IsInf(x, 0) {
			t.Errorf("NormFloat64() case %d is not finite", i)
			return
		}
		sum += x
		sumSq += x * x
	}
	mean := sum / normDraws
	if mean < -0.02 || mean > 0.02 {
		t.Errorf("NormFloat64() mean = %g, want near 0", mean)
	}
	stddev := math.Sqrt(sumSq/normDraws - mean*mean)
	if stddev < 0.98 || stddev > 1.02 {
		t.Errorf("NormFloat64() stddev = %g, want near 1", stddev)
	}
}

func TestNormFloat64Tail(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("NormFloat64 needs a hosted environment")
		return
	}
	// About 0.03% of the results are past the tail on each side. The slow path
	// of NormFloat64 produces them, so the test proves that the slow path runs
	// and gives values of both signs.
	p := rand.NewPCG(3, 4)
	r := rand.New(&p)
	lo, hi := 0, 0
	for i := 0; i < normDraws; i++ {
		x := r.NormFloat64()
		if x > tail {
			hi++
		} else if x < -tail {
			lo++
		}
	}
	if hi < 5 {
		t.Errorf("NormFloat64() gave %d values above the tail, want at least 5", hi)
	}
	if lo < 5 {
		t.Errorf("NormFloat64() gave %d values below the tail, want at least 5", lo)
	}
}

func TestNormFloat64Stream(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("NormFloat64 needs a hosted environment")
		return
	}
	// Two generators with the same seed give the same stream.
	p1 := rand.NewPCG(5, 6)
	r1 := rand.New(&p1)
	p2 := rand.NewPCG(5, 6)
	r2 := rand.New(&p2)
	for i := 0; i < 100; i++ {
		if got, want := r1.NormFloat64(), r2.NormFloat64(); got != want {
			t.Errorf("NormFloat64() case %d = %g, want %g", i, got, want)
			return
		}
	}
}

func TestNormFloat64Global(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("NormFloat64 needs a hosted environment")
		return
	}
	// The global generator takes a random seed, so the test checks the range of
	// the results.
	for i := 0; i < 1000; i++ {
		x := rand.NormFloat64()
		if math.IsNaN(x) || math.IsInf(x, 0) {
			t.Errorf("NormFloat64() case %d is not finite", i)
			return
		}
	}
}
