package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
)

func TestParseAddr_IPv4(t *testing.T) {
	ip4, err := netip.ParseAddr("192.168.140.255")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var a4 [4]byte
	a4 = ip4.As4(a4)
	if a4 != [4]byte{192, 168, 140, 255} {
		t.Error("unexpected IPv4 bytes")
	}
}

func TestParseAddr_IPv6(t *testing.T) {
	ip6, err := netip.ParseAddr("fd7a:115c::626b:430b")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var a16 [16]byte
	a16 = ip6.As16(a16)
	if a16 != [16]byte{0xfd, 0x7a, 0x11, 0x5c, 12: 0x62, 0x6b, 0x43, 0x0b} {
		t.Error("unexpected IPv6 bytes")
	}
}

func TestAddr_String(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ip := netip.MustParseAddr("10.0.0.1")
	if ip.String(buf[:]) != "10.0.0.1" {
		t.Error("Addr.String IPv4")
	}
	ip = netip.MustParseAddr("2001:db8::1")
	if ip.String(buf[:]) != "2001:db8::1" {
		t.Error("Addr.String IPv6")
	}
}

func TestAddr_Is(t *testing.T) {
	ip4 := netip.MustParseAddr("1.2.3.4")
	if !ip4.Is4() {
		t.Error("Is4")
	}
	if ip4.Is6() {
		t.Error("Is6 for v4")
	}
	ip6 := netip.MustParseAddr("::1")
	if ip6.Is4() {
		t.Error("Is4 for v6")
	}
	if !ip6.Is6() {
		t.Error("Is6")
	}
}

func TestAddr_Props(t *testing.T) {
	if !netip.MustParseAddr("127.0.0.1").IsLoopback() {
		t.Error("IsLoopback v4")
	}
	if !netip.MustParseAddr("::1").IsLoopback() {
		t.Error("IsLoopback v6")
	}
	if !netip.MustParseAddr("10.0.0.1").IsPrivate() {
		t.Error("IsPrivate")
	}
	if !netip.MustParseAddr("224.0.0.1").IsMulticast() {
		t.Error("IsMulticast")
	}
}

func TestAddr_Compare(t *testing.T) {
	a := netip.MustParseAddr("1.2.3.4")
	b := netip.MustParseAddr("1.2.3.5")
	if a.Compare(b) != -1 {
		t.Error("Compare less")
	}
	if b.Compare(a) != 1 {
		t.Error("Compare greater")
	}
	if a.Compare(a) != 0 {
		t.Error("Compare equal")
	}
}

func TestAddr_NextPrev(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ip := netip.MustParseAddr("1.2.3.4")
	next := ip.Next()
	if next.String(buf[:]) != "1.2.3.5" {
		t.Error("Addr.Next")
	}
	prev := next.Prev()
	if !prev.Equal(ip) {
		t.Error("Addr.Prev")
	}
}

func TestAddr_Unmap(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ip := netip.MustParseAddr("::ffff:1.2.3.4")
	if !ip.Is4In6() {
		t.Error("Is4In6")
	}
	unmapped := ip.Unmap()
	if !unmapped.Is4() {
		t.Error("Unmap Is4")
	}
	if unmapped.String(buf[:]) != "1.2.3.4" {
		t.Error("Unmap String")
	}
}

func TestAddrFrom(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ip4 := netip.AddrFrom4([4]byte{10, 20, 30, 40})
	if ip4.String(buf[:]) != "10.20.30.40" {
		t.Error("AddrFrom4")
	}
	ip6 := netip.AddrFrom16([16]byte{0x20, 0x01, 0x0d, 0xb8, 15: 0x01})
	if ip6.String(buf[:]) != "2001:db8::1" {
		t.Error("AddrFrom16")
	}
}
