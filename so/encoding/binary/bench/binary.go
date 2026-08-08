package binary_bench

import (
	"solod.dev/so/encoding/binary"
	"solod.dev/so/testing"
)

type Sink struct {
	buf [8]byte
	res []byte
}

//so:volatile
var sink Sink

//so:volatile
var sinkInt int

func BenchmarkBE_PutUint64_So(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		binary.BigEndian.PutUint64(sink.buf[:8], uint64(i))
		sinkInt = i
	}
}

func BenchmarkBE_AppendUint64_So(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		sink.res = binary.BigEndian.AppendUint64(sink.buf[:0], uint64(i))
	}
}

func BenchmarkLE_PutUint64_So(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		binary.LittleEndian.PutUint64(sink.buf[:8], uint64(i))
		sinkInt = i
	}
}

func BenchmarkLE_AppendUint64_So(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		sink.res = binary.LittleEndian.AppendUint64(sink.buf[:0], uint64(i))
	}
}
