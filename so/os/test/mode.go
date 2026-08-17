package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestFileMode_IsRegular(t *testing.T) {
	// Only the permission bits describe a regular file.
	if !os.FileMode(0o644).IsRegular() {
		t.Error("IsRegular(0644): want true")
	}
	if !os.FileMode(0).IsRegular() {
		t.Error("IsRegular(0): want true")
	}

	// Every type bit describes something else.
	modes := []os.FileMode{
		os.ModeDir,
		os.ModeSymlink,
		os.ModeNamedPipe,
		os.ModeSocket,
		os.ModeDevice,
		os.ModeCharDevice,
		os.ModeIrregular,
	}
	for i, mode := range modes {
		if (mode | 0o644).IsRegular() {
			t.Errorf("IsRegular(type bit %d): want false", i)
		}
	}
}

func TestFileMode_IsDir(t *testing.T) {
	if !(os.ModeDir | 0o755).IsDir() {
		t.Error("IsDir(ModeDir): want true")
	}
	if os.FileMode(0o755).IsDir() {
		t.Error("IsDir(0755): want false")
	}
	if os.ModeSymlink.IsDir() {
		t.Error("IsDir(ModeSymlink): want false")
	}
}

func TestFileMode_Perm(t *testing.T) {
	// Perm drops the type bits and keeps the permission bits.
	if got := (os.ModeDir | 0o755).Perm(); got != 0o755 {
		t.Errorf("Perm(ModeDir|0755) = %o, want 755", int(got))
	}
	if got := (os.ModeSetuid | 0o600).Perm(); got != 0o600 {
		t.Errorf("Perm(ModeSetuid|0600) = %o, want 600", int(got))
	}
	if got := os.ModeSymlink.Perm(); got != 0 {
		t.Errorf("Perm(ModeSymlink) = %o, want 0", int(got))
	}
	if os.ModePerm != 0o777 {
		t.Errorf("ModePerm = %o, want 777", int(os.ModePerm))
	}
}
