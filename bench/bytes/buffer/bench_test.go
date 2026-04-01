package main

import (
	"bytes"
	"testing"

	"solod.dev/so/unicode/utf8"
)

func BenchmarkReadString(b *testing.B) {
	b.ReportAllocs()
	const n = 32 << 10

	data := make([]byte, n)
	data[n-1] = 'x'
	b.SetBytes(int64(n))
	for b.Loop() {
		buf := bytes.NewBuffer(data)
		_, err := buf.ReadString('x')
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteByte(b *testing.B) {
	const n = 4 << 10
	b.SetBytes(n)
	buf := bytes.NewBuffer(make([]byte, n))
	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteByte('x')
		}
	}
}

func BenchmarkWriteRune(b *testing.B) {
	const n = 4 << 10
	const r = '☺'
	b.SetBytes(int64(n * utf8.RuneLen(r)))
	buf := bytes.NewBuffer(make([]byte, n*utf8.UTFMax))
	for b.Loop() {
		buf.Reset()
		for range n {
			buf.WriteRune(r)
		}
	}
}

func BenchmarkWriteBlock(b *testing.B) {
	block := make([]byte, 1024)
	const n = 2 << 16
	b.ReportAllocs()
	for b.Loop() {
		var bb bytes.Buffer
		for bb.Len() < n {
			bb.Write(block)
		}
	}
}
