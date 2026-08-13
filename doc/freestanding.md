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

`fmt.Print`, `fmt.Println`, and `fmt.Printf` format the text and then drop the bytes, because there is no standard output to write them to. Assign another writer to `fmt.Output` — a UART or a host import — to get the output back.

## Stdlib packages

These packages work in freestanding mode with no restrictions:

```text
bufio  bytealg  bytes  c  cmp  encoding  encoding/binary
encoding/hex  encoding/json  errors  io  maps  math/bits  math/rand
mem  path  runtime  slices  strconv  strings  sync/atomic  unicode
unicode/utf8  unsafe
```

The `fmt` package works with these restrictions:

- `Scanf`, `Sscanf`, and `Fscanf` read through the stdio of the host, so a freestanding call panics.
- `Print`, `Println`, and `Printf` drop the bytes unless you set `fmt.Output`, as [No stdio](#no-stdio) describes. `Sprintf` and `Fprintf` are unaffected.

The `net/netip` package works with one restriction:

- A zone given as an interface name (`fe80::1%eth0`) resolves to no zone, because a freestanding environment has no network interfaces. A numeric zone (`fe80::1%2`) works.

The `time` package works with these restrictions:

- `Now`, `Since`, and `Until` are not available.
- `Time.Format` and `Time.Parse` only support named layouts (such as `RFC3339` or `DateOnly`), not custom layouts.

These packages require a hosted environment and will produce a compile-time error if imported:

```text
conc  crypto/crand  flag  log/slog  math  net  os  sync  testing  uuid
```
