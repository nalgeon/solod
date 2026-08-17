// Differential test of the format walk against Go's fmt. It extends the test
// of format.go from a single verb to a whole format string, so it covers the
// flags, the width, the precision and '*' as the walk reads them.
//
// A verb of the ported engine sometimes reads as another verb in Go. %u is
// Go's %d with an unsigned value, and %a is Go's %x of a float. Every case
// therefore carries both spellings.

package fmt

import (
	"bytes"
	gofmt "fmt"
	"math"
	"strings"
	"testing"
	"unsafe"
)

func TestPrintfVerbs(t *testing.T) {
	cases := verbCases()
	for _, s := range specs() {
		for _, c := range cases {
			difff(t, s.format(c.verb), []arg{c.arg}, s.format(c.goVerb), c.goArg)
		}
	}
}

func TestPrintfLiteral(t *testing.T) {
	difff(t, "no verbs", nil, "no verbs")
	difff(t, "", nil, "")
	difff(t, "100%%", nil, "100%%")
	difff(t, "%%%%", nil, "%%%%")
	difff(t, "a%db", []arg{{kind: kindInt, i: 1}}, "a%db", 1)
	difff(t, "héllo %s!", []arg{{kind: kindString, s: "世界"}}, "héllo %s!", "世界")
	difff(t, "%d%d%d", []arg{{kind: kindInt, i: 1}, {kind: kindInt, i: 2}, {kind: kindInt, i: 3}},
		"%d%d%d", 1, 2, 3)
	difff(t, "%s=%d (%6.2f)",
		[]arg{{kind: kindString, s: "x"}, {kind: kindInt, i: 42}, {kind: kindFloat, f: 3.14159}},
		"%s=%d (%6.2f)", "x", 42, 3.14159)
}

func TestPrintfStar(t *testing.T) {
	widths := []int{-99999999, -8, -1, 0, 1, 6, 12}
	precs := []int{-1, 0, 1, 3, 8}
	for _, wid := range widths {
		difff(t, "%*d", []arg{{kind: kindInt, i: wid}, {kind: kindInt, i: 42}}, "%*d", wid, 42)
		difff(t, "%-*d|", []arg{{kind: kindInt, i: wid}, {kind: kindInt, i: 42}}, "%-*d|", wid, 42)
		difff(t, "%0*d", []arg{{kind: kindInt, i: wid}, {kind: kindInt, i: 42}}, "%0*d", wid, 42)
		difff(t, "%*s", []arg{{kind: kindInt, i: wid}, {kind: kindString, s: "ab"}}, "%*s", wid, "ab")
		for _, prec := range precs {
			difff(t, "%.*f", []arg{{kind: kindInt, i: prec}, {kind: kindFloat, f: 3.14159}},
				"%.*f", prec, 3.14159)
			difff(t, "%*.*f",
				[]arg{{kind: kindInt, i: wid}, {kind: kindInt, i: prec}, {kind: kindFloat, f: 3.14159}},
				"%*.*f", wid, prec, 3.14159)
			difff(t, "%*.*s",
				[]arg{{kind: kindInt, i: wid}, {kind: kindInt, i: prec}, {kind: kindString, s: "hello"}},
				"%*.*s", wid, prec, "hello")
		}
	}
}

func TestPrintfBadFormat(t *testing.T) {
	cases := []struct {
		format string
		args   []arg
		goArgs []any
	}{
		// A format that ends inside a verb.
		{"%", nil, nil},
		{"abc%", nil, nil},
		{"%12.3", nil, nil},
		{"%-", nil, nil},

		// A verb with no argument.
		{"%d", nil, nil},
		{"%s", nil, nil},
		{"a%fb", nil, nil},

		// A width or a precision with no argument.
		{"%*d", nil, nil},
		{"%.*d", nil, nil},

		// A width or a precision whose argument is not an integer.
		{"%*d", []arg{{kind: kindString, s: "x"}}, []any{"x"}},
		{"%.*d", []arg{{kind: kindString, s: "x"}}, []any{"x"}},

		// A number too long for a width or a precision.
		{"%*d", []arg{{kind: kindInt, i: 99999999}, {kind: kindInt, i: 1}}, []any{99999999, 1}},
	}
	for _, c := range cases {
		difff(t, c.format, c.args, c.format, c.goArgs...)
	}
}

func TestPrintfUnknownVerb(t *testing.T) {
	// Covers a verb the engine does not know. The C shim cannot collect
	// an argument for it, so the walk stops there. Go writes the same
	// text but keeps walking, so these cases carry their own expectation.
	cases := []struct {
		format string
		args   []arg
		want   string
	}{
		{"%q", nil, "%!q(MISSING)"},
		{"%v", nil, "%!v(MISSING)"},
		{"a%qb", nil, "a%!q(MISSING)"},
		{"%d %q %d", []arg{{kind: kindInt, i: 1}}, "1 %!q(MISSING)"},
		{"%世", nil, "%!世(MISSING)"},
	}
	for _, c := range cases {
		got := sprintf(t, c.format, c.args)
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.format, got, c.want)
		}
	}
}

func TestPrintfExtra(t *testing.T) {
	// Covers an argument that the format string does not use. Go writes
	// a %!(EXTRA ...) tail for it, and the port drops the tail: the C shim
	// reads one argument per verb, so it never collects a spare one.
	cases := []struct {
		format string
		args   []arg
		want   string
	}{
		{"%d", []arg{{kind: kindInt, i: 1}, {kind: kindInt, i: 2}}, "1"},
		{"none", []arg{{kind: kindInt, i: 1}}, "none"},
		// A number too long for a width or a precision cuts the format
		// string off, so the argument stays unused.
		{"%99999999d", []arg{{kind: kindInt, i: 1}}, "%!(NOVERB)"},
		{"%.99999999d", []arg{{kind: kindInt, i: 1}}, "%!(NOVERB)"},
	}
	for _, c := range cases {
		got := sprintf(t, c.format, c.args)
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.format, got, c.want)
		}
	}
}

func TestSprint(t *testing.T) {
	cases := []struct {
		size   int
		format string
		args   []arg
		want   string
	}{
		{16, "%d", []arg{{kind: kindInt, i: 42}}, "42"},
		{16, "", nil, ""},
		{0, "%d", []arg{{kind: kindInt, i: 42}}, ""},
		{1, "%d", []arg{{kind: kindInt, i: 42}}, "4"},
		{8, "%s", []arg{{kind: kindString, s: "0123456789"}}, "01234567"},
		{8, "%s", []arg{{kind: kindString, s: "01234567"}}, "01234567"},
		{8, "%s", []arg{{kind: kindString, s: "012345"}}, "012345"},
		// Output longer than the scratch space of the buffer.
		{600, "%f", []arg{{kind: kindFloat, f: 1e300}}, gofmt.Sprintf("%f", 1e300)},
		{300, "%f", []arg{{kind: kindFloat, f: 1e300}}, gofmt.Sprintf("%f", 1e300)[:300]},
	}
	for _, c := range cases {
		store := make([]byte, c.size)
		got := vsprint(store, c.format, c.args)
		if got != c.want {
			t.Errorf("%q into %d bytes: got %q, want %q", c.format, c.size, got, c.want)
		}
	}
}

func TestJoin(t *testing.T) {
	cases := []struct {
		args    []string
		newline bool
		want    string
	}{
		{nil, false, ""},
		{nil, true, "\n"},
		{[]string{"a"}, false, "a"},
		{[]string{"a"}, true, "a\n"},
		{[]string{"a", "b", "c"}, false, "a b c"},
		{[]string{"a", "b", "c"}, true, "a b c\n"},
		{[]string{"", ""}, false, " "},
	}
	for _, c := range cases {
		args := make([]arg, len(c.args))
		for i, s := range c.args {
			args[i] = arg{kind: kindString, s: s}
		}
		var sink bytes.Buffer
		n, err := vjoin(&sink, args, c.newline)
		if err != nil {
			t.Fatalf("%v: write error: %v", c.args, err)
		}
		if sink.String() != c.want {
			t.Errorf("%v: got %q, want %q", c.args, sink.String(), c.want)
		}
		if n != len(c.want) {
			t.Errorf("%v: count = %d, want %d", c.args, n, len(c.want))
		}
	}
}

func TestArgKind(t *testing.T) {
	cases := map[rune]int{
		'd': kindInt, 'b': kindInt, 'o': kindInt, 'O': kindInt, 'x': kindInt, 'X': kindInt,
		'u': kindUint,
		'e': kindFloat, 'E': kindFloat, 'f': kindFloat, 'F': kindFloat,
		'g': kindFloat, 'G': kindFloat, 'a': kindFloat, 'A': kindFloat,
		'c': kindRune, 's': kindString, 't': kindBool, 'p': kindPtr,
		'v': kindNone, 'q': kindNone, 'U': kindNone, 'w': kindNone,
		'%': kindNone, ' ': kindNone, 0: kindNone, '世': kindNone,
	}
	for verb, want := range cases {
		if got := argKind(verb); got != want {
			t.Errorf("argKind(%q) = %d, want %d", verb, got, want)
		}
	}
}

func TestArgKinds(t *testing.T) {
	cases := []struct {
		format string
		stop   bool // an unknown verb stops the walk
	}{
		{"", false}, {"no verbs", false}, {"%%", false}, {"100%% done", false},
		{"%d", false}, {"%s", false}, {"%t", false}, {"%c", false},
		{"%p", false}, {"%u", false}, {"%f", false},
		{"%d %s %f %c %t %p %u", false},
		{"a%db%sc", false}, {"%+08.3f", false}, {"%-#5x|", false},
		{"%*d", false}, {"%.*f", false}, {"%*.*g", false}, {"%-*s|", false},
		{"%*d %s %.*f", false},
		{"%", false}, {"% ", false}, {"%-", false}, {"abc%", false},
		{"%99999999d", false}, {"%.99999999d", false}, {"%d %99999999d %s", false},
		{"%c%%%s", false},
		{"%q", true}, {"%v", true}, {"%d %q %d", true}, {"%世d", true},
		{"%.", true}, // a point at the end is the verb, not a precision
	}
	for _, c := range cases {
		n := argKinds(c.format, nil)
		kinds := make([]int, n)
		if got := argKinds(c.format, kinds); got != n {
			t.Errorf("%q: second count = %d, want %d", c.format, got, n)
		}
		args := make([]arg, n)
		for i, kind := range kinds {
			if kind == kindNone {
				t.Errorf("%q: argument %d has no kind", c.format, i)
			}
			args[i] = sampleArg(kind)
		}

		// The walk must read every collected argument, and no more.
		var sink bytes.Buffer
		var p printer
		p.init(&sink)
		p.doPrintf(c.format, args)
		p.buf.flush()
		if p.argNum != n {
			t.Errorf("%q: the walk read %d arguments, the collector gave %d", c.format, p.argNum, n)
		}
		if !c.stop && strings.Contains(sink.String(), missingString) {
			t.Errorf("%q: output reports a missing argument: %q", c.format, sink.String())
		}
	}
}

// verbCase is one verb with its argument, and the format that gives the same
// output in Go's fmt.
type verbCase struct {
	verb   rune
	goVerb rune
	arg    arg
	goArg  any
}

// verbCases returns one case for every verb and value the engine supports.
func verbCases() []verbCase {
	var out []verbCase

	ints := []int{0, 1, -1, 7, 42, -42, 255, -255, 1 << 40, math.MinInt64, math.MaxInt64}
	for _, verb := range []rune{'d', 'b', 'o', 'O', 'x', 'X'} {
		for _, v := range ints {
			out = append(out, verbCase{verb, verb, arg{kind: kindInt, i: v}, v})
		}
	}

	uints := []uint64{0, 1, 7, 42, 255, 1 << 40, math.MaxUint64}
	for _, v := range uints {
		out = append(out, verbCase{'u', 'd', arg{kind: kindUint, i: int(v)}, v})
	}

	floats := []float64{
		0, math.Copysign(0, -1), 1, -1, 0.5, 3.14159, -3.14159,
		1234.5678, 1e21, 1e300, 1e-300, 5e-324,
		math.NaN(), math.Inf(1), math.Inf(-1),
	}
	goVerbs := map[rune]rune{'e': 'e', 'E': 'E', 'f': 'f', 'F': 'F', 'g': 'g', 'G': 'G', 'a': 'x', 'A': 'X'}
	for verb, goVerb := range goVerbs {
		for _, v := range floats {
			out = append(out, verbCase{verb, goVerb, arg{kind: kindFloat, f: v}, v})
		}
	}

	for _, v := range []rune{0, 'A', '0', 'é', '世', 0x10FFFF} {
		out = append(out, verbCase{'c', 'c', arg{kind: kindRune, i: int(v)}, v})
	}

	for _, v := range []string{"", "a", "hello", "héllo", "世界"} {
		out = append(out, verbCase{'s', 's', arg{kind: kindString, s: v}, v})
	}

	for _, v := range []bool{true, false} {
		i := 0
		if v {
			i = 1
		}
		out = append(out, verbCase{'t', 't', arg{kind: kindBool, i: i}, v})
	}

	ptr := new(int)
	out = append(out, verbCase{'p', 'p', arg{kind: kindPtr, i: int(uintptr(unsafe.Pointer(ptr)))}, ptr})

	return out
}

// sprintf runs the format walk and returns its output.
func sprintf(t *testing.T, format string, args []arg) string {
	t.Helper()
	var sink bytes.Buffer
	n, err := vfprint(&sink, format, args)
	if err != nil {
		t.Fatalf("%q: write error: %v", format, err)
	}
	if n != sink.Len() {
		t.Fatalf("%q: count = %d, want %d", format, n, sink.Len())
	}
	return sink.String()
}

// difff compares the output of the format walk to Go's fmt.
func difff(t *testing.T, format string, args []arg, goFormat string, goArgs ...any) {
	t.Helper()
	got := sprintf(t, format, args)
	want := gofmt.Sprintf(goFormat, goArgs...)
	if got != want {
		t.Errorf("%q of %v: got %q, want %q", format, goArgs, got, want)
	}
}

// sampleArg returns an argument of the given kind.
func sampleArg(kind int) arg {
	switch kind {
	case kindInt:
		return arg{kind: kind, i: -42}
	case kindUint:
		return arg{kind: kind, i: 42}
	case kindFloat:
		return arg{kind: kind, f: 1.5}
	case kindRune:
		return arg{kind: kind, i: 'A'}
	case kindString:
		return arg{kind: kind, s: "hi"}
	case kindBool:
		return arg{kind: kind, i: 1}
	case kindPtr:
		return arg{kind: kind, i: 4096}
	}
	return arg{}
}
