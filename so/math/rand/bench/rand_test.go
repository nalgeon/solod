package rand_bench

import (
	"math/rand/v2"
	"testing"
)

func testRand() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}

func BenchmarkSourceUint64_Go(b *testing.B) {
	s := rand.NewPCG(1, 2)
	var t uint64
	for b.Loop() {
		t += s.Uint64()
	}
	sink = uint64(t)
}

func BenchmarkGlobalUint64_Go(b *testing.B) {
	var t uint64
	for b.Loop() {
		t += rand.Uint64()
	}
	sink = t
}

func BenchmarkUint64_Go(b *testing.B) {
	r := testRand()
	var t uint64
	for b.Loop() {
		t += r.Uint64()
	}
	sink = t
}

func BenchmarkInt64N1e9_Go(b *testing.B) {
	r := testRand()
	var t int64
	for b.Loop() {
		t += r.Int64N(1e9)
	}
	sink = uint64(t)
}

func BenchmarkInt64N1e18_Go(b *testing.B) {
	r := testRand()
	var t int64
	for b.Loop() {
		t += r.Int64N(1e18)
	}
	sink = uint64(t)
}

func BenchmarkInt64N4e18_Go(b *testing.B) {
	r := testRand()
	var t int64
	for b.Loop() {
		t += r.Int64N(4e18)
	}
	sink = uint64(t)
}

func BenchmarkFloat64_Go(b *testing.B) {
	r := testRand()
	var t float64
	for b.Loop() {
		t += r.Float64()
	}
	sink = uint64(t)
}
