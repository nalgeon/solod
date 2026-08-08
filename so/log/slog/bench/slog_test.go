package slog_bench

import (
	"context"
	"log/slog"
	"testing"
)

type goHandler struct{}

func (h *goHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}
func (h *goHandler) Handle(_ context.Context, r slog.Record) error {
	sink += len(r.Message)
	return nil
}
func (h *goHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return new(goHandler)
}
func (h *goHandler) WithGroup(name string) slog.Handler {
	return new(goHandler)
}

func BenchmarkNoAttr_Go(b *testing.B) {
	ctx := context.Background()
	h := &goHandler{}
	l := slog.New(h)
	b.ReportAllocs()
	for b.Loop() {
		l.Log(ctx, slog.LevelInfo, "msg")
	}
}

func BenchmarkAttr_Go(b *testing.B) {
	ctx := context.Background()
	h := &goHandler{}
	l := slog.New(h)
	b.ReportAllocs()
	for b.Loop() {
		l.Log(ctx, slog.LevelInfo, "msg",
			slog.Int("a", 1), slog.String("b", "two"), slog.Bool("c", true))
	}
}
