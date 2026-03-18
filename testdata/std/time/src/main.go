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
		if t.Location().String() != "UTC" {
			panic("unexpected Time.Location")
		}
		println(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location().String())
	}
}
