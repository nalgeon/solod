// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/net/netip"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

// The error a bad address text must give.
// A test table names one of these codes.
const (
	errAny    = iota // any error does
	errIP            // netip.ErrIP
	errIPv4          // netip.ErrIPv4
	errIPv6          // netip.ErrIPv6
	errIPPort        // netip.ErrIPPort
	errPort          // netip.ErrPort
	errPrefix        // netip.ErrPrefix
)

// A parseCase is one ParseAddr case with a valid address.
//
// The tests use a numeric zone. An interface name resolves to an index only in
// a hosted build, and the index depends on the host, so a name gives no stable
// expected value.
type parseCase struct {
	in   string // the address text
	hi   uint64 // the top 64 bits of the address
	lo   uint64 // the bottom 64 bits of the address
	bits int    // the address length in bits
	zone string // the zone as text, "" for an address without a zone
	str  string // the String result, "" if String gives in back
}

var parseCases = []parseCase{
	// Basic zero IPv4 address.
	{in: "0.0.0.0", hi: 0, lo: 0xffff00000000, bits: 32},
	// Basic non-zero IPv4 address.
	{in: "192.168.140.255", hi: 0, lo: 0xffffc0a88cff, bits: 32},
	// IPv4 broadcast address.
	{in: "255.255.255.255", hi: 0, lo: 0xffffffffffff, bits: 32},
	// Basic zero IPv6 address.
	{in: "::", hi: 0, lo: 0, bits: 128},
	// Localhost IPv6.
	{in: "::1", hi: 0, lo: 1, bits: 128},
	// Fully expanded IPv6 address.
	{
		in:   "fd7a:115c:a1e0:ab12:4843:cd96:626b:430b",
		hi:   0xfd7a115ca1e0ab12,
		lo:   0x4843cd96626b430b,
		bits: 128,
	},
	// IPv6 with elided fields in the middle.
	{
		in:   "fd7a:115c::626b:430b",
		hi:   0xfd7a115c00000000,
		lo:   0x00000000626b430b,
		bits: 128,
	},
	// IPv6 with elided fields at the end.
	{
		in:   "fd7a:115c:a1e0:ab12:4843:cd96::",
		hi:   0xfd7a115ca1e0ab12,
		lo:   0x4843cd9600000000,
		bits: 128,
	},
	// IPv6 with single elided field at the end.
	{
		in:   "fd7a:115c:a1e0:ab12:4843:cd96:626b::",
		hi:   0xfd7a115ca1e0ab12,
		lo:   0x4843cd96626b0000,
		bits: 128,
		str:  "fd7a:115c:a1e0:ab12:4843:cd96:626b:0",
	},
	// IPv6 with single elided field in the middle.
	{
		in:   "fd7a:115c:a1e0::4843:cd96:626b:430b",
		hi:   0xfd7a115ca1e00000,
		lo:   0x4843cd96626b430b,
		bits: 128,
		str:  "fd7a:115c:a1e0:0:4843:cd96:626b:430b",
	},
	// IPv6 with the trailing 32 bits written as IPv4 dotted decimal (4in6).
	{
		in:   "::ffff:192.168.140.255",
		hi:   0,
		lo:   0x0000ffffc0a88cff,
		bits: 128,
	},
	// The same address in hex form.
	{
		in:   "::ffff:c0a8:8cff",
		hi:   0,
		lo:   0x0000ffffc0a88cff,
		bits: 128,
		str:  "::ffff:192.168.140.255",
	},
	// IPv6 with a zone.
	{
		in:   "fd7a:115c:a1e0:ab12:4843:cd96:626b:430b%10",
		hi:   0xfd7a115ca1e0ab12,
		lo:   0x4843cd96626b430b,
		bits: 128,
		zone: "10",
	},
	// IPv6 with dotted decimal and a zone.
	{
		in:   "1:2::ffff:192.168.140.255%11",
		hi:   0x0001000200000000,
		lo:   0x0000ffffc0a88cff,
		bits: 128,
		zone: "11",
		str:  "1:2::ffff:c0a8:8cff%11",
	},
	// 4-in-6 with a zone.
	{
		in:   "::ffff:192.168.140.255%11",
		hi:   0,
		lo:   0x0000ffffc0a88cff,
		bits: 128,
		zone: "11",
	},
	// A zone at the top of the range.
	{
		in:   "::1%4294967295",
		hi:   0,
		lo:   1,
		bits: 128,
		zone: "4294967295",
	},
	// IPv6 with capital letters.
	{
		in:   "FD9E:1A04:F01D::1",
		hi:   0xfd9e1a04f01d0000,
		lo:   0x1,
		bits: 128,
		str:  "fd9e:1a04:f01d::1",
	},
	// The highest IPv6 address.
	{
		in:   "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		hi:   0xffffffffffffffff,
		lo:   0xffffffffffffffff,
		bits: 128,
	},
	// An ellipsis that covers a single group is written out.
	{
		in:   "1:2:3:4:5:6::8",
		hi:   0x0001000200030004,
		lo:   0x0005000600000008,
		bits: 128,
		str:  "1:2:3:4:5:6:0:8",
	},
}

// A badCase is one parse case with an invalid address.
type badCase struct {
	in  string // the address text
	err int    // the expected error
}

var badCases = []badCase{
	// Empty string.
	{"", errIP},
	// Garbage non-IP.
	{"bad", errIP},
	{"localhost", errIP},
	// Single number. Some parsers accept this as an IPv4 address in
	// big-endian uint32 form, but we don't.
	{"1234", errIP},
	// IPv4 address in windows-style "print all the digits" form.
	{"010.000.015.001", errIPv4},
	// IPv4 address with a silly amount of leading zeros.
	{"000001.00000002.00000003.000000004", errIPv4},
	// IPv4 with a zone.
	{"1.2.3.4%1", errIPv4},
	// IPv4 field must have at least one digit.
	{".1.2.3", errIPv4},
	{"1.2.3.", errIPv4},
	{"1..2.3", errIPv4},
	// IPv4 address too long.
	{"1.2.3.4.5", errIPv4},
	{"192.168.0.1.5.6", errIPv4},
	// IPv4 in dotted octal form.
	{"0300.0250.0214.0377", errIPv4},
	// IPv4 in dotted hex form.
	{"0xc0.0xa8.0x8c.0xff", errIPv4},
	// IPv4 in class B form.
	{"192.168.12345", errIPv4},
	// IPv4 in class B form, with a small enough number to be
	// parseable as a regular dotted decimal field.
	{"127.0.1", errIPv4},
	// IPv4 in class A form.
	{"192.1234567", errIPv4},
	// IPv4 in class A form, with a small enough number to be
	// parseable as a regular dotted decimal field.
	{"127.1", errIPv4},
	// IPv4 field has a value above 255.
	{"192.168.300.1", errIPv4},
	{"500.0.0.1", errIPv4},
	// 4-in-6 with an octet with a leading zero.
	{"::ffff:1.2.03.4", errIPv4},
	// 4-in-6 with an unexpected character in an octet.
	{"::ffff:1.2.3.z", errIPv4},
	// IPv6 with invalid embedded IPv4.
	{"::ffff:192.168.140.bad", errIPv4},
	// IPv6 with not enough fields.
	{"1:2:3:4:5:6:7", errIPv6},
	// IPv6 with too many fields.
	{"1:2:3:4:5:6:7:8:9", errIPv6},
	// IPv6 with 8 fields and a :: expander.
	{"1:2:3:4::5:6:7:8", errIPv6},
	// IPv6 with a field above 16 bits.
	{"fe801::1", errIPv6},
	// IPv6 with non-hex values in a field.
	{"fe80:tail:scal:e::", errIPv6},
	{"::gggg%1", errIPv6},
	// IPv6 with a zone delimiter but no zone.
	{"fe80::1%", errIPv6},
	{"fe80::1cc0:3e8c:119f:c2e1%", errIPv6},
	// A zone with no address.
	{"%", errIPv6},
	{"%1", errIPv6},
	// IPv6 (without ellipsis) with too many fields for a trailing IPv4.
	{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:192.168.140.255", errIPv6},
	// IPv6 (with ellipsis) with too many fields for a trailing IPv4.
	{"ffff::ffff:ffff:ffff:ffff:ffff:ffff:192.168.140.255", errIPv6},
	// IPv6 with multiple ellipsis.
	{"fe80::1::1", errIPv6},
	// IPv6 with an invalid non hex or colon character.
	{"fe80:1?:1", errIPv6},
	// IPv6 with truncated bytes after a single colon.
	{"fe80:", errIPv6},
	{"::1:", errIPv6},
	// IPv6 with 5 zeros in the last group.
	{"0:0:0:0:0:ffff:0:00000", errIPv6},
	// IPv6 with 5 zeros in one group and an embedded IPv4.
	{"0:0:0:0:00000:ffff:127.1.2.3", errIPv6},
	// An ellipsis that covers no group.
	{"1:2:3:4:5:6:7::8", errIPv6},
}

func TestParseAddr(t *testing.T) {
	var zoneBuf [netip.MaxZoneLen]byte
	for _, tc := range parseCases {
		ip, err := netip.ParseAddr(tc.in)
		if err != nil {
			t.Errorf("ParseAddr(%s) failed: %s", tc.in, err.Error())
			continue
		}

		hi, lo := addrParts(ip)
		if hi != tc.hi || lo != tc.lo {
			t.Errorf("ParseAddr(%s) = %x %x, want %x %x", tc.in, hi, lo, tc.hi, tc.lo)
		}
		if got := ip.BitLen(); got != tc.bits {
			t.Errorf("ParseAddr(%s).BitLen() = %d, want %d", tc.in, got, tc.bits)
		}
		if got := zoneOf(ip, zoneBuf[:]); got != tc.zone {
			t.Errorf("ParseAddr(%s).Zone() = %s, want %s", tc.in, got, tc.zone)
		}
		if !ip.IsValid() {
			t.Errorf("ParseAddr(%s).IsValid() = false, want true", tc.in)
		}

		// ParseAddr is a pure function.
		ip2, err := netip.ParseAddr(tc.in)
		if err != nil {
			t.Errorf("ParseAddr(%s) failed on the second call: %s", tc.in, err.Error())
			continue
		}
		if !ip.Equal(ip2) {
			t.Errorf("ParseAddr(%s) gave two different results", tc.in)
		}

		// The address formats as expected.
		want := tc.str
		if want == "" {
			want = tc.in
		}
		var buf [bufLen]byte
		s := ip.String(buf[:])
		if s != want {
			t.Errorf("ParseAddr(%s).String() = %s, want %s", tc.in, s, want)
		}

		// ParseAddr(ip.String()) is the identity function.
		ip3, err := netip.ParseAddr(s)
		if err != nil {
			t.Errorf("ParseAddr(%s) failed: %s", s, err.Error())
			continue
		}
		if !ip.Equal(ip3) {
			t.Errorf("ParseAddr(%s) != ParseAddr(ParseAddr(%s).String())", tc.in, tc.in)
		}
	}
}

func TestParseAddrBad(t *testing.T) {
	for _, tc := range badCases {
		ip, err := netip.ParseAddr(tc.in)
		if err == nil {
			var buf [bufLen]byte
			t.Errorf("ParseAddr(%s) = %s, want an error", tc.in, ip.String(buf[:]))
			continue
		}
		if want := wantErr(tc.err); want != nil && err != want {
			t.Errorf("ParseAddr(%s) = %s, want %s", tc.in, err.Error(), want.Error())
		}
		if ip.IsValid() {
			t.Errorf("ParseAddr(%s) failed, but the address is valid", tc.in)
		}
	}
}

// A stringCase is one String case. It parses in, then formats the address.
type stringCase struct {
	in   string // the address text
	want string // the String result
}

// The cases follow section 4 of RFC 5952: no unnecessary zeros, "::" elides the
// longest run of zero groups, and "::" does not compact a single zero group.
var stringCases = []stringCase{
	{"1:2:3:4:5:6:7:8", "1:2:3:4:5:6:7:8"},
	{"0:0:0:0:0:0:0:0", "::"},
	{"0:0:0:0:0:0:0:1", "::1"},
	{"1:0:0:0:0:0:0:0", "1::"},
	// A single zero group stays.
	{"1:0:2:3:4:5:6:7", "1:0:2:3:4:5:6:7"},
	{"1:2:3:4:5:6:0:8", "1:2:3:4:5:6:0:8"},
	// The longest run wins.
	{"1:0:0:2:0:0:0:3", "1:0:0:2::3"},
	{"1:0:0:0:2:0:0:3", "1::2:0:0:3"},
	// The first of two equal runs wins.
	{"1:0:0:2:0:0:3:4", "1::2:0:0:3:4"},
	// No leading zeros in a group.
	{"0001:0002:0003:0004:0005:0006:0007:0008", "1:2:3:4:5:6:7:8"},
	{"ff:f:0:00:000:0000:1:1", "ff:f::1:1"},
	// A 4-in-6 address formats as dotted decimal.
	{"::ffff:1.2.3.4", "::ffff:1.2.3.4"},
	{"::ffff:0.0.0.0", "::ffff:0.0.0.0"},
	{"::ffff:255.255.255.255", "::ffff:255.255.255.255"},
	// An address close to 4-in-6 stays in hex.
	{"::fffe:c000:280", "::fffe:c000:280"},
	{"::1:ffff:c000:280", "::1:ffff:c000:280"},
	// IPv4.
	{"0.0.0.0", "0.0.0.0"},
	{"1.2.3.4", "1.2.3.4"},
	{"255.255.255.255", "255.255.255.255"},
}

func TestAddrString(t *testing.T) {
	for _, tc := range stringCases {
		ip := netip.MustParseAddr(tc.in)
		var buf [bufLen]byte
		if got := ip.String(buf[:]); got != tc.want {
			t.Errorf("ParseAddr(%s).String() = %s, want %s", tc.in, got, tc.want)
		}
	}

	var zero netip.Addr
	var buf [bufLen]byte
	if got := zero.String(buf[:]); got != "invalid IP" {
		t.Errorf("Addr{}.String() = %s, want invalid IP", got)
	}
}

func TestAddrAppendText(t *testing.T) {
	for _, tc := range stringCases {
		ip := netip.MustParseAddr(tc.in)
		var buf [bufLen]byte
		b := append(buf[:0], "ip="...)
		b, err := ip.AppendText(b)
		if err != nil {
			t.Errorf("AppendText(%s) failed: %s", tc.in, err.Error())
			continue
		}
		if string(b[:3]) != "ip=" {
			t.Errorf("AppendText(%s) overwrote the buffer", tc.in)
		}
		if got := string(b[3:]); got != tc.want {
			t.Errorf("AppendText(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}

	// The zero address appends nothing.
	var zero netip.Addr
	var buf [bufLen]byte
	b, err := zero.AppendText(buf[:0])
	if err != nil {
		t.Errorf("Addr{}.AppendText() failed: %s", err.Error())
		return
	}
	if len(b) != 0 {
		t.Errorf("Addr{}.AppendText() = %s, want an empty result", string(b))
	}
}

func TestAddrFrom4(t *testing.T) {
	a4 := [4]byte{1, 2, 3, 4}
	ip := netip.AddrFrom4(a4)
	if !ip.Equal(netip.MustParseAddr("1.2.3.4")) {
		t.Error("AddrFrom4(1.2.3.4) != ParseAddr(1.2.3.4)")
	}
	if !ip.Is4() {
		t.Error("AddrFrom4().Is4() = false, want true")
	}
	if ip.BitLen() != 32 {
		t.Errorf("AddrFrom4().BitLen() = %d, want 32", ip.BitLen())
	}
	hi, lo := addrParts(ip)
	if hi != 0 || lo != 0xffff01020304 {
		t.Errorf("AddrFrom4() = %x %x, want 0 ffff01020304", hi, lo)
	}
}

func TestAddrFrom16(t *testing.T) {
	// A raw IPv6 address.
	a16 := [16]byte{15: 1}
	ip := netip.AddrFrom16(a16)
	hi, lo := addrParts(ip)
	if hi != 0 || lo != 1 || ip.BitLen() != 128 {
		t.Errorf("AddrFrom16(::1) = %x %x/%d, want 0 1/128", hi, lo, ip.BitLen())
	}

	// A 4-in-6 address stays an IPv6 address.
	a16 = [16]byte{10: 0xff, 11: 0xff, 12: 1, 13: 2, 14: 3, 15: 4}
	ip = netip.AddrFrom16(a16)
	hi, lo = addrParts(ip)
	if hi != 0 || lo != 0xffff01020304 || ip.BitLen() != 128 {
		t.Errorf("AddrFrom16(::ffff:1.2.3.4) = %x %x/%d, want 0 ffff01020304/128", hi, lo, ip.BitLen())
	}
	if ip.Is4() {
		t.Error("AddrFrom16(::ffff:1.2.3.4).Is4() = true, want false")
	}
	if !ip.Is4In6() {
		t.Error("AddrFrom16(::ffff:1.2.3.4).Is4In6() = false, want true")
	}
}

func TestAddrFromSlice(t *testing.T) {
	var buf [16]byte

	// 4 bytes give an IPv4 address.
	b := append(buf[:0], 10, 0, 0, 1)
	ip := netip.AddrFromSlice(b)
	if !ip.Equal(netip.MustParseAddr("10.0.0.1")) {
		t.Error("AddrFromSlice(4 bytes) is not 10.0.0.1")
	}

	// 16 bytes give an IPv6 address.
	b = buf[:16]
	clear(b)
	b[0] = 0xfe
	b[1] = 0x80
	b[15] = 0x01
	ip = netip.AddrFromSlice(b)
	if !ip.Equal(netip.MustParseAddr("fe80::1")) {
		t.Error("AddrFromSlice(16 bytes) is not fe80::1")
	}

	// Any other length gives the zero address.
	ip = netip.AddrFromSlice(buf[:3])
	if ip.IsValid() {
		t.Error("AddrFromSlice(3 bytes) is valid, want the zero address")
	}
	ip = netip.AddrFromSlice(nil)
	if ip.IsValid() {
		t.Error("AddrFromSlice(nil) is valid, want the zero address")
	}
	ip = netip.AddrFromSlice(buf[:0])
	if ip.IsValid() {
		t.Error("AddrFromSlice(0 bytes) is valid, want the zero address")
	}
}

func TestAddrAs4(t *testing.T) {
	var a4 [4]byte
	var a16 [16]byte

	a4 = netip.MustParseAddr("1.2.3.4").As4(a4)
	if a4 != [4]byte{1, 2, 3, 4} {
		t.Error("As4(1.2.3.4) is wrong")
	}

	// A 4-in-6 address gives its IPv4 bytes.
	a4 = netip.MustParseAddr("::ffff:1.2.3.4").As4(a4)
	if a4 != [4]byte{1, 2, 3, 4} {
		t.Error("As4(::ffff:1.2.3.4) is wrong")
	}

	a4 = netip.MustParseAddr("0.0.0.0").As4(a4)
	if a4 != [4]byte{0, 0, 0, 0} {
		t.Error("As4(0.0.0.0) is wrong")
	}

	// As16 keeps every byte of the address.
	a16 = netip.MustParseAddr("fd7a:115c:a1e0:ab12:4843:cd96:626b:430b").As16(a16)
	want := [16]byte{
		0xfd, 0x7a, 0x11, 0x5c, 0xa1, 0xe0, 0xab, 0x12,
		0x48, 0x43, 0xcd, 0x96, 0x62, 0x6b, 0x43, 0x0b,
	}
	if a16 != want {
		t.Error("As16(fd7a:...:430b) is wrong")
	}

	// An IPv4 address gives its 4-in-6 form.
	a16 = netip.MustParseAddr("1.2.3.4").As16(a16)
	want = [16]byte{10: 0xff, 11: 0xff, 12: 1, 13: 2, 14: 3, 15: 4}
	if a16 != want {
		t.Error("As16(1.2.3.4) is wrong")
	}

	// The zero address gives all zeros.
	var zero netip.Addr
	a16 = zero.As16(a16)
	var zero16 [16]byte
	if a16 != zero16 {
		t.Error("As16(Addr{}) is not all zeros")
	}

	// As16 drops the zone.
	a16 = netip.MustParseAddr("::1%10").As16(a16)
	want = [16]byte{15: 1}
	if a16 != want {
		t.Error("As16(::1%10) is wrong")
	}
}

func TestAddrAsSlice(t *testing.T) {
	var buf [16]byte

	got := netip.MustParseAddr("1.2.3.4").AsSlice(buf[:])
	if !bytes.Equal(got, []byte{1, 2, 3, 4}) {
		t.Error("AsSlice(1.2.3.4) is wrong")
	}

	got = netip.MustParseAddr("ffff::1").AsSlice(buf[:])
	want := [16]byte{0: 0xff, 1: 0xff, 15: 1}
	if !bytes.Equal(got, want[:]) {
		t.Error("AsSlice(ffff::1) is wrong")
	}

	var zero netip.Addr
	got = zero.AsSlice(buf[:])
	if got != nil {
		t.Errorf("AsSlice(Addr{}) has %d bytes, want nil", len(got))
	}
}

// An isCase is one case of the Is4, Is6 and Is4In6 methods.
type isCase struct {
	in     string // the address text
	is4    bool   // the Is4 result
	is6    bool   // the Is6 result
	is4in6 bool   // the Is4In6 result
	unmap  string // the Unmap result as text
}

var isCases = []isCase{
	{"1.2.3.4", true, false, false, "1.2.3.4"},
	{"127.0.0.2", true, false, false, "127.0.0.2"},
	{"0.0.0.0", true, false, false, "0.0.0.0"},
	{"::1", false, true, false, "::1"},
	{"::", false, true, false, "::"},
	{"::1%10", false, true, false, "::1%10"},
	{"::ffff:192.0.2.128", false, true, true, "192.0.2.128"},
	{"::ffff:c000:0280", false, true, true, "192.0.2.128"},
	{"::ffff:192.0.2.128%10", false, true, true, "192.0.2.128"},
	{"::ffff:127.1.2.3", false, true, true, "127.1.2.3"},
	{"::ffff:7f01:0203", false, true, true, "127.1.2.3"},
	{"0:0:0:0:0000:ffff:127.1.2.3", false, true, true, "127.1.2.3"},
	{"0:0:0:0::ffff:127.1.2.3", false, true, true, "127.1.2.3"},
	// One bit off the 4-in-6 prefix.
	{"::fffe:c000:0280", false, true, false, "::fffe:c000:280"},
	{"::1:ffff:c000:0280", false, true, false, "::1:ffff:c000:280"},
}

func TestAddrIs(t *testing.T) {
	for _, tc := range isCases {
		ip := netip.MustParseAddr(tc.in)
		if got := ip.Is4(); got != tc.is4 {
			t.Errorf("Is4(%s) = %t, want %t", tc.in, got, tc.is4)
		}
		if got := ip.Is6(); got != tc.is6 {
			t.Errorf("Is6(%s) = %t, want %t", tc.in, got, tc.is6)
		}
		if got := ip.Is4In6(); got != tc.is4in6 {
			t.Errorf("Is4In6(%s) = %t, want %t", tc.in, got, tc.is4in6)
		}

		var buf [bufLen]byte
		u := ip.Unmap()
		if got := u.String(buf[:]); got != tc.unmap {
			t.Errorf("Unmap(%s) = %s, want %s", tc.in, got, tc.unmap)
		}
		// Unmap drops the zone of a 4-in-6 address, and keeps it otherwise.
		if tc.is4in6 && u.Is6() {
			t.Errorf("Unmap(%s).Is6() = true, want false", tc.in)
		}
	}

	// The zero address is neither IPv4 nor IPv6.
	var zero netip.Addr
	if zero.Is4() || zero.Is6() || zero.Is4In6() {
		t.Error("Addr{} reports an address family")
	}
	if zero.Unmap().IsValid() {
		t.Error("Addr{}.Unmap() is valid, want the zero address")
	}
}

func TestAddrBitLen(t *testing.T) {
	var zero netip.Addr
	if got := zero.BitLen(); got != 0 {
		t.Errorf("Addr{}.BitLen() = %d, want 0", got)
	}
	for _, tc := range isCases {
		ip := netip.MustParseAddr(tc.in)
		want := 128
		if tc.is4 {
			want = 32
		}
		if got := ip.BitLen(); got != want {
			t.Errorf("BitLen(%s) = %d, want %d", tc.in, got, want)
		}
	}
}

func TestAddrZone(t *testing.T) {
	var zoneBuf [netip.MaxZoneLen]byte

	// An IPv6 address takes a zone.
	ip := netip.MustParseAddr("fe80::1")
	if zoneOf(ip, zoneBuf[:]) != "" {
		t.Error("an address without a zone reports one")
	}
	z := ip.WithZone("42")
	if got := zoneOf(z, zoneBuf[:]); got != "42" {
		t.Errorf("WithZone(42).Zone() = %s, want 42", got)
	}
	if !z.Equal(netip.MustParseAddr("fe80::1%42")) {
		t.Error("WithZone(42) != ParseAddr(fe80::1%42)")
	}

	// An empty zone removes the zone.
	if got := zoneOf(z.WithZone(""), zoneBuf[:]); got != "" {
		t.Errorf("WithZone(\"\").Zone() = %s, want an empty zone", got)
	}
	if !z.WithZone("").Equal(ip) {
		t.Error("WithZone(\"\") does not give the address back")
	}

	// A zone of zero is no zone.
	if got := zoneOf(ip.WithZone("0"), zoneBuf[:]); got != "" {
		t.Errorf("WithZone(0).Zone() = %s, want an empty zone", got)
	}

	// An IPv4 address takes no zone.
	ip4 := netip.MustParseAddr("1.2.3.4")
	if got := zoneOf(ip4.WithZone("42"), zoneBuf[:]); got != "" {
		t.Errorf("IPv4 WithZone(42).Zone() = %s, want an empty zone", got)
	}

	// The zero address takes no zone.
	var zero netip.Addr
	if zero.WithZone("42").IsValid() {
		t.Error("Addr{}.WithZone(42) is valid, want the zero address")
	}

	// A zone is part of the address.
	if z.Equal(ip) {
		t.Error("an address with a zone equals the same address without one")
	}
	if netip.MustParseAddr("::1%1").Equal(netip.MustParseAddr("::1%2")) {
		t.Error("two addresses with different zones are equal")
	}
}

// A nextCase is one case of the Next and Prev methods.
// An empty want means the zero address.
type nextCase struct {
	in   string // the address text
	next string // the Next result as text
	prev string // the Prev result as text
}

var nextCases = []nextCase{
	{"10.0.0.1", "10.0.0.2", "10.0.0.0"},
	{"10.0.0.255", "10.0.1.0", "10.0.0.254"},
	{"127.0.0.1", "127.0.0.2", "127.0.0.0"},
	{"254.255.255.255", "255.0.0.0", "254.255.255.254"},
	{"255.255.255.255", "", "255.255.255.254"},
	{"0.0.0.0", "0.0.0.1", ""},
	{"::", "::1", ""},
	{"::%10", "::1%10", ""},
	{"::1", "::2", "::"},
	{"::ffff:ffff:ffff:ffff", "0:0:0:1::", "::ffff:ffff:ffff:fffe"},
	{"0:0:0:1::", "0:0:0:1::1", "::ffff:ffff:ffff:ffff"},
	{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", "", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"},
	// A 4-in-6 address counts as an IPv6 address.
	{"::ffff:255.255.255.255", "::1:0:0:0", "::ffff:255.255.255.254"},
}

func TestAddrNextPrev(t *testing.T) {
	for _, tc := range nextCases {
		ip := netip.MustParseAddr(tc.in)

		var next netip.Addr
		if tc.next != "" {
			next = netip.MustParseAddr(tc.next)
		}
		if got := ip.Next(); !got.Equal(next) {
			var buf [bufLen]byte
			t.Errorf("Next(%s) = %s, want %s", tc.in, got.String(buf[:]), tc.next)
		}

		var prev netip.Addr
		if tc.prev != "" {
			prev = netip.MustParseAddr(tc.prev)
		}
		if got := ip.Prev(); !got.Equal(prev) {
			var buf [bufLen]byte
			t.Errorf("Prev(%s) = %s, want %s", tc.in, got.String(buf[:]), tc.prev)
		}

		// Next and Prev undo each other.
		if next.IsValid() && !next.Prev().Equal(ip) {
			t.Errorf("Next(%s).Prev() is not %s", tc.in, tc.in)
		}
		if prev.IsValid() && !prev.Next().Equal(ip) {
			t.Errorf("Prev(%s).Next() is not %s", tc.in, tc.in)
		}
	}

	// The zero address stays zero.
	var zero netip.Addr
	if zero.Next().IsValid() || zero.Prev().IsValid() {
		t.Error("Addr{}.Next() or Addr{}.Prev() is valid")
	}
}

// A compareCase is one case of the Compare and Less methods.
// An empty address text means the zero address.
type compareCase struct {
	a    string // the first address
	b    string // the second address
	want int    // the Compare result
}

var compareCases = []compareCase{
	{"", "", 0},
	{"", "1.2.3.4", -1},
	{"1.2.3.4", "", 1},

	// A shorter address sorts first.
	{"1.2.3.4", "0102:0304::0", -1},
	{"0102:0304::0", "1.2.3.4", 1},
	{"0.0.0.0", "::", -1},
	{"::", "0.0.0.0", 1},

	{"1.2.3.4", "1.2.3.4", 0},
	{"1.2.3.4", "1.2.3.5", -1},
	{"1.2.3.5", "1.2.3.4", 1},
	{"0.0.0.0", "255.255.255.255", -1},

	{"::1", "::2", -1},
	{"::2", "::3", -1},
	{"::", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", -1},
	{"1::", "2::", -1},

	// A zone sorts after the same address without one.
	{"::1", "::1%1", -1},
	{"::1%1", "::2", -1},
	{"::1%1", "::1%2", -1},
	{"::1%1", "::1%1", 0},
	{"::1%2", "::1%1", 1},

	// A 4-in-6 address is not the IPv4 address.
	{"::ffff:11.1.1.12", "11.1.1.12", 1},
}

func TestAddrCompare(t *testing.T) {
	for _, tc := range compareCases {
		var a, b netip.Addr
		if tc.a != "" {
			a = netip.MustParseAddr(tc.a)
		}
		if tc.b != "" {
			b = netip.MustParseAddr(tc.b)
		}

		got := a.Compare(b)
		if got != tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
		// Compare is antisymmetric.
		if back := b.Compare(a); back != -tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.b, tc.a, back, -tc.want)
		}
		// Less agrees with Compare.
		if less := a.Less(b); less != (tc.want == -1) {
			t.Errorf("Less(%s, %s) = %t, want %t", tc.a, tc.b, less, tc.want == -1)
		}
		if less := b.Less(a); less != (tc.want == 1) {
			t.Errorf("Less(%s, %s) = %t, want %t", tc.b, tc.a, less, tc.want == 1)
		}
		// Equal agrees with Compare.
		if eq := a.Equal(b); eq != (tc.want == 0) {
			t.Errorf("Equal(%s, %s) = %t, want %t", tc.a, tc.b, eq, tc.want == 0)
		}
	}
}

func TestAddrSort(t *testing.T) {
	var zero netip.Addr
	values := []netip.Addr{
		netip.MustParseAddr("::1"),
		netip.MustParseAddr("::2"),
		zero,
		netip.MustParseAddr("1.2.3.4"),
		netip.MustParseAddr("8.8.8.8"),
		netip.MustParseAddr("::1%42"),
	}
	want := []string{"invalid IP", "1.2.3.4", "8.8.8.8", "::1", "::1%42", "::2"}

	slices.SortFunc(values, compareAddrs)
	for i, v := range values {
		var buf [bufLen]byte
		if got := v.String(buf[:]); got != want[i] {
			t.Errorf("sorted[%d] = %s, want %s", i, got, want[i])
		}
	}
}

func TestAddrWellKnown(t *testing.T) {
	var buf [bufLen]byte

	if got := netip.IPv4Unspecified().String(buf[:]); got != "0.0.0.0" {
		t.Errorf("IPv4Unspecified() = %s, want 0.0.0.0", got)
	}
	if got := netip.IPv6Unspecified().String(buf[:]); got != "::" {
		t.Errorf("IPv6Unspecified() = %s, want ::", got)
	}
	if got := netip.IPv6Loopback().String(buf[:]); got != "::1" {
		t.Errorf("IPv6Loopback() = %s, want ::1", got)
	}
	if got := netip.IPv6LinkLocalAllNodes().String(buf[:]); got != "ff02::1" {
		t.Errorf("IPv6LinkLocalAllNodes() = %s, want ff02::1", got)
	}
	if got := netip.IPv6LinkLocalAllRouters().String(buf[:]); got != "ff02::2" {
		t.Errorf("IPv6LinkLocalAllRouters() = %s, want ff02::2", got)
	}

	if !netip.IPv4Unspecified().Is4() {
		t.Error("IPv4Unspecified() is not an IPv4 address")
	}
	if !netip.IPv6Unspecified().Is6() {
		t.Error("IPv6Unspecified() is not an IPv6 address")
	}
	if !netip.IPv4Unspecified().IsUnspecified() || !netip.IPv6Unspecified().IsUnspecified() {
		t.Error("an unspecified address does not report itself")
	}
	if !netip.IPv6Loopback().IsLoopback() {
		t.Error("IPv6Loopback() is not a loopback address")
	}
}

func TestAddrMaxLen(t *testing.T) {
	// Write the longest address text into a buffer of the documented size.
	// The longest IPv4 text.
	var buf4 [netip.MaxAddr4Len]byte
	ip4 := netip.MustParseAddr("255.255.255.255")
	if got := ip4.String(buf4[:]); got != "255.255.255.255" {
		t.Errorf("String() = %s, want 255.255.255.255", got)
	}

	// The longest 4-in-6 text.
	var buf4In6 [netip.MaxAddr4In6Len]byte
	ip4In6 := netip.MustParseAddr("::ffff:255.255.255.255").WithZone("4294967295")
	want4In6 := "::ffff:255.255.255.255%4294967295"
	if got := ip4In6.String(buf4In6[:]); got != want4In6 {
		t.Errorf("String() = %s, want %s", got, want4In6)
	}

	// The longest IPv6 text.
	var buf6 [netip.MaxAddr6Len]byte
	ip6 := netip.MustParseAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff").WithZone("4294967295")
	want6 := "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%4294967295"
	if got := ip6.String(buf6[:]); got != want6 {
		t.Errorf("String() = %s, want %s", got, want6)
	}

	// The longest zone text.
	var bufZone [netip.MaxZoneLen]byte
	if got := ip6.Zone(bufZone[:]); got != "4294967295" {
		t.Errorf("Zone() = %s, want 4294967295", got)
	}

	// AppendText needs the same room.
	var bufAppend [netip.MaxAddrLen]byte
	b, err := ip6.AppendText(bufAppend[:0])
	if err != nil {
		t.Errorf("AppendText() failed: %s", err.Error())
		return
	}
	if got := string(b); got != want6 {
		t.Errorf("AppendText() = %s, want %s", got, want6)
	}
}
