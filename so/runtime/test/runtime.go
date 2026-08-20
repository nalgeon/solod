package runtime_test

import (
	"solod.dev/so/runtime"
	"solod.dev/so/testing"
)

func TestGOOS(t *testing.T) {
	os := runtime.GOOS
	if os != "bare" && os != "darwin" && os != "linux" && os != "windows" && os != "wasip1" {
		t.Error("Unexpected GOOS")
	}
}

func TestGOARCH(t *testing.T) {
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" && arch != "386" && arch != "riscv64" && arch != "wasm" {
		t.Error("Unexpected GOARCH")
	}
}

func TestHosted(t *testing.T) {
	bare := runtime.GOOS == "bare"
	if runtime.Hosted == bare {
		t.Error("Hosted must disagree with a bare GOOS")
	}
}

func TestFileName(t *testing.T) {
	if len(runtime.FileName) == 0 {
		t.Error("Empty FileName")
	}
}

func TestLine(t *testing.T) {
	first := runtime.Line
	second := runtime.Line
	if first < 1 {
		t.Error("Line must be >= 1")
	}
	// Line expands at the point of use, so the two reads differ.
	if second <= first {
		t.Error("Line must grow down the file")
	}
}

func TestFuncName(t *testing.T) {
	if len(runtime.FuncName) == 0 {
		t.Error("Empty FuncName")
	}
}

func TestNumCPU(t *testing.T) {
	if runtime.NumCPU() < 1 {
		t.Error("NumCPU must be >= 1")
	}
}

func TestSeed(t *testing.T) {
	// Seed reads a 0 from the so_crand_read hook as a broken hook, so 0 is
	// the one value Seed never returns.
	first := runtime.Seed()
	if first == 0 {
		t.Error("Seed must not be 0")
	}
	// Two calls return different values.
	second := runtime.Seed()
	if second == first {
		t.Error("Seed must not repeat")
	}
}

func TestVersion(t *testing.T) {
	v := runtime.Version()
	if len(v) == 0 {
		t.Error("Empty version")
	}
}
