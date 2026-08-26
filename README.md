# ☀️ Solod: Go can be a better C

Solod (**So**) is a strict subset of Go that translates to regular C.

Highlights:

- Go in, C out. You write regular Go code and get readable C11 as output.
- Zero runtime. No garbage collection, no reference counting, no hidden allocations.
- Rich standard library. Use familiar types and functions ported from Go's stdlib.
- Native C interop. Call C from So and So from C — no CGO, no overhead.
- Go tooling works out of the box. Syntax highlighting, LSP, linting and modules.

So supports structs, methods, interfaces, slices, maps, multiple returns, and defer. Everything is stack-allocated by default; heap is opt-in through the standard library. There is limited support for generics, and concurrency is provided by the standard library instead of being built into the language.

So is for Go developers who want systems-level control without learning a new language. And for C programmers who like Go's safety, structure, and tooling.

[Example](#example) •
[Install](#installation) •
[Use](#usage) •
[Learn](#learn-the-language) •
[Build&test](#build-and-test) •
[FAQ](#frequently-asked-questions) •
[Interop](#interop) •
[Compatibility](#compatibility) •
[Roadmap](#roadmap) •
[Contributing](#contributing)

## Example

This Go code in a file `main.go`:

```go
package main

import (
    "solod.dev/so/conc"
    "solod.dev/so/mem"
    "solod.dev/so/sync/atomic"
)

// Account is a thread-safe money account.
type Account struct {
    Balance atomic.Int64
}

// Deposit adds an amount to the balance.
func (a *Account) Deposit(amount int64) {
    a.Balance.Add(amount)
}

// pay deposits $10 into the shared account.
func pay(arg any) {
    acc := arg.(*Account)
    acc.Deposit(10)
}

func main() {
    var acc Account

    // Run 100 payments across 4 worker threads.
    opts := conc.PoolOptions{NumThreads: 4}
    pool := conc.NewPool(mem.System, opts)
    defer pool.Free()
    for range 100 {
        pool.Go(pay, &acc)
    }
    pool.Wait()

    println("balance is", acc.Balance.Load())
}
```

Translates to a header file `main.h`:

```c
#pragma once
#include "so/builtin/builtin.h"
#include "so/conc/conc.h"
#include "so/mem/mem.h"
#include "so/sync/atomic/atomic.h"

// Account is a thread-safe money account.
typedef struct main_Account {
    atomic_Int64 Balance;
} main_Account;

// Deposit adds an amount to the balance.
void main_Account_Deposit(void* self, int64_t amount);
```

Plus an implementation file `main.c`:

```c
#include "main.h"

// Deposit adds an amount to the balance.
void main_Account_Deposit(void* self, int64_t amount) {
    main_Account* a = self;
    atomic_Int64_Add(&a->Balance, amount);
}

// pay deposits $10 into the shared account.
static void pay(void* arg) {
    main_Account* acc = (main_Account*)arg;
    main_Account_Deposit(acc, 10);
}

int main(void) {
    main_Account acc = {0};
    // Run 100 payments across 4 worker threads.
    conc_PoolOptions opts = (conc_PoolOptions){.NumThreads = 4};
    conc_Pool* pool = conc_NewPool(mem_System, opts);
    for (so_int _i = 0; _i < 100; _i++) {
        conc_Pool_Go(pool, pay, &acc);
    }
    conc_Pool_Wait(pool);
    so_println("%s %" PRId64, "balance is", atomic_Int64_Load(&acc.Balance));
    conc_Pool_Free(pool);
    return 0;
}
```

Check out more examples in [So by example](https://github.com/solod-dev/example) and learn about the supported language features in the [language tour](doc/spec.md).

## Installation

Install the So command line tool:

```
go install solod.dev/cmd/so@latest
```

Create a new Go project and add the Solod dependency to use the So standard library:

```
go mod init example
go get solod.dev@latest
```

Use `main` or a specific commit hash instead of `latest` to install the newest development version, not the latest tagged release.

## Usage

Write regular Go code, but use So packages instead of the standard Go packages:

```go
package main

import "solod.dev/so/math"

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

You can also transpile to C and compile the code to a binary in one step. This uses the C compiler set by the `CC` environment variable (default `cc`):

```
so build -o main .
```

Or you can transpile, compile, and run without saving the binary:

```
so run .
```

All commands work with Go modules, not individual files (`so run .`, not `so run main.go`).

## Learn the language

**[Language tour](./doc/spec.md)**. To learn about So's features and limitations, check out the brief overview of the language.

**[Standard library](./doc/stdlib.md)**. So provides a growing set of packages similar to Go's stdlib.

**[Playground](https://codapi.org/so/)**. Try So online without installing anything. You can run the code or view the translated C output.

**[So by example](https://github.com/solod-dev/example)**. If you like learning by doing, try a hands-on introduction to So with annotated example programs.

## Build and test

**[Building](./doc/building.md)**. Compile So programs with `so build` and manage safety features.

**[Testing](doc/testing.md)**. Write tests with the `so test` command and the `so/testing` package. Since So code is also valid Go code, you can still use `go test` where it fits — those tests are never transpiled, so they can use all Go features.

**[AI skill](https://github.com/solod-dev/ai)**. You can have a clanker write So code for you. But where's the fun in that?

## Frequently asked questions

**[Principles](doc/design.md)**. So is highly opinionated. Simplicity is key. Heap allocations are explicit. Strictly Go syntax.

**[Benchmarks](doc/benchmarks.md)**. So truly shines when it comes to C interop, but it's also quite fast on regular Go code — typically on par with or faster than Go.

**[FAQ](doc/faq.md)**. I have heard these several times, so it's worth answering.

## Interop

C interop in So is easy to write and has zero runtime overhead. See the [interop](doc/interop.md) guide for details.

So offers native bindings for popular C libraries like [libuv](https://github.com/solod-dev/libuv), [raylib](https://github.com/solod-dev/raylib), [sodium](https://github.com/solod-dev/sodium), and [sqlite](https://github.com/solod-dev/sqlite).

To automatically generate bindings for any C library from its headers, use the [sobind](https://github.com/solod-dev/sobind) tool.

## Compatibility

So generates C11 code that relies on several GCC/Clang extensions:

- Binary literals (`0b1010`) in generated code.
- Statement expressions (`({...})`) in macros.
- `__attribute__((constructor))` for package-level initialization.
- `__auto_type` for local type inference in generated code.
- `__typeof__` for type inference in generic macros.
- `alloca` and VLAs for `make()` and other dynamic stack allocations.

Supported compilers: GCC, Clang, Emscripten, and `zig cc`. MSVC is not supported.

Supported operating systems: Linux, BSDs, macOS, and [Windows](doc/windows.md).

Supported platforms: amd64, arm64, riscv64, i386, and wasm32.

So can also target [freestanding](doc/freestanding.md) environments.

## Roadmap

✓ [v0.1](https://github.com/solod-dev/solod/releases/tag/v0.1.0) —
Core language features and stdlib packages.

✓ [v0.2](https://github.com/solod-dev/solod/releases/tag/v0.2.0) —
Networking, uuids and more targets: WebAssembly, 32-bit, freestanding.

✓ [v0.3](https://github.com/solod-dev/solod/releases/tag/v0.3.0) — Concurrency, JSON, testing and benchmarking, more safety (escape analysis, leak checking, stack traces).

⏳ [v0.4](./doc/changelog.md#solod-04-in-progress) — work in progress.

Future plans:

- High-level JSON API.
- HTTP.
- SQL.

## Contributing

Bug fixes are welcome. For anything other than bug fixes, please open an issue first to discuss your proposed changes. To prevent feature bloat, it's important to discuss any new features before adding them.

AI-assisted submissions are fine on one condition: you, the human, have read all the code and fully understand what it does. Code reviewed only by another AI will not suffice.

Make sure to add or update tests as needed.

## License

Go stdlib code by the [Go Authors](https://github.com/golang/go).

Transpiler and So stdlib code by [Anton Zhiyanov](https://antonz.org/).

Released under the BSD 3-Clause License.
