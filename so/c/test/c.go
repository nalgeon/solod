package c_test

import (
	"solod.dev/so/c"
	"solod.dev/so/testing"
)

//so:embed testc.h
var testc_h string

//so:extern
func is_alpha(ch int32) bool {
	_ = ch
	return false
}

func TestExtern(t *testing.T) {
	if !is_alpha('a') {
		t.Error("is_alpha('a') = false")
	}
	if is_alpha('1') {
		t.Error("is_alpha('1') = true")
	}
}

func TestBitcast(t *testing.T) {
	// A float and its IEEE 754 bits convert both ways.
	if got := c.Bitcast[uint64](1.0); got != 0x3FF0000000000000 {
		t.Errorf("Bitcast[uint64](1.0) = %x, want 3ff0000000000000", got)
	}
	if got := c.Bitcast[float64](uint64(0x3FF0000000000000)); got != 1.0 {
		t.Errorf("Bitcast[float64](0x3ff0000000000000) = %v, want 1", got)
	}
	if got := c.Bitcast[uint32](float32(1.0)); got != 0x3F800000 {
		t.Errorf("Bitcast[uint32](1.0) = %x, want 3f800000", got)
	}
	if got := c.Bitcast[float32](uint32(0x3F800000)); got != 1.0 {
		t.Errorf("Bitcast[float32](0x3f800000) = %v, want 1", got)
	}

	// A round trip returns the original bits.
	var bits uint64 = 0x7FF8000000000001
	if got := c.Bitcast[uint64](c.Bitcast[float64](bits)); got != bits {
		t.Errorf("round trip = %x, want %x", got, bits)
	}

	// A signed value keeps its bit pattern.
	if got := c.Bitcast[uint32](int32(-1)); got != 0xFFFFFFFF {
		t.Errorf("Bitcast[uint32](-1) = %x, want ffffffff", got)
	}
}

func TestAssert(t *testing.T) {
	_ = t
	a, b := 11, 11
	c.Assert(a == b, "a != b")
	// A false condition panics, and Solod has no recover,
	// so a test cannot take the failing branch.
}

func TestAssume(t *testing.T) {
	// Assume generates no code, so a true condition changes nothing.
	p := c.Alloca[int32](1)
	c.Assume(p != nil)
	*p = 42
	if *p != 42 {
		t.Errorf("*p = %d, want 42", *p)
	}
}

func TestVal(t *testing.T) {
	// A C macro.
	if n := c.Val[c.Int]("TEST_ANSWER"); n != 42 {
		t.Errorf("TEST_ANSWER = %d, want 42", n)
	}
	// A C expression.
	if x := c.Val[float64]("1.0 / 4.0"); x != 0.25 {
		t.Errorf("1.0 / 4.0 = %f, want 0.25", x)
	}
	// A C function call.
	if n := c.Val[c.Int]("is_alpha('a')"); n == 0 {
		t.Error("is_alpha('a') = 0")
	}
}

func TestRaw(t *testing.T) {
	var b int
	c.Raw(`
	int a = 7;
	b = a * a;
	`)
	if b != 49 {
		t.Errorf("Raw block: b = %d, want 49", b)
	}
}

func TestRawVars(t *testing.T) {
	// The block reads the Solod variables of the function. Go does not see the
	// read, so the variable needs a blank assignment to stay in use.
	a := 7
	_ = a
	var b int
	c.Raw(`
	b = a + 1;
	`)
	if b != 8 {
		t.Errorf("Raw block: b = %d, want 8", b)
	}
}
