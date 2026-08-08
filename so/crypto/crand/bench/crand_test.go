package crand_bench

import (
	"crypto/rand"
	"testing"
)

func BenchmarkRead_4_Go(b *testing.B) {
	const size = 4
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := rand.Read(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead_32_Go(b *testing.B) {
	const size = 32
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := rand.Read(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead_4K_Go(b *testing.B) {
	const size = 4 << 10
	b.SetBytes(int64(size))
	buf := make([]byte, size)
	for b.Loop() {
		if _, err := rand.Read(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkText_Go(b *testing.B) {
	b.SetBytes(26)
	for b.Loop() {
		sinkStr = rand.Text()
	}
}
