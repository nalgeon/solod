// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"unsafe"

	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// checkBuilder reports whether the builder holds want. Len must give the length
// of the string, and Cap must not be below it.
func checkBuilder(t *testing.T, b *strings.Builder, want string, step string) {
	got := b.String()
	if got != want {
		t.Errorf("%s: String() = %s, want %s", step, got, want)
		return
	}
	if n := b.Len(); n != len(got) {
		t.Errorf("%s: Len() = %d, want %d", step, n, len(got))
	}
	if n := b.Cap(); n < len(got) {
		t.Errorf("%s: Cap() = %d, below the length %d", step, n, len(got))
	}
}

func TestBuilder(t *testing.T) {
	// The zero builder is ready to use.
	var b strings.Builder
	defer b.Free()
	checkBuilder(t, &b, "", "zero")

	n, err := b.WriteString("hello")
	if err != nil || n != 5 {
		t.Errorf("WriteString() = %d, want 5", n)
	}
	checkBuilder(t, &b, "hello", "after WriteString")

	if err := b.WriteByte(' '); err != nil {
		t.Error("WriteByte() failed")
	}
	checkBuilder(t, &b, "hello ", "after WriteByte")

	n, err = b.WriteString("world")
	if err != nil || n != 5 {
		t.Errorf("WriteString() = %d, want 5", n)
	}
	checkBuilder(t, &b, "hello world", "after the second WriteString")
}

func TestBuilderAllocator(t *testing.T) {
	// A builder of NewBuilder takes the memory from the allocator, and Free
	// gives every byte back.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	for range 100 {
		b.WriteString("abc")
	}
	if b.Len() != 300 {
		t.Errorf("Len() = %d, want 300", b.Len())
	}
	b.Free()
}

// The write methods of the builder, by number.
const (
	writeBytes = iota
	writeRune
	writeRuneWide
	writeString
)

// The string of the write cases. It holds a code point of one byte and a code
// point of three bytes.
const writeInput = "hello 世界"

// writeOnce calls the write method with the number and returns the number of
// written bytes.
func writeOnce(b *strings.Builder, kind int) (int, error) {
	switch kind {
	case writeBytes:
		return b.Write([]byte(writeInput))
	case writeRune:
		return b.WriteRune('a')
	case writeRuneWide:
		return b.WriteRune('世')
	}
	return b.WriteString(writeInput)
}

// writeWant returns the string that the write method with the number appends.
func writeWant(kind int) string {
	switch kind {
	case writeBytes:
		return writeInput
	case writeRune:
		return "a"
	case writeRuneWide:
		return "世"
	}
	return writeInput
}

func TestBuilderWrite(t *testing.T) {
	// Every write method appends to the builder and returns the number of
	// written bytes.
	alloc := t.Allocator()
	for kind := writeBytes; kind <= writeString; kind++ {
		b := strings.NewBuilder(alloc)
		want := writeWant(kind)

		n, err := writeOnce(&b, kind)
		if err != nil {
			t.Errorf("kind %d: the first write failed", kind)
			b.Free()
			continue
		}
		if n != len(want) {
			t.Errorf("kind %d: the first write = %d bytes, want %d", kind, n, len(want))
		}
		checkBuilder(t, &b, want, "after the first write")

		n, err = writeOnce(&b, kind)
		if err != nil {
			t.Errorf("kind %d: the second write failed", kind)
			b.Free()
			continue
		}
		if n != len(want) {
			t.Errorf("kind %d: the second write = %d bytes, want %d", kind, n, len(want))
		}
		checkBuilder(t, &b, want+want, "after the second write")
		b.Free()
	}
}

func TestBuilderWriteByte(t *testing.T) {
	// WriteByte writes a NUL byte like every other byte.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	defer b.Free()
	if err := b.WriteByte('a'); err != nil {
		t.Error("WriteByte(a) failed")
	}
	if err := b.WriteByte(0); err != nil {
		t.Error("WriteByte(0) failed")
	}
	checkBuilder(t, &b, "a\x00", "after WriteByte")
}

func TestBuilderWriteInvalidRune(t *testing.T) {
	// An invalid code point becomes RuneError.
	alloc := t.Allocator()
	invalid := []rune{-1, utf8.MaxRune + 1, 0xD800}
	for _, r := range invalid {
		b := strings.NewBuilder(alloc)
		n, _ := b.WriteRune(r)
		if n != 3 {
			t.Errorf("WriteRune(%d) = %d bytes, want 3", r, n)
		}
		checkBuilder(t, &b, "\ufffd", "after WriteRune")
		b.Free()
	}
}

func TestBuilderReset(t *testing.T) {
	// Reset empties the builder and keeps the buffer.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	defer b.Free()
	b.WriteString("aaa")
	checkBuilder(t, &b, "aaa", "before Reset")

	c := b.Cap()
	b.Reset()
	checkBuilder(t, &b, "", "after Reset")
	if b.Cap() != c {
		t.Errorf("Cap() after Reset = %d, want %d", b.Cap(), c)
	}

	b.WriteString("bbb")
	checkBuilder(t, &b, "bbb", "after the write that follows Reset")
}

func TestBuilderFree(t *testing.T) {
	// Free empties the builder, and the builder works again after it.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	b.WriteString("aaa")
	b.Free()
	checkBuilder(t, &b, "", "after Free")
	if b.Cap() != 0 {
		t.Errorf("Cap() after Free = %d, want 0", b.Cap())
	}
	b.WriteString("bbb")
	checkBuilder(t, &b, "bbb", "after the write that follows Free")
	b.Free()
}

func TestBuilderGrow(t *testing.T) {
	// Grow gives space for n more bytes without a change of the string.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	defer b.Free()
	b.WriteString("abc")
	b.Grow(1000)
	if b.Cap() < 1003 {
		t.Errorf("Cap() after Grow(1000) = %d, want 1003 or more", b.Cap())
	}
	checkBuilder(t, &b, "abc", "after Grow")

	// Grow(0) changes nothing.
	c := b.Cap()
	b.Grow(0)
	if b.Cap() != c {
		t.Errorf("Cap() after Grow(0) = %d, want %d", b.Cap(), c)
	}
}

func TestBuilderString(t *testing.T) {
	// String gives a view of the buffer, not a copy.
	alloc := t.Allocator()
	b := strings.NewBuilder(alloc)
	defer b.Free()
	b.WriteString("alpha")
	s1 := b.String()
	s2 := b.String()
	if unsafe.StringData(s1) != unsafe.StringData(s2) {
		t.Error("String() copied the bytes")
	}
}

func TestBuilderFixed(t *testing.T) {
	// A fixed builder writes into the buffer of the caller and allocates
	// nothing.
	var buf [16]byte
	b := strings.FixedBuilder(buf[:])
	if b.Cap() != 16 {
		t.Errorf("Cap() = %d, want 16", b.Cap())
	}
	b.WriteString("hello")
	b.WriteByte(' ')
	b.WriteRune('世')
	checkBuilder(t, &b, "hello 世", "after the writes")
	if unsafe.StringData(b.String()) != &buf[0] {
		t.Error("the fixed builder does not write into the buffer of the caller")
	}

	// Reset empties the builder and keeps the buffer of the caller.
	b.Reset()
	if b.Len() != 0 || b.Cap() != 16 {
		t.Errorf("after Reset: Len() = %d, Cap() = %d, want 0 and 16", b.Len(), b.Cap())
	}

	// Free drops the buffer of the caller and frees nothing.
	b.Free()
	if b.Cap() != 0 {
		t.Errorf("Cap() after Free = %d, want 0", b.Cap())
	}
}
