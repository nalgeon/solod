package io_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

type reader struct {
	b []byte
}

func (r *reader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

type writer struct {
	b []byte
}

func (w *writer) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}

func TestCopy(t *testing.T) {
	r := reader{b: []byte("hello world")}
	w := writer{b: make([]byte, 0, 11)}
	if _, err := io.Copy(&w, &r); err != nil {
		t.Fatal("Copy failed")
		return
	}
	if string(w.b) != "hello world" {
		t.Error("Copy: wrong output")
	}
}

func TestCopyN(t *testing.T) {
	r := reader{b: []byte("hello world")}
	w := writer{b: make([]byte, 0, 5)}
	if _, err := io.CopyN(&w, &r, 5); err != nil {
		t.Fatal("CopyN failed")
		return
	}
	if string(w.b) != "hello" {
		t.Error("CopyN: wrong output")
	}
}

func TestReadAll(t *testing.T) {
	r := reader{b: []byte("hello world")}
	buf, err := io.ReadAll(mem.System, &r)
	if err != nil {
		t.Fatal("ReadAll failed")
		return
	}
	defer mem.FreeSlice(mem.System, buf)

	if string(buf) != "hello world" {
		t.Error("ReadAll: wrong output")
	}
}

func TestReadFull(t *testing.T) {
	r := reader{b: []byte("hello world")}
	buf := make([]byte, 11)
	if _, err := io.ReadFull(&r, buf); err != nil {
		t.Fatal("ReadFull failed")
		return
	}
	if string(buf) != "hello world" {
		t.Error("ReadFull: wrong output")
	}
}

func TestWriteString(t *testing.T) {
	w := writer{b: make([]byte, 0, 11)}
	n, err := io.WriteString(&w, "hello world")
	if err != nil {
		t.Fatal("WriteString failed")
		return
	}
	if n != 11 || string(w.b) != "hello world" {
		t.Error("WriteString: wrong output")
	}
}

func TestLimitReader(t *testing.T) {
	r := reader{b: []byte("hello world")}
	lr := io.LimitReader(&r, 5)
	buf := make([]byte, 5)
	if _, err := lr.Read(buf); err != nil {
		t.Fatal("LimitReader failed")
		return
	}
	if string(buf) != "hello" {
		t.Error("LimitReader: wrong output")
	}
}
