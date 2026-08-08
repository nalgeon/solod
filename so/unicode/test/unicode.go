package unicode_test

import (
	"solod.dev/so/unicode"
	"solod.dev/so/testing"
)

func TestIs(t *testing.T) {
	if !unicode.IsDigit('0') {
		t.Error("IsDigit failed")
	}
	if !unicode.IsLetter('a') {
		t.Error("IsLetter failed")
	}
	if !unicode.IsLower('a') {
		t.Error("IsLower failed")
	}
	if !unicode.IsSpace(' ') {
		t.Error("IsSpace failed")
	}
	if !unicode.IsTitle('ᾭ') {
		t.Error("IsTitle failed")
	}
	if !unicode.IsUpper('A') {
		t.Error("IsUpper failed")
	}
}

func TestTo(t *testing.T) {
	if unicode.ToLower('A') != 'a' {
		t.Error("ToLower failed")
	}
	if unicode.ToTitle('a') != 'A' {
		t.Error("ToTitle failed")
	}
	if unicode.ToUpper('a') != 'A' {
		t.Error("ToUpper failed")
	}
	if unicode.To(unicode.UpperCase, 'a') != 'A' {
		t.Error("To failed")
	}
}
