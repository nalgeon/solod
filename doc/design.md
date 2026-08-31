# Design principles

Solod is highly opinionated.

## First-level goals

**Simplicity is key**. Fewer features are always better. Every new feature is strongly discouraged by default and should be added only if there are very convincing real-world use cases to support it. This applies to the standard library too — Solod tries to export as little of Go's stdlib API as possible while still remaining highly useful for real-world use cases.

**No heap allocations** are allowed in language built-ins (like maps, slices, new, or append). Heap allocations are allowed in the standard library, but they have to be explicit. If a function or type allocates memory, it must take an allocator and clearly state ownership in the documentation.

**Fast and easy C interop**. Even though Solod uses Go syntax, it's basically C with a Go-like standard library. Calling C from Solod, and Solod from C, should always be simple to write and run efficiently. The Solod standard library (translated to C) should be easy to add to any C project.

**Go compatibility**. Solod code is syntactically valid Go code, with no exceptions. Semantics may differ.

## Second-level goals

**Performance**. You can definitely write C code by hand that runs faster than code produced by Solod. Also, some features, like interfaces, are currently implemented in a way that's not the most efficient, mainly to keep things simple. Still, performance matters: the code is benchmarked and optimized to match or beat Go in speed and resource usage.

**Readability**. There are several languages that claim they can transpile to readable C code. Unfortunately, the C code they generate is usually unreadable or barely readable at best. Solod isn't perfect in this area either (though it's arguably better than others), but it aims to produce C code that's as readable as possible.

## Non-goals

**Hiding C entirely**. Solod is a cleaner way to write C, not a replacement for it. You should be familiar with C to use Solod effectively.

**Go feature parity**. Less is more. Iterators aren't coming, and neither are generic methods. Concurrency features are part of the standard library, not the language itself.
