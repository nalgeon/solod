package main

import (
	"solod.dev/so/bytes"
	"solod.dev/so/fmt"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

//so:embed bench.h
var bench_h string

var arena *mem.Arena

//so:extern nodecay
var (
	sink    []byte
	sinkInt int
)

func Clone(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte{'a'}, 1024)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.Clone(a, src)
		mem.FreeSlice(a, data)
		if arena != nil {
			arena.Reset()
		}
	}
}

func Compare(b *testing.B) {
	src1 := bytes.Repeat(nil, []byte("01234567890abcdef"), 64)
	src2 := bytes.Repeat(nil, []byte("01234567890abcdef"), 64)
	defer mem.FreeSlice(nil, src1)
	defer mem.FreeSlice(nil, src2)
	for b.Loop() {
		sinkInt = bytes.Compare(src1, src2)
		if arena != nil {
			arena.Reset()
		}
	}
}

func Index(b *testing.B) {
	buf := bytes.NewBuffer(nil, nil)
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("xyz")
	src := buf.Bytes()
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sinkInt = bytes.Index(src, []byte("xyz"))
		if arena != nil {
			arena.Reset()
		}
	}
}

func IndexByte(b *testing.B) {
	buf := bytes.NewBuffer(nil, nil)
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteByte('x')
	src := buf.Bytes()
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sinkInt = bytes.IndexByte(src, 'x')
		if arena != nil {
			arena.Reset()
		}
	}
}

func Repeat(b *testing.B) {
	a := b.Allocator()
	src := []byte("0123456789abcdef")
	for b.Loop() {
		data := bytes.Repeat(a, src, 64)
		mem.FreeSlice(a, data)
		if arena != nil {
			arena.Reset()
		}
	}
}

func ReplaceAll(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("0123456789abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.Replace(a, src, []byte("a"), []byte("AB"), -1)
		mem.FreeSlice(a, data)
		if arena != nil {
			arena.Reset()
		}
	}
}

func Split(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		fields := bytes.Split(a, src, []byte("abc"))
		mem.FreeSlice(a, fields)
		if arena != nil {
			arena.Reset()
		}
	}
}

func ToUpper(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.ToUpper(a, src)
		mem.FreeSlice(a, data)
		if arena != nil {
			arena.Reset()
		}
	}
}

func Trim(b *testing.B) {
	buf := bytes.NewBuffer(nil, nil)
	buf.WriteString("jklmnopqrstuvwxyz")
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("jklmnopqrstuvwxyz")
	src := buf.Bytes()
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sink = bytes.Trim(src, "jklmnopqrstuvwxyz")
	}
}

func TrimSuffix(b *testing.B) {
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	suffix := []byte("01234567890abcdef")
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sink = bytes.TrimSuffix(src, suffix)
	}
}

func main() {
	benchs := []testing.Benchmark{
		{Name: "Clone", F: Clone},
		{Name: "Compare", F: Compare},
		{Name: "Index", F: Index},
		{Name: "IndexByte", F: IndexByte},
		{Name: "Repeat", F: Repeat},
		{Name: "ReplaceAll", F: ReplaceAll},
		{Name: "Split", F: Split},
		{Name: "ToUpper", F: ToUpper},
		{Name: "Trim", F: Trim},
		{Name: "TrimSuffix", F: TrimSuffix},
	}

	fmt.Println("Malloc-based allocator:")
	testing.RunBenchmarks(mem.System, benchs)

	fmt.Println("Arena allocator:")
	var buf [4096]byte
	a := mem.NewArena(buf[:])
	arena = &a
	testing.RunBenchmarks(arena, benchs)
}
