# Frequently asked questions

_Why not Rust/Zig/Odin/other language?_

Because I like C and Go.

_Why not TinyGo?_

TinyGo is lightweight, but it still has a garbage collector, a runtime, and aims to support all Go features. What I'm after is something even simpler, with no runtime at all, source-level C interop, and eventually, Go's standard library ported to plain C so it can be used in regular C projects.

_How does So handle memory?_

Everything is stack-allocated by default. There's no garbage collector or reference counting. The standard library provides explicit heap allocation in the `so/mem` package when you need it.

_Is it safe?_

So has extra safeguards beyond Go's default type checking:

- It will panic on out-of-bounds array access and nil dereference.
- It won't let you return stack-allocated memory in common situations.
- Programs can detect memory leaks with a tracking allocator.

However, escape analysis doesn't catch every case, and the leak checker won't detect double-free or use-after-free errors by itself.

Most memory-related problems can be caught with AddressSanitizer in modern compilers. I strongly recommend turning on sanitizers with the `-sanitize` flag while developing. Or set the flags in `CFLAGS` yourself:

```text
-g -fno-omit-frame-pointer -fsanitize=address,undefined
```

_What about concurrency?_

Right now, concurrency tools like threads, channels, and worker pools are available through the standard library (the `so/conc` package), not built into the language itself. As the standard library matures, these features might eventually be accessible using Go's standard `go`, `chan` and `select` keywords.

_Can I use So code from C (and vice versa)?_

Yes. So compiles to plain C, therefore calling So from C is just calling C from C. Calling C from So is equally straightforward — see the language tour for details.

_Can I compile existing Go packages with So?_

Not really. Go uses automatic memory management, while So uses manual memory management. So also supports far fewer features than Go. Neither Go's standard library nor third-party packages will work with So without changes.

_How stable is this?_

Not for production at the moment.

_Where's the standard library?_

There is a growing set of high-level packages (`so/bytes`, `so/mem`, `so/slices`, ...), and a low-level `so/c` package to help with C interop. Check out the standard library overview for more details.
