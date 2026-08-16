// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

// A prefixCase is one ParsePrefix case with a valid prefix.
type prefixCase struct {
	in   string // the prefix text
	addr string // the address part
	bits int    // the prefix length
	str  string // the String result, "" if String gives in back
}

var prefixCases = []prefixCase{
	{in: "192.168.0.0/24", addr: "192.168.0.0", bits: 24},
	{in: "192.168.1.1/32", addr: "192.168.1.1", bits: 32},
	{in: "100.64.0.0/10", addr: "100.64.0.0", bits: 10},
	{in: "0.0.0.0/0", addr: "0.0.0.0", bits: 0},
	{in: "255.255.255.255/32", addr: "255.255.255.255", bits: 32},
	// The host bits stay as they are.
	{in: "1.2.3.4/24", addr: "1.2.3.4", bits: 24},
	{in: "2001:db8::/96", addr: "2001:db8::", bits: 96},
	{in: "::/0", addr: "::", bits: 0},
	{in: "2000::/3", addr: "2000::", bits: 3},
	{in: "::1/128", addr: "::1", bits: 128},
	{
		in:   "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
		addr: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		bits: 128,
	},
	// A 4-in-6 address keeps its dotted decimal form.
	{in: "::ffff:1.2.3.4/120", addr: "::ffff:1.2.3.4", bits: 120},
	{
		in:   "::ffff:c000:0280/96",
		addr: "::ffff:192.0.2.128",
		bits: 96,
		str:  "::ffff:192.0.2.128/96",
	},
}

// A badPrefixCase is one parse case with an invalid prefix.
type badPrefixCase struct {
	in string // the prefix text
}

var badPrefixCases = []badPrefixCase{
	// No slash.
	{"192.168.0.0"},
	{""},
	// No address.
	{"/24"},
	// A bad address.
	{"1.257.1.1/24"},
	{"bad/24"},
	// No bits.
	{"1.1.1.0/"},
	// Bad bits.
	{"1.1.1.0/q"},
	{"1.1.1.0/-1"},
	{"1.1.1.0/+32"},
	{"1.1.1.0/-32"},
	// A leading zero in the bits.
	{"1.1.1.0/032"},
	{"1.1.1.0/0032"},
	// Bits out of range.
	{"1.1.1.0/33"},
	{"2001::/129"},
	// A zone is not allowed.
	{"1.1.1.0%a/24"},
	{"2001:db8::%10/32"},
	// Two slashes.
	{"1.1.1.0/24/24"},
}

func TestParsePrefix(t *testing.T) {
	for _, tc := range prefixCases {
		p, err := netip.ParsePrefix(tc.in)
		if err != nil {
			t.Errorf("ParsePrefix(%s) failed: %s", tc.in, err.Error())
			continue
		}

		addr := p.Addr()
		if !addr.Equal(netip.MustParseAddr(tc.addr)) {
			var buf [bufLen]byte
			t.Errorf("ParsePrefix(%s).Addr() = %s, want %s",
				tc.in, addr.String(buf[:]), tc.addr)
		}
		if got := p.Bits(); got != tc.bits {
			t.Errorf("ParsePrefix(%s).Bits() = %d, want %d", tc.in, got, tc.bits)
		}
		if !p.IsValid() {
			t.Errorf("ParsePrefix(%s).IsValid() = false, want true", tc.in)
		}

		want := tc.str
		if want == "" {
			want = tc.in
		}
		var buf [bufLen]byte
		s := p.String(buf[:])
		if s != want {
			t.Errorf("ParsePrefix(%s).String() = %s, want %s", tc.in, s, want)
		}

		// ParsePrefix(p.String()) is the identity function.
		p2, err := netip.ParsePrefix(s)
		if err != nil {
			t.Errorf("ParsePrefix(%s) failed: %s", s, err.Error())
			continue
		}
		if !p.Equal(p2) {
			t.Errorf("ParsePrefix(%s) != ParsePrefix(ParsePrefix(%s).String())", tc.in, tc.in)
		}
	}
}

func TestParsePrefixBad(t *testing.T) {
	for _, tc := range badPrefixCases {
		p, err := netip.ParsePrefix(tc.in)
		if err == nil {
			var buf [bufLen]byte
			t.Errorf("ParsePrefix(%s) = %s, want an error", tc.in, p.String(buf[:]))
			continue
		}
		if err != netip.ErrPrefix {
			t.Errorf("ParsePrefix(%s) = %s, want %s", tc.in, err.Error(), netip.ErrPrefix.Error())
		}
		if p.IsValid() {
			t.Errorf("ParsePrefix(%s) failed, but the prefix is valid", tc.in)
		}
	}
}

// A bitsCase is one PrefixFrom case.
type bitsCase struct {
	addr string // the address text, "" for the zero address
	in   int    // the bits argument
	want int    // the Bits result, -1 if the prefix is invalid
}

var bitsCases = []bitsCase{
	{"1.2.3.4", 0, 0},
	{"1.2.3.4", 1, 1},
	{"1.2.3.4", 31, 31},
	{"1.2.3.4", 32, 32},
	{"1.2.3.4", 33, -1},
	{"1.2.3.4", 254, -1},
	{"1.2.3.4", 255, -1},
	{"1.2.3.4", 256, -1},
	{"1.2.3.4", -1, -1},
	{"1.2.3.4", -2, -1},
	{"1.2.3.4", -5, -1},

	{"66::66", 0, 0},
	{"66::66", 33, 33},
	{"66::66", 127, 127},
	{"66::66", 128, 128},
	{"66::66", 129, -1},
	{"66::66", -1, -1},
	{"66::66", -5, -1},

	// The zero address gives an invalid prefix for every bit count.
	{"", 0, -1},
	{"", 32, -1},
	{"", 128, -1},
	{"", -1, -1},
}

func TestPrefixFrom(t *testing.T) {
	var zoneBuf [netip.MaxZoneLen]byte

	for _, tc := range bitsCases {
		var ip netip.Addr
		if tc.addr != "" {
			ip = netip.MustParseAddr(tc.addr)
		}
		p := netip.PrefixFrom(ip, tc.in)

		if got := p.Bits(); got != tc.want {
			t.Errorf("PrefixFrom(%s, %d).Bits() = %d, want %d", tc.addr, tc.in, got, tc.want)
		}
		if got := p.IsValid(); got != (tc.want >= 0) {
			t.Errorf("PrefixFrom(%s, %d).IsValid() = %t, want %t",
				tc.addr, tc.in, got, tc.want >= 0)
		}
		if !p.Addr().Equal(ip) {
			t.Errorf("PrefixFrom(%s, %d).Addr() is not the address", tc.addr, tc.in)
		}

		// There is only one invalid prefix per address.
		if tc.want < 0 {
			invalid := netip.PrefixFrom(p.Addr(), -1)
			if !p.Equal(invalid) {
				t.Errorf("PrefixFrom(%s, %d) != PrefixFrom(%s, -1)", tc.addr, tc.in, tc.addr)
			}
		}
	}

	// PrefixFrom drops the zone.
	ip := netip.MustParseAddr("::1%10")
	p := netip.PrefixFrom(ip, 128)
	if zoneOf(p.Addr(), zoneBuf[:]) != "" {
		t.Error("PrefixFrom kept the zone")
	}
	if !p.Addr().Equal(netip.MustParseAddr("::1")) {
		t.Error("PrefixFrom(::1%10).Addr() is not ::1")
	}
}

// A maskCase is one case of the Addr.Prefix and Prefix.Masked methods.
type maskCase struct {
	addr string // the address text
	bits int    // the prefix length
	want string // the masked prefix text
}

var maskCases = []maskCase{
	{"192.0.2.0", 16, "192.0.0.0/16"},
	{"192.168.0.255", 24, "192.168.0.0/24"},
	{"255.255.255.255", 20, "255.255.240.0/20"},
	{"255.255.255.255", 0, "0.0.0.0/0"},
	{"255.255.255.255", 32, "255.255.255.255/32"},
	// One byte holds both masked and unmasked bits.
	{"100.98.156.66", 10, "100.64.0.0/10"},
	{"2001:db8::1", 32, "2001:db8::/32"},
	{"2100::", 3, "2000::/3"},
	{"fe80::dead:beef:dead:beef", 96, "fe80::dead:beef:0:0/96"},
	{"aaaa::", 4, "a000::/4"},
	{"::", 63, "::/63"},
	{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", 65, "ffff:ffff:ffff:ffff:8000::/65"},
	{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", 128, "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
	{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", 0, "::/0"},
}

func TestAddrPrefix(t *testing.T) {
	var zoneBuf [netip.MaxZoneLen]byte

	for _, tc := range maskCases {
		ip := netip.MustParseAddr(tc.addr)
		p, err := ip.Prefix(tc.bits)
		if err != nil {
			t.Errorf("Prefix(%s, %d) failed: %s", tc.addr, tc.bits, err.Error())
			continue
		}
		var buf [bufLen]byte
		if got := p.String(buf[:]); got != tc.want {
			t.Errorf("Prefix(%s, %d) = %s, want %s", tc.addr, tc.bits, got, tc.want)
		}

		// The prefix is already masked, so Masked gives it back.
		var buf2 [bufLen]byte
		if got := p.Masked().String(buf2[:]); got != tc.want {
			t.Errorf("Prefix(%s, %d).Masked() = %s, want %s", tc.addr, tc.bits, got, tc.want)
		}

		// Masked gives the same prefix as Prefix.
		unmasked := netip.PrefixFrom(ip, tc.bits)
		var buf3 [bufLen]byte
		if got := unmasked.Masked().String(buf3[:]); got != tc.want {
			t.Errorf("PrefixFrom(%s, %d).Masked() = %s, want %s", tc.addr, tc.bits, got, tc.want)
		}
	}

	// A bit count out of range gives an error.
	ip4 := netip.MustParseAddr("1.2.3.4")
	if _, err := ip4.Prefix(33); err != netip.ErrLargePrefix {
		t.Error("Prefix(1.2.3.4, 33) does not report a too large prefix")
	}
	ip6 := netip.MustParseAddr("::1")
	if _, err := ip6.Prefix(129); err != netip.ErrLargePrefix {
		t.Error("Prefix(::1, 129) does not report a too large prefix")
	}
	if _, err := ip4.Prefix(-1); err != netip.ErrNegativePrefix {
		t.Error("Prefix(1.2.3.4, -1) does not report a negative prefix")
	}
	if _, err := ip6.Prefix(-1); err != netip.ErrNegativePrefix {
		t.Error("Prefix(::1, -1) does not report a negative prefix")
	}

	// The zero address gives the zero prefix and no error.
	var zero netip.Addr
	p, err := zero.Prefix(16)
	if err != nil {
		t.Errorf("Prefix(Addr{}, 16) failed: %s", err.Error())
		return
	}
	if p.IsValid() {
		t.Error("Prefix(Addr{}, 16) is valid, want the zero prefix")
	}
	if _, err := zero.Prefix(-1); err != netip.ErrNegativePrefix {
		t.Error("Prefix(Addr{}, -1) does not report a negative prefix")
	}

	// The prefix of an address with a zone has no zone.
	p, err = netip.MustParseAddr("2001:db8::1%10").Prefix(32)
	if err != nil {
		t.Errorf("Prefix(2001:db8::1%%10, 32) failed: %s", err.Error())
		return
	}
	if zoneOf(p.Addr(), zoneBuf[:]) != "" {
		t.Error("Prefix() kept the zone")
	}
	var buf [bufLen]byte
	if got := p.String(buf[:]); got != "2001:db8::/32" {
		t.Errorf("Prefix(2001:db8::1%%10, 32) = %s, want 2001:db8::/32", got)
	}
}

func TestPrefixMasked(t *testing.T) {
	// An invalid prefix gives the zero prefix.
	p := netip.PrefixFrom(netip.MustParseAddr("2000::"), 129)
	if p.Masked().IsValid() {
		t.Error("Masked() of an invalid prefix is valid")
	}
	p = netip.PrefixFrom(netip.MustParseAddr("1.2.3.4"), 33)
	if p.Masked().IsValid() {
		t.Error("Masked() of an invalid prefix is valid")
	}
	var zero netip.Prefix
	if zero.Masked().IsValid() {
		t.Error("Prefix{}.Masked() is valid")
	}
}

// A containsCase is one case of the Prefix.Contains method.
// An empty address text means the zero address.
type containsCase struct {
	prefix string // the prefix text
	addr   string // the address text
	want   bool   // the Contains result
}

var containsCases = []containsCase{
	{"9.8.7.6/0", "9.8.7.6", true},
	{"9.8.7.6/0", "0.0.0.0", true},
	{"9.8.7.6/0", "255.255.255.255", true},
	{"9.8.7.6/16", "9.8.7.6", true},
	{"9.8.7.6/16", "9.8.6.4", true},
	{"9.8.7.6/16", "9.9.7.6", false},
	{"9.8.7.6/32", "9.8.7.6", true},
	{"9.8.7.6/32", "9.8.7.7", false},
	{"192.168.0.0/24", "192.168.0.1", true},
	{"192.168.0.0/24", "192.168.0.55", true},
	{"192.168.0.0/24", "192.168.1.1", false},
	{"192.168.0.0/24", "1.1.1.1", false},
	// A prefix that is not a multiple of 8.
	{"100.64.0.0/10", "100.64.0.0", true},
	{"100.64.0.0/10", "100.64.0.1", true},
	{"100.64.0.0/10", "100.81.251.94", true},
	{"100.64.0.0/10", "100.100.100.100", true},
	{"100.64.0.0/10", "100.127.255.254", true},
	{"100.64.0.0/10", "100.127.255.255", true},
	{"100.64.0.0/10", "100.63.255.255", false},
	{"100.64.0.0/10", "100.128.0.0", false},

	{"::1/0", "::1", true},
	{"::1/0", "::2", true},
	{"::1/127", "::1", true},
	{"::1/127", "::2", false},
	{"::1/128", "::1", true},
	{"2001:db8::/96", "2001:db8::1", true},
	{"2001:db8::/96", "2001:db8::aaaa:bbbb", true},
	{"2001:db8::/96", "2001:db8::1:aaaa:bbbb", false},
	{"2001:db8::/96", "2001:db9::", false},
	{"2000::/3", "2001:db8::1", true},
	{"2000::/3", "fe80::1", false},
	{"::/0", "::1", true},
	{"::/0", "2001:db8::1", true},

	// The zero address is in no prefix.
	{"::1/0", "", false},
	{"1.2.3.4/0", "", false},
	// An address with a zone is in no prefix.
	{"::1/128", "::1%10", false},
	{"::/0", "::1%10", false},
	// One address family does not match the other.
	{"::1/0", "1.2.3.4", false},
	{"1.2.3.4/0", "::1", false},
	{"::/0", "192.0.2.1", false},
	{"0.0.0.0/0", "2001:db8::1", false},
	// A 4-in-6 address does not match an IPv4 prefix.
	{"1.2.3.0/24", "::ffff:1.2.3.4", false},
	{"::ffff:1.2.3.0/120", "1.2.3.4", false},
	{"::ffff:1.2.3.0/120", "::ffff:1.2.3.4", true},
}

func TestPrefixContains(t *testing.T) {
	for _, tc := range containsCases {
		p := netip.MustParsePrefix(tc.prefix)
		var ip netip.Addr
		if tc.addr != "" {
			ip = netip.MustParseAddr(tc.addr)
		}
		if got := p.Contains(ip); got != tc.want {
			t.Errorf("Contains(%s, %s) = %t, want %t", tc.prefix, tc.addr, got, tc.want)
		}
	}

	// An invalid prefix contains nothing.
	ip4 := netip.MustParseAddr("1.2.3.4")
	ip6 := netip.MustParseAddr("::1")
	var zero netip.Addr
	if netip.PrefixFrom(ip4, 33).Contains(ip4) {
		t.Error("an invalid prefix contains an address")
	}
	if netip.PrefixFrom(ip6, 129).Contains(ip6) {
		t.Error("an invalid prefix contains an address")
	}
	if netip.PrefixFrom(zero, 0).Contains(ip4) {
		t.Error("the zero prefix contains an address")
	}
	if netip.PrefixFrom(zero, 32).Contains(ip4) {
		t.Error("the zero prefix contains an address")
	}
	if netip.PrefixFrom(zero, 128).Contains(ip6) {
		t.Error("the zero prefix contains an address")
	}

	// The zone of the prefix address is dropped, so it does not matter.
	p := netip.PrefixFrom(ip4.WithZone("10"), 32)
	if !p.Contains(ip4) {
		t.Error("a prefix from an address with a zone contains nothing")
	}
	p = netip.PrefixFrom(ip6.WithZone("10"), 128)
	if !p.Contains(ip6) {
		t.Error("a prefix from an address with a zone contains nothing")
	}
}

// An overlapsCase is one case of the Prefix.Overlaps method.
type overlapsCase struct {
	a    string // the first prefix
	b    string // the second prefix
	want bool   // the Overlaps result
}

var overlapsCases = []overlapsCase{
	// Different address families.
	{"::/3", "0.0.0.0/3", false},

	{"1.2.0.0/16", "1.2.0.0/16", true},
	{"1.2.0.0/16", "1.2.3.0/24", true},
	{"1.2.0.0/16", "1.2.3.0/32", true},
	{"1.2.0.0/16", "1.3.0.0/16", false},
	{"1.2.3.0/32", "1.2.3.1/32", false},

	// A prefix of zero bits overlaps everything of its family.
	{"1.2.3.0/32", "0.0.0.0/0", true},
	// The address of the prefix needs no masking.
	{"1.2.3.0/32", "5.5.5.5/0", true},

	{"5::1/128", "5::/8", true},
	{"1::1/128", "2::2/128", false},
	{"100::/8", "::1/128", false},
	{"2001:db8::/32", "2001:db8:1::/48", true},

	// A 4-in-6 address does not overlap an IPv4 address.
	{"::ffff:1.2.0.0/16", "1.2.3.0/24", false},
}

func TestPrefixOverlaps(t *testing.T) {
	for _, tc := range overlapsCases {
		a := netip.MustParsePrefix(tc.a)
		b := netip.MustParsePrefix(tc.b)
		if got := a.Overlaps(b); got != tc.want {
			t.Errorf("Overlaps(%s, %s) = %t, want %t", tc.a, tc.b, got, tc.want)
		}
		// Overlaps is commutative.
		if got := b.Overlaps(a); got != tc.want {
			t.Errorf("Overlaps(%s, %s) = %t, want %t", tc.b, tc.a, got, tc.want)
		}
	}

	// An invalid prefix overlaps nothing.
	var zero netip.Prefix
	p := netip.MustParsePrefix("1.2.0.0/16")
	if zero.Overlaps(p) || p.Overlaps(zero) {
		t.Error("the zero prefix overlaps a prefix")
	}
	bad := netip.PrefixFrom(netip.MustParseAddr("1.2.3.4"), 33)
	if bad.Overlaps(p) || p.Overlaps(bad) {
		t.Error("an invalid prefix overlaps a prefix")
	}
	bad = netip.PrefixFrom(netip.MustParseAddr("2000::"), 129)
	p = netip.MustParsePrefix("2000::/64")
	if bad.Overlaps(p) || p.Overlaps(bad) {
		t.Error("an invalid prefix overlaps a prefix")
	}
}

func TestPrefixIsSingleIP(t *testing.T) {
	singles := []string{"127.0.0.1/32", "0.0.0.0/32", "::1/128", "::/128"}
	for _, s := range singles {
		if !netip.MustParsePrefix(s).IsSingleIP() {
			t.Errorf("IsSingleIP(%s) = false, want true", s)
		}
	}

	many := []string{"127.0.0.1/31", "127.0.0.1/0", "::1/127", "::1/0"}
	for _, s := range many {
		if netip.MustParsePrefix(s).IsSingleIP() {
			t.Errorf("IsSingleIP(%s) = true, want false", s)
		}
	}

	var zero netip.Prefix
	if zero.IsSingleIP() {
		t.Error("Prefix{}.IsSingleIP() = true, want false")
	}
	bad := netip.PrefixFrom(netip.MustParseAddr("1.2.3.4"), 33)
	if bad.IsSingleIP() {
		t.Error("an invalid prefix holds a single address")
	}
}

// A prefixCompareCase is one case of the Prefix.Compare method.
// An empty prefix text means the zero prefix.
type prefixCompareCase struct {
	a    string // the first prefix
	b    string // the second prefix
	want int    // the Compare result
}

var prefixCompareCases = []prefixCompareCase{
	{"", "", 0},
	{"", "1.2.3.0/24", -1},

	{"1.2.3.0/24", "1.2.3.0/24", 0},
	{"fe80::/64", "fe80::/64", 0},

	// The masked address sorts first.
	{"1.2.3.0/24", "1.2.4.0/24", -1},
	{"fe80::/64", "fe90::/64", -1},
	// Then the prefix length.
	{"1.2.0.0/16", "1.2.0.0/24", -1},
	{"fe80::/48", "fe80::/64", -1},
	// An IPv4 prefix sorts before an IPv6 prefix.
	{"1.2.3.0/24", "fe80::/8", -1},
	// Then the unmasked address.
	{"1.2.3.0/24", "1.2.3.4/24", -1},
	{"1.2.3.0/24", "1.2.3.0/28", -1},
}

func TestPrefixCompare(t *testing.T) {
	for _, tc := range prefixCompareCases {
		var a, b netip.Prefix
		if tc.a != "" {
			a = netip.MustParsePrefix(tc.a)
		}
		if tc.b != "" {
			b = netip.MustParsePrefix(tc.b)
		}

		if got := a.Compare(b); got != tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
		if back := b.Compare(a); back != -tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.b, tc.a, back, -tc.want)
		}
		if eq := a.Equal(b); eq != (tc.want == 0) {
			t.Errorf("Equal(%s, %s) = %t, want %t", tc.a, tc.b, eq, tc.want == 0)
		}
	}
}

func TestPrefixSort(t *testing.T) {
	var zero netip.Prefix
	values := []netip.Prefix{
		netip.MustParsePrefix("1.2.3.0/24"),
		netip.MustParsePrefix("fe90::/64"),
		netip.MustParsePrefix("fe80::/64"),
		netip.MustParsePrefix("1.2.0.0/16"),
		zero,
		netip.MustParsePrefix("fe80::/48"),
		netip.MustParsePrefix("1.2.0.0/24"),
		netip.MustParsePrefix("1.2.3.4/24"),
		netip.MustParsePrefix("1.2.3.0/28"),
	}
	want := []string{
		"invalid Prefix", "1.2.0.0/16", "1.2.0.0/24", "1.2.3.0/24",
		"1.2.3.4/24", "1.2.3.0/28", "fe80::/48", "fe80::/64", "fe90::/64",
	}

	slices.SortFunc(values, comparePrefixes)
	for i, v := range values {
		var buf [bufLen]byte
		if got := v.String(buf[:]); got != want[i] {
			t.Errorf("sorted[%d] = %s, want %s", i, got, want[i])
		}
	}
}

// The special purpose prefixes of IANA, in the order the registry lists them.
// The sort must give the conventional order of the registry.
//
// See https://www.iana.org/assignments/iana-ipv4-special-registry
// and https://www.iana.org/assignments/ipv6-address-space
var ianaPrefixes = [...]string{
	"0.0.0.0/8",
	"127.0.0.0/8",
	"10.0.0.0/8",
	"203.0.113.0/24",
	"169.254.0.0/16",
	"192.0.0.0/24",
	"240.0.0.0/4",
	"192.0.2.0/24",
	"192.0.0.170/32",
	"198.18.0.0/15",
	"192.0.0.8/32",
	"0.0.0.0/32",
	"192.0.0.9/32",
	"198.51.100.0/24",
	"192.168.0.0/16",
	"192.0.0.10/32",
	"192.175.48.0/24",
	"192.52.193.0/24",
	"100.64.0.0/10",
	"255.255.255.255/32",
	"192.31.196.0/24",
	"172.16.0.0/12",
	"192.0.0.0/29",
	"192.88.99.0/24",
	"fec0::/10",
	"6000::/3",
	"fe00::/9",
	"8000::/3",
	"0000::/8",
	"0400::/6",
	"f800::/6",
	"e000::/4",
	"ff00::/8",
	"a000::/3",
	"fc00::/7",
	"1000::/4",
	"0800::/5",
	"4000::/3",
	"0100::/8",
	"c000::/3",
	"fe80::/10",
	"0200::/7",
	"f000::/5",
	"2000::/3",
}

// The IANA prefixes in sorted order.
var ianaSorted = [...]string{
	"0.0.0.0/8",
	"0.0.0.0/32",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.0.0/29",
	"192.0.0.8/32",
	"192.0.0.9/32",
	"192.0.0.10/32",
	"192.0.0.170/32",
	"192.0.2.0/24",
	"192.31.196.0/24",
	"192.52.193.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"192.175.48.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
	"255.255.255.255/32",
	"::/8",
	"100::/8",
	"200::/7",
	"400::/6",
	"800::/5",
	"1000::/4",
	"2000::/3",
	"4000::/3",
	"6000::/3",
	"8000::/3",
	"a000::/3",
	"c000::/3",
	"e000::/4",
	"f000::/5",
	"f800::/6",
	"fc00::/7",
	"fe00::/9",
	"fe80::/10",
	"fec0::/10",
	"ff00::/8",
}

func TestPrefixSortIANA(t *testing.T) {
	var values [len(ianaPrefixes)]netip.Prefix
	for i, s := range ianaPrefixes {
		values[i] = netip.MustParsePrefix(s)
	}

	slices.SortFunc(values[:], comparePrefixes)
	for i, v := range values {
		var buf [bufLen]byte
		if got := v.String(buf[:]); got != ianaSorted[i] {
			t.Errorf("sorted[%d] = %s, want %s", i, got, ianaSorted[i])
		}
	}
}

func TestPrefixString(t *testing.T) {
	var buf [bufLen]byte

	var zero netip.Prefix
	if got := zero.String(buf[:]); got != "invalid Prefix" {
		t.Errorf("Prefix{}.String() = %s, want invalid Prefix", got)
	}

	var zeroAddr netip.Addr
	p := netip.PrefixFrom(zeroAddr, 8)
	if got := p.String(buf[:]); got != "invalid Prefix" {
		t.Errorf("PrefixFrom(Addr{}, 8).String() = %s, want invalid Prefix", got)
	}

	p = netip.PrefixFrom(netip.MustParseAddr("1.2.3.4"), 88)
	if got := p.String(buf[:]); got != "invalid Prefix" {
		t.Errorf("PrefixFrom(1.2.3.4, 88).String() = %s, want invalid Prefix", got)
	}
}

func TestPrefixAppendText(t *testing.T) {
	for _, tc := range prefixCases {
		p := netip.MustParsePrefix(tc.in)
		var buf [bufLen]byte
		b, err := p.AppendText(buf[:0])
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

	// The zero prefix appends nothing.
	var zero netip.Prefix
	var buf [bufLen]byte
	b, err := zero.AppendText(buf[:0])
	if err != nil {
		t.Errorf("Prefix{}.AppendText() failed: %s", err.Error())
		return
	}
	if len(b) != 0 {
		t.Errorf("Prefix{}.AppendText() = %s, want an empty result", string(b))
	}

	// An invalid prefix with an address appends the invalid text.
	p := netip.PrefixFrom(netip.MustParseAddr("1.2.3.4"), 88)
	b, err = p.AppendText(buf[:0])
	if err != nil {
		t.Errorf("AppendText() failed: %s", err.Error())
		return
	}
	if got := string(b); got != "invalid Prefix" {
		t.Errorf("AppendText() = %s, want invalid Prefix", got)
	}
}

func TestPrefixMaxLen(t *testing.T) {
	// Write the longest prefix text into a buffer of the documented size.
	var buf [netip.MaxPrefixLen]byte
	p := netip.MustParsePrefix("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128")
	want := "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"
	if got := p.String(buf[:]); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}

	var bufAppend [netip.MaxPrefixLen]byte
	b, err := p.AppendText(bufAppend[:0])
	if err != nil {
		t.Errorf("AppendText() failed: %s", err.Error())
		return
	}
	if got := string(b); got != want {
		t.Errorf("AppendText() = %s, want %s", got, want)
	}
}
