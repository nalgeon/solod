package slog_test

import (
	"solod.dev/so/errors"
	"solod.dev/so/log/slog"
	"solod.dev/so/strings"
	"solod.dev/so/time"
)

// refSec is the Unix second of the fixed instant of the tests. refText is the
// RFC3339 text of the same instant.
const (
	refSec  = 1700000000
	refText = "2023-11-14T22:13:20Z"
)

// refTime returns the fixed instant of the tests.
func refTime() time.Time {
	return time.Unix(refSec, 0)
}

// levels holds every named level. levelNames holds the name of each level.
var (
	levels     = [...]slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	levelNames = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}
)

// strVals holds the string values of the tests. The list includes the values a
// text handler quotes.
var strVals = [...]string{
	"",
	"a",
	"hello",
	"hello world",
	" ",
	"  lead and trail  ",
	"a=b",
	"=",
	"q\"uo",
	"\"",
	"tab\there",
	"line\nfeed",
	"\x00",
	"\xff",
	"日本語",
}

// errWriteFailed is the error of the writer that fails on purpose.
var errWriteFailed = errors.New("slog_test: write failed")

// errWriter fails every write.
type errWriter struct{}

func (*errWriter) Write(p []byte) (int, error) {
	_ = p
	return 0, errWriteFailed
}

// countWriter counts the bytes and drops them.
type countWriter struct {
	bytes int
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

// maxRecAttrs is the number of the attrs recHandler keeps.
const maxRecAttrs = 8

// recHandler records the last record it handles. It implements [slog.Handler].
type recHandler struct {
	minLevel slog.Level // the lowest level the handler accepts
	ret      error      // the error Handle returns
	calls    int        // the number of the Handle calls
	gotTime  time.Time
	gotLevel slog.Level
	gotMsg   string
	gotAttrs [maxRecAttrs]slog.Attr
	nAttrs   int
}

// Enabled reports whether the handler accepts the level.
func (h *recHandler) Enabled(level slog.Level) bool {
	return level >= h.minLevel
}

// Handle records the fields of r. It copies the attrs, because the slice of
// the caller does not outlive the call.
func (h *recHandler) Handle(r slog.Record) error {
	h.calls++
	h.gotTime = r.Time
	h.gotLevel = r.Level
	h.gotMsg = r.Message
	h.nAttrs = min(len(r.Attrs), maxRecAttrs)
	for i := range h.nAttrs {
		h.gotAttrs[i] = r.Attrs[i]
	}
	return h.ret
}

// join writes the parts into buf and returns the result. The result is a view
// into buf.
//
// The generator never folds a string constant concatenation, so "a" + "b"
// allocates. A test builds an expected value with join instead.
func join(buf []byte, parts ...string) string {
	b := strings.FixedBuilder(buf)
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

// record returns a record with the fixed time and the level LevelInfo.
func record(msg string, attrs []slog.Attr) slog.Record {
	return slog.Record{Time: refTime(), Message: msg, Level: slog.LevelInfo, Attrs: attrs}
}

// handle formats the record with a text handler over buf and returns the text.
// The text is a view into buf.
func handle(buf []byte, r slog.Record) string {
	b := strings.FixedBuilder(buf)
	h := slog.NewTextHandler(&b, slog.LevelDebug)
	_ = h.Handle(r)
	return b.String()
}

// handleAttr formats a record with one attr and returns the text of the attr
// alone. The text is a view into buf.
func handleAttr(buf []byte, a slog.Attr) string {
	var attrs [1]slog.Attr
	attrs[0] = a
	got := handle(buf, record("m", attrs[:]))
	// The prefix is the time, the level and the message, and the suffix is the
	// newline.
	head := len(refText) + len(" INFO m ")
	if len(got) < head+1 {
		return got
	}
	return got[head : len(got)-1]
}
