package main

import (
	"bytes"
	"testing"
)

func Benchmark_Clone(b *testing.B) {
	b.ReportAllocs()
	var src = bytes.Repeat([]byte{'a'}, 1024)
	for b.Loop() {
		sink = bytes.Clone(src)
	}
}

func Benchmark_Compare(b *testing.B) {
	b.ReportAllocs()
	src1 := bytes.Repeat([]byte("01234567890abcdef"), 64)
	src2 := bytes.Repeat([]byte("01234567890abcdef"), 64)
	for b.Loop() {
		sinkInt = bytes.Compare(src1, src2)
	}
}

func Benchmark_Index(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("xyz")
	src := buf.Bytes()
	for b.Loop() {
		sinkInt = bytes.Index(src, []byte("xyz"))
	}
}

func Benchmark_IndexByte(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("x")
	src := buf.Bytes()
	for b.Loop() {
		sinkInt = bytes.IndexByte(src, 'x')
	}
}

func Benchmark_Repeat(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = bytes.Repeat([]byte("0123456789abcdef"), 64)
	}
}

func Benchmark_ReplaceAll(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("0123456789abcdef"), 16)
	for b.Loop() {
		sink = bytes.Replace(src, []byte("a"), []byte("AB"), -1)
	}
}

func Benchmark_Split(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	for b.Loop() {
		fields := bytes.Split(src, []byte("abc"))
		sink = fields[0]
	}
}

func Benchmark_ToUpper(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	for b.Loop() {
		sink = bytes.ToUpper(src)
	}
}

func Benchmark_Trim(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	buf.WriteString("jklmnopqrstuvwxyz")
	for range 64 {
		buf.WriteString("01234567890abcdef")
	}
	buf.WriteString("jklmnopqrstuvwxyz")
	src := buf.Bytes()
	for b.Loop() {
		sink = bytes.Trim(src, "jklmnopqrstuvwxyz")
	}
}

func Benchmark_TrimSuffix(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	suffix := []byte("01234567890abcdef")
	for b.Loop() {
		sink = bytes.TrimSuffix(src, suffix)
	}
}
