package path_bench

import (
	"solod.dev/so/mem"
	"solod.dev/so/path"
	"solod.dev/so/testing"
)

//so:volatile
var sinkBool bool

//so:volatile
var sinkErr error

//so:volatile
var sinkStr string

func BenchmarkJoin_So(b *testing.B) {
	a := b.Allocator()
	parts := []string{"one", "two", "three", "four"}
	for b.Loop() {
		sinkStr = path.Join(a, parts...)
		mem.FreeString(a, sinkStr)
	}
}

func BenchmarkMatchTrue_So(b *testing.B) {
	pattern := "a*b*c*d*e*/f"
	s := "axbxcxdxexxx/f"
	for b.Loop() {
		sinkBool, sinkErr = path.Match(pattern, s)
	}
}

func BenchmarkMatchFalse_So(b *testing.B) {
	pattern := "a*b*c*d*e*/f"
	s := "axbxcxdxexxx/fff"
	for b.Loop() {
		sinkBool, sinkErr = path.Match(pattern, s)
	}
}
