// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex_test

import (
	"solod.dev/so/encoding/hex"
	"solod.dev/so/errors"
	"solod.dev/so/io"
	"solod.dev/so/strings"
)

// The error codes of the tables.
const (
	errNone = iota
	errLength
	errInvalidByte
	errDumperClosed
	errEOF
	errUnexpectedEOF
	errWrite
	errOther
)

// errWriteFailed is the error of the writers that fail on purpose.
var errWriteFailed = errors.New("hex_test: write failed")

// errCode returns the code of the error.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	switch err {
	case hex.ErrLength:
		return errLength
	case hex.ErrInvalidByte:
		return errInvalidByte
	case hex.ErrDumperClosed:
		return errDumperClosed
	case io.EOF:
		return errEOF
	case io.ErrUnexpectedEOF:
		return errUnexpectedEOF
	case errWriteFailed:
		return errWrite
	}
	return errOther
}

// errName returns the name of the error with the code.
func errName(code int) string {
	switch code {
	case errNone:
		return "nil"
	case errLength:
		return "ErrLength"
	case errInvalidByte:
		return "ErrInvalidByte"
	case errDumperClosed:
		return "ErrDumperClosed"
	case errEOF:
		return "EOF"
	case errUnexpectedEOF:
		return "ErrUnexpectedEOF"
	case errWrite:
		return "errWriteFailed"
	}
	return "other"
}

// encDecTest is a decoded value and the hexadecimal encoding of the value.
type encDecTest struct {
	enc string
	dec string
}

var encDecTests = [...]encDecTest{
	{"", ""},
	{"0001020304050607", "\x00\x01\x02\x03\x04\x05\x06\x07"},
	{"08090a0b0c0d0e0f", "\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f"},
	{"f0f1f2f3f4f5f6f7", "\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7"},
	{"f8f9fafbfcfdfeff", "\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"},
	{"67", "g"},
	{"e3a1", "\xe3\xa1"},
}

// maxDec is the length of the longest decoded value of encDecTests.
const maxDec = 8

// errTest is a decode input, the decoded prefix, and the error code.
type errTest struct {
	in  string
	out string
	err int
}

var errTests = [...]errTest{
	{"", "", errNone},
	{"0", "", errLength},
	{"zd4aa", "", errInvalidByte},
	{"d4aaz", "\xd4\xaa", errInvalidByte},
	{"30313", "01", errLength},
	{"0g", "", errInvalidByte},
	{"00gg", "\x00", errInvalidByte},
	{"0\x01", "", errInvalidByte},
	{"ffeed", "\xff\xee", errLength},
}

// hexVal returns the value of a hexadecimal character, and -1 for any other
// character.
func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// hexDigit returns the lowercase hexadecimal character of a value below 16.
func hexDigit(v byte) byte {
	if v < 10 {
		return '0' + v
	}
	return 'a' + v - 10
}

// encodeBrute encodes src into dst the simple way. It returns the number of
// the written bytes.
func encodeBrute(dst, src []byte) int {
	j := 0
	for _, v := range src {
		dst[j] = hexDigit(v >> 4)
		dst[j+1] = hexDigit(v & 0x0f)
		j += 2
	}
	return j
}

// decodeBrute decodes src into dst the simple way. It returns the number of
// the written bytes and the error code.
func decodeBrute(dst []byte, src string) (int, int) {
	n := 0
	for i := 0; i+1 < len(src); i += 2 {
		hi := hexVal(src[i])
		lo := hexVal(src[i+1])
		if hi < 0 || lo < 0 {
			return n, errInvalidByte
		}
		dst[n] = byte(hi<<4 | lo)
		n++
	}
	if len(src)%2 == 1 {
		// An invalid character is an earlier problem than the odd length.
		if hexVal(src[len(src)-1]) < 0 {
			return n, errInvalidByte
		}
		return n, errLength
	}
	return n, errNone
}

// dumpChar returns the character a dump prints for a byte.
func dumpChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

// dumpBrute writes the hexadecimal dump of data into b the simple way.
// The format matches the output of `hexdump -C`.
func dumpBrute(b *strings.Builder, data []byte) {
	for i := 0; i < len(data); i += 16 {
		end := min(i+16, len(data))
		line := data[i:end]

		// The offset of the line, in 8 hexadecimal digits.
		for shift := 28; shift >= 0; shift -= 4 {
			b.WriteByte(hexDigit(byte(i>>uint(shift)) & 0x0f))
		}
		b.WriteString("  ")

		// The 16 byte slots. An absent byte gives two spaces.
		for j := range 16 {
			if j < len(line) {
				b.WriteByte(hexDigit(line[j] >> 4))
				b.WriteByte(hexDigit(line[j] & 0x0f))
			} else {
				b.WriteString("  ")
			}
			b.WriteByte(' ')
			if j == 7 {
				b.WriteByte(' ')
			}
			if j == 15 {
				b.WriteString(" |")
			}
		}

		// The ASCII column.
		for j := range len(line) {
			b.WriteByte(dumpChar(line[j]))
		}
		b.WriteString("|\n")
	}
}

// The alphabet and the longest word of the decode sweep. The alphabet holds a
// digit, both cases of a hexadecimal letter, the last hexadecimal letter, a
// space, a letter above the hexadecimal range, and the two edge bytes.
const (
	decAlpha   = "0aA fz\x00\xff"
	maxDecWord = 4
)

// wordCount returns the number of words of alpha with the length n.
func wordCount(alpha string, n int) int {
	count := 1
	for range n {
		count *= len(alpha)
	}
	return count
}

// wordTotal returns the number of words of alpha with a length up to max.
func wordTotal(alpha string, max int) int {
	total := 0
	for n := 0; n <= max; n++ {
		total += wordCount(alpha, n)
	}
	return total
}

// wordAt writes the word number i of alpha into buf and returns the result.
// The shorter words come first, and every word with a length up to max appears
// once. The caller must keep i below wordTotal(alpha, max).
func wordAt(buf []byte, alpha string, max, i int) string {
	for n := 0; n <= max; n++ {
		count := wordCount(alpha, n)
		if i < count {
			for k := 0; k < n; k++ {
				buf[k] = alpha[i%len(alpha)]
				i /= len(alpha)
			}
			return string(buf[:n])
		}
		i -= count
	}
	return ""
}

// errWriter fails every write.
type errWriter struct{}

func (*errWriter) Write(p []byte) (int, error) {
	_ = p
	return 0, errWriteFailed
}

// lateErrWriter accepts the first left writes and fails after them.
type lateErrWriter struct {
	left int
}

func (w *lateErrWriter) Write(p []byte) (int, error) {
	if w.left == 0 {
		return 0, errWriteFailed
	}
	w.left--
	return len(p), nil
}

// oneByteReader gives one byte at each read.
type oneByteReader struct {
	s string
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if len(r.s) == 0 {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.s[0]
	r.s = r.s[1:]
	return 1, nil
}
