// This case checks that runtime.Seed never returns 0 and never repeats a value.
//
// The case defines no so_crand_read hook. A freestanding build then takes the
// weak definition in builtin.c, which reads no bytes, and Seed draws from its
// own counter. Only that build covers the counter, so run
// `make run-case name=runtime-seed mode=bare` as well as the hosted run of test-lang.
package main

import (
	"solod.dev/so/runtime"
)

// seedCount is the number of seeds that the case collects.
const seedCount = 16

func check(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func main() {
	var seeds [seedCount]uint64
	for i := range seeds {
		seeds[i] = runtime.Seed()
	}

	// The fallback Seed implementation must never return 0.
	for _, seed := range seeds {
		check(seed != 0, "runtime: zero seed")
	}

	// The whole program draws from one counter, so no value repeats.
	for i := range seeds {
		for j := i + 1; j < seedCount; j++ {
			check(seeds[i] != seeds[j], "runtime: repeated seed")
		}
	}

	println("ok")
}
