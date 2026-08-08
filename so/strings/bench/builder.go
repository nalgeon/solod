package strings_bench

import (
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

var someStr = "some string sdljlk jsklj3lkjlk djlkjw"

const numWrite = 16

func BenchmarkWriteB_AutoGrow_So(b *testing.B) {
	a := b.Allocator()
	someBytes := []byte(someStr)
	for b.Loop() {
		buf := strings.NewBuilder(a)
		for range numWrite {
			buf.Write(someBytes)
		}
		sink = buf.String()
		buf.Free()
	}
}

func BenchmarkWriteB_PreGrow_So(b *testing.B) {
	a := b.Allocator()
	someBytes := []byte(someStr)
	for b.Loop() {
		buf := strings.NewBuilder(a)
		buf.Grow(len(someBytes) * numWrite)
		for range numWrite {
			buf.Write(someBytes)
		}
		sink = buf.String()
		buf.Free()
	}
}

func BenchmarkWriteS_AutoGrow_So(b *testing.B) {
	a := b.Allocator()
	for b.Loop() {
		buf := strings.NewBuilder(a)
		for range numWrite {
			buf.WriteString(someStr)
		}
		sink = buf.String()
		buf.Free()
	}
}

func BenchmarkWriteS_PreGrow_So(b *testing.B) {
	a := b.Allocator()
	for b.Loop() {
		buf := strings.NewBuilder(a)
		buf.Grow(len(someStr) * numWrite)
		for range numWrite {
			buf.WriteString(someStr)
		}
		sink = buf.String()
		buf.Free()
	}
}
