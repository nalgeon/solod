package strconv_test

import (
	"solod.dev/so/math"
	"solod.dev/so/math/rand"
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

// An atofCase is a test case of ParseFloat. The out field holds the text that
// the shortest format writes for the parsed value.
type atofCase struct {
	in  string
	out string
	err int
}

var atofCases = []atofCase{
	{"", "0", errSyntax},
	{"1", "1", errNone},
	{"+1", "1", errNone},
	{"1x", "0", errSyntax},
	{"1.1.", "0", errSyntax},
	{"1e23", "1e+23", errNone},
	{"1E23", "1e+23", errNone},
	{"100000000000000000000000", "1e+23", errNone},
	{"1e-100", "1e-100", errNone},
	{"123456700", "1.234567e+08", errNone},
	{"99999999999999974834176", "9.999999999999997e+22", errNone},
	{"100000000000000000000001", "1.0000000000000001e+23", errNone},
	{"100000000000000008388608", "1.0000000000000001e+23", errNone},
	{"100000000000000016777215", "1.0000000000000001e+23", errNone},
	{"100000000000000016777216", "1.0000000000000003e+23", errNone},
	{"-1", "-1", errNone},
	{"-0.1", "-0.1", errNone},
	{"-0", "-0", errNone},
	{"1e-20", "1e-20", errNone},
	{"625e-3", "0.625", errNone},

	// Hexadecimal floating-point.
	{"0x1p0", "1", errNone},
	{"0x1p1", "2", errNone},
	{"0x1p-1", "0.5", errNone},
	{"0x1ep-1", "15", errNone},
	{"-0x1ep-1", "-15", errNone},
	{"-0x1_ep-1", "-15", errNone},
	{"0x1p-200", "6.223015277861142e-61", errNone},
	{"0x1p200", "1.6069380442589903e+60", errNone},
	{"0x1fFe2.p0", "131042", errNone},
	{"0x1fFe2.P0", "131042", errNone},
	{"-0x2p3", "-16", errNone},
	{"0x0.fp4", "15", errNone},
	{"0x0.fp0", "0.9375", errNone},
	{"0x1e2", "0", errSyntax},
	{"1p2", "0", errSyntax},

	// zeros
	{"0", "0", errNone},
	{"0e0", "0", errNone},
	{"-0e0", "-0", errNone},
	{"+0e0", "0", errNone},
	{"0e-0", "0", errNone},
	{"-0e-0", "-0", errNone},
	{"+0e-0", "0", errNone},
	{"0e+0", "0", errNone},
	{"-0e+0", "-0", errNone},
	{"+0e+0", "0", errNone},
	{"0e+01234567890123456789", "0", errNone},
	{"0.00e-01234567890123456789", "0", errNone},
	{"-0e+01234567890123456789", "-0", errNone},
	{"-0.00e-01234567890123456789", "-0", errNone},
	{"0x0p+01234567890123456789", "0", errNone},
	{"0x0.00p-01234567890123456789", "0", errNone},
	{"-0x0p+01234567890123456789", "-0", errNone},
	{"-0x0.00p-01234567890123456789", "-0", errNone},

	{"0e291", "0", errNone}, // issue 15364
	{"0e292", "0", errNone}, // issue 15364
	{"0e347", "0", errNone}, // issue 15364
	{"0e348", "0", errNone}, // issue 15364
	{"-0e291", "-0", errNone},
	{"-0e292", "-0", errNone},
	{"-0e347", "-0", errNone},
	{"-0e348", "-0", errNone},
	{"0x0p126", "0", errNone},
	{"0x0p127", "0", errNone},
	{"0x0p128", "0", errNone},
	{"0x0p129", "0", errNone},
	{"0x0p130", "0", errNone},
	{"0x0p1022", "0", errNone},
	{"0x0p1023", "0", errNone},
	{"0x0p1024", "0", errNone},
	{"0x0p1025", "0", errNone},
	{"0x0p1026", "0", errNone},
	{"-0x0p126", "-0", errNone},
	{"-0x0p127", "-0", errNone},
	{"-0x0p128", "-0", errNone},
	{"-0x0p129", "-0", errNone},
	{"-0x0p130", "-0", errNone},
	{"-0x0p1022", "-0", errNone},
	{"-0x0p1023", "-0", errNone},
	{"-0x0p1024", "-0", errNone},
	{"-0x0p1025", "-0", errNone},
	{"-0x0p1026", "-0", errNone},

	// NaNs
	{"nan", "NaN", errNone},
	{"NaN", "NaN", errNone},
	{"NAN", "NaN", errNone},

	// Infs
	{"inf", "+Inf", errNone},
	{"-Inf", "-Inf", errNone},
	{"+INF", "+Inf", errNone},
	{"-Infinity", "-Inf", errNone},
	{"+INFINITY", "+Inf", errNone},
	{"Infinity", "+Inf", errNone},

	// largest float64
	{"1.7976931348623157e308", "1.7976931348623157e+308", errNone},
	{"-1.7976931348623157e308", "-1.7976931348623157e+308", errNone},
	{"0x1.fffffffffffffp1023", "1.7976931348623157e+308", errNone},
	{"-0x1.fffffffffffffp1023", "-1.7976931348623157e+308", errNone},
	{"0x1fffffffffffffp+971", "1.7976931348623157e+308", errNone},
	{"-0x1fffffffffffffp+971", "-1.7976931348623157e+308", errNone},
	{"0x.1fffffffffffffp1027", "1.7976931348623157e+308", errNone},
	{"-0x.1fffffffffffffp1027", "-1.7976931348623157e+308", errNone},

	// next float64 - too large
	{"1.7976931348623159e308", "+Inf", errRange},
	{"-1.7976931348623159e308", "-Inf", errRange},
	{"0x1p1024", "+Inf", errRange},
	{"-0x1p1024", "-Inf", errRange},
	{"0x2p1023", "+Inf", errRange},
	{"-0x2p1023", "-Inf", errRange},
	{"0x.1p1028", "+Inf", errRange},
	{"-0x.1p1028", "-Inf", errRange},
	{"0x.2p1027", "+Inf", errRange},
	{"-0x.2p1027", "-Inf", errRange},

	// the border is ...158079
	// borderline - okay
	{"1.7976931348623158e308", "1.7976931348623157e+308", errNone},
	{"-1.7976931348623158e308", "-1.7976931348623157e+308", errNone},
	{"0x1.fffffffffffff7fffp1023", "1.7976931348623157e+308", errNone},
	{"-0x1.fffffffffffff7fffp1023", "-1.7976931348623157e+308", errNone},
	// borderline - too large
	{"1.797693134862315808e308", "+Inf", errRange},
	{"-1.797693134862315808e308", "-Inf", errRange},
	{"0x1.fffffffffffff8p1023", "+Inf", errRange},
	{"-0x1.fffffffffffff8p1023", "-Inf", errRange},
	{"0x1fffffffffffff.8p+971", "+Inf", errRange},
	{"-0x1fffffffffffff8p+967", "-Inf", errRange},
	{"0x.1fffffffffffff8p1027", "+Inf", errRange},
	{"-0x.1fffffffffffff9p1027", "-Inf", errRange},

	// a little too large
	{"1e308", "1e+308", errNone},
	{"2e308", "+Inf", errRange},
	{"1e309", "+Inf", errRange},
	{"0x1p1025", "+Inf", errRange},

	// way too large
	{"1e310", "+Inf", errRange},
	{"-1e310", "-Inf", errRange},
	{"1e400", "+Inf", errRange},
	{"-1e400", "-Inf", errRange},
	{"1e400000", "+Inf", errRange},
	{"-1e400000", "-Inf", errRange},
	{"0x1p1030", "+Inf", errRange},
	{"0x1p2000", "+Inf", errRange},
	{"0x1p2000000000", "+Inf", errRange},
	{"-0x1p1030", "-Inf", errRange},
	{"-0x1p2000", "-Inf", errRange},
	{"-0x1p2000000000", "-Inf", errRange},

	// denormalized
	{"1e-305", "1e-305", errNone},
	{"1e-306", "1e-306", errNone},
	{"1e-307", "1e-307", errNone},
	{"1e-308", "1e-308", errNone},
	{"1e-309", "1e-309", errNone},
	{"1e-310", "1e-310", errNone},
	{"1e-322", "1e-322", errNone},
	// smallest denormal
	{"5e-324", "5e-324", errNone},
	{"4e-324", "5e-324", errNone},
	{"3e-324", "5e-324", errNone},
	// too small
	{"2e-324", "0", errNone},
	// way too small
	{"1e-350", "0", errNone},
	{"1e-400000", "0", errNone},

	// Near denormals and denormals.
	{"0x2.00000000000000p-1010", "1.8227805048890994e-304", errNone}, // 0x00e0000000000000
	{"0x1.fffffffffffff0p-1010", "1.8227805048890992e-304", errNone}, // 0x00dfffffffffffff
	{"0x1.fffffffffffff7p-1010", "1.8227805048890992e-304", errNone}, // rounded down
	{"0x1.fffffffffffff8p-1010", "1.8227805048890994e-304", errNone}, // rounded up
	{"0x1.fffffffffffff9p-1010", "1.8227805048890994e-304", errNone}, // rounded up

	{"0x2.00000000000000p-1022", "4.450147717014403e-308", errNone},  // 0x0020000000000000
	{"0x1.fffffffffffff0p-1022", "4.4501477170144023e-308", errNone}, // 0x001fffffffffffff
	{"0x1.fffffffffffff7p-1022", "4.4501477170144023e-308", errNone}, // rounded down
	{"0x1.fffffffffffff8p-1022", "4.450147717014403e-308", errNone},  // rounded up
	{"0x1.fffffffffffff9p-1022", "4.450147717014403e-308", errNone},  // rounded up

	{"0x1.00000000000000p-1022", "2.2250738585072014e-308", errNone}, // 0x0010000000000000
	{"0x0.fffffffffffff0p-1022", "2.225073858507201e-308", errNone},  // 0x000fffffffffffff
	{"0x0.ffffffffffffe0p-1022", "2.2250738585072004e-308", errNone}, // 0x000ffffffffffffe
	{"0x0.ffffffffffffe7p-1022", "2.2250738585072004e-308", errNone}, // rounded down
	{"0x1.ffffffffffffe8p-1023", "2.225073858507201e-308", errNone},  // rounded up
	{"0x1.ffffffffffffe9p-1023", "2.225073858507201e-308", errNone},  // rounded up

	{"0x0.00000003fffff0p-1022", "2.072261e-317", errNone},  // 0x00000000003fffff
	{"0x0.00000003456780p-1022", "1.694649e-317", errNone},  // 0x0000000000345678
	{"0x0.00000003456787p-1022", "1.694649e-317", errNone},  // rounded down
	{"0x0.00000003456788p-1022", "1.694649e-317", errNone},  // rounded down (half to even)
	{"0x0.00000003456790p-1022", "1.6946496e-317", errNone}, // 0x0000000000345679
	{"0x0.00000003456789p-1022", "1.6946496e-317", errNone}, // rounded up

	{"0x0.0000000345678800000000000000000000000001p-1022", "1.6946496e-317", errNone}, // rounded up

	{"0x0.000000000000f0p-1022", "7.4e-323", errNone}, // 0x000000000000000f
	{"0x0.00000000000060p-1022", "3e-323", errNone},   // 0x0000000000000006
	{"0x0.00000000000058p-1022", "3e-323", errNone},   // rounded up
	{"0x0.00000000000057p-1022", "2.5e-323", errNone}, // rounded down
	{"0x0.00000000000050p-1022", "2.5e-323", errNone}, // 0x0000000000000005

	{"0x0.00000000000010p-1022", "5e-324", errNone},  // 0x0000000000000001
	{"0x0.000000000000081p-1022", "5e-324", errNone}, // rounded up
	{"0x0.00000000000008p-1022", "0", errNone},       // rounded down
	{"0x0.00000000000007fp-1022", "0", errNone},      // rounded down

	// try to overflow exponent
	{"1e-4294967296", "0", errNone},
	{"1e+4294967296", "+Inf", errRange},
	{"1e-18446744073709551616", "0", errNone},
	{"1e+18446744073709551616", "+Inf", errRange},
	{"0x1p-4294967296", "0", errNone},
	{"0x1p+4294967296", "+Inf", errRange},
	{"0x1p-18446744073709551616", "0", errNone},
	{"0x1p+18446744073709551616", "+Inf", errRange},

	// Parse errors
	{"1e", "0", errSyntax},
	{"1e-", "0", errSyntax},
	{".e-1", "0", errSyntax},
	{"1\x00.2", "0", errSyntax},
	{"0x", "0", errSyntax},
	{"0x.", "0", errSyntax},
	{"0x1", "0", errSyntax},
	{"0x.1", "0", errSyntax},
	{"0x1p", "0", errSyntax},
	{"0x.1p", "0", errSyntax},
	{"0x1p+", "0", errSyntax},
	{"0x.1p+", "0", errSyntax},
	{"0x1p-", "0", errSyntax},
	{"0x.1p-", "0", errSyntax},
	{"0x1p+2", "4", errNone},
	{"0x.1p+2", "0.25", errNone},
	{"0x1p-2", "0.25", errNone},
	{"0x.1p-2", "0.015625", errNone},

	// https://www.exploringbinary.com/java-hangs-when-converting-2-2250738585072012e-308/
	{"2.2250738585072012e-308", "2.2250738585072014e-308", errNone},
	// https://www.exploringbinary.com/php-hangs-on-numeric-value-2-2250738585072011e-308/
	{"2.2250738585072011e-308", "2.225073858507201e-308", errNone},

	// A very large number (initially wrongly parsed by the fast algorithm).
	{"4.630813248087435e+307", "4.630813248087435e+307", errNone},

	// A different kind of very large number.
	{"22.222222222222222", "22.22222222222222", errNone},
	{"0x1.1111111111111p222", "7.18931911124017e+66", errNone},
	{"0x2.2222222222222p221", "7.18931911124017e+66", errNone},

	// Exactly halfway between 1 and the float64 after it.
	// Round to even (down).
	{"1.00000000000000011102230246251565404236316680908203125", "1", errNone},
	{"0x1.00000000000008p0", "1", errNone},
	// Slightly lower; still round down.
	{"1.00000000000000011102230246251565404236316680908203124", "1", errNone},
	{"0x1.00000000000007Fp0", "1", errNone},
	// Slightly higher; round up.
	{"1.00000000000000011102230246251565404236316680908203126", "1.0000000000000002", errNone},
	{"0x1.000000000000081p0", "1.0000000000000002", errNone},
	{"0x1.00000000000009p0", "1.0000000000000002", errNone},

	// Halfway between the two float64 values after 1.
	// Round to even (up).
	{"1.00000000000000033306690738754696212708950042724609375", "1.0000000000000004", errNone},
	{"0x1.00000000000018p0", "1.0000000000000004", errNone},

	// Halfway between 1090544144181609278303144771584 and 1090544144181609419040633126912
	// (15497564393479157p+46, should round to even 15497564393479156p+46, issue 36657)
	{"1090544144181609348671888949248", "1.0905441441816093e+30", errNone},
	// slightly above, rounds up
	{"1090544144181609348835077142190", "1.0905441441816094e+30", errNone},

	// Underscores.
	{"1_23.50_0_0e+1_2", "1.235e+14", errNone},
	{"-_123.5e+12", "0", errSyntax},
	{"+_123.5e+12", "0", errSyntax},
	{"_123.5e+12", "0", errSyntax},
	{"1__23.5e+12", "0", errSyntax},
	{"123_.5e+12", "0", errSyntax},
	{"123._5e+12", "0", errSyntax},
	{"123.5_e+12", "0", errSyntax},
	{"123.5__0e+12", "0", errSyntax},
	{"123.5e_+12", "0", errSyntax},
	{"123.5e+_12", "0", errSyntax},
	{"123.5e_-12", "0", errSyntax},
	{"123.5e-_12", "0", errSyntax},
	{"123.5e+1__2", "0", errSyntax},
	{"123.5e+12_", "0", errSyntax},

	{"0x_1_2.3_4_5p+1_2", "74565", errNone},
	{"-_0x12.345p+12", "0", errSyntax},
	{"+_0x12.345p+12", "0", errSyntax},
	{"_0x12.345p+12", "0", errSyntax},
	{"0x__12.345p+12", "0", errSyntax},
	{"0x1__2.345p+12", "0", errSyntax},
	{"0x12_.345p+12", "0", errSyntax},
	{"0x12._345p+12", "0", errSyntax},
	{"0x12.3__45p+12", "0", errSyntax},
	{"0x12.345_p+12", "0", errSyntax},
	{"0x12.345p_+12", "0", errSyntax},
	{"0x12.345p+_12", "0", errSyntax},
	{"0x12.345p_-12", "0", errSyntax},
	{"0x12.345p-_12", "0", errSyntax},
	{"0x12.345p+1__2", "0", errSyntax},
	{"0x12.345p+12_", "0", errSyntax},

	{"1e100x", "0", errSyntax},
	{"1e1000x", "0", errSyntax},
}

// atof32Cases run through ParseFloat with a bit size of 32 alone.
var atof32Cases = []atofCase{
	// Hex
	{"0x1p-100", "7.888609e-31", errNone},
	{"0x1p100", "1.2676506e+30", errNone},

	// Exactly halfway between 1 and the next float32.
	// Round to even (down).
	{"1.000000059604644775390625", "1", errNone},
	{"0x1.000001p0", "1", errNone},
	// Slightly lower.
	{"1.000000059604644775390624", "1", errNone},
	{"0x1.0000008p0", "1", errNone},
	{"0x1.000000fp0", "1", errNone},
	// Slightly higher.
	{"1.000000059604644775390626", "1.0000001", errNone},
	{"0x1.000002p0", "1.0000001", errNone},
	{"0x1.0000018p0", "1.0000001", errNone},
	{"0x1.0000011p0", "1.0000001", errNone},

	// largest float32: (1<<128) * (1 - 2^-24)
	{"340282346638528859811704183484516925440", "3.4028235e+38", errNone},
	{"-340282346638528859811704183484516925440", "-3.4028235e+38", errNone},
	{"0x.ffffffp128", "3.4028235e+38", errNone},
	{"-0x.ffffffp128", "-3.4028235e+38", errNone},
	// next float32 - too large
	{"3.4028236e38", "+Inf", errRange},
	{"-3.4028236e38", "-Inf", errRange},
	{"0x1.0p128", "+Inf", errRange},
	{"-0x1.0p128", "-Inf", errRange},
	// the border is 3.40282356779...e+38
	// borderline - okay
	{"3.402823567e38", "3.4028235e+38", errNone},
	{"-3.402823567e38", "-3.4028235e+38", errNone},
	{"0x.ffffff7fp128", "3.4028235e+38", errNone},
	{"-0x.ffffff7fp128", "-3.4028235e+38", errNone},
	// borderline - too large
	{"3.4028235678e38", "+Inf", errRange},
	{"-3.4028235678e38", "-Inf", errRange},
	{"0x.ffffff8p128", "+Inf", errRange},
	{"-0x.ffffff8p128", "-Inf", errRange},

	// Denormals: less than 2^-126
	{"1e-38", "1e-38", errNone},
	{"1e-39", "1e-39", errNone},
	{"1e-40", "1e-40", errNone},
	{"1e-41", "1e-41", errNone},
	{"1e-42", "1e-42", errNone},
	{"1e-43", "1e-43", errNone},
	{"1e-44", "1e-44", errNone},
	{"6e-45", "6e-45", errNone}, // 4p-149 = 5.6e-45
	{"5e-45", "6e-45", errNone},

	// Smallest denormal
	{"1e-45", "1e-45", errNone}, // 1p-149 = 1.4e-45
	{"2e-45", "1e-45", errNone},
	{"3e-45", "3e-45", errNone},

	// Near denormals and denormals.
	{"0x0.89aBcDp-125", "1.2643093e-38", errNone},  // 0x0089abcd
	{"0x0.8000000p-125", "1.1754944e-38", errNone}, // 0x00800000
	{"0x0.1234560p-125", "1.671814e-39", errNone},  // 0x00123456
	{"0x0.1234567p-125", "1.671814e-39", errNone},  // rounded down
	{"0x0.1234568p-125", "1.671814e-39", errNone},  // rounded down
	{"0x0.1234569p-125", "1.671815e-39", errNone},  // rounded up
	{"0x0.1234570p-125", "1.671815e-39", errNone},  // 0x00123457
	{"0x0.0000010p-125", "1e-45", errNone},         // 0x00000001
	{"0x0.00000081p-125", "1e-45", errNone},        // rounded up
	{"0x0.0000008p-125", "0", errNone},             // rounded down
	{"0x0.0000007p-125", "0", errNone},             // rounded down

	// 2^92 = 8388608p+69 = 4951760157141521099596496896 (4.9517602e27)
	// is an exact power of two that needs 8 decimal digits to be correctly
	// parsed back.
	// The float32 before is 16777215p+68 = 4.95175986e+27
	// The halfway is 4.951760009. A bad algorithm that thinks the previous
	// float32 is 8388607p+69 will shorten incorrectly to 4.95176e+27.
	{"4951760157141521099596496896", "4.9517602e+27", errNone},
}

func TestAtof(t *testing.T) {
	buf := make([]byte, bufLen)
	for i, tc := range atofCases {
		out, err := strconv.ParseFloat(tc.in, 64)
		outs := strconv.FormatFloat(buf, out, 'g', -1, 64)
		if outs != tc.out || errCode(err) != tc.err {
			t.Errorf("case %d: ParseFloat(%s, 64) = %s, %s, want %s, %s",
				i, tc.in, outs, errName(errCode(err)), tc.out, errName(tc.err))
		}
		// A result that a float32 holds exactly must parse the same way at
		// the narrower width.
		if float64(float32(out)) != out {
			continue
		}
		out, err = strconv.ParseFloat(tc.in, 32)
		out32 := float32(out)
		if float64(out32) != out {
			t.Errorf("case %d: ParseFloat(%s, 32) = %g, not a float32", i, tc.in, out)
			continue
		}
		outs = strconv.FormatFloat(buf, float64(out32), 'g', -1, 32)
		if outs != tc.out || errCode(err) != tc.err {
			t.Errorf("case %d: ParseFloat(%s, 32) = %s, %s, want %s, %s",
				i, tc.in, outs, errName(errCode(err)), tc.out, errName(tc.err))
		}
	}
}

func TestAtof32(t *testing.T) {
	buf := make([]byte, bufLen)
	for i, tc := range atof32Cases {
		out, err := strconv.ParseFloat(tc.in, 32)
		out32 := float32(out)
		if float64(out32) != out {
			t.Errorf("case %d: ParseFloat(%s, 32) = %g, not a float32", i, tc.in, out)
			continue
		}
		outs := strconv.FormatFloat(buf, float64(out32), 'g', -1, 32)
		if outs != tc.out || errCode(err) != tc.err {
			t.Errorf("case %d: ParseFloat(%s, 32) = %s, %s, want %s, %s",
				i, tc.in, outs, errName(errCode(err)), tc.out, errName(tc.err))
		}
	}
}

// atofRandomCount is the number of random float64 values of TestAtofRandom.
const atofRandomCount = 10000

// TestAtofRandom formats a random bit pattern and parses the text back. The
// round trip must return the value the format started from.
func TestAtofRandom(t *testing.T) {
	buf := make([]byte, bufLen)
	for range atofRandomCount {
		bits := uint64(rand.Uint32())<<32 | uint64(rand.Uint32())
		x := math.Float64frombits(bits)
		s := strconv.FormatFloat(buf, x, 'g', -1, 64)
		got, _ := strconv.ParseFloat(s, 64)
		if got == x || (math.IsNaN(x) && math.IsNaN(got)) {
			continue
		}
		t.Errorf("number %s badly parsed as %g (expected %g)", s, got, x)
	}
}

func TestRoundTrip(t *testing.T) {
	// Two values that an 80 bit register would round twice. Issue 2917.
	buf := make([]byte, bufLen)
	// The values are 8865794286000691<<39 and 8865794286000692<<39, which no
	// integer constant of C can hold.
	lo := 8865794286000691 * pow2(39)
	hi := 8865794286000692 * pow2(39)
	if s := strconv.FormatFloat(buf, lo, 'g', -1, 64); s != "4.87402195346389e+27" {
		t.Errorf("FormatFloat(8865794286000691<<39) = %s, want 4.87402195346389e+27", s)
	}
	if s := strconv.FormatFloat(buf, hi, 'g', -1, 64); s != "4.8740219534638903e+27" {
		t.Errorf("FormatFloat(8865794286000692<<39) = %s, want 4.8740219534638903e+27", s)
	}
	if f, err := strconv.ParseFloat("4.87402195346389e+27", 64); f != lo || err != nil {
		t.Errorf("ParseFloat(4.87402195346389e+27) = %g, want %g", f, lo)
	}
	if f, err := strconv.ParseFloat("4.8740219534638903e+27", 64); f != hi || err != nil {
		t.Errorf("ParseFloat(4.8740219534638903e+27) = %g, want %g", f, hi)
	}
}

// roundTrip32Step is the distance between the float32 values of the sweep. A
// step of one would run every finite positive float32, which takes too long.
const roundTrip32Step = 9973

func TestRoundTrip32(t *testing.T) {
	// Format a fraction of all finite float32 values and parse the text back.
	// The round trip must return the value the format started from.
	buf := make([]byte, bufLen)
	for i := uint32(0); i < 0xff<<23; i += roundTrip32Step {
		f := math.Float32frombits(i)
		if i&1 == 1 {
			f = -f // negative
		}
		s := strconv.FormatFloat(buf, float64(f), 'g', -1, 32)
		parsed, err := strconv.ParseFloat(s, 32)
		parsed32 := float32(parsed)
		if err != nil {
			t.Errorf("ParseFloat(%s, 32) gave an error", s)
			continue
		}
		if float64(parsed32) != parsed {
			t.Errorf("ParseFloat(%s, 32) = %g, not a float32", s, parsed)
			continue
		}
		if parsed32 != f {
			t.Errorf("ParseFloat(%s, 32) = %g, want %g", s, float64(parsed32), float64(f))
		}
	}
}

// parseFloatBitSizes are the bit sizes that ParseFloat accepts next to 32 and
// 64. A lot of code in the wild calls ParseFloat(s, 10) or ParseFloat(s, 0).
// Issue 42297.
var parseFloatBitSizes = []int{0, 10, 100, 128}

func TestParseFloatIncorrectBitSize(t *testing.T) {
	const in = "1.5e308"
	const want = 1.5e308
	for _, bitSize := range parseFloatBitSizes {
		f, err := strconv.ParseFloat(in, bitSize)
		if err != nil {
			t.Errorf("ParseFloat(%s, %d) gave an error", in, bitSize)
			continue
		}
		if f != want {
			t.Errorf("ParseFloat(%s, %d) = %g, want %g", in, bitSize, f, float64(want))
		}
	}
}
