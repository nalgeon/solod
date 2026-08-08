package hex_bench

import (
	"solod.dev/so/bytes"
	"solod.dev/so/encoding/hex"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

//so:volatile
var sink []byte

func BenchmarkEncode_256_So(b *testing.B) {
	encode(b, 256)
}

func BenchmarkEncode_1024_So(b *testing.B) {
	encode(b, 1024)
}

func BenchmarkEncode_4096_So(b *testing.B) {
	encode(b, 4096)
}

func BenchmarkEncode_16384_So(b *testing.B) {
	encode(b, 16384)
}

func BenchmarkDecode_256_So(b *testing.B) {
	decode(b, 256)
}

func BenchmarkDecode_1024_So(b *testing.B) {
	decode(b, 1024)
}

func BenchmarkDecode_4096_So(b *testing.B) {
	decode(b, 4096)
}

func BenchmarkDecode_16384_So(b *testing.B) {
	decode(b, 16384)
}

func encode(b *testing.B, size int) {
	b.SetBytes(int64(size))
	alloc := b.Allocator()
	data := []byte{2, 3, 5, 7, 9, 11, 13, 17}
	src := bytes.Repeat(alloc, data, size/8)
	sink = slices.Make[byte](alloc, 2*size)
	for b.Loop() {
		hex.Encode(sink, src)
	}
	slices.Free(alloc, src)
}

func decode(b *testing.B, size int) {
	b.SetBytes(int64(size))
	alloc := b.Allocator()
	data := []byte{'2', 'b', '7', '4', '4', 'f', 'a', 'a'}
	src := bytes.Repeat(alloc, data, size/8)
	sink = slices.Make[byte](alloc, size/2)
	for b.Loop() {
		hex.Decode(sink, src)
	}
	slices.Free(alloc, src)
}
