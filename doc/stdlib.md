# So standard library

So provides low-level packages that wrap the libc API (`so/c/*`) and a growing set of high-level packages. For full API details, see the [package documentation](https://pkg.go.dev/github.com/nalgeon/solod/so).

[so/errors](#soerrors) •
[so/io](#soio) •
[so/math/bits](#somathbits) •
[so/mem](#somem) •
[so/slices](#soslices) •
[so/unicode](#sounicode) •
[so/unicode/utf8](#sounicodeutf8) •
[so/c](#soc) •
[so/c/assert](#socassert) •
[so/c/ctype](#socctype) •
[so/c/cstring](#soccstring) •
[so/c/math](#socmath) •
[so/c/stdio](#socstdio) •
[so/c/stdlib](#socstdlib) •
[so/c/time](#soctime)

## [so/bytes](https://pkg.go.dev/github.com/nalgeon/solod/so/bytes)

Byte slice operations. Offers an API similar to Go's `bytes` package, but with fewer features.

Notable mentions:

- `Clone` returns a copy of a slice.
- `Compare` and `Equal` compare two slices lexicographically.
- `Contains` reports whether a subslice is within a slice.
- `Count` counts the number of non-overlapping instances of a subslice in a slice.
- `Fields`, `Split` and `SplitN` split a slice into subslices.
- `HasPrefix` and `HasSuffix` reports whethe a byte slice begins with a prefix.
- `Index` and `LastIndex` search for a subslice within a slice.
- `Join` concatenates two slices.
- `Replace` and `ReplaceAll` replace subslices within a slice.
- `Runes` converts a byte slice to a rune slice.

## [so/errors](https://pkg.go.dev/github.com/nalgeon/solod/so/errors)

Error creation from text messages.

- `New(text string) error` - create a new error with the given message.

To avoid heap allocations, `New` can only be used at the package level.

## [so/io](https://pkg.go.dev/github.com/nalgeon/solod/so/io)

Basic interfaces to I/O primitives. Offers an API similar to Go's `io` package, but with fewer features.

Notable mentions:

- `Reader` and `Writer` are basic concepts for anything that reads or writes bytes.
- `Copy`, `CopyBuffer` and `CopyN` copy data from a reader to a writer.
- `ReadAll`, `ReadFull` and `ReadAtLeast` read data from a reader.
- `LimitReader` and `TeeReader` implement specialized readers.

## [so/math/bits](https://pkg.go.dev/github.com/nalgeon/solod/so/math/bits)

Bit counting and manipulation functions. Offers the same API as Go's `math/bits` package.

## [so/mem](https://pkg.go.dev/github.com/nalgeon/solod/so/mem)

Memory allocation with a pluggable allocator interface.

- `Alloc` / `Free` - allocate/free a single value.
- `AllocSlice` / `FreeSlice` - allocate/free a slice.
- `Allocator` interface - custom allocator support (`Alloc`, `Realloc`, `Free`).
- `SystemAllocator` - default allocator backed by C `calloc`/`realloc`/`free`.

## [so/slices](https://pkg.go.dev/github.com/nalgeon/solod/so/slices)

Operations on slices.

- `Append` - append elements to a heap slice, growing if needed.
- `Extend` - append another slice to a heap slice, growing if needed.

## [so/unicode](https://pkg.go.dev/github.com/nalgeon/solod/so/unicode)

Data and functions to test certain properties of Unicode code points. Offers an API similar to Go's `unicode` package, but with fewer Unicode features (no support for graphic characters, punctuation, symbols, etc.).

Notable mentions:

- `IsDigit`, `IsLetter` and `IsSpace` check for different character classes.
- `IsLower`, `IsUpper` and `IsTitle` check for character case.
- `ToLower`, `ToUpper` and `ToTitle` change the character case.

## [so/unicode/utf8](https://pkg.go.dev/github.com/nalgeon/solod/so/unicode/utf8)

Functions to convert between runes and UTF-8 byte sequences. Offers the same API as Go's `unicode/utf8` package.

Notable mentions:

- `DecodeRune` and `DecodeRuneInString` unpacks a UTF-8-encoded rune from a byte slice or a string.
- `EncodeRune` writes a UTF-8-encoded rune into a byte slice.
- `RuneCount` and `RuneCountInString` return the number of runes in a byte slice or a string.
- `ValidString` reports whether a string consists entirely of valid UTF-8-encoded runes.

## [so/c](https://pkg.go.dev/github.com/nalgeon/solod/so/c)

C-to-So type bridge for pointers and strings.

- `Bytes` - wrap a C pointer and length as a byte slice.
- `String` - convert a null-terminated C string to a So string.
- `CharPtr` - cast a `*byte` (`uint8_t*`) to `char*` for C interop.

## [so/c/assert](https://pkg.go.dev/github.com/nalgeon/solod/so/c/assert)

Runtime assertions (wraps C `<assert.h>`).

- `Assert` / `Assertf` - abort if a condition is false.
- `Enabled` - whether assertions are active.

## [so/c/ctype](https://pkg.go.dev/github.com/nalgeon/solod/so/c/ctype)

Character classification and conversion (wraps C `<ctype.h>`).

- `IsAlpha`, `IsDigit`, `IsAlnum`, `IsSpace`, `IsUpper`, `IsLower`, `IsPrint`, `IsPunct`, `IsGraph`, `IsCntrl`, `IsBlank`, `IsXDigit` - classify a character.
- `ToUpper` / `ToLower` - convert case.

## [so/c/cstring](https://pkg.go.dev/github.com/nalgeon/solod/so/c/cstring)

Raw memory block operations (wraps C `<string.h>`).

- `Memcpy` - copy n bytes (non-overlapping).
- `Memmove` - copy n bytes (may overlap).
- `Memset` - fill n bytes with value.
- `Memcmp` - compare n bytes.

## [so/c/math](https://pkg.go.dev/github.com/nalgeon/solod/so/c/math)

Math constants and functions (wraps C `<math.h>`).

Constants: `Pi`, `E`, `Inf`.

Functions:

- `Abs`, `Sqrt`, `Pow`, `Floor`, `Ceil`, `Round` - basic operations.
- `Log`, `Log2`, `Log10`, `Exp` - logarithms and exponentials.
- `Sin`, `Cos`, `Atan2` - trigonometry.
- `Fmin`, `Fmax`, `Fmod` - min, max, remainder.

## [so/c/stdio](https://pkg.go.dev/github.com/nalgeon/solod/so/c/stdio)

File I/O and formatted I/O (wraps C `<stdio.h>`).

Streams: `Stdin`, `Stdout`, `Stderr`, `File` type, `EOF`.

Seek constants: `SeekSet`, `SeekCur`, `SeekEnd`.

File operations:

- `Fopen` / `Fclose` - open/close files.
- `Fread` / `Fwrite` - binary I/O.
- `Fgetc` / `Fputc` - character I/O.
- `Fgets` / `Fputs` - string I/O.
- `Fseek`, `Ftell`, `Fflush`, `Feof`, `Ferror` - stream control.

Formatted I/O:

- `Printf` / `Fprintf` - print.
- `Snprintf` - print to buffer.
- `Scanf`, `Fscanf`, `Sscanf` - scan formatted input.

## [so/c/stdlib](https://pkg.go.dev/github.com/nalgeon/solod/so/c/stdlib)

Process control, memory, and string conversion (wraps C `<stdlib.h>`).

- `Exit` - terminate the program.
- `Malloc` / `Calloc` / `Realloc` / `Free` - raw memory management.
- `Atoi` / `Atof` - string-to-number conversion.
- `Getenv` - read an environment variable.
- `ExitSuccess`, `ExitFailure` - standard exit codes.

## [so/c/time](https://pkg.go.dev/github.com/nalgeon/solod/so/c/time)

Calendar time, broken-down time, and formatting (wraps C `<time.h>`).

Constants: `ClocksPerSec` - number of CPU clock ticks per second.

Types: `TimeT` (calendar time), `Tm` (broken-down time with individual fields).

Functions:

- `Time` - current calendar time.
- `Clock` - processor clock ticks.
- `Difftime` - time difference in seconds.
- `Gmtime` - convert calendar to broken-down time.
- `Mktime` - convert broken-down time to calendar time.
- `Strftime` - format time string.
