package binary_test

import (
	"solod.dev/so/encoding/binary"
	"solod.dev/so/testing"
)

func TestBigEndian(t *testing.T) {
	const n1 uint64 = 0x0123456789abcdef
	const n2 uint64 = 0xfedcba9876543210

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n1)
	if binary.BigEndian.Uint64(buf) != n1 {
		t.Error("BigEndian: invalid decoded n1")
	}

	buf = binary.BigEndian.AppendUint64(buf[:0], n2)
	if binary.BigEndian.Uint64(buf) != n2 {
		t.Error("BigEndian: invalid decoded n2")
	}
}

func TestLittleEndian(t *testing.T) {
	const n1 uint64 = 0x0123456789abcdef
	const n2 uint64 = 0xfedcba9876543210

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n1)
	if binary.LittleEndian.Uint64(buf) != n1 {
		t.Error("LittleEndian: invalid decoded n1")
	}

	buf = binary.LittleEndian.AppendUint64(buf[:0], n2)
	if binary.LittleEndian.Uint64(buf) != n2 {
		t.Error("LittleEndian: invalid decoded n2")
	}
}
