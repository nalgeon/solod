# Freestanding mode

So can target freestanding (bare-metal) environments where no C standard library is available.

[Compiling](#compiling) •
[Stdlib packages](#stdlib-packages) •
[Target hooks](#target-hooks) •
[Limitations](#limitations)

## Compiling

Set `CC` and `CFLAGS` to target a freestanding environment, and pass `-freestanding`. For example, using `zig cc` to target bare `wasm32`:

```sh
export CC="zig cc"
export CFLAGS="-Oz --target=wasm32-freestanding -nostdlib -Wl,--no-entry -Wl,--export=main"
so build -freestanding -o main.wasm .
```

`CFLAGS` names the target. The `-freestanding` flag settles the hosted or freestanding question:

- It passes `-ffreestanding` to the C compiler. The compiler then sets `__STDC_HOSTED__` to 0, and the stdlib takes its freestanding branch.
- It drops the libc dependencies declared with `so:link` in the standard library. It does not affect the libraries declared with `so:link` in user code.
- It drops the panic trace flags, which a freestanding build has no use for.

`so test`, `so bench` and `so run` take the same flag. `so translate` does not, because it never invokes the C compiler.

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

## Stdlib packages

### Fully freestanding

These packages work in freestanding mode with no restrictions:

```text
bufio  bytealg  bytes  c  cmp  encoding  encoding/binary
encoding/hex  encoding/json  errors  io  maps  math/bits  math/rand
mem  path  runtime  slices  strconv  strings  unicode
unicode/utf8  unsafe
```

### Freestanding with restrictions

**crypto/crand**

A freestanding environment has no CSPRNG, so the package draws from the `so_crand_read` hook. See [Target hooks](#target-hooks).

**fmt**

`Scanf`, `Sscanf`, and `Fscanf` read through the stdio of the host, so a freestanding call panics.

`Print`, `Println`, and `Printf` write through the `so_write_out` hook (see [Target hooks](#target-hooks)). A target with no hook drops the bytes. `Sprintf` and `Fprintf` are unaffected.

**math**

Only the part that needs no libm works. `Abs`, `Copysign`, `Dim`, `Inf`, `IsInf`, `IsNaN`, `Max`, `Min`, `NaN`, `Pow10`, `RoundToEven`, `Signbit`, the `Float64bits` family and every constant work. Every other function panics.

**net/netip**

A zone given as an interface name (`fe80::1%eth0`) resolves to no zone, because a freestanding environment has no network interfaces. A numeric zone (`fe80::1%2`) works.

**sync/atomic**

An atomic operation needs a lock-free instruction of its own width. A target with no such instruction calls libatomic instead. A freestanding build doesn't link it, so linking fails with an undefined `__atomic_*` symbol.

The available widths depend on the target:

| Target                                           | Types that work                      | Types that fail   |
| ------------------------------------------------ | ------------------------------------ | ----------------- |
| wasm32, x86-64, ARM64, RV64 with the A extension | all                                  | none              |
| ARMv7-M, ARMv8-M, RV32 with the A extension      | `Bool`, `Int32`, `Uint32`, `Pointer` | `Int64`, `Uint64` |
| ARMv6-M, RISC-V without the A extension          | none                                 | all               |

On ARM, the M profile is the restricted one. ARMv6, ARMv7-A and ARMv7-R have every width. On RISC-V, the A extension decides the available widths, not the word size. RV64I without the A extension has no lock-free width. RV32 with the A extension has the 32-bit widths.

This means `sync/atomic` won't work on ARMv6-M or on a RISC-V target without the A extension. A build for such a target fails with a compile-time error that names the cause.

It can still work on ARMv7-M, ARMv8-M and RV32 with the A extension if both of these conditions are met:

- You don't use 64-bit atomic types.
- You build with `-ffunction-sections -fdata-sections -Wl,--gc-sections` to remove unused symbols.

**testing**

The test report goes to the `so_write_out` hook. A target with no hook reports nothing.

A failed run traps instead of exiting with a non-zero status, because a freestanding environment has no exit status.

A benchmark reads the clock, so `so bench` needs the `so_time_wall` and `so_time_mono` hooks.

The runner returns the heap to the position it holds before the first test, after every test.

**time**

`Now`, `Since`, `Until`, and `Sleep` read the clock through the `so_time_wall`, `so_time_mono`, and `so_time_sleep` hooks. See [Target hooks](#target-hooks).

`Time.Format` and `Time.Parse` only support named layouts (such as `RFC3339` or `DateOnly`), not custom layouts.

**uuid**

`New` and `NewV4` hold random data only, so they need the `so_crand_read` hook. `NewV7` also reads the clock, so it needs `so_time_wall` as well. See [Target hooks](#target-hooks).

### Hosted only

These packages require a hosted environment and will produce a compile-time error if imported:

```text
conc  flag  log/slog  net  os  sync
```

## Target hooks

A freestanding environment has no standard output, no entropy source and no clock, and only the target knows how to reach its own hardware. So declares a C function for each of these, and the target defines the ones its program needs:

| Hook            | Description                     | With no definition                            |
| --------------- | ------------------------------- | --------------------------------------------- |
| `so_crand_read` | send some bytes to the output   | `crypto/crand` panics, `runtime.Seed` repeats |
| `so_time_wall`  | read some random bytes          | `time.Now` panics                             |
| `so_time_mono`  | get the current wall clock time | no monotonic clock                            |
| `so_time_sleep` | get the current monotonic time  | `time.Sleep` panics                           |
| `so_write_out`  | pause for a given duration      | `panic` and `fmt` print nothing               |

Every hook has a weak default definition in `builtin.c`, so a program that never calls the package still links, and a program that calls it gets the behavior of the last column. A definition in the target replaces the default. Define the hooks in a C file and add it to the build, or embed it with `so:embed`:

```c
so_int so_crand_read(uint8_t* buf, so_int size) {
    return board_rng_fill(buf, size);
}

so_R_i64_i32 so_time_wall(void) {
    return (so_R_i64_i32){.val = board_rtc_unix_seconds(), .val2 = 0};
}

int64_t so_time_mono(void) {
    return board_uptime_ms() * 1000000;
}

void so_time_sleep(int64_t ns) {
    board_delay_ms(ns / 1000000);
}

so_int so_write_out(const uint8_t* buf, so_int size) {
    return board_uart_write(buf, size);
}
```

Notes on each hook:

`so_crand_read` fills `buf` with `size` bytes and returns the number of bytes written. `crypto/crand` has no deterministic fallback on purpose. A generator that quietly returns predictable bytes gives every device the same keys and identifiers, and nothing reports the failure.

`runtime.Seed` reads the same hook, so one definition also seeds `math/rand` and the hash of `maps`. `Seed` does fall back, as [Deterministic random](#deterministic-random) describes, because these packages promise nothing about unpredictability and must work on a board with no entropy source.

`so_time_wall` returns seconds and nanoseconds since the Unix epoch. A board that counts elapsed time but does not know the date returns 0 seconds. Every `Time` then dates at the epoch, and `Since` and `Until` stay exact, because they measure with the monotonic clock.

`so_time_mono` returns nanoseconds from an arbitrary origin. The count must never decrease and must never be 0, because `time.Now` reads 0 as the absence of a monotonic clock. Convert the tick of the board to nanoseconds here, and widen a counter that wraps: a 32-bit counter at 1 kHz wraps after 49 days.

A target with no `so_time_mono` still works. `time.Now` returns a wall clock reading alone, and `Since` and `Until` measure with the wall clock.

`so_write_out` writes `size` bytes from `buf` and returns the number of bytes written. `panic` and the `fmt` package share the hook, so one definition gives the program both a panic message and printed output.

## Limitations

### Bump allocator

In freestanding mode, `mem.System` is implemented as a simple bump allocator backed by a static buffer. It's off by default, but you can enable it by setting the heap size with `-DSO_HEAP_SIZE=<bytes>` at compile time.

In this implementation `free` is a no-op; memory is never reclaimed. `realloc` allocates a new bump region and copies data from the old one; the old region is not freed.

The entire program shares a single heap of `SO_HEAP_SIZE` bytes.

`malloc` takes the next range of the heap with an atomic compare and exchange, so more than one thread can allocate. A target with no lock-free atomics of pointer width gets a plain read and write. On such a target, `malloc` (and hence `mem.System`) is not thread-safe.

It's best not to use `mem.System` in freestanding mode. Instead, use `mem.Arena` to control the heap size and reset it when needed.

### Allocation statistics

`mem.Tracker` counts allocated bytes and objects in 64-bit counters. A target with a lock-free 64-bit atomic adds to a counter atomically, so the tracker stats are thread-safe. A target with no such atomic gets a plain add. On such a target, tracker's stats could be inaccurate if it's shared across multiple threads.

### Deterministic random

If the `so_crand_read` hook is not defined, `runtime.Seed` mixes a counter into a 64-bit value. Each call returns a different value, but the sequence is the same on every run.

The entire program uses one shared counter, so `math/rand` and `maps` draw from the same sequence. The counter uses an atomic add, so more than one thread can call `runtime.Seed`. A target with no lock-free 32-bit atomics gets a plain add. On such a target, the counter is not thread-safe.

A target that defines `so_crand_read` never reads the counter. `runtime.Seed` is then thread-safe only if the hook is thread-safe.

Packages that depend on `runtime.Seed` (like `math/rand` and `maps`) work but produce repeatable output.

### No stdio

A freestanding environment has no standard output, so text goes to the `so_write_out` hook of the target. Define it and point it at a UART or a host import.

`panic` prints its message and location through the hook, then traps. `fmt.Print`, `fmt.Println`, and `fmt.Printf` format the text and write it through the hook. The `testing` package writes its report through the hook too. A target with no hook drops every one of these. A panic then traps with no message.

`print` and `println` are no-ops in a freestanding build, whatever the hook does. Both format with the stdio of the host, which the environment does not have.
