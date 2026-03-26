# Solod: Go can be a better C

Solod (**So**) is a strict subset of Go that translates to regular C — with zero runtime, manual memory management, and source-level interop.

Highlights:

- Go in, C out. You write regular Go code and get readable C11 as output.
- Zero runtime. No garbage collection, no reference counting, no hidden allocations.
- Everything is stack-allocated by default. Heap is opt-in through the standard library.
- Native C interop. Call C from So and So from C — no CGO, no overhead.
- Go tooling works out of the box — syntax highlighting, LSP, linting and "go test".

So supports structs, methods, interfaces, slices, multiple returns, and defer. To keep things simple, there are no channels, goroutines, closures, or generics.

So is for systems programming in C, but with Go's syntax, type safety, and tooling.

[Example](#example) •
[Installation and usage](#installation-and-usage) •
[Language tour](doc/spec.md) •
[Stdlib](doc/stdlib.md) •
[Playground](https://codapi.org/so/) •
[So by example](example/README.md) •
[Testing](#testing) •
[Compatibility](#compatibility) •
[Design decisions](#design-decisions) •
[FAQ](#frequently-asked-questions) •
[Roadmap](#roadmap) •
[Contributing](#contributing)

## Example

This Go code in a file `main.go`:

```go
package main

type Person struct {
    Name string
    Age  int
    Nums [3]int
}

func (p *Person) Sleep() int {
    p.Age += 1
    return p.Age
}

func main() {
    p := Person{Name: "Alice", Age: 30}
    p.Sleep()
    println(p.Name, "is now", p.Age, "years old.")

    p.Nums[0] = 42
    println("1st lucky number is", p.Nums[0])
}
```

Translates to a header file `main.h`:

```c
#pragma once
#include "so/builtin/builtin.h"

typedef struct main_Person {
    so_String Name;
    so_int Age;
    so_int Nums[3];
} main_Person;

so_int main_Person_Sleep(void* self);
```

Plus an implementation file `main.c`:

```c
#include "main.h"

so_int main_Person_Sleep(void* self) {
    main_Person* p = (main_Person*)self;
    p->Age += 1;
    return p->Age;
}

int main(void) {
    main_Person p = (main_Person){.Name = so_str("Alice"), .Age = 30};
    main_Person_Sleep(&p);
    so_println("%.*s %s %" PRId64 " %s", p.Name.len, p.Name.ptr, "is now", p.Age, "years old.");
    p.Nums[0] = 42;
    so_println("%s %" PRId64, "1st lucky number is", p.Nums[0]);
}
```

Check out more examples in [So by example](example/README.md) and learn about the supported language features in the [language tour](doc/spec.md).

## Installation and usage

Install the So command line tool:

```
go install solod.dev/cmd/so@latest
```

Create a new Go project and add the Solod dependency to use the So standard library:

```
go mod init example
go get solod.dev@latest
```

Write regular Go code, but use So packages instead of the standard Go packages:

```go
package main

import "solod.dev/so/c/math"

func main() {
    ans := math.Sqrt(1764)
    println("Hello, world! The answer is", int(ans))
}
```

Transpile to C:

```
so translate -o generated .
```

The translated C code will be saved in the `generated` directory.

You can also transpile to C and compile the code to a binary in one step. This uses the C compiler set by the `CC` environment variable:

```
so build -o main .
```

Or you can transpile, compile, and run without saving the binary:

```
so run .
```

All commands work with Go modules, not individual files (`so run .`, not `so run main.go`).

Keep in mind that So is new, so it's still a bit rough around the edges.

## Language tour

To learn about So's features and limitations, check out the brief [overview of the language](./doc/spec.md).

## Standard library

So provides a growing set of [high-level packages](./doc/stdlib.md) similar to Go's stdlib, as well as low-level packages that wrap the libc API.

## Playground

[Try So online](https://codapi.org/so/) without installing anything. You can run the code or view the translated C output.

## So by example

If you like learning by doing, try a [hands-on introduction](./example/README.md) to So with annotated example programs.

## Testing

So doesn't have its own testing framework. Since So code is valid Go code, you can just use `go test` like you normally would. Plus, your tests can use all Go features because they're never transpiled.

The transpilation logic is covered by the So compiler's own tests.

## Compatibility

So generates C11 code that relies on several GCC/Clang extensions:

- Binary literals (`0b1010`) in generated code.
- Statement expressions (`({...})`) in macros.
- `__attribute__((constructor))` for package-level initialization.
- `__auto_type` for local type inference in generated code.
- `__typeof__` for type inference in generic macros.
- `alloca` for `make()` and other dynamic stack allocations.

You can use GCC, Clang, or `zig cc` to compile the transpiled C code. MSVC is not supported.

Supported operating systems: Linux, macOS, and Windows (core language only).

## Design decisions

So is highly opinionated.

**Simplicity is key**. Fewer features are always better. Every new feature is strongly discouraged by default and should be added only if there are very convincing real-world use cases to support it. This applies to the standard library too — So tries to export as little of Go's stdlib API as possible while still remaining highly useful for real-world use cases.

**No heap allocations** are allowed in language built-ins (like maps, slices, new, or append). Heap allocations are allowed in the standard library, but they must clearly state when an allocation happens and who owns the allocated data.

**Fast and easy C interop**. Even though So uses Go syntax, it's basically C with its own standard library. Calling C from So, and So from C, should always be simple to write and run efficiently. The So standard library (translated to C) should be easy to add to any C project.

**Readability**. There are several languages that claim they can transpile to readable C code. Unfortunately, the C code they generate is usually unreadable or barely readable at best. So isn't perfect in this area either (though it's arguably better than others), but it aims to produce C code that's as readable as possible.

**Go compatibility**. So code is syntactically valid Go code, with no exceptions. Semantics may differ.

Non-goals:

**Raw performance**. You can definitely write C code by hand that runs faster than code produced by So. Also, some features in So, like interfaces, are currently implemented in a way that's not very efficient, mainly to keep things simple.

**Hiding C entirely**. So is a cleaner way to write C, not a replacement for it. You should know C to use So effectively.

**Go feature parity**. Less is more. Iterators aren't coming, and neither are generic methods.

## Frequently asked questions

_Why not Rust/Zig/Odin/other language?_

Because I like C and Go.

_Why not TinyGo?_

TinyGo is lightweight, but it still has a garbage collector, a runtime, and aims to support all Go features. What I'm after is something even simpler, with no runtime at all, source-level C interop, and eventually, Go's standard library ported to plain C so it can be used in regular C projects.

_How does So handle memory?_

Everything is stack-allocated by default. There's no garbage collector or reference counting. The standard library provides explicit heap allocation in the `so/mem` package when you need it.

_Is it safe?_

So itself has few safeguards other than the default Go type checking. It will panic on out-of-bounds array access, but it won't stop you from returning a dangling pointer or forgetting to free allocated memory.

Most memory-related problems can be caught with AddressSanitizer in modern compilers, so I recommend enabling it during development by adding `-fsanitize=address` to your `CFLAGS`.

_Can I use So code from C (and vice versa)?_

Yes. So compiles to plain C, therefore calling So from C is just calling C from C. Calling C from So is equally straightforward — see the language tour for details.

_Can I compile existing Go packages with So?_

Not really. Go uses automatic memory management, while So uses manual memory management. So also supports far fewer features than Go. Neither Go's standard library nor third-party packages will work with So without changes.

_How stable is this?_

Not for production at the moment.

_Where's the standard library?_

There is a growing set of high-level packages (`so/bytes`, `so/mem`, `so/slices`, ...). There are also low-level packages that wrap the libc API (`so/c/stdlib`, `so/c/stdio`, `so/c/cstring`, ...). Check out the standard library overview for more details.

## Roadmap

✅ Transpiler with basic Go features.

✅ Low-level stdlib (libc wrappers). Done for now; I will add more if needed.

⏳ Core stdlib packages: fmt, io, strings, time, ...

⏳ Maps.

⬜ Hardened transpiler.

⬜ Real-world examples.

⬜ More stdlib packages: crypto, http, json, ...

⬜ Full Windows support.

## Contributing

Bug fixes are welcome. For anything other than bug fixes, please open an issue first to discuss your proposed changes. To prevent feature bloat, it's important to discuss any new features before adding them.

AI-assisted submissions are fine on one condition: you, the human, have read all the code and fully understand what it does. Code reviewed only by another AI will not suffice.

Make sure to add or update tests as needed.

## License

Go stdlib code by the [Go Authors](https://github.com/golang/go).

Transpiler and So stdlib code by [Anton Zhiyanov](https://antonz.org/).

Released under the BSD 3-Clause License.
