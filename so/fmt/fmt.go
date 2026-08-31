// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package fmt formats and scans text. The print family formats with Go's verbs
and writes the bytes Go writes. The scan family reads through the stdio of the
host, so it needs a hosted environment; everything else works freestanding.

The verbs:

	%%	a literal percent sign

	%b	integer, base 2
	%d	integer, base 10, signed
	%u	integer, base 10, unsigned
	%o	integer, base 8
	%O	integer, base 8, with a 0o prefix
	%x	integer, base 16, lowercase
	%X	integer, base 16, uppercase

	%e %E	float, decimal exponent notation
	%f %F	float, decimal notation
	%g %G	float, exponent notation for a large exponent, %f notation otherwise
	%a %A	float, hexadecimal exponent notation

	%c	the character of a code point
	%s	string
	%t	bool
	%p	pointer, base 16 notation, with a 0x prefix

The flags are Go's: '+', '-', '#', ' ' and '0'. A width and a precision are
decimal numbers, and a '*' takes the value from an argument.

Two differences against Go. Solod has no reflection, so the verbs that need type
information are absent: %v, %T, %w, %q and %U. And %u is added, because a call
carries no type information either, so nothing else can tell a signed value
from an unsigned one.

# The verb picks the type

Each verb takes exactly one type, and a print call cannot check the argument
against it. A wrong verb is undefined behavior, as in C.

An integer verb takes an int, and %u takes a uint. A float verb takes a
float64, %c takes a rune, %p takes a pointer, and %s takes a string. A []byte
needs string(b).

A narrower type needs no conversion. The print family is extern nodecay, so
every scalar widens at the call site:

	var i32 int32 = 42
	fmt.Printf("%d", i32)

An unknown verb stops the walk over the format string. The output holds the
text up to that verb, then %!<verb>(MISSING).
*/
package fmt

import (
	"fmt" // for testing

	"solod.dev/so/errors"
	"solod.dev/so/io"
)

//so:embed fmt.h
var fmt_h string

//so:embed fmt.c
var fmt_c string

//so:extern
var (
	ErrPrint = errors.New("print failure")
	ErrScan  = errors.New("scan failure")
)

// Print writes its arguments to [Output], separated by spaces.
// It returns the number of bytes written and any write error encountered.
//
// Since Print only accepts string arguments, most of the time you'd want
// to use the print built-in function instead.
//
//so:extern nodecay
func Print(a ...string) (int, error) {
	// Go's Print adds a space between two operands only when neither is a
	// string, so the separator goes in here.
	args := make([]any, 0, 2*len(a))
	for i, s := range a {
		if i > 0 {
			args = append(args, " ")
		}
		args = append(args, s)
	}
	return fmt.Print(args...)
}

// Println is like Print but adds a newline at the end.
//
// Since Println only accepts string arguments, most of the time you'd want
// to use the println built-in function instead.
//
//so:extern nodecay
func Println(a ...string) (int, error) {
	args := make([]any, len(a))
	for i, s := range a {
		args[i] = s
	}
	return fmt.Println(args...)
}

// Printf formats according to a format specifier and writes to [Output].
// It returns the number of bytes written and any write error encountered.
//
//so:extern nodecay
func Printf(format string, a ...any) (int, error) {
	return fmt.Printf(format, a...)
}

// Sprintf formats according to a format specifier, outputs to buf,
// and returns the resulting string.
// If the output size exceeds buf length, it silently truncates the output.
//
//so:extern nodecay
func Sprintf(buf []byte, format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
//
//so:extern nodecay
func Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

// Scanf scans text read from standard input, storing successive
// space-separated values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
// Scanf requires a hosted environment. A freestanding call panics.
//
//so:extern
func Scanf(format string, a ...any) (int, error) {
	return fmt.Scanf(format, a...)
}

// Sscanf scans the argument string, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
// Sscanf requires a hosted environment. A freestanding call panics.
//
//so:extern
func Sscanf(str string, format string, a ...any) (int, error) {
	return fmt.Sscanf(str, format, a...)
}

// Fscanf scans text read from r, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
// Fscanf requires a hosted environment. A freestanding call panics.
//
//so:extern
func Fscanf(r io.Reader, format string, a ...any) (int, error) {
	return fmt.Fscanf(r, format, a...)
}
