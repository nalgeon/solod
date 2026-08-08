package io_bench

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/testing"
)

func BenchmarkCopyNSmall_So(b *testing.B) {
	a := b.Allocator()

	bs := make([]byte, 512+1)
	rd := bytes.NewReader(bs)
	buf := bytes.NewBuffer(a, nil)
	defer buf.Free()

	for b.Loop() {
		io.CopyN(&buf, &rd, 512)
		rd.Reset(bs)
	}
}

func BenchmarkCopyNLarge_So(b *testing.B) {
	a := b.Allocator()

	bs := make([]byte, 32*1024+1)
	rd := bytes.NewReader(bs)
	buf := bytes.NewBuffer(a, nil)
	defer buf.Free()

	for b.Loop() {
		io.CopyN(&buf, &rd, 32*1024)
		rd.Reset(bs)
	}
}
