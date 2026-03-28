# So standard library

Solod provides a growing set of high-level packages similar to Go's stdlib, as well as low-level packages that wrap the libc API. For full API details, see the [package documentation](https://pkg.go.dev/solod.dev/so).

High-level:
[bytes](#sobytes) •
[errors](#soerrors) •
[fmt](#sofmt) •
[io](#soio) •
[maps](#somaps) •
[math](#somath) •
[math/bits](#somathbits) •
[mem](#somem) •
[os](#soos) •
[slices](#soslices) •
[strconv](#sostrconv) •
[strings](#sostrings) •
[time](#sotime) •
[unicode](#sounicode) •
[unicode/utf8](#sounicodeutf8)

Low-level:
[c](#soc) •
[c/assert](#socassert) •
[c/ctype](#socctype) •
[c/cstring](#soccstring) •
[c/math](#socmath) •
[c/stdio](#socstdio) •
[c/stdlib](#socstdlib) •
[c/time](#soctime)

## [so/bytes](https://pkg.go.dev/solod.dev/so/bytes)

Byte slice operations. Offers an API similar to Go's `bytes` package, but with fewer features.

Functions:

- `Clone` returns a copy of a slice.
- `Compare` and `Equal` compare two slices lexicographically.
- `Contains` reports whether a subslice is within a slice.
- `Count` counts the number of non-overlapping instances of a subslice in a slice.
- `Cut` slices around the first instance of a separator.
- `HasPrefix` and `HasSuffix` report whether a slice begins/ends with a prefix/suffix.
- `Index` and `IndexByte` search for a subslice or byte within a slice.
- `Join` concatenates slices with a separator.
- `Replace` replaces occurrences of a subslice within a slice.
- `Runes` converts a byte slice to a rune slice.
- `Split` and `SplitN` split a slice into subslices.
- `String` creates a string from a byte slice.
- `ToLower` and `ToUpper` return a copy with all letters lowercased/uppercased.
- `TrimLeft`, `TrimRight` and `TrimSpace` trim characters from a slice.
- `TrimPrefix` and `TrimSuffix` trim a prefix/suffix from a slice.

Types:

- `Buffer` is a variable-sized buffer of bytes with `Read` and `Write` methods.
- `Reader` reads data from a byte slice.

## [so/errors](https://pkg.go.dev/solod.dev/so/errors)

Error creation from text messages.

- `New(text string) error` - create a new error with the given message.

So only supports sentinel errors, which are defined at the package level using `New`.

## [so/fmt](https://pkg.go.dev/solod.dev/so/fmt)

Formatted I/O with functions analogous to C's printf and scanf. Uses C format verbs (not Go verbs).

- `Print` and `Println` write strings to standard output.
- `Printf` formats and writes to standard output.
- `Fprintf` formats and writes to an `io.Writer`.
- `Scanf` scans formatted text from standard input.
- `Sscanf` scans formatted text from a string.
- `Fscanf` scans formatted text from an `io.Reader`.

## [so/io](https://pkg.go.dev/solod.dev/so/io)

Basic interfaces to I/O primitives. Offers an API similar to Go's `io` package, but with fewer features.

Functions:

- `Copy` and `CopyN` copy data from a reader to a writer.
- `ReadAll` and `ReadFull` read data from a reader.

Types:

- `Reader`, `Writer`, and `Closer` are basic concepts for anything that does I/O.
- `LimitedReader` and `SectionReader` implement specialized readerss.
- `Discard` is a no-op writer.

## [so/maps](https://pkg.go.dev/solod.dev/so/maps)

Generic hashmap similar to Go's built-in `map[K]V`, backed by a Robin Hood hash table with automatic grow/shrink.

Functions:

- `New` creates a new `Map` with a given allocator.

Types:

- `Map[K, V]` is a generic hashmap with `Get`, `Set`, and `Delete` methods.

## [so/math](https://pkg.go.dev/solod.dev/so/math)

Mathematical functions and constants. Offers the same API as Go's `math` package.

## [so/math/bits](https://pkg.go.dev/solod.dev/so/math/bits)

Bit counting and manipulation functions. Offers the same API as Go's `math/bits` package.

## [so/mem](https://pkg.go.dev/solod.dev/so/mem)

Memory allocation with a pluggable allocator interface.

Functions:

- `Alloc` / `Free` - allocate/free a single value.
- `AllocSlice` / `FreeSlice` - allocate/free a slice.

Types:

- `Allocator` interface - custom allocator support (`Alloc`, `Realloc`, `Free`).
- `SystemAllocator` - default allocator backed by C `calloc`/`realloc`/`free`.
- `Arena` - bump allocator backed by a fixed buffer (`Alloc`, `Realloc`, `Reset`).

## [so/os](https://pkg.go.dev/solod.dev/so/os)

File I/O and filesystem operations. Offers an API similar to Go's `os` package, but with fewer features.

Functions:

- `Create` creates or truncates a file for reading and writing.
- `Open` opens a file for reading.
- `ReadFile` reads an entire file into a byte slice.
- `WriteFile` writes a byte slice to a file, creating or truncating it.
- `Rename` renames (moves) a file.
- `Remove` removes a file or empty directory.
- `Getenv` returns the value of an environment variable.
- `Exit` terminates the program with the given status code.

Types:

- `File` represents an open file with methods for reading and writing data.

`so/os` is currently built on libc's `stdio.h`, so it's much more limited than Go's `os` package.

## [so/slices](https://pkg.go.dev/solod.dev/so/slices)

Operations on slices:

- `Make` and `MakeCap` allocate a slice, `Free` deallocates it.
- `Append` appends elements to a heap slice, growing if needed.
- `Extend` appends another slice to a heap slice, growing if needed.
- `Clone` creates a shallow copy of the slice.
- `Equal` reports whether two slices are equal.

## [so/strconv](https://pkg.go.dev/solod.dev/so/strconv)

Conversions between numbers and strings. Offers an API similar to Go's `strconv` package, but with fewer features.

- `ParseBool` and `FormatBool` convert bool ↔ string.
- `Atoi`, `Itoa`, `ParseInt` and `FormatInt` convert signed integer ↔ string.
- `ParseUint` and `FormatUint` convert unsigned integer ↔ string.
- `ParseFloat` and `FormatFloat` convert float ↔ string.

## [so/strings](https://pkg.go.dev/solod.dev/so/strings)

String operations. Offers an API similar to Go's `strings` package, but with fewer features.

Functions:

- `Clone` returns a fresh copy of a string.
- `Compare` compares two strings lexicographically.
- `Contains` and `ContainsFunc` report whether a substring is within a string.
- `Count` counts the number of non-overlapping instances of a substring in a string.
- `Cut` slices a string around a separator.
- `Fields` and `FieldsFunc` split a string around whitespace or a predicate.
- `HasPrefix` and `HasSuffix` report whether a string begins/ends with a prefix/suffix.
- `Index` and `IndexFunc` search for a substring within a string.
- `Join` concatenates string slices with a separator.
- `Repeat` returns a string consisting of count copies of a string.
- `Replace` and `ReplaceAll` replace occurrences of a substring within a string.
- `Split` and `SplitN` split a string into substrings.
- `ToLower` and `ToUpper` return a copy with all letters lowercased/uppercased.
- `Trim`, `TrimFunc` and `TrimSpace` trim characters from a string.
- `TrimPrefix` and `TrimSuffix` trim a prefix/suffix from a string.

Types:

- `Builder` efficiently builds a string, minimizing memory copying.
- `Reader` reads data from a string.

## [so/time](https://pkg.go.dev/solod.dev/so/time)

Measuring and displaying time. Offers an API similar to Go's `time` package, but handles locations, formatting, and parsing differently.

Time is always stored as UTC internally. Formatting and parsing use C strftime/strptime verbs (e.g. `%Y-%m-%d %H:%M:%S`).

Constants:

- `UTC` - zero offset (UTC).
- `Nanosecond`, `Microsecond`, `Millisecond`, `Second`, `Minute`, `Hour` - common durations.

Functions:

- `Now` returns the current time in UTC (with monotonic clock reading).
- `Date` returns the Time for a given year, month, day, hour, min, sec, nsec, and offset (seconds east of UTC).
- `Unix`, `UnixMilli`, `UnixMicro` create a Time from a Unix timestamp.
- `Since` and `Until` return the duration elapsed since or until a given time.
- `Parse` parses a time string per layout (strptime verbs) with a given offset, returning a Time.

Types:

- `Time` represents an instant in time with nanosecond precision. Always UTC.
- `Duration` represents elapsed time as an int64 nanosecond count.
- `CalDate` is a date specified by year, month, and day.
- `CalClock` is a time of day specified by hour, minute, and second.
- `Offset` represents a fixed offset from UTC in seconds.

## [so/unicode](https://pkg.go.dev/solod.dev/so/unicode)

Data and functions to test certain properties of Unicode code points. Offers an API similar to Go's `unicode` package, but with fewer Unicode features (no support for graphic characters, punctuation, symbols, etc.).

- `IsDigit`, `IsLetter` and `IsSpace` check for different character classes.
- `IsLower`, `IsUpper` and `IsTitle` check for character case.
- `ToLower`, `ToUpper` and `ToTitle` change the character case.

## [so/unicode/utf8](https://pkg.go.dev/solod.dev/so/unicode/utf8)

Functions to convert between runes and UTF-8 byte sequences. Offers the same API as Go's `unicode/utf8` package.

- `DecodeRune` and `DecodeRuneInString` unpacks a UTF-8-encoded rune from a byte slice or a string.
- `EncodeRune` writes a UTF-8-encoded rune into a byte slice.
- `RuneCount` and `RuneCountInString` return the number of runes in a byte slice or a string.
- `ValidString` reports whether a string consists entirely of valid UTF-8-encoded runes.

## [so/c](https://pkg.go.dev/solod.dev/so/c)

C-to-So type bridge for pointers and strings.

- `Bytes` - wrap a C pointer and length as a byte slice.
- `String` - convert a null-terminated C string to a So string.
- `CharPtr` - cast a `*byte` (`uint8_t*`) to `char*` for C interop.

## [so/c/assert](https://pkg.go.dev/solod.dev/so/c/assert)

Runtime assertions (wraps C `<assert.h>`).

- `Assert` / `Assertf` - abort if a condition is false.
- `Enabled` - whether assertions are active.

## [so/c/ctype](https://pkg.go.dev/solod.dev/so/c/ctype)

Character classification and conversion (wraps C `<ctype.h>`).

- `IsAlpha`, `IsDigit`, `IsAlnum`, `IsSpace`, `IsUpper`, `IsLower`, `IsPrint`, `IsPunct`, `IsGraph`, `IsCntrl`, `IsBlank`, `IsXDigit` - classify a character.
- `ToUpper` / `ToLower` - convert case.

## [so/c/cstring](https://pkg.go.dev/solod.dev/so/c/cstring)

Raw memory block operations (wraps C `<string.h>`).

- `Memcpy` - copy n bytes (non-overlapping).
- `Memmove` - copy n bytes (may overlap).
- `Memset` - fill n bytes with value.
- `Memcmp` - compare n bytes.

## [so/c/math](https://pkg.go.dev/solod.dev/so/c/math)

Math constants and functions (wraps C `<math.h>`).

Constants: `Pi`, `E`, `Inf`.

Functions:

- `Abs`, `Sqrt`, `Pow`, `Floor`, `Ceil`, `Round` - basic operations.
- `Log`, `Log2`, `Log10`, `Exp` - logarithms and exponentials.
- `Sin`, `Cos`, `Atan2` - trigonometry.
- `Fmin`, `Fmax`, `Fmod` - min, max, remainder.

## [so/c/stdio](https://pkg.go.dev/solod.dev/so/c/stdio)

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

## [so/c/stdlib](https://pkg.go.dev/solod.dev/so/c/stdlib)

Process control, memory, and string conversion (wraps C `<stdlib.h>`).

- `Exit` - terminate the program.
- `Malloc` / `Calloc` / `Realloc` / `Free` - raw memory management.
- `Atoi` / `Atof` - string-to-number conversion.
- `Getenv` - read an environment variable.
- `ExitSuccess`, `ExitFailure` - standard exit codes.

## [so/c/time](https://pkg.go.dev/solod.dev/so/c/time)

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
