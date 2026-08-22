# Solod 0.4 (in progress)

This document lists the main changes in the So version in development.

- Language:
  [Type parameters](#type-parameters) ·
  [Constraint interfaces](#constraint-interfaces) ·
  [Interface methods](#interface-methods) ·
  [Interface comparison](#interface-comparison) ·
  [Type embedding](#type-embedding) ·
  [Switch statement](#switch-statement) ·
  [Empty structs](#empty-structs) ·
  [String literals](#string-and-character-literals) ·
  [Integer constants](#integer-constants) ·
  [Integer literals](#integer-literals) ·
  [Float constants](#float-constants) ·
  [Float literals](#float-literals)
- Interop:
  [Name overrides](#c-field-name-override) ·
  [Variadic nodecay](#variadic-nodecay) ·
  [Target-width types](#target-width-c-types) ·
  [Assume](#assume)
- Safety:
  [Assertions](#assertions)
- Stdlib:
  [crypto/crand](#cryptocrand) ·
  [encoding/json](#encodingjson) ·
  [fmt](#fmt) ·
  [math](#math) ·
  [math/rand](#mathrand) ·
  [net/netip](#netnetip) ·
  [os](#os) ·
  [runtime](#runtime) ·
  [testing](#testing) ·
  [time](#time) ·
  [uuid](#uuid)
- Tooling:
  [so test](#so-test) ·
  [so translate-test](#so-translate-test) ·
  [Freestanding flag](#freestanding-flag) ·
  [Test naming](#test-and-bench-package-naming) ·
  [Windows](#windows)

## Language

### Type parameters

Type parameters are now rejected where C has nothing to emit.

A type parameter works for Go type-checking and as a macro argument, so a type parameter can appear only inside an `so:inline` macro. A declaration with a bare `T` in C used to produce invalid output. It is now an error:

```go
type Pair[T any] struct{ a, b T } // rejected: no C type for T

type Stack[T any] struct{ items []T } // OK: so_Slice is type-erased
```

A generic function or method must be `so:inline` or `so:extern`, even when the signature does not mention the type parameter.

[9501bae](https://github.com/solod-dev/solod/commit/9501baef6386471516b70c22f64ec69b442d008f)

### Constraint interfaces

Constraint interfaces are no longer emitted. Go allows such an interface only as a type parameter constraint. A constraint interface is never a value type, so C has nothing to represent.

```go
type Number interface{ ~int | ~float64 } // not emitted
```

[4f12848](https://github.com/solod-dev/solod/commit/4f12848353acb9c4564948ee05a1aa5e6342511a)

### Interface methods

Interface methods with value receivers are now rejected. A concrete type with value receivers does not convert to an interface:

```go
func (r Rect) Area() int { // value receiver
    return r.width * r.height
}

var s Shape = r // rejected: method Rect.Area has a value receiver
```

A pointer receiver (`func (r *Rect) Area() int`) works.

[7e8e8b2](https://github.com/solod-dev/solod/commit/7e8e8b2125af949db9c78d2fab9cf9c1720cfe2a)

### Interface comparison

Two interfaces are equal when they hold the same pointer. An interface compares with `nil` as expected. Comparing an interface to a concrete type is not supported:

```go
r := Rect{2, 4}
var s Shape = &r
if s == nil { }   // supported
if s == other { } // supported, other is a Shape
if s == &r { }    // not supported
```

An `any` compares with `nil`, with a pointer, or with another `any`. Comparing `any` to a value is not supported, because an `any` holds the address of the value:

```go
var a any = n
if a == nil { }  // supported
if a == &n { }   // supported, &n is a pointer
if a == n { }    // not supported
```

[06b8ab8](https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

### Type embedding

Type embedding is rejected. Struct embedding and interface embedding are errors now:

```go
type number struct {
    base // rejected: embedded field
}

type readWriter interface {
    reader // rejected: embedded interface
    write(v int)
}
```

Declare a named field or list the methods.

[d0e48e0](https://github.com/solod-dev/solod/commit/d0e48e0a5786e47420b9aea1477390db7f2c72fb)

### Switch statement

A switch translates to an if/else-if chain, which repeats the tag in every comparison. The tag now goes into a temporary variable first, so a call in the tag runs once, as in Go:

```go
switch inc() { // used to call inc() once per case
case 2:
    println("two")
default:
    println("other")
}
```

A switch on a struct or an array is not supported.

[06b8ab8](https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

Switch case bodies reject `break` and `fallthrough`. A case body used to keep both statements as written. The emitted C gave incorrect behavior or did not compile. Both statements are errors now.

[fd8659b](https://github.com/solod-dev/solod/commit/fd8659b66bbb2e24f2b70062c5b08b84abffe551)

### Empty structs

A struct with no fields now works as a value, as a slice element, and as a map value:

```go
type empty struct{}

var e empty   // was invalid C, now empty e = {};
p := new(empty)

set := make(map[string]empty, 4)
set["a"] = empty{}
v := set["a"]

es := make([]empty, 1, 4)
es[0] = empty{}
es = append(es, empty{})
```

The zero value of an aggregate is now the empty initializer `{}`. It used to be `{0}`, which sets the first member of the aggregate. A struct with no fields has no member to set, so C rejected `{0}` for an `empty` value, an `[N]empty` array, and any struct with an `empty` first field.

A struct with no fields has a size of zero, like in Go.

[1cb1cb3](https://github.com/solod-dev/solod/commit/1cb1cb3919e4d81e72027651e83b993ea78598c2) ·
[79eeb51](https://github.com/solod-dev/solod/commit/79eeb51d8aa2085e91c310f56cd9454a83c0a755)

### String and character literals

A literal is decoded to its value and encoded again for C. The output no longer keeps the literal as written. A string keeps printable ASCII and UTF-8 as is, and escapes other bytes in octal. A byte or rune literal stays a character literal for printable ASCII. Any other value becomes a number:

```text
"日本語"   ->  "日本語"
"a\xffb"  ->  "a\377b"
'a'       ->  'a'
'世'      ->  0x4e16
```

[f5a91b1](https://github.com/solod-dev/solod/commit/f5a91b1c4902d19787f83b5d8ccc5804d69f2c08)

### Integer constants

A constant integer expression is now emitted as its value (folded) when C cannot do the arithmetic step by step:

```go
var mask uint64 = 1<<64 - 1 // was ((int64_t)1 << 64) - 1, now 18446744073709551615u
```

An expression that C computes correctly is unchanged:

```go
var flags int64 = 1<<20 | 1<<10 // (((int64_t)1 << 20) | ((int64_t)1 << 10))
```

[a4d08c0](https://github.com/solod-dev/solod/commit/a4d08c0dae92ef6ea11972b1ffb81666df8ec702) ·
[5829d2c](https://github.com/solod-dev/solod/commit/5829d2cfbc9d68f93f76a723cefc90bb8d23089b)

An untyped integer constant above `MaxInt64` is now declared as `uint64_t`. The previous declaration was `int64_t`:

```go
const maxUint64 = 18446744073709551615
var third uint64 = maxUint64 / 3 // was 0, now 6148914691236517205
```

A constant that does not fit `uint64_t` is rejected:

```go
const huge = 1 << 200 // rejected: constant huge does not fit in int64 or uint64
```

[bca0714](https://github.com/solod-dev/solod/commit/bca071476d5d5d09eb1a9097f8ddba07847bdcc9)

An untyped integer constant is now converted to the type of the use, to give C the correct Go type used by the expression:

```go
const deBruijn32 = 0x077CB531 // declared as int64_t

func TrailingZeros32(x uint32) int {
    // was x * deBruijn32 in int64, now (uint32_t)deBruijn32 wraps in uint32
    return int(deBruijn32tab[(x&-x)*deBruijn32>>(32-5)])
}
```

[f0b39d0](https://github.com/solod-dev/solod/commit/f0b39d0d9a89f2b2e839f7f459a8c0cd65f7e27c)

A constant shift with a negative left operand is now folded:

```go
var low int64 = -1 << 63 // was (-1 << 63), now INT64_MIN
```

The left operand of a constant shift now gets a cast to the C type of the shift:

```go
var big int64 = (1 + 1) << 62 // was ((1 + 1) << 62), now ((int64_t)(1 + 1) << 62)
```

[0d26001](https://github.com/solod-dev/solod/commit/0d26001eb5f6c85928c3757107626b97e57cfc52)

### Integer literals

An integer literal above `MaxInt64` now gets a `u` suffix in C:

```go
var n uint64 = 18446744073709551615 // -> 18446744073709551615u
```

C gives an unsuffixed decimal literal a signed type. Without the suffix the compiler warned that the value does not fit.

[d18d68d](https://github.com/solod-dev/solod/commit/d18d68d92b65cd37c1cdd096e2a97fc957f629d3)

### Float constants

A constant float expression is emitted as its value (folded). The output no longer shows the operators:

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

[0d81c8b](https://github.com/solod-dev/solod/commit/0d81c8b88c96d569f0e1428ed0ef9c1d5a22d4ce)

### Float literals

A `float32` literal now gets an `f` suffix in C:

```go
var x float32 = 0.1
_ = x == 0.1 // was false, now true
```

Without the suffix the literal was a `double`. C then promoted the other operand to `double`, while Go computed the same expression in `float`.

[0d81c8b](https://github.com/solod-dev/solod/commit/0d81c8b88c96d569f0e1428ed0ef9c1d5a22d4ce)

## Interop

### C field name override

A `c:"..."` struct tag sets the C name of a field. An extern struct can then match a C header field with a Go keyword as the name:

```go
//so:extern SDL_CommonEvent
type SDL_CommonEvent struct {
    etype uint32 `c:"type"` // emitted as .type in C
}
```

[fc25bb8](https://github.com/solod-dev/solod/commit/fc25bb875b10e4d64a6f223ad3f5ec647ed2733e)

### Variadic nodecay

Nodecay on a variadic passes So types. A plain extern variadic is a C variadic: every argument decays, and the callee reads C types. A nodecay extern variadic is not. Each argument goes to the C `...` on its own, at its So type, and every scalar widens:

| So type                            | C type read by `va_arg` |
| ---------------------------------- | ----------------------- |
| any signed integer, `rune`, `bool` | `so_int`                |
| any unsigned integer               | `so_uint`               |
| `float32`, `float64`               | `double`                |
| `string`                           | `so_String`             |
| anything else                      | the type itself         |

```go
//so:extern nodecay
func measure(kinds string, args ...any) int

var n int32 = 7
measure("is", n, "abc")
// measure(so_str("is"), (so_int)(n), so_str("abc"))
```

The call must list its arguments explicitly rather than using spread syntax. An `any` argument is an error.

[efd7892](https://github.com/solod-dev/solod/commit/efd78920df8d065726c629b471ec8320682f5292)

### Target-width C types

`so/c` now supports more common C types:

```text
size_t      - c.Size
ssize_t     - c.SSize
ptrdiff_t   - c.Ptrdiff
intptr_t    - c.Intptr
long double - c.LongDouble
```

[c06b294](https://github.com/solod-dev/solod/commit/c06b29494cda8c822a4191843120fb6a6091a25d)

### Assume

`c.Assume` states a fact for the C compiler. It generates no code in any build, and `-assert` does not affect it:

```go
for i < len(hdib) {
    elem := c.PtrAt(hdib, i)
    c.Assume(elem != nil) // the loop runs only when hdib is non-nil
    // ...
}
```

The behavior is undefined if the condition is false. Use `c.Assume` only for conditions that are provably true, such as when a pointer is known to be non-null but the compiler cannot see it. Use `c.Assert` for all other conditions.

[1488dbc](https://github.com/solod-dev/solod/commit/1488dbc2840a9a1cdee264ee6a76e5992d801568)

## Safety

### Assertions

Assertions are no longer tied to `NDEBUG`. Assertions are on by default. The new `-assert` flag removes them:

```sh
so build -assert=off .
```

`NDEBUG` removed these checks before. A C project that defined `NDEBUG` would turn the So safety checks off by accident.

[f67d8ca](https://github.com/solod-dev/solod/commit/f67d8ca3660437d6f648dd905d7784a5fbb180fa)

## Standard library

### crypto/crand

The package now works in freestanding mode. It rejected a freestanding build with a compile-time error, because no freestanding environment has a CSPRNG. The build now succeeds, and the target supplies the entropy. Define the C function `so_crand_read` and point it at the hardware random number generator of the board or at a host import:

```c
so_int so_crand_read(uint8_t* buf, so_int size) {
    return board_rng_fill(buf, size);
}
```

The hook has a weak default definition, so a program that never calls `crypto/crand` still links. A call with no definition in the target panics.

[451b213](https://github.com/solod-dev/solod/commit/451b213cb61d91c65b56665ba6a12587471afc51) ·
[2d55a04](https://github.com/solod-dev/solod/commit/2d55a04beff2321d5d4f889eb18807e6bc4a1cf5)

### encoding/json

The package now works in freestanding mode. It used `math.IsNaN` and `math.IsInf` to reject a non-finite float. The `math` package requires a hosted environment, so that import alone made `encoding/json` hosted. The package now uses a private finite check and doesn't import `math`.

[211ccd7](https://github.com/solod-dev/solod/commit/211ccd73d8986da72262956f6a63d7ed564bb9b4)

### fmt

The package now works in freestanding mode. The print family used to wrap C's `vsnprintf`, so it printed C's text with C's verbs, and the `<stdio.h>` include made the whole package hosted. It now runs a formatting engine ported from Go's `fmt` and writes the bytes Go writes. Hosted and freestanding builds produce the same output.

The verbs are Go's, with two differences. So has no reflection, so the verbs that need type information are absent: `%v`, `%T`, `%w`, `%q`, and `%U`. And `%u` is added for an unsigned integer, because a print call carries no type information either, so nothing else can tell a signed value from an unsigned one. `%t` for a bool and `%O` for octal with a `0o` prefix are new as well.

`Print`, `Println`, and `Printf` write to the the standard output of the host, or to the `so_write_out` hook in a freestanding environment. The scan family (`Scanf`, `Sscanf`, `Fscanf`) reads through the stdio of the host, so it stays hosted and a freestanding call panics.

⚠️ `BufSize` and `ErrSize` are gone. The output has no size limit any more, so `Fprintf` never returns `ErrSize`.

⚠️ `Buffer`, `NewBuffer` and `BufferFrom` are gone. `Sprintf` takes the destination as a byte slice:

```go
buf := make([]byte, 64)
s := fmt.Sprintf(buf, "%d apples", n)
```

`Buffer` existed because a `[]byte` argument decayed to a bare pointer, which lost the length. The print family is nodecay now, so the slice arrives whole.

[efd7892](https://github.com/solod-dev/solod/commit/efd78920df8d065726c629b471ec8320682f5292) ·
[f7a1eaa](https://github.com/solod-dev/solod/commit/f7a1eaade656450bd84804a24a8b60b82518a7c8)

### math

The package now works in freestanding mode. It rejected a freestanding build with a compile-time error, because a freestanding environment has no libm. The build now succeeds, and only the part that needs libm panics. `Abs`, `Copysign`, `Dim`, `Inf`, `IsInf`, `IsNaN`, `Max`, `Min`, `NaN`, `Pow10`, `RoundToEven`, `Signbit`, the `Float64bits` family and every constant work.

`Abs`, `Copysign` and `Signbit` used to call libm. All three read the bits of the float now, the way Go does.

⚠️ `Max` and `Min` follow Go for NaN. Both wrapped C's `fmax` and `fmin`, which ignore a NaN operand, so `Max(2, NaN)` returned 2. Both are ported from Go now, and they match the documented special cases: a NaN operand gives NaN, `Max` gives `+Inf` before it gives NaN, and `Min` gives `-Inf` before it gives NaN.

[fbfc173](https://github.com/solod-dev/solod/commit/fbfc173e1918d81f968a481f6d221a7c0b9ab453)

### math/rand

The new `NormFloat64` function returns a normally distributed value with a mean of 0 and a standard deviation of 1. It is available as a method of `Rand` and as a top-level function over the global generator:

```go
sample := rand.NormFloat64()*desiredStdDev + desiredMean
```

`NormFloat64` calls `math.Log` and `math.Exp` for a small part of the results, so it requires a hosted environment.

[76b74a5](https://github.com/solod-dev/solod/commit/76b74a5c5a5eee50722773043c3aa56231213d52)

### net/netip

The package now works in freestanding mode. It included `<net/if.h>` for `if_nametoindex`, and that header made the whole package hosted. The include now sits behind a hosted guard, and a freestanding build gets a stub that returns 0. A zone given as an interface name resolves to no zone, the same result a hosted `if_nametoindex` gives for a name that matches no interface. A numeric zone works everywhere.

[758b40d](https://github.com/solod-dev/solod/commit/758b40de473dac084668c3282d87eab3491d51ce)

### os

A `File` writes through a buffered C stream, so a program that ends abnormally loses the buffered data. The new `File.Sync` method flushes the stream:

```go
f.WriteString("progress\n")
f.Sync() // the line is out of the buffer now
```

The data can still wait in an operating system cache, so `Sync` does not guarantee that the data reached the storage device.

[b10bac5](https://github.com/solod-dev/solod/commit/b10bac59bfe030cf159676f453a4ab78022a5b8c)

### runtime

The new `Hosted` constant reports whether the program is running in a hosted environment (one with a C standard library). Use it to skip tests in freestanding mode:

```go
if !runtime.Hosted {
    t.Skip("needs a clock")
    return
}
```

[b69ffb5](https://github.com/solod-dev/solod/commit/b69ffb5eb99cc7d7d9483dd236275e2cbafdf672)

`Seed` now reads the target entropy in freestanding mode. `Seed` used a deterministic generator with a fixed initial state, so `math/rand` and the hash of `maps` repeated on every run. `Seed` now reads the same `so_crand_read` hook as `crypto/crand`, so one definition covers both. A target with no hook keeps the deterministic generator: `math/rand` and `maps` promise nothing about unpredictability and must work on a board with no entropy source.

[55c6936](https://github.com/solod-dev/solod/commit/55c69361731adaafad125f8305c9e9970d88eb7a)

### testing

The package now works in freestanding mode. It imported `os` for the standard output and for `os.Exit`, and that import made it hosted. The test report goes to the standard output (in a hosted environment) or to the `so_write_out` hook (in a freestanding environment). The runner also resets the heap (a static buffer in a freestanding environment) after every test.

`RunSuites`, `RunTests` and `RunBenchmarks` take an [Options](https://pkg.go.dev/solod.dev/so/testing#Options) value in place of the runner arguments. `so test` and `so bench` read `-run` from the command line themselves and write the value into the generated runner.

⚠️ If you have a main package of your own that calls `RunTests` or `RunBenchmarks`, pass a `testing.Options` value instead of `os.Args`.

[c69d8c0](https://github.com/solod-dev/solod/commit/c69d8c0f34e4a44221104e72a87c961850826da6)

### time

The package now reads the clock in freestanding mode. `Now`, `Since`, `Until`, and `Sleep` used to panic. The target now supplies the clock through three hooks, the same way `crypto/crand` supplies entropy:

```c
so_R_i64_i32 so_time_wall(void) {
    return (so_R_i64_i32){.val = board_rtc_unix_seconds(), .val2 = 0};
}

int64_t so_time_mono(void) {
    return board_uptime_ms() * 1000000;
}

void so_time_sleep(int64_t ns) {
    board_delay_ms(ns / 1000000);
}
```

A board that counts elapsed time but does not know the date returns 0 seconds from `so_time_wall`. Every `Time` then dates at the epoch, and `Since` and `Until` stay exact, because they measure with the monotonic clock. A target with no `so_time_mono` still works: `time.Now` returns a wall clock reading alone.

[f2c7963](https://github.com/solod-dev/solod/commit/f2c7963cf423014b9a27cd0c48a6fa85e64b44a9) ·
[2d55a04](https://github.com/solod-dev/solod/commit/2d55a04beff2321d5d4f889eb18807e6bc4a1cf5)

### uuid

The package now works in freestanding mode. It imports `crypto/crand`, so `New` and `NewV4` now work in a freestanding environment as soon as `so_crand_read` is defined, and `NewV7` works once `so_time_wall` is defined too.

[451b213](https://github.com/solod-dev/solod/commit/451b213cb61d91c65b56665ba6a12587471afc51)

## Tooling

### so test

`so test` can now run tests from multiple packages in a single run. A pattern that ends with `...` selects every package with a `test` subdirectory below its base directory:

```sh
so test ./so/...
```

The whole run costs one translate, one compile and one execution, which is much faster than one run per package. Narrow the pattern to select a group of packages: `so test ./so/net/...` runs the tests of `so/net` and `so/net/netip`.

The packages share a process, so a hard crash in one package stops the packages after it.

[b10bac5](https://github.com/solod-dev/solod/commit/b10bac59bfe030cf159676f453a4ab78022a5b8c)

The `-pkg-file` flag limits the run to the packages a file lists:

```sh
# freestanding.txt
so/bytes
so/mem
```

```sh
so test -pkg-file=freestanding.txt ./so/...
```

The file holds one package per line, as a path relative to the module root. Blank lines and the text after a `#` are ignored. The list is a filter over the packages the pattern selects, so a listed package that the pattern does not select is an error.

[c69d8c0](https://github.com/solod-dev/solod/commit/c69d8c0f34e4a44221104e72a87c961850826da6)

### so translate-test

The new `translate-test` command writes the C of the test program without a compile or a run, the way `so translate` does for an ordinary package:

```sh
so translate-test -pkg-file=freestanding.txt -run=TestBuffer -o out ./so/...
```

[c69d8c0](https://github.com/solod-dev/solod/commit/c69d8c0f34e4a44221104e72a87c961850826da6)

### Freestanding flag

The new `-freestanding` flag tells `so build`, `so test`, `so bench` and `so run` that the program targets a freestanding environment:

```sh
export CC=clang
export CFLAGS="--target=wasm32-freestanding -nostdlib -Wl,--no-entry -Wl,--export=main"
so build -freestanding -o main.wasm .
```

The flag drops the libc dependencies declared with `so:link` in the standard library, because a freestanding target has no host C library. It does not affect the libraries declared with `so:link` in user code.

The target still comes from `CFLAGS`; the flag doesn't affect it.

[5bda474](https://github.com/solod-dev/solod/commit/5bda4746a4b85d3dc14ca43286e034568017212c)

### Test and bench package naming

A test or bench directory declared `package main` before. The generated runner must import the package now, so `package main` is rejected. Two test packages of one run must also have distinct names, because the So compiler prefixes an exported C name with the package name.

A common convention is to name a test package after the package under test, with a `_test` or `_bench` suffix: `so/sync/test` declares `package sync_test`, and `so/sync/bench` declares `package sync_bench`.

The runner is no longer written to disk. `so test` and `so bench` pass the generated runner to the Go loader as an in-memory overlay, so the committed `test/main.go` and `bench/main.go` files are gone. Adding, renaming or removing a `TestXxx` or `BenchmarkXxx` needs no other step.

⚠️ If you have test or benchmark subpackages, rename them from `main` to `{package}_test` or `{package}_bench`, and remove the `main.go` files.

[b10bac5](https://github.com/solod-dev/solod/commit/b10bac59bfe030cf159676f453a4ab78022a5b8c)

### Windows

The standard library now builds for `windows/amd64` and `windows/arm64`. Every package of the freestanding set works. The packages that need POSIX (`conc`, `flag`, `log/slog`, `net`, `os`, `sync`) are not supported.

Use `zig cc` to cross-compile for Windows:

```sh
export CC="zig cc"
export CFLAGS="--target=x86_64-windows-gnu"
export LDFLAGS="-lbcrypt -liphlpapi"
so build -o app.exe .
```
