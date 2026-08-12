# Solod 0.4 (in progress)

This document lists the main changes in the So version in development.

## Language

**Type parameters are rejected where C has nothing to emit**. A type parameter works for Go type-checking and as a macro argument, so a type parameter can appear only inside an `so:inline` macro. A declaration with a bare `T` in C used to produce invalid output. It is now an error:

```go
type Pair[T any] struct{ a, b T } // rejected: no C type for T

type Stack[T any] struct{ items []T } // OK: so_Slice is type-erased
```

A generic function or method must be `so:inline` or `so:extern`, even when the signature does not mention the type parameter. See the [generics guide](./generics.md) for details.

[9501bae](https://github.com/solod-dev/solod/commit/9501baef6386471516b70c22f64ec69b442d008f)

**Constraint interfaces are no longer emitted**. Go allows such an interface only as a type parameter constraint. A constraint interface is never a value type, so C has nothing to represent.

```go
type Number interface{ ~int | ~float64 } // not emitted
```

[4f12848](https://github.com/solod-dev/solod/commit/4f12848353acb9c4564948ee05a1aa5e6342511a)

**Interface methods with value receivers are rejected**. A concrete type with value receivers does not convert to an interface:

```go
func (r Rect) Area() int { // value receiver
    return r.width * r.height
}

var s Shape = r // rejected: method Rect.Area has a value receiver
```

A pointer receiver (`func (r *Rect) Area() int`) works.

[7e8e8b2](https://github.com/solod-dev/solod/commit/7e8e8b2125af949db9c78d2fab9cf9c1720cfe2a)

**Type embedding is rejected**. Struct embedding and interface embedding are errors now:

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

**Switch tag evaluation**. A switch translates to an if/else-if chain, which repeats the tag in every comparison. The tag now goes into a temporary variable first, so a call in the tag runs once, as in Go:

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

**Switch case bodies reject break and fallthrough**. A case body used to keep both statements as written. The emitted C gave incorrect behavior or did not compile. Both statements are errors now.

[fd8659b](https://github.com/solod-dev/solod/commit/fd8659b66bbb2e24f2b70062c5b08b84abffe551)

**Interface comparison**. Two interfaces are equal when they hold the same pointer. An interface compares with `nil` as expected. Comparing an interface to a concrete type is not supported:

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

**String and character literals**. A literal is decoded to its value and encoded again for C. The output no longer keeps the literal as written. A string keeps printable ASCII and UTF-8 as is, and escapes other bytes in octal. A byte or rune literal stays a character literal for printable ASCII. Any other value becomes a number:

```text
"日本語"   ->  "日本語"
"a\xffb"  ->  "a\377b"
'a'       ->  'a'
'世'      ->  0x4e16
```

[f5a91b1](https://github.com/solod-dev/solod/commit/f5a91b1c4902d19787f83b5d8ccc5804d69f2c08)

**Integer constants**. A constant integer expression is now emitted as its value (folded) when C cannot do the arithmetic step by step:

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

**Integer literals**. An integer literal above `MaxInt64` now gets a `u` suffix in C:

```go
var n uint64 = 18446744073709551615 // -> 18446744073709551615u
```

C gives an unsuffixed decimal literal a signed type. Without the suffix the compiler warned that the value does not fit.

[d18d68d](https://github.com/solod-dev/solod/commit/d18d68d92b65cd37c1cdd096e2a97fc957f629d3)

**Float constants**. A constant float expression is emitted as its value (folded). The output no longer shows the operators:

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

**Float32 literals**. A `float32` literal now gets an `f` suffix in C:

```go
var x float32 = 0.1
_ = x == 0.1 // was false, now true
```

Without the suffix the literal was a `double`. C then promoted the other operand to `double`, while Go computed the same expression in `float`.

[0d81c8b](https://github.com/solod-dev/solod/commit/0d81c8b88c96d569f0e1428ed0ef9c1d5a22d4ce)

**Empty struct**. A struct with no fields now works as a value, as a slice element, and as a map value:

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

## Interop

**C field name override**. A `c:"..."` struct tag sets the C name of a field. An extern struct can then match a C header field with a Go keyword as the name:

```go
//so:extern SDL_CommonEvent
type SDL_CommonEvent struct {
    etype uint32 `c:"type"` // emitted as .type in C
}
```

[fc25bb8](https://github.com/solod-dev/solod/commit/fc25bb875b10e4d64a6f223ad3f5ec647ed2733e)

**Target-width C types**. `so/c` now supports more common C types:

```text
size_t      - c.Size
ssize_t     - c.SSize
ptrdiff_t   - c.Ptrdiff
intptr_t    - c.Intptr
long double - c.LongDouble
```

[c06b294](https://github.com/solod-dev/solod/commit/c06b29494cda8c822a4191843120fb6a6091a25d)

**c.Assume states a fact for the C compiler**. It generates no code in any build, and `-assert` does not affect it:

```go
for i < len(hdib) {
    elem := c.PtrAt(hdib, i)
    c.Assume(elem != nil) // the loop runs only when hdib is non-nil
    // ...
}
```

The behavior is undefined if the condition is false. Use `c.Assume` only for conditions that are provably true, such as when a pointer is known to be non-null but the compiler cannot see it. Use `c.Assert` for all other conditions.

## Safety

**Assertions are no longer tied to NDEBUG**. Assertions are on by default. The new `-assert` flag removes them:

```sh
so build -assert=off .
```

`NDEBUG` removed these checks before. A C project that defined `NDEBUG` would turn the So safety checks off by accident.

[f67d8ca](https://github.com/solod-dev/solod/commit/f67d8ca3660437d6f648dd905d7784a5fbb180fa)

## Standard library

**os.File.Sync**. A `File` writes through a buffered C stream, so a program that ends abnormally loses the buffered data. `Sync` flushes the stream:

```go
f.WriteString("progress\n")
f.Sync() // the line is out of the buffer now
```

The data can still wait in an operating system cache, so `Sync` does not guarantee that the data reached the storage device.

**net/netip works in freestanding mode**. The package included `<net/if.h>` for `if_nametoindex`, and that header made the whole package hosted. The include now sits behind a hosted guard, and a freestanding build gets a stub that returns 0. A zone given as an interface name resolves to no zone, the same result a hosted `if_nametoindex` gives for a name that matches no interface. A numeric zone works everywhere. See [freestanding mode](freestanding.md).

**encoding/json works in freestanding mode**. The package used `math.IsNaN` and `math.IsInf` to reject a non-finite float. The `math` package requires a hosted environment, so that import alone made `encoding/json` hosted. The package now uses a private finite check and doesn't import `math`. See [freestanding mode](freestanding.md).

## Tooling

**Multi-package so test**. A pattern that ends with `...` selects every package with a `test` subdirectory below its base directory:

```sh
so test ./so/...
```

The whole run costs one translate, one compile and one execution, which is much faster than one run per package. Narrow the pattern to select a group of packages: `so test ./so/net/...` runs the tests of `so/net` and `so/net/netip`.

The packages share a process, so a hard crash in one package stops the packages after it.

**Test and bench package naming**. A test or bench directory declared `package main` before. The generated runner must import the package now, so `package main` is rejected. Two test packages of one run must also have distinct names, because the So compiler prefixes an exported C name with the package name.

A common convention is to name a test package after the package under test, with a `_test` or `_bench` suffix: `so/sync/test` declares `package sync_test`, and `so/sync/bench` declares `package sync_bench`.

The runner is no longer written to disk. `so test` and `so bench` pass the generated runner to the Go loader as an in-memory overlay, so the committed `test/main.go` and `bench/main.go` files are gone. Adding, renaming or removing a `TestXxx` or `BenchmarkXxx` needs no other step.

⚠️ This is a breaking change. If you have test or benchmark subpackages, rename them from `main` to `{package}_test` or `{package}_bench`, and remove the `main.go` files.
