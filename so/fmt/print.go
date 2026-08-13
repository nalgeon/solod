// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ported from Go's fmt/print.go: the walk over the format string of doPrintf,
// and parsenum. Everything below them in Go is reflection over any.
//
// Dropped against Go: the argument index (%[2]d), because a va_list has no
// random access; the fast path for a simple verb, which only saves time; and
// the %!(EXTRA ...) tail, because the C shim reads one argument per verb and
// cannot collect a spare one.

package fmt

import (
	"solod.dev/so/io"
	"solod.dev/so/unicode/utf8"
)

// The error texts of Go's fmt.
const (
	percentBangString = "%!"
	missingString     = "(MISSING)"
	badWidthString    = "%!(BADWIDTH)"
	badPrecString     = "%!(BADPREC)"
	noVerbString      = "%!(NOVERB)"
)

// The kind of the value that an argument holds. The C shim reads the kind of
// a verb from [argKind] and selects the va_arg type with it.
//
//so:promote
const (
	kindNone   = 0 // unknown verb
	kindInt    = 1
	kindUint   = 2
	kindFloat  = 3
	kindRune   = 4
	kindString = 5
	kindBool   = 6
	kindPtr    = 7
)

// arg is one argument, collected by the C shim. So has no unions, so the
// signed, the unsigned, the rune, the boolean and the pointer values share
// the integer field.
//
//so:promote
type arg struct {
	kind int
	i    int
	f    float64
	s    string
}

// argKind returns the kind of the value that verb takes, or kindNone if the
// verb is unknown. A verb takes exactly one type, because a C variadic
// carries no type.
//
//so:promote
func argKind(verb rune) int {
	switch verb {
	case 'd', 'b', 'o', 'O', 'x', 'X':
		return kindInt
	case 'u':
		return kindUint
	case 'e', 'E', 'f', 'F', 'g', 'G', 'a', 'A':
		return kindFloat
	case 'c':
		return kindRune
	case 's':
		return kindString
	case 't':
		return kindBool
	case 'p':
		return kindPtr
	}
	return kindNone
}

// maxNum is the largest width and the largest precision. A larger number in
// the format string is a mistake.
const maxNum = 1000000

// tooLarge reports whether x is too large for a width or a precision.
func tooLarge(x int) bool {
	return x > maxNum || x < -maxNum
}

// printer walks one format string and writes the result to a buffer.
type printer struct {
	f   formatter
	buf buffer

	format string
	args   []arg
	pos    int // position in the format string
	argNum int // position in the arguments
}

// init prepares the printer to write to w.
func (p *printer) init(w io.Writer) {
	p.buf.init(w)
	p.f.init(&p.buf)
}

// parsenum reads a decimal number at the current position and steps over it.
// It reports whether a number is present.
func (p *printer) parsenum() (int, bool) {
	num := 0
	isnum := false
	for p.pos < len(p.format) && '0' <= p.format[p.pos] && p.format[p.pos] <= '9' {
		if tooLarge(num) {
			p.pos = len(p.format) // a number this long is a mistake
			return 0, false
		}
		num = num*10 + int(p.format[p.pos]-'0')
		isnum = true
		p.pos++
	}
	return num, isnum
}

// intFromArg takes the next argument as a width or a precision. It reports
// whether the argument holds an integer.
func (p *printer) intFromArg() (int, bool) {
	if p.argNum >= len(p.args) {
		return 0, false
	}
	a := p.args[p.argNum]
	p.argNum++
	if a.kind != kindInt || tooLarge(a.i) {
		return 0, false
	}
	return a.i, true
}

// scanFlags reads the flags of one verb and steps over them.
func (p *printer) scanFlags() {
	for p.pos < len(p.format) {
		c := p.format[p.pos]
		if c == '#' {
			p.f.flags.sharp = true
		} else if c == '0' {
			p.f.flags.zero = true
		} else if c == '+' {
			p.f.flags.plus = true
		} else if c == '-' {
			p.f.flags.minus = true
		} else if c == ' ' {
			p.f.flags.space = true
		} else {
			return
		}
		p.pos++
	}
}

// star reports whether a '*' is at the current position, and steps over it.
func (p *printer) star() bool {
	if p.pos < len(p.format) && p.format[p.pos] == '*' {
		p.pos++
		return true
	}
	return false
}

// dot reports whether a precision follows, and steps over the point. A point
// at the end of the format string is not a precision.
func (p *printer) dot() bool {
	if p.pos+1 >= len(p.format) || p.format[p.pos] != '.' {
		return false
	}
	p.pos++
	return true
}

// verb reads the verb at the current position and steps over it.
func (p *printer) verb() rune {
	r, size := utf8.DecodeRuneInString(p.format[p.pos:])
	p.pos += size
	return r
}

// scanWidth reads the width of one verb, from the format string or from an
// argument, and steps over it.
func (p *printer) scanWidth() {
	if p.star() {
		wid, ok := p.intFromArg()
		if !ok {
			p.buf.writeString(badWidthString)
		}
		// A negative width is a positive width with the minus flag.
		if wid < 0 {
			wid = -wid
			p.f.flags.minus = true
			p.f.flags.zero = false // do not pad with zeros to the right
		}
		p.f.wid = wid
		p.f.flags.widPresent = ok
		return
	}
	wid, ok := p.parsenum()
	p.f.wid = wid
	p.f.flags.widPresent = ok
}

// scanPrec reads the precision of one verb, from the format string or from an
// argument, and steps over it.
func (p *printer) scanPrec() {
	if !p.dot() {
		return
	}
	if p.star() {
		prec, ok := p.intFromArg()
		// A negative precision has no meaning.
		if prec < 0 {
			prec = 0
			ok = false
		}
		if !ok {
			p.buf.writeString(badPrecString)
		}
		p.f.prec = prec
		p.f.flags.precPresent = ok
		return
	}
	// A point with no number is a precision of 0.
	prec, _ := p.parsenum()
	p.f.prec = prec
	p.f.flags.precPresent = true
}

// missingArg writes the text for a verb with no argument.
func (p *printer) missingArg(verb rune) {
	p.buf.writeString(percentBangString)
	p.buf.writeRune(verb)
	p.buf.writeString(missingString)
}

// fmtInt writes an integer in the base of verb.
func (p *printer) fmtInt(u uint64, verb rune, isSigned bool) {
	switch verb {
	case 'b':
		p.f.fmtInteger(u, 2, isSigned, verb, ldigits)
	case 'o', 'O':
		p.f.fmtInteger(u, 8, isSigned, verb, ldigits)
	case 'x':
		p.f.fmtInteger(u, 16, isSigned, verb, ldigits)
	case 'X':
		p.f.fmtInteger(u, 16, isSigned, verb, udigits)
	default:
		p.f.fmtInteger(u, 10, isSigned, verb, ldigits)
	}
}

// fmtFloat writes a float in the format of verb. A verb with no precision of
// its own keeps 6 places, as Go does.
func (p *printer) fmtFloat(v float64, verb rune) {
	switch verb {
	case 'e', 'E', 'f':
		p.f.fmtFloat(v, 64, verb, 6)
	case 'F':
		p.f.fmtFloat(v, 64, 'f', 6)
	case 'a':
		p.f.fmtFloat(v, 64, 'x', -1)
	case 'A':
		p.f.fmtFloat(v, 64, 'X', -1)
	default:
		p.f.fmtFloat(v, 64, verb, -1)
	}
}

// fmtPointer writes a pointer in hexadecimal with a leading "0x". The sharp
// flag removes the prefix, which is the opposite of what the flag does for an
// integer. Go does the same.
func (p *printer) fmtPointer(u uint64, verb rune) {
	sharp := p.f.flags.sharp
	p.f.flags.sharp = !sharp
	p.f.fmtInteger(u, 16, unsignedInt, verb, ldigits)
	p.f.flags.sharp = sharp
}

// printArg writes one argument. The kind selects the formatter, and the verb
// selects the base of an integer and the format of a float.
func (p *printer) printArg(a arg, verb rune) {
	switch a.kind {
	case kindInt:
		p.fmtInt(uint64(a.i), verb, signedInt)
	case kindUint:
		// The unsigned value sits in a signed field, so the conversion goes
		// through uint. A direct uint64(a.i) extends the sign on a target
		// where an int is narrower than 64 bits.
		p.fmtInt(uint64(uint(a.i)), verb, unsignedInt)
	case kindFloat:
		p.fmtFloat(a.f, verb)
	case kindRune:
		p.f.fmtC(uint64(a.i))
	case kindString:
		p.f.fmtS(a.s)
	case kindBool:
		p.f.fmtBoolean(a.i != 0)
	case kindPtr:
		p.fmtPointer(uint64(uint(a.i)), verb)
	}
}

// doPrintf walks the format string and writes each verb with its argument.
func (p *printer) doPrintf(format string, args []arg) {
	p.format = format
	p.args = args
	p.pos = 0
	p.argNum = 0

	end := len(format)
	for p.pos < end {
		// Copy the text up to the next verb.
		start := p.pos
		for p.pos < end && format[p.pos] != '%' {
			p.pos++
		}
		if p.pos > start {
			p.buf.writeString(format[start:p.pos])
		}
		if p.pos >= end {
			return
		}
		p.pos++ // step over the '%'

		p.f.clearflags()
		p.scanFlags()
		p.scanWidth()
		p.scanPrec()

		if p.pos >= end {
			p.buf.writeString(noVerbString)
			return
		}
		verb := p.verb()
		if verb == '%' {
			// A percent takes no argument and ignores the width and the
			// precision.
			p.buf.writeByte('%')
			continue
		}
		if argKind(verb) == kindNone {
			// The C shim stops at a verb it does not know, so the arguments
			// after this verb are missing as well.
			p.missingArg(verb)
			return
		}
		if p.argNum >= len(p.args) {
			p.missingArg(verb)
			continue
		}
		p.printArg(p.args[p.argNum], verb)
		p.argNum++
	}
}

// setKind writes kind at index n, if the index is inside kinds.
func setKind(kinds []int, n int, kind int) {
	if n < len(kinds) {
		kinds[n] = kind
	}
}

// argKinds fills kinds with the kind of every argument that format needs, in
// order, and returns the number of arguments. A width or a precision from an
// argument takes a kindInt slot before the slot of its verb. The walk stops at
// an unknown verb, because the collector cannot pull a value that it cannot
// name.
//
// argKinds writes no more kinds than kinds holds, and it returns the full
// count. The collector calls it twice: once with an empty slice for the count,
// and once with an array of that size. It steps over the format string with
// the methods of [printer.doPrintf], so the collector and the walk agree about
// every argument.
//
//so:promote
func argKinds(format string, kinds []int) int {
	var p printer // a scratch printer, only for the scanning methods
	p.format = format
	end := len(format)
	n := 0
	for p.pos < end {
		for p.pos < end && format[p.pos] != '%' {
			p.pos++
		}
		if p.pos >= end {
			return n
		}
		p.pos++ // step over the '%'

		p.scanFlags()
		if p.star() {
			setKind(kinds, n, kindInt)
			n++
		} else {
			p.parsenum()
		}
		if p.dot() {
			if p.star() {
				setKind(kinds, n, kindInt)
				n++
			} else {
				p.parsenum()
			}
		}

		if p.pos >= end {
			return n
		}
		verb := p.verb()
		if verb == '%' {
			continue
		}
		kind := argKind(verb)
		if kind == kindNone {
			return n
		}
		setKind(kinds, n, kind)
		n++
	}
	return n
}

// vfprint formats args with format and writes the result to w. It returns the
// number of bytes written and the first write error.
//
//so:promote
func vfprint(w io.Writer, format string, args []arg) (int, error) {
	var p printer
	p.init(w)
	p.doPrintf(format, args)
	p.buf.flush()
	return p.buf.total, p.buf.err
}

// vsprint formats args with format and writes the result to buf. It returns
// the result as a string, which points into buf. Output that does not fit buf
// is dropped.
//
//so:promote
func vsprint(buf []byte, format string, args []arg) string {
	var w bufWriter
	w.init(buf)
	var p printer
	p.init(&w)
	p.doPrintf(format, args)
	p.buf.flush()
	return w.output()
}

// vjoin writes args to w, separated by spaces, and adds a newline if newline
// is true. Every argument is a string. It returns the number of bytes written
// and the first write error.
//
//so:promote
func vjoin(w io.Writer, args []arg, newline bool) (int, error) {
	var b buffer
	b.init(w)
	for i, a := range args {
		if i > 0 {
			b.writeByte(' ')
		}
		b.writeString(a.s)
	}
	if newline {
		b.writeByte('\n')
	}
	b.flush()
	return b.total, b.err
}
