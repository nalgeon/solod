package time_bench

import (
	"testing"
	"time"
)

var (
	goInt  int64
	goStr  string
	goTime time.Time
)

func BenchmarkDate_Go(b *testing.B) {
	t := time.Now()
	for b.Loop() {
		year, _, _ := t.Date()
		goInt = int64(year)
	}
}

func BenchmarkFormat_Go(b *testing.B) {
	t := time.Unix(1265346057, 0)
	for b.Loop() {
		goStr = t.Format(time.RFC3339)
	}
}

func BenchmarkFormatCustom_Go(b *testing.B) {
	t := time.Unix(1265346057, 0)
	for b.Loop() {
		goStr = t.Format("02.01.2006")
	}
}

func BenchmarkISOWeek_Go(b *testing.B) {
	t := time.Now()
	for b.Loop() {
		_, week := t.ISOWeek()
		goInt = int64(week)
	}
}

func BenchmarkNow_Go(b *testing.B) {
	for b.Loop() {
		goTime = time.Now()
	}
}

func BenchmarkParse_Go(b *testing.B) {
	str := "2011-11-18T15:56:35Z"
	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	for b.Loop() {
		goTime, _ = time.Parse(time.RFC3339, str)
	}
}

func BenchmarkParseCustom_Go(b *testing.B) {
	str := "15.03.2024"
	_, err := time.Parse("02.01.2006", str)
	if err != nil {
		panic(err)
	}
	for b.Loop() {
		goTime, _ = time.Parse("02.01.2006", str)
	}
}

func BenchmarkSince_Go(b *testing.B) {
	start := time.Now()
	for b.Loop() {
		goInt = int64(time.Since(start))
	}
}

func BenchmarkUnixNano_Go(b *testing.B) {
	for b.Loop() {
		goInt = time.Now().UnixNano()
	}
}

func BenchmarkUntil_Go(b *testing.B) {
	end := time.Now().Add(1 * time.Hour)
	for b.Loop() {
		goInt = int64(time.Until(end))
	}
}
