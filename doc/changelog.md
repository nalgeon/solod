# Changelog

This document outlines the main changes in different So versions.

## Solod 0.4 (in progress)

### Language

**Type parameters are rejected where C has nothing to emit**. A type parameter exists for Go type-checking and as a macro argument, so it may only appear inside an `so:inline` macro. Declarations that would have emitted a bare `T` into C are now errors instead of invalid output:

```go
type Pair[T any] struct{ a, b T } // rejected: no C type for T

type Stack[T any] struct{ items []T } // OK: so_Slice is type-erased
```

A generic function or method must be `so:inline` or `so:extern`, even when its signature never mentions the type parameter. See the [generics guide](./generics.md) for details.

[9501bae](https://github.com/solod-dev/solod/commit/9501baef6386471516b70c22f64ec69b442d008f)

**Constraint interfaces are no longer emitted**. Go allows such an interface only as a type parameter constraint, never as a value type, so it has nothing to represent in C.

```go
type Number interface{ ~int | ~float64 } // not emitted
```

**Interface methods with value receivers are rejected**. If a concrete type uses value receivers, converting it to an interface will fail:

```go
func (r Rect) Area() int { // value receiver
    return r.width * r.height
}

var s Shape = r // rejected: method Rect.Area has a value receiver
```

Using pointer receivers (`func (r *Rect) Area() int`) works fine.

**Type embedding is rejected**. Both struct and interface embedding are errors now:

```go
type number struct {
    base // rejected: embedded field
}

type readWriter interface {
    reader // rejected: embedded interface
    write(v int)
}
```

Declare a named field or list the methods instead.

**Switch tag evaluation**. A switch translates to an if/else-if chain, which repeats the tag in every comparison. The tag now goes into a temporary variable first, so a call in the tag runs once, as in Go:

```go
switch inc() { // used to call inc() once per case
case 2:
    println("two")
default:
    println("other")
}
```

Switching on a struct or array is not supported.

[06b8ab8]https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

**Switch case bodies reject `break` and `fallthrough`**. Both used to be emitted as-is, producing C that either lead to incorrect behavior or did not compile at all. Both are errors now.

[fd8659b](https://github.com/solod-dev/solod/commit/fd8659b66bbb2e24f2b70062c5b08b84abffe551)

**Interface comparison**. Two interfaces are equal when they hold the same pointer, and an interface compares with `nil` as expected. Comparing an interface with a concrete type is not supported:

```go
r := Rect{2, 4}
var s Shape = &r
if s == nil { }   // supported
if s == other { } // supported, other is a Shape
if s == &r { }    // not supported
```

An `any` compares with `nil`, with a pointer, or with another `any`. Comparing it with a value is not supported, because an `any` holds the address of the value:

```go
var a any = n
if a == nil { }  // supported
if a == &n { }   // supported, &n is a pointer
if a == n { }    // not supported
```

[06b8ab8]https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

**String and character literals**. A literal is decoded to its value and re-encoded for C instead of being copied into the output as written. A string keeps printable ASCII and UTF-8 as is and escapes other bytes in octal. A byte or rune literal stays a character literal only for printable ASCII, and becomes a number otherwise:

```text
"日本語"   ->  "日本語"
"a\xffb"  ->  "a\377b"
'a'       ->  'a'
'世'      ->  0x4e16
```

**Integer literals above MaxInt64**. Such a literal now gets a `u` suffix in C. C gives an unsuffixed decimal literal a signed type, so without the suffix the compiler warned that the value does not fit:

```go
var n uint64 = 18446744073709551615 // -> 18446744073709551615u
```

**Untyped integer constants above MaxInt64**. Such a constant is now declared as `uint64_t` instead of `int64_t`:

```go
const maxUint64 = 18446744073709551615
var third uint64 = maxUint64 / 3 // was 0, now 6148914691236517205
```

A constant that does not fit `uint64_t` either is rejected:

```go
const huge = 1 << 200 // rejected: constant huge does not fit in int64 or uint64
```

**Float constant expressions**. A constant float expression is emitted as its value, not as the operators used to produce it:

```go
const pi = 3.14159
const twoPi = 2 * pi
```

```c
static const so_unused double pi = 3.14159;
static const so_unused double twoPi = 6.28318;
```

A constant that does not fit the C float type is rejected:

```go
const huge = 1e200 * 1e200 // rejected: constant 1e+400 overflows float64
```

**Float32 literals**. Such a literal now gets an `f` suffix in C. Without it the literal is a `double`, and C promotes the other operand to match instead of computing in `float` as Go does:

```go
var x float32 = 0.1
_ = x == 0.1 // was false, now true
```

**Integer literals in a float context**. Such a literal is now emitted as a float literal. C has no integer type wide enough for the larger ones:

```go
var x float64 = 10000000000000000000000 // was 10000000000000000000000, now 1e+22
```

### Interop

**C field name override**. A `c:"..."` struct tag sets the C name of a field, so an extern struct can match a C header field whose name is a Go keyword:

```go
//so:extern SDL_CommonEvent
type SDL_CommonEvent struct {
    etype uint32 `c:"type"` // emitted as .type in C
}
```

[fc25bb8]https://github.com/solod-dev/solod/commit/fc25bb875b10e4d64a6f223ad3f5ec647ed2733e)

**Target-width C types**. `so/c` now supports more common C types:

```text
size_t      - c.Size
ssize_t     - c.SSize
ptrdiff_t   - c.Ptrdiff
intptr_t    - c.Intptr
long double - c.LongDouble
```

[c06b294]https://github.com/solod-dev/solod/commit/c06b29494cda8c822a4191843120fb6a6091a25d)

### Safety

**Assertions are no longer tied to NDEBUG**. They are on by default and removed only by the new `-assert` flag:

```sh
so build -assert=off .
```

Previously, `NDEBUG` removed these checks. This meant that a C project that defines it could accidentally turn So's safety net off.

[f67d8ca]https://github.com/solod-dev/solod/commit/f67d8ca3660437d6f648dd905d7784a5fbb180fa)

## Solod 0.3

### Language

**Type declaration order**. Types are now emitted in dependency order, so a type can use a type declared later in the source:

```go
type Rect struct {
    Min, Max Point // Point is declared below
}

type Point struct {
    X, Y int
}
```

[dd21515](https://github.com/solod-dev/solod/commit/dd21515ce358c3af0d815ac30178af63226d3e5a)

**Promoted symbols**. The `//so:promote` directive promotes an unexported type, function, method, var, or const into the package header with the package prefix. This lets an exported inline function or type reference an unexported one without exporting it and polluting the public API:

```go
type Stats struct { c counter }

//so:inline
func NewStats() Stats {
	return Stats{c: newCounter()}
}

//so:promote
type counter struct { val int }

//so:promote
func newCounter() counter { ... }
```

[4dc1364](https://github.com/solod-dev/solod/commit/4dc1364488e125f2e96b57d4cffcdce32aa8805a)

**Anonymous functions**. You can now use anonymous functions as variable types and function parameters:

```go
// func parameter
func apply(n int, f func(int) int) int { return f(n) }

// func variable
var fn func(int) int = calc
```

[d926b2a](https://github.com/solod-dev/solod/commit/d926b2a56f5f56fd22a9b580a929cb5159236b0c)

### Safety

**Escape analysis**. If a function tries to return a stack-allocated value, the program won't compile:

```go
type Point struct{ x, y int }

func newPoint(x, y int) *Point {
    return &Point{x: x, y: y}
	//     |
	// compile-time error: stack-allocated
	// value escapes function frame
}

func main() {
    p := newPoint(3, 4)
    println(p.x, p.y)
}
```

The check isn't thorough and only covers a few common cases.

[3f2a2cc](https://github.com/solod-dev/solod/commit/3f2a2cc0afc13c2fc1eb66cbaa34c50581f72a97)

**Diagnosable assertions**. Assertions (slice bounds, index out of range, slice-to-array length, ...) and `c.Assert` now panic instead of calling C's `assert`. They go through a single `so_assert` macro, so a failure reports the calling function and honors `-panic=trace`:

```text
panic: index out of range
  main.c:10 (func boundsFail)
0   app    0x00000001045f7264 boundsFail + 160
1   app    0x00000001045f716c main + 448
```

Defining `NDEBUG` removes assertions. Other runtime checks, like calling `append` beyond capacity, still panic.

[2f86ef3](https://github.com/solod-dev/solod/commit/2f86ef33f739f4d211ea33faf8bf517e9cf37b3b)

**Stack traces**. The `panic` flag controls how a panic terminates the program: `trace` (default) prints a stack trace before exiting, `exit` calls `exit(1)`, and `abort` raises `SIGABRT` for a debugger or core dump. The default fits glibc and macOS; on musl (empty trace) or freestanding (always traps) pass `-panic=exit` or `-panic=abort`. See [building](building.md).

[8ed7f48](https://github.com/solod-dev/solod/commit/8ed7f48e66d2632d55309f810072617ee22b80ac)

**Nil checks**. A nil pointer dereference (or other invalid memory access) is now caught at runtime in POSIX hosted builds and reported as a panic that honors `-panic`, rather than emitting a per-dereference check that clutters the generated C.

⚠️ This removes the `-check-nil` flag.

[b829c4f](https://github.com/solod-dev/solod/commit/b829c4f77bb2b8880ea5a3a96caf5a273c99e1b1)

**Divide-by-zero checks**. Integer division or modulo by a zero divisor now panics instead of relying on hardware. This closes a portability gap: division by zero is undefined in C, and on arm64 it silently yields 0 rather than trapping.

[af173d8](https://github.com/solod-dev/solod/commit/af173d8e7975327a0994a4bc8eb982a3aa24995d)

**Sanitizer flag**. The `-sanitize` flag turns on C sanitizers for `build` and other commands. Bare `-sanitize` enables `address,undefined`; a comma-separated list picks a specific set. See [building](building.md).

[f87f8f4](https://github.com/solod-dev/solod/commit/f87f8f4abc421661fefc0d5cdd8b25d43025939a)

**Reserved names**. Local variables and parameters whose names conflict with C keywords or macros (`long`, `bool`, ...) are now mangled automatically instead of producing invalid C. Reserved names as struct fields or package-level declarations are rejected instead.

[7f1bb70](https://github.com/solod-dev/solod/commit/7f1bb702ebb28e0fb6d941e428deb3d476eb7188)

### Standard library

**Concurrency tools**. The `conc` package provides basic tools for concurrent programming, backed by pthreads.

- `Chan[T]` — a thread-safe FIFO channel (buffered) or rendezvous (unbuffered).
- `Pool` — a bounded worker pool for fork-join parallelism.
- `Thread` — an operating system thread.

**Synchronization primitives**. The `sync` package provides basic synchronization primitives, backed by pthreads.

- `Cond` — a condition variable.
- `Mutex` — a mutual exclusion lock.
- `Once` — runs a function exactly once.

[22e7e78](https://github.com/solod-dev/solod/commit/22e7e782cb3edc56789c08306e08e6f71739fddf) ·
[f5ae958](https://github.com/solod-dev/solod/commit/f5ae958ba9ee34a135dc006f6d0b30063a3d1479)

**Atomic types**. The `sync/atomic` package provides lock-free atomic operations. Offers atomic values like `Int64`, `Uint64`, `Bool`, and `Pointer[T]`.

[71a49a0](https://github.com/solod-dev/solod/commit/71a49a0413622a4cc5a1f53e88d98fbcaceb3496)

**Low-level JSON API**. The `encoding/json` package provides token-level JSON encoding and decoding. With no reflection, there is no `Marshal`/`Unmarshal`; you read and write one token at a time using `Decoder` and `Encoder`. Both types support streaming and use minimal allocations.

[54161e2](https://github.com/solod-dev/solod/commit/54161e27f5d7bf08a80cc50d21bb1c063c8ccf23)

**Source locations**. The `runtime` package provides `FileName`, `Line` and `FuncName` to report the current source location.

[7f0a8c0](https://github.com/solod-dev/solod/commit/7f0a8c06483e133b499e207cceee0e866cdf404b)

**Test leak checking**. `testing.T` provides an `Allocator` method that returns a tracking allocator. A test fails if memory allocated through it is not freed by the time the test returns. See the [testing guide](testing.md).

[9440bf5](https://github.com/solod-dev/solod/commit/9440bf5ec011f0469aa0ff4d14386588304b499a)

### Tools

**Tests**. The `so test` command runs tests from a package's `test` subdirectory. It discovers `TestXxx(t *testing.T)` functions, generates a runner that dispatches them via `testing.RunTests`, and runs them. See the [testing guide](testing.md).

[ca13759](https://github.com/solod-dev/solod/commit/ca1375959866ca7fc7c0b38b60f5b84fd085e6bc) ·
[163afcb](https://github.com/solod-dev/solod/commit/163afcb8662359cb6b93a4043f4572b4bec64d7b)

**Benchmarks**. The `so bench` command runs benchmarks from a package's `bench` subdirectory, mirroring `so test`. It discovers `BenchmarkXxx(b *testing.B)` functions, generates a runner that dispatches them via `testing.RunBenchmarks` with the system allocator, and runs them. See the [testing guide](testing.md).

[c374069](https://github.com/solod-dev/solod/commit/c37406988e712a7f266c87990107d6f491a566f7) ·
[e042c23](https://github.com/solod-dev/solod/commit/e042c233b77c32d4e6e05e7dfbea8a9661e598c1) ·
[2bc8dd9](https://github.com/solod-dev/solod/commit/2bc8dd90ce321b6bd36f9ac9cb13350021bc0e3f)

**Automatic linking**. The `so:link` directive declares which C library a package requires. `so build` and other commands collect these across all transpiled packages and pass them to the C compiler. The standard library already uses `so:link`, so importing `so/math` links `-lm` and `so/sync` or `so/conc` links `-lpthread` without setting `LDFLAGS` by hand. See the [interop guide](interop.md#linking).

[4cb27cd](https://github.com/solod-dev/solod/commit/4cb27cd4eed149348c84a9a01eff9df7c0e5d67f)

## Solod 0.2

Networking, new targets, and friendlier interop.

### Language

New directives: `so:volatile`, `so:thread_local`, `so:attr`.<br>
[600e881](https://github.com/solod-dev/solod/commit/600e881fe72cf5f9857745b489c6dedf9a864ea3)

Implement `error` as a regular interface (it was special-cased before).<br>
[6c8f0bd](https://github.com/solod-dev/solod/commit/6c8f0bd68e4ba8693d22be59f763676889270070)

Type aliases.<br>
[deeccb9](https://github.com/solod-dev/solod/commit/deeccb98d22f342e6bdceb9d7827e9d464af9603)

⚠️ Auto-generate result types for `(T, error)` multi-return values where `T` is a custom struct type. This is a breaking change: previously, you had to manually define a `T{Result}` type for any such `T`.<br>
[745b174](https://github.com/solod-dev/solod/commit/745b174e11c08bee91cdaeaf8a8b2aa083863b61)

⚠️ Block-scoped `defer` is no longer supported. This is a breaking change.<br>
[fb49cca](https://github.com/solod-dev/solod/commit/fb49ccab2316815308f690f2690e1c3bf19ee59b)

### Standard library

`encoding/hex` package.<br>
[42a5cf0](https://github.com/solod-dev/solod/commit/42a5cf0d7f6f08ba9e862bd5a738cf74448b2711)

`net` package.<br>
[ee89acb](https://github.com/solod-dev/solod/commit/ee89acb4ab1e1a05861c186ca5b8fc588a9ec268)
[658ce66](https://github.com/solod-dev/solod/commit/658ce6693272f85641f4b5f6228e78b40f322fd8)
[2ec017b](https://github.com/solod-dev/solod/commit/2ec017bed64e611aa2e975a3aa6bee9d9b2bcc89)

`net/netip` package.<br>
[5f87292](https://github.com/solod-dev/solod/commit/5f87292ac5b334cd1080a20ca75cc5a1c2c3ea59)

`uuid` package.<br>
[fc8f2fa](https://github.com/solod-dev/solod/commit/fc8f2fabdac147f576fcede45bd178b313a7e25a)

Numeric C types in the `so/c` package for better interop.<br>
[5914a75](https://github.com/solod-dev/solod/commit/5914a7591bc44335b4556893a3f848e1a6c9cc8c)

`mem/Arena.Free` reclaims the last allocation if the pointer matches.<br>
[f9adba6](https://github.com/solod-dev/solod/commit/f9adba6baca67b2ed332e3aaaf1e59b44113d1db)

### Tools

Support for third-party packages (`go get` or vendoring) and multi-module projects.<br>
[bba8265](https://github.com/solod-dev/solod/commit/bba8265883b10814803510518693b224b70d2d98)

Report So source location (file and line number) instead of C source location when panicking (`track-source` flag, off by default).<br>
[fb78b7a](https://github.com/solod-dev/solod/commit/fb78b7af20525055e320e5f01cb5bb8198ab18ff)

Check for nil pointer dereference on struct pointer field access and interface method calls (`check-nil` flag, off by default).<br>
[426961e](https://github.com/solod-dev/solod/commit/426961e0ef463cc2390e6d1a930555f2db581f7e)

### Targets

32-bit target support.<br>
[deac815](https://github.com/solod-dev/solod/commit/deac815a5100f119765ffcf8b5961ef579c7a766)
[de30cde](https://github.com/solod-dev/solod/commit/de30cdec169be0f7f8835853ccde5f78e3e4c233)

WebAssembly support (WASI).<br>
[3d0791b](https://github.com/solod-dev/solod/commit/3d0791b69e8fd5053fd508dbbb8c9cebfb0b3ff7)

Freestanding mode (no libc dependency).<br>
[1cfc8c7](https://github.com/solod-dev/solod/commit/1cfc8c7cd602a379332e6c128ebd2bde007c9a63)
