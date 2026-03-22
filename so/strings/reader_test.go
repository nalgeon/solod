// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"fmt"
	"sync"
	"testing"

	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/strings"
)

func TestReader(t *testing.T) {
	r := strings.NewReader("0123456789")
	tests := []struct {
		off     int64
		seek    int
		n       int
		want    string
		wantpos int64
		readerr error
		seekerr string
	}{
		{seek: io.SeekStart, off: 0, n: 20, want: "0123456789"},
		{seek: io.SeekStart, off: 1, n: 1, want: "1"},
		{seek: io.SeekCurrent, off: 1, wantpos: 3, n: 2, want: "34"},
		{seek: io.SeekStart, off: -1, seekerr: io.ErrOffset.Error()},
		{seek: io.SeekStart, off: 1 << 33, wantpos: 1 << 33, readerr: io.EOF},
		{seek: io.SeekCurrent, off: 1, wantpos: 1<<33 + 1, readerr: io.EOF},
		{seek: io.SeekStart, n: 5, want: "01234"},
		{seek: io.SeekCurrent, n: 5, want: "56789"},
		{seek: io.SeekEnd, off: -1, n: 1, wantpos: 9, want: "9"},
	}

	for i, tt := range tests {
		pos, err := r.Seek(tt.off, tt.seek)
		if err == nil && tt.seekerr != "" {
			t.Errorf("%d. want seek error %q", i, tt.seekerr)
			continue
		}
		if err != nil && err.Error() != tt.seekerr {
			t.Errorf("%d. seek error = %q; want %q", i, err.Error(), tt.seekerr)
			continue
		}
		if tt.wantpos != 0 && tt.wantpos != pos {
			t.Errorf("%d. pos = %d, want %d", i, pos, tt.wantpos)
		}
		buf := make([]byte, tt.n)
		n, err := r.Read(buf)
		if err != tt.readerr {
			t.Errorf("%d. read = %v; want %v", i, err, tt.readerr)
			continue
		}
		got := string(buf[:n])
		if got != tt.want {
			t.Errorf("%d. got %q; want %q", i, got, tt.want)
		}
	}
}

func TestReadAfterBigSeek(t *testing.T) {
	r := strings.NewReader("0123456789")
	if _, err := r.Seek(1<<31+5, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if n, err := r.Read(make([]byte, 10)); n != 0 || err != io.EOF {
		t.Errorf("Read = %d, %v; want 0, EOF", n, err)
	}
}

func TestReaderAt(t *testing.T) {
	r := strings.NewReader("0123456789")
	tests := []struct {
		off     int64
		n       int
		want    string
		wanterr any
	}{
		{0, 10, "0123456789", nil},
		{1, 10, "123456789", io.EOF},
		{1, 9, "123456789", nil},
		{11, 10, "", io.EOF},
		{0, 0, "", nil},
		{-1, 0, "", io.ErrOffset.Error()},
	}
	for i, tt := range tests {
		b := make([]byte, tt.n)
		rn, err := r.ReadAt(b, tt.off)
		got := string(b[:rn])
		if got != tt.want {
			t.Errorf("%d. got %q; want %q", i, got, tt.want)
		}
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.wanterr) {
			t.Errorf("%d. got error = %v; want %v", i, err, tt.wanterr)
		}
	}
}

func TestReaderAtConcurrent(t *testing.T) {
	// Test for the race detector, to verify ReadAt doesn't mutate
	// any state.
	r := strings.NewReader("0123456789")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var buf [1]byte
			r.ReadAt(buf[:], int64(i))
		}(i)
	}
	wg.Wait()
}

func TestEmptyReaderConcurrent(t *testing.T) {
	// Test for the race detector, to verify a Read that doesn't yield any bytes
	// is okay to use from multiple goroutines. This was our historic behavior.
	// See golang.org/issue/7856
	r := strings.NewReader("")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			var buf [1]byte
			r.Read(buf[:])
		}()
		go func() {
			defer wg.Done()
			r.Read(nil)
		}()
	}
	wg.Wait()
}

func TestWriteTo(t *testing.T) {
	const str = "0123456789"
	for i := 0; i <= len(str); i++ {
		s := str[i:]
		r := strings.NewReader(s)
		var b bytes.Buffer
		n, err := r.WriteTo(&b)
		if expect := int64(len(s)); n != expect {
			t.Errorf("got %v; want %v", n, expect)
		}
		if err != nil {
			t.Errorf("for length %d: got error = %v; want nil", len(s), err)
		}
		if b.String() != s {
			t.Errorf("got string %q; want %q", b.String(), s)
		}
		if r.Len() != 0 {
			t.Errorf("reader contains %v bytes; want 0", r.Len())
		}
	}
}

// tests that Len is affected by reads, but Size is not.
func TestReaderLenSize(t *testing.T) {
	r := strings.NewReader("abc")
	io.CopyN(io.Discard, &r, 1)
	if r.Len() != 2 {
		t.Errorf("Len = %d; want 2", r.Len())
	}
	if r.Size() != 3 {
		t.Errorf("Size = %d; want 3", r.Size())
	}
}

func TestReaderReset(t *testing.T) {
	r := strings.NewReader("世界")
	if res := r.ReadRune(); res.Err != nil {
		t.Errorf("ReadRune: unexpected error: %v", res.Err)
	}

	const want = "abcdef"
	r.Reset(want)
	if err := r.UnreadRune(); err == nil {
		t.Errorf("UnreadRune: expected error, got nil")
	}
	buf, err := io.ReadAll(nil, &r)
	if err != nil {
		t.Errorf("ReadAll: unexpected error: %v", err)
	}
	if got := string(buf); got != want {
		t.Errorf("ReadAll: got %q, want %q", got, want)
	}
}

func TestReaderZero(t *testing.T) {
	if l := (&strings.Reader{}).Len(); l != 0 {
		t.Errorf("Len: got %d, want 0", l)
	}

	if n, err := (&strings.Reader{}).Read(nil); n != 0 || err != io.EOF {
		t.Errorf("Read: got %d, %v; want 0, io.EOF", n, err)
	}

	if n, err := (&strings.Reader{}).ReadAt(nil, 11); n != 0 || err != io.EOF {
		t.Errorf("ReadAt: got %d, %v; want 0, io.EOF", n, err)
	}

	if b, err := (&strings.Reader{}).ReadByte(); b != 0 || err != io.EOF {
		t.Errorf("ReadByte: got %d, %v; want 0, io.EOF", b, err)
	}

	if res := (&strings.Reader{}).ReadRune(); res.Rune != 0 || res.Size != 0 || res.Err != io.EOF {
		t.Errorf("ReadRune: got %d, %d, %v; want 0, 0, io.EOF", res.Rune, res.Size, res.Err)
	}

	if offset, err := (&strings.Reader{}).Seek(11, io.SeekStart); offset != 11 || err != nil {
		t.Errorf("Seek: got %d, %v; want 11, nil", offset, err)
	}

	if s := (&strings.Reader{}).Size(); s != 0 {
		t.Errorf("Size: got %d, want 0", s)
	}

	if (&strings.Reader{}).UnreadByte() == nil {
		t.Errorf("UnreadByte: got nil, want error")
	}

	if (&strings.Reader{}).UnreadRune() == nil {
		t.Errorf("UnreadRune: got nil, want error")
	}

	if n, err := (&strings.Reader{}).WriteTo(io.Discard); n != 0 || err != nil {
		t.Errorf("WriteTo: got %d, %v; want 0, nil", n, err)
	}
}

func TestReadByte(t *testing.T) {
	testStrings := []string{"", abcd, faces, commas}
	for _, s := range testStrings {
		reader := strings.NewReader(s)
		if e := reader.UnreadByte(); e == nil {
			t.Errorf("Unreading %q at beginning: expected error", s)
		}
		var res bytes.Buffer
		for {
			b, e := reader.ReadByte()
			if e == io.EOF {
				break
			}
			if e != nil {
				t.Errorf("Reading %q: %s", s, e)
				break
			}
			res.WriteByte(b)
			// unread and read again
			e = reader.UnreadByte()
			if e != nil {
				t.Errorf("Unreading %q: %s", s, e)
				break
			}
			b1, e := reader.ReadByte()
			if e != nil {
				t.Errorf("Reading %q after unreading: %s", s, e)
				break
			}
			if b1 != b {
				t.Errorf("Reading %q after unreading: want byte %q, got %q", s, b, b1)
				break
			}
		}
		if res.String() != s {
			t.Errorf("Reader(%q).ReadByte() produced %q", s, res.String())
		}
	}
}

func TestReadRune(t *testing.T) {
	testStrings := []string{"", abcd, faces, commas}
	for _, s := range testStrings {
		reader := strings.NewReader(s)
		if e := reader.UnreadRune(); e == nil {
			t.Errorf("Unreading %q at beginning: expected error", s)
		}
		res := ""
		for {
			rr := reader.ReadRune()
			if rr.Err == io.EOF {
				break
			}
			if rr.Err != nil {
				t.Errorf("Reading %q: %s", s, rr.Err)
				break
			}
			res += string(rr.Rune)
			// unread and read again
			e := reader.UnreadRune()
			if e != nil {
				t.Errorf("Unreading %q: %s", s, e)
				break
			}
			rr1 := reader.ReadRune()
			if rr1.Err != nil {
				t.Errorf("Reading %q after unreading: %s", s, rr1.Err)
				break
			}
			if rr1.Rune != rr.Rune {
				t.Errorf("Reading %q after unreading: want rune %q, got %q", s, rr.Rune, rr1.Rune)
				break
			}
			if rr1.Size != rr.Size {
				t.Errorf("Reading %q after unreading: want size %d, got %d", s, rr.Size, rr1.Size)
				break
			}
		}
		if res != s {
			t.Errorf("Reader(%q).ReadRune() produced %q", s, res)
		}
	}
}

func TestUnreadRuneError(t *testing.T) {
	var tests = []struct {
		name string
		f    func(*strings.Reader)
	}{
		{"Read", func(r *strings.Reader) { r.Read([]byte{0}) }},
		{"ReadByte", func(r *strings.Reader) { r.ReadByte() }},
		{"UnreadRune", func(r *strings.Reader) { r.UnreadRune() }},
		{"Seek", func(r *strings.Reader) { r.Seek(0, io.SeekCurrent) }},
		{"WriteTo", func(r *strings.Reader) { r.WriteTo(&bytes.Buffer{}) }},
	}

	for _, tt := range tests {
		reader := strings.NewReader("0123456789")
		if res := reader.ReadRune(); res.Err != nil {
			// should not happen
			t.Fatal(res.Err)
		}
		tt.f(&reader)
		err := reader.UnreadRune()
		if err == nil {
			t.Errorf("Unreading after %s: expected error", tt.name)
		}
	}
}
