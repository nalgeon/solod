## Solod 0.4 (in progress)

This document outlines the main changes in the latest So version currently in development.

## Language

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

[4f12848](https://github.com/solod-dev/solod/commit/4f12848353acb9c4564948ee05a1aa5e6342511a)

**Interface methods with value receivers are rejected**. If a concrete type uses value receivers, converting it to an interface will fail:

```go
func (r Rect) Area() int { // value receiver
    return r.width * r.height
}

var s Shape = r // rejected: method Rect.Area has a value receiver
```

Using pointer receivers (`func (r *Rect) Area() int`) works fine.

[7e8e8b2](https://github.com/solod-dev/solod/commit/7e8e8b2125af949db9c78d2fab9cf9c1720cfe2a)

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

Switching on a struct or array is not supported.

[06b8ab8](https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

**Switch case bodies reject break and fallthrough**. Both used to be emitted as-is, producing C that either lead to incorrect behavior or did not compile at all. Both are errors now.

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

[06b8ab8](https://github.com/solod-dev/solod/commit/06b8ab85bf4cb065b760c5a0e750b1bf6658d076)

**String and character literals**. A literal is decoded to its value and re-encoded for C instead of being copied into the output as written. A string keeps printable ASCII and UTF-8 as is and escapes other bytes in octal. A byte or rune literal stays a character literal only for printable ASCII, and becomes a number otherwise:

```text
"日本語"   ->  "日本語"
"a\xffb"  ->  "a\377b"
'a'       ->  'a'
'世'      ->  0x4e16
```

[f5a91b1](https://github.com/solod-dev/solod/commit/f5a91b1c4902d19787f83b5d8ccc5804d69f2c08)

**Integer literals above MaxInt64**. Such a literal now gets a `u` suffix in C. C gives an unsuffixed decimal literal a signed type, so without the suffix the compiler warned that the value does not fit:

```go
var n uint64 = 18446744073709551615 // -> 18446744073709551615u
```

[d18d68d](https://github.com/solod-dev/solod/commit/d18d68d92b65cd37c1cdd096e2a97fc957f629d3)

**Untyped integer constants above MaxInt64**. Such a constant is now declared as `uint64_t` instead of `int64_t`:

```go
const maxUint64 = 18446744073709551615
var third uint64 = maxUint64 / 3 // was 0, now 6148914691236517205
```

A constant that does not fit `uint64_t` either is rejected:

```go
const huge = 1 << 200 // rejected: constant huge does not fit in int64 or uint64
```

[bca0714](https://github.com/solod-dev/solod/commit/bca071476d5d5d09eb1a9097f8ddba07847bdcc9)

**Integer constant folding**. A constant integer expression is now emitted as its value when C cannot compute it step by step:

```go
var mask uint64 = 1<<64 - 1 // was ((int64_t)1 << 64) - 1, now 18446744073709551615u
```

Expressions that C computes correctly are unchanged:

```go
var flags int64 = 1<<20 | 1<<10 // (((int64_t)1 << 20) | ((int64_t)1 << 10))
```

**The smallest int64**. The value -9223372036854775808 is now emitted as `INT64_MIN`. C has no negative literals, so as a unary minus applied to a value that no signed C type can hold, it was read as unsigned:

```go
const minInt64 = -(1 << 63) // was -9223372036854775808, now INT64_MIN
```

[a4d08c0](https://github.com/solod-dev/solod/commit/a4d08c0dae92ef6ea11972b1ffb81666df8ec702)

**Float constant folding**. A constant float expression is emitted as its value, not as the operators used to produce it:

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

[0d81c8b](https://github.com/solod-dev/solod/commit/0d81c8b88c96d569f0e1428ed0ef9c1d5a22d4ce)

## Interop

**C field name override**. A `c:"..."` struct tag sets the C name of a field, so an extern struct can match a C header field whose name is a Go keyword:

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

## Safety

**Assertions are no longer tied to NDEBUG**. They are on by default and removed only by the new `-assert` flag:

```sh
so build -assert=off .
```

Previously, `NDEBUG` removed these checks. This meant that a C project that defines it could accidentally turn So's safety net off.

[f67d8ca](https://github.com/solod-dev/solod/commit/f67d8ca3660437d6f648dd905d7784a5fbb180fa)
