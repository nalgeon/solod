// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"testing"

	"solod.dev/so/io"
)

// A panicReader panics from Read if panic is set.
type panicReader struct{ panic bool }

func (r panicReader) Read(p []byte) (int, error) {
	if r.panic {
		panic("oops")
	}
	return 0, io.EOF
}

func TestReadFromPanicReader(t *testing.T) {
	// A reader that gives no bytes leaves the buffer empty.
	var buf Buffer
	n, err := buf.ReadFrom(panicReader{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("ReadFrom = %d, want 0", n)
	}
	if buf.Len() != 0 {
		t.Fatalf("buf.Len() = %d, want 0", buf.Len())
	}

	// A reader that panics also leaves the buffer empty.
	var buf2 Buffer
	defer func() {
		recover()
		if buf2.Len() != 0 {
			t.Errorf("after the panic buf2.Len() = %d, want 0", buf2.Len())
		}
	}()
	buf2.ReadFrom(panicReader{panic: true})
}

func TestGrowOverflow(t *testing.T) {
	defer func() {
		want := "bytes: buffer overflow"
		if reason := recover(); reason != want {
			t.Errorf("after a too large Grow, recover() = %v, want %v", reason, want)
		}
	}()

	buf := NewBuffer(nil, make([]byte, 1))
	const maxInt = int(^uint(0) >> 1)
	buf.Grow(maxInt)
}

func TestGrowNegative(t *testing.T) {
	defer func() {
		if reason := recover(); reason == nil {
			t.Error("Grow(-1) did not panic")
		}
	}()
	var buf Buffer
	buf.Grow(-1)
}
