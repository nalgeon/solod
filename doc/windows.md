# Windows

So builds for `windows/amd64` and `windows/arm64`. The build requires GCC or Clang toolchain, not MSVC. MinGW works, and so does `zig cc` cross-compilation:

```sh
export CC="zig cc"
export CFLAGS="--target=x86_64-windows-gnu"
export LDFLAGS="-lbcrypt -liphlpapi"
so build -o app.exe .
```

Make sure to link the system libraries manually, as shown above, because `so build` can't link them automatically yet.

Every package of the [freestanding set](freestanding.md#stdlib-packages) works on Windows. The packages that need POSIX are not supported.

## Libraries

A Windows build manually links two system libraries:

| Library    | Needed by              | Symbol            |
| ---------- | ---------------------- | ----------------- |
| `bcrypt`   | `crypto/crand`, `maps` | `BCryptGenRandom` |
| `iphlpapi` | `netip.Addr.WithZone`  | `if_nametoindex`  |

Link only what the program actually uses. For example, if a program doesn't use maps or random, don't link `bcrypt`.

## Behavior

The following behavior on Windows differs from that on other targets:

`time.Parse` supports only named layouts, such as `RFC3339` and `DateOnly`. It panics when given a custom layout. `time.Format` accepts custom layouts as usual.

`runtime.NumCPU` always returns 1, no matter how many CPUs the machine actually has.
