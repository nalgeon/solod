package c_test

import (
	"solod.dev/so/c"
	"solod.dev/so/testing"
)

//so:extern
func str_len(s string) c.Size {
	_ = s
	return 0
}

//so:extern
func str_index(s string, ch c.Char) c.SSize {
	_, _ = s, ch
	return 0
}

//so:extern
func ptr_diff(a *c.Char, b *c.Char) c.Ptrdiff {
	_, _ = a, b
	return 0
}

//so:extern
func ptr_addr(p any) c.Intptr {
	_ = p
	return 0
}

//so:extern
func find_first(items *c.ConstVoid, count c.Size, size c.Size, match func(item *c.ConstVoid) bool) c.SSize {
	_, _, _, _ = items, count, size, match
	return 0
}

//so:extern
func ld_half(x c.LongDouble) c.LongDouble {
	_ = x
	return 0
}

func TestCharTypes(t *testing.T) {
	if n := c.Sizeof[c.Char](); n != 1 {
		t.Errorf("Sizeof[Char] = %d, want 1", n)
	}
	if n := c.Sizeof[c.ConstChar](); n != 1 {
		t.Errorf("Sizeof[ConstChar] = %d, want 1", n)
	}
	if n := c.Sizeof[c.SChar](); n != 1 {
		t.Errorf("Sizeof[SChar] = %d, want 1", n)
	}
	if n := c.Sizeof[c.UChar](); n != 1 {
		t.Errorf("Sizeof[UChar] = %d, want 1", n)
	}

	// SChar is signed, UChar is unsigned.
	var s c.SChar = -1
	if s >= 0 {
		t.Errorf("SChar(-1) = %d, want a negative value", s)
	}
	var u c.UChar = 255
	u++
	if u != 0 {
		t.Errorf("UChar(255) + 1 = %d, want 0", u)
	}
}

func TestShortTypes(t *testing.T) {
	if n := c.Sizeof[c.Short](); n != 2 {
		t.Errorf("Sizeof[Short] = %d, want 2", n)
	}
	if n := c.Sizeof[c.UShort](); n != 2 {
		t.Errorf("Sizeof[UShort] = %d, want 2", n)
	}

	var s c.Short = -1
	if s >= 0 {
		t.Errorf("Short(-1) = %d, want a negative value", s)
	}
	var u c.UShort = 65535
	u++
	if u != 0 {
		t.Errorf("UShort(65535) + 1 = %d, want 0", u)
	}
}

func TestIntTypes(t *testing.T) {
	if n := c.Sizeof[c.Int](); n != 4 {
		t.Errorf("Sizeof[Int] = %d, want 4", n)
	}
	if n := c.Sizeof[c.UInt](); n != 4 {
		t.Errorf("Sizeof[UInt] = %d, want 4", n)
	}

	var i c.Int = -1
	if i >= 0 {
		t.Errorf("Int(-1) = %d, want a negative value", i)
	}
	var u c.UInt = 4294967295
	u++
	if u != 0 {
		t.Errorf("UInt(4294967295) + 1 = %x, want 0", u)
	}
}

func TestLongTypes(t *testing.T) {
	// C long is 8 bytes on 64-bit Unix and 4 bytes on Windows and wasm32,
	// so the test stays inside 32 bits.
	var l c.Long = -1
	if l >= 0 {
		t.Errorf("Long(-1) = %d, want a negative value", l)
	}
	var ul c.ULong = 1
	ul <<= 31
	if ul>>31 != 1 {
		t.Errorf("ULong(1 << 31) >> 31 = %x, want 1", ul)
	}

	// C long long is 8 bytes everywhere.
	if n := c.Sizeof[c.LongLong](); n != 8 {
		t.Errorf("Sizeof[LongLong] = %d, want 8", n)
	}
	if n := c.Sizeof[c.ULongLong](); n != 8 {
		t.Errorf("Sizeof[ULongLong] = %d, want 8", n)
	}
	var ll c.LongLong = 1
	ll <<= 62
	if ll>>62 != 1 {
		t.Errorf("LongLong(1 << 62) >> 62 = %d, want 1", ll)
	}
	var ull c.ULongLong = 1
	ull <<= 63
	if ull>>63 != 1 {
		t.Errorf("ULongLong(1 << 63) >> 63 = %x, want 1", ull)
	}
}

func TestLongDouble(t *testing.T) {
	if x := ld_half(9); x != 4.5 {
		t.Errorf("ld_half(9) = %f, want 4.5", float64(x))
	}
}

func TestSizeTypes(t *testing.T) {
	// Size follows the width of C size_t.
	if n := str_len("hello"); n != 5 {
		t.Errorf("str_len(hello) = %d, want 5", int(n))
	}
	if n := str_len(""); n != 0 {
		t.Errorf("str_len() = %d, want 0", int(n))
	}

	// SSize is signed, so it carries the -1 of a C function.
	if i := str_index("hello", 'l'); i != 2 {
		t.Errorf("str_index(hello, l) = %d, want 2", i)
	}
	if i := str_index("hello", 'z'); i != -1 {
		t.Errorf("str_index(hello, z) = %d, want -1", i)
	}
}

func TestPtrTypes(t *testing.T) {
	p := c.CString("hello")
	q := c.PtrAt(p, 3)

	// Ptrdiff holds the distance between two pointers.
	if d := ptr_diff(q, p); d != 3 {
		t.Errorf("ptr_diff(q, p) = %d, want 3", d)
	}
	if d := ptr_diff(p, q); d != -3 {
		t.Errorf("ptr_diff(p, q) = %d, want -3", d)
	}

	// Intptr holds an address.
	if ptr_addr(p) == 0 {
		t.Error("ptr_addr(p) = 0")
	}
	if ptr_addr(q)-ptr_addr(p) != 3 {
		t.Error("ptr_addr(q) - ptr_addr(p) != 3")
	}
}

// isOdd reports whether the item is an odd number.
func isOdd(item *c.ConstVoid) bool {
	return *c.PtrAs[int32](item)%2 != 0
}

func TestConstVoid(t *testing.T) {
	// ConstVoid matches a C const void* parameter,
	// in a function and in a function pointer.
	nums := []int32{4, 8, 9, 12}
	items := c.SliceData[c.ConstVoid](nums)
	size := c.Size(c.Sizeof[int32]())

	if i := find_first(items, c.Size(len(nums)), size, isOdd); i != 2 {
		t.Errorf("find_first(nums, isOdd) = %d, want 2", i)
	}
	if i := find_first(items, 2, size, isOdd); i != -1 {
		t.Errorf("find_first(nums[:2], isOdd) = %d, want -1", i)
	}
}
