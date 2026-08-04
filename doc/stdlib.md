# So standard library

So provides a growing set of packages similar to Go's standard library, plus a low-level package for C interop. This document gives a general overview of what each package does and how it differs from its Go counterpart.

For the full package API, see the [package documentation](https://pkg.go.dev/solod.dev/so) or run `go doc` (for example, `go doc -all solod.dev/so/slices`).

Three language traits shape the whole library:

- **Memory is manual.** There is no garbage collector. Anything that allocates takes a `mem.Allocator`, and you free it yourself.
- **There is no reflection.** Values carry no runtime type information, so there is no `Marshal`/`Unmarshal` and no `%v`. Where Go inspects a type at runtime, So asks you to pass a function instead.
- **There is no user-space concurrency.** Concurrency is OS threads, provided by packages rather than by the language.

## Memory

[mem](https://pkg.go.dev/solod.dev/so/mem) allocates and frees single values and slices through an `Allocator` interface. It ships a system allocator backed by C `calloc`/`realloc`/`free`, an arena that bump-allocates from a fixed buffer, and a tracker that wraps another allocator and records what was allocated and freed.

[slices](https://pkg.go.dev/solod.dev/so/slices) creates, grows, sorts, and searches slices. Functions that allocate or deallocate (`Make`, `Append`, `Extend`, `Free`) take the allocator as their first argument.

[maps](https://pkg.go.dev/solod.dev/so/maps) is a generic hashmap standing in for Go's built-in `map[K]V`. It uses Robin Hood hashing and grows automatically, and like the slice functions it takes an allocator.

## Concurrency

[conc](https://pkg.go.dev/solod.dev/so/conc) covers what Go's language-level concurrency does: `Go` starts a thread, `Pool` runs tasks on a bounded set of worker threads, and `Chan[T]` is a thread-safe FIFO channel with optional buffering and send/receive timeouts. A channel carries values by copy.

[sync](https://pkg.go.dev/solod.dev/so/sync) provides mutexes, condition variables, and once-only initialization, all backed by pthreads. Unlike Go, each primitive has `Init` and `Free` methods; the zero value is not ready to use.

[sync/atomic](https://pkg.go.dev/solod.dev/so/sync/atomic) provides lock-free atomic types (`Int32`, `Int64`, `Uint32`, `Uint64`, `Bool`, `Pointer[T]`). Here the zero value is ready to use, but a value must not be copied once used.

## Strings and text

[strings](https://pkg.go.dev/solod.dev/so/strings) searches, splits, trims, joins, and cases strings, and provides a `Builder` for cheap concatenation and a `Reader` for reading a string as a stream. A smaller API than Go's, with no regexp-adjacent helpers.

[bytes](https://pkg.go.dev/solod.dev/so/bytes) is the same set of operations over byte slices, plus a growable `Buffer` and a `Reader`.

[strconv](https://pkg.go.dev/solod.dev/so/strconv) converts between strings and bools, integers, and floats.

[unicode](https://pkg.go.dev/solod.dev/so/unicode) classifies code points by category and case, and converts case. It covers letters, digits, spaces, and case, but not the wider Unicode tables (graphic characters, punctuation, symbols).

[unicode/utf8](https://pkg.go.dev/solod.dev/so/unicode/utf8) encodes and decodes UTF-8 byte sequences and counts runes. Same API as Go's.

## Input and output

[fmt](https://pkg.go.dev/solod.dev/so/fmt) formats and scans text through the `Printf`/`Sprintf`/`Fprintf` and `Scanf`/`Sscanf`/`Fscanf` families. The verbs are C's, not Go's, and without reflection there is no `%v`. `Print` and `Println` take only strings, so for anything else use the built-in `print` and `println`.

[io](https://pkg.go.dev/solod.dev/so/io) defines the `Reader`, `Writer`, and `Closer` interfaces the rest of the library builds on, along with `Copy`, `ReadAll`, and a few reader wrappers.

[bufio](https://pkg.go.dev/solod.dev/so/bufio) adds buffering to a reader or writer, and provides a `Scanner` that splits input into lines, words, bytes, runes, or a custom token.

[os](https://pkg.go.dev/solod.dev/so/os) opens, reads, and writes files, walks directories, inspects and changes file metadata, and reaches the process environment (arguments, environment variables, working directory, ids, exit). Built on POSIX APIs.

[path](https://pkg.go.dev/solod.dev/so/path) manipulates slash-separated paths lexically: join, split, clean, and shell-pattern matching.

## Encoding

[encoding](https://pkg.go.dev/solod.dev/so/encoding) holds the marshaler and unmarshaler interfaces that other packages implement.

[encoding/binary](https://pkg.go.dev/solod.dev/so/encoding/binary) reads and writes fixed-width integers in little- or big-endian order.

[encoding/hex](https://pkg.go.dev/solod.dev/so/encoding/hex) encodes and decodes hexadecimal, in place, into a string, or as a stream, and dumps data in `hexdump -C` format.

[encoding/json](https://pkg.go.dev/solod.dev/so/encoding/json) reads and writes JSON one token at a time. With no reflection there is no `Marshal`/`Unmarshal`: a `Decoder` walks the document and pulls values through typed getters, and an `Encoder` writes objects, arrays, and values with explicit begin/end calls.

## Networking

[net](https://pkg.go.dev/solod.dev/so/net) does TCP, UDP, and Unix domain socket networking: resolve an address, dial a connection, or listen for one. A small subset of Go's `net`, with no concurrent server support and no DNS beyond what the system resolver provides.

[net/netip](https://pkg.go.dev/solod.dev/so/net/netip) provides small value types for IP addresses, address-port pairs, and CIDR prefixes. IPv6 zones are numeric scope ids, not strings.

## Time

[time](https://pkg.go.dev/solod.dev/so/time) measures, formats, and parses time, and sleeps. A `Time` is always stored as UTC; instead of Go's locations, you pass a fixed UTC offset. Formatting and parsing use C strftime/strptime verbs (`%Y-%m-%d`) rather than Go's reference layout.

## Math and randomness

[math](https://pkg.go.dev/solod.dev/so/math) provides the usual mathematical constants and functions. Same API as Go's.

[math/bits](https://pkg.go.dev/solod.dev/so/math/bits) counts and manipulates bits. Same API as Go's.

[math/rand](https://pkg.go.dev/solod.dev/so/math/rand) generates pseudo-random numbers from a PCG source, either through a `Rand` you own or through top-level functions backed by a global one. Based on Go's `math/rand/v2`. Not suitable for security-sensitive work.

[crypto/crand](https://pkg.go.dev/solod.dev/so/crypto/crand) generates cryptographically secure random bytes and strings, which is what you want for keys, tokens, and identifiers.

## Program support

[errors](https://pkg.go.dev/solod.dev/so/errors) creates errors from a message. So natively supports only sentinel errors, defined once at package level. There is no error wrapping or `Is`/`As`. You can create custom error types, but if you want to return them as `error` from functions, you'll need to allocate them.

[cmp](https://pkg.go.dev/solod.dev/so/cmp) compares ordered values. Beyond Go's `Compare`, it provides `Func`, an untyped comparison function, and `FuncFor[T]`, which produces one for a given type. This is how sorting and searching work without reflection.

[flag](https://pkg.go.dev/solod.dev/so/flag) defines and parses command-line flags, either on the default set or on your own `FlagSet`.

[log/slog](https://pkg.go.dev/solod.dev/so/log/slog) does leveled, key-value logging with zero-allocation formatting. Inspired by Go's `log/slog`, but smaller: attributes are built with typed constructors like `String` and `Int`, and the only built-in handler writes text.

[runtime](https://pkg.go.dev/solod.dev/so/runtime) reports the build target (`GOOS`, `GOARCH`), the compiler version, the current source location, and the CPU count, and provides a secure random seed.

[uuid](https://pkg.go.dev/solod.dev/so/uuid) generates and parses UUIDs (RFC 9562), version 4 (random) and version 7 (time-ordered). Random components come from a cryptographically secure source.

## Testing

[testing](https://pkg.go.dev/solod.dev/so/testing) is the minimal `T` API for So tests and benchmarks, which live in a package's `test` and `bench` subdirectories and are run with `so test` and `so bench`. See the [testing guide](testing.md).

## C interop

[c](https://pkg.go.dev/solod.dev/so/c) is the escape hatch into C: types for C's numeric and character types, conversions between So and C strings, slices, and pointers, stack allocation, size and alignment queries, and a way to emit raw C. See the [interop guide](interop.md).
