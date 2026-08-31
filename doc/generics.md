# Generics

Solod supports generic functions as [extern declarations](#generic-extern-declarations) and [inline macros](#generic-inline-macros), and also supports [generic types](#generic-types). However, these features are very limited, and you should only use generics for the simplest cases (or, even better — don't use them at all).

## Generic extern declarations

Generic type parameters are translated to C macro arguments, prepended before the regular arguments.

Generic extern functions:

```go
//so:extern
func newObj[T any]() *T {
    return nil
}

//so:extern
func freeObj[T any](ptr *T) {
}
```

Calling with explicit or inferred type arguments:

```go
v := newObj[int]()  // newObj(so_int)
freeObj(v)          // freeObj(so_int, v) - type argument inferred
```

Generic extern types:

```go
//so:extern
type Map[K comparable, V any] struct {
    // ...
}
```

Generic extern methods:

```go
//so:extern
func (m *Map[K, V]) Len() int {
    return m.len
}
```

Method calls prepend the receiver's type arguments:

```go
// go
m := newMap[string, int](10)
l := m.Len()
```

```c
// c
main_Map m = newMap(so_String, so_int, 10);
so_int l = main_Map_Len(so_String, so_int, &m);
```

On the C side, generic functions and methods should be manually implemented as macros that receive the type arguments:

```c
#define newObj(T) (alloca(sizeof(T)))
#define freeObj(T, ptr) ((void)(ptr))

#define newMap(K, V, size) ((main_Map){0})
#define main_Map_Len(K, V, m) ((m)->len)

typedef struct {
    // ...
} main_Map;
```

Constraints (`any`, `comparable`, etc.) are not emitted in C. See [Type constraints](#type-constraints).

## Generic inline macros

When `//so:inline` is applied to a generic function, the transpiler automatically generates a C `#define` macro instead of a `static inline` function. This is the primary mechanism for writing type-generic code directly in Solod without hand-writing C macros.

```go
//so:inline
func identity[T any](val T) T {
    return val
}
```

Produces:

```c
#define identity(T, val_) ({ \
    (val_); \
})
```

The macro is emitted in the `.h` file so it is available to all translation units.

**How it works.** The transpiler:

1. Collects all type parameters and prepends them as macro parameters.
2. Appends a `_` suffix to each non-type parameter name to avoid collisions with struct fields and other identifiers (e.g. `val` becomes `val_`).
3. Wraps references to those parameters in parentheses (`(val_)`) to prevent operator-precedence bugs when expressions are passed as arguments.
4. Uses a GCC/Clang statement expression `({ ... })` for functions that return a value, or `do { ... } while (0)` for void functions.
5. Translates `return expr` into a bare `expr;` (the last expression in a statement expression becomes its value).

**Generic methods** work the same way. The receiver's type parameters are included, and the receiver itself becomes a macro parameter:

```go
//so:extern
type Box[T any] struct {
    val T
}

//so:inline
func (b *Box[T]) set(val T) {
    b.val = val
}
```

Produces:

```c
#define main_Box_set(T, b_, val_) do { \
    (b_)->val = (val_); \
} while (0)
```

**Call-site translation.** Generic calls pass the resolved C type as the first argument:

```go
x := identity(42)       // identity(so_int, 42)
x := identity[int](42)  // identity(so_int, 42)
```

Type arguments can be explicit or inferred by the Go type checker - the C output is the same.

**Restrictions**

_No early returns_. A `return` inside a macro body does not return from the macro - it becomes a bare expression. Multiple returns (e.g. `if ... { return x } return y`) will not work correctly. Structure the logic so there is exactly one return at the end, or use local variables and conditionals to compute the result.

_Argument evaluation_. Arguments are not automatically assigned to temporaries. If an argument with side effects (e.g. `i++` or a function call) is referenced more than once in the macro body, it will be evaluated multiple times. Assign arguments to local variables at the top of the body to avoid this (see recommended practices below).

_No defer_. Deferred calls are not supported inside macro bodies.

_No control flow that escapes the macro_. `break`, `continue`, and `goto` in the caller's context cannot be used from within a macro body.

**Recommended practices**

_Prefix local variables with underscore_. The transpiler does not add a hygiene prefix to variables declared inside the macro body. If a local variable has the same name as one in the caller's scope, it will shadow or conflict. Using an underscore prefix (e.g. `_result`, `_key`) greatly reduces collision risk:

```go
//so:inline
func increment[T int](n T) T {
    _n := n        // copy argument to local to avoid double evaluation
    _n = _n + 1
    return _n
}
```

_Copy arguments to locals_. If a parameter is used more than once, assign it to a `_`-prefixed local at the top. This prevents double evaluation and makes the macro behave like a function:

```go
//so:inline
func (m *Map[K, V]) Has(key K) bool {
    _key := key    // evaluated once
    _m := m.bm
    // ... use _key and _m throughout ...
}
```

_Keep macros short_. Because every call site is expanded inline, large macros increase binary size and compile time. If a function is longer than ~15 lines, consider moving the heavy logic into a regular (non-generic) helper and using the macro only as a thin typed wrapper.

_Avoid macro call chains_. When one inline macro calls another, and that one calls a third, the preprocessor expands everything at the call site into a single deeply nested expression. This produces unreadable compiler errors, makes debugging nearly impossible, and can hit compiler limits. If you need a chain of calls, make the intermediate functions regular (non-inline) and only use `//so:inline` on the outermost typed wrapper.

_Single return at the end_. Structure the function so there is one `return` as the last statement:

```go
// Good - single return at the end
//so:inline
func max[T int](a, b T) T {
    _a := a
    _b := b
    _r := _a
    if _b > _a {
        _r = _b
    }
    return _r
}

// Bad - multiple returns (won't work correctly as a macro)
//so:inline
func max[T int](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

## Generic types

A type parameter never reaches the generated C. It exists for Go type-checking and as a macro argument, nothing else. A generic type declaration is therefore allowed only when its C representation does not depend on the type parameters:

```go
// OK: one C struct serves every instantiation.
type Stack[T any] struct {
    items []T          // so_Slice items;
    index map[string]T // so_Map* index;
}

// Rejected: there is no C type to emit for T.
type Pair[T any] struct{ a, b T }
type Ref[T any] struct{ p *T }
type List[T any] func(T)
```

Slices and maps are type-erased in C (`so_Slice`, `so_Map*`), so a `[]T` or `map[K]V` field carries no dependency on the type parameters. A field typed `T`, `*T`, or a function mentioning `T` does.

The usual pattern uses a phantom type parameter: the Go type is generic for type safety, the C struct doesn't include the type parameter, and the type parameter is used by `so:inline` methods:

```go
type Chan[T any] struct {
    buf *Buffer     // no field mentions T
    rdv *Rendezvous
}

//so:inline
func (ch *Chan[T]) Send(v T) {
    // T is a macro argument here
}
```

If the type is marked with `so:extern`, you can use type parameters in fields without any restrictions, because extern type definitions aren't emitted (they come from hand-written C code):

```go
// OK: type definiton comes from C.
//so:extern atomic_Pointer
type Pointer[T any] struct {
	v *T
}
```

The same rule covers functions and methods. A generic one must be `so:inline` or `so:extern`, because a regular C function has nowhere to declare a type parameter while its call sites still pass one:

```go
//so:inline
func (s *Stack[T]) First() T { ... } // OK: #define main_Stack_First(T, s_)

func (s *Stack[T]) First() T { ... } // rejected
```

This holds even when the signature itself never mentions `T`: the receiver's type parameters are still prepended at every call site.

## Type constraints

Constraints are used only for Go type-checking and are not emitted in C. This covers the predeclared ones (`any`, `comparable`) and constraint interfaces you declare yourself:

```go
// Neither of these produces any C.
type Number interface {
    ~int | ~float64
}

type ordered interface {
    comparable
    ~int | ~string
}
```

Go allows such an interface only as a type parameter constraint, never as a value type, so it has nothing to represent in C.
