package main

import (
	"solod.dev/so/time"
	"solod.dev/so/testing"
)

func TestFormat_RFC3339(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, time.RFC3339Len)
	s := tm.Format(buf, time.RFC3339, time.UTC)
	if s != "2024-03-15T14:30:45Z" {
		t.Error("unexpected RFC3339 format")
	}
}

func TestFormat_RFC3339Nano(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 123456789, time.UTC)
	buf := make([]byte, time.RFC3339NanoLen)
	s := tm.Format(buf, time.RFC3339Nano, time.UTC)
	if s != "2024-03-15T14:30:45.123456789Z" {
		t.Error("unexpected RFC3339Nano format")
	}
}

func TestFormat_DateTime(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, time.DateTimeLen)
	s := tm.Format(buf, time.DateTime, time.UTC)
	if s != "2024-03-15 14:30:45" {
		t.Error("unexpected DateTime format")
	}
}

func TestFormat_DateOnly(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, time.DateOnlyLen)
	s := tm.Format(buf, time.DateOnly, time.UTC)
	if s != "2024-03-15" {
		t.Error("unexpected DateOnly format")
	}
}

func TestFormat_TimeOnly(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, time.TimeOnlyLen)
	s := tm.Format(buf, time.TimeOnly, time.UTC)
	if s != "14:30:45" {
		t.Error("unexpected TimeOnly format")
	}
}

func TestFormat_Custom(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, len("15.03.2024")+1)
	s := tm.Format(buf, "%d.%m.%Y", time.UTC)
	if s != "15.03.2024" {
		t.Error("unexpected custom format")
	}
}

func TestString(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	buf := make([]byte, time.RFC3339Len)
	s := tm.String(buf)
	if s != "2024-03-15T14:30:45Z" {
		t.Error("unexpected String format")
	}
}

func TestFormat_EmptyBuf(t *testing.T) {
	tm := time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC)
	var buf []byte
	if tm.Format(buf, time.RFC3339, time.UTC) != "" {
		t.Error("Format with nil buf: want empty")
	}
	if tm.Format(buf, "%Y", time.UTC) != "" {
		t.Error("Format with nil buf and strftime layout: want empty")
	}
}
