package crand_test

import (
	"solod.dev/so/crypto/crand"
	"solod.dev/so/testing"
)

func TestRead(t *testing.T) {
	buf := make([]byte, 16)
	n, err := crand.Read(buf)
	if err != nil {
		t.Fatal("Read failed")
		return
	}
	if n != len(buf) {
		t.Error("short read of random data")
	}
}

func TestRead_Empty(t *testing.T) {
	buf := make([]byte, 0)
	n, err := crand.Read(buf)
	if err != nil {
		t.Fatal("Read failed")
		return
	}
	if n != 0 {
		t.Error("non-zero read of empty slice")
	}
}

func TestReader(t *testing.T) {
	buf := make([]byte, 16)
	n, err := crand.Reader.Read(buf)
	if err != nil {
		t.Fatal("Reader.Read failed")
		return
	}
	if n != len(buf) {
		t.Error("short read of random data")
	}
}

func TestText(t *testing.T) {
	buf := make([]byte, 26)
	s := crand.Text(buf)
	if len(s) != 26 {
		t.Error("unexpected length of random text")
	}
}
