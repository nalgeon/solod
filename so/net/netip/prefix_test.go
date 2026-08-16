// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import (
	gonetip "net/netip"
	"strings"
	"testing"
)

func TestMustParsePrefixPanic(t *testing.T) {
	mustParse := func(s string) (gotPanic bool) {
		defer func() {
			if recover() != nil {
				gotPanic = true
			}
		}()
		MustParsePrefix(s)
		return
	}

	tests := []struct {
		in        string
		wantPanic bool
	}{
		{in: "1.2.3.0/24"},
		{in: "::/0"},
		{in: "", wantPanic: true},
		{in: "1.2.3.0", wantPanic: true},
		{in: "1.2.3.0/33", wantPanic: true},
	}
	for _, tt := range tests {
		if got := mustParse(tt.in); got != tt.wantPanic {
			t.Errorf("MustParsePrefix(%q) panic = %v; want %v", tt.in, got, tt.wantPanic)
		}
	}
}

// fuzzPrefixes are the seed inputs of the prefix fuzzers.
var fuzzPrefixes = []string{
	"",
	"/",
	"1.2.3.0",
	"1.2.3.0/",
	"1.2.3.0/0",
	"1.2.3.0/24",
	"1.2.3.4/24",
	"1.2.3.0/32",
	"1.2.3.0/33",
	"1.2.3.0/033",
	"1.2.3.0/+24",
	"1.2.3.0/-1",
	"0.0.0.0/0",
	"255.255.255.255/32",
	"100.64.0.0/10",
	"::/0",
	"::1/128",
	"2001:db8::/32",
	"2001:db8::1/32",
	"2000::/3",
	"2001:db8::/129",
	"::ffff:1.2.3.4/120",
	"::ffff:1.2.3.4/24",
	"1.2.3.0/24/24",
}

func FuzzParsePrefix(f *testing.F) {
	for _, s := range fuzzPrefixes {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if strings.ContainsRune(s, '%') {
			return
		}

		soP, soErr := ParsePrefix(s)
		goP, goErr := gonetip.ParsePrefix(s)
		if (soErr == nil) != (goErr == nil) {
			t.Fatalf("ParsePrefix(%q) err = %v; Go err = %v", s, soErr, goErr)
		}
		if soErr != nil {
			return
		}

		var a16 [16]byte
		if got, want := soP.Addr().As16(a16), goP.Addr().As16(); got != want {
			t.Errorf("ParsePrefix(%q).Addr() = %v; Go = %v", s, got, want)
		}
		if got, want := soP.Bits(), goP.Bits(); got != want {
			t.Errorf("ParsePrefix(%q).Bits() = %v; Go = %v", s, got, want)
		}
		if got, want := soP.IsSingleIP(), goP.IsSingleIP(); got != want {
			t.Errorf("ParsePrefix(%q).IsSingleIP() = %v; Go = %v", s, got, want)
		}

		var buf [MaxPrefixLen]byte
		if got, want := soP.String(buf[:]), goP.String(); got != want {
			t.Errorf("ParsePrefix(%q).String() = %q; Go = %q", s, got, want)
		}
		if got, want := soP.Masked().String(buf[:]), goP.Masked().String(); got != want {
			t.Errorf("ParsePrefix(%q).Masked() = %q; Go = %q", s, got, want)
		}
	})
}

func FuzzPrefixContains(f *testing.F) {
	for _, s := range fuzzPrefixes {
		f.Add(s, "1.2.3.4/24")
		f.Add(s, "::1/128")
	}
	f.Fuzz(func(t *testing.T, sa, sb string) {
		if strings.ContainsRune(sa, '%') || strings.ContainsRune(sb, '%') {
			return
		}

		soA, err := ParsePrefix(sa)
		if err != nil {
			return
		}
		soB, err := ParsePrefix(sb)
		if err != nil {
			return
		}
		goA, err := gonetip.ParsePrefix(sa)
		if err != nil {
			return
		}
		goB, err := gonetip.ParsePrefix(sb)
		if err != nil {
			return
		}

		if got, want := soA.Contains(soB.Addr()), goA.Contains(goB.Addr()); got != want {
			t.Errorf("Contains(%q, %q) = %v; Go = %v", sa, sb, got, want)
		}
		if got, want := soA.Overlaps(soB), goA.Overlaps(goB); got != want {
			t.Errorf("Overlaps(%q, %q) = %v; Go = %v", sa, sb, got, want)
		}
	})
}
