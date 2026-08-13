// Exact-output tests for the formatting verbs. Each case holds a format, an
// argument and the exact text that the engine must write.

package fmt_test

import (
	"solod.dev/so/fmt"
	"solod.dev/so/math"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestVerbInt(t *testing.T) {
	buf := make([]byte, 128)
	cases := []intCase{
		{"%d", 42, "42"},
		{"%d", -42, "-42"},
		{"%d", 0, "0"},
		{"%5d", 42, "   42"},
		{"%-5d|", 42, "42   |"},
		{"%05d", 42, "00042"},
		{"%+d", 42, "+42"},
		{"% d", 42, " 42"},
		{"%.5d", 42, "00042"},
		{"%x", 255, "ff"},
		{"%X", 255, "FF"},
		{"%#x", 255, "0xff"},
		{"%08x", 255, "000000ff"},
		{"%o", 8, "10"},
		{"%#o", 8, "010"},

		// A negative value in a base other than 10 prints a sign and the
		// magnitude, not the two's complement.
		{"%x", -1, "-1"},
		{"%o", -1, "-1"},
	}
	for _, c := range cases {
		check(t, c.format, fmt.Sprintf(buf, c.format, c.arg), c.want)
	}

	// %d reads the full width of an int. The width follows the target, so the
	// limits come from a variable: a constant that needs 64 bits does not fit
	// an int on a 32-bit target.
	var zero uint
	maxInt := int(^zero >> 1)
	minInt := -maxInt - 1
	maxWant, minWant := "9223372036854775807", "-9223372036854775808"
	if maxInt == 2147483647 {
		maxWant, minWant = "2147483647", "-2147483648"
	}
	check(t, "%d", fmt.Sprintf(buf, "%d", maxInt), maxWant)
	check(t, "%d", fmt.Sprintf(buf, "%d", minInt), minWant)
}

func TestVerbUint(t *testing.T) {
	buf := make([]byte, 128)
	// %u reads the full width of a uint, same as %d. See TestVerbInt for the
	// reason the limit comes from a variable.
	var zero uint
	maxUint := ^zero
	maxWant := "18446744073709551615"
	if maxUint == 4294967295 {
		maxWant = "4294967295"
	}

	cases := []uintCase{
		{"%u", 42, "42"},
		{"%x", 255, "ff"},
		{"%u", maxUint, maxWant},
	}
	for _, c := range cases {
		check(t, c.format, fmt.Sprintf(buf, c.format, c.arg), c.want)
	}
}

func TestVerbFloat(t *testing.T) {
	buf := make([]byte, 128)
	cases := []floatCase{
		{"%f", 3.14159, "3.141590"},
		{"%f", 1.5, "1.500000"},
		{"%.2f", 3.14159, "3.14"},
		{"%.0f", 3.14159, "3"},
		{"%8.2f", 3.14159, "    3.14"},
		{"%-8.2f|", 3.14159, "3.14    |"},
		{"%08.2f", 3.14159, "00003.14"},
		{"%+f", 3.14159, "+3.141590"},
		{"%e", 1234.5678, "1.234568e+03"},
		{"%E", 1234.5678, "1.234568E+03"},
		{"%.3e", 1234.5678, "1.235e+03"},
		{"%.3g", 1234.5678, "1.23e+03"},
		{"%g", 100000.0, "100000"},
		{"%g", 1000000.0, "1e+06"},
		{"%G", 1e-7, "1E-07"},
		{"%g", 1e300, "1e+300"},
		{"%g", 1e-300, "1e-300"},
		{"%g", math.Copysign(0, -1), "-0"},
		{"%f", math.Copysign(0, -1), "-0.000000"},

		// %g without a precision writes the digits a reader needs to get the
		// value back, not the 6 significant digits of C.
		{"%g", 1234.5678, "1234.5678"},
		{"%g", 3.14159265358979, "3.14159265358979"},

		// A hexadecimal float writes at least two exponent digits in Go.
		{"%a", 1.0, "0x1p+00"},

		// The two special values keep the Go spelling, and an infinity
		// carries a sign.
		{"%f", math.NaN(), "NaN"},
		{"%e", math.NaN(), "NaN"},
		{"%f", math.Inf(1), "+Inf"},
		{"%f", math.Inf(-1), "-Inf"},
		{"%g", math.Inf(1), "+Inf"},
		{"%8.2f", math.NaN(), "     NaN"},
		{"%08.2f", math.Inf(1), "    +Inf"},
	}
	for _, c := range cases {
		check(t, c.format, fmt.Sprintf(buf, c.format, c.arg), c.want)
	}
}

func TestVerbString(t *testing.T) {
	buf := make([]byte, 128)
	cases := []strCase{
		{"%s", "hello", "hello"},
		{"%s", "", ""},
		{"%10s|", "hi", "        hi|"},
		{"%-10s|", "hi", "hi        |"},
		{"%.3s", "hello", "hel"},
		{"%10.3s|", "hello", "       hel|"},
	}
	for _, c := range cases {
		check(t, c.format, fmt.Sprintf(buf, c.format, c.arg), c.want)
	}
}

func TestVerbStringNUL(t *testing.T) {
	// An %s argument with an embedded NUL byte. A So string
	// carries its length, so the byte is part of the output.
	buf := make([]byte, 32)
	if fmt.Sprintf(buf, "[%s]", "a\x00b") != "[a\x00b]" {
		t.Error("%s with a NUL byte: wrong output")
	}
}

func TestVerbRune(t *testing.T) {
	buf := make([]byte, 128)
	cases := []runeCase{
		{"%c", 'A', "A"},
		{"%c", '0', "0"},
		{"%c", '\u20ac', "\u20ac"},
		{"%5c|", 'A', "    A|"},
	}
	for _, c := range cases {
		check(t, c.format, fmt.Sprintf(buf, c.format, c.arg), c.want)
	}
}

func TestVerbBool(t *testing.T) {
	buf := make([]byte, 32)
	check(t, "%t", fmt.Sprintf(buf, "%t", true), "true")
	check(t, "%t", fmt.Sprintf(buf, "%t", false), "false")
	check(t, "%6t|", fmt.Sprintf(buf, "%6t|", true), "  true|")
}

func TestVerbOctalO(t *testing.T) {
	buf := make([]byte, 32)
	check(t, "%O", fmt.Sprintf(buf, "%O", 8), "0o10")
	check(t, "%O", fmt.Sprintf(buf, "%O", -8), "-0o10")
}

func TestVerbPointer(t *testing.T) {
	buf := make([]byte, 32)
	x := 42
	out := fmt.Sprintf(buf, "%p", &x)
	if !strings.HasPrefix(out, "0x") {
		t.Error("%p: no 0x prefix")
	}
	if len(out) < 4 {
		t.Error("%p: too short")
	}
	// A sign extended address holds more digits than a pointer has.
	var zero uint
	digits := 16
	if ^zero == 4294967295 {
		digits = 8
	}
	if len(out) > 2+digits {
		t.Error("%p: too long")
	}
}

func TestVerbUnknown(t *testing.T) {
	// The collector stops at unknown verb, so
	// every argument after it is missing as well.
	buf := make([]byte, 64)
	check(t, "%q", fmt.Sprintf(buf, "%q", "hi"), "%!q(MISSING)")
	check(t, "%d %q", fmt.Sprintf(buf, "%d %q", 42, "hi"), "42 %!q(MISSING)")
	check(t, "%d %v", fmt.Sprintf(buf, "%d %v", 42, 42), "42 %!v(MISSING)")
}

func TestVerbPercent(t *testing.T) {
	buf := make([]byte, 128)
	check(t, "%%", fmt.Sprintf(buf, "100%%"), "100%")
	check(t, "literal", fmt.Sprintf(buf, "no verbs"), "no verbs")
}

func TestVerbWidth(t *testing.T) {
	buf := make([]byte, 128)
	check(t, "%*d", fmt.Sprintf(buf, "%*d", 5, 42), "   42")
	check(t, "%-*d|", fmt.Sprintf(buf, "%-*d|", 5, 42), "42   |")
	check(t, "%.*f", fmt.Sprintf(buf, "%.*f", 2, 3.14159), "3.14")
}

func TestVerbLongFloat(t *testing.T) {
	// An output longer than the usual buffer: %f of 1e300 needs 308 bytes.
	buf := make([]byte, 512)
	out := fmt.Sprintf(buf, "%f", 1e300)
	if len(out) != 308 {
		t.Error("%f of 1e300: wrong length")
	}
	if !strings.HasPrefix(out, "1000000000000000052504760255204420248704468581108") {
		t.Error("%f of 1e300: wrong digits")
	}
	if !strings.HasSuffix(out, ".000000") {
		t.Error("%f of 1e300: wrong tail")
	}
}

func TestVerbTruncate(t *testing.T) {
	// Sprintf truncates silently if an output does not fit the buffer.
	buf := make([]byte, 8)
	check(t, "truncated", fmt.Sprintf(buf, "%s", "0123456789"), "01234567")
}

type intCase struct {
	format string
	arg    int
	want   string
}

type uintCase struct {
	format string
	arg    uint
	want   string
}

type floatCase struct {
	format string
	arg    float64
	want   string
}

type strCase struct {
	format string
	arg    string
	want   string
}

type runeCase struct {
	format string
	arg    rune
	want   string
}

// check reports a case whose output does not match.
func check(t *testing.T, format string, got string, want string) {
	if got == want {
		return
	}
	var sb strings.Builder
	defer sb.Free()
	sb.WriteString(format)
	sb.WriteString(": got [")
	sb.WriteString(got)
	sb.WriteString("], want [")
	sb.WriteString(want)
	sb.WriteString("]")
	t.Error(sb.String())
}
