// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/math/rand"
	"solod.dev/so/mem"
	"solod.dev/so/strconv"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

func TestReadWriteRune(t *testing.T) {
	const nRune = 1000
	alloc := t.Allocator()
	byteBuf := bytes.NewBuffer(alloc, nil)
	w := bufio.NewWriter(alloc, &byteBuf)

	// Write the runes with WriteRune.
	var buf [utf8.UTFMax]byte
	for r := range rune(nRune) {
		size := utf8.EncodeRune(buf[:], r)
		nbytes, err := w.WriteRune(r)
		if err != nil {
			t.Errorf("WriteRune(%x) = %s, want nil", r, errName(errCode(err)))
			break
		}
		if nbytes != size {
			t.Errorf("WriteRune(%x) = %d bytes, want %d", r, nbytes, size)
			break
		}
	}
	w.Flush()

	// Read the runes back with ReadRune.
	r := bufio.NewReader(alloc, &byteBuf)
	for r1 := range rune(nRune) {
		size := utf8.EncodeRune(buf[:], r1)
		res := r.ReadRune()
		if res.Rune != r1 || res.Size != size || res.Err != nil {
			t.Errorf("ReadRune() = %x, %d, %s, want %x, %d, nil",
				res.Rune, res.Size, errName(errCode(res.Err)), r1, size)
			break
		}
	}

	r.Free()
	w.Free()
	byteBuf.Free()
}

func TestWriteInvalidRune(t *testing.T) {
	runes := []rune{-1, utf8.MaxRune + 1}
	for i := range runes {
		var arr [8]byte
		buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
		w := bufio.NewWriter(t.Allocator(), &buf)
		w.WriteRune(runes[i])
		w.Flush()
		if s := buf.String(); s != "�" {
			t.Errorf("WriteRune(%d) wrote %s, want the replacement character",
				runes[i], s)
		}
		w.Free()
	}
}

func TestWriter(t *testing.T) {
	// Write a number of bytes with every buffer size and check
	// that the writer passes the right bytes to the underlying writer.
	const dataSize = 8192
	const arenaSize = 8 * 1024
	alloc := t.Allocator()

	data := mem.AllocSlice[byte](alloc, dataSize, dataSize)
	for i := range data {
		data[i] = byte(' ' + i%('~'-' '))
	}
	out := mem.AllocSlice[byte](alloc, 0, dataSize)
	w := bytes.NewBuffer(alloc, out)
	backing := mem.AllocSlice[byte](alloc, arenaSize, arenaSize)
	arena := mem.NewArena(backing)

	for i := range bufSizes {
		for j := range bufSizes {
			nwrite := bufSizes[i]
			bs := bufSizes[j]

			// Write nwrite bytes with the buffer size bs.
			w.Reset()
			buf := bufio.NewWriterSize(&arena, &w, bs)
			n, err := buf.Write(data[0:nwrite])
			if err != nil || n != nwrite {
				t.Errorf("nwrite=%d bufsize=%d: Write() = %d, %s, want %d, nil",
					nwrite, bs, n, errName(errCode(err)), nwrite)
				buf.Free()
				arena.Reset()
				continue
			}
			if err := buf.Flush(); err != nil {
				t.Errorf("nwrite=%d bufsize=%d: Flush() = %s, want nil",
					nwrite, bs, errName(errCode(err)))
				buf.Free()
				arena.Reset()
				continue
			}
			written := w.Bytes()
			if len(written) != nwrite {
				t.Errorf("nwrite=%d bufsize=%d: wrote %d bytes, want %d",
					nwrite, bs, len(written), nwrite)
			} else if string(written) != string(data[:len(written)]) {
				t.Errorf("nwrite=%d bufsize=%d: wrote the wrong bytes", nwrite, bs)
			}
			buf.Free()
			arena.Reset()
		}
	}

	mem.FreeSlice(alloc, backing)
	w.Free()
	mem.FreeSlice(alloc, data)
}

func TestWriterAppend(t *testing.T) {
	// The largest number and the space take 11 bytes. One more byte
	// leaves room for a shift of the available buffer.
	const maxNumLen = 12
	const wantSize = 4096
	alloc := t.Allocator()

	got := bytes.NewBuffer(alloc, nil)
	want := mem.AllocSlice[byte](alloc, 0, wantSize)
	pcg := rand.NewPCG(0, 0)
	rn := rand.New(&pcg)
	w := bufio.NewWriterSize(alloc, &got, 64)

	for range 100 {
		// Obtain a buffer to append to.
		b := w.AvailableBuffer()
		if w.Available() != cap(b) {
			t.Errorf("Available() = %d, want %d", w.Available(), cap(b))
			break
		}

		// Solod has no append that grows a slice, so the buffer must have
		// room for the whole number.
		if cap(b) < maxNumLen {
			w.Flush()
			b = w.AvailableBuffer()
		}

		// While not recommended, it is valid to append to a shifted buffer.
		// This makes Write copy the input.
		if rn.IntN(8) == 0 && cap(b) > 0 {
			b = b[1:1]
		}

		// Append a random integer of a varying width.
		n := int64(rn.IntN(1 << rn.IntN(30)))
		want = append(strconv.AppendInt(want, n, 10), ' ')
		b = append(strconv.AppendInt(b, n, 10), ' ')
		w.Write(b)
	}
	w.Flush()

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("output = %s, want %s", string(got.Bytes()), string(want))
	}

	w.Free()
	mem.FreeSlice(alloc, want)
	got.Free()
}

// writeErrorCase is a case of the write errors test.
type writeErrorCase struct {
	n, m   int
	err    int
	expect int
}

var writeErrorCases = []writeErrorCase{
	{0, 1, errNone, errShortWrite},
	{1, 2, errNone, errShortWrite},
	{1, 1, errNone, errNone},
	{0, 1, errClosedPipe, errClosedPipe},
	{1, 2, errClosedPipe, errClosedPipe},
	{1, 1, errClosedPipe, errClosedPipe},
}

func TestWriteErrors(t *testing.T) {
	for i := range writeErrorCases {
		tc := writeErrorCases[i]
		w := errorWriter{n: tc.n, m: tc.m, err: tc.err}
		buf := bufio.NewWriter(t.Allocator(), &w)
		if _, err := buf.Write([]byte("hello world")); err != nil {
			t.Errorf("case %d: Write() = %s, want nil", i, errName(errCode(err)))
			buf.Free()
			continue
		}
		// Two flushes, to check that the error is sticky.
		for j := range 2 {
			err := buf.Flush()
			if errCode(err) != tc.expect {
				t.Errorf("case %d: Flush %d/2 = %s, want %s",
					i, j+1, errName(errCode(err)), errName(tc.expect))
			}
		}
		buf.Free()
	}
}

func TestWriteString(t *testing.T) {
	const bufSize = 8
	var arr [64]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	b := bufio.NewWriterSize(t.Allocator(), &buf, bufSize)
	defer b.Free()

	b.WriteString("0")                         // easy
	b.WriteString("123456")                    // still easy
	b.WriteString("7890")                      // easy after flush
	b.WriteString("abcdefghijklmnopqrstuvwxy") // hard
	b.WriteString("z")
	if err := b.Flush(); err != nil {
		t.Errorf("Flush() = %s, want nil", errName(errCode(err)))
	}
	const want = "01234567890abcdefghijklmnopqrstuvwxyz"
	if buf.String() != want {
		t.Errorf("WriteString() wrote %s, want %s", buf.String(), want)
	}
}

func TestNewWriterSizeIdempotent(t *testing.T) {
	const bufSize = 1000
	alloc := t.Allocator()
	var arr [1]byte
	out := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	b := bufio.NewWriterSize(alloc, &out, bufSize)

	// The writer must recognize itself.
	b1 := bufio.NewWriterSize(alloc, &b, bufSize)
	if b1.Size() != b.Size() {
		t.Errorf("NewWriterSize: Size() = %d, want %d", b1.Size(), b.Size())
	}

	// The writer must wrap a buffer that is too small.
	b2 := bufio.NewWriterSize(alloc, &b, 2*bufSize)
	if b2.Size() == b.Size() {
		t.Errorf("NewWriterSize: Size() = %d, want a larger buffer", b2.Size())
	}

	b2.Free()
	b.Free()
}

func TestWriterReadFrom(t *testing.T) {
	const size = 8192
	alloc := t.Allocator()
	input := mem.AllocSlice[byte](alloc, size, size)
	fillTestInput(input)
	outBuf := mem.AllocSlice[byte](alloc, 0, size)

	modes := []int{wrapFull, wrapDataErr}
	for i := range modes {
		out := bytes.NewBuffer(mem.NoAlloc, outBuf)
		w := bufio.NewWriter(alloc, &out)
		br := bytes.NewReader(input)
		wrap := wrapReader{r: &br, mode: modes[i]}

		n, err := w.ReadFrom(&wrap)
		if err != nil || int(n) != len(input) {
			t.Errorf("mode %d: ReadFrom() = %d, %s, want %d, nil",
				i, n, errName(errCode(err)), len(input))
		} else if err := w.Flush(); err != nil {
			t.Errorf("mode %d: Flush() = %s, want nil", i, errName(errCode(err)))
		} else if string(out.Bytes()) != string(input) {
			t.Errorf("mode %d: ReadFrom() wrote the wrong bytes", i)
		}
		w.Free()
	}

	mem.FreeSlice(alloc, outBuf)
	mem.FreeSlice(alloc, input)
}

// readerFromCase is a case of the ReadFrom errors test.
type readerFromCase struct {
	rn, wn   int
	rerr     int
	werr     int
	expected int
}

var readerFromCases = []readerFromCase{
	{0, 1, errEOF, errNone, errNone},
	{1, 1, errEOF, errNone, errNone},
	{0, 1, errClosedPipe, errNone, errClosedPipe},
	{0, 0, errClosedPipe, errShortWrite, errClosedPipe},
	{1, 0, errNone, errShortWrite, errShortWrite},
}

func TestWriterReadFromErrors(t *testing.T) {
	for i := range readerFromCases {
		tc := readerFromCases[i]
		rw := errorReadWriter{rn: tc.rn, wn: tc.wn, rerr: tc.rerr, werr: tc.werr}
		w := bufio.NewWriter(t.Allocator(), &rw)
		_, err := w.ReadFrom(&rw)
		if errCode(err) != tc.expected {
			t.Errorf("case %d: ReadFrom() = %s, want %s",
				i, errName(errCode(err)), errName(tc.expected))
		}
		w.Free()
	}
}

func TestWriterReadFromCounts(t *testing.T) {
	// Check that a copy into a writer does not flush the buffer too early.
	// For example, when the writes go to a network socket,
	// the program must avoid the excessive network writes.
	alloc := t.Allocator()
	xs := mem.AllocSlice[byte](alloc, 1200, 1200)
	for i := range xs {
		xs[i] = 'x'
	}

	var w0 countingWriter
	b0 := bufio.NewWriterSize(alloc, &w0, 1234)
	b0.WriteString(string(xs[:1000]))
	if w0.count != 0 {
		t.Errorf("write 1000 xs: writes = %d, want 0", w0.count)
	}
	b0.WriteString(string(xs[:200]))
	if w0.count != 0 {
		t.Errorf("write 1200 xs: writes = %d, want 0", w0.count)
	}
	sr := strings.NewReader(string(xs[:30]))
	io.Copy(&b0, &sr)
	if w0.count != 0 {
		t.Errorf("write 1230 xs: writes = %d, want 0", w0.count)
	}
	sr = strings.NewReader(string(xs[:9]))
	io.Copy(&b0, &sr)
	if w0.count != 1 {
		t.Errorf("write 1239 xs: writes = %d, want 1", w0.count)
	}
	b0.Free()

	var w1 countingWriter
	b1 := bufio.NewWriterSize(alloc, &w1, 1234)
	b1.WriteString(string(xs[:1200]))
	b1.Flush()
	if w1.count != 1 {
		t.Errorf("flush 1200 xs: writes = %d, want 1", w1.count)
	}
	b1.WriteString(string(xs[:89]))
	if w1.count != 1 {
		t.Errorf("write 1200 + 89 xs: writes = %d, want 1", w1.count)
	}
	sr = strings.NewReader(string(xs[:700]))
	io.Copy(&b1, &sr)
	if w1.count != 1 {
		t.Errorf("write 1200 + 789 xs: writes = %d, want 1", w1.count)
	}
	sr = strings.NewReader(string(xs[:600]))
	io.Copy(&b1, &sr)
	if w1.count != 2 {
		t.Errorf("write 1200 + 1389 xs: writes = %d, want 2", w1.count)
	}
	b1.Flush()
	if w1.count != 3 {
		t.Errorf("flush 1200 + 1389 xs: writes = %d, want 3", w1.count)
	}
	b1.Free()

	mem.FreeSlice(alloc, xs)
}

func TestWriterReadFromWhileFull(t *testing.T) {
	var arr [32]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	w := bufio.NewWriterSize(t.Allocator(), &buf, 10)
	defer w.Free()

	// Fill the buffer exactly.
	n, err := w.Write([]byte("0123456789"))
	if n != 10 || err != nil {
		t.Errorf("Write() = %d, %s, want 10, nil", n, errName(errCode(err)))
		return
	}

	// Read some data with ReadFrom.
	sr := strings.NewReader("abcdef")
	n2, err := w.ReadFrom(&sr)
	if n2 != 6 || err != nil {
		t.Errorf("ReadFrom() = %d, %s, want 6, nil", n2, errName(errCode(err)))
	}
}

func TestWriterReadFromUntilEOF(t *testing.T) {
	var arr [32]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	w := bufio.NewWriterSize(t.Allocator(), &buf, 5)
	defer w.Free()

	// Fill the buffer in part.
	n, err := w.Write([]byte("0123"))
	if n != 4 || err != nil {
		t.Errorf("Write() = %d, %s, want 4, nil", n, errName(errCode(err)))
		return
	}

	// Read some data with ReadFrom.
	sr := strings.NewReader("abcd")
	r := emptyThenNonEmptyReader{r: &sr, n: 3}
	n2, err := w.ReadFrom(&r)
	if n2 != 4 || err != nil {
		t.Errorf("ReadFrom() = %d, %s, want 4, nil", n2, errName(errCode(err)))
		return
	}
	w.Flush()
	if buf.String() != "0123abcd" {
		t.Errorf("ReadFrom() wrote %s, want 0123abcd", buf.String())
	}
}

func TestWriterReadFromErrNoProgress(t *testing.T) {
	var arr [32]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])
	w := bufio.NewWriterSize(t.Allocator(), &buf, 5)
	defer w.Free()

	// Fill the buffer in part.
	n, err := w.Write([]byte("0123"))
	if n != 4 || err != nil {
		t.Errorf("Write() = %d, %s, want 4, nil", n, errName(errCode(err)))
		return
	}

	// Read some data with ReadFrom.
	sr := strings.NewReader("abcd")
	r := emptyThenNonEmptyReader{r: &sr, n: 100}
	n2, err := w.ReadFrom(&r)
	if n2 != 0 || errCode(err) != errNoProgress {
		t.Errorf("ReadFrom() = %d, %s, want 0, ErrNoProgress", n2, errName(errCode(err)))
	}
}

func TestWriterReset(t *testing.T) {
	alloc := t.Allocator()
	var arr1, arr2, arr3, arr4 [8]byte
	buf1 := bytes.NewBuffer(mem.NoAlloc, arr1[:0])
	buf2 := bytes.NewBuffer(mem.NoAlloc, arr2[:0])
	buf3 := bytes.NewBuffer(mem.NoAlloc, arr3[:0])
	buf4 := bytes.NewBuffer(mem.NoAlloc, arr4[:0])

	w := bufio.NewWriter(alloc, &buf1)
	w.WriteString("foo")

	w.Reset(&buf2) // the writer drops the buffered data
	w.WriteString("bar")
	w.Flush()
	if buf1.String() != "" {
		t.Errorf("buf1 = %s, want the empty string", buf1.String())
	}
	if buf2.String() != "bar" {
		t.Errorf("buf2 = %s, want bar", buf2.String())
	}
	w.Free()

	// Reset on the zero value of a writer allocates the internal buffer.
	w = bufio.Writer{}
	w.Reset(&buf3)
	w.WriteString("bar")
	w.Flush()
	if buf3.String() != "bar" {
		t.Errorf("buf3 = %s, want bar", buf3.String())
	}

	// A reset of a writer to itself does nothing.
	w.Reset(&buf4)
	w.Reset(&w)
	w.WriteString("recur")
	w.Flush()
	if buf4.String() != "recur" {
		t.Errorf("buf4 = %s, want recur", buf4.String())
	}
	w.Free()
}

func TestWriterSize(t *testing.T) {
	alloc := t.Allocator()
	var arr [1]byte
	buf := bytes.NewBuffer(mem.NoAlloc, arr[:0])

	w := bufio.NewWriter(alloc, &buf)
	if got := w.Size(); got != bufio.DefaultBufSize {
		t.Errorf("NewWriter: Size() = %d, want %d", got, bufio.DefaultBufSize)
	}
	w.Free()

	w = bufio.NewWriterSize(alloc, &buf, 1234)
	if got := w.Size(); got != 1234 {
		t.Errorf("NewWriterSize: Size() = %d, want 1234", got)
	}
	w.Free()
}

func TestWriterReadFromMustReturnUnderlyingError(t *testing.T) {
	var ew errorOnlyWriter
	wr := bufio.NewWriter(t.Allocator(), &ew)
	defer wr.Free()

	const s = "test1"
	if _, err := wr.WriteString(s); err != nil {
		t.Errorf("WriteString() = %s, want nil", errName(errCode(err)))
		return
	}
	if err := wr.Flush(); err == nil {
		t.Error("Flush() = nil, want an error")
	}
	sr := strings.NewReader("test2")
	if _, err := wr.ReadFrom(&sr); err == nil {
		t.Error("ReadFrom() = nil, want an error")
	}
	if buffered := wr.Buffered(); buffered != len(s) {
		t.Errorf("Buffered() = %d, want %d", buffered, len(s))
	}
}
