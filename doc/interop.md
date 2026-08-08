# C interop

So provides several tools for easy C interop.

[Includes](#includes) •
[Linking](#linking) •
[Extern declarations](#extern-declarations) •
[Inlining](#inlining) •
[Promoting](#promoting) •
[Qualifiers](#qualifiers) •
[Embeds](#embeds) •
[Raw C](#raw-c-experimental) •
[Helpers](#helpers)

## Includes

Include a C header file. By default, `so:include` emits in the `.h` file, making the header visible to consumers:

```go
//so:include <stdint.h>
```

Use `so:include.c` when the include is purely an implementation detail that should only appear in the `.c` file:

```go
//so:include.c "internal_helper.h"
```

## Linking

When a package uses a C library that must be linked explicitly, declare it with `so:link`. The name is the library as passed to the linker's `-l` flag, without the prefix:

```go
//so:include <pthread.h>
//so:link pthread
```

`so build`, `so run`, and `so test` collect the `so:link` libraries from every transpiled package, deduplicate them, and pass them to the C compiler (`-lpthread` above) after your `LDFLAGS`.

The standard library already uses `so:link` for its packages. For example, importing `so/math` automatically links `-lm`, and importing `so/sync` or `so/conc` automatically links `-lpthread`.

The flags are always emitted. On platforms where a library is already part of libc (for example pthreads and libm on macOS) the extra `-l` is a harmless no-op.

## Extern declarations

Declare an external C type (excluded from emission) with `so:extern`:

```go
//so:extern
type Account struct {
    name    string
    balance int64
    flags   []uint8
}
```

Declare an external C function:

```go
//so:extern
func dec_balance(acc *Account, amount int64) int64 {
    return 42 // for testing
}
```

When calling extern functions, `string` and `[]T` arguments are automatically decayed to their C equivalents: string literals become raw C strings (`"hello"`), string values become `char*` (`.ptr`), and slices become raw pointers (`.ptr`). This means C macros don't need to extract `.ptr` themselves:

```go
//so:extern
func fopen(path string, mode string) *File { return nil }

// Go call:
f := fopen("/tmp/test.txt", "w")

// Generated C:
// fopen("/tmp/test.txt", "w")
// not fopen(so_str("/tmp/test.txt"), so_str("w"))
```

The `so:extern` directive supports two optional parameters: a C name override and the `nodecay` flag.

Methods can be extern too.

### Extern options

_Name override_ specifies the C name to use instead of the default package-prefixed name. Useful for extern types that must match a C header:

```go
//so:extern Account
type Account struct {
    name    string
    balance int64
}
// Uses "Account" in C instead of "main_Account"
```

_Nodecay_ skips the automatic decay of So types (`so_String`, `so_Slice`) to raw C types. Use this for C functions that are "So-aware" and accept So types directly:

```go
//so:extern nodecay
func set_name(acc *Account, name string)

// Generated C passes so_String directly:
// set_name(&acc, name)
// not set_name(&acc, so_cstr(name))
```

Both options can be combined:

```go
//so:extern MyFunc nodecay
func MyFunc(s string)
```

### Field names

A `c:"..."` struct tag overrides the C name of a single field. This is needed when a C struct has a field whose name is a Go keyword, such as `type`:

```go
//so:extern SDL_CommonEvent
type SDL_CommonEvent struct {
    etype     uint32 `c:"type"`
    timestamp uint64
}
```

So sees the field as `etype`; the generated C uses `type` everywhere, including field accesses in packages that import the struct. The tag value must be a valid C identifier that is not a C keyword, and it may not collide with another field's C name in the same struct.

The tag is honored only on the fields of a named struct type, not on anonymous or function-local structs. It works on any named struct, not just extern ones, but its main use is matching an external C layout.

## Inlining

Force a function to be emitted as `static inline` in the header file using `//so:inline`. This is useful for small, frequently used functions when the compiler won't inline them automatically:

```go
//so:inline
func add(a, b int) int {
    return a + b
}
```

The function body is emitted directly in the `.h` file and skipped from the `.c` file. Works with both functions and methods.

## Promoting

By default an unexported symbol (lowercase name) stays in the `.c` file with its bare name. You can promote it into the header with `//so:promote`, which also gives it the package prefix:

```go
//so:promote
type counter struct { val int }

//so:promote
func newCounter() counter { ... }

//so:promote
func (e *counter) inc() { ... }
```

```c
// pkg.h
typedef struct pkg_counter { so_int val; } pkg_counter;
pkg_counter pkg_newCounter(void);
void pkg_counter_inc(void* self);
```

Types are emitted in full; functions and methods get a header prototype while their body stays in the `.c` file; variables become `extern`; constants are emitted with their value. A method's C name comes from its receiver type, so an `so:promote` method requires the receiver type to be exported or `so:promote` too; otherwise it is rejected.

This is needed when an exported `so:inline` function (whose body lives in the header) calls an unexported helper, or when an exported type has a field of an unexported type:

```go
type Stats struct { c counter }

//so:inline
func NewStats() Stats {
	return Stats{c: newCounter()}
}
```

Without `so:promote`, the header would reference a name it never declares. The alternative (exporting the helper) pollutes the public API; `so:promote` keeps it out of the Go API while still making the C declaration visible.

`so:promote` works on types, functions, methods, vars, and consts. It is rejected on exported declarations (already in the header, so redundant) and cannot combine with `so:inline` (which already emits the body in the header).

## Qualifiers

### Volatile

Mark a package-level variable as `volatile` using `//so:volatile`:

```go
//so:volatile
var counter int
```

Only allowed on `var` declarations.

### Thread-local storage

Mark a package-level variable as thread-local using `//so:thread_local`. Uses C11 `_Thread_local`:

```go
//so:thread_local
var perThread int
```

Can be combined with `//so:volatile`:

```go
//so:volatile
//so:thread_local
var flags int
```

Only allowed on `var` declarations.

### Attributes

Add GCC/Clang `__attribute__` annotations using `//so:attr`. The text after `so:attr` is used as the attribute value:

```go
//so:attr packed
type header struct {
    version byte
    length  int
}
```

Multiple `//so:attr` lines on the same declaration are combined:

```go
//so:attr packed
//so:attr aligned(16)
type aligned struct {
    x int
}
```

Allowed on `var`, `const`, `type`, and `func` declarations.

## Embeds

Embed C files directly into the generated output using `//so:embed`:

```go
//so:embed main.h
var main_h string

//so:embed main.c
var main_c string
```

`.h` files are embedded into the generated header, `.c` files into the generated implementation. The embed variable declarations are not emitted as C variables - they serve as markers only.

## Raw C (experimental)

For ad-hoc C interop, the `so/c` package provides two compiler intrinsics that emit their string argument as raw C code. The argument must be a string literal.

`c.Val[T](expr)` emits a typed C expression. Use it to access C constants, macros, or call C functions inline:

```go
nan := c.Val[float64]("NAN")
x := c.Val[float64]("sqrt(49)")
```

`c.Raw(code)` emits a raw block of C code as a statement:

```go
var b int
c.Raw(`
int a = 7;
b = a * a;
`)
```

Be careful when using `c.Val` and `c.Raw`. C code written as string literals bypasses the type system and is hard to maintain, so it's usually better to use `so:extern` and `so:embed` instead.

## Helpers

The `so/c` package also provides low-level interop helpers for pointers, strings, and type information.

Functions:

- `Alignof` and `Sizeof` return the alignment and size of type T.
- `Alloca` allocates an array on the stack.
- `Assert` panics with a message if a condition is false.
- `Assume` tells the C compiler that a condition is always true.
- `Bytes`, `Slice` and `String` wrap C pointers to So types.
- `CString` converts a So string to a null-terminated C string.
- `PtrAdd`, `PtrAs` and `PtrAt` manipulate pointers.
- `Zero` returns the zero value of type T.

Types:

- `Char` and `ConstChar` represent a C `char` type.
- `Int`, `UInt`, `Long`, `ULong`, etc. represent numeric C types.
- `Size`, `SSize`, `Ptrdiff` and `Intptr` represent C types whose width follows the target.
