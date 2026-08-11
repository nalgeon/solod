# Freestanding mode

So can target freestanding (bare-metal) environments where no C standard library is available.

## Compiling

Set `CC` and `CFLAGS` to target a freestanding environment. For example, using `zig cc` to target bare `wasm32`:

```sh
export CC="zig cc"
export CFLAGS="-Oz --target=wasm32-freestanding -nostdlib -Wl,--no-entry -Wl,--export=main"
so build -o main.wasm .
```

Or transpile to C first and compile separately:

```sh
so translate -o generated .
zig cc -Oz \
    --target=wasm32-freestanding \
    -nostdlib \
    -Wl,--no-entry \
    -Wl,--export=main \
    -o main.wasm \
    generated/**/*.c
```

## Limitations

### Bump allocator

In freestanding mode, `mem.System` is implemented as a simple bump allocator backed by a static buffer. It's off by default, but you can enable it by setting the heap size with `-DSO_HEAP_SIZE=<bytes>` at compile time.

In this implementation `free` is a no-op; memory is never reclaimed. `realloc` allocates a new bump region and copies data from the old one; the old region is not freed.

It's best not to use `mem.System` in freestanding mode. Instead, use `mem.Arena` so you can control the heap size and reset it when needed.

### Deterministic random

`runtime.Seed` uses a deterministic generator with a fixed initial state, instead of getting randomness from the operating system. Each call returns a different value, but the sequence is always the same every time you run the program.

Packages that depend on `runtime.Seed` (like `math/rand`) work but produce repeatable output.

### No stdio

`panic` silently traps instead of printing a message. `print` and `println` are no-ops.

## Stdlib packages

These packages work in freestanding mode with no restrictions:

```text
bufio  bytes  bytealg  c  cmp  encoding/binary  encoding/hex
errors  io  maps  math/bits  math/rand  mem  path runtime
slices  strconv  strings  unicode unicode/utf8  unsafe
```

The `time` package works with these restrictions:

- `Now`, `Since`, and `Until` are not available.
- `Time.Format` and `Time.Parse` only support named layouts (such as `RFC3339` or `DateOnly`), not custom layouts.

These packages require a hosted environment and will produce a compile-time error if imported:

```text
crypto/crand  flag  fmt  log/slog  math  os  testing
```
