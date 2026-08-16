package slog_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/errors"
	"solod.dev/so/log/slog"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// errHandleFailed is the error the recording handler gives on request.
var errHandleFailed = errors.New("slog_test: handle failed")

func TestNew(t *testing.T) {
	var buf [bufLen]byte
	b := strings.FixedBuilder(buf[:])
	h := slog.NewTextHandler(&b, slog.LevelInfo)
	l := slog.New(&h)

	var want slog.Handler = &h
	if l.Handler() != want {
		t.Error("Handler() is not the handler of New()")
	}
}

func TestLoggerEnabled(t *testing.T) {
	var h recHandler
	for i, minLevel := range levels {
		h.minLevel = minLevel
		l := slog.New(&h)
		for j, level := range levels {
			want := j >= i
			if got := l.Enabled(level); got != want {
				t.Errorf("logger level %s: Enabled(%s) = %t, want %t",
					levelNames[i], levelNames[j], got, want)
			}
		}
	}
}

func TestLoggerLog(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelDebug
	l := slog.New(&h)

	for i, level := range levels {
		h.calls = 0
		l.Log(level, levelNames[i], slog.Int("n", i))
		if h.calls != 1 {
			t.Errorf("%s: Handle() got %d calls, want 1", levelNames[i], h.calls)
			continue
		}
		if h.gotLevel != level {
			t.Errorf("%s: the record has the level %d, want %d",
				levelNames[i], int(h.gotLevel), int(level))
		}
		if h.gotMsg != levelNames[i] {
			t.Errorf("%s: the record has the message %s", levelNames[i], h.gotMsg)
		}
		if h.nAttrs != 1 || h.gotAttrs[0].Value.Int() != i {
			t.Errorf("%s: the record has the wrong attrs", levelNames[i])
		}
	}
}

func TestLoggerLevelMethods(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelDebug
	l := slog.New(&h)

	for i, level := range levels {
		h.calls = 0
		switch level {
		case slog.LevelDebug:
			l.Debug("m")
		case slog.LevelInfo:
			l.Info("m")
		case slog.LevelWarn:
			l.Warn("m")
		case slog.LevelError:
			l.Error("m")
		}
		if h.calls != 1 {
			t.Errorf("%s: Handle() got %d calls, want 1", levelNames[i], h.calls)
			continue
		}
		if h.gotLevel != level {
			t.Errorf("%s: the record has the level %d, want %d",
				levelNames[i], int(h.gotLevel), int(level))
		}
		if h.gotMsg != "m" {
			t.Errorf("%s: the record has the message %s, want m", levelNames[i], h.gotMsg)
		}
	}
}

func TestLoggerFiltered(t *testing.T) {
	for i, minLevel := range levels {
		var h recHandler
		h.minLevel = minLevel
		l := slog.New(&h)
		for j, level := range levels {
			l.Log(level, "m")
			want := 0
			if j >= i {
				want = 1
			}
			if h.calls != want {
				t.Errorf("handler level %s, record level %s: Handle() got %d calls, want %d",
					levelNames[i], levelNames[j], h.calls, want)
			}
			h.calls = 0
		}
	}
}

func TestLoggerAttrs(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelInfo
	l := slog.New(&h)

	l.Info("m", slog.String("s", "v"), slog.Int("i", 42), slog.Bool("b", true))
	if h.nAttrs != 3 {
		t.Errorf("the record has %d attrs, want 3", h.nAttrs)
		return
	}
	if h.gotAttrs[0].Key != "s" || h.gotAttrs[0].Value.String() != "v" {
		t.Errorf("attr #0 is %s, want s=v", h.gotAttrs[0].Key)
	}
	if h.gotAttrs[1].Key != "i" || h.gotAttrs[1].Value.Int() != 42 {
		t.Errorf("attr #1 is %s, want i=42", h.gotAttrs[1].Key)
	}
	if h.gotAttrs[2].Key != "b" || !h.gotAttrs[2].Value.Bool() {
		t.Errorf("attr #2 is %s, want b=true", h.gotAttrs[2].Key)
	}
}

func TestLoggerNoAttrs(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelInfo
	l := slog.New(&h)

	l.Info("m")
	if h.nAttrs != 0 {
		t.Errorf("the record has %d attrs, want 0", h.nAttrs)
	}
}

func TestLoggerTime(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelInfo
	l := slog.New(&h)

	before := time.Now()
	l.Info("m")
	after := time.Now()
	if h.gotTime.Before(before) {
		t.Error("the record is older than the call")
	}
	if h.gotTime.After(after) {
		t.Error("the record is younger than the call")
	}
}

func TestLoggerHandleError(t *testing.T) {
	var h recHandler
	h.minLevel = slog.LevelInfo
	h.ret = errHandleFailed
	l := slog.New(&h)

	l.Info("m")
	if h.calls != 1 {
		t.Errorf("Handle() got %d calls, want 1", h.calls)
	}
}

func TestLoggerText(t *testing.T) {
	var buf [bufLen]byte
	b := strings.FixedBuilder(buf[:])
	h := slog.NewTextHandler(&b, slog.LevelInfo)
	l := slog.New(&h)

	l.Info("hello world", slog.String("user", "john"), slog.Int("count", 42))
	l.Debug("hidden") // filtered out, below the handler level
	l.Warn("caution")
	l.Error("failure", slog.Float64("elapsed", 1.5), slog.Bool("retry", true))
	l.Info("test quoting", slog.String("msg", "hello world"))

	out := b.String()
	if !strings.Contains(out, "INFO hello world user=john count=42\n") {
		t.Error("missing info line")
	}
	if strings.Contains(out, "hidden") {
		t.Error("debug line should be filtered out")
	}
	if !strings.Contains(out, "WARN caution\n") {
		t.Error("missing warn line")
	}
	if !strings.Contains(out, "ERROR failure elapsed=1.5 retry=true\n") {
		t.Error("missing error line")
	}
	if !strings.Contains(out, `INFO test quoting msg="hello world"`) {
		t.Error("string with spaces should be quoted")
	}

	// Every line carries a time stamp and ends with a newline.
	lines := 0
	for i := range len(out) {
		if out[i] == '\n' {
			lines++
		}
	}
	if lines != 4 {
		t.Errorf("the handler wrote %d lines, want 4", lines)
	}
	if out[len(out)-1] != '\n' {
		t.Error("the last line has no newline")
	}
}

func TestDefault(t *testing.T) {
	d := slog.Default()
	if d == nil {
		t.Fatal("Default() is nil")
		return
	}
	if slog.Default() != d {
		t.Error("Default() gave two different loggers")
	}
	if !d.Enabled(slog.LevelInfo) {
		t.Error("the default logger rejects LevelInfo")
	}
	if d.Enabled(slog.LevelDebug) {
		t.Error("the default logger accepts LevelDebug")
	}
	if d.Handler() == nil {
		t.Error("the default logger has no handler")
	}
}

func TestSetDefault(t *testing.T) {
	saved := slog.Default()
	defer slog.SetDefault(saved)

	var h recHandler
	h.minLevel = slog.LevelDebug
	l := slog.New(&h)
	slog.SetDefault(&l)

	if slog.Default() != &l {
		t.Error("Default() is not the logger of SetDefault()")
		return
	}
	if !slog.Default().Enabled(slog.LevelDebug) {
		t.Error("the default logger rejects LevelDebug")
	}
}

func TestPackageFuncs(t *testing.T) {
	saved := slog.Default()
	defer slog.SetDefault(saved)

	var h recHandler
	h.minLevel = slog.LevelDebug
	l := slog.New(&h)
	slog.SetDefault(&l)

	slog.Debug("d", slog.Int("n", 0))
	slog.Info("i", slog.Int("n", 1))
	slog.Warn("w", slog.Int("n", 2))
	slog.Error("e", slog.Int("n", 3))
	slog.Log(slog.LevelWarn, "l", slog.Int("n", 4))

	if h.calls != 5 {
		t.Errorf("Handle() got %d calls, want 5", h.calls)
		return
	}
	if h.gotLevel != slog.LevelWarn {
		t.Errorf("the last record has the level %d, want %d",
			int(h.gotLevel), int(slog.LevelWarn))
	}
	if h.gotMsg != "l" {
		t.Errorf("the last record has the message %s, want l", h.gotMsg)
	}
	if h.nAttrs != 1 || h.gotAttrs[0].Value.Int() != 4 {
		t.Error("the last record has the wrong attrs")
	}
}

func TestPackageFuncsFiltered(t *testing.T) {
	saved := slog.Default()
	defer slog.SetDefault(saved)

	var h recHandler
	h.minLevel = slog.LevelWarn
	l := slog.New(&h)
	slog.SetDefault(&l)

	slog.Debug("d")
	slog.Info("i")
	if h.calls != 0 {
		t.Errorf("Handle() got %d calls, want 0", h.calls)
	}
	slog.Warn("w")
	slog.Error("e")
	if h.calls != 2 {
		t.Errorf("Handle() got %d calls, want 2", h.calls)
	}
}

// getDefault returns the process-global default logger.
func getDefault(arg any) any {
	_ = arg
	return slog.Default()
}

func TestDefaultConcurrentInit(t *testing.T) {
	// The default logger lazy init must run exactly once,
	// so every thread must observe the same non-nil *Logger.
	const n = 8
	var threads [n]conc.Thread
	for i := range n {
		threads[i] = conc.Go(getDefault, nil)
	}

	var first *slog.Logger
	for i := range n {
		got := threads[i].Wait().(*slog.Logger)
		if got == nil {
			t.Fatal("default logger is nil")
			return
		}
		if i == 0 {
			first = got
		} else if got != first {
			t.Error("default logger differs between threads")
		}
	}
}
