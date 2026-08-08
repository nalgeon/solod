package slog_bench

import (
	"solod.dev/so/log/slog"
	"solod.dev/so/testing"
)

//so:volatile
var sink int

type soHandler struct{}

func (h *soHandler) Enabled(level slog.Level) bool {
	return true
}
func (h *soHandler) Handle(r slog.Record) error {
	sink += len(r.Message)
	return nil
}

func BenchmarkNoAttr_So(b *testing.B) {
	h := &soHandler{}
	l := slog.New(h)
	for b.Loop() {
		l.Log(slog.LevelInfo, "msg")
	}
}

func BenchmarkAttr_So(b *testing.B) {
	h := &soHandler{}
	l := slog.New(h)
	for b.Loop() {
		l.Log(slog.LevelInfo, "msg",
			slog.Int("a", 1), slog.String("b", "two"), slog.Bool("c", true))
	}
}
