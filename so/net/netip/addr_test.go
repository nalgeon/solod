// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import (
	gonetip "net/netip"
	"strings"
	"testing"
)

func TestIPv6Accessor(t *testing.T) {
	var a [16]byte
	for i := range a {
		a[i] = uint8(i) + 1
	}
	ip := AddrFrom16(a)
	for i := range a {
		if got, want := ip.v6(uint8(i)), uint8(i)+1; got != want {
			t.Errorf("v6(%v) = %v; want %v", i, got, want)
		}
	}
}

func TestAs4Panic(t *testing.T) {
	as4 := func(ip Addr) (gotPanic bool) {
		defer func() {
			if recover() != nil {
				gotPanic = true
			}
		}()
		var a4 [4]byte
		ip.As4(a4)
		return
	}

	tests := []struct {
		ip        Addr
		wantPanic bool
	}{
		{ip: MustParseAddr("1.2.3.4")},
		{ip: MustParseAddr("::ffff:1.2.3.4")},
		{ip: MustParseAddr("0.0.0.0")},
		{ip: Addr{}, wantPanic: true},
		{ip: MustParseAddr("::1"), wantPanic: true},
	}
	for _, tt := range tests {
		if got := as4(tt.ip); got != tt.wantPanic {
			var buf [MaxAddrLen]byte
			t.Errorf("As4(%v) panic = %v; want %v", tt.ip.String(buf[:]), got, tt.wantPanic)
		}
	}
}

func TestMustParseAddrPanic(t *testing.T) {
	mustParse := func(s string) (gotPanic bool) {
		defer func() {
			if recover() != nil {
				gotPanic = true
			}
		}()
		MustParseAddr(s)
		return
	}

	tests := []struct {
		in        string
		wantPanic bool
	}{
		{in: "1.2.3.4"},
		{in: "::1"},
		{in: "", wantPanic: true},
		{in: "1.2.3.4.5", wantPanic: true},
		{in: "::gggg", wantPanic: true},
	}
	for _, tt := range tests {
		if got := mustParse(tt.in); got != tt.wantPanic {
			t.Errorf("MustParseAddr(%q) panic = %v; want %v", tt.in, got, tt.wantPanic)
		}
	}
}

// fuzzAddrs are the seed inputs of the address fuzzers.
var fuzzAddrs = []string{
	"",
	" ",
	"1",
	"1.2.3",
	"1.2.3.4",
	"0.0.0.0",
	"255.255.255.255",
	"1.2.3.400",
	"01.2.3.4",
	"1.2.3.4.5",
	"::",
	"::1",
	"::ffff:1.2.3.4",
	"::ffff:c000:0280",
	"2001:db8::1",
	"fe80::1",
	"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	"::ffff:0:0",
	"1:2:3:4:5:6:7:8",
	"1:2:3:4:5:6:7:8:9",
	"1::2::3",
	"::gggg",
	":::",
	"[::1]",
	"1.2.3.4:80",
	"[::1]:80",
	"1.2.3.0/24",
	"::/0",
	"2001:db8::/129",
	"1.1.1.0/033",
}

// FuzzParseAddr compares ParseAddr against the Go original.
// It skips an input with a zone: So stores a scope id, Go stores the zone
// text, so the two results cannot be compared.
func FuzzParseAddr(f *testing.F) {
	for _, s := range fuzzAddrs {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if strings.ContainsRune(s, '%') {
			return
		}

		soIP, soErr := ParseAddr(s)
		goIP, goErr := gonetip.ParseAddr(s)
		if (soErr == nil) != (goErr == nil) {
			t.Fatalf("ParseAddr(%q) err = %v; Go err = %v", s, soErr, goErr)
		}
		if soErr != nil {
			return
		}

		var a16 [16]byte
		if got, want := soIP.As16(a16), goIP.As16(); got != want {
			t.Errorf("ParseAddr(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.Is4(), goIP.Is4(); got != want {
			t.Errorf("ParseAddr(%q).Is4() = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.Is4In6(), goIP.Is4In6(); got != want {
			t.Errorf("ParseAddr(%q).Is4In6() = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.BitLen(), goIP.BitLen(); got != want {
			t.Errorf("ParseAddr(%q).BitLen() = %v; Go = %v", s, got, want)
		}

		var buf [MaxAddrLen]byte
		if got, want := soIP.String(buf[:]), goIP.String(); got != want {
			t.Errorf("ParseAddr(%q).String() = %q; Go = %q", s, got, want)
		}
	})
}

// FuzzAddrProperties compares the property methods against the Go original.
func FuzzAddrProperties(f *testing.F) {
	for _, s := range fuzzAddrs {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if strings.ContainsRune(s, '%') {
			return
		}

		soIP, soErr := ParseAddr(s)
		goIP, goErr := gonetip.ParseAddr(s)
		if soErr != nil || goErr != nil {
			return
		}

		if got, want := soIP.IsGlobalUnicast(), goIP.IsGlobalUnicast(); got != want {
			t.Errorf("IsGlobalUnicast(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsInterfaceLocalMulticast(), goIP.IsInterfaceLocalMulticast(); got != want {
			t.Errorf("IsInterfaceLocalMulticast(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsLinkLocalMulticast(), goIP.IsLinkLocalMulticast(); got != want {
			t.Errorf("IsLinkLocalMulticast(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsLinkLocalUnicast(), goIP.IsLinkLocalUnicast(); got != want {
			t.Errorf("IsLinkLocalUnicast(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsLoopback(), goIP.IsLoopback(); got != want {
			t.Errorf("IsLoopback(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsMulticast(), goIP.IsMulticast(); got != want {
			t.Errorf("IsMulticast(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsPrivate(), goIP.IsPrivate(); got != want {
			t.Errorf("IsPrivate(%q) = %v; Go = %v", s, got, want)
		}
		if got, want := soIP.IsUnspecified(), goIP.IsUnspecified(); got != want {
			t.Errorf("IsUnspecified(%q) = %v; Go = %v", s, got, want)
		}
	})
}
