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
		mode    string
		cc      string
		def     string
		flags   []string
		wantErr bool
	}{
		{"", "cc", traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "gcc", traceDef, []string{"-fno-omit-frame-pointer", "-rdynamic"}, false},
		{"trace", "x86_64-w64-mingw32-gcc", traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"trace", `C:\MinGW\bin\gcc.exe`, traceDef, []string{"-fno-omit-frame-pointer"}, false},
		{"exit", "cc", "-DSO_PANIC_MODE=SO_PANIC_EXIT", nil, false},
		{"abort", "cc", "-DSO_PANIC_MODE=SO_PANIC_ABORT", nil, false},
		{"none", "cc", "", nil, true},
	}
	for _, test := range tests {
		def, flags, err := panicMode(test.mode, test.cc)
		be.Equal(t, err != nil, test.wantErr)
		be.Equal(t, def, test.def)
		be.Equal(t, flags, test.flags)
	}
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
