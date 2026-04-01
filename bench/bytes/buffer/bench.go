package main

import (
	"solod.dev/so/bytes"
	"solod.dev/so/fmt"
	"solod.dev/so/mem"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

//so:embed bench.h
var bench_h string

//so:extern nodecay
var sink string

var arena *mem.Arena

func ReadString(b *testing.B) {
	const n = 32 << 10
	b.SetBytes(int64(n))

	a := b.Allocator()
	data := make([]byte, n)
	data[n-1] = 'x'

	for b.Loop() {
		buf := bytes.NewBuffer(a, data)
		s, err := buf.ReadString('x')
		sink = s
		if err != nil {
			panic(err)
		}
		mem.FreeString(a, s)
		if arena != nil {
			arena.Reset()
		}
	}
}

func WriteByte(b *testing.B) {
	const n = 4 << 10
	b.SetBytes(n)

	a := b.Allocator()
	data := make([]byte, n)
	buf := bytes.NewBuffer(a, data)
	defer buf.Free()

	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteByte('x')
		}
	}
}

func WriteRune(b *testing.B) {
	const n = 4 << 10
	const r = '☺'
	b.SetBytes(int64(n * utf8.RuneLen(r)))

	a := b.Allocator()
	data := make([]byte, n*utf8.UTFMax)
	buf := bytes.NewBuffer(a, data)
	defer buf.Free()

	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteRune(r)
		}
	}
}

func WriteBlock(b *testing.B) {
	block := make([]byte, 1024)
	const n = 2 << 16
	a := b.Allocator()
	for b.Loop() {
		buf := bytes.NewBuffer(a, nil)
		for buf.Len() < n {
			buf.Write(block)
		}
		buf.Free()
		if arena != nil {
			arena.Reset()
		}
	}
}

func main() {
	benchs := []testing.Benchmark{
		{Name: "ReadString", F: ReadString},
		{Name: "WriteByte", F: WriteByte},
		{Name: "WriteRune", F: WriteRune},
		{Name: "WriteBlock", F: WriteBlock},
	}

	fmt.Println("Malloc-based allocator:")
	testing.RunBenchmarks(mem.System, benchs)

	fmt.Println("Arena allocator:")
	buf := slices.Make[byte](nil, 2<<20)
	defer slices.Free(nil, buf)
	a := mem.NewArena(buf)
	arena = &a
	testing.RunBenchmarks(arena, benchs)
}
