package stdio

import (
	_ "embed"
)

//so:embed stdio.h
var stdio_h string

// EOF is returned by several functions to indicate end-of-file
// or an error condition.
//
//so:extern
var EOF int

// SeekWhence represents the reference point for file seeking operations.
//
//so:extern
const SeekSet int = 0 // SEEK_SET
//so:extern
const SeekCur int = 1 // SEEK_CUR
//so:extern
const SeekEnd int = 2 // SEEK_END

// File represents a file stream.
//
//so:extern
type File struct{}

// Stdin represents the standard input stream.
//
//so:extern
var Stdin *File

// Stdout represents the standard output stream.
//
//so:extern
var Stdout *File

// Stderr represents the standard error stream.
//
//so:extern
var Stderr *File

// Fopen opens a file indicated by path and returns a pointer to the file stream
// associated with that file. mode is used to determine the file access mode.
//
// If successful, returns a pointer to the new file stream. The stream is fully
// buffered unless filename refers to an interactive device. On error, returns a null pointer.
//
//so:extern
func Fopen(path string, mode string) *File { _, _ = path, mode; return nil }

// Fclose closes the given file stream. Any unread buffered data are discarded.
//
// Whether or not the operation succeeds, the stream is no longer associated with
// a file, and the buffer allocated by setbuf or setvbuf, if any, is also disassociated
// and deallocated if automatic allocation was used.
//
// The behavior is undefined if the value of the pointer stream is used after Fclose returns.
//
// Returns zero on success, [EOF] otherwise.
//
//so:extern
func Fclose(stream *File) int { _ = stream; return 0 }

// Fflush writes any unwritten data from the stream's buffer to the associated output device.
//
// If stream is a null pointer, all open output streams are flushed, including
// the ones manipulated within library packages or otherwise not directly
// accessible to the program.
//
// Returns zero on success, [EOF] otherwise.
//
//so:extern
func Fflush(stream *File) int { _ = stream; return 0 }

// Fseek sets the file position indicator for the stream to the value pointed to by offset.
//
// If the stream is open in binary mode, the new position is exactly offset bytes
// measured from the beginning of the file if origin is [SeekSet], from the current
// file position if origin is [SeekCur], and from the end of the file if origin is [SeekEnd].
// Binary streams are not required to support [SeekEnd], in particular if additional
// null bytes are output.
//
// If the stream is open in text mode, the only supported values for offset are zero
// (which works with any origin) and a value returned by an earlier call to ftell on
// a stream associated with the same file (which only works with origin of [SeekSet]).
//
// If the stream is wide-oriented, the restrictions of both text and binary streams
// apply (result of ftell is allowed with [SeekSet] and zero offset is allowed from
// [SeekSet] and [SeekCur], but not [SeekEnd]).
//
// In addition to changing the file position indicator, Fseek undoes the effects
// of ungetc and clears the end-of-file status, if applicable.
//
// If a read or write error occurs, the error indicator for the stream ([Ferror])
// is set and the file position is unaffected.
//
// Returns zero on success, [EOF] otherwise.
//
//so:extern
func Fseek(stream *File, offset int, whence int) int {
	_, _, _ = stream, offset, whence
	return 0
}

// Fgetc reads the next character from the input stream.
//
// On success, returns the obtained character as an unsigned char
// converted to an int. On failure, returns EOF.
//
//so:extern
func Fgetc(stream *File) int { _ = stream; return 0 }

// Fputc writes a character to the output stream.
//
// On success, returns the character written as an unsigned char
// converted to an int. On failure, returns EOF.
//
//so:extern
func Fputc(ch int, stream *File) int { _, _ = ch, stream; return 0 }

// Fgets reads at most n-1 characters from the input stream into s.
//
// Parsing stops if a newline character is found (in which case str will contain
// that newline character) or if end-of-file occurs. If bytes are read and no errors
// occur, writes a null character at the position immediately after the last
// character written to str.
//
// Returns s on success, and a null pointer on error or when
// end-of-file occurs while no characters have been read.
//
//so:extern
func Fgets(s *byte, n int, stream *File) *byte { _, _, _ = s, n, stream; return nil }

// Fputs writes a string to the output stream.
//
// On success, returns a non-negative value. On failure, returns EOF.
//
//so:extern
func Fputs(s string, stream *File) int { _, _ = s, stream; return 0 }

// Fread reads up to count elements of size bytes from the input stream into ptr.
// The file position indicator for the stream is advanced by the number of bytes read.
// If an error occurs, the resulting value of the file position indicator for the
// stream is indeterminate.
//
// Returns number of objects read successfully, which may be less than count
// if an error or end-of-file occurs.
//
// If size or count is zero, returns zero and performs no other action.
// Fread does not distinguish between end-of-file and error, and callers
// must use [Feof] and [Ferror] to determine which occurred.
//
//so:extern
func Fread(ptr *byte, size int, count int, stream *File) int {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

// Fwrite writes count elements of size bytes from ptr to the output stream.
// The file position indicator for the stream is advanced by the number of bytes written.
// If an error occurs, the resulting value of the file position indicator for the
// stream is indeterminate.
//
// Returns the number of objects written successfully, which may be less than count
// if an error occurs. If size or count is zero, returns zero and performs no other action.
//
//so:extern
func Fwrite(ptr *byte, size int, count int, stream *File) int {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

// Feof reports whether the end of the stream has been reached.
//
//so:extern
func Feof(stream *File) bool { _ = stream; return false }

// Ferror reports whether the stream has encountered an error.
//
//so:extern
func Ferror(stream *File) bool { _ = stream; return false }
