package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestBuffer_Stack(t *testing.T) {
	buf := bytes.NewBuffer(mem.System, []byte("hello world"))
	if buf.String() != "hello world" {
		t.Error("Buffer.String() != hello world")
	}
	rdbuf := make([]byte, 5)
	n, err := buf.Read(rdbuf)
	if n != 5 || string(rdbuf) != "hello" || err != nil {
		t.Error("Buffer.Read() != hello")
	}
	if buf.String() != " world" {
		t.Error("Buffer.Read() did not advance the buffer")
	}
}

func TestBuffer_Heap(t *testing.T) {
	buf := bytes.NewBuffer(mem.System, nil)
	defer buf.Free()

	buf.WriteString("hello")
	buf.WriteString(" world")
	if buf.String() != "hello world" {
		t.Error("Buffer.WriteString() != hello world")
	}

	buf.Grow(64)
	if buf.Cap() < 64 {
		t.Error("Buffer.Grow(64) did not grow capacity")
	}

	rdbuf := make([]byte, 5)
	n, err := buf.Read(rdbuf)
	if n != 5 || string(rdbuf) != "hello" || err != nil {
		t.Error("Buffer.Read() != hello")
	}
	if buf.String() != " world" {
		t.Error("Buffer.Read() did not advance the buffer")
	}
}
