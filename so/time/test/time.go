package time_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

func TestDate(t *testing.T) {
	tm := time.Date(2021, time.May, 10, 12, 33, 44, 777888999, time.UTC)
	if tm.Year() != 2021 {
		t.Error("unexpected Time.Year")
	}
	if tm.Month() != time.May {
		t.Error("unexpected Time.Month")
	}
	if tm.Day() != 10 {
		t.Error("unexpected Time.Day")
	}
	if tm.Hour() != 12 {
		t.Error("unexpected Time.Hour")
	}
	if tm.Minute() != 33 {
		t.Error("unexpected Time.Minute")
	}
	if tm.Second() != 44 {
		t.Error("unexpected Time.Second")
	}
	if tm.Nanosecond() != 777888999 {
		t.Error("unexpected Time.Nanosecond")
	}
}

func TestNow(t *testing.T) {
	tm := time.Now()
	if tm.IsZero() {
		t.Error("unexpected Time.IsZero")
	}
}

func TestSleep(t *testing.T) {
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		t.Error("Sleep returned before the duration elapsed")
	}
	// No upper bound is asserted: Sleep only guarantees a lower bound, and a
	// loaded machine (e.g. a CI runner) may schedule the thread much later.
}

func TestSleepNonPositive(t *testing.T) {
	start := time.Now()
	// Returns immediately without blocking.
	time.Sleep(0)
	time.Sleep(-1 * time.Second)
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Error("Sleep should return immediately")
	}
}
