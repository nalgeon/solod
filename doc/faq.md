# Frequently asked questions

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
