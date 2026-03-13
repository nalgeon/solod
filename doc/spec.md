# So language description

Solod (So) is a strict subset of Go that transpiles to regular C. This document lists the features it supports. If a feature isn't listed, it's not supported.

[Values](#values) •
[Variables](#variables) •
[Strings](#strings) •
[Arrays](#arrays) •
[Slices](#slices) •
[If/else](#ifelse) •
[For](#for) •
[Goto](#goto) •
[Functions](#functions) •
[Multiple returns](#multiple-return-values) •
[Variadic functions](#variadic-functions) •
[Structs](#structs) •
[Methods](#methods) •
[Interfaces](#interfaces) •
[Enums](#enums) •
[Errors](#errors) •
[Panic](#panic) •
[Defer](#defer) •
[C interop](#c-interop) •
[Embeds](#embeds) •
[Packages](#packages)

## Values

So supports basic Go types:

```go
// Integers.
const d1 = 123
const d2 = 100_000
const d3 = 0b1010
const d4 = 0o600
const d5 = 0xBadFace
const d6 = 0x_67_7a_2f_cc_40_c6

// Floating-point numbers.
const f1 = 3.14
const f2 = 0.25
const f3 = 1e-9
const f4 = 6.022e23
const f5 = 1e6

// Runes.
const r1 = 'a'
const r2 = 'ä'
const r3 = '本'
const r4 = '\xff'
const r5 = '\u12e4'
```

In C, the default type for integers is `so_int` (`int64_t`), for floats it's `double`, and for runes it's `int32_t`.

Complex numbers are not supported.

Constants are translated to C `const` qualifiers.

## Variables

So supports all the main ways to declare and initialize a variable in Go.

`var` with an explicit or inferred type:

```go
var vInt int = 42
var vFloat float64 = 3.14
var vBool bool = true
var vByte byte = 'x'
var vRune rune = '本'
var vString = "hello"
var vSlice = []int{1, 2, 3}
var vStruct = person{age: 42}
var vPtr = &vStruct
var vAnyVal any = 42
var vAnyPtr any = vPtr
var vNil any = nil
```

Short variable declaration:

```go
vInt := 42
vFloat := 3.14
vBool := true
vByte := 'x'
vRune := '本'
vString := "hello"
vSlice := []int{1, 2, 3}
vStruct := person{age: 42}
vPtr := &vStruct
vAnyVal := any(42)
vAnyPtr := any(vPtr)
vNil := any(nil)
```

`byte` is translated to `so_byte` (`uint8_t`), `rune` to `so_rune` (`int32_t`), and `int` to `so_int` (`int64_t`).

`any` is not treated as an interface. Instead, it's translated to `void*`. This makes handling pointers much easier and removes the need for `unsafe.Pointer`.

`nil` is translated to `NULL`.

As in Go, all variables are implicitly initialized to zero values:

```go
var vInt int        // 0
var vFloat float64  // 0
var vBool bool      // false
var vByte byte      // 0
var vRune rune      // 0
var vString string  // "", len=0
var vSlice []int    // len=0, cap=0
var vStruct person  // all fields are set to zero values
var vPtr *person    // NULL
var vNil any        // NULL
```

## Strings

Strings are represented as `so_String` type in C (a struct with a `ptr` and `len`). String literals are wrapped in `so_str()`.

Indexing a string returns a byte (`uint8_t`):

```go
str := "Hi 世界!"
chr := str[0] // byte value
```

Iterating over a string with `range` decodes UTF-8 runes:

```go
for i, r := range str {
    println("i =", i, "r =", r)
}
```

Slicing a string returns a new string (zero-copy):

```go
s := "hello"
s1 := s[:]    // "hello"
s2 := s[2:]   // "llo"
s3 := s[:3]   // "hel"
s4 := s[1:4]  // "ell"
```

String comparison uses dedicated functions (`so_string_eq`, etc.) instead of C operators:

```go
s1 := "hello"
s2 := "world"
if s1 == s2 || s1 < s2 {
    println("ok")
}
```

Converting a string to a byte or rune slice:

```go
s := "1世3"
bs := []byte(s)  // byte slice
rs := []rune(s)  // rune slice
```

String concatenation with `+` is supported for string literals (but not for variables).

## Arrays

Arrays are represented as plain C arrays (`T name[N]`). They are value types - copied on struct assignment and support direct indexing.

Array literals:

```go
var a [5]int                       // zero-initialized
b := [5]int{1, 2, 3, 4, 5}         // explicit values
c := [...]int{1, 2, 3, 4, 5}       // inferred size
d := [...]int{100, 3: 400, 500}    // designated initializers
```

Named array types:

```go
type IntArray [3]int
var arr IntArray
```

Arrays can be struct fields:

```go
type Box struct {
    nums [3]int
}
```

`len()` and `cap()` on arrays are emitted as compile-time constants.

Slicing an array produces a `so_Slice`:

```go
nums := [...]int{1, 2, 3, 4, 5}
s := nums[1:4]  // s is a so_Slice
```

Limitations:

- Arrays decay to pointers when passed to functions (no value semantics on calls).
- Cannot return arrays from functions.
- Array assignment uses `memcpy`.

## Slices

Slices are represented as `so_Slice` type in C (a struct with a data pointer, `len`, and `cap`).

Slice literals:

```go
strs := []string{"a", "b", "c"}
twoD := [][]int{{1, 2, 3}, {4, 5, 6}}
```

Slicing:

```go
s1 := nums[:]    // full slice
s2 := nums[2:]   // from index 2
s3 := nums[:3]   // up to index 3
s4 := nums[1:4]  // from 1 to 4
```

Full slice expressions are not supported.

Built-in operations:

```go
s := make([]int, 4)         // allocate with len=4, cap=4
s = make([]int, 0, 8)       // allocate with len=0, cap=8
s = append(s, 1)            // append a single value
s = append(s, 2, 3)         // append multiple values
s = append(s, other...)     // append another slice
n := copy(dst, src)         // copy elements
l := len(s)                 // length
c := cap(s)                 // capacity
x := s[2]                   // index access
```

`make()` allocates a fixed amount of memory on the stack (`sizeof(T)*cap`). `append()` only works up to the initial capacity and panics if it's exceeded. There's no automatic reallocation. Use the standard library for heap allocation and dynamic arrays.

Iterating over a slice with `range`:

```go
for i, v := range nums {
    println(i, v)
}
```

Arithmetic and bitwise compound assignments work on slice elements:

```go
s[1] += 10
s[1] <<= 2
s[1]++
```

`clear` is not supported.

## If/else

Standard `if`, `else if`, and `else`:

```go
if 7%2 == 0 {
    println("even")
} else {
    println("odd")
}
```

Chained conditions:

```go
if x > 0 {
    println("positive")
} else if x < 0 {
    println("negative")
} else {
    println("zero")
}
```

Init statement (scoped to the if block):

```go
if num := 9; num < 10 {
    println(num, "has 1 digit")
}
```

`switch` is not supported. Use `if-then-else` instead.

## For

Traditional for loop:

```go
for j := 0; j < 3; j++ {
    println(j)
}
```

While-style loop:

```go
for i <= 3 {
    println(i)
    i = i + 1
}
```

Infinite loop:

```go
for {
    println("loop")
    break
}
```

Range over an integer:

```go
for k := range 3 {
    println(k)
}
```

Range over a slice and range over a string are also supported.

`break` and `continue` work as expected.

## Goto

Labels and `goto` map directly to C:

```go
for i := range 10 {
    if i%2 == 0 {
        goto next
    }
next:
    fails++
    if fails > 2 {
        goto fallback
    }
}

fallback:
    println("done")
```

## Functions

Regular function declarations:

```go
func sumABC(a, b, c int) int {
    return a + b + c
}
```

Named function types and function variables:

```go
type SumFn func(int, int, int) int

fn1 := sumABC           // infer type
var fn2 SumFn = sumABC  // explicit type
s := fn2(7, 8, 9)
```

Function literals (anonymous functions / closures) are not supported. Use named types instead, like `SumFn` in the example above.

Exported functions (capitalized) become public C symbols prefixed with the package name (`package_Func`). Unexported functions are `static`.

Exported functions must only use exported types in their signatures (parameters and return types).

## Multiple return values

So supports two-value multiple returns in two patterns: `(T, error)` and `(T1, T2)`.

The `(T, error)` pattern - the second value is `error`:

```go
func divide(a, b int) (int, error) {
    return a / b, nil
}
```

The `(T1, T2)` pattern - two values of any supported type:

```go
func divmod(a, b int) (int, int) {
    return a / b, a % b
}
```

Destructuring:

```go
q, err := divide(10, 3)     // new variables
q, err = divide(20, 7)      // reassign existing
_, err2 := divide(10, 3)    // blank identifier
r, _ := divide(10, 3)       // ignore second value

d, m := divmod(10, 3)       // two values
_, m2 := divmod(10, 3)      // blank identifier
```

If-init with multi-return:

```go
if n, err := f.Read(64); err != nil {
    println("error")
}
```

Forwarding a multi-return call:

```go
func forwardCall() (int, error) {
    return divide(10, 3)
}

func forwardDivmod() (int, int) {
    return divmod(10, 3)
}
```

Supported return types:

```go
bool
byte
float64
int
int64
rune
string
[]T
*T
```

Not supported: returning struct values or interface values, named return values.

## Variadic functions

Variadic functions use the standard `...` syntax:

```go
func sum(nums ...int) int {
    total := 0
    for _, num := range nums {
        total += num
    }
    return total
}
```

Calling with individual arguments or spreading a slice:

```go
sum(1, 2)
sum(1, 2, 3)

nums := []int{1, 2, 3, 4}
sum(nums...)
```

## Structs

Struct type declarations:

```go
type Person struct {
    name string
    age  int
}
```

Struct literals (positional, named fields, or partial):

```go
bob := Person{"Bob", 20}
alice := Person{name: "Alice", age: 30}
fred := Person{name: "Fred"}
```

Pointer to struct:

```go
ann := &Person{name: "Ann", age: 40}
```

Field access (automatically uses `->` for pointers in C):

```go
ann.age = 41
sp := &sean
sp.age = 51
```

Anonymous structs:

```go
dog := struct {
    name   string
    isGood bool
}{"Rex", true}
```

`new()` works with types and values:

```go
n := new(int)           // *int, zero-initialized
p := new(point)         // *point, zero-initialized
n2 := new(42)           // *int with value 42
p2 := new(point{1, 2})  // *point with values
```

## Methods

Methods are defined on struct types with pointer receivers:

```go
type Rect struct {
    width, height int
}

func (r *Rect) Area() int {
    return r.width * r.height
}

func (r *Rect) perim(n int) int {
    return n * (2*r.width + 2*r.height)
}
```

Calling methods on values and pointers:

```go
r := Rect{width: 10, height: 5}
r.Area()    // called on value (address taken automatically)

rp := &r
rp.Area()   // called on pointer
```

Methods on primitive/named types are not supported.

## Interfaces

Interfaces in So are more like Rust traits (static) than Go interfaces (dynamic) because they don't include type information at runtime.

Interface declarations list the required methods:

```go
type Shape interface {
    Area() int
    Perim(n int) int
}
```

In C, an interface is a struct with a `void* self` pointer and function pointers for each method (less efficient than using a static method table, but simpler; this might change in the future).

Converting a concrete type to an interface:

```go
s := Shape(r)
var s2 Shape = r
var s3 Shape = &r
```

Passing a concrete type to functions that accept interfaces:

```go
func calcShape(s Shape) int {
    return s.Perim(2) + s.Area()
}

calcShape(r)        // implicit conversion
calcShape(Shape(r)) // explicit conversion
```

Type assertions:

```go
_, ok := s.(Rect)     // comma-ok pattern (checks without panic)
r := s.(Rect)         // direct assertion

_, ok := l.(*Rect)    // pointer type assertion
r := l.(*Rect)

// But not both; this is not supported.
// r, ok := l.(*Rect)
```

Value receivers and pointer receivers are both supported. Pointer receiver interfaces require passing a pointer:

```go
type Line interface {
    Length() int
}

func (r *Rect) Length() int { ... }

l := Line(&r) // must pass pointer
```

Empty interfaces (`interface{}` and `any`) are translated to `void*`.

Converting between interfaces is not supported: no type assertions like `iface.(AnotherIface)` and no type switches.

## Enums

So supports typed constant groups as enums:

```go
type HttpStatus int

const (
    StatusOK       HttpStatus = 200
    StatusNotFound HttpStatus = 404
    StatusError    HttpStatus = 500
)
```

String-based enums:

```go
type ServerState string

const (
    StateIdle      ServerState = "idle"
    StateConnected ServerState = "connected"
    StateError     ServerState = "error"
)
```

Each constant is emitted as a C `const`. Exported constants are public, unexported ones are `static`.

`iota` is supported for integer-typed constants:

```go
type Day int

const (
    Sunday Day = iota
    Monday
    Tuesday
)
```

Iota values are evaluated at compile time and translated to integer literals:

```c
typedef so_int main_Day;

const main_Day main_Sunday = 0;
const main_Day main_Monday = 1;
const main_Day main_Tuesday = 2;
```

## Errors

Errors use the `so_Error` type (a pointer). So only supports sentinel errors, which are defined at the package level using `errors.New`:

```go
import "github.com/nalgeon/solod/so/errors"

var ErrOutOfTea = errors.New("no more tea available")
```

Returning and checking errors:

```go
func makeTea(arg int) error {
    if arg == 42 {
        return ErrOutOfTea
    }
    return nil
}

err := makeTea(42)
if err != nil {
    println("got error")
}
if err == ErrOutOfTea {
    println("out of tea")
}
```

Errors are compared using `==`. This is an O(1) operation (compares pointers, not strings).

Dynamic errors (`fmt.Errorf`), local error variables (`errors.New` inside functions), and error wrapping are not supported.

The zero value of `error` is `nil` (`NULL` in C).

## Panic

`panic()` accepts a string literal, string variable, or error value and immediately terminates the program:

```go
panic("something went wrong")

msg := "runtime error"
panic(msg)

var err = errors.New("not found")
panic(err)
```

In C, this is emitted as a macro call `so_panic(...)`.

`recover` is not supported.

## Defer

`defer` calls a function with a single argument at the end of the enclosing function:

```go
func main() {
    x := 42
    xopen(&x)
    defer xclose(&x)
    println("working...")
}
```

The deferred function must accept exactly one argument. Multi-argument defer calls are not supported. As C2Y `defer` becomes more widely used, So might switch to it to support more complex `defer` statements.

## C interop

Include a C header file:

```go
//so:include "person.ext.h"
```

Declare a struct that is defined in C (excluded from emission):

```go
//so:extern
type Account struct {
    name    string
    balance int64
    flags   []uint8
}
```

Declare an external C function (no body or `so:extern`):

```go
func inc_balance(acc *Account, amount int64) int64

//so:extern
func dec_balance(acc *Account, amount int64) int64 {
    return 42 // for testing
}
```

When calling extern functions, `string` and `[]T` arguments are automatically decayed to their C equivalents: string literals become raw C strings (`"hello"`), string values become `char*` (`.ptr`), and slices become raw pointers (`.ptr`). This means C macros don't need to extract `.ptr` themselves:

```go
//so:extern
func Fopen(path string, mode string) *File { return nil }

// Go call:
f := Fopen("/tmp/test.txt", "w")

// Generated C:
// Fopen("/tmp/test.txt", "w")
// not Fopen(so_str("/tmp/test.txt"), so_str("w"))
```

The `so/c` package includes helpers for converting C pointers back to So string and slice types: `c.String(ptr)` and `c.Bytes(ptr, n)`. It also provides `c.CharPtr(ptr)` to cast a `*byte` (`uint8_t*`) to `char*` for C functions that expect `char*` (e.g. `strftime`).

## Embeds

Embed C files directly into the generated output using `//so:embed`:

```go
import _ "embed"

//so:embed main.h
var main_h string

//so:embed main.c
var main_c string
```

`.h` files are embedded into the generated header, `.c` files into the generated implementation. The embed variable declarations are not emitted as C variables — they serve as markers only.

## Packages

Each Go package is translated into a single `.h` + `.c` pair, regardless of how many `.go` files it contains. Multiple `.go` files in the same package are merged into one `.c` file, separated by `// -- filename.go --` comments.

Exported symbols (capitalized names) are prefixed with the package name:

```go
// geom/geom.go
package geom

const Pi = 3.14159

func RectArea(width, height float64) float64 {
    return width * height
}
```

Becomes:

```c
// geom.h
extern const double geom_Pi;
double geom_RectArea(double width, double height);

// geom.c
const double geom_Pi = 3.14159;
double geom_RectArea(double width, double height) { ... }
```

Unexported symbols (lowercase names) keep their original names and are marked `static`:

```c
static double rectArea(double width, double height);
```

Exported symbols are declared in the `.h` file (with `extern` for variables). Unexported symbols only appear in the `.c` file as forward declarations.

Importing a package translates to a C `#include`:

```go
import "example/geom"
```

```c
#include "geom/geom.h"
```

Calling imported symbols uses the package prefix:

```go
a := geom.RectArea(5, 10)
_ = geom.Pi
```

```c
double a = geom_RectArea(5, 10);
(void)geom_Pi;
```
