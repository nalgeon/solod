package path_bench

import (
	"path"
	"testing"
)

func BenchmarkJoin_Go(b *testing.B) {
	b.ReportAllocs()
	parts := []string{"one", "two", "three", "four"}
	for b.Loop() {
		sinkStr = path.Join(parts...)
	}
}

func BenchmarkMatchTrue_Go(b *testing.B) {
	pattern := "a*b*c*d*e*/f"
	s := "axbxcxdxexxx/f"
	b.ReportAllocs()
	for b.Loop() {
		sinkBool, sinkErr = path.Match(pattern, s)
	}
}

func BenchmarkMatchFalse_Go(b *testing.B) {
	pattern := "a*b*c*d*e*/f"
	s := "axbxcxdxexxx/fff"
	b.ReportAllocs()
	for b.Loop() {
		sinkBool, sinkErr = path.Match(pattern, s)
	}
}
