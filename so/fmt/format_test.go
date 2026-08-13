// Differential test of the ported formatter against Go's fmt. A So functional
// test cannot reach Go's fmt, so this check has to be a Go test. The So tests
// in so/fmt/test own the shipped behavior.
//
// Every case builds a format string from a spec, formats a value with the
// ported formatter, and compares the output to what Go's fmt gives for the
// same format string and value.

package fmt

import (
	"bytes"
	gofmt "fmt"
	"math"
	"testing"
)

func TestFmtInteger(t *testing.T) {
	values := []int64{0, 1, -1, 7, 42, -42, 255, -255, 1 << 40, math.MinInt64, math.MaxInt64}
	bases := map[rune]int{'d': 10, 'b': 2, 'o': 8, 'x': 16, 'X': 16}
	for _, s := range specs() {
		for verb, base := range bases {
			digits := ldigits
			if verb == 'X' {
				digits = udigits
			}
			for _, v := range values {
				diff(t, s, verb, v, func(f *formatter) {
					f.fmtInteger(uint64(v), base, signedInt, verb, digits)
				})
			}
		}
	}
}

func TestFmtUnsigned(t *testing.T) {
	values := []uint64{0, 1, 7, 42, 255, 1 << 40, math.MaxUint64}
	bases := map[rune]int{'d': 10, 'b': 2, 'o': 8, 'x': 16, 'X': 16}
	for _, s := range specs() {
		for verb, base := range bases {
			digits := ldigits
			if verb == 'X' {
				digits = udigits
			}
			for _, v := range values {
				diff(t, s, verb, v, func(f *formatter) {
					f.fmtInteger(v, base, unsignedInt, verb, digits)
				})
			}
		}
	}
}

func TestFmtFloat(t *testing.T) {
	values := []float64{
		0, math.Copysign(0, -1), 1, -1, 0.5, 3.14159, -3.14159,
		1234.5678, 1e21, 1e300, 1e-300, 5e-324,
		math.NaN(), math.Inf(1), math.Inf(-1),
	}
	verbs := []rune{'e', 'E', 'f', 'g', 'G', 'x', 'X'}
	for _, s := range specs() {
		for _, verb := range verbs {
			// Go's fmt keeps 6 digits for %e and %f, and the shortest form
			// that reads back exactly for the others.
			prec := -1
			if verb == 'e' || verb == 'E' || verb == 'f' {
				prec = 6
			}
			for _, v := range values {
				diff(t, s, verb, v, func(f *formatter) {
					f.fmtFloat(v, 64, verb, prec)
				})
			}
		}
	}
}

func TestFmtString(t *testing.T) {
	values := []string{"", "a", "hello", "héllo", "世界", "a\x00b"}
	for _, s := range specs() {
		for _, v := range values {
			diff(t, s, 's', v, func(f *formatter) {
				f.fmtS(v)
			})
		}
	}
}

func TestFmtRune(t *testing.T) {
	values := []rune{0, 'A', '0', 'é', '世', 0x10FFFF}
	for _, s := range specs() {
		for _, v := range values {
			diff(t, s, 'c', v, func(f *formatter) {
				f.fmtC(uint64(v))
			})
		}
	}
}

func TestFmtRuneInvalid(t *testing.T) {
	// Covers a code point above the Unicode range, which Go's
	// fmt cannot pass to %c as a rune.
	for _, s := range specs() {
		got := render(t, s, func(f *formatter) {
			f.fmtC(0x110000)
		})
		want := gofmt.Sprintf(s.format('c'), '�')
		if got != want {
			t.Errorf("%s of an invalid rune: got %q, want %q", s.format('c'), got, want)
		}
	}
}

func TestFmtBoolean(t *testing.T) {
	for _, s := range specs() {
		for _, v := range []bool{true, false} {
			diff(t, s, 't', v, func(f *formatter) {
				f.fmtBoolean(v)
			})
		}
	}
}

func TestBufferFlush(t *testing.T) {
	// An output longer than the scratch space.
	var sink bytes.Buffer
	var b buffer
	b.init(&sink)
	for i := range bufSize * 3 {
		b.writeByte(byte('a' + i%26))
	}
	b.writeString("tail")
	b.flush()
	if sink.Len() != bufSize*3+4 {
		t.Fatalf("wrote %d bytes, want %d", sink.Len(), bufSize*3+4)
	}
	if b.total != sink.Len() {
		t.Errorf("count = %d, want %d", b.total, sink.Len())
	}
	out := sink.String()
	for i := range bufSize * 3 {
		if out[i] != byte('a'+i%26) {
			t.Fatalf("byte %d = %q, want %q", i, out[i], byte('a'+i%26))
		}
	}
	if out[bufSize*3:] != "tail" {
		t.Errorf("tail = %q, want %q", out[bufSize*3:], "tail")
	}
}

func TestBufferError(t *testing.T) {
	w := &errWriter{n: bufSize + 10}
	var b buffer
	b.init(w)
	for range bufSize * 3 {
		b.writeByte('x')
	}
	b.flush()
	if b.err == nil {
		t.Fatal("no error reported")
	}
	if b.total != bufSize+10 {
		t.Errorf("count = %d, want %d", b.total, bufSize+10)
	}
}

// spec is one format specification without the verb.
type spec struct {
	minus bool
	plus  bool
	sharp bool
	space bool
	zero  bool
	wid   int // -1 means absent
	prec  int // -1 means absent
}

// format returns the spec as a format string with the verb appended.
func (s spec) format(verb rune) string {
	out := "%"
	if s.minus {
		out += "-"
	}
	if s.plus {
		out += "+"
	}
	if s.sharp {
		out += "#"
	}
	if s.space {
		out += " "
	}
	if s.zero {
		out += "0"
	}
	if s.wid >= 0 {
		out += gofmt.Sprint(s.wid)
	}
	if s.prec >= 0 {
		out += "." + gofmt.Sprint(s.prec)
	}
	return out + string(verb)
}

// apply sets the spec on the formatter.
func (s spec) apply(f *formatter) {
	f.flags = fmtFlags{
		widPresent:  s.wid >= 0,
		precPresent: s.prec >= 0,
		minus:       s.minus,
		plus:        s.plus,
		sharp:       s.sharp,
		space:       s.space,
		zero:        s.zero,
	}
	f.wid = max(s.wid, 0)
	f.prec = max(s.prec, 0)
}

// specs returns every combination of the flags with a few widths and
// precisions. A width of 0 is missing because "%0d" reads as the zero flag,
// not as a width, so Go's fmt cannot express it.
func specs() []spec {
	widths := []int{-1, 1, 6, 12}
	precs := []int{-1, 0, 1, 3, 10}
	var out []spec
	for flags := range 32 {
		for _, wid := range widths {
			for _, prec := range precs {
				out = append(out, spec{
					minus: flags&1 != 0,
					plus:  flags&2 != 0,
					sharp: flags&4 != 0,
					space: flags&8 != 0,
					zero:  flags&16 != 0,
					wid:   wid,
					prec:  prec,
				})
			}
		}
	}
	return out
}

// render formats one value with the ported formatter and returns the output.
func render(t *testing.T, s spec, write func(f *formatter)) string {
	t.Helper()
	var sink bytes.Buffer
	var b buffer
	b.init(&sink)
	var f formatter
	f.init(&b)
	s.apply(&f)
	write(&f)
	b.flush()
	if b.err != nil {
		t.Fatalf("write error: %v", b.err)
	}
	if b.total != sink.Len() {
		t.Fatalf("count = %d, want %d", b.total, sink.Len())
	}
	return sink.String()
}

// diff compares the output of the ported formatter to Go's fmt.
func diff(t *testing.T, s spec, verb rune, arg any, write func(f *formatter)) {
	t.Helper()
	format := s.format(verb)
	got := render(t, s, write)
	want := gofmt.Sprintf(format, arg)
	if got != want {
		t.Errorf("%s of %v: got %q, want %q", format, arg, got, want)
	}
}

// errWriter fails every write after the first n bytes.
type errWriter struct {
	n int
}

func (w *errWriter) Write(p []byte) (int, error) {
	if w.n <= 0 {
		return 0, gofmt.Errorf("full")
	}
	if len(p) > w.n {
		n := w.n
		w.n = 0
		return n, gofmt.Errorf("full")
	}
	w.n -= len(p)
	return len(p), nil
}
