package compiler

import "strings"

// target is a C compiler target value, in the spelling that clang and
// zig cc accept after --target=. An empty target means the host.
//
// The position of a component is not fixed: zig writes arch-os-abi, LLVM
// writes arch-vendor-os-abi, and both spellings reach Solod. The methods below
// search every component, so they read the same target either way.
type target string

// has reports whether one component of the target value starts with prefix.
func (t target) has(prefix string) bool {
	for part := range strings.SplitSeq(string(t), "-") {
		if strings.HasPrefix(part, prefix) {
			return true
		}
	}
	return false
}

// freestanding reports whether the target has no C standard library.
// zig names such a target "freestanding", LLVM names it "none".
func (t target) freestanding() bool {
	return t.has("freestanding") || t.has("none")
}

// windows reports whether the target is Windows.
func (t target) windows() bool {
	return t.has("windows")
}

// musl reports whether the target uses the musl C library. The name carries
// an ABI suffix on some targets, as in "musleabihf".
func (t target) musl() bool {
	return t.has("musl")
}
