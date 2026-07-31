# Building

So programs are built with the `so build` command, which transpiles the package to C and compiles it with a system C compiler. The `run`, `test`, and `bench` commands build the same way before running. This guide covers the build options they share.

## Compiler and flags

So invokes the compiler named by the `CC` environment variable (default `cc`) and passes along `CFLAGS` and `LDFLAGS`:

```sh
export CC=clang
export CFLAGS="-O2"
so build -o app .
```

`so build` writes the executable to the path given by `-o`, or to the package directory's basename if `-o` is omitted.

For freestanding (bare-metal) targets, see [Freestanding mode](freestanding.md).

## Panic mode

The `-panic` flag selects how a panic terminates the program after printing its message. It applies to `build`, `run`, `test`, and `bench`:

```
so run -panic=trace .
```

- `trace` (default): print a symbolized backtrace, then `exit(1)`.
- `exit`: call `exit(1)`. Clean, deterministic exit code.
- `abort`: call `abort()`, raising `SIGABRT` so a debugger, AddressSanitizer, or core dump can report the stack.

Trace mode adds `-rdynamic -fno-omit-frame-pointer` to the C build so frames can be unwound and named. The trace shows C symbols (`package_Func`), which map directly onto So functions; combine it with `-track-source` to relate the panic site back to So source. The default fits glibc and macOS. Use `-panic=exit` or `-panic=abort` on musl, where the trace comes out empty, and on freestanding, which always traps.

`-rdynamic` is an ELF-only option that the MinGW driver rejects, and So omits it when `CC` looks like a MinGW compiler (`CC=x86_64-w64-mingw32-gcc`). Trace mode still builds for Windows, but the trace comes out empty there: MinGW has no `execinfo.h`.

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

## Sanitizers

The `-sanitize` flag turns on C sanitizers for a build so memory errors like out-of-bounds access, use-after-free, and undefined behavior are caught at runtime:

```
so test -sanitize .              # address,undefined
so test -sanitize=address .      # a specific set
```

Bare `-sanitize` enables `address,undefined`; passing a comma-separated list selects a specific set. The flag also adds `-g` and `-fno-omit-frame-pointer` so reports carry readable `file:line` stack traces. Pair `-sanitize` with `-panic=abort` to hand a failing check straight to the sanitizer's own reporter.
