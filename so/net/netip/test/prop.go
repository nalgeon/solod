// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
)

// A propCase is one case of the address property methods.
// An empty address text means the zero address.
// A field that is not set means the property is false.
type propCase struct {
	in     string // the address text
	gu     bool   // IsGlobalUnicast
	ilm    bool   // IsInterfaceLocalMulticast
	llm    bool   // IsLinkLocalMulticast
	llu    bool   // IsLinkLocalUnicast
	lo     bool   // IsLoopback
	mc     bool   // IsMulticast
	priv   bool   // IsPrivate
	unspec bool   // IsUnspecified
}

var propCases = []propCase{
	// The zero address has no property.
	{in: ""},

	// Unicast.
	{in: "192.0.2.1", gu: true},
	{in: "::ffff:192.0.2.1", gu: true},
	{in: "2001:db8::1", gu: true},
	{in: "2001:db8::1%10", gu: true},
	// Not in 2000::/3, but still a global unicast address.
	{in: "4000::1", gu: true},
	{in: "0.0.0.1", gu: true},
	// A site-local address is not a link-local address.
	{in: "fec0::1", gu: true},

	// Multicast.
	{in: "224.0.0.1", llm: true, mc: true},
	{in: "::ffff:224.0.0.1", llm: true, mc: true},
	{in: "ff02::1", llm: true, mc: true},
	{in: "ff02::1%10", llm: true, mc: true},
	// The flags of the multicast address do not matter.
	{in: "ff12::1", llm: true, mc: true},
	{in: "ff11::1", ilm: true, mc: true},
	// Multicast, but neither interface-local nor link-local.
	{in: "224.0.1.1", mc: true},
	{in: "239.255.255.255", mc: true},
	{in: "ff05::1", mc: true},

	// Link-local unicast.
	{in: "169.254.0.1", llu: true},
	{in: "::ffff:169.254.0.1", llu: true},
	{in: "fe80::1", llu: true},
	{in: "fe80::1%10", llu: true},
	{in: "febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff", llu: true},

	// Loopback.
	{in: "127.0.0.1", lo: true},
	{in: "127.255.255.255", lo: true},
	{in: "::1", lo: true},
	{in: "::ffff:127.0.0.1", lo: true},

	// Interface-local multicast.
	{in: "ff01::1", ilm: true, mc: true},
	{in: "ff01::1%10", ilm: true, mc: true},

	// Private.
	{in: "10.0.0.1", gu: true, priv: true},
	{in: "172.16.0.1", gu: true, priv: true},
	{in: "172.31.255.255", gu: true, priv: true},
	{in: "192.168.1.1", gu: true, priv: true},
	{in: "fd00::1", gu: true, priv: true},
	{in: "fc00::1", gu: true, priv: true},
	{in: "fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", gu: true, priv: true},
	{in: "::ffff:10.0.0.1", gu: true, priv: true},
	{in: "::ffff:172.16.0.1", gu: true, priv: true},
	{in: "::ffff:192.168.1.1", gu: true, priv: true},
	// Just outside the private ranges.
	{in: "172.15.0.1", gu: true},
	{in: "172.32.0.1", gu: true},
	{in: "fbff::1", gu: true},
	{in: "fe00::1", gu: true},

	// Unspecified.
	{in: "0.0.0.0", unspec: true},
	{in: "::", unspec: true},
	// A 4-in-6 address is not the same value as the IPv4 address,
	// so the mapped form is not unspecified.
	{in: "::ffff:0.0.0.0"},

	// The broadcast address is not a global unicast address.
	{in: "255.255.255.255"},
	{in: "::ffff:255.255.255.255"},
}

func TestAddrProperties(t *testing.T) {
	for _, tc := range propCases {
		var ip netip.Addr
		if tc.in != "" {
			ip = netip.MustParseAddr(tc.in)
		}

		if got := ip.IsGlobalUnicast(); got != tc.gu {
			t.Errorf("IsGlobalUnicast(%s) = %t, want %t", tc.in, got, tc.gu)
		}
		if got := ip.IsInterfaceLocalMulticast(); got != tc.ilm {
			t.Errorf("IsInterfaceLocalMulticast(%s) = %t, want %t", tc.in, got, tc.ilm)
		}
		if got := ip.IsLinkLocalMulticast(); got != tc.llm {
			t.Errorf("IsLinkLocalMulticast(%s) = %t, want %t", tc.in, got, tc.llm)
		}
		if got := ip.IsLinkLocalUnicast(); got != tc.llu {
			t.Errorf("IsLinkLocalUnicast(%s) = %t, want %t", tc.in, got, tc.llu)
		}
		if got := ip.IsLoopback(); got != tc.lo {
			t.Errorf("IsLoopback(%s) = %t, want %t", tc.in, got, tc.lo)
		}
		if got := ip.IsMulticast(); got != tc.mc {
			t.Errorf("IsMulticast(%s) = %t, want %t", tc.in, got, tc.mc)
		}
		if got := ip.IsPrivate(); got != tc.priv {
			t.Errorf("IsPrivate(%s) = %t, want %t", tc.in, got, tc.priv)
		}
		if got := ip.IsUnspecified(); got != tc.unspec {
			t.Errorf("IsUnspecified(%s) = %t, want %t", tc.in, got, tc.unspec)
		}
	}
}

func TestAddrPropertiesZone(t *testing.T) {
	// A zone does not change a property.
	texts := []string{
		"2001:db8::1", "ff02::1", "ff01::1", "fe80::1", "::1", "fd00::1",
	}
	for _, s := range texts {
		plain := netip.MustParseAddr(s)
		zoned := plain.WithZone("42")

		if zoned.IsGlobalUnicast() != plain.IsGlobalUnicast() {
			t.Errorf("IsGlobalUnicast(%s) changes with a zone", s)
		}
		if zoned.IsInterfaceLocalMulticast() != plain.IsInterfaceLocalMulticast() {
			t.Errorf("IsInterfaceLocalMulticast(%s) changes with a zone", s)
		}
		if zoned.IsLinkLocalMulticast() != plain.IsLinkLocalMulticast() {
			t.Errorf("IsLinkLocalMulticast(%s) changes with a zone", s)
		}
		if zoned.IsLinkLocalUnicast() != plain.IsLinkLocalUnicast() {
			t.Errorf("IsLinkLocalUnicast(%s) changes with a zone", s)
		}
		if zoned.IsLoopback() != plain.IsLoopback() {
			t.Errorf("IsLoopback(%s) changes with a zone", s)
		}
		if zoned.IsMulticast() != plain.IsMulticast() {
			t.Errorf("IsMulticast(%s) changes with a zone", s)
		}
		if zoned.IsPrivate() != plain.IsPrivate() {
			t.Errorf("IsPrivate(%s) changes with a zone", s)
		}
	}

	// IsUnspecified is the one property a zone changes: the zone makes the
	// address a different value.
	zeroZoned := netip.MustParseAddr("::").WithZone("42")
	if zeroZoned.IsUnspecified() {
		t.Error("IsUnspecified(::%42) = true, want false")
	}
}
