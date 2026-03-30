# Design principles

So is highly opinionated.

**Simplicity is key**. Fewer features are always better. Every new feature is strongly discouraged by default and should be added only if there are very convincing real-world use cases to support it. This applies to the standard library too — So tries to export as little of Go's stdlib API as possible while still remaining highly useful for real-world use cases.

**No heap allocations** are allowed in language built-ins (like maps, slices, new, or append). Heap allocations are allowed in the standard library, but they must clearly state when an allocation happens and who owns the allocated data.

**Fast and easy C interop**. Even though So uses Go syntax, it's basically C with its own standard library. Calling C from So, and So from C, should always be simple to write and run efficiently. The So standard library (translated to C) should be easy to add to any C project.

**Readability**. There are several languages that claim they can transpile to readable C code. Unfortunately, the C code they generate is usually unreadable or barely readable at best. So isn't perfect in this area either (though it's arguably better than others), but it aims to produce C code that's as readable as possible.

**Go compatibility**. So code is syntactically valid Go code, with no exceptions. Semantics may differ.

Non-goals:

**Raw performance**. You can definitely write C code by hand that runs faster than code produced by So. Also, some features in So, like interfaces, are currently implemented in a way that's not very efficient, mainly to keep things simple.

**Hiding C entirely**. So is a cleaner way to write C, not a replacement for it. You should know C to use So effectively.

**Go feature parity**. Less is more. Iterators aren't coming, and neither are generic methods.
