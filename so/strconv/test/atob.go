package strconv_test

import (
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

// An atobCase is a test case of ParseBool.
type atobCase struct {
	in  string
	out bool
	err int
}

var atobCases = []atobCase{
	{"", false, errSyntax},
	{"asdf", false, errSyntax},
	{"0", false, errNone},
	{"f", false, errNone},
	{"F", false, errNone},
	{"FALSE", false, errNone},
	{"false", false, errNone},
	{"False", false, errNone},
	{"1", true, errNone},
	{"t", true, errNone},
	{"T", true, errNone},
	{"TRUE", true, errNone},
	{"true", true, errNone},
	{"True", true, errNone},
}

func TestParseBool(t *testing.T) {
	for _, tc := range atobCases {
		out, err := strconv.ParseBool(tc.in)
		if out != tc.out || errCode(err) != tc.err {
			t.Errorf("ParseBool(%s) = %t, %s, want %t, %s",
				tc.in, out, errName(errCode(err)), tc.out, errName(tc.err))
		}
	}
}

func TestFormatBool(t *testing.T) {
	if s := strconv.FormatBool(true); s != "true" {
		t.Errorf("FormatBool(true) = %s, want true", s)
	}
	if s := strconv.FormatBool(false); s != "false" {
		t.Errorf("FormatBool(false) = %s, want false", s)
	}
	// MaxBoolLen holds the longer of the two words.
	if strconv.MaxBoolLen != len("false") {
		t.Errorf("MaxBoolLen = %d, want %d", strconv.MaxBoolLen, len("false"))
	}
}

func TestAppendBool(t *testing.T) {
	buf := make([]byte, 0, bufLen)
	buf = append(buf, 'f', 'o', 'o', ' ')
	if b := strconv.AppendBool(buf, true); string(b) != "foo true" {
		t.Errorf("AppendBool(foo , true) = %s, want foo true", string(b))
	}
	if b := strconv.AppendBool(buf, false); string(b) != "foo false" {
		t.Errorf("AppendBool(foo , false) = %s, want foo false", string(b))
	}
}
