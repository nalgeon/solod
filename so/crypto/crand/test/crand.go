// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crand_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/crypto/crand"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

// readSize is the number of bytes that a distribution test reads.
const readSize = 100000

// alphabet is the RFC 4648 base32 alphabet that Text uses.
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// textLen is the number of characters that Text writes.
const textLen = 26

// readInto fills b through Read or through Reader.Read. The two entry points give
// the same guarantees, so every read test checks both.
func readInto(useReader bool, b []byte) (int, error) {
	if useReader {
		return crand.Reader.Read(b)
	}
	return crand.Read(b)
}

// readName returns the name of the entry point that readInto uses.
func readName(useReader bool) string {
	if useReader {
		return "Reader.Read"
	}
	return "Read"
}

func TestRead(t *testing.T) {
	checkRead(t, false)
}

func TestReaderRead(t *testing.T) {
	checkRead(t, true)
}

// checkRead reads a large block and checks the distribution of the bytes.
func checkRead(t *testing.T, useReader bool) {
	name := readName(useReader)
	alloc := t.Allocator()
	b := mem.AllocSlice[byte](alloc, readSize, readSize)
	defer mem.FreeSlice(alloc, b)

	n, err := readInto(useReader, b)
	if err != nil {
		t.Errorf("%s(buf) failed", name)
		return
	}
	if n != len(b) {
		t.Errorf("%s(buf) = %d, want %d", name, n, len(b))
		return
	}
	checkUniform(t, name, b)
}

// chiLo and chiHi bound the chi-square statistic of a histogram of 256 buckets.
// The statistic has 255 degrees of freedom, so a uniform stream gives a value
// near 255 with a standard deviation of 22.6. The bounds are more than five
// standard deviations away, so a uniform stream almost never fails them. A
// stream with a pattern gives a value in the thousands.
const chiLo = 128.0
const chiHi = 448.0

// checkUniform checks that b holds a uniform stream of bytes. Go compresses the
// stream with flate and rejects a stream that gets smaller. So has no
// compression package, so the check is a chi-square test over the byte values
// and over the differences of the neighboring bytes. A constant stream fails
// the first test, and a counter fails the second one.
func checkUniform(t *testing.T, name string, b []byte) {
	var values, diffs [256]int
	for i, v := range b {
		values[v]++
		if i > 0 {
			diffs[v-b[i-1]]++
		}
	}
	for v := range 256 {
		if values[v] == 0 {
			t.Errorf("%s: byte %d is absent from %d bytes", name, v, len(b))
			return
		}
	}
	checkChiSquare(t, name, "the values", values[:], len(b))
	checkChiSquare(t, name, "the differences", diffs[:], len(b)-1)
}

// checkChiSquare checks the chi-square statistic of a histogram of 256 buckets
// over n values.
func checkChiSquare(t *testing.T, name string, what string, hist []int, n int) {
	exp := float64(n) / 256
	sum := 0.0
	for _, count := range hist {
		d := float64(count) - exp
		sum += d * d / exp
	}
	if sum < chiLo || sum > chiHi {
		t.Errorf("%s: chi-square of %s = %g, want %g to %g", name, what, sum, chiLo, chiHi)
	}
}

// byteValueReads bounds the number of one byte reads that a byte value test
// takes. The expected number is 1568, so the bound is far above it.
const byteValueReads = 20000

func TestReadByteValues(t *testing.T) {
	checkReadByteValues(t, false)
}

func TestReaderReadByteValues(t *testing.T) {
	checkReadByteValues(t, true)
}

// checkReadByteValues checks that one byte reads give every byte value.
func checkReadByteValues(t *testing.T, useReader bool) {
	name := readName(useReader)
	var seen [256]bool
	count := 0
	b := make([]byte, 1)
	for range byteValueReads {
		n, err := readInto(useReader, b)
		if err != nil || n != 1 {
			t.Errorf("%s(b) = %d, want 1", name, n)
			return
		}
		if seen[b[0]] {
			continue
		}
		seen[b[0]] = true
		count++
		if count == 256 {
			return
		}
	}
	t.Errorf("%s: %d byte values after %d reads, want 256", name, count, byteValueReads)
}

func TestReadEmpty(t *testing.T) {
	checkReadEmpty(t, false)
}

func TestReaderReadEmpty(t *testing.T) {
	checkReadEmpty(t, true)
}

// checkReadEmpty checks the result of a read into an empty buffer. So gives one
// representation to a nil slice and to an empty slice, so one check covers both.
func checkReadEmpty(t *testing.T, useReader bool) {
	name := readName(useReader)
	var b []byte
	n, err := readInto(useReader, b)
	if err != nil {
		t.Errorf("%s(nil) failed", name)
		return
	}
	if n != 0 {
		t.Errorf("%s(nil) = %d, want 0", name, n)
	}
}

// readSizes are the buffer sizes that TestReadSizes reads. The wasm path of
// crand_read reads in blocks of 256 bytes, so the sizes cross that boundary.
var readSizes = []int{1, 2, 16, 31, 255, 256, 257, 512, 1000}

// guard is the number of bytes that TestReadSizes keeps past the buffer end.
const guard = 16

// fill is the byte that TestReadSizes writes before every read.
const fill = 0xAA

func TestReadSizes(t *testing.T) {
	const bufSize = 1000 + guard
	alloc := t.Allocator()
	buf := mem.AllocSlice[byte](alloc, bufSize, bufSize)
	defer mem.FreeSlice(alloc, buf)

	for _, size := range readSizes {
		for i := range buf {
			buf[i] = fill
		}
		n, err := crand.Read(buf[:size])
		if err != nil || n != size {
			t.Errorf("Read(buf[:%d]) = %d, want %d", size, n, size)
			return
		}
		if !allEqual(buf[size:], fill) {
			t.Errorf("Read(buf[:%d]) wrote past the end of the buffer", size)
			return
		}
		// A read of 16 bytes or more keeps every byte of the fill with a
		// probability of 2^-128, so an unchanged buffer means a failed read.
		if size >= 16 && allEqual(buf[:size], fill) {
			t.Errorf("Read(buf[:%d]) wrote no random byte", size)
			return
		}
	}
}

// allEqual reports whether every byte of b is v.
func allEqual(b []byte, v byte) bool {
	for _, c := range b {
		if c != v {
			return false
		}
	}
	return true
}

func TestReaderReadFull(t *testing.T) {
	buf := make([]byte, 64)
	n, err := io.ReadFull(crand.Reader, buf)
	if err != nil {
		t.Fatal("ReadFull failed")
		return
	}
	if n != len(buf) {
		t.Errorf("ReadFull() = %d, want %d", n, len(buf))
	}
}

// textRounds bounds the number of Text calls that TestText takes. The chance to
// miss a character in a position after 1000 rounds is (31/32)^1000 = 1.6e-14,
// so the test reaches every character in every position long before the bound.
const textRounds = 1000

func TestText(t *testing.T) {
	alloc := t.Allocator()
	buf := mem.AllocSlice[byte](alloc, textLen, textLen)
	defer mem.FreeSlice(alloc, buf)
	// seen holds the results of the earlier rounds, one after another.
	seen := mem.AllocSlice[byte](alloc, textRounds*textLen, textRounds*textLen)
	defer mem.FreeSlice(alloc, seen)

	// hasChar[i][j] records the alphabet character j at the position i, and
	// distinct[i] counts the characters that the position i has reached.
	var hasChar [textLen][32]bool
	var distinct [textLen]int

	done := false
	for round := 0; round < textRounds && !done; round++ {
		s := crand.Text(buf)
		if len(s) != textLen {
			t.Errorf("len(Text()) = %d, want %d", len(s), textLen)
			return
		}
		done = true
		for i := range textLen {
			j := alphabetIndex(s[i])
			if j < 0 {
				t.Errorf("Text()[%d] = %c, outside of the base32 alphabet", i, s[i])
				return
			}
			if !hasChar[i][j] {
				hasChar[i][j] = true
				distinct[i]++
			}
			if distinct[i] != 32 {
				done = false
			}
		}
		// Every result carries 128 bits of randomness, so a repeated result
		// means a broken generator.
		prev := seen[:round*textLen]
		for k := 0; k < len(prev); k += textLen {
			if bytes.Equal(prev[k:k+textLen], buf) {
				t.Errorf("Text() = %s, a duplicate of an earlier result", s)
				return
			}
		}
		copy(seen[round*textLen:], buf)
	}

	if done {
		return
	}
	t.Errorf("Text() missed a character in a position after %d rounds", textRounds)
	for i := range textLen {
		if distinct[i] != 32 {
			t.Errorf("position %d has %d of 32 characters", i, distinct[i])
		}
	}
}

// alphabetIndex returns the position of c in the base32 alphabet. It returns -1
// if the alphabet does not hold c.
func alphabetIndex(c byte) int {
	for i := range len(alphabet) {
		if alphabet[i] == c {
			return i
		}
	}
	return -1
}
