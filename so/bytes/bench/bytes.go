package bytes_bench

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

//so:volatile
var sink []byte

//so:volatile
var sinkInt int

func BenchmarkClone_So(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte{'a'}, 1024)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.Clone(a, src)
		mem.FreeSlice(a, data)
	}
}

func BenchmarkCompare_So(b *testing.B) {
	src1 := bytes.Repeat(nil, []byte("01234567890abcdef"), 64)
	src2 := bytes.Repeat(nil, []byte("01234567890abcdef"), 64)
	defer mem.FreeSlice(nil, src1)
	defer mem.FreeSlice(nil, src2)
	for b.Loop() {
		sinkInt = bytes.Compare(src1, src2)
	}
}

func BenchmarkIndex_So(b *testing.B) {
	buf := bytes.NewBuffer(nil, nil)
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("xyz")
	src := buf.Bytes()
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sinkInt = bytes.Index(src, []byte("xyz"))
	}
}

func BenchmarkIndexByte_So(b *testing.B) {
	buf := bytes.NewBuffer(nil, nil)
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteByte('x')
	src := buf.Bytes()
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sinkInt = bytes.IndexByte(src, 'x')
	}
}

func BenchmarkRepeat_So(b *testing.B) {
	a := b.Allocator()
	src := []byte("0123456789abcdef")
	for b.Loop() {
		data := bytes.Repeat(a, src, 64)
		mem.FreeSlice(a, data)
	}
}

func BenchmarkReplaceAll_So(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("0123456789abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.Replace(a, src, []byte("a"), []byte("AB"), -1)
		mem.FreeSlice(a, data)
	}
}

func BenchmarkSplit_So(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		fields := bytes.Split(a, src, []byte("abc"))
		mem.FreeSlice(a, fields)
	}
}

func BenchmarkToUpper_So(b *testing.B) {
	a := b.Allocator()
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		data := bytes.ToUpper(a, src)
		mem.FreeSlice(a, data)
	}
}

func BenchmarkTrim_So(b *testing.B) {
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

func BenchmarkTrimSuffix_So(b *testing.B) {
	src := bytes.Repeat(nil, []byte("01234567890abcdef"), 16)
	suffix := []byte("01234567890abcdef")
	defer mem.FreeSlice(nil, src)
	for b.Loop() {
		sink = bytes.TrimSuffix(src, suffix)
	}
}
