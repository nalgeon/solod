package compiler

import (
	"slices"
	"testing"

	"github.com/nalgeon/be"
)

func TestCheckMode(t *testing.T) {
	warn := []string{"-Wall", "-Wextra", "-Werror", "-Wno-shadow", "-Wno-unused-label"}
	tests := []struct {
		mode    string
		cc      string
		tgt     target
		want    []string
		wantErr bool
	}{
		{"", "cc", "", nil, false},
		{"off", "cc", "", nil, false},
		{"warn", "cc", "", warn, false},
		{"sanitize", "cc", "", slices.Concat(warn, []string{
			"-g", "-fsanitize=address,undefined",
			"-fno-sanitize-recover=all", "-fno-omit-frame-pointer",
		}), false},
		{"analyze", "cc", "", slices.Concat(warn, []string{"-fanalyzer"}), false},
		// A sanitizer runtime needs a hosted target.
		{"sanitize", "cc", "wasm32-freestanding", nil, true},
		// Only GCC has -fanalyzer.
		{"analyze", "clang", "", nil, true},
		{"analyze", "zig cc", "", nil, true},
		{"none", "cc", "", nil, true},
	}
	for _, test := range tests {
		flags, err := checkMode(test.mode, test.cc, test.tgt)
		be.Equal(t, err != nil, test.wantErr)
		be.Equal(t, flags, test.want)
	}
}

func TestPanicMode(t *testing.T) {
	const traceDef = "-DSO_PANIC_MODE=SO_PANIC_TRACE"
	const exitDef = "-DSO_PANIC_MODE=SO_PANIC_EXIT"
	tests := []struct {
		mode    string
		cc      string
		tgt     target
		def     string
		flags   []string
		wantErr bool
	}{
		{"", "cc", "", traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "gcc", "", traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "x86_64-w64-mingw32-gcc", "", traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"trace", `C:\MinGW\bin\gcc.exe`, "", traceDef, []string{"-fno-omit-frame-pointer"}, false},
		// A windows target rejects -rdynamic whatever the C compiler is called.
		{"trace", "zig cc", "x86_64-windows-gnu", traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"", "cc", "wasm32-freestanding", traceDef, nil, false},
		{"trace", "gcc", "wasm32-freestanding", traceDef, nil, false},
		// musl leaves the trace empty, so it defaults to exit. An explicit
		// mode still wins.
		{"", "cc", "x86_64-linux-musl", exitDef, nil, false},
		{"trace", "cc", "x86_64-linux-musl", traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"exit", "cc", "", exitDef, nil, false},
		{"abort", "cc", "", "-DSO_PANIC_MODE=SO_PANIC_ABORT", nil, false},
		{"none", "cc", "", "", nil, true},
	}
	for _, test := range tests {
		def, flags, err := panicMode(test.mode, test.cc, test.tgt)
		be.Equal(t, err != nil, test.wantErr)
		be.Equal(t, def, test.def)
		be.Equal(t, flags, test.flags)
	}
}

func TestTargetFlags(t *testing.T) {
	// The panic mode is "exit" so that the trace flags, which depend on the
	// C compiler name, stay out of the comparison.
	t.Setenv("CC", "zig cc")

	hosted, err := newCompileOptions(Options{PanicMode: "exit"})
	be.Err(t, err, nil)
	be.Equal(t, hosted.flags, nil)

	bare, err := newCompileOptions(Options{PanicMode: "exit", Target: "wasm32-freestanding"})
	be.Err(t, err, nil)
	be.Equal(t, bare.flags, []string{"--target=wasm32-freestanding", "-ffreestanding"})

	// A hosted target passes the value through and adds nothing.
	linux, err := newCompileOptions(Options{PanicMode: "exit", Target: "riscv64-linux"})
	be.Err(t, err, nil)
	be.Equal(t, linux.flags, []string{"--target=riscv64-linux"})
}

func TestTargetNeedsClang(t *testing.T) {
	// GCC selects the target with a separate cross compiler, not with --target.
	t.Setenv("CC", "gcc-16")
	_, err := newCompileOptions(Options{Target: "riscv64-linux"})
	be.True(t, err != nil)

	// An ambiguous name passes, so the C compiler reports the problem itself.
	t.Setenv("CC", "cc")
	_, err = newCompileOptions(Options{PanicMode: "exit", Target: "riscv64-linux"})
	be.Err(t, err, nil)
}

func TestAssertMode(t *testing.T) {
	tests := []struct {
		mode    string
		want    []string
		wantErr bool
	}{
		{"", nil, false},
		{"on", nil, false},
		{"off", []string{"-DSO_NO_ASSERT"}, false},
		{"none", nil, true},
	}
	for _, test := range tests {
		got, err := assertMode(test.mode)
		be.Equal(t, err != nil, test.wantErr)
		be.Equal(t, got, test.want)
	}
}
