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

func TestAssert(t *testing.T) {
	_ = t
	a, b := 11, 11
	c.Assert(a == b, "a != b")
	// A false condition panics, and So has no recover,
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
	// The block reads the So variables of the function. Go does not see the
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
