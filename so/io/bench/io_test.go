package io_bench

import (
	"bytes"
	"io"
	"testing"
)

// A version of bytes.Buffer without ReadFrom and WriteTo
type Buffer struct {
	bytes.Buffer
	io.ReaderFrom // conflicts with and hides bytes.Buffer's ReaderFrom.
	io.WriterTo   // conflicts with and hides bytes.Buffer's WriterTo.
}

func BenchmarkCopyNSmall_Go(b *testing.B) {
	b.ReportAllocs()

	bs := make([]byte, 512+1)
	rd := bytes.NewReader(bs)
	buf := new(bytes.Buffer)

	for b.Loop() {
		io.CopyN(buf, rd, 512)
		rd.Reset(bs)
	}
}

func BenchmarkCopyNLarge_Go(b *testing.B) {
	b.ReportAllocs()

	bs := make([]byte, 32*1024+1)
	rd := bytes.NewReader(bs)
	buf := new(Buffer)

	for b.Loop() {
		io.CopyN(buf, rd, 32*1024)
		rd.Reset(bs)
	}
}
