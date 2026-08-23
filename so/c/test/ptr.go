package c_test

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/testing"
)

type point struct {
	x int32
	y int32
}

func TestSizeof(t *testing.T) {
	if n := c.Sizeof[byte](); n != 1 {
		t.Errorf("Sizeof[byte] = %d, want 1", n)
	}
	if n := c.Sizeof[int32](); n != 4 {
		t.Errorf("Sizeof[int32] = %d, want 4", n)
	}
	if n := c.Sizeof[int64](); n != 8 {
		t.Errorf("Sizeof[int64] = %d, want 8", n)
	}
	if n := c.Sizeof[float64](); n != 8 {
		t.Errorf("Sizeof[float64] = %d, want 8", n)
	}
	if n := c.Sizeof[point](); n != 8 {
		t.Errorf("Sizeof[point] = %d, want 8", n)
	}
}

func TestAlignof(t *testing.T) {
	if n := c.Alignof[byte](); n != 1 {
		t.Errorf("Alignof[byte] = %d, want 1", n)
	}
	if n := c.Alignof[int32](); n != 4 {
		t.Errorf("Alignof[int32] = %d, want 4", n)
	}
	// A struct aligns like its widest field.
	if n := c.Alignof[point](); n != 4 {
		t.Errorf("Alignof[point] = %d, want 4", n)
	}
}

func TestZero(t *testing.T) {
	if v := c.Zero[int32](); v != 0 {
		t.Errorf("Zero[int32] = %d, want 0", v)
	}
	if v := c.Zero[float64](); v != 0 {
		t.Errorf("Zero[float64] = %f, want 0", v)
	}
	if p := c.Zero[point](); p.x != 0 || p.y != 0 {
		t.Errorf("Zero[point] = {%d, %d}, want {0, 0}", p.x, p.y)
	}
}

func TestAlloca(t *testing.T) {
	// Alloca gives room for n elements of T.
	p := c.Alloca[int32](4)
	for i := range 4 {
		*c.PtrAt(p, i) = int32(i * 10)
	}
	for i := range 4 {
		if v := *c.PtrAt(p, i); v != int32(i*10) {
			t.Errorf("p[%d] = %d, want %d", i, v, i*10)
		}
	}
}

func TestPtrAdd(t *testing.T) {
	p := c.Alloca[int32](4)
	*c.PtrAt(p, 2) = 42

	// The offset counts elements, not bytes.
	if v := *c.PtrAdd(p, 2); v != 42 {
		t.Errorf("*PtrAdd(p, 2) = %d, want 42", v)
	}
	// One element of int32 is four bytes.
	if c.PtrAs[byte](c.PtrAdd(p, 1)) != c.PtrAdd(c.PtrAs[byte](p), 4) {
		t.Error("PtrAdd does not scale the offset by the element size")
	}
	// A zero offset returns the same pointer.
	if c.PtrAdd(p, 0) != p {
		t.Error("PtrAdd(p, 0) != p")
	}
}

func TestPtrAt(t *testing.T) {
	p := c.Alloca[int32](4)
	// PtrAt and PtrAdd are the same operation.
	for i := range 4 {
		if c.PtrAt(p, i) != c.PtrAdd(p, i) {
			t.Errorf("PtrAt(p, %d) != PtrAdd(p, %d)", i, i)
		}
	}
}

func TestPtrAs(t *testing.T) {
	var x int32 = -1

	// The cast keeps the address and reads the same bits as another type.
	u := c.PtrAs[uint32](&x)
	if *u != 4294967295 {
		t.Errorf("*PtrAs[uint32](&x) = %x, want ffffffff", *u)
	}
	*u = 0
	if x != 0 {
		t.Errorf("x = %d, want 0", x)
	}
}

func TestBytes(t *testing.T) {
	p := c.Alloca[byte](4)
	b := c.Bytes(p, 4)
	if len(b) != 4 {
		t.Errorf("len(Bytes(p, 4)) = %d, want 4", len(b))
	}

	// The slice shares the memory of the pointer, without a copy.
	b[0] = 7
	if *p != 7 {
		t.Errorf("*p = %d, want 7", *p)
	}
	*c.PtrAt(p, 3) = 9
	if b[3] != 9 {
		t.Errorf("b[3] = %d, want 9", b[3])
	}
}

func TestBytesNil(t *testing.T) {
	var p *byte
	b := c.Bytes(p, 4)
	if b != nil {
		t.Error("Bytes(nil, 4) != nil")
	}
	if len(b) != 0 {
		t.Errorf("len(Bytes(nil, 4)) = %d, want 0", len(b))
	}
}

func TestSlice(t *testing.T) {
	p := c.Alloca[int32](4)

	// The length and the capacity are separate.
	s := c.Slice(p, 2, 4)
	if len(s) != 2 || cap(s) != 4 {
		t.Errorf("Slice(p, 2, 4): len = %d, cap = %d, want 2 and 4", len(s), cap(s))
	}

	// The slice shares the memory of the pointer, without a copy.
	s[1] = 42
	if v := *c.PtrAt(p, 1); v != 42 {
		t.Errorf("p[1] = %d, want 42", v)
	}

	// A reslice reaches the capacity.
	s = s[:4]
	*c.PtrAt(p, 3) = 7
	if s[3] != 7 {
		t.Errorf("s[3] = %d, want 7", s[3])
	}
}

func TestSliceNil(t *testing.T) {
	var p *int32
	s := c.Slice(p, 2, 4)
	if s != nil {
		t.Error("Slice(nil, 2, 4) != nil")
	}
	if len(s) != 0 || cap(s) != 0 {
		t.Errorf("Slice(nil, 2, 4): len = %d, cap = %d, want 0 and 0", len(s), cap(s))
	}
}

func TestSliceData(t *testing.T) {
	b := []byte{1, 2, 3}

	// The pointer reads the slice data as a C type.
	if n := sum_bytes(c.SliceData[c.UChar](b), c.Size(len(b))); n != 6 {
		t.Errorf("sum_bytes(b) = %d, want 6", n)
	}

	// The pointer shares the memory of the slice, without a copy.
	p := c.SliceData[byte](b)
	*p = 9
	if b[0] != 9 {
		t.Errorf("b[0] = %d, want 9", b[0])
	}
	if c.SliceData[byte](b) != unsafe.SliceData(b) {
		t.Error("SliceData returns a copy, want the slice data")
	}
}

func TestSliceDataType(t *testing.T) {
	// The element type of the slice and the pointer type are independent,
	// so a []int32 also reads as bytes. The byte order does not change
	// the sum, so the test does not depend on the endianness.
	nums := []int32{1, 2, 3, 4}
	n := c.Size(len(nums) * c.Sizeof[int32]())
	if sum := sum_bytes(c.SliceData[c.UChar](nums), n); sum != 10 {
		t.Errorf("sum_bytes(nums) = %d, want 10", sum)
	}
}

func TestSliceDataNil(t *testing.T) {
	var b []byte
	if c.SliceData[byte](b) != nil {
		t.Error("SliceData(nil) != nil")
	}
}
