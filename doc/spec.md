# So language description

Solod (So) is a strict subset of Go that transpiles to regular C. This document lists the features it supports. If a feature isn't listed, it's not supported.

[Values](#values) •
[Constants](#constants) •
[Variables](#variables) •
[Strings](#strings) •
[Arrays](#arrays) •
[Slices](#slices) •
[Maps](#maps) •
[If/else](#ifelse) •
[Switch](#switch) •
[For](#for) •
[Goto](#labels-and-goto) •
[Functions](#functions) •
[Multiple returns](#multiple-return-values) •
[Variadic functions](#variadic-functions) •
[Structs](#structs) •
[Methods](#methods) •
[Interfaces](#interfaces) •
[Any](#any) •
[Enums](#enums) •
[Errors](#errors) •
[Panic](#panic) •
[Defer](#defer) •
[C interop](#c-interop) •
[Generics](#generics) •
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

In C, the default type for integers is `so_int`, for floats it's `double`, and for runes it's `int32_t`.

Complex numbers are not supported.

## Constants

Constants are translated to C `const` qualifiers.

A constant integer expression is emitted as its value when C cannot compute it step by step. Expressions that C computes correctly are emitted as operators:

```go
var mask uint64 = 1<<64 - 1      // folds to 18446744073709551615u
var flags int64 = 1<<20 | 1<<10  // emits as is
```

A constant float expression is always emitted as its value, not as the operators used to produce it:

```go
const pi = 3.14159
const twoPi = 2 * pi  // folds to 6.28318
```

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

`byte` is translated to `so_byte` (`uint8_t`), `rune` to `so_rune` (`int32_t`), `int` to `so_int`, and `uint` to `so_uint`.

The width of `so_int` and `so_uint` equals the pointer width of the target, either 32 bits or 64 bits. Portable code must not assume more than 32 bits. Use `int64` or `uint64` for a value above 32 bits.

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

Strings are represented as `so_String` type in C:

```c
typedef struct {
    const char* ptr;
    so_int len;
} so_String;
```

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

Comparing strings (uses `memcmp`):

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
bs := []byte(s)  // zero-copy view of s
rs := []rune(s)  // allocates with alloca
```

Converting a byte or a rune slice to a string:

```go
s1 := string(bs)  // zero-copy view of bs
s2 := string(rs)  // allocates with alloca
```

`string([]byte)` and `[]byte(string)` are zero-copy views that alias the original data. Modifying the byte slice will affect the string and vice versa. Clone the data if you need an independent copy.

Converting a byte or rune to a string:

```go
var b byte = 'A'
s1 := string(b)  // "A"
var r rune = '世'
s2 := string(r)  // "世" (UTF-8 encoded)
```

String concatenation with `+` and `+=` is supported for both literals and variables. Adding string variables allocates memory on the stack, so avoid using them for large strings or strings that should be on the heap. Instead, use the `so/strings` package.

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

Arrays decay to pointers when passed to functions (no value semantics on calls).

Array assignment uses `memcpy`.

An array-typed element of a composite literal must be an array literal:

```go
b1 := Box{nums: [3]int{1, 2, 3}}  // ok
b2 := Box{nums: arr}              // error: use an array literal
var b3 Box; b3.nums = arr         // ok
```

Slices of arrays (`[][3]int`) are not supported. Use a slice of slices, or wrap
the array in a struct.

## Slices

Slices are represented as `so_Slice` type in C:

```c
typedef struct {
    void* ptr;
    so_int len;
    so_int cap;
} so_Slice;
```

Slice literals:

```go
strs := []string{"a", "b", "c"}
twoD := [][]int{{1, 2, 3}, {4, 5, 6}}
```

Unlike in Go, a nil slice and an empty slice are the same thing:

```go
// Both emit `(so_Slice){}`.
var nils []int = nil
var empty []int = []int{}
```

In general, the slice is nil if and only if its capacity is 0.

Slicing:

```go
s1 := nums[:]    // full slice
s2 := nums[2:]   // from index 2
s3 := nums[:3]   // up to index 3
s4 := nums[1:4]  // from 1 to 4
```

Full slice expressions (`s[low:high:max]`) are supported to limit the capacity of the resulting slice:

```go
s := nums[1:3:4]  // len=2, cap=3
```

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

`make()` allocates a fixed amount of memory on the stack (`sizeof(T)*cap`) with `alloca`. `append()` only works up to the initial capacity and panics if it's exceeded. There's no automatic reallocation. Use the `so/slices` package instead of `make` and `append` for heap allocation and dynamic arrays.

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

`clear` zeros all elements of a slice to their zero value. Length and capacity are unchanged.

## Maps

Maps are fixed-size and stack-allocated, backed by "mask-step-index" hashtables. They are pointer-based reference types, represented as `so_Map*` in C. No delete, no resize.

```c
typedef struct {
    void* keys;
    void* vals;
    so_int len;
    so_int cap;
} so_Map;
```

Only use maps when you have a small, fixed number of items (<1024). For anything else, use heap-allocated maps from the `so/maps` package.

Map literals:

```go
m1 := map[string]int{"a": 11, "b": 22}
m2 := map[int]string{11: "a", 22: "b"}
```

Creating a map with `make`:

```go
m := make(map[string]int, 2)
```

The capacity argument is required and determines the fixed size of the map. `make()` allocates key and value arrays on the stack with `alloca`.

Setting and getting values:

```go
m["a"] = 11
v := m["a"]
```

Comma-ok pattern to check if a key exists:

```go
v, ok := m["a"]
if !ok {
    println("not found")
}
```

If the key is not found, the value is the zero value for the element type and `ok` is `false`.

Iterating over a map with `range`:

```go
for k, v := range m {
    println(k, v)
}
```

Supported key types: all integer types, `bool`, `float32`, `float64`, `string`, and pointers.

A `nil` map emits as `NULL` in C.

Limitations:

- Maps have a fixed capacity set at creation time. Setting a key when the map is full panics.
- Compound assignment on map index (`m["a"] += 1`) is not supported.
- Arrays as map value types are not supported.
- `delete` is not supported.
- `clear` is not supported with maps.

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

## Switch

Switch statements are translated to if/else-if/else chains.

Tagged switch:

```go
switch x {
case 1:
    println("one")
case 2, 3:
    println("two or three")
default:
    println("other")
}
```

Tagless switch (bool conditions):

```go
switch {
case x > 100:
    println("big")
case x > 0:
    println("positive")
}
```

Init statement (scoped to the switch block):

```go
switch n := compute(); n {
case 42:
    println("answer")
}
```

The tag is evaluated once, before any case expression, and the case expressions are compared to it in order.

Not supported: switching on structs or arrays, type switches, `fallthrough`, and unlabeled `break` in a case body (use a labeled break).

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

Regular `break` and `continue` work as expected.

## Labels and goto

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

Labeled `break` in a loop works as expected:

```go
sum := 0
outer:
for i := range 5 {
    for j := range 5 {
        if i+j > 3 {
            break outer
        }
        sum += i + j
    }
}
```

Labeled `continue` is not supported.

## Functions

Regular function declarations:

```go
func sumABC(a, b, c int) int {
    return a + b + c
}
```

Named function types and function variables:

```go
type CalcFunc func(int) int
func calc(n int) int { return n*2 }

fn1 := calc              // infer type by signature
var fn2 CalcFunc = calc  // explicit type
n := fn2(7)
```

Anonymous function types can be used as variable types and function parameters:

```go
// func parameter
func apply(n int, f func(int) int) int { return f(n) }

// func variable
var fn func(int) int = calc
```

Anonymous function types are not supported as return types; use a named type
like `CalcFunc` there. Function literals (anonymous functions / closures) are not
supported either.

Exported functions (capitalized) become public C symbols prefixed with the package name (`package_Func`). Unexported functions are `static`.

Exported functions must only use exported types in their signatures (parameters and return types).

### Call order

Go evaluates the calls in an expression from left to right. So does not define the order because it passes each expression to the C compiler, and C does not specify the order of operands and arguments. As a result, two different C compilers might choose different orders for the same generated code:

```go
sum(next(), next())  // either call can run first
arr[next()] = next()
```

Write one call per statement when the order is important:

```go
x := next()
y := next()
sum(x, y)
```

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
bool byte float64
int int64 rune
string []T *T
```

So also supports the `(T, error)` pattern, where `T` is a custom struct type:

```go
func create(size int) (File, error) {
    return File{size: size}, nil
}
```

The compiler auto-generates `{T}Result` structs for these `T` types, so don't name your own types `{T}Result` if `T` is a struct type that's returned as `(T, error)`.

Automatic `{T}Result` generation for a custom `T` only works if the function returning `(T, error)` is defined in the same package as `T`, or if there's at least one function in `T`'s package that also returns `(T, error)`.

Otherwise, you'll have to manually define a struct type called `{T}Result` with two fields — `val T` and `err error`, like this:

```go
type FileResult struct {
    val File
    err error
}
```

Named return values are not supported.

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

Variadic methods work the same way:

```go
func (l *Logger) Info(msg string, attrs ...Attr) {
    // attrs is a []Attr slice
}

l.Info("hello", attr1, attr2)
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

Inner structs (anonymous struct fields):

```go
type Benchmark struct {
    name string
    loop struct {
        n int
        i int
    }
}

b := Benchmark{name: "Test", loop: struct{ n, i int }{n: 200, i: 10}}
b.loop.n = 100
```

Structs without fields (`type T struct{}`) are supported.

Anonymous structs are only supported as local variables (the `dog` example) and as inner struct fields (the `Benchmark` example). In other cases — slice/array elements, params, returns — use a named type instead.

Embedded fields are not supported; declare named fields instead.

Struct comparison (`==`, `!=`) is not supported.

`new()` works with types and values:

```go
n := new(int)           // *int, zero-initialized
p := new(point)         // *point, zero-initialized
n2 := new(42)           // *int with value 42
p2 := new(point{1, 2})  // *point with values
```

## Methods

Methods are defined on struct types with pointer or value receivers:

```go
type Rect struct {
    width, height int
}

func (r *Rect) Area() int {
    return r.width * r.height
}

func (r Rect) resize(x int) Rect {
    r.height *= x
    r.width *= x
    return r
}
```

A method translates to a regular function in C; the receiver is passed as the first argument. Pointer receivers are passed as `void*`, value receivers are passed as a typed value:

```c
so_int main_Rect_Area(void* self)
static main_Rect main_Rect_resize(main_Rect r, so_int x)
```

Calling methods on values and pointers:

```go
r := Rect{width: 10, height: 5}
r.Area()      // called on value (address taken automatically)
r.resize(2)   // called on value (passed by value)

rp := &r
rp.Area()     // called on pointer
rp.resize(2)  // called on pointer (dereferenced automatically)
```

Methods on named primitive types are also supported:

```go
type HttpStatus int

func (s HttpStatus) String() string {
    // ...
}
```

Method expressions (`T.method` or `(*T).method`) are supported. Method values (`v.method`) are not supported.

## Interfaces

Interfaces in So are like Go interfaces, but they don't include runtime type information.

Interface declarations list the required methods:

```go
type Shape interface {
    Area() int
    Perim(n int) int
}
```

In C, an interface is a struct with a `void* self` pointer and function pointers for each method (less efficient than using a static method table, but simpler; this might change in the future).

```c
typedef struct main_Shape {
    void* self;
    so_int (*Area)(void* self);
    so_int (*Perim)(void* self, so_int n);
} main_Shape;
```

Interface methods on concrete types must use pointer receivers, since the vtable uses `void* self` function pointers. If a concrete type uses value receivers, converting it to an interface will fail:

```go
func (r Rect) Area() int { // value receiver
    return r.width * r.height
}

var s Shape = r // rejected: method Rect.Area has a value receiver
```

Using pointer receivers works fine:

```go
func (r *Rect) Area() int { // pointer receiver
    return r.width * r.height
}
```

Converting a concrete type to an interface requires passing a pointer:

```go
r := Rect{2, 4}
s := Shape(&r)
var s2 Shape = &r
```

Passing a concrete type to functions that accept interfaces:

```go
func calcShape(s Shape) int {
    return s.Perim(2) + s.Area()
}

calcShape(&r)         // implicit conversion
calcShape(Shape(&r))  // explicit conversion
```

Type assertions:

```go
_, ok := s.(*Rect)    // comma-ok pattern (checks without panic)
r := s.(*Rect)        // direct assertion

// But not both; this is not supported.
// r, ok := s.(*Rect)
```

Empty interfaces (`interface{}` and `any`) are translated to `void*`.

Converting between named interfaces is not supported: no type assertions like `iface.(AnotherIface)` and no type switches.

Embedded interfaces are not supported; list the methods explicitly instead.

Two interfaces are equal when they hold the same pointer, and an interface compares with `nil` as expected. Comparing an interface with a concrete type is not supported:

```go
var s Shape = &r
if s == nil { }   // supported
if s == other { } // supported, other is a Shape
if s == &r { }    // not supported
```

## Any

`any` is not implemented as a regular interface. Instead, it's translated to `void*`.

An `any` can hold any value:

```go
var a any // in C: void* a = NULL

// Primitive value.
var n int = 42
a = n // in C: a = &n

// String or slice.
var s string = "hello"
a = s // in C: a = &s

// Struct value.
var r Rect = Rect{5, 10}
a = r // in C: a = &r

// Pointer.
var rp *Rect = &Rect{5, 10}
a = rp // in C: a = rp

// Unsafe pointer (also a pointer, so it is stored as is).
var up = unsafe.Pointer(rp)
a = up // in C: a = up

// Named interface value.
var sh Shape = &r
a = sh // in C: a = &sh
```

If an `any` holds a named interface value, it can be asserted back to that interface:

```go
var r1 *Rect = &Rect{5, 10}
var sh1 Shape = r1
a = sh1
sh2 := a.(Shape) // works fine
r2 := a.(*Rect)  // DO NOT do this
```

Because `any` carries no runtime type information, the assertion `a.(Shape)` is unchecked — it trusts that `a` holds a `Shape`. Unlike Go, you should never assert `a.(*Rect)`; doing so will give you an incorrectly typed pointer. Once an interface is boxed into an `any`, you have to assert it back to the interface type (`Shape`), not the concrete pointer type inside it (`*Rect`).

An `any` compares (`==`, `!=`) with `nil`, with a pointer, or with another `any`. Comparing it with a value is not supported: an `any` holds the address of the value, not the value itself:

```go
var a any = n
if a == nil { }  // supported
if a == &n { }   // supported, &n is a pointer
if a == n { }    // not supported
```

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

Each constant is emitted as a C `const`.

`iota` is supported for integer- and float-typed constants:

```go
type Day int

const (
    Sunday Day = iota
    Monday
    Tuesday
)
```

Iota values are evaluated at compile time and translated to numeric literals.

## Errors

The `error` type is a regular interface with an `Error() string` method. In C, it is represented as `so_Error` an interface struct, following the same pattern as other named interfaces:

```c
typedef struct {
    void* self;
    so_String (*Error)(void* self);
} so_Error;
```

Use `errors.New` to create sentinel errors at the package level:

```go
import "solod.dev/so/errors"

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

func main() {
    err := makeTea(42)
    if err != nil {
        println("got error")
    }
    if err == ErrOutOfTea {
        println("out of tea")
    }
}
```

Errors are compared using `==`. This is an O(1) operation (compares `.self` pointers, not strings).

Dynamic errors (`fmt.Errorf`), local error variables (`errors.New` inside functions), and error wrapping are not supported.

The zero value of `error` is `nil` (`{}` in C).

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

A panic prints its message with a source location, then terminates the program. How it terminates (the panic mode), how traces are symbolized, and how to report the original So source location are build options; see [Building](building.md).

`recover` is not supported.

## Defer

`defer` schedules a function or method call to run at the end of the function:

```go
func main() {
    xopen(&state)
    defer xclose(&state)
    println("working...")
    // xclose(&state) runs here
}
```

Deferred calls are emitted inline (before returns, panics, and function end) in LIFO order. The return value is evaluated before the deferred calls run.

Defer can only use variables declared at the top level of a function, not inside nested scopes like bare blocks, `for`, or `if`.

## C interop

So provides several tools for easy C interop. See the [interop guide](./interop.md) for details.

## Generics

So supports generic functions as extern declarations and inline macros, and also supports generic types. However, these features are very limited, and you should only use generics for the simplest cases (or, even better — don't use them at all). See the [generics guide](./generics.md) for details.

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

You can promote an unexported symbol to the header using `//so:promote`, which also gives it the package prefix. This is necessary when an exported inline function or type needs to reference an unexported symbol. See the [interop guide](./interop.md) for details.

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

Constants and variables are emitted in source order, so a constant or variable can't refer to one that's declared after it:

```go
const a = b // won't compile: b is declared below
const b = 1
```

Types are emitted in dependency order, so a type can refer to a type declared later in the source:

```go
type Rect struct {
    Min, Max Point // Point is declared below
}

type Point struct {
    X, Y int
}
```

A recursive type only works if the cycle goes through a struct, because that's what C forward declarations support. For example, `type Node struct { next *Node }` is allowed, but `type StateFn func() StateFn` and `type Tree [2]*Tree` are not.

### Init functions

Each package can have an `init()` function (with no arguments or return values) that runs automatically before `main()`. Unlike Go, only one `init` function is allowed per package.

Init functions can be used to initialize package-level variables with non-static values.

```go
var state int

func init() {
    state = 42
}
```

If the program has multiple packages, each with its own `init` function, the order in which the `init` functions are called is not guaranteed.
