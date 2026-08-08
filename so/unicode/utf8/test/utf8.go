package utf8_test

import (
	"solod.dev/so/unicode/utf8"
	"solod.dev/so/testing"
)

func TestDecodeLastRune(t *testing.T) {
	b := []byte("Hello, 世界")
	r, size := utf8.DecodeLastRune(b)
	if r != '界' || size != 3 {
		t.Error("DecodeLastRune failed")
	}
}

func TestDecodeLastRuneInString(t *testing.T) {
	str := "Hello, 世界"
	r, size := utf8.DecodeLastRuneInString(str)
	if r != '界' || size != 3 {
		t.Error("DecodeLastRuneInString failed")
	}
}

func TestDecodeRune(t *testing.T) {
	b := []byte("Hello, 世界")
	r, size := utf8.DecodeRune(b)
	if r != 'H' || size != 1 {
		t.Error("DecodeRune failed")
	}
}

func TestDecodeRuneInString(t *testing.T) {
	str := "Hello, 世界"
	r, size := utf8.DecodeRuneInString(str)
	if r != 'H' || size != 1 {
		t.Error("DecodeRuneInString failed")
	}
}

func TestEncodeRune(t *testing.T) {
	buf := make([]byte, 3)
	n := utf8.EncodeRune(buf, '界')
	if n != 3 || string(buf) != "界" {
		t.Error("EncodeRune failed")
	}
}

func TestRuneCount(t *testing.T) {
	n := utf8.RuneCount([]byte("Hello, 世界"))
	if n != 9 {
		t.Error("RuneCount failed")
	}
}

func TestRuneCountInString(t *testing.T) {
	n := utf8.RuneCountInString("Hello, 世界")
	if n != 9 {
		t.Error("RuneCountInString failed")
	}
}

func TestRuneLen(t *testing.T) {
	n := utf8.RuneLen('界')
	if n != 3 {
		t.Error("RuneLen failed")
	}
}

func TestValidString(t *testing.T) {
	if !utf8.ValidString("Hello, 世界") {
		t.Error("ValidString failed")
	}
}

func TestAppendRune(t *testing.T) {
	buf := make([]byte, 7, 10)
	copy(buf, []byte("Hello, "))
	buf = utf8.AppendRune(buf, '界')
	if string(buf) != "Hello, 界" {
		t.Error("AppendRune failed")
	}
}
