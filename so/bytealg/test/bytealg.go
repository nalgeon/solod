package bytealg_test

import (
	"solod.dev/so/bytealg"
	"solod.dev/so/testing"
)

// Compares byte slices lexicographically.
func TestCompare(t *testing.T) {
	b := []byte("abc")
	if bytealg.Compare(b, []byte("abb")) <= 0 {
		t.Error("Compare(abc, abb) <= 0")
	}
	if bytealg.Compare(b, []byte("abd")) >= 0 {
		t.Error("Compare(abc, abd) >= 0")
	}
	if bytealg.Compare(b, []byte("abc")) != 0 {
		t.Error("Compare(abc, abc) != 0")
	}
}

// Counts byte occurrences in a byte slice and in a string.
func TestCount(t *testing.T) {
	if n := bytealg.Count([]byte("hello world"), 'o'); n != 2 {
		t.Error("Count(hello world, o) != 2")
	}
	if n := bytealg.CountString("hello world", 'o'); n != 2 {
		t.Error("CountString(hello world, o) != 2")
	}
}

// Reports whether two byte slices are equal.
func TestEqual(t *testing.T) {
	a := []byte("hello")
	if !bytealg.Equal(a, []byte("hello")) {
		t.Error("Equal(hello, hello) = false")
	}
	if bytealg.Equal(a, []byte("world")) {
		t.Error("Equal(hello, world) = true")
	}
}

// Finds a substring with the Rabin-Karp algorithm.
func TestIndexRabinKarp(t *testing.T) {
	b := []byte("go is fun")
	if idx := bytealg.IndexRabinKarp(b, []byte("is")); idx != 3 {
		t.Error("IndexRabinKarp(go is fun, is) != 3")
	}
}

// Finds the last occurrence of a substring with Rabin-Karp.
func TestLastIndexRabinKarp(t *testing.T) {
	b := []byte("hello")
	if idx := bytealg.LastIndexRabinKarp(b, []byte("l")); idx != 3 {
		t.Error("LastIndexRabinKarp(hello, l) != 3")
	}
}

// Finds a byte in a byte slice and in a string.
func TestIndexByte(t *testing.T) {
	if idx := bytealg.IndexByte([]byte("hello"), 'l'); idx != 2 {
		t.Error("IndexByte(hello, l) != 2")
	}
	if idx := bytealg.IndexByteString("hello", 'l'); idx != 2 {
		t.Error("IndexByteString(hello, l) != 2")
	}
}

// Finds the last occurrence of a byte in a byte slice and in a string.
func TestLastIndexByte(t *testing.T) {
	if idx := bytealg.LastIndexByte([]byte("hello"), 'l'); idx != 3 {
		t.Error("LastIndexByte(hello, l) != 3")
	}
	if idx := bytealg.LastIndexByteString("hello", 'l'); idx != 3 {
		t.Error("LastIndexByteString(hello, l) != 3")
	}
}
