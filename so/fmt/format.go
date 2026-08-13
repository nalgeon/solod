// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ported from Go's fmt/format.go.
//
// Dropped against Go: fmtUnicode for %U, fmtQ and fmtQc for %q, and the
// formatters for byte slices and hexadecimal strings. A verb picks the type
// of its argument, so no verb can reach them.

package fmt

import (
	"solod.dev/so/strconv"
	"solod.dev/so/unicode/utf8"
)

const (
	ldigits = "0123456789abcdefx"
	udigits = "0123456789ABCDEFX"
)

// C reserves the words "signed" and "unsigned", so the constants that select
// the integer kind carry a suffix.
const (
	signedInt   = true
	unsignedInt = false
)

// maxFloatIntDigits is the number of digits before the decimal point in %f of
// the largest float64.
const maxFloatIntDigits = 309

// fmtFlags holds the flags of one verb, in a struct of its own for easy
// clearing.
type fmtFlags struct {
	widPresent  bool
	precPresent bool
	minus       bool
	plus        bool
	sharp       bool
	space       bool
	zero        bool
}

// formatter writes one value to a buffer, with the flags, the width and the
// precision of its verb.
type formatter struct {
	buf *buffer

	flags fmtFlags

	wid  int // width
	prec int // precision

	// intbuf is large enough to store %b of an int64 with a sign.
	intbuf [68]byte
}

// init prepares the formatter to write to buf.
func (f *formatter) init(buf *buffer) {
	f.buf = buf
	f.clearflags()
}

// clearflags resets the flags, the width and the precision.
func (f *formatter) clearflags() {
	f.flags = fmtFlags{}
	f.wid = 0
	f.prec = 0
}

// writePadding writes n padding bytes.
func (f *formatter) writePadding(n int) {
	if n <= 0 {
		return
	}
	// Zero padding is allowed only to the left.
	padByte := byte(' ')
	if f.flags.zero && !f.flags.minus {
		padByte = byte('0')
	}
	f.buf.writeRepeat(padByte, n)
}

// pad writes b, padded on the left (!minus) or on the right (minus).
func (f *formatter) pad(b []byte) {
	if !f.flags.widPresent || f.wid == 0 {
		f.buf.write(b)
		return
	}
	width := f.wid - utf8.RuneCount(b)
	if !f.flags.minus {
		f.writePadding(width)
		f.buf.write(b)
		return
	}
	f.buf.write(b)
	f.writePadding(width)
}

// padString writes s, padded on the left (!minus) or on the right (minus).
func (f *formatter) padString(s string) {
	if !f.flags.widPresent || f.wid == 0 {
		f.buf.writeString(s)
		return
	}
	width := f.wid - utf8.RuneCountInString(s)
	if !f.flags.minus {
		f.writePadding(width)
		f.buf.writeString(s)
		return
	}
	f.buf.writeString(s)
	f.writePadding(width)
}

// fmtBoolean formats a boolean.
func (f *formatter) fmtBoolean(v bool) {
	if v {
		f.padString("true")
	} else {
		f.padString("false")
	}
}

// fmtInteger formats a signed or an unsigned integer.
func (f *formatter) fmtInteger(u uint64, base int, isSigned bool, verb rune, digits string) {
	negative := isSigned && int64(u) < 0
	if negative {
		u = -u
	}

	// The 68 bytes of intbuf are enough for an integer with no width and no
	// precision.
	buf := f.intbuf[0:]
	if f.flags.widPresent || f.flags.precPresent {
		// Account 3 extra bytes for a sign and "0x".
		width := 3 + f.wid + f.prec // wid and prec are always positive
		if width > len(buf) {
			buf = make([]byte, width)
		}
	}

	// Two ways to ask for extra leading zero digits: %.3d or %03d.
	// If the format has both, the zero flag is ignored and the value gets
	// space padding.
	prec := 0
	if f.flags.precPresent {
		prec = f.prec
		// A precision of 0 and a value of 0 mean "print the padding only".
		if prec == 0 && u == 0 {
			oldZero := f.flags.zero
			f.flags.zero = false
			f.writePadding(f.wid)
			f.flags.zero = oldZero
			return
		}
	} else if f.flags.zero && !f.flags.minus && f.flags.widPresent {
		// Zero padding is allowed only to the left.
		prec = f.wid
		if negative || f.flags.plus || f.flags.space {
			prec-- // leave room for the sign
		}
	}

	// Format u into buf, ending at buf[i]. Right to left is easier.
	i := len(buf)
	switch base {
	case 10:
		for u >= 10 {
			i--
			next := u / 10
			buf[i] = byte('0' + u - next*10)
			u = next
		}
	case 16:
		for u >= 16 {
			i--
			buf[i] = digits[u&0xF]
			u >>= 4
		}
	case 8:
		for u >= 8 {
			i--
			buf[i] = byte('0' + u&7)
			u >>= 3
		}
	case 2:
		for u >= 2 {
			i--
			buf[i] = byte('0' + u&1)
			u >>= 1
		}
	default:
		panic("fmt: unknown base; can't happen")
	}
	i--
	buf[i] = digits[u]
	for i > 0 && prec > len(buf)-i {
		i--
		buf[i] = '0'
	}

	// Prefixes: 0x, -, and so on.
	if f.flags.sharp {
		switch base {
		case 2:
			i--
			buf[i] = 'b'
			i--
			buf[i] = '0'
		case 8:
			if buf[i] != '0' {
				i--
				buf[i] = '0'
			}
		case 16:
			i--
			buf[i] = digits[16]
			i--
			buf[i] = '0'
		}
	}
	if verb == 'O' {
		i--
		buf[i] = 'o'
		i--
		buf[i] = '0'
	}

	if negative {
		i--
		buf[i] = '-'
	} else if f.flags.plus {
		i--
		buf[i] = '+'
	} else if f.flags.space {
		i--
		buf[i] = ' '
	}

	// The code above already wrote the left zero padding, or an explicit
	// precision made the zero flag void.
	oldZero := f.flags.zero
	f.flags.zero = false
	f.pad(buf[i:])
	f.flags.zero = oldZero
}

// truncateString cuts s to the precision, if the format has one.
func (f *formatter) truncateString(s string) string {
	if !f.flags.precPresent {
		return s
	}
	n := f.prec
	i := 0
	for i < len(s) {
		if n == 0 {
			return s[:i]
		}
		n--
		_, wid := utf8.DecodeRuneInString(s[i:])
		i += wid
	}
	return s
}

// fmtS formats a string.
func (f *formatter) fmtS(s string) {
	s = f.truncateString(s)
	f.padString(s)
}

// fmtC formats an integer as a character.
// An invalid code point prints as '�'.
func (f *formatter) fmtC(c uint64) {
	// A conversion of a uint64 to a rune can lose the bits that mark the
	// overflow, so compare against MaxRune first.
	r := rune(c)
	if c > utf8.MaxRune {
		r = utf8.RuneError
	}
	buf := f.intbuf[:0]
	f.pad(utf8.AppendRune(buf, r))
}

// floatBufLen returns the buffer size that fmtFloat needs for verb and prec.
// AppendFloat asks for prec+4 bytes, and that bound holds for a format with
// an exponent. The 'f' verb writes every digit before the decimal point. The
// 'g' verb writes up to prec digits on both sides of the point. The sharp
// flag then restores up to prec trailing zeros.
func floatBufLen(verb rune, prec int) int {
	digits := prec
	if digits < 0 {
		digits = strconv.MaxFloat64Len
	}
	size := 3*digits + 16
	if verb == 'f' {
		size += maxFloatIntDigits
	}
	return size
}

// fmtFloat formats a float64. verb must be a format specifier of
// [strconv.AppendFloat], so it fits into a byte.
func (f *formatter) fmtFloat(v float64, size int, verb rune, prec int) {
	var tailBuf [6]byte // room for "e+123" or "p-1023"

	// An explicit precision in the format overrules the default precision.
	if f.flags.precPresent {
		prec = f.prec
	}
	// Format the number, reserving space for a leading + sign.
	// The buffer lives until fmtFloat returns, which is long enough: the
	// value reaches the output buffer below.
	num := make([]byte, 1, floatBufLen(verb, prec))
	num = strconv.AppendFloat(num, v, byte(verb), prec, size)
	if num[1] == '-' || num[1] == '+' {
		num = num[1:]
	} else {
		num[0] = '+'
	}
	// The space flag asks for a leading space in place of a "+" sign, unless
	// the plus flag asks for the sign.
	if f.flags.space && num[0] == '+' && !f.flags.plus {
		num[0] = ' '
	}
	// An infinity and a NaN do not look like a number, so they get no zero
	// padding.
	if num[1] == 'I' || num[1] == 'N' {
		oldZero := f.flags.zero
		f.flags.zero = false
		// Remove the sign before a NaN if the format does not ask for it.
		if num[1] == 'N' && !f.flags.space && !f.flags.plus {
			num = num[1:]
		}
		f.pad(num)
		f.flags.zero = oldZero
		return
	}
	// The sharp flag forces a decimal point in a format that is not binary,
	// and it keeps the trailing zeros, which the code below restores.
	if f.flags.sharp && verb != 'b' {
		digits := 0
		switch verb {
		case 'v', 'g', 'G', 'x':
			digits = prec
			// A format with no explicit precision keeps 6 digits.
			if digits == -1 {
				digits = 6
			}
		}

		tail := tailBuf[:0]

		hasDecimalPoint := false
		sawNonzeroDigit := false
		// Start at i = 1 to skip the sign at num[0].
		for i := 1; i < len(num); i++ {
			c := num[i]
			isExponent := c == 'p' || c == 'P'
			// In a hexadecimal float 'e' and 'E' are digits.
			if (c == 'e' || c == 'E') && verb != 'x' && verb != 'X' {
				isExponent = true
			}
			if c == '.' {
				hasDecimalPoint = true
			} else if isExponent {
				// Cut the exponent off and put it back after the zeros.
				tail = append(tail, num[i:]...)
				num = num[:i]
			} else {
				if c != '0' {
					sawNonzeroDigit = true
				}
				// Count the significant digits after the first digit that is
				// not zero.
				if sawNonzeroDigit {
					digits--
				}
			}
		}
		if !hasDecimalPoint {
			// A leading digit 0 counts once.
			if len(num) == 2 && num[1] == '0' {
				digits--
			}
			num = append(num, '.')
		}
		for digits > 0 {
			num = append(num, '0')
			digits--
		}
		num = append(num, tail...)
	}
	// Write a sign if the format asks for one, or if the sign is not positive.
	if f.flags.plus || num[0] != '+' {
		// Zero padding to the left must come after the sign. Write the sign
		// out, then pad the unsigned number.
		if f.flags.zero && !f.flags.minus && f.flags.widPresent && f.wid > len(num) {
			f.buf.writeByte(num[0])
			f.writePadding(f.wid - len(num))
			f.buf.write(num[1:])
			return
		}
		f.pad(num)
		return
	}
	// The number is positive and needs no sign. Write the unsigned number.
	f.pad(num[1:])
}
