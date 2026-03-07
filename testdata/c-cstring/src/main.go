package main

import (
	"unsafe"

	"github.com/nalgeon/solod/so/c/cstring"
	"github.com/nalgeon/solod/so/c/stdlib"
)

func main() {
	size := 4 * unsafe.Sizeof(byte(0))

	// Memset: fill memory with a value.
	buf := stdlib.Malloc(size)
	if buf == nil {
		panic("malloc failed")
	}
	cstring.Memset(buf, 65, size)

	// Memcpy: copy to a new buffer.
	buf2 := stdlib.Malloc(size)
	if buf2 == nil {
		panic("malloc failed")
	}
	cstring.Memcpy(buf2, buf, size)

	// Memcmp: compare the two buffers.
	result := cstring.Memcmp(buf, buf2, size)
	if result != 0 {
		panic("want result == 0")
	}

	// Memmove: overlapping copy within buf.
	cstring.Memmove(buf, buf, size)
	result = cstring.Memcmp(buf, buf2, size)
	if result != 0 {
		panic("want result == 0")
	}

	stdlib.Free(buf)
	stdlib.Free(buf2)
}
