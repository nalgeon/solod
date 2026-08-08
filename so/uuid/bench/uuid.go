package uuid_bench

import (
	"solod.dev/so/testing"
	"solod.dev/so/uuid"
)

//so:volatile
var sink uuid.UUID

//so:volatile
var sinkStr string

func BenchmarkNewV4_So(b *testing.B) {
	for b.Loop() {
		sink = uuid.NewV4()
	}
}

func BenchmarkNewV7_So(b *testing.B) {
	for b.Loop() {
		sink = uuid.NewV7()
	}
}

func BenchmarkString_So(b *testing.B) {
	u := uuid.MustParse("f81d4fae-7dec-11d0-a765-00a0c91e6bf6")
	buf := make([]byte, uuid.UUIDLen)
	for b.Loop() {
		sinkStr = u.String(buf)
	}
}

func BenchmarkParseSuccess_So(b *testing.B) {
	for b.Loop() {
		sink = uuid.MustParse("f81d4fae-7dec-11d0-a765-00a0c91e6bf6")
	}
}

func BenchmarkParseError_So(b *testing.B) {
	for b.Loop() {
		sink, _ = uuid.Parse("00000000-0000-0000-0000-00000000000X")
	}
}
