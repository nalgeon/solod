package c_test

import (
	"solod.dev/so/c"
	"solod.dev/so/testing"
)

//so:embed testc.h
var testc_h string

//so:extern
func isalpha(ch int32) bool {
	_ = ch
	return false
}

//so:extern
func get_cstring(s string) *c.ConstChar {
	_ = s
	return nil
}

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
func ld_half(x c.LongDouble) c.LongDouble {
	_ = x
	return 0
}

func TestAssert(t *testing.T) {
	_ = t
	a, b := 11, 11
	c.Assert(a == b, "a != b")
}

func TestString(t *testing.T) {
	cstr := get_cstring("Hello, C!")
	str := c.String(cstr)
	if str != "Hello, C!" {
		t.Error("String = " + str + ", want Hello, C!")
	}
}

func TestExtern(t *testing.T) {
	if !isalpha('a') {
		t.Error("isalpha('a') = false")
	}
}

func TestVal(t *testing.T) {
	nan := c.Val[float64]("NAN")
	if nan == nan {
		t.Error("NAN == NAN")
	}
	x := c.Val[float64]("sqrt(49)")
	if x != 7 {
		t.Error("sqrt(49) != 7")
	}
}

func TestRaw(t *testing.T) {
	var b int
	c.Raw(`
	int a = 7;
	b = a * a;
	b = sqrt(b);
	`)
	if b != 7 {
		t.Error("Raw block: b != 7")
	}
}

func TestCString(t *testing.T) {
	_ = t
	s := "hello"
	p := (*c.ConstChar)(c.CString(s))
	_ = p
}

func TestNumericTypes(t *testing.T) {
	var i c.Int = 42
	var u c.UInt = c.UInt(i)
	var l c.Long = c.Long(u)
	var ul c.ULong = c.ULong(l)
	var ll c.LongLong = c.LongLong(ul)
	var ull c.ULongLong = c.ULongLong(ll)
	if ull != 42 {
		t.Error("ull != 42")
	}
}

func TestSizeTypes(t *testing.T) {
	if n := str_len("hello"); n != 5 {
		t.Error("str_len(hello) != 5")
	}
	if i := str_index("hello", 'l'); i != 2 {
		t.Error("str_index(hello, l) != 2")
	}
	if i := str_index("hello", 'z'); i != -1 {
		t.Error("str_index(hello, z) != -1")
	}
}

func TestPtrTypes(t *testing.T) {
	p := c.CString("hello")
	q := c.PtrAt(p, 3)
	if d := ptr_diff(q, p); d != 3 {
		t.Error("ptr_diff(q, p) != 3")
	}
	if ptr_addr(p) == 0 {
		t.Error("ptr_addr(p) = 0")
	}
}

func TestLongDouble(t *testing.T) {
	x := ld_half(9)
	if x != 4.5 {
		t.Error("ld_half(9) != 4.5")
	}
}
