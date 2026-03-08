# So standard library

So provides low-level packages that wrap the libc API (`so/c/*`) and a growing set of high-level packages. For full API details, see the [package documentation](https://pkg.go.dev/github.com/nalgeon/solod/so).

[so/errors](#soerrors) •
[so/mem](#somem) •
[so/c](#soc) •
[so/c/assert](#socassert) •
[so/c/ctype](#socctype) •
[so/c/cstring](#soccstring) •
[so/c/math](#socmath) •
[so/c/stdio](#socstdio) •
[so/c/stdlib](#socstdlib)

## so/errors

Error creation from text messages.

- `New(text string) error` - create a new error with the given message.

To avoid heap allocations, `New` can only be used at the package level.

## so/mem

Memory allocation with a pluggable allocator interface.

- `New` / `Free` - allocate/free a single value (system allocator).
- `NewSlice` / `FreeSlice` - allocate/free a slice (system allocator).
- `Alloc` / `Dealloc` - allocate/free with a custom allocator.
- `AllocSlice` / `DeallocSlice` - same for slices.
- `Allocator` interface - custom allocator support (`Alloc`, `Realloc`, `Dealloc`).
- `SystemAllocator` - default allocator backed by C `calloc`/`realloc`/`free`.

## so/c

C-to-So type bridge for pointers and strings.

- `Bytes` - wrap a C pointer and length as a byte slice.
- `String` - convert a null-terminated C string to a So string.
- `CharPtr` - cast a `*byte` (`uint8_t*`) to `char*` for C interop.

## so/c/assert

Runtime assertions (wraps C `<assert.h>`).

- `Assert` / `Assertf` - abort if a condition is false.
- `Enabled` - whether assertions are active.

## so/c/ctype

Character classification and conversion (wraps C `<ctype.h>`).

- `IsAlpha`, `IsDigit`, `IsAlnum`, `IsSpace`, `IsUpper`, `IsLower`, `IsPrint`, `IsPunct`, `IsGraph`, `IsCntrl`, `IsBlank`, `IsXDigit` - classify a character.
- `ToUpper` / `ToLower` - convert case.

## so/c/cstring

Raw memory block operations (wraps C `<string.h>`).

- `Memcpy` - copy n bytes (non-overlapping).
- `Memmove` - copy n bytes (may overlap).
- `Memset` - fill n bytes with value.
- `Memcmp` - compare n bytes.

## so/c/math

Math constants and functions (wraps C `<math.h>`).

Constants: `Pi`, `E`, `Inf`.

Functions:

- `Abs`, `Sqrt`, `Pow`, `Floor`, `Ceil`, `Round` - basic operations.
- `Log`, `Log2`, `Log10`, `Exp` - logarithms and exponentials.
- `Sin`, `Cos`, `Atan2` - trigonometry.
- `Fmin`, `Fmax`, `Fmod` - min, max, remainder.

## so/c/stdio

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

## so/c/stdlib

Process control, memory, and string conversion (wraps C `<stdlib.h>`).

- `Exit` - terminate the program.
- `Malloc` / `Calloc` / `Realloc` / `Free` - raw memory management.
- `Atoi` / `Atof` - string-to-number conversion.
- `Getenv` - read an environment variable.
- `ExitSuccess`, `ExitFailure` - standard exit codes.
