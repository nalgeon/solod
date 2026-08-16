// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import (
	gonetip "net/netip"
	"strings"
	"testing"
)

func TestMustParseAddrPortPanic(t *testing.T) {
	mustParse := func(s string) (gotPanic bool) {
		defer func() {
			if recover() != nil {
				gotPanic = true
			}
		}()
		MustParseAddrPort(s)
		return
	}

	tests := []struct {
		in        string
		wantPanic bool
	}{
		{in: "1.2.3.4:80"},
		{in: "[::1]:80"},
		{in: "", wantPanic: true},
		{in: "1.2.3.4", wantPanic: true},
		{in: "1.2.3.4:65536", wantPanic: true},
	}
	for _, tt := range tests {
		if got := mustParse(tt.in); got != tt.wantPanic {
			t.Errorf("MustParseAddrPort(%q) panic = %v; want %v", tt.in, got, tt.wantPanic)
		}
	}
}

// fuzzAddrPorts are the seed inputs of the address-port fuzzer.
var fuzzAddrPorts = []string{
	"",
	":",
	":80",
	"1.2.3.4",
	"1.2.3.4:",
	"1.2.3.4:0",
	"1.2.3.4:80",
	"1.2.3.4:65535",
	"1.2.3.4:65536",
	"1.2.3.4:+80",
	"1.2.3.4:-80",
	"1.2.3.400:80",
	"[1.2.3.4]:80",
	"[::1]:80",
	"[::1]:",
	"[::1]",
	"[::1:80",
	"::1:80",
	"[]:80",
	"[::ffff:1.2.3.4]:80",
	"[::ffff:c000:0280]:65535",
	"[bad]:80",
	"[::gggg]:80",
}

func FuzzParseAddrPort(f *testing.F) {
	for _, s := range fuzzAddrPorts {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if strings.ContainsRune(s, '%') {
			return
		}

		soAP, soErr := ParseAddrPort(s)
		goAP, goErr := gonetip.ParseAddrPort(s)
		if (soErr == nil) != (goErr == nil) {
			t.Fatalf("ParseAddrPort(%q) err = %v; Go err = %v", s, soErr, goErr)
		}
		if soErr != nil {
			return
		}

		var a16 [16]byte
		if got, want := soAP.Addr().As16(a16), goAP.Addr().As16(); got != want {
			t.Errorf("ParseAddrPort(%q).Addr() = %v; Go = %v", s, got, want)
		}
		if got, want := soAP.Port(), goAP.Port(); got != want {
			t.Errorf("ParseAddrPort(%q).Port() = %v; Go = %v", s, got, want)
		}

		var buf [MaxAddrPortLen]byte
		if got, want := soAP.String(buf[:]), goAP.String(); got != want {
			t.Errorf("ParseAddrPort(%q).String() = %q; Go = %q", s, got, want)
		}
	})
}
