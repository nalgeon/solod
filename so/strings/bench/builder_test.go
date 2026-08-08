package strings_bench

import (
	"strings"
	"testing"
)

func BenchmarkWriteB_AutoGrow_Go(b *testing.B) {
	someBytes := []byte(someStr)
	b.ReportAllocs()
	for b.Loop() {
		var buf strings.Builder
		for range numWrite {
			buf.Write(someBytes)
		}
		sink = buf.String()
	}
}

func BenchmarkWriteB_PreGrow_Go(b *testing.B) {
	someBytes := []byte(someStr)
	b.ReportAllocs()
	for b.Loop() {
		var buf strings.Builder
		buf.Grow(len(someBytes) * numWrite)
		for range numWrite {
			buf.Write(someBytes)
		}
		sink = buf.String()
	}
}

func BenchmarkWriteS_AutoGrow_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf strings.Builder
		for range numWrite {
			buf.WriteString(someStr)
		}
		sink = buf.String()
	}
}

func BenchmarkWriteS_PreGrow_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf strings.Builder
		buf.Grow(len(someStr) * numWrite)
		for range numWrite {
			buf.WriteString(someStr)
		}
		sink = buf.String()
	}
}
