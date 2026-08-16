package strconv_test

import (
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

// A parseUint64Case is a test case of ParseUint with a base and a bit size.
type parseUint64Case struct {
	in   string
	base int
	out  uint64
	err  int
}

// parseUint64Cases run with base 10.
var parseUint64Cases = []parseUint64Case{
	{"", 10, 0, errSyntax},
	{"0", 10, 0, errNone},
	{"1", 10, 1, errNone},
	{"12345", 10, 12345, errNone},
	{"012345", 10, 12345, errNone},
	{"12345x", 10, 0, errSyntax},
	{"98765432100", 10, 98765432100, errNone},
	{"18446744073709551615", 10, 1<<64 - 1, errNone},
	{"18446744073709551616", 10, 1<<64 - 1, errRange},
	{"18446744073709551620", 10, 1<<64 - 1, errRange},
	{"1_2_3_4_5", 10, 0, errSyntax}, // base=10 so no underscores allowed
	{"_12345", 10, 0, errSyntax},
	{"1__2345", 10, 0, errSyntax},
	{"12345_", 10, 0, errSyntax},
	{"-0", 10, 0, errSyntax},
	{"-1", 10, 0, errSyntax},
	{"+1", 10, 0, errSyntax},
}

// parseUint64BaseCases run with the base of the case.
var parseUint64BaseCases = []parseUint64Case{
	{"", 0, 0, errSyntax},
	{"0", 0, 0, errNone},
	{"0x", 0, 0, errSyntax},
	{"0X", 0, 0, errSyntax},
	{"1", 0, 1, errNone},
	{"12345", 0, 12345, errNone},
	{"012345", 0, 012345, errNone},
	{"0x12345", 0, 0x12345, errNone},
	{"0X12345", 0, 0x12345, errNone},
	{"12345x", 0, 0, errSyntax},
	{"0xabcdefg123", 0, 0, errSyntax},
	{"123456789abc", 0, 0, errSyntax},
	{"98765432100", 0, 98765432100, errNone},
	{"18446744073709551615", 0, 1<<64 - 1, errNone},
	{"18446744073709551616", 0, 1<<64 - 1, errRange},
	{"18446744073709551620", 0, 1<<64 - 1, errRange},
	{"0xFFFFFFFFFFFFFFFF", 0, 1<<64 - 1, errNone},
	{"0x10000000000000000", 0, 1<<64 - 1, errRange},
	{"01777777777777777777777", 0, 1<<64 - 1, errNone},
	{"01777777777777777777778", 0, 0, errSyntax},
	{"02000000000000000000000", 0, 1<<64 - 1, errRange},
	{"0200000000000000000000", 0, 1 << 61, errNone},
	{"0b", 0, 0, errSyntax},
	{"0B", 0, 0, errSyntax},
	{"0b101", 0, 5, errNone},
	{"0B101", 0, 5, errNone},
	{"0o", 0, 0, errSyntax},
	{"0O", 0, 0, errSyntax},
	{"0o377", 0, 255, errNone},
	{"0O377", 0, 255, errNone},

	// underscores allowed with base == 0 only
	{"1_2_3_4_5", 0, 12345, errNone}, // base 0 => 10
	{"_12345", 0, 0, errSyntax},
	{"1__2345", 0, 0, errSyntax},
	{"12345_", 0, 0, errSyntax},

	{"1_2_3_4_5", 10, 0, errSyntax}, // base 10
	{"_12345", 10, 0, errSyntax},
	{"1__2345", 10, 0, errSyntax},
	{"12345_", 10, 0, errSyntax},

	{"0x_1_2_3_4_5", 0, 0x12345, errNone}, // base 0 => 16
	{"_0x12345", 0, 0, errSyntax},
	{"0x__12345", 0, 0, errSyntax},
	{"0x1__2345", 0, 0, errSyntax},
	{"0x1234__5", 0, 0, errSyntax},
	{"0x12345_", 0, 0, errSyntax},

	{"1_2_3_4_5", 16, 0, errSyntax}, // base 16
	{"_12345", 16, 0, errSyntax},
	{"1__2345", 16, 0, errSyntax},
	{"1234__5", 16, 0, errSyntax},
	{"12345_", 16, 0, errSyntax},

	{"0_1_2_3_4_5", 0, 012345, errNone}, // base 0 => 8 (0377)
	{"_012345", 0, 0, errSyntax},
	{"0__12345", 0, 0, errSyntax},
	{"01234__5", 0, 0, errSyntax},
	{"012345_", 0, 0, errSyntax},

	{"0o_1_2_3_4_5", 0, 012345, errNone}, // base 0 => 8 (0o377)
	{"_0o12345", 0, 0, errSyntax},
	{"0o__12345", 0, 0, errSyntax},
	{"0o1234__5", 0, 0, errSyntax},
	{"0o12345_", 0, 0, errSyntax},

	{"0_1_2_3_4_5", 8, 0, errSyntax}, // base 8
	{"_012345", 8, 0, errSyntax},
	{"0__12345", 8, 0, errSyntax},
	{"01234__5", 8, 0, errSyntax},
	{"012345_", 8, 0, errSyntax},

	{"0b_1_0_1", 0, 5, errNone}, // base 0 => 2 (0b101)
	{"_0b101", 0, 0, errSyntax},
	{"0b__101", 0, 0, errSyntax},
	{"0b1__01", 0, 0, errSyntax},
	{"0b10__1", 0, 0, errSyntax},
	{"0b101_", 0, 0, errSyntax},

	{"1_0_1", 2, 0, errSyntax}, // base 2
	{"_101", 2, 0, errSyntax},
	{"1_01", 2, 0, errSyntax},
	{"10_1", 2, 0, errSyntax},
	{"101_", 2, 0, errSyntax},
}

// A parseInt64Case is a test case of ParseInt with a base and a bit size.
type parseInt64Case struct {
	in   string
	base int
	out  int64
	err  int
}

// parseInt64Cases run with base 10.
var parseInt64Cases = []parseInt64Case{
	{"", 10, 0, errSyntax},
	{"0", 10, 0, errNone},
	{"-0", 10, 0, errNone},
	{"+0", 10, 0, errNone},
	{"1", 10, 1, errNone},
	{"-1", 10, -1, errNone},
	{"+1", 10, 1, errNone},
	{"12345", 10, 12345, errNone},
	{"-12345", 10, -12345, errNone},
	{"012345", 10, 12345, errNone},
	{"-012345", 10, -12345, errNone},
	{"98765432100", 10, 98765432100, errNone},
	{"-98765432100", 10, -98765432100, errNone},
	{"9223372036854775807", 10, 1<<63 - 1, errNone},
	{"-9223372036854775807", 10, -(1<<63 - 1), errNone},
	{"9223372036854775808", 10, 1<<63 - 1, errRange},
	{"-9223372036854775808", 10, -9223372036854775808, errNone},
	{"9223372036854775809", 10, 1<<63 - 1, errRange},
	{"-9223372036854775809", 10, -9223372036854775808, errRange},
	{"-1_2_3_4_5", 10, 0, errSyntax}, // base=10 so no underscores allowed
	{"-_12345", 10, 0, errSyntax},
	{"_12345", 10, 0, errSyntax},
	{"1__2345", 10, 0, errSyntax},
	{"12345_", 10, 0, errSyntax},
	{"123%45", 10, 0, errSyntax},
}

// parseInt64BaseCases run with the base of the case.
var parseInt64BaseCases = []parseInt64Case{
	{"", 0, 0, errSyntax},
	{"0", 0, 0, errNone},
	{"-0", 0, 0, errNone},
	{"1", 0, 1, errNone},
	{"-1", 0, -1, errNone},
	{"12345", 0, 12345, errNone},
	{"-12345", 0, -12345, errNone},
	{"012345", 0, 012345, errNone},
	{"-012345", 0, -012345, errNone},
	{"0x12345", 0, 0x12345, errNone},
	{"-0X12345", 0, -0x12345, errNone},
	{"12345x", 0, 0, errSyntax},
	{"-12345x", 0, 0, errSyntax},
	{"98765432100", 0, 98765432100, errNone},
	{"-98765432100", 0, -98765432100, errNone},
	{"9223372036854775807", 0, 1<<63 - 1, errNone},
	{"-9223372036854775807", 0, -(1<<63 - 1), errNone},
	{"9223372036854775808", 0, 1<<63 - 1, errRange},
	{"-9223372036854775808", 0, -9223372036854775808, errNone},
	{"9223372036854775809", 0, 1<<63 - 1, errRange},
	{"-9223372036854775809", 0, -9223372036854775808, errRange},

	// other bases
	{"g", 17, 16, errNone},
	{"10", 25, 25, errNone},
	{"holycow", 35, 32544027072, errNone},
	{"holycow", 36, 38493362624, errNone},

	// base 2
	{"0", 2, 0, errNone},
	{"-1", 2, -1, errNone},
	{"1010", 2, 10, errNone},
	{"1000000000000000", 2, 1 << 15, errNone},
	{"111111111111111111111111111111111111111111111111111111111111111", 2, 1<<63 - 1, errNone},
	{"1000000000000000000000000000000000000000000000000000000000000000", 2, 1<<63 - 1, errRange},
	{"-1000000000000000000000000000000000000000000000000000000000000000", 2, -9223372036854775808, errNone},
	{"-1000000000000000000000000000000000000000000000000000000000000001", 2, -9223372036854775808, errRange},

	// base 8
	{"-10", 8, -8, errNone},
	{"57635436545", 8, 057635436545, errNone},
	{"100000000", 8, 1 << 24, errNone},

	// base 16
	{"10", 16, 16, errNone},
	{"-123456789abcdef", 16, -0x123456789abcdef, errNone},
	{"7fffffffffffffff", 16, 1<<63 - 1, errNone},

	// underscores
	{"-0x_1_2_3_4_5", 0, -0x12345, errNone},
	{"0x_1_2_3_4_5", 0, 0x12345, errNone},
	{"-_0x12345", 0, 0, errSyntax},
	{"_-0x12345", 0, 0, errSyntax},
	{"_0x12345", 0, 0, errSyntax},
	{"0x__12345", 0, 0, errSyntax},
	{"0x1__2345", 0, 0, errSyntax},
	{"0x1234__5", 0, 0, errSyntax},
	{"0x12345_", 0, 0, errSyntax},

	{"-0_1_2_3_4_5", 0, -012345, errNone}, // octal
	{"0_1_2_3_4_5", 0, 012345, errNone},   // octal
	{"-_012345", 0, 0, errSyntax},
	{"_-012345", 0, 0, errSyntax},
	{"_012345", 0, 0, errSyntax},
	{"0__12345", 0, 0, errSyntax},
	{"01234__5", 0, 0, errSyntax},
	{"012345_", 0, 0, errSyntax},

	{"+0xf", 0, 0xf, errNone},
	{"-0xf", 0, -0xf, errNone},
	{"0x+f", 0, 0, errSyntax},
	{"0x-f", 0, 0, errSyntax},
}

// A parseUint32Case is a test case of ParseUint with a bit size of 32.
type parseUint32Case struct {
	in  string
	out uint32
	err int
}

var parseUint32Cases = []parseUint32Case{
	{"", 0, errSyntax},
	{"0", 0, errNone},
	{"1", 1, errNone},
	{"12345", 12345, errNone},
	{"012345", 12345, errNone},
	{"12345x", 0, errSyntax},
	{"987654321", 987654321, errNone},
	{"4294967295", 1<<32 - 1, errNone},
	{"4294967296", 1<<32 - 1, errRange},
	{"1_2_3_4_5", 0, errSyntax}, // base=10 so no underscores allowed
	{"_12345", 0, errSyntax},
	{"1__2345", 0, errSyntax},
	{"12345_", 0, errSyntax},
}

// A parseInt32Case is a test case of ParseInt with a bit size of 32.
type parseInt32Case struct {
	in  string
	out int32
	err int
}

var parseInt32Cases = []parseInt32Case{
	{"", 0, errSyntax},
	{"0", 0, errNone},
	{"-0", 0, errNone},
	{"1", 1, errNone},
	{"-1", -1, errNone},
	{"12345", 12345, errNone},
	{"-12345", -12345, errNone},
	{"012345", 12345, errNone},
	{"-012345", -12345, errNone},
	{"12345x", 0, errSyntax},
	{"-12345x", 0, errSyntax},
	{"987654321", 987654321, errNone},
	{"-987654321", -987654321, errNone},
	{"2147483647", 1<<31 - 1, errNone},
	{"-2147483647", -(1<<31 - 1), errNone},
	{"2147483648", 1<<31 - 1, errRange},
	{"-2147483648", -2147483648, errNone},
	{"2147483649", 1<<31 - 1, errRange},
	{"-2147483649", -2147483648, errRange},
	{"-1_2_3_4_5", 0, errSyntax}, // base=10 so no underscores allowed
	{"-_12345", 0, errSyntax},
	{"_12345", 0, errSyntax},
	{"1__2345", 0, errSyntax},
	{"12345_", 0, errSyntax},
	{"123%45", 0, errSyntax},
}

// A parseArgCase is a test case of an invalid base or bit size.
type parseArgCase struct {
	arg int
	err int
}

var parseBitSizeCases = []parseArgCase{
	{-1, errBitSize},
	{0, errNone},
	{64, errNone},
	{65, errBitSize},
}

var parseBaseCases = []parseArgCase{
	{-1, errBase},
	{0, errNone},
	{1, errBase},
	{2, errNone},
	{36, errNone},
	{37, errBase},
}

// failUint reports the failure of a ParseUint case.
func failUint(t *testing.T, in string, base int, out uint64, err int, want uint64, wantErr int) {
	gotBuf := make([]byte, bufLen)
	wantBuf := make([]byte, bufLen)
	t.Errorf("ParseUint(%s, %d) = %s, %s, want %s, %s",
		in, base,
		strconv.FormatUint(gotBuf, out, 10), errName(err),
		strconv.FormatUint(wantBuf, want, 10), errName(wantErr))
}

// failInt reports the failure of a ParseInt case.
func failInt(t *testing.T, in string, base int, out int64, err int, want int64, wantErr int) {
	gotBuf := make([]byte, bufLen)
	wantBuf := make([]byte, bufLen)
	t.Errorf("ParseInt(%s, %d) = %s, %s, want %s, %s",
		in, base,
		strconv.FormatInt(gotBuf, out, 10), errName(err),
		strconv.FormatInt(wantBuf, want, 10), errName(wantErr))
}

func TestParseUint32(t *testing.T) {
	for _, tc := range parseUint32Cases {
		out, err := strconv.ParseUint(tc.in, 10, 32)
		if out != uint64(tc.out) || errCode(err) != tc.err {
			failUint(t, tc.in, 10, out, errCode(err), uint64(tc.out), tc.err)
		}
	}
}

func TestParseUint64(t *testing.T) {
	for _, tc := range parseUint64Cases {
		out, err := strconv.ParseUint(tc.in, 10, 64)
		if out != tc.out || errCode(err) != tc.err {
			failUint(t, tc.in, 10, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseUint64Base(t *testing.T) {
	for _, tc := range parseUint64BaseCases {
		out, err := strconv.ParseUint(tc.in, tc.base, 64)
		if out != tc.out || errCode(err) != tc.err {
			failUint(t, tc.in, tc.base, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseInt32(t *testing.T) {
	for _, tc := range parseInt32Cases {
		out, err := strconv.ParseInt(tc.in, 10, 32)
		if out != int64(tc.out) || errCode(err) != tc.err {
			failInt(t, tc.in, 10, out, errCode(err), int64(tc.out), tc.err)
		}
	}
}

func TestParseInt64(t *testing.T) {
	for _, tc := range parseInt64Cases {
		out, err := strconv.ParseInt(tc.in, 10, 64)
		if out != tc.out || errCode(err) != tc.err {
			failInt(t, tc.in, 10, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseInt64Base(t *testing.T) {
	for _, tc := range parseInt64BaseCases {
		out, err := strconv.ParseInt(tc.in, tc.base, 64)
		if out != tc.out || errCode(err) != tc.err {
			failInt(t, tc.in, tc.base, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseUintDefaultBitSize(t *testing.T) {
	// Check that a bit size of 0 means the width of an int, which differs
	// between a hosted and a freestanding target.
	if strconv.IntSize == 32 {
		for _, tc := range parseUint32Cases {
			out, err := strconv.ParseUint(tc.in, 10, 0)
			if out != uint64(tc.out) || errCode(err) != tc.err {
				failUint(t, tc.in, 10, out, errCode(err), uint64(tc.out), tc.err)
			}
		}
		return
	}
	for _, tc := range parseUint64Cases {
		out, err := strconv.ParseUint(tc.in, 10, 0)
		if out != tc.out || errCode(err) != tc.err {
			failUint(t, tc.in, 10, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseIntDefaultBitSize(t *testing.T) {
	// Check that a bit size of 0 means the width of an int, which differs
	// between a hosted and a freestanding target.
	if strconv.IntSize == 32 {
		for _, tc := range parseInt32Cases {
			out, err := strconv.ParseInt(tc.in, 10, 0)
			if out != int64(tc.out) || errCode(err) != tc.err {
				failInt(t, tc.in, 10, out, errCode(err), int64(tc.out), tc.err)
			}
		}
		return
	}
	for _, tc := range parseInt64Cases {
		out, err := strconv.ParseInt(tc.in, 10, 0)
		if out != tc.out || errCode(err) != tc.err {
			failInt(t, tc.in, 10, out, errCode(err), tc.out, tc.err)
		}
	}
}

func TestAtoi(t *testing.T) {
	if strconv.IntSize == 32 {
		for _, tc := range parseInt32Cases {
			out, err := strconv.Atoi(tc.in)
			if int64(out) != int64(tc.out) || errCode(err) != tc.err {
				failInt(t, tc.in, 10, int64(out), errCode(err), int64(tc.out), tc.err)
			}
		}
		return
	}
	for _, tc := range parseInt64Cases {
		out, err := strconv.Atoi(tc.in)
		if int64(out) != tc.out || errCode(err) != tc.err {
			failInt(t, tc.in, 10, int64(out), errCode(err), tc.out, tc.err)
		}
	}
}

func TestParseIntBitSize(t *testing.T) {
	for _, tc := range parseBitSizeCases {
		_, err := strconv.ParseInt("0", 0, tc.arg)
		if errCode(err) != tc.err {
			t.Errorf("ParseInt(0, 0, %d) = 0, %s, want 0, %s",
				tc.arg, errName(errCode(err)), errName(tc.err))
		}
	}
}

func TestParseUintBitSize(t *testing.T) {
	for _, tc := range parseBitSizeCases {
		_, err := strconv.ParseUint("0", 0, tc.arg)
		if errCode(err) != tc.err {
			t.Errorf("ParseUint(0, 0, %d) = 0, %s, want 0, %s",
				tc.arg, errName(errCode(err)), errName(tc.err))
		}
	}
}

func TestParseIntBase(t *testing.T) {
	for _, tc := range parseBaseCases {
		_, err := strconv.ParseInt("0", tc.arg, 0)
		if errCode(err) != tc.err {
			t.Errorf("ParseInt(0, %d, 0) = 0, %s, want 0, %s",
				tc.arg, errName(errCode(err)), errName(tc.err))
		}
	}
}

func TestParseUintBase(t *testing.T) {
	for _, tc := range parseBaseCases {
		_, err := strconv.ParseUint("0", tc.arg, 0)
		if errCode(err) != tc.err {
			t.Errorf("ParseUint(0, %d, 0) = 0, %s, want 0, %s",
				tc.arg, errName(errCode(err)), errName(tc.err))
		}
	}
}
