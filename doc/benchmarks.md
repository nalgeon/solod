# Solod vs. Go benchmarks

Here are some benchmarks that show how Solod performs on common tasks compared to Go.

[bufio](#buffered-io) •
[bytes](#byte-functions) •
[conc](#concurrency) •
[crypto/crand](#cryptographic-random) •
[encoding/binary](#binary-encoding) •
[encoding/hex](#hex-encoding) •
[encoding/json](#json-encoding) •
[io](#stream-copying) •
[log/slog](#structured-logging) •
[maps](#maps) •
[math/rand](#pseudorandom-numbers) •
[net/netip](#ip-addresses) •
[path](#path-manipulation) •
[strconv](#string-conversion) •
[strings](#string-functions) •
[sync](#synchronization) •
[time](#time) •
[uuid](#uuid)

## Buffered I/O

Solod is ~3x faster than Go for reading and writing, and ~4x faster for scanning.

| Benchmark           |     Go |  Solod | Winner           |
| ------------------- | -----: | -----: | ---------------- |
| Reader (buffered)   | 3089ns | 1073ns | **Solod** - 2.9x |
| Reader (unbuffered) | 1269ns |  412ns | **Solod** - 3.1x |
| Writer (buffered)   | 2988ns | 1038ns | **Solod** - 2.9x |
| Writer (unbuffered) | 4928ns | 1537ns | **Solod** - 3.2x |
| Scanner             |  443ns |  112ns | **Solod** - 4.0x |

## Byte functions

Solod is generally ~1.5x faster than Go, except for Index operations.
Memory usage is the same for both.

| Benchmark  |    Go | Solod | Winner           |
| ---------- | ----: | ----: | ---------------- |
| Clone      | 102ns |  41ns | **Solod** - 2.5x |
| Compare    |  34ns |  25ns | **Solod** - 1.4x |
| Index      |  21ns |  32ns | Go - 1.5x        |
| IndexByte  |  16ns |  25ns | Go - 1.6x        |
| Repeat     | 106ns |  56ns | **Solod** - 1.9x |
| ReplaceAll | 247ns | 258ns | ~same            |
| Split      | 510ns | 422ns | **Solod** - 1.2x |
| ToUpper    | 322ns | 176ns | **Solod** - 1.8x |
| Trim       |  47ns |  44ns | **Solod** - 1.1x |
| TrimSuffix |   4ns |   2ns | **Solod** - 1.8x |

## Byte buffer

Solod reads 1.3x faster and writes 2-4x faster than Go.
Memory usage is the same for both.

| Benchmark  |      Go |  Solod | Winner           |
| ---------- | ------: | -----: | ---------------- |
| ReadString |  2329ns | 1757ns | **Solod** - 1.3x |
| WriteByte  |  8858ns | 2608ns | **Solod** - 3.4x |
| WriteRune  | 15110ns | 3902ns | **Solod** - 3.8x |
| WriteBlock | 17238ns | 7830ns | **Solod** - 2.2x |

## Concurrency

### Pool

`conc.Pool` is a fixed set of worker threads draining a shared task queue, built
on Solod's `Mutex` and `Cond`. Each dispatch crosses into the kernel to wake a
worker (see [Cond](#cond)), so the pool suits coarse-grained tasks: on realistic
workloads that per-task cost is amortized and Solod stays within ~1.1x of Go.

The benchmarks run 8 workers on both sides - Solod's `conc.Pool` against an
equivalent Go pool of persistent goroutines draining a buffered channel. Each
CPU-bound task runs computations of ~40µs; each IO-bound task blocks for 1ms,
standing in for a network or disk round-trip.

| Benchmark                    |  Go | Solod | Winner    |
| ---------------------------- | --: | ----: | --------- |
| Work: 1000 CPU tasks (~40µs) | 7ms |   8ms | Go - 1.1x |
| IO: 64 IO tasks (1ms block)  | 9ms |  10ms | Go - 1.1x |

Each number is the time to run the whole batch of tasks through the pool, not a
single task.

For CPU-bound work Solod's faster compute nearly offsets its heavier dispatch; for
IO-bound work the dispatch cost hides behind the blocking waits. Note the pool
is capped at `NumThreads` OS threads, so unlike Go's goroutines it cannot fan a
single batch out to thousands of concurrent IO waits.

### Channels

`conc.Chan` is a mutex+cond ring buffer when buffered and a rendezvous when
unbuffered. Every blocking cross-thread transfer requires a kernel wakeup,
while Go handles channel wakeups in user space. Because of this, Solod falls
behind Go when threads actually hand off work.

Figures are per value moved through the channel (one send plus its matching receive).

| Benchmark              |    Go | Solod | Winner           |
| ---------------------- | ----: | ----: | ---------------- |
| Uncontended (1 thread) |  24ns |  21ns | **Solod** - 1.1x |
| Unbuffered handoff     | 130ns | 3.0µs | Go - 23x         |
| Buffered handoff (10)  |  44ns | 400ns | Go - 9.1x        |
| Buffered handoff (100) |  33ns |  70ns | Go - 2.1x        |

The uncontended case fills then drains a buffer from a single thread, so nothing
ever blocks; it is just lock plus copy, and Solod's thin pthread mutex edges ahead.
The handoff rows move values between a producer and a consumer thread, where Solod
pays a wakeup on every transfer that parks. The gap is largest for the unbuffered
channel, where every value is a rendezvous with two wakeups; it narrows to ~2x
once a buffer of 100 lets most sends land without parking.

## Cryptographic random

Solod is faster than Go for small reads and random text, and about the same for large reads.

| Benchmark |     Go |  Solod | Winner           |
| --------- | -----: | -----: | ---------------- |
| Read 4B   |   69ns |   40ns | **Solod** - 1.7x |
| Read 32B  |  242ns |  211ns | **Solod** - 1.1x |
| Read 4KB  | 1215ns | 1184ns | ~same            |
| Text      |  264ns |  213ns | **Solod** - 1.2x |

## Binary encoding

Solod encodes fixed-size integers about 2x faster than Go.

| Benchmark       |     Go |  Solod | Winner           |
| --------------- | -----: | -----: | ---------------- |
| BE PutUint64    | 0.63ns | 0.32ns | **Solod** - 2.0x |
| BE AppendUint64 | 1.77ns | 0.95ns | **Solod** - 1.9x |
| LE PutUint64    | 0.63ns | 0.31ns | **Solod** - 2.0x |
| LE AppendUint64 | 1.73ns | 0.95ns | **Solod** - 1.8x |

## Hex encoding

Solod encodes ~1.1x and decodes ~1.4x faster than Go. The ratios hold across buffer sizes from 256B to 16KB; representative figures:

| Benchmark   |     Go |  Solod | Winner           |
| ----------- | -----: | -----: | ---------------- |
| Encode 256B |  193ns |  171ns | **Solod** - 1.1x |
| Encode 4KB  | 2940ns | 2607ns | **Solod** - 1.1x |
| Decode 256B |  127ns |   96ns | **Solod** - 1.3x |
| Decode 4KB  | 1963ns | 1422ns | **Solod** - 1.4x |

## JSON encoding

Solod's JSON API is token-level, with no reflection. Decoding is a fair comparison
against Go's `Decoder.Token` stream: Solod is ~13x faster and allocates once per
document (the scratch buffer for an escaped string) rather than boxing every
token. Encoding is not like-for-like: Go has no token-level encoder, so its
`Encoder` marshals a whole value by reflection, while Solod makes an individual
call per token.

| Benchmark      |     Go | Solod | Winner           |
| -------------- | -----: | ----: | ---------------- |
| Decode         | 6836ns | 528ns | **Solod** - 13x  |
| Decode Unicode |  461ns |  47ns | **Solod** - 9.8x |
| Encode         |  345ns | 322ns | **Solod** - 1.1x |

Benchmarks:

- Decode walks a small document that carries every token kind, pulling each value.
- Decode Unicode decodes a string spelled as a UTF-16 surrogate pair.
- Encode builds an equivalent document.

As for allocations, Go's `Token` boxes every value in an `any`, so its cost
scales with the document. Solod returns values through typed getters and touches
the heap at most once per call.

| Benchmark      | Go allocs | Solod allocs | Go bytes | Solod bytes |
| -------------- | --------: | -----------: | -------: | ----------: |
| Decode         |       187 |            1 |   4128 B |         7 B |
| Decode Unicode |         8 |            1 |   2476 B |        12 B |
| Encode         |         1 |            0 |    112 B |         0 B |

## Stream copying

Solod's `io.CopyN` is ~1.2-1.3x faster than Go and, routed through an allocator, reports no per-op allocations.

| Benchmark   |      Go |   Solod | Winner           |
| ----------- | ------: | ------: | ---------------- |
| CopyN small |   487ns |   419ns | **Solod** - 1.2x |
| CopyN large | 21419ns | 16004ns | **Solod** - 1.3x |

## Structured logging

Solod is 4-7x faster than Go, and logging with attributes allocates nothing in Solod versus three allocations in Go.

| Benchmark       |    Go | Solod | Winner           |
| --------------- | ----: | ----: | ---------------- |
| No attributes   | 166ns |  39ns | **Solod** - 4.3x |
| With attributes | 259ns |  38ns | **Solod** - 6.8x |

## Maps

### Int keys

For heap-allocated maps, Solod is ~1.4x faster than Go across all operations.

Solod's built-in map is even faster, but it's only useful in certain situations — it's fixed size and stack-allocated.

| Benchmark |      Go |   Solod | Solod (built-in) | Winner           |
| --------- | ------: | ------: | ---------------: | ---------------- |
| Set       | 35645ns | 26333ns |              n/a | **Solod** - 1.4x |
| Set (pre) |  9676ns |  8813ns |           3109ns | **Solod** - 1.1x |
| Get       |  5594ns |  1581ns |           2577ns | **Solod** - 3.5x |
| Delete    | 23968ns | 14889ns |              n/a | **Solod** - 1.6x |

### String keys

Solod modifications are ~1.4x faster than Go, while lookups are slightly slower.

| Benchmark |      Go |   Solod | Solod (built-in) | Winner           |
| --------- | ------: | ------: | ---------------: | ---------------- |
| Set       | 47805ns | 31055ns |              n/a | **Solod** - 1.5x |
| Set (pre) | 14699ns | 12101ns |           6585ns | **Solod** - 1.2x |
| Get       |  9216ns | 10170ns |          10531ns | Go - 1.1x        |
| Delete    | 33819ns | 24227ns |              n/a | **Solod** - 1.4x |

## Pseudorandom numbers

Solod's raw source generator is ~1.6x faster, but the package-level helpers (global source, bounded ints, floats) are about 2x slower than Go.

| Benchmark     |    Go | Solod | Winner           |
| ------------- | ----: | ----: | ---------------- |
| Source Uint64 | 4.7ns | 2.8ns | **Solod** - 1.6x |
| Global Uint64 | 4.8ns | 8.8ns | Go - 1.8x        |
| Uint64        | 4.5ns | 8.8ns | Go - 2.0x        |
| Int64N (1e9)  | 4.6ns | 9.1ns | Go - 2.0x        |
| Int64N (4e18) | 9.1ns |  12ns | Go - 1.3x        |
| Float64       | 4.4ns | 9.3ns | Go - 2.1x        |

## IP addresses

Solod parses IPv6 ~1.4-1.5x faster and formats addresses 2-4x faster than Go, allocating nothing. The exception is parsing a zoned IPv6 address, which makes an `if_nametoindex` syscall and is far slower.

Parsing:

| Benchmark     |   Go |   Solod | Winner           |
| ------------- | ---: | ------: | ---------------- |
| Parse v4      | 18ns |    16ns | **Solod** - 1.1x |
| Parse v6      | 81ns |    55ns | **Solod** - 1.5x |
| Parse v6e     | 47ns |    33ns | **Solod** - 1.4x |
| Parse v6+v4   | 48ns |    40ns | **Solod** - 1.2x |
| Parse v6+zone | 64ns | 19087ns | Go - syscall     |

Formatting:

| Benchmark      |   Go | Solod | Winner           |
| -------------- | ---: | ----: | ---------------- |
| String v4      | 20ns |   9ns | **Solod** - 2.3x |
| String v6      | 53ns |  17ns | **Solod** - 3.1x |
| String v6+v4   | 23ns |  11ns | **Solod** - 2.0x |
| String v6+zone | 60ns |  14ns | **Solod** - 4.3x |

## Path manipulation

Slash paths are roughly on par with Go. Matching and joining is marginally slower in Solod.

| Benchmark   |    Go | Solod | Winner    |
| ----------- | ----: | ----: | --------- |
| Join        |  61ns |  73ns | Go - 1.2x |
| Match true  | 105ns | 113ns | Go - 1.1x |
| Match false | 106ns | 114ns | Go - 1.1x |

## String conversion

### Parsing

Solod parses floats ~1.5x faster and ints ~2x faster than Go.

| Benchmark       |   Go | Solod | Winner           |
| --------------- | ---: | ----: | ---------------- |
| Atof64 decimal  | 21ns |  12ns | **Solod** - 1.7x |
| Atof64 float    | 24ns |  15ns | **Solod** - 1.6x |
| Atof64 exp      | 25ns |  21ns | **Solod** - 1.2x |
| Atof64 big      | 38ns |  25ns | **Solod** - 1.5x |
| ParseInt 7-bit  | 10ns |   4ns | **Solod** - 2.5x |
| ParseInt 26-bit | 14ns |   7ns | **Solod** - 2.0x |
| ParseInt 31-bit | 16ns |   9ns | **Solod** - 1.9x |
| ParseInt 56-bit | 24ns |  15ns | **Solod** - 1.6x |
| ParseInt 62-bit | 26ns |  17ns | **Solod** - 1.6x |

### Formatting

Solod formats floats ~1.2x faster and ints ~2x faster than Go.

| Benchmark           |   Go | Solod | Winner           |
| ------------------- | ---: | ----: | ---------------- |
| FormatFloat decimal | 30ns |  27ns | **Solod** - 1.1x |
| FormatFloat float   | 43ns |  34ns | **Solod** - 1.3x |
| FormatFloat exp     | 35ns |  30ns | **Solod** - 1.2x |
| FormatFloat big     | 39ns |  33ns | **Solod** - 1.2x |
| FormatInt 7-bit     | 14ns |   5ns | **Solod** - 3.0x |
| FormatInt 26-bit    | 17ns |   7ns | **Solod** - 2.3x |
| FormatInt 31-bit    | 20ns |   8ns | **Solod** - 2.3x |
| FormatInt 56-bit    | 24ns |  12ns | **Solod** - 2.0x |
| FormatInt 62-bit    | 26ns |  13ns | **Solod** - 2.0x |

## String functions

Solod is generally ~1.3x faster than Go, except for Index operations.
Memory usage is the same for both.

| Benchmark  |     Go |  Solod | Winner           |
| ---------- | -----: | -----: | ---------------- |
| Clone      |   99ns |   42ns | **Solod** - 2.4x |
| Compare    |   47ns |   36ns | **Solod** - 1.3x |
| Fields     | 1524ns |  908ns | **Solod** - 1.7x |
| Index      |   25ns |   35ns | Go - 1.4x        |
| IndexByte  |   22ns |   33ns | Go - 1.5x        |
| Repeat     |  127ns |   64ns | **Solod** - 1.9x |
| ReplaceAll |  243ns |  200ns | **Solod** - 1.2x |
| Split      | 1899ns | 1399ns | **Solod** - 1.3x |
| ToUpper    | 2066ns | 1602ns | **Solod** - 1.3x |
| Trim       |  501ns |  373ns | **Solod** - 1.3x |

## String builder

Solod is 2-4x faster than Go and uses 10%-20% less memory.

| Benchmark                |    Go | Solod | Winner           |
| ------------------------ | ----: | ----: | ---------------- |
| Write bytes (auto-grow)  | 245ns | 118ns | **Solod** - 2.1x |
| Write bytes (pre-grow)   | 109ns |  29ns | **Solod** - 3.8x |
| Write string (auto-grow) | 224ns | 116ns | **Solod** - 1.9x |
| Write string (pre-grow)  | 113ns |  29ns | **Solod** - 3.9x |

## Synchronization

Solod's synchronization primitives are built on POSIX threads: `Mutex` and `Cond`
wrap a pthread mutex and condition variable. The mutex beats Go's for short,
spin-friendly critical sections but loses once contention forces threads to park
in the kernel. `Cond` is slower because it always parks threads in the kernel
instead of a user-space scheduler. `Once` takes a lock-free atomic fast path,
so uncontended it is close to Go; under contention it inherits the same kernel
dispatch cost as `Cond`.

The contended benchmarks run 8 worker threads that share one primitive, using a
persistent thread pool on the Solod side and an equivalent persistent goroutine pool
on the Go side.

### Mutex

Uncontended lock/unlock is ~1.6x faster than Go. Under contention the result
depends on how long the lock is held. With an empty critical section (the _spin_
row) a waiting thread reacquires the lock while still spinning and almost never
parks, so Solod's thin pthread wrapper wins by ~2.8x. Give the critical section a
small (~1µs) amount of real work (the _work_ row), and waiters exhaust their spin
budget and park in the kernel; every handoff then costs a wakeup syscall, and Solod
drops to ~1.8x behind Go. The _work_ critical section runs identically on both sides
single-threaded, so the gap is purely the parking cost, not the work.

| Benchmark           |    Go | Solod | Winner           |
| ------------------- | ----: | ----: | ---------------- |
| Uncontended         |  14ns |   9ns | **Solod** - 1.6x |
| TryLock             |  14ns |   9ns | **Solod** - 1.6x |
| Contended spin (8t) | 600µs | 215µs | **Solod** - 2.8x |
| Contended work (8t) |   9ms |  16ms | Go - 1.8x        |

The uncontended rows are per `Lock`/`Unlock` pair on a single thread; the
contended rows are the total time for 8 threads to run a fixed batch of
lock/unlock rounds over the shared mutex.

### Cond

Solod's condition variable is ~7-10x slower than Go across waiter counts: each
wakeup crosses into the kernel, while Go wakes goroutines in user space. Figures
are per 1000 rendezvous rounds.

| Benchmark  |     Go | Solod | Winner    |
| ---------- | -----: | ----: | --------- |
| 1 waiter   | 0.15ms | 1.5ms | Go - 10x  |
| 2 waiters  | 0.39ms | 2.9ms | Go - 7.4x |
| 4 waiters  | 0.87ms | 7.3ms | Go - 8.4x |
| 8 waiters  |  2.0ms |  14ms | Go - 7.0x |
| 16 waiters |  3.9ms |  28ms | Go - 7.2x |
| 32 waiters |  9.0ms |  60ms | Go - 6.7x |

### Once

Solod's `Do` takes a lock-free atomic fast path: once the initializer has run,
every call is just an atomic load. Uncontended, both sides do that single load
and land within ~1.2x of each other. Under contention the gap is because of
`conc.Pool` dispatch: waking the eight workers crosses into the kernel, the
same cost that makes `Cond` slow, rather than anything in `Once`.

| Benchmark      |    Go | Solod | Winner    |
| -------------- | ----: | ----: | --------- |
| Uncontended    | 2.1ns | 2.6ns | Go - 1.2x |
| Contended (8t) | 6.0µs |  32µs | Go - 5.3x |

The uncontended row is a single `Do` call; the contended row is one round of 8
threads calling `Do` on the same `Once`.

### Atomic

Solod's atomic types map directly to the compiler's `__atomic` builtins - the same
hardware instructions Go emits - so performance is on par with Go across the board.

Single-value ops use `Uint64`; the contended row runs 8 threads adding to one counter.

| Benchmark       |    Go | Solod | Winner |
| --------------- | ----: | ----: | ------ |
| Load            |   2ns |   2ns | ~same  |
| Store           |   2ns |   2ns | ~same  |
| Add             |   7ns |   7ns | ~same  |
| Swap            |   7ns |   6ns | ~same  |
| CompareAndSwap  |  13ns |  13ns | ~same  |
| Add (8 threads) | 180µs | 180µs | ~same  |

## Time

Regular time functions and methods in Solod are slightly slower than in Go.
In parsing and formatting, Solod is 5x faster for predefined layouts (RFC3339, DateTime, etc.),
about the same for custom parsing, and 5x slower for custom formatting (due to strftime overhead).

| Benchmark    |   Go | Solod | Winner           |
| ------------ | ---: | ----: | ---------------- |
| Date         |  7ns |   2ns | **Solod** - 3.2x |
| ISOWeek      |  9ns |   2ns | **Solod** - 4.3x |
| Now          | 34ns |  39ns | Go - 1.1x        |
| Since        | 17ns |  25ns | Go - 1.5x        |
| UnixNano     | 34ns |  38ns | Go - 1.1x        |
| Until        | 17ns |  24ns | Go - 1.4x        |
| Format       | 39ns |   4ns | **Solod** - 8.8x |
| FormatCustom | 55ns | 250ns | Go - 4.5x        |
| Parse        | 27ns |   6ns | **Solod** - 4.9x |
| ParseCustom  | 55ns |  45ns | **Solod** - 1.2x |

## UUID

Solod generates v4 UUIDs a bit faster and formats them ~4x faster; v7 generation and parsing are on par with Go.

| Benchmark     |    Go | Solod | Winner           |
| ------------- | ----: | ----: | ---------------- |
| NewV4         | 251ns | 212ns | **Solod** - 1.2x |
| NewV7         |  72ns |  79ns | Go - 1.1x        |
| String        |  34ns |   9ns | **Solod** - 3.9x |
| Parse (ok)    |  29ns |  29ns | ~same            |
| Parse (error) |  26ns |  29ns | Go - 1.1x        |

Go 1.27

## Methodology

All benchmarks run on an Apple M1 CPU running macOS. The C code is compiled with Clang 16 using these CFLAGS and mimalloc as the system allocator:

```text
-Ofast -march=native -flto -funroll-loops
```

The Go benchmarks use Go 1.26 (unless stated otherwise) and run with `go test -bench=.`.
