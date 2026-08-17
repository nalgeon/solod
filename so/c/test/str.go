package c_test

import (
	"solod.dev/so/c"
	"solod.dev/so/testing"
)

//so:extern
func get_cstring(s string) *c.ConstChar {
	_ = s
	return nil
}

func TestString(t *testing.T) {
	cstr := get_cstring("Hello, C!")
	if str := c.String(cstr); str != "Hello, C!" {
		t.Errorf("String = %s, want Hello, C!", str)
	}
}

func TestStringEmpty(t *testing.T) {
	if str := c.String(get_cstring("")); str != "" {
		t.Errorf("String of an empty C string = %s, want an empty string", str)
	}
}

func TestStringNil(t *testing.T) {
	var p *c.ConstChar
	if str := c.String(p); str != "" {
		t.Errorf("String(nil) = %s, want an empty string", str)
	}
}

func TestStringChar(t *testing.T) {
	// String takes a *Char as well as a *ConstChar.
	p := c.CString("hello")
	if str := c.String(p); str != "hello" {
		t.Errorf("String = %s, want hello", str)
	}
}

func TestCString(t *testing.T) {
	s := "hello"
	// The C string ends with a null byte, so String finds the same content.
	if str := c.String(c.CString(s)); str != s {
		t.Errorf("String(CString(%s)) = %s, want %s", s, str, s)
	}
	if str := c.String(c.CString("")); str != "" {
		t.Errorf("String(CString()) = %s, want an empty string", str)
	}
}

func TestCStringCopy(t *testing.T) {
	s := "hello"

	// Each call copies the string to the stack, so two calls give two buffers.
	p := c.CString(s)
	q := c.CString(s)
	if p == q {
		t.Error("CString returns the string data, want a copy")
	}

	// The copy is writable, and the write leaves the So string alone.
	*c.PtrAt(p, 0) = '!'
	if str := c.String(p); str != "!ello" {
		t.Errorf("String(p) = %s, want !ello", str)
	}
	if s != "hello" {
		t.Errorf("s = %s, want hello", s)
	}
	if str := c.String(q); str != "hello" {
		t.Errorf("String(q) = %s, want hello", str)
	}
}
