package time_bench

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

//so:volatile
var sinkInt int64

//so:volatile
var sinkStr string

//so:volatile
var sinkTime time.Time

func BenchmarkDate_So(b *testing.B) {
	t := time.Now()
	for b.Loop() {
		date := t.Date(time.UTC)
		sinkInt = int64(date.Year)
	}
}

func BenchmarkFormat_So(b *testing.B) {
	buf := make([]byte, 64)
	t := time.Unix(1265346057, 0)
	for b.Loop() {
		sinkStr = t.Format(buf, time.RFC3339, time.UTC)
	}
}

func BenchmarkFormatCustom_So(b *testing.B) {
	buf := make([]byte, 64)
	t := time.Unix(1265346057, 0)
	for b.Loop() {
		sinkStr = t.Format(buf, "%d.%m.%Y", time.UTC)
	}
}

func BenchmarkISOWeek_So(b *testing.B) {
	t := time.Now()
	for b.Loop() {
		_, week := t.ISOWeek()
		sinkInt = int64(week)
	}
}

func BenchmarkNow_So(b *testing.B) {
	for b.Loop() {
		sinkTime = time.Now()
	}
}

func BenchmarkParse_So(b *testing.B) {
	str := "2011-11-18T15:56:35Z"
	_, err := time.Parse(time.RFC3339, str, time.UTC)
	if err != nil {
		panic(err)
	}
	for b.Loop() {
		sinkTime, _ = time.Parse(time.RFC3339, str, time.UTC)
	}
}

func BenchmarkParseCustom_So(b *testing.B) {
	str := "15.03.2024"
	_, err := time.Parse("%d.%m.%Y", str, time.UTC)
	if err != nil {
		panic(err)
	}
	for b.Loop() {
		sinkTime, _ = time.Parse("%d.%m.%Y", str, time.UTC)
	}
}

func BenchmarkSince_So(b *testing.B) {
	start := time.Now()
	for b.Loop() {
		sinkInt = int64(time.Since(start))
	}
}

func BenchmarkUnixNano_So(b *testing.B) {
	for b.Loop() {
		sinkInt = time.Now().UnixNano()
	}
}

func BenchmarkUntil_So(b *testing.B) {
	end := time.Now().Add(1 * time.Hour)
	for b.Loop() {
		sinkInt = int64(time.Until(end))
	}
}
