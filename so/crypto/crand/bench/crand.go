package crand_bench

import (
	"solod.dev/so/crypto/crand"
	"solod.dev/so/testing"
)

//so:volatile
var sinkSlice []byte

//so:volatile
var sinkStr string

func BenchmarkRead_4_So(b *testing.B) {
	const size = 4
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}
	}
}

func BenchmarkRead_32_So(b *testing.B) {
	const size = 32
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}
	}
}

func BenchmarkRead_4K_So(b *testing.B) {
	const size = 4 << 10
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}
	}
}

func BenchmarkText_So(b *testing.B) {
	const size = 26
	b.SetBytes(size)
	buf := make([]byte, size)
	for b.Loop() {
		sinkStr = crand.Text(buf)
	}
}
