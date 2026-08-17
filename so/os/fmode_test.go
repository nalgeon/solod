package os

import "testing"

func TestMakePosixMode(t *testing.T) {
	tests := []struct {
		fmode FileMode
		want  mode_t
	}{
		{0o644, 0o644},
		{0o777, 0o777},
		{0, 0},
		{ModeDir | 0o755, 0o755},     // the type bits do not go to mode_t
		{ModeSymlink | 0o777, 0o777}, // same
		{ModeSetuid | 0o600, 0o4600},
		{ModeSetgid | 0o700, 0o2700},
		{ModeSticky | 0o777, 0o1777},
		{ModeSetuid | ModeSetgid | ModeSticky | 0o600, 0o7600},
	}
	for _, test := range tests {
		got := makePosixMode(test.fmode)
		if got != test.want {
			t.Errorf("makePosixMode(%v) = %o, want %o", test.fmode, got, test.want)
		}
	}
}

func TestToFileMode(t *testing.T) {
	tests := []struct {
		pmode mode_t
		want  FileMode
	}{
		{sIFREG | 0o644, 0o644},
		{sIFDIR | 0o755, ModeDir | 0o755},
		{sIFLNK | 0o777, ModeSymlink | 0o777},
		{sIFIFO | 0o600, ModeNamedPipe | 0o600},
		{sIFSOCK | 0o600, ModeSocket | 0o600},
		{sIFBLK | 0o600, ModeDevice | 0o600},
		{sIFCHR | 0o600, ModeCharDevice | 0o600},
		{sIFREG | 0o4600, ModeSetuid | 0o600},
		{sIFDIR | 0o2700, ModeDir | ModeSetgid | 0o700},
		{sIFDIR | 0o1777, ModeDir | ModeSticky | 0o777},
		{sIFREG | 0o7600, ModeSetuid | ModeSetgid | ModeSticky | 0o600},
	}
	for _, test := range tests {
		got := test.pmode.toFileMode()
		if got != test.want {
			t.Errorf("mode_t(%o).toFileMode() = %v, want %v", test.pmode, got, test.want)
		}
	}
}

func TestModeRoundTrip(t *testing.T) {
	// The special bits and the permission bits survive both conversions.
	modes := []FileMode{
		0o644,
		0o600 | ModeSetuid,
		0o700 | ModeSetgid,
		0o777 | ModeSticky,
		0o600 | ModeSetuid | ModeSetgid | ModeSticky,
	}
	for _, want := range modes {
		got := (makePosixMode(want) | sIFREG).toFileMode()
		if got != want {
			t.Errorf("round trip of %v = %v", want, got)
		}
	}
}
