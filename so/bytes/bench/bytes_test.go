package bytes_bench

import (
	"bytes"
	"testing"
)

func BenchmarkClone_Go(b *testing.B) {
	b.ReportAllocs()
	var src = bytes.Repeat([]byte{'a'}, 1024)
	for b.Loop() {
		sink = bytes.Clone(src)
	}
}

func BenchmarkCompare_Go(b *testing.B) {
	b.ReportAllocs()
	src1 := bytes.Repeat([]byte("01234567890abcdef"), 64)
	src2 := bytes.Repeat([]byte("01234567890abcdef"), 64)
	for b.Loop() {
		sinkInt = bytes.Compare(src1, src2)
	}
}

func BenchmarkIndex_Go(b *testing.B) {
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

func BenchmarkIndexByte_Go(b *testing.B) {
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

func BenchmarkRepeat_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = bytes.Repeat([]byte("0123456789abcdef"), 64)
	}
}

func BenchmarkReplaceAll_Go(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("0123456789abcdef"), 16)
	for b.Loop() {
		sink = bytes.Replace(src, []byte("a"), []byte("AB"), -1)
	}
}

func BenchmarkSplit_Go(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	for b.Loop() {
		fields := bytes.Split(src, []byte("abc"))
		sink = fields[0]
	}
}

func BenchmarkToUpper_Go(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	for b.Loop() {
		sink = bytes.ToUpper(src)
	}
}

func BenchmarkTrim_Go(b *testing.B) {
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

func BenchmarkTrimSuffix_Go(b *testing.B) {
	b.ReportAllocs()
	src := bytes.Repeat([]byte("01234567890abcdef"), 16)
	suffix := []byte("01234567890abcdef")
	for b.Loop() {
		sink = bytes.TrimSuffix(src, suffix)
	}
}
