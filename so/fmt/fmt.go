/*
Package fmt implements formatted I/O with functions analogous
to C's printf and scanf. The format 'verbs' are the same as in C
(not the ones used in Go):

	%%	literal percent sign

	%d	integer, base 10, signed
	%u	integer, base 10, unsigned
	%o	integer, base 8, unsigned
	%x	integer, base 16, unsigned

	%f	floating-point, decimal notation
	%e	floating-point, decimal exponent notation
	%a	floating-point, hexadecimal exponent notation
	%g	floating-point, decimal or exponent notation as needed

	%c	single literal character
	%s	character string

	%p	pointer, base 16 notation, with leading 0x
*/
package fmt

import (
	"fmt" // for testing

	"solod.dev/so/c"
	"solod.dev/so/errors"
	"solod.dev/so/io"
)

//so:embed fmt.h
var fmt_h string

//so:embed fmt.c
var fmt_c string

// BufSize is the size of the internal formatting buffer in bytes.
//
//so:extern
const BufSize = 1024

//so:extern
var (
	ErrPrint = errors.New("print failure")
	ErrScan  = errors.New("scan failure")
	ErrSize  = errors.New("buffer size exceeded")
)

// Buffer is a fixed-size stack-allocated buffer
// for formatted output and scanning.
type Buffer struct {
	Ptr *byte
	Len int
}

// NewBuffer creates a new stack-allocated Buffer of the given size.
//
//so:extern
func NewBuffer(size int) Buffer {
	b := make([]byte, size)
	return Buffer{
		Ptr: &b[0],
		Len: size,
	}
}

// BufferFrom creates a Buffer that uses the provided byte slice as its storage.
// The buffer doesn't take ownership of the slice and doesn't free it.
func BufferFrom(buf []byte) Buffer {
	return Buffer{
		Ptr: &buf[0],
		Len: len(buf),
	}
}

// String returns the contents of the Buffer as a string,
// up to the first null byte.
func (b Buffer) String() string {
	return c.String(b.Ptr)
}

// Print writes its arguments to standard output, separated by spaces.
// It returns the number of bytes written and any write error encountered.
//
// Since Print only accepts string arguments, most of the time you'd want
// to use the print built-in function instead.
//
//so:extern
func Print(a ...string) (int, error) {
	args := make([]any, len(a))
	for i, s := range a {
		args[i] = s
	}
	return fmt.Print(args...)
}

// Println is like Print but adds a newline at the end.
//
// Since Println only accepts string arguments, most of the time you'd want
// to use the println built-in function instead.
//
//so:extern
func Println(a ...string) (int, error) {
	args := make([]any, len(a))
	for i, s := range a {
		args[i] = s
	}
	return fmt.Println(args...)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
//
//so:extern
func Printf(format string, a ...any) (int, error) {
	return fmt.Printf(format, a...)
}

// Sprintf formats according to a format specifier, outputs to buf,
// and returns the resulting string.
// If the output size exceeds buf length, it silently truncates the output.
//
//so:extern
func Sprintf(buf Buffer, format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
// Returns [ErrSize] if the output size exceeds BufSize.
//
//so:extern
func Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

// Scanf scans text read from standard input, storing successive
// space-separated values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
//so:extern
func Scanf(format string, a ...any) (int, error) {
	return fmt.Scanf(format, a...)
}

// Sscanf scans the argument string, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
//so:extern
func Sscanf(str string, format string, a ...any) (int, error) {
	return fmt.Sscanf(str, format, a...)
}

// Fscanf scans text read from r, storing successive space-separated
// values into successive arguments as determined by the format.
// It returns the number of items successfully scanned.
//
//so:extern
func Fscanf(r io.Reader, format string, a ...any) (int, error) {
	return fmt.Fscanf(r, format, a...)
}
