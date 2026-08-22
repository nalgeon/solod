package compiler

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestTarget(t *testing.T) {
	tests := []struct {
		value                         string
		freestanding, windows, isMusl bool
	}{
		{"", false, false, false},
		// zig writes arch-os-abi.
		{"wasm32-freestanding", true, false, false},
		{"x86_64-windows-gnu", false, true, false},
		{"x86_64-linux-musl", false, false, true},
		{"x86_64-linux-gnu", false, false, false},
		{"riscv64-linux", false, false, false},
		{"wasm32-wasi", false, false, false},
		// LLVM writes arch-vendor-os-abi, and names a bare target "none".
		{"x86_64-unknown-linux-musl", false, false, true},
		{"x86_64-pc-windows-gnu", false, true, false},
		{"arm-none-eabi", true, false, false},
		{"riscv32-unknown-none-elf", true, false, false},
		// The musl name carries an ABI suffix on some targets.
		{"arm-linux-musleabihf", false, false, true},
	}
	for _, test := range tests {
		tgt := target(test.value)
		be.Equal(t, tgt.freestanding(), test.freestanding)
		be.Equal(t, tgt.windows(), test.windows)
		be.Equal(t, tgt.musl(), test.isMusl)
	}
}
