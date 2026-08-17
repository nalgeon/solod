// Every case holds its format string in a variable, which [check] takes as the
// label. go vet parses a format literal against Go's verbs, and it does not
// know %u.

package fmt_test

import (
	"solod.dev/so/fmt"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestArgsMixed(t *testing.T) {
	buf := make([]byte, 128)

	// Every kind in one call.
	f := "%d|%u|%f|%c|%s|%t|%*.*f"
	check(t, f, fmt.Sprintf(buf, f, -1, uint(2), 3.5, 'A', "s", true, 8, 2, 1.5),
		"-1|2|3.500000|A|s|true|    1.50")

	// A string between values of another kind. A string is a struct in the
	// va_list, so it is the argument most likely to move the ones after it.
	f = "%t %s %g %s %u"
	check(t, f, fmt.Sprintf(buf, f, true, "a", 1.5, "b", uint(7)), "true a 1.5 b 7")

	// The kinds in the reverse order.
	f = "%s|%t|%c|%g|%u|%d"
	check(t, f, fmt.Sprintf(buf, f, "s", false, 'A', 3.5, uint(2), -1), "s|false|A|3.5|2|-1")
}

func TestArgsPointer(t *testing.T) {
	// A pointer slot must use up exactly one argument, so the verb after it
	// reads the argument after it.
	buf := make([]byte, 128)
	x := 42
	f := "%d %p %s"
	got := fmt.Sprintf(buf, f, 1, &x, "end")
	if !strings.HasPrefix(got, "1 0x") {
		t.Errorf("%s: got [%s], want a \"1 0x\" prefix", f, got)
	}
	if !strings.HasSuffix(got, " end") {
		t.Errorf("%s: got [%s], want an \" end\" suffix", f, got)
	}
}

func TestArgsWiden(t *testing.T) {
	buf := make([]byte, 128)

	// The print family is nodecay, so every scalar widens at the call site. A
	// signed value keeps its sign, and an unsigned value keeps its value.
	var i8 int8 = -128
	var i16 int16 = -32768
	var i32 int32 = -2147483648
	d := "%d"
	check(t, "int8 "+d, fmt.Sprintf(buf, d, i8), "-128")
	check(t, "int16 "+d, fmt.Sprintf(buf, d, i16), "-32768")
	check(t, "int32 "+d, fmt.Sprintf(buf, d, i32), "-2147483648")

	var u8 uint8 = 255
	var u16 uint16 = 65535
	var u32 uint32 = 4294967295
	u := "%u"
	check(t, "uint8 "+u, fmt.Sprintf(buf, u, u8), "255")
	check(t, "uint16 "+u, fmt.Sprintf(buf, u, u16), "65535")
	check(t, "uint32 "+u, fmt.Sprintf(buf, u, u32), "4294967295")

	var f32 float32 = 1.5
	ff, g := "%f", "%g"
	check(t, "float32 "+ff, fmt.Sprintf(buf, ff, f32), "1.500000")
	check(t, "float32 "+g, fmt.Sprintf(buf, g, f32), "1.5")

	yes := true
	tt := "%t"
	check(t, "bool "+tt, fmt.Sprintf(buf, tt, yes), "true")

	// A widened value keeps the arguments after it in step.
	f := "%d %s %u %f"
	check(t, f, fmt.Sprintf(buf, f, i8, "x", u8, f32), "-128 x 255 1.500000")
}

func TestArgsStar(t *testing.T) {
	buf := make([]byte, 128)

	// A '*' uses a slot of its own before the slot of its verb.
	f := "%*d"
	check(t, f, fmt.Sprintf(buf, f, 5, 42), "   42")
	f = "%*u"
	check(t, f, fmt.Sprintf(buf, f, 4, uint(7)), "   7")
	f = "%.*f"
	check(t, f, fmt.Sprintf(buf, f, 2, 3.14159), "3.14")
	f = "%*.*f"
	check(t, f, fmt.Sprintf(buf, f, 9, 2, 3.14159), "     3.14")
	f = "%*s"
	check(t, f, fmt.Sprintf(buf, f, 5, "ab"), "   ab")
	f = "%*.*s"
	check(t, f, fmt.Sprintf(buf, f, 6, 3, "hello"), "   hel")
	f = "%*c"
	check(t, f, fmt.Sprintf(buf, f, 3, 'A'), "  A")
	f = "%*t"
	check(t, f, fmt.Sprintf(buf, f, 6, true), "  true")

	// A negative width from an argument is a positive width with the minus
	// flag.
	f = "%*d|"
	check(t, f, fmt.Sprintf(buf, f, -5, 42), "42   |")

	// A '*' slot between two verbs.
	f = "%d %*s %d"
	check(t, f, fmt.Sprintf(buf, f, 1, 4, "ab", 2), "1   ab 2")
}

func TestArgsBadNumber(t *testing.T) {
	buf := make([]byte, 64)

	// A width or a precision from an argument the engine cannot use. The verb
	// still writes its own argument.
	f := "%*d"
	check(t, f, fmt.Sprintf(buf, f, 99999999, 42), "%!(BADWIDTH)42")
	f = "%.*f"
	check(t, f, fmt.Sprintf(buf, f, -1, 3.14159), "%!(BADPREC)3.141590")

	// A number in the format string that is too long cuts the format off.
	f = "%99999999d"
	check(t, f, fmt.Sprintf(buf, f, 42), "%!(NOVERB)")

	// A format that ends inside a verb.
	f = "%12.3"
	check(t, f, fmt.Sprintf(buf, f, 42), "%!(NOVERB)")
	f = "%"
	check(t, f, fmt.Sprintf(buf, f, 42), "%!(NOVERB)")
}

func TestArgsByteSlice(t *testing.T) {
	// A []byte has no verb of its own. string(b) is a zero-copy view that
	// carries the length, so %s writes every byte.
	buf := make([]byte, 32)
	b := []byte{'h', 'i', 0, '!'}
	check(t, "[%s] of a []byte", fmt.Sprintf(buf, "[%s]", string(b)), "[hi\x00!]")
}

func TestArgsFormatNUL(t *testing.T) {
	// The format string carries its length, so a NUL byte inside it is
	// ordinary text. The C shim never reads the format as a C string.
	buf := make([]byte, 32)
	check(t, "NUL in the format", fmt.Sprintf(buf, "a\x00b %d", 7), "a\x00b 7")
}
