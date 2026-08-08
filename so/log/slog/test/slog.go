package slog_test

import (
	"solod.dev/so/conc"
	"solod.dev/so/log/slog"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestEnabled(t *testing.T) {
	var sb strings.Builder
	defer sb.Free()

	h := slog.NewTextHandler(&sb, slog.LevelInfo)
	l := slog.New(&h)

	if l.Enabled(slog.LevelDebug) {
		t.Error("debug should not be enabled")
	}
	if !l.Enabled(slog.LevelInfo) {
		t.Error("info should be enabled")
	}
}

func TestText(t *testing.T) {
	var sb strings.Builder
	defer sb.Free()

	h := slog.NewTextHandler(&sb, slog.LevelInfo)
	l := slog.New(&h)

	l.Info("hello world", slog.String("user", "john"), slog.Int("count", 42))
	l.Debug("hidden") // filtered out, below the handler level
	l.Warn("caution")
	l.Error("failure", slog.Float64("elapsed", 1.5), slog.Bool("retry", true))
	l.Info("test quoting", slog.String("msg", "hello world"))

	out := sb.String()
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
