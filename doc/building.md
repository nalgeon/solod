# Building

So programs are built with the `so build` command, which transpiles the package to C and compiles it with a system C compiler. The `run`, `test`, and `bench` commands build the same way before running. This guide covers the build options they share.

[Compiler and flags](#compiler-and-flags) •
[Target](#target) •
[Checks](#checks) •
[Panic mode](#panic-mode) •
[Assertions](#assertions) •
[Source locations](#source-locations)

POSIX targets are fully supported. Other targets have some restrictions. For more information, see these links:

- [Freestanding](freestanding.md) (bare-metal) targets.
- [Windows](windows.md).

## Compiler and flags

So invokes the compiler named by the `CC` environment variable (default `cc`) and passes along `CFLAGS` and `LDFLAGS`:

```sh
export CC=clang
export CFLAGS="-Ofast"
so build -o app .
```

`so build` writes the executable to the path given by `-o`, or to the package directory's basename if `-o` is omitted.

The default optimization level is `-O2`. A level from `CFLAGS` (like in the example above) replaces it.

## Target

The `-target` flag names the target of a cross-compilation. It takes the value that `clang` and `zig cc` accept after `--target=`, and So passes it through unchanged:

```sh
export CC="zig cc"
so build -target=x86_64-windows-gnu -o app.exe .
so build -target=riscv64-linux -o app .
so build -target=wasm32-freestanding -o app.wasm .
```

The value is `arch-os` or `arch-os-abi`. Run `zig targets` for the full list. `-target` needs `clang` or `zig cc`; GCC uses a separate cross compiler for each target, so it has no `--target` option.

## Checks

The `-check` flag turns on compile-time or run-time checking of the build. It is off by default:

```sh
so test -check=warn .        # warnings are errors
so test -check=sanitize .    # AddressSanitizer and UndefinedBehaviorSanitizer
so test -check=analyze .     # the GCC static analyzer
```

| Mode       | C compiler flags                                                                                 |
| ---------- | ------------------------------------------------------------------------------------------------ |
| `off`      | none (default)                                                                                   |
| `warn`     | `-Wall -Wextra -Werror -Wno-shadow -Wno-unused-label`                                            |
| `sanitize` | `warn`, plus `-g -fsanitize=address,undefined -fno-sanitize-recover=all -fno-omit-frame-pointer` |
| `analyze`  | `warn`, plus `-fanalyzer`                                                                        |

`sanitize` catches memory errors like out-of-bounds access and use-after-free at run time. It adds `-g` so the reports carry readable `file:line` stack traces. Pair it with `-panic=abort` to hand a failing check to the sanitizer's own reporter. It needs a hosted target, because the sanitizer runtimes need a C library.

`analyze` runs the GCC static analyzer, which reports at compile time. It needs GCC; `clang` has no `-fanalyzer`.

Use `CFLAGS` to customize which checks are enabled:

```sh
CFLAGS="-fsanitize=thread" so test .
```

## Panic mode

The `-panic` flag selects how a panic terminates the program after printing its message. It applies to `build`, `run`, `test`, and `bench`:

```
so run -panic=trace .
```

- `trace` (default): print a symbolized backtrace, then `exit(1)`.
- `exit`: call `exit(1)`. Clean, deterministic exit code.
- `abort`: call `abort()`, raising `SIGABRT` so a debugger, AddressSanitizer, or core dump can report the stack.

Trace mode adds `-rdynamic -fno-omit-frame-pointer` to the C build so frames can be unwound and named. The trace shows C symbols (`package_Func`), which map directly onto So functions; combine it with `-track-source` to relate the panic site back to So source.

## Assertions

Assertions check preconditions like slice bounds and index-out-of-range, plus the `c.Assert` calls the program makes itself. They are on by default. The `-assert` flag removes them for `build`, `run`, `test`, and `bench`:

```sh
so build -assert=off .
```

A removed assertion doesn't evaluate its condition at all, so conditions must be free of side effects. Only turn assertions off when you're sure the program is correct.

`NDEBUG` doesn't affect assertions, so a C project that defines it can't remove them by accident.

## Source locations

By default, panic messages report the C file and line number. Use the `-track-source` flag to print the original So source locations instead:

```
so build -track-source .
so run -track-source .
```

When `-track-source` is enabled, the reported source location may be off by a few lines for panics that occur inside complex statements (e.g., multi-line expressions or nested calls).
