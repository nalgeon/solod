package strconv_test

import (
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

// An itoaCase is a test case of FormatInt and AppendInt.
type itoaCase struct {
	in   int64
	base int
	out  string
}

var itoaCases = []itoaCase{
	{0, 10, "0"},
	{1, 10, "1"},
	{-1, 10, "-1"},
	{12345678, 10, "12345678"},
	{-987654321, 10, "-987654321"},
	{1<<31 - 1, 10, "2147483647"},
	{-2147483647, 10, "-2147483647"},
	{1 << 31, 10, "2147483648"},
	{-2147483648, 10, "-2147483648"},
	{1<<31 + 1, 10, "2147483649"},
	{-2147483649, 10, "-2147483649"},
	{1<<32 - 1, 10, "4294967295"},
	{-4294967295, 10, "-4294967295"},
	{1 << 32, 10, "4294967296"},
	{-4294967296, 10, "-4294967296"},
	{1<<32 + 1, 10, "4294967297"},
	{-4294967297, 10, "-4294967297"},
	{1 << 50, 10, "1125899906842624"},
	{1<<63 - 1, 10, "9223372036854775807"},
	{-9223372036854775807, 10, "-9223372036854775807"},
	{-9223372036854775808, 10, "-9223372036854775808"},

	{0, 2, "0"},
	{10, 2, "1010"},
	{-1, 2, "-1"},
	{1 << 15, 2, "1000000000000000"},

	{-8, 8, "-10"},
	{057635436545, 8, "57635436545"},
	{1 << 24, 8, "100000000"},

	{16, 16, "10"},
	{-0x123456789abcdef, 16, "-123456789abcdef"},
	{1<<63 - 1, 16, "7fffffffffffffff"},
	{1<<63 - 1, 2, "111111111111111111111111111111111111111111111111111111111111111"},
	{-9223372036854775808, 2, "-1000000000000000000000000000000000000000000000000000000000000000"},

	{16, 17, "g"},
	{25, 25, "10"},
	{32544027072, 35, "holycow"},
	{38493362624, 36, "holycow"},
}

// A uitoaCase is a test case of FormatUint and AppendUint.
type uitoaCase struct {
	in   uint64
	base int
	out  string
}

var uitoaCases = []uitoaCase{
	{1<<63 - 1, 10, "9223372036854775807"},
	{1 << 63, 10, "9223372036854775808"},
	{1<<63 + 1, 10, "9223372036854775809"},
	{1<<64 - 2, 10, "18446744073709551614"},
	{1<<64 - 1, 10, "18446744073709551615"},
	{1<<64 - 1, 2, "1111111111111111111111111111111111111111111111111111111111111111"},
}

// varlenUints hold one digit more than the value before them. They walk every
// length that the base 10 formatter handles.
var varlenUints = []uitoaCase{
	{1, 10, "1"},
	{12, 10, "12"},
	{123, 10, "123"},
	{1234, 10, "1234"},
	{12345, 10, "12345"},
	{123456, 10, "123456"},
	{1234567, 10, "1234567"},
	{12345678, 10, "12345678"},
	{123456789, 10, "123456789"},
	{1234567890, 10, "1234567890"},
	{12345678901, 10, "12345678901"},
	{123456789012, 10, "123456789012"},
	{1234567890123, 10, "1234567890123"},
	{12345678901234, 10, "12345678901234"},
	{123456789012345, 10, "123456789012345"},
	{1234567890123456, 10, "1234567890123456"},
	{12345678901234567, 10, "12345678901234567"},
	{123456789012345678, 10, "123456789012345678"},
	{1234567890123456789, 10, "1234567890123456789"},
	{12345678901234567890, 10, "12345678901234567890"},
}

func TestFormatInt(t *testing.T) {
	buf := make([]byte, bufLen)
	dst := make([]byte, 0, bufLen)
	dst = append(dst, prefix...)
	for i, tc := range itoaCases {
		if s := strconv.FormatInt(buf, tc.in, tc.base); s != tc.out {
			t.Errorf("case %d: FormatInt(base %d) = %s, want %s", i, tc.base, s, tc.out)
		}
		x := strconv.AppendInt(dst, tc.in, tc.base)
		if appended(x) != tc.out {
			t.Errorf("case %d: AppendInt(%s, base %d) = %s, want %s%s",
				i, prefix, tc.base, string(x), prefix, tc.out)
		}
	}
}

func TestFormatIntAsUint(t *testing.T) {
	// Run the non-negative cases of FormatInt through the
	// unsigned functions, which must write the same text.
	buf := make([]byte, bufLen)
	dst := make([]byte, 0, bufLen)
	for i, tc := range itoaCases {
		if tc.in < 0 {
			continue
		}
		if s := strconv.FormatUint(buf, uint64(tc.in), tc.base); s != tc.out {
			t.Errorf("case %d: FormatUint(base %d) = %s, want %s", i, tc.base, s, tc.out)
		}
		if x := strconv.AppendUint(dst, uint64(tc.in), tc.base); string(x) != tc.out {
			t.Errorf("case %d: AppendUint(base %d) = %s, want %s", i, tc.base, string(x), tc.out)
		}
	}
}

func TestItoa(t *testing.T) {
	// Run the base 10 cases that fit in an int through Itoa.
	buf := make([]byte, bufLen)
	for i, tc := range itoaCases {
		if tc.base != 10 || int64(int(tc.in)) != tc.in {
			continue
		}
		if s := strconv.Itoa(buf, int(tc.in)); s != tc.out {
			t.Errorf("case %d: Itoa = %s, want %s", i, s, tc.out)
		}
	}
}

func TestFormatUint(t *testing.T) {
	buf := make([]byte, bufLen)
	dst := make([]byte, 0, bufLen)
	dst = append(dst, prefix...)
	for i, tc := range uitoaCases {
		if s := strconv.FormatUint(buf, tc.in, tc.base); s != tc.out {
			t.Errorf("case %d: FormatUint(base %d) = %s, want %s", i, tc.base, s, tc.out)
		}
		x := strconv.AppendUint(dst, tc.in, tc.base)
		if appended(x) != tc.out {
			t.Errorf("case %d: AppendUint(%s, base %d) = %s, want %s%s",
				i, prefix, tc.base, string(x), prefix, tc.out)
		}
	}
}

func TestFormatUintVarlen(t *testing.T) {
	buf := make([]byte, bufLen)
	for i, tc := range varlenUints {
		if s := strconv.FormatUint(buf, tc.in, 10); s != tc.out {
			t.Errorf("case %d: FormatUint(base 10) = %s, want %s", i, s, tc.out)
		}
	}
}

// A maxLenCase is a base with a documented buffer length, next to that length
// for a signed and for an unsigned value.
type maxLenCase struct {
	base    int
	maxInt  int
	maxUint int
}

var maxLenBases = []maxLenCase{
	{2, strconv.MaxIntBase2Len, strconv.MaxUintBase2Len},
	{8, strconv.MaxIntBase8Len, strconv.MaxUintBase8Len},
	{10, strconv.MaxIntBase10Len, strconv.MaxUintBase10Len},
	{16, strconv.MaxIntBase16Len, strconv.MaxUintBase16Len},
}

func TestMaxLen(t *testing.T) {
	// Check that the documented buffer lengths hold the longest text that
	// the integer formatters write. The extreme values of the two types give
	// the longest text of every base.
	buf := make([]byte, bufLen)
	for _, tc := range maxLenBases {
		if s := strconv.FormatInt(buf, -9223372036854775808, tc.base); len(s) > tc.maxInt {
			t.Errorf("len(FormatInt(-1<<63, %d)) = %d, over the limit of %d",
				tc.base, len(s), tc.maxInt)
		}
		if s := strconv.FormatInt(buf, 1<<63-1, tc.base); len(s) > tc.maxInt {
			t.Errorf("len(FormatInt(1<<63-1, %d)) = %d, over the limit of %d",
				tc.base, len(s), tc.maxInt)
		}
		if s := strconv.FormatUint(buf, 1<<64-1, tc.base); len(s) > tc.maxUint {
			t.Errorf("len(FormatUint(1<<64-1, %d)) = %d, over the limit of %d",
				tc.base, len(s), tc.maxUint)
		}
	}
}
