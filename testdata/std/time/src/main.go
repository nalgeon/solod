package main

import (
	"github.com/nalgeon/solod/so/time"
)

func main() {
	{
		// time.Date and time.Time properties.
		t := time.Date(2021, time.May, 10, 12, 33, 44, 777888999, time.UTC)
		if t.Year() != 2021 {
			panic("unexpected Time.Year")
		}
		if t.Month() != time.May {
			panic("unexpected Time.Month")
		}
		if t.Day() != 10 {
			panic("unexpected Time.Day")
		}
		if t.Hour() != 12 {
			panic("unexpected Time.Hour")
		}
		if t.Minute() != 33 {
			panic("unexpected Time.Minute")
		}
		if t.Second() != 44 {
			panic("unexpected Time.Second")
		}
		if t.Nanosecond() != 777888999 {
			panic("unexpected Time.Nanosecond")
		}
		println(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
	}
	{
		// time.Time.Format and time.Time.String.
		t := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
		var buf [64]byte
		s := t.Format("%Y-%m-%d", time.UTC, buf[:])
		if s != "2024-03-15" {
			panic("unexpected Format")
		}
		s = t.String(buf[:])
		if s != "2024-03-15T14:30:45Z" {
			panic("unexpected String")
		}
	}
	{
		// time.Parse.
		t, err := time.Parse("%Y-%m-%d %H:%M:%S", "2024-03-15 14:30:45", time.UTC)
		if err != nil {
			panic("unexpected Parse error")
		}
		date := t.Date(time.UTC)
		clock := t.Clock(time.UTC)
		if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
			panic("unexpected Parse date")
		}
		if clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45 {
			panic("unexpected Parse clock")
		}
	}
	{
		// time.Parse error.
		_, err := time.Parse("%Y-%m-%d", "not-a-date", time.UTC)
		if err == nil {
			panic("expected Parse error")
		}
	}
	{
		// Time.Now.
		t := time.Now()
		if t.IsZero() {
			panic("unexpected Time.IsZero")
		}
		println("UTC:", t.String(make([]byte, 64)))
		utc5 := time.Offset(5 * 3600)
		println("UTC+5:", t.Format(time.RFC3339, utc5, make([]byte, 64)))
	}
}
