package rand_bench

import (
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
)

//so:volatile
var sink uint64

func BenchmarkSourceUint64_So(b *testing.B) {
	s := rand.NewPCG(1, 2)
	var t uint64
	for b.Loop() {
		t += s.Uint64()
	}
	sink = uint64(t)
}

func BenchmarkGlobalUint64_So(b *testing.B) {
	var t uint64
	for b.Loop() {
		t += rand.Uint64()
	}
	sink = t
}

func BenchmarkUint64_So(b *testing.B) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	var t uint64
	for b.Loop() {
		t += r.Uint64()
	}
	sink = t
}

func BenchmarkInt64N1e9_So(b *testing.B) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	var t int64
	for b.Loop() {
		t += r.Int64N(1e9)
	}
	sink = uint64(t)
}

func BenchmarkInt64N1e18_So(b *testing.B) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	var t int64
	for b.Loop() {
		t += r.Int64N(1e18)
	}
	sink = uint64(t)
}

func BenchmarkInt64N4e18_So(b *testing.B) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	var t int64
	for b.Loop() {
		t += r.Int64N(4e18)
	}
	sink = uint64(t)
}

func BenchmarkFloat64_So(b *testing.B) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	var t float64
	for b.Loop() {
		t += r.Float64()
	}
	sink = uint64(t)
}
