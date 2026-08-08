package binary_bench

import (
	"encoding/binary"
	"testing"
)

var putbuf = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func BenchmarkBE_PutUint64_Go(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		binary.BigEndian.PutUint64(putbuf[:8], uint64(i))
	}
}

func BenchmarkBE_AppendUint64_Go(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		putbuf = binary.BigEndian.AppendUint64(putbuf[:0], uint64(i))
	}
}

func BenchmarkLE_PutUint64_Go(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		binary.LittleEndian.PutUint64(putbuf[:8], uint64(i))
	}
}

func BenchmarkLE_AppendUint64_Go(b *testing.B) {
	b.SetBytes(8)
	for i := range b.N {
		putbuf = binary.LittleEndian.AppendUint64(putbuf[:0], uint64(i))
	}
}
