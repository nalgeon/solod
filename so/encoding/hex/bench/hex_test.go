package hex_bench

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func BenchmarkEncode_Go(b *testing.B) {
	data := []byte{2, 3, 5, 7, 9, 11, 13, 17}
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat(data, size/8)
		sink = make([]byte, 2*size)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				hex.Encode(sink, src)
			}
		})
	}
}

func BenchmarkDecode_Go(b *testing.B) {
	data := []byte{'2', 'b', '7', '4', '4', 'f', 'a', 'a'}
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat(data, size/8)
		sink = make([]byte, size/2)
		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				hex.Decode(sink, src)
			}
		})
	}
}
