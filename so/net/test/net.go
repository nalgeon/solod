package net_test

import (
	"solod.dev/so/net"
	"solod.dev/so/testing"
)

func TestSplitHostPort(t *testing.T) {
	hp, err := net.SplitHostPort("127.0.0.1:8080")
	if err != nil || hp.Host != "127.0.0.1" || hp.Port != "8080" {
		t.Error("unexpected SplitHostPort result")
	}
}

func TestJoinHostPort(t *testing.T) {
	var buf [64]byte
	if net.JoinHostPort(buf[:], "::1", "80") != "[::1]:80" {
		t.Error("unexpected JoinHostPort result")
	}
}
