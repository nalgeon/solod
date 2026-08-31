package net

import (
	stdnet "net"
	"testing"
)

// The address texts the fuzzers start from.
var hostPortSeeds = []string{
	"",
	":",
	":80",
	"localhost:http",
	"localhost%lo0:80",
	"127.0.0.1:80",
	"127.0.0.1:",
	"[::1]:80",
	"[::1%lo0]:http",
	"[]:",
	"golang.org",
	"::1",
	"fe80::1%lo0:80",
	"[foo:bar]baz",
	"[foo]:[bar]:baz",
	"foo[bar]:baz",
	"foo]bar:baz",
	"[foo:80",
}

// sameErr reports whether the error of this package and the error of Go's net
// package name the same defect. Solod merges Go's two bracket errors into
// ErrUnexpectedBracket, so that sentinel accepts either message.
func sameErr(soErr, goErr error) bool {
	if soErr == nil || goErr == nil {
		return soErr == nil && goErr == nil
	}
	ae, ok := goErr.(*stdnet.AddrError)
	if !ok {
		return false
	}
	switch soErr {
	case ErrMissingPort:
		return ae.Err == "missing port in address"
	case ErrTooManyColons:
		return ae.Err == "too many colons in address"
	case ErrMissingBracket:
		return ae.Err == "missing ']' in address"
	case ErrUnexpectedBracket:
		return ae.Err == "unexpected '[' in address" || ae.Err == "unexpected ']' in address"
	}
	return false
}

func FuzzSplitHostPort(f *testing.F) {
	for _, s := range hostPortSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, hostport string) {
		hp, err := SplitHostPort(hostport)
		goHost, goPort, goErr := stdnet.SplitHostPort(hostport)
		if !sameErr(err, goErr) {
			t.Fatalf("SplitHostPort(%q) error = %v; Go = %v", hostport, err, goErr)
		}
		if hp.Host != goHost || hp.Port != goPort {
			t.Fatalf("SplitHostPort(%q) = %q, %q; Go = %q, %q",
				hostport, hp.Host, hp.Port, goHost, goPort)
		}
	})
}

func FuzzJoinHostPort(f *testing.F) {
	for _, s := range hostPortSeeds {
		host, port, err := stdnet.SplitHostPort(s)
		if err == nil {
			f.Add(host, port)
		}
	}
	f.Fuzz(func(t *testing.T, host, port string) {
		buf := make([]byte, len(host)+len(port)+4)
		got := JoinHostPort(buf, host, port)
		if want := stdnet.JoinHostPort(host, port); got != want {
			t.Fatalf("JoinHostPort(%q, %q) = %q; Go = %q", host, port, got, want)
		}
		if len(got) > len(buf) || string(buf[:len(got)]) != got {
			t.Fatalf("JoinHostPort(%q, %q) did not build into buf", host, port)
		}
	})
}
