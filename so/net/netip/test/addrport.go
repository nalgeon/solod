// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

// A portCase is one ParseAddrPort case with a valid pair.
type portCase struct {
	in   string // the pair text
	addr string // the address part
	port uint16 // the port part
	str  string // the String result, "" if String gives in back
}

var portCases = []portCase{
	{in: "1.2.3.4:1234", addr: "1.2.3.4", port: 1234},
	{in: "127.0.0.1:80", addr: "127.0.0.1", port: 80},
	{in: "1.2.3.4:0", addr: "1.2.3.4", port: 0},
	{in: "255.255.255.255:65535", addr: "255.255.255.255", port: 65535},
	{in: "[::1]:1234", addr: "::1", port: 1234},
	{in: "[::]:8080", addr: "::", port: 8080},
	{in: "[0000::0]:8080", addr: "::", port: 8080, str: "[::]:8080"},
	{in: "[FFFF::1]:8080", addr: "ffff::1", port: 8080, str: "[ffff::1]:8080"},
	{
		in:   "[fd7a:115c:a1e0:ab12:4843:cd96:626b:430b]:80",
		addr: "fd7a:115c:a1e0:ab12:4843:cd96:626b:430b",
		port: 80,
	},
	// A 4-in-6 address keeps its dotted decimal form.
	{in: "[::ffff:1.2.3.4]:80", addr: "::ffff:1.2.3.4", port: 80},
	{
		in:   "[::ffff:c000:0280]:65535",
		addr: "::ffff:192.0.2.128",
		port: 65535,
		str:  "[::ffff:192.0.2.128]:65535",
	},
	// A zone is allowed.
	{in: "[::1%10]:80", addr: "::1%10", port: 80},
	{in: "[::ffff:1.2.3.4%10]:1", addr: "::ffff:1.2.3.4%10", port: 1},
}

// A badPortCase is one parse case with an invalid pair.
type badPortCase struct {
	in  string // the pair text
	err int    // the expected error
}

var badPortCases = []badPortCase{
	// No colon at all.
	{"", errIPPort},
	{"1.2.3.4", errIPPort},
	// No address.
	{":0", errIP},
	{"[]:80", errIP},
	// No port.
	{"1.2.3.4:", errPort},
	{"[::1]:", errPort},
	// A port out of range.
	{"1.1.1.1:123456", errPort},
	{"1.1.1.1:-123", errPort},
	{"1.1.1.1:65536", errPort},
	{"1.1.1.1:+80", errPort},
	{"1.1.1.1:8o", errPort},
	// A missing bracket.
	{"[::1:80", errIP},
	{"[::1]80", errIP},
	// An IPv4 address in brackets.
	{"[1.2.3.4]:1234", errIPPort},
	// An IPv6 address without brackets.
	{"fe80::1:1234", errIPPort},
	{"::1:80", errIPPort},
	// A bad address.
	{"[bad]:80", errIP},
	{"1.2.3.400:80", errIPv4},
	{"[::gggg]:80", errIPv6},
}

func TestParseAddrPort(t *testing.T) {
	for _, tc := range portCases {
		ap, err := netip.ParseAddrPort(tc.in)
		if err != nil {
			t.Errorf("ParseAddrPort(%s) failed: %s", tc.in, err.Error())
			continue
		}

		addr := ap.Addr()
		if !addr.Equal(netip.MustParseAddr(tc.addr)) {
			var buf [bufLen]byte
			t.Errorf("ParseAddrPort(%s).Addr() = %s, want %s",
				tc.in, addr.String(buf[:]), tc.addr)
		}
		if got := ap.Port(); got != tc.port {
			t.Errorf("ParseAddrPort(%s).Port() = %d, want %d", tc.in, got, tc.port)
		}
		if !ap.IsValid() {
			t.Errorf("ParseAddrPort(%s).IsValid() = false, want true", tc.in)
		}

		want := tc.str
		if want == "" {
			want = tc.in
		}
		var buf [bufLen]byte
		s := ap.String(buf[:])
		if s != want {
			t.Errorf("ParseAddrPort(%s).String() = %s, want %s", tc.in, s, want)
		}

		// ParseAddrPort(p.String()) is the identity function.
		ap2, err := netip.ParseAddrPort(s)
		if err != nil {
			t.Errorf("ParseAddrPort(%s) failed: %s", s, err.Error())
			continue
		}
		if ap.Compare(ap2) != 0 {
			t.Errorf("ParseAddrPort(%s) != ParseAddrPort(ParseAddrPort(%s).String())", tc.in, tc.in)
		}
	}
}

func TestParseAddrPortBad(t *testing.T) {
	for _, tc := range badPortCases {
		ap, err := netip.ParseAddrPort(tc.in)
		if err == nil {
			var buf [bufLen]byte
			t.Errorf("ParseAddrPort(%s) = %s, want an error", tc.in, ap.String(buf[:]))
			continue
		}
		if want := wantErr(tc.err); want != nil && err != want {
			t.Errorf("ParseAddrPort(%s) = %s, want %s", tc.in, err.Error(), want.Error())
		}
		if ap.IsValid() {
			t.Errorf("ParseAddrPort(%s) failed, but the pair is valid", tc.in)
		}
	}
}

func TestAddrPortFrom(t *testing.T) {
	ip := netip.MustParseAddr("1.2.3.4")
	ap := netip.AddrPortFrom(ip, 8080)
	if !ap.Addr().Equal(ip) {
		t.Error("AddrPortFrom().Addr() is not the address")
	}
	if ap.Port() != 8080 {
		t.Errorf("AddrPortFrom().Port() = %d, want 8080", ap.Port())
	}
	if !ap.IsValid() {
		t.Error("AddrPortFrom().IsValid() = false, want true")
	}

	// Every port is valid, zero included.
	if !netip.AddrPortFrom(ip, 0).IsValid() {
		t.Error("AddrPortFrom(ip, 0).IsValid() = false, want true")
	}

	// The zero address gives an invalid pair.
	var zero netip.Addr
	if netip.AddrPortFrom(zero, 80).IsValid() {
		t.Error("AddrPortFrom(Addr{}, 80).IsValid() = true, want false")
	}
}

func TestAddrPortString(t *testing.T) {
	var buf [bufLen]byte

	var zero netip.AddrPort
	if got := zero.String(buf[:]); got != "invalid AddrPort" {
		t.Errorf("AddrPort{}.String() = %s, want invalid AddrPort", got)
	}
	var zeroAddr netip.Addr
	if got := netip.AddrPortFrom(zeroAddr, 80).String(buf[:]); got != "invalid AddrPort" {
		t.Errorf("AddrPortFrom(Addr{}, 80).String() = %s, want invalid AddrPort", got)
	}
	if zero.IsValid() {
		t.Error("AddrPort{}.IsValid() = true, want false")
	}
}

func TestAddrPortAppendText(t *testing.T) {
	for _, tc := range portCases {
		ap := netip.MustParseAddrPort(tc.in)
		var buf [bufLen]byte
		b, err := ap.AppendText(buf[:0])
		if err != nil {
			t.Errorf("AppendText(%s) failed: %s", tc.in, err.Error())
			continue
		}
		want := tc.str
		if want == "" {
			want = tc.in
		}
		if got := string(b); got != want {
			t.Errorf("AppendText(%s) = %s, want %s", tc.in, got, want)
		}
	}

	// The zero pair appends nothing.
	var zero netip.AddrPort
	var buf [bufLen]byte
	b, err := zero.AppendText(buf[:0])
	if err != nil {
		t.Errorf("AddrPort{}.AppendText() failed: %s", err.Error())
		return
	}
	if len(b) != 0 {
		t.Errorf("AddrPort{}.AppendText() = %s, want an empty result", string(b))
	}
}

// A portCompareCase is one case of the AddrPort Compare method.
// An empty address text means the zero pair.
type portCompareCase struct {
	a    string // the first pair
	b    string // the second pair
	want int    // the Compare result
}

var portCompareCases = []portCompareCase{
	{"", "", 0},
	{"", "1.2.3.4:80", -1},
	{"1.2.3.4:80", "", 1},

	{"1.2.3.4:80", "1.2.3.4:80", 0},
	{"[::1]:80", "[::1]:80", 0},

	// The address sorts before the port.
	{"1.2.3.4:80", "2.3.4.5:22", -1},
	{"[::1]:80", "[::2]:22", -1},
	{"1.2.3.4:80", "1.2.3.4:443", -1},
	{"[::1]:80", "[::1]:443", -1},

	// An IPv4 address sorts before an IPv6 address.
	{"1.2.3.4:80", "[0102:0304::0]:80", -1},

	// A zone sorts after the same address without one.
	{"[::1]:80", "[::1%1]:80", -1},
}

func TestAddrPortCompare(t *testing.T) {
	for _, tc := range portCompareCases {
		var a, b netip.AddrPort
		if tc.a != "" {
			a = netip.MustParseAddrPort(tc.a)
		}
		if tc.b != "" {
			b = netip.MustParseAddrPort(tc.b)
		}

		if got := a.Compare(b); got != tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
		if back := b.Compare(a); back != -tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.b, tc.a, back, -tc.want)
		}
	}
}

func TestAddrPortSort(t *testing.T) {
	var zero netip.AddrPort
	values := []netip.AddrPort{
		netip.MustParseAddrPort("[::1]:80"),
		netip.MustParseAddrPort("[::2]:80"),
		zero,
		netip.MustParseAddrPort("1.2.3.4:443"),
		netip.MustParseAddrPort("8.8.8.8:8080"),
		netip.MustParseAddrPort("[::1%42]:1024"),
	}
	want := []string{
		"invalid AddrPort", "1.2.3.4:443", "8.8.8.8:8080",
		"[::1]:80", "[::1%42]:1024", "[::2]:80",
	}

	slices.SortFunc(values, compareAddrPorts)
	for i, v := range values {
		var buf [bufLen]byte
		if got := v.String(buf[:]); got != want[i] {
			t.Errorf("sorted[%d] = %s, want %s", i, got, want[i])
		}
	}
}

func TestAddrPortMaxLen(t *testing.T) {
	// Write the longest pair text into a buffer of the documented size.
	ip := netip.MustParseAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff").WithZone("4294967295")
	ap := netip.AddrPortFrom(ip, 65535)
	want := "[ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%4294967295]:65535"

	var buf [netip.MaxAddrPortLen]byte
	if got := ap.String(buf[:]); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}

	var bufAppend [netip.MaxAddrPortLen]byte
	b, err := ap.AppendText(bufAppend[:0])
	if err != nil {
		t.Errorf("AppendText() failed: %s", err.Error())
		return
	}
	if got := string(b); got != want {
		t.Errorf("AppendText() = %s, want %s", got, want)
	}
}
