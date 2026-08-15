package compiler

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestSanitizeFlags(t *testing.T) {
	tests := []struct {
		list string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{",", nil},
		{"address", []string{"-g", "-fno-omit-frame-pointer", "-fsanitize=address"}},
		{"address,undefined", []string{"-g", "-fno-omit-frame-pointer", "-fsanitize=address", "-fsanitize=undefined"}},
		{" address , undefined ", []string{"-g", "-fno-omit-frame-pointer", "-fsanitize=address", "-fsanitize=undefined"}},
	}
	for _, test := range tests {
		be.Equal(t, sanitizeFlags(test.list), test.want)
	}
}

func TestPanicMode(t *testing.T) {
	const traceDef = "-DSO_PANIC_MODE=SO_PANIC_TRACE"
	tests := []struct {
		mode         string
		cc           string
		freestanding bool
		def          string
		flags        []string
		wantErr      bool
	}{
		{"", "cc", false, traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "gcc", false, traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "x86_64-w64-mingw32-gcc", false, traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"trace", `C:\MinGW\bin\gcc.exe`, false, traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"", "cc", true, traceDef, nil, false},
		{"trace", "gcc", true, traceDef, nil, false},
		{"exit", "cc", false, "-DSO_PANIC_MODE=SO_PANIC_EXIT", nil, false},
		{"abort", "cc", false, "-DSO_PANIC_MODE=SO_PANIC_ABORT", nil, false},
		{"none", "cc", false, "", nil, true},
	}
	for _, test := range tests {
		def, flags, err := panicMode(test.mode, test.cc, test.freestanding)
		be.Equal(t, err != nil, test.wantErr)
		be.Equal(t, def, test.def)
		be.Equal(t, flags, test.flags)
	}
}

func TestFreestandingFlags(t *testing.T) {
	// The panic mode is "exit" so that the trace flags, which depend on the
	// C compiler name, stay out of the comparison.
	hosted, err := newCompileOptions(Options{PanicMode: "exit"})
	be.Err(t, err, nil)
	be.Equal(t, hosted.flags, nil)

	bare, err := newCompileOptions(Options{PanicMode: "exit", Freestanding: true})
	be.Err(t, err, nil)
	be.Equal(t, bare.flags, []string{"-ffreestanding"})
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

func TestSplitList(t *testing.T) {
	tests := []struct {
		s    string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{",", nil},
		{",,", nil},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{" a , b ", []string{"a", "b"}},
		{"a,,b", []string{"a", "b"}},
	}
	for _, test := range tests {
		be.Equal(t, splitList(test.s), test.want)
	}
}
