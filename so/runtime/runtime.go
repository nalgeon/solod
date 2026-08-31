// Package runtime provides information about
// the environment where the program was compiled.
//
// # Source locations
//
// [FileName], [Line], and [FuncName] map to the C preprocessor's __FILE__,
// __LINE__, and __func__. They expand at the place they're written, not where
// the function is called. So if you use them inside a logging helper, they'll
// show the helper's info, not the caller's:
//
//	// Always reports the line of the Printf below, not the caller's.
//	func logf(msg string) {
//	    fmt.Printf("%s:%d %s\n", runtime.FileName, runtime.Line, msg)
//	}
//
// By default, they describe the generated C code: FileName is the name of
// the .c file, and Line is the line number in that file. Use -track-source
// to map both back to the original source. FuncName is always the generated
// C function name, so a package-level function will look like "package_Func".
package runtime

//so:embed runtime.h
var runtime_h string

//so:embed runtime.c
var runtime_c string

// GOOS is the running program's operating system target:
// one of darwin, linux, windows, and so on.
//
//so:extern
const GOOS string = "bare"

// GOARCH is the running program's architecture target:
// one of amd64, arm64, and so on.
//
//so:extern
const GOARCH string = "unknown"

// Recognized GOOS/GOARCH pairs are:
// GOOS		GOARCH
// bare		*
// darwin	amd64
// darwin	arm64
// linux	amd64
// linux	arm64
// linux	386
// linux	riscv64
// wasip1	wasm
// windows	amd64
// windows	arm64

// Hosted reports whether the program runs in a hosted environment.
// A freestanding build sets Hosted to false.
//
//so:extern
const Hosted bool = true

//so:extern runtime_buildVersion
var buildVersion string

// FileName is the name of the source file at the point of use.
//
//so:extern so_str(__FILE__)
var FileName string

// Line is the line number at the point of use.
//
//so:extern __LINE__
var Line int

// FuncName is the name of the enclosing function at the point of use.
//
//so:extern so_str(__func__)
var FuncName string

// NumCPU returns the number of logical CPUs usable by the program.
// The result is always >= 1. It reports the number of online CPUs and does
// not account for GOMAXPROCS or scheduler affinity. On freestanding targets
// and platforms without a CPU count query it returns 1.
//
//so:extern
func NumCPU() int {
	return 1
}

// Seed returns a random 64-bit seed.
//
// In a hosted environment, Seed reads the cryptographic random of the
// operating system, so it panics on a platform with no such source.
// Seed is thread-safe there.
//
// In a freestanding environment, Seed reads the so_crand_read hook of the
// target, and is thread-safe if the hook is thread-safe. A target with no hook
// gets a deterministic sequence, which repeats on every run. The sequence is
// thread-safe only on a target with lock-free 32-bit atomics.
//
//so:extern
func Seed() uint64 {
	return 0
}

// Version returns the Solod tree's version string.
// It is either the commit hash and date at the time of the build or,
// when possible, a release tag like "v0.1.0".
func Version() string {
	return buildVersion
}
