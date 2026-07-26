# So standard library

Solod provides a growing set of high-level packages similar to Go's stdlib, and a low-level package to help with C interop. For full API details, see the [package documentation](https://pkg.go.dev/solod.dev/so).

[bufio](#sobufio) •
[bytes](#sobytes) •
[c](#soc) •
[cmp](#socmp) •
[conc](#soconc) •
[crypto/crand](#socryptocrand) •
[encoding/binary](#soencodingbinary) •
[encoding/hex](#soencodinghex) •
[encoding/json](#soencodingjson) •
[errors](#soerrors) •
[flag](#soflag) •
[fmt](#sofmt) •
[io](#soio) •
[log/slog](#sologslog) •
[maps](#somaps) •
[math](#somath) •
[math/bits](#somathbits) •
[math/rand](#somathrand) •
[mem](#somem) •
[net](#sonet) •
[net/netip](#sonetnetip) •
[os](#soos) •
[path](#sopath) •
[runtime](#soruntime) •
[slices](#soslices) •
[strconv](#sostrconv) •
[strings](#sostrings) •
[sync](#sosync) •
[sync/atomic](#sosyncatomic) •
[testing](#sotesting) •
[time](#sotime) •
[unicode](#sounicode) •
[unicode/utf8](#sounicodeutf8) •
[uuid](#souuid)

## [so/bufio](https://pkg.go.dev/solod.dev/so/bufio)

Buffered I/O. Wraps an `io.Reader` or `io.Writer` with buffering and helpers for textual I/O. Based on Go's `bufio` package.

Functions:

- `NewReader` and `NewReaderSize` create a buffered reader.
- `NewWriter` and `NewWriterSize` create a buffered writer.
- `NewReadWriter` combines a Reader and Writer into a single `io.ReadWriter`.
- `NewScanner` creates a scanner for token-based reading (lines, words, bytes, or custom split functions).
- `ScanLines`, `ScanWords`, `ScanBytes` and `ScanRunes` are built-in split functions for `Scanner`.

Types:

- `Reader` wraps an `io.Reader` with buffering.
- `Writer` wraps an `io.Writer` with buffering.
- `ReadWriter` combines a Reader and Writer.
- `Scanner` reads tokens from an `io.Reader` using a `SplitFunc`.

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

## [so/c](https://pkg.go.dev/solod.dev/so/c)

Low-level C interop helpers for pointers, strings, and type information.

Functions:

- `Alignof` and `Sizeof` return the alignment and size of type T.
- `Alloca` allocates an array on the stack.
- `Assert` panics with a message if a condition is false.
- `Bytes`, `Slice` and `String` wrap C pointers to So types.
- `CString` converts a So string to a null-terminated C string.
- `PtrAdd`, `PtrAs` and `PtrAt` manipulate pointers.
- `Val` and `Raw` emit raw C code.
- `Zero` returns the zero value of type T.

Types:

- `Char`, and `ConstChar` represent a C `char` type.
- `Int`, `UInt`, `Long`, `ULong`, etc. represent numeric C types.
- `Size`, `SSize`, `Ptrdiff` and `Intptr` represent C types whose width follows the target.

## [so/cmp](https://pkg.go.dev/solod.dev/so/cmp)

Comparing ordered values. Based on Go's `cmp` package.

Functions:

- `Compare` returns -1, 0, or +1 for two ordered values.
- `Equal` reports whether two comparable values are equal.
- `Less` reports whether x is less than y.

Types:

- `Func` is a comparison function `func(a, b any) int`.
- `FuncFor` returns the appropriate comparison function for type T.

## [so/conc](https://pkg.go.dev/solod.dev/so/conc)

Basic primitives for concurrent programming, backed by pthreads.
Meant to be used in place of language-level concurrency features.

`Chan[T]` is a thread-safe FIFO channel, similar to Go's built-in `chan T`. It carries values by copy: a sender copies a value into the channel and a receiver copies one out.

- `NewChan[T]` creates a new channel, either buffered or unbuffered (rendezvous).
- `Chan.Send` copies a value into the channel (blocks until delivered).
- `Chan.Recv` copies a value out of the channel (blocks until a value or close).
- `Chan.SendTimeout` sends with a deadline, returning a status; a zero duration makes it non-blocking.
- `Chan.RecvTimeout` receives with a deadline, returning the value and a status; a zero duration makes it non-blocking.
- `Chan.Close` closes the channel.
- `Chan.Free` releases the channel's resources.

`Thread` is a handle to a single OS thread running a `func(any) any`:

- `Go` and `GoWith` launch a thread and return a handle to it.
- `Thread.Wait` blocks until the thread terminates.
- `Thread.Detach` hands the thread's resources to the runtime.

`Pool` is a bounded pool of worker threads for tasks of type `func(any)`:

- `NewPool` creates a pool of workers and starts them.
- `Pool.Go` submits a task for execution.
- `Pool.Wait` blocks until all submitted tasks finish; the pool stays usable afterward.
- `Pool.Free` drains queued tasks, joins the workers, and releases the pool.

## [so/crypto/crand](https://pkg.go.dev/solod.dev/so/crypto/crand)

Cryptographically secure random number generation.

- `Read` fills a slice with cryptographically secure random bytes.
- `Reader` is a global instance of a cryptographically secure RNG.
- `Text` returns a cryptographically random string using the base32 alphabet.

## [so/encoding/binary](https://pkg.go.dev/solod.dev/so/encoding/binary)

Translation between numbers and byte sequences. Based on Go's `encoding/binary` package.

Types:

- `ByteOrder` specifies how to convert byte slices into unsigned integers.
- `AppendByteOrder` specifies how to append unsigned integers into a byte slice.
- `LittleEndian` and `BigEndian` implement `ByteOrder` and `AppendByteOrder`.

## [so/encoding/hex](https://pkg.go.dev/solod.dev/so/encoding/hex)

Hexadecimal encoding and decoding. Based on Go's `encoding/hex` package.

Functions:

- `EncodedLen` and `DecodedLen` return the encoded/decoded length for n bytes.
- `Encode` and `Decode` encode to or decode from hexadecimal in place.
- `AppendEncode` and `AppendDecode` append the encoded/decoded form to a slice.
- `EncodeToString` and `DecodeString` convert between a byte slice and a hexadecimal string.
- `Dump` returns a hex dump of data in `hexdump -C` format.
- `NewEncoder` and `NewDecoder` wrap an `io.Writer`/`io.Reader` for streaming encoding/decoding.
- `NewDumper` returns a writer that hex-dumps everything written to it.

Types:

- `Encoder` writes hexadecimal characters to an underlying `io.Writer`.
- `Decoder` reads and decodes hexadecimal characters from an underlying `io.Reader`.
- `Dumper` writes a hex dump of all data written to it.

## [so/encoding/json](https://pkg.go.dev/solod.dev/so/encoding/json)

Token-level JSON encoding and decoding. With no reflection, there is no
`Marshal`/`Unmarshal`; you read and write one token at a time.

Types:

- `Decoder` walks a document with `Next`/`Kind` and pulls values through typed
  getters (`Str`, `Int`, `Float`, `Bool`).
- `Encoder` writes a document with `BeginObject`/`EndObject`, `BeginArray`/`EndArray`,
  and the value tokens, streaming to an `io.Writer`.

## [so/errors](https://pkg.go.dev/solod.dev/so/errors)

Error creation from text messages.

- `New(text string) error` - create a new error with the given message.

So only supports sentinel errors, which are defined at the package level using `New`.

## [so/flag](https://pkg.go.dev/solod.dev/so/flag)

Command-line flag parsing. Based on Go's `flag` package.

Functions:

- `BoolVar`, `IntVar`, `UintVar`, `Float64Var` and `StringVar` define typed flags.
- `Var` defines a flag with a custom `Value` implementation.
- `Parse` parses command-line flags from `os.Args`.
- `Args` returns the non-flag command-line arguments after parsing.

Types:

- `FlagSet` represents a set of defined flags with its own error handling and output.
- `Flag` represents a single flag.
- `Value` is the interface for custom flag values.

## [so/fmt](https://pkg.go.dev/solod.dev/so/fmt)

Formatted I/O with functions analogous to C's printf and scanf. Uses C format verbs (not Go verbs).

- `Print` and `Println` write strings to standard output.
- `Printf` formats and writes to standard output.
- `Sprintf` formats and writes to a string.
- `Fprintf` formats and writes to an `io.Writer`.
- `Scanf` scans formatted text from standard input.
- `Sscanf` scans formatted text from a string.
- `Fscanf` scans formatted text from an `io.Reader`.

Since `Print` and `Println` only take string arguments, you'll usually want to use the built-in functions `print` and `println` instead.

## [so/io](https://pkg.go.dev/solod.dev/so/io)

Basic interfaces to I/O primitives. Offers an API similar to Go's `io` package, but with fewer features.

Functions:

- `Copy` and `CopyN` copy data from a reader to a writer.
- `ReadAll` and `ReadFull` read data from a reader.

Types:

- `Reader`, `Writer`, and `Closer` are basic concepts for anything that does I/O.
- `LimitedReader` and `SectionReader` implement specialized readerss.
- `Discard` is a no-op writer.

## [so/log/slog](https://pkg.go.dev/solod.dev/so/log/slog)

Simplified structured logging, inspired by Go's `log/slog` package. Provides leveled, key-value logging with zero-allocation formatting.

Functions:

- `New` creates a Logger from a Handler.
- `SetDefault` and `Default` manage the default logger.
- `Debug`, `Info`, `Warn`, `Error` log at the corresponding level using the default logger.
- `String`, `Int`, `Float64`, `Bool`, `Time` and `Duration` create key-value attributes.

Types:

- `Logger` logs messages through a `Handler`.
- `TextHandler` formats log records as `timestamp LEVEL message key=value ...` lines.
- `Record` holds a log event (time, level, message, attributes).
- `Attr` is a key-value pair. `Value` is a tagged union holding the value.
- `Level` represents severity.

## [so/maps](https://pkg.go.dev/solod.dev/so/maps)

Generic hashmap similar to Go's built-in `map[K]V`, backed by a Robin Hood hash table with automatic grow.

Functions:

- `New` creates a new `Map` with a given allocator.

Types:

- `Map` is a generic hashmap with `Get`, `Set`, and `Delete` methods.
- `Iter` is an iterator over a map's key-value pairs.

## [so/math](https://pkg.go.dev/solod.dev/so/math)

Mathematical functions and constants. Offers the same API as Go's `math` package.

## [so/math/bits](https://pkg.go.dev/solod.dev/so/math/bits)

Bit counting and manipulation functions. Offers the same API as Go's `math/bits` package.

## [so/math/rand](https://pkg.go.dev/solod.dev/so/math/rand)

Pseudo-random number generation. Based on Go's `math/rand/v2` package.

Top-level functions use a global `Rand` with a `PCG` source seeded by `runtime.Seed`.

Functions:

- `Int`, `Int32`, `Int64` return non-negative pseudo-random integers.
- `Uint`, `Uint32`, `Uint64` return pseudo-random unsigned integers.
- `IntN`, `Int32N`, `Int64N` return non-negative pseudo-random integers in [0,n).
- `UintN`, `Uint32N`, `Uint64N` return pseudo-random unsigned integers in [0,n).
- `Float32` and `Float64` return pseudo-random floats in [0.0,1.0).

Types:

- `Source` is an interface for a source of pseudo-random `uint64` values.
- `PCG` is a PCG generator with 128 bits of internal state. Implements `Source`.
- `Rand` wraps a `Source` and provides the same methods as the top-level functions.

## [so/mem](https://pkg.go.dev/solod.dev/so/mem)

Memory allocation with a pluggable allocator interface.

Functions:

- `Alloc` / `Free` - allocate/free a single value.
- `AllocSlice` / `FreeSlice` - allocate/free a slice.

Types:

- `Allocator` interface - custom allocator support (`Alloc`, `Realloc`, `Free`).
- `SystemAllocator` - default allocator backed by C `calloc`/`realloc`/`free`.
- `Arena` - bump allocator backed by a fixed buffer (`Alloc`, `Realloc`, `Reset`).
- `Tracker` - tracks all allocations and deallocations made through an allocator.
- `Stats` - records statistics about the memory allocator.

## [so/net](https://pkg.go.dev/solod.dev/so/net)

Basic TCP, UDP, and Unix domain socket networking. There is no concurrent server support.

Functions:

- `ResolveTCPAddr` resolves a `host:port` string to a `TCPAddr`; `ResolveUDPAddr` does the same for a `UDPAddr`; `ResolveUnixAddr` carries a socket path into a `UnixAddr`.
- `DialTCP` connects to a TCP address; `ListenTCP` announces on a local one.
- `DialUDP` creates a connected UDP socket (fixed peer, `Read`/`Write`); `ListenUDP` creates an unconnected one (any peer, `ReadFrom`/`WriteTo`).
- `DialUnix` connects to a Unix socket; `ListenUnix` announces a stream listener; `ListenUnixgram` binds an unconnected datagram socket.
- `SplitHostPort` splits a `host:port` address into a `HostPort`; `JoinHostPort` does the reverse into a caller buffer.

Types:

- `TCPAddr` is the address of a TCP endpoint.
- `TCPConn` is a TCP connection; implements `io.Reader` and `io.Writer`.
- `TCPListener` is a TCP listener; its `Accept` method returns the next `TCPConn`.
- `UDPAddr` is the address of a UDP endpoint.
- `UDPConn` is a UDP socket; connected (`Read`/`Write`) or unconnected (`ReadFrom`/`WriteTo`).
- `UnixAddr` is the path of a Unix domain socket endpoint.
- `UnixConn` is a Unix socket; a connected stream/datagram (`Read`/`Write`) or an unconnected datagram (`ReadFrom`/`WriteTo`).
- `UnixListener` is a Unix stream listener; its `Accept` method returns the next `UnixConn`.

## [so/net/netip](https://pkg.go.dev/solod.dev/so/net/netip)

Small value types for IP addresses, address-port pairs, and CIDR prefixes. IPv6 zones are stored as numeric scope IDs, not as strings.

Functions:

- `AddrFrom4` and `AddrFrom16` create an `Addr` from a fixed-size byte array.
- `AddrFromSlice` creates an `Addr` from a 4- or 16-byte slice.
- `ParseAddr` and `MustParseAddr` parse an IP address from a string.
- `AddrPortFrom` creates an `AddrPort` from an `Addr` and port.
- `ParseAddrPort` and `MustParseAddrPort` parse an address-port pair from a string.
- `PrefixFrom` creates a `Prefix` from an `Addr` and bit length.
- `ParsePrefix` and `MustParsePrefix` parse a CIDR prefix from a string.

Types:

- `Addr` is an IPv4 or IPv6 address.
- `AddrPort` is an IP address and port number.
- `Prefix` is a CIDR prefix.

## [so/os](https://pkg.go.dev/solod.dev/so/os)

File I/O and filesystem operations. Offers an API similar to Go's `os` package, built on POSIX APIs.

Functions:

- `Create`, `Open`, `OpenFile` open files for reading and/or writing.
- `ReadFile` and `WriteFile` read or write an entire file.
- `ReadDir` reads a directory and returns its entries.
- `Stat` and `Lstat` return file information.
- `Chmod`, `Chown`, `Lchown`, `Chtimes` change file attributes.
- `Rename` renames (moves) a file.
- `Remove` removes a file or empty directory.
- `Mkdir` creates a directory.
- `TempDir`, `CreateTemp`, `MkdirTemp` work with temporary files and directories.
- `Getwd` and `Chdir` manage the current working directory.
- `Getenv`, `LookupEnv`, `Setenv` and `Unsetenv` manage environment variables.
- `Getpid`, `Getppid`, `Getuid`, `Geteuid`, `Getgid`, `Getegid` return process/user info.
- `Exit` terminates the program with the given status code.

Variables:

- `Args` holds the command-line arguments, starting with the program name.

Types:

- `File` represents an open file with methods for reading and writing data.
- `FileInfo` describes a file (returned by `Stat` and `Lstat`).
- `FileMode` represents a file's mode and permission bits.
- `DirEntry` describes an entry in a directory (returned by `ReadDir`).

## [so/path](https://pkg.go.dev/solod.dev/so/path)

Utility routines for manipulating slash-separated paths. Based on Go's `path` package.

- `Base` returns the last element of a path.
- `Clean` returns the shortest equivalent path by lexical processing.
- `Dir` returns all but the last element of a path.
- `Ext` returns the file name extension used by a path.
- `IsAbs` reports whether a path is absolute.
- `Join` joins path elements into a single path.
- `Match` reports whether a name matches a shell pattern.
- `Split` splits a path into directory and file components.

## [so/runtime](https://pkg.go.dev/solod.dev/so/runtime)

Information about the environment where the program was compiled, and runtime utilities.

- `GOOS` and `GOARCH` specify the target operating system and architecture.
- `FileName`, `Line` and `FuncName` report the current source location.
- `NumCPU` NumCPU returns the number of logical CPUs usable by the program.
- `Seed` returns a cryptographically secure random 64-bit seed.
- `Version` returns So's compiler version (git commit hash or tag).

## [so/slices](https://pkg.go.dev/solod.dev/so/slices)

Operations on slices:

- `Make` and `MakeCap` allocate a slice, `Free` deallocates it.
- `Append` appends elements to a heap slice, growing if needed.
- `Extend` appends another slice to a heap slice, growing if needed.
- `Clone` creates a shallow copy of the slice.
- `Equal` reports whether two slices are equal.
- `Contains` and `Index` search for value in a slice.
- `Min`, `MinFunc`, `Max` and `MaxFunc` return the minimum/maximum element.
- `Sort`, `SortFunc` and `SortStableFunc` sort slices.

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

## [so/sync](https://pkg.go.dev/solod.dev/so/sync)

Basic synchronization primitives, backed by pthreads.

`Mutex` is a mutual exclusion lock:

- `Mutex.Init` prepares the mutex for use, leaving it unlocked.
- `Mutex.Lock` and `Mutex.Unlock` acquire and release the lock.
- `Mutex.TryLock` tries to acquire the lock and reports whether it succeeded.
- `Mutex.Free` releases the mutex's resources.

`Cond` is a condition variable tied to a `*Mutex`:

- `Cond.Init` prepares the condition variable, guarded by the given mutex.
- `Cond.Wait` atomically unlocks the mutex and blocks until signaled, then re-locks.
- `Cond.WaitFor` waits like `Cond.Wait` but gives up after a given duration.
- `Cond.Signal` and `Cond.Broadcast` wake one or all waiting threads.
- `Cond.Free` releases the condition variable's resources.

`Once` runs a function exactly once, even when called concurrently:

- `Once.Init` prepares the once for use.
- `Once.Do` runs the given function on the first call only.
- `Once.Free` releases the once's resources.

## [so/sync/atomic](https://pkg.go.dev/solod.dev/so/sync/atomic)

Lock-free atomic operations. Each type's zero value is ready to use,
and must not be copied after first use.

`Int32`, `Int64`, `Uint32`, `Uint64`, `Bool`, and `Pointer[T]` wrap a single value:

- `Load` and `Store` atomically read and write the value.
- `Swap` stores a new value and returns the previous one.
- `CompareAndSwap` sets a new value only if the current one matches, reporting whether it did.
- `Add` (numeric types only) adds a delta and returns the new value.

## [so/testing](https://pkg.go.dev/solod.dev/so/testing)

Minimal testing support, mirroring Go's `testing` package. Tests live in a package's `test` subdirectory and are run with the `so test` command; benchmarks live in a `bench` subdirectory and are run with `so bench`. See the [testing guide](testing.md) for details.

Functions:

- `RunTests` runs a list of tests, prints per-test results, and exits non-zero on failure. It is called from the generated test runner.

Types:

- `T` is passed to each test to record failure and skip state.
- `Test` pairs a test name with its function.

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
- `Sleep` pauses the current thread for at least the given duration.

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

## [so/uuid](https://pkg.go.dev/solod.dev/so/uuid)

Generating and manipulating universally unique identifiers (UUIDs), as specified in RFC 9562. Random components are generated with a cryptographically secure RNG.

Constants:

- `UUIDLen` - length of a canonical UUID string (36).

Functions:

- `New` returns a new UUID using an algorithm suitable for most purposes (currently `NewV4`).
- `NewV4` returns a random version 4 UUID.
- `NewV7` returns a time-ordered version 7 UUID.
- `Nil` and `Max` return the Nil and Max UUIDs.
- `Parse` and `MustParse` parse a UUID from a string.

Types:

- `UUID` is a 128-bit universally unique identifier.
