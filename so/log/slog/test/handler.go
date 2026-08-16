package slog_test

import (
	"solod.dev/so/log/slog"
	"solod.dev/so/math"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// bufLen is the length of the buffers the format tests write into.
const bufLen = 256

func TestLevelString(t *testing.T) {
	for i, level := range levels {
		if got := level.String(); got != levelNames[i] {
			t.Errorf("Level(%d).String() = %s, want %s", int(level), got, levelNames[i])
		}
	}
}

func TestLevelStringUnknown(t *testing.T) {
	for n := -16; n <= 16; n++ {
		level := slog.Level(n)
		named := false
		for _, l := range levels {
			if level == l {
				named = true
			}
		}
		if named {
			continue
		}
		if got := level.String(); got != "UNKNOWN" {
			t.Errorf("Level(%d).String() = %s, want UNKNOWN", n, got)
		}
	}
}

func TestLevelOrder(t *testing.T) {
	for i := 1; i < len(levels); i++ {
		if levels[i-1] >= levels[i] {
			t.Errorf("%s is not below %s", levelNames[i-1], levelNames[i])
		}
	}
}

func TestHandlerEnabled(t *testing.T) {
	var buf [bufLen]byte
	for hl := -8; hl <= 12; hl++ {
		b := strings.FixedBuilder(buf[:])
		h := slog.NewTextHandler(&b, slog.Level(hl))
		for rl := -8; rl <= 12; rl++ {
			want := rl >= hl
			if got := h.Enabled(slog.Level(rl)); got != want {
				t.Errorf("handler level %d: Enabled(%d) = %t, want %t", hl, rl, got, want)
				return
			}
		}
	}
}

func TestHandleNoAttrs(t *testing.T) {
	var buf, wbuf [bufLen]byte
	got := handle(buf[:], record("hello", nil))
	want := join(wbuf[:], refText, " INFO hello\n")
	if got != want {
		t.Errorf("Handle() wrote %s, want %s", got, want)
	}
}

func TestHandleLevel(t *testing.T) {
	var buf, wbuf [bufLen]byte
	for i, level := range levels {
		r := record("m", nil)
		r.Level = level
		got := handle(buf[:], r)
		want := join(wbuf[:], refText, " ", levelNames[i], " m\n")
		if got != want {
			t.Errorf("Handle() wrote %s, want %s", got, want)
		}
	}
}

func TestHandleLevelUnknown(t *testing.T) {
	var buf, wbuf [bufLen]byte
	r := record("m", nil)
	r.Level = slog.Level(42)
	got := handle(buf[:], r)
	want := join(wbuf[:], refText, " UNKNOWN m\n")
	if got != want {
		t.Errorf("Handle() wrote %s, want %s", got, want)
	}
}

func TestHandleTime(t *testing.T) {
	secs := [...]int64{0, 1, 1700000000, -1, -86400, 253370764800}
	stamps := [...]string{
		"1970-01-01T00:00:00Z",
		"1970-01-01T00:00:01Z",
		"2023-11-14T22:13:20Z",
		"1969-12-31T23:59:59Z",
		"1969-12-31T00:00:00Z",
		"9999-01-01T00:00:00Z",
	}
	var buf, wbuf [bufLen]byte
	for i, sec := range secs {
		r := record("m", nil)
		r.Time = time.Unix(sec, 999999999)
		got := handle(buf[:], r)
		want := join(wbuf[:], stamps[i], " INFO m\n")
		if got != want {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, want)
		}
	}
}

func TestHandleMessage(t *testing.T) {
	var buf, wbuf [bufLen]byte
	for i, msg := range strVals {
		got := handle(buf[:], record(msg, nil))
		want := join(wbuf[:], refText, " INFO ", msg, "\n")
		if got != want {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, want)
		}
	}
}

func TestHandleAttrKinds(t *testing.T) {
	attrs := [...]slog.Attr{
		slog.String("s", "plain"),
		slog.Int("i", 42),
		slog.Int("i", -42),
		slog.Int64("i", math.MinInt64),
		slog.Int64("i", math.MaxInt64),
		slog.Uint64("u", 0),
		slog.Uint64("u", math.MaxUint64),
		slog.Float64("f", 1.5),
		slog.Float64("f", 0),
		slog.Float64("f", -0.25),
		slog.Float64("f", 1e21),
		slog.Float64("f", math.Inf(1)),
		slog.Float64("f", math.Inf(-1)),
		slog.Float64("f", math.NaN()),
		slog.Bool("b", true),
		slog.Bool("b", false),
		slog.Time("t", time.Unix(refSec, 0)),
		slog.Duration("d", 0),
		slog.Duration("d", time.Nanosecond),
		slog.Duration("d", 3*time.Second+500*time.Millisecond),
		slog.Duration("d", -90*time.Minute),
	}
	wants := [...]string{
		"s=plain",
		"i=42",
		"i=-42",
		"i=-9223372036854775808",
		"i=9223372036854775807",
		"u=0",
		"u=18446744073709551615",
		"f=1.5",
		"f=0",
		"f=-0.25",
		"f=1e+21",
		"f=+Inf",
		"f=-Inf",
		"f=NaN",
		"b=true",
		"b=false",
		"t=2023-11-14T22:13:20Z",
		"d=0s",
		"d=1ns",
		"d=3.5s",
		"d=-1h30m0s",
	}
	if len(attrs) != len(wants) {
		t.Fatal("the tables have different lengths")
		return
	}
	var buf [bufLen]byte
	for i := range len(attrs) {
		if got := handleAttr(buf[:], attrs[i]); got != wants[i] {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, wants[i])
		}
	}
}

func TestHandleQuote(t *testing.T) {
	vals := [...]string{
		"plain", "", " ", "a b", "a=b", "=", "a\"b", "\"",
		"tab\there", "line\nfeed", "\x00", "\xff", "日本語",
	}
	wants := [...]string{
		"k=plain",
		"k=\"\"",
		"k=\" \"",
		"k=\"a b\"",
		"k=\"a=b\"",
		"k=\"=\"",
		// TODO: the handler does not escape an embedded quote, so the text of
		// a value with a quote is ambiguous.
		"k=\"a\"b\"",
		"k=\"\"\"",
		"k=tab\there",
		"k=line\nfeed",
		"k=\x00",
		"k=\xff",
		"k=日本語",
	}
	if len(vals) != len(wants) {
		t.Fatal("the tables have different lengths")
		return
	}
	var buf [bufLen]byte
	for i := range len(vals) {
		if got := handleAttr(buf[:], slog.String("k", vals[i])); got != wants[i] {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, wants[i])
		}
	}
}

func TestHandleKey(t *testing.T) {
	var buf, wbuf [bufLen]byte
	for i, key := range strVals {
		got := handleAttr(buf[:], slog.Int(key, 1))
		want := join(wbuf[:], key, "=1")
		if got != want {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, want)
		}
	}
}

func TestHandleManyAttrs(t *testing.T) {
	const digits = "01234567"
	var attrs [len(digits)]slog.Attr
	var wbuf [bufLen]byte
	w := strings.FixedBuilder(wbuf[:])
	w.WriteString(refText)
	w.WriteString(" INFO m")
	for i := range len(attrs) {
		key := digits[i : i+1]
		attrs[i] = slog.Int(key, i)
		w.WriteString(" ")
		w.WriteString(key)
		w.WriteString("=")
		w.WriteString(key)
	}
	w.WriteString("\n")

	var buf [bufLen]byte
	got := handle(buf[:], record("m", attrs[:]))
	if want := w.String(); got != want {
		t.Errorf("Handle() wrote %s, want %s", got, want)
	}
}

func TestHandleEmptyAttrs(t *testing.T) {
	var attrs [1]slog.Attr
	var bufNil, bufEmpty [bufLen]byte
	gotNil := handle(bufNil[:], record("m", nil))
	gotEmpty := handle(bufEmpty[:], record("m", attrs[:0]))
	if gotNil != gotEmpty {
		t.Errorf("Handle() wrote %s for an empty slice, want %s", gotEmpty, gotNil)
	}
}

func TestHandleWriteError(t *testing.T) {
	var w errWriter
	h := slog.NewTextHandler(&w, slog.LevelInfo)
	if err := h.Handle(record("m", nil)); err != errWriteFailed {
		t.Error("Handle() did not give the error of the writer")
	}
}

func TestHandleIface(t *testing.T) {
	var buf, wbuf [bufLen]byte
	b := strings.FixedBuilder(buf[:])
	th := slog.NewTextHandler(&b, slog.LevelWarn)
	var h slog.Handler = &th

	if h.Enabled(slog.LevelInfo) {
		t.Error("Enabled(LevelInfo) = true, want false")
	}
	if !h.Enabled(slog.LevelWarn) {
		t.Error("Enabled(LevelWarn) = false, want true")
	}
	r := record("m", nil)
	r.Level = slog.LevelWarn
	if err := h.Handle(r); err != nil {
		t.Error("Handle() gave an error")
	}
	want := join(wbuf[:], refText, " WARN m\n")
	if got := b.String(); got != want {
		t.Errorf("Handle() wrote %s, want %s", got, want)
	}
}

func TestHandleWriter(t *testing.T) {
	var w countWriter
	h := slog.NewTextHandler(&w, slog.LevelInfo)
	var attrs [1]slog.Attr
	attrs[0] = slog.Int("n", 1)
	_ = h.Handle(record("m", attrs[:]))
	want := len(refText) + len(" INFO m n=1\n")
	if w.bytes != want {
		t.Errorf("Handle() wrote %d bytes, want %d", w.bytes, want)
	}
}
