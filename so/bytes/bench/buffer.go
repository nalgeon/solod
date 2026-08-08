package bytes_bench

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

//so:volatile
var sinkStr string

func BenchmarkReadString_So(b *testing.B) {
	const n = 32 << 10
	b.SetBytes(int64(n))

	a := b.Allocator()
	data := make([]byte, n)
	data[n-1] = 'x'

	for b.Loop() {
		buf := bytes.NewBuffer(a, data)
		s, err := buf.ReadString('x')
		sinkStr = s
		if err != nil {
			panic(err)
		}
		mem.FreeString(a, s)
	}
}

func BenchmarkWriteByte_So(b *testing.B) {
	const n = 4 << 10
	b.SetBytes(n)

	a := b.Allocator()
	data := mem.AllocSlice[byte](a, n, n)
	buf := bytes.NewBuffer(a, data)
	defer buf.Free()

	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteByte('x')
		}
	}
}

func BenchmarkWriteRune_So(b *testing.B) {
	const n = 4 << 10
	const r = '☺'
	b.SetBytes(int64(n * utf8.RuneLen(r)))

	a := b.Allocator()
	data := mem.AllocSlice[byte](a, n*utf8.UTFMax, n*utf8.UTFMax)
	buf := bytes.NewBuffer(a, data)
	defer buf.Free()

	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteRune(r)
		}
	}
}

func BenchmarkWriteBlock_So(b *testing.B) {
	block := make([]byte, 1024)
	const n = 2 << 16
	a := b.Allocator()
	for b.Loop() {
		buf := bytes.NewBuffer(a, nil)
		for buf.Len() < n {
			buf.Write(block)
		}
		buf.Free()
	}
}
