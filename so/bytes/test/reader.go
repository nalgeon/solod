package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestReader_Read(t *testing.T) {
	s := "hello world"
	r := bytes.NewReader([]byte(s))
	if r.Len() != len(s) {
		t.Error("Reader.Len() != len(input) before read")
	}

	b, err := io.ReadAll(nil, &r)
	if err != nil {
		t.Fatal("ReadAll failed")
		return
	}
	defer mem.FreeSlice(nil, b)

	if string(b) != s {
		t.Error("Reader read wrong content")
	}
	if r.Len() != 0 {
		t.Error("Reader.Len() != 0 after read")
	}
}
