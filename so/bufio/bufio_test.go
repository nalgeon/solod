// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio

import (
	stdbufio "bufio"
	stdio "io"
	stdstrings "strings"
	"testing"

	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
)

// The classes of the errors the fuzzers compare. An error of so/bufio and the
// matching error of bufio are different values, so the fuzzers compare the
// class of an error instead of the error.
const (
	kindNil = iota
	kindEOF
	kindBufferFull
	kindTooLong
	kindBadReadCount
	kindOther
)

// errKind returns the class of the error.
func errKind(err error) int {
	switch err {
	case nil:
		return kindNil
	case io.EOF, stdio.EOF:
		return kindEOF
	case ErrBufferFull, stdbufio.ErrBufferFull:
		return kindBufferFull
	case ErrTooLong, stdbufio.ErrTooLong:
		return kindTooLong
	case ErrBadReadCount, stdbufio.ErrBadReadCount:
		return kindBadReadCount
	}
	return kindOther
}

type negativeReader int

func (r *negativeReader) Read([]byte) (int, error) { return -1, nil }

func TestNegativeRead(t *testing.T) {
	// The panic must point at the reader, not at bufio itself. It must not be
	// a slice index panic, for example.
	b := NewReader(nil, new(negativeReader))
	defer func() {
		switch err := recover().(type) {
		case nil:
			t.Fatal("read did not panic")
		case error:
			if !strings.Contains(err.Error(), "reader returned negative count from Read") {
				t.Fatalf("wrong panic: %v", err)
			}
		default:
			t.Fatalf("unexpected panic value: %T(%v)", err, err)
		}
	}()
	b.Read(make([]byte, 100))
}

func FuzzReader(f *testing.F) {
	// The size of the buffer and the size of a read decide how often the
	// reader fills the buffer, so the fuzzer varies both.
	f.Add([]byte(""), uint8(16), uint8(1))
	f.Add([]byte("hello, world."), uint8(16), uint8(1))
	f.Add([]byte("hello, world."), uint8(4), uint8(200))
	f.Add([]byte("first line\nsecond line\n"), uint8(200), uint8(3))
	f.Add(make([]byte, 4096), uint8(64), uint8(255))

	f.Fuzz(func(t *testing.T, data []byte, bufSize, readSize uint8) {
		size := int(bufSize)%128 + 1
		chunk := int(readSize)%64 + 1

		sr := strings.NewReader(string(data))
		r := NewReaderSize(mem.System, &sr, size)
		defer r.Free()

		refSr := stdstrings.NewReader(string(data))
		refR := stdbufio.NewReaderSize(refSr, size)

		buf := make([]byte, chunk)
		refBuf := make([]byte, chunk)
		for {
			n, err := r.Read(buf)
			refN, refErr := refR.Read(refBuf)
			if n != refN || errKind(err) != errKind(refErr) {
				t.Fatalf("Read() = %d, %v, want %d, %v", n, err, refN, refErr)
			}
			if string(buf[:n]) != string(refBuf[:refN]) {
				t.Fatal("Read() read the wrong bytes")
			}
			if err != nil {
				break
			}
		}
	})
}

func FuzzReaderReadSlice(f *testing.F) {
	// A line longer than the buffer gives ErrBufferFull, so the size of the
	// buffer decides where the reader stops.
	f.Add([]byte(""), byte('\n'), uint8(16))
	f.Add([]byte("hello, world."), byte('\n'), uint8(16))
	f.Add([]byte("first line\nsecond line\n"), byte('\n'), uint8(16))
	f.Add([]byte("a,b,,c,"), byte(','), uint8(4))

	f.Fuzz(func(t *testing.T, data []byte, delim byte, bufSize uint8) {
		size := int(bufSize)%128 + 1

		sr := strings.NewReader(string(data))
		r := NewReaderSize(mem.System, &sr, size)
		defer r.Free()

		refSr := stdstrings.NewReader(string(data))
		refR := stdbufio.NewReaderSize(refSr, size)

		for {
			line, err := r.ReadSlice(delim)
			refLine, refErr := refR.ReadSlice(delim)
			if errKind(err) != errKind(refErr) {
				t.Fatalf("ReadSlice() error = %v, want %v", err, refErr)
			}
			if string(line) != string(refLine) {
				t.Fatalf("ReadSlice() = %q, want %q", line, refLine)
			}
			if err != nil {
				break
			}
		}
	})
}

func FuzzReaderReadLine(f *testing.F) {
	f.Add([]byte(""), uint8(16))
	f.Add([]byte("hello, world."), uint8(16))
	f.Add([]byte("first line\r\nsecond line\n\n"), uint8(16))
	f.Add([]byte("a long line that does not fit the buffer\n"), uint8(4))

	f.Fuzz(func(t *testing.T, data []byte, bufSize uint8) {
		size := int(bufSize)%128 + 1

		sr := strings.NewReader(string(data))
		r := NewReaderSize(mem.System, &sr, size)
		defer r.Free()

		refSr := stdstrings.NewReader(string(data))
		refR := stdbufio.NewReaderSize(refSr, size)

		for {
			res := r.ReadLine()
			refLine, refPrefix, refErr := refR.ReadLine()
			if res.IsPrefix != refPrefix || errKind(res.Err) != errKind(refErr) {
				t.Fatalf("ReadLine() = %v, %v, want %v, %v", res.IsPrefix, res.Err, refPrefix, refErr)
			}
			if string(res.Line) != string(refLine) {
				t.Fatalf("ReadLine() = %q, want %q", res.Line, refLine)
			}
			if res.Err != nil {
				break
			}
		}
	})
}

func FuzzWriter(f *testing.F) {
	// The size of the buffer and the size of a write decide how often the
	// writer flushes the buffer, so the fuzzer varies both.
	f.Add([]byte(""), uint8(16), uint8(1))
	f.Add([]byte("hello, world."), uint8(16), uint8(1))
	f.Add([]byte("hello, world."), uint8(4), uint8(200))
	f.Add(make([]byte, 4096), uint8(64), uint8(255))

	f.Fuzz(func(t *testing.T, data []byte, bufSize, writeSize uint8) {
		size := int(bufSize)%128 + 1
		chunk := int(writeSize)%64 + 1

		var dst stdstrings.Builder
		w := NewWriterSize(mem.System, &dst, size)
		defer w.Free()

		var refDst stdstrings.Builder
		refW := stdbufio.NewWriterSize(&refDst, size)

		for i := 0; i < len(data); i += chunk {
			end := min(i+chunk, len(data))
			n, err := w.Write(data[i:end])
			refN, refErr := refW.Write(data[i:end])
			if n != refN || errKind(err) != errKind(refErr) {
				t.Fatalf("Write() = %d, %v, want %d, %v", n, err, refN, refErr)
			}
			if w.Buffered() != refW.Buffered() {
				t.Fatalf("Buffered() = %d, want %d", w.Buffered(), refW.Buffered())
			}
			if dst.String() != refDst.String() {
				t.Fatal("Write() wrote the wrong bytes")
			}
		}

		if err := w.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
		if err := refW.Flush(); err != nil {
			t.Fatalf("std Flush() error = %v", err)
		}
		if dst.String() != refDst.String() {
			t.Fatal("Flush() wrote the wrong bytes")
		}
	})
}
