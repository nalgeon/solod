package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
)

func TestPrefix(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	pfx, err := netip.ParsePrefix("192.168.1.0/24")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if pfx.Bits() != 24 {
		t.Error("Prefix.Bits")
	}
	if pfx.String(buf[:]) != "192.168.1.0/24" {
		t.Error("Prefix.String")
	}
	if !pfx.Contains(netip.MustParseAddr("192.168.1.100")) {
		t.Error("Prefix.Contains true")
	}
	if pfx.Contains(netip.MustParseAddr("192.168.2.1")) {
		t.Error("Prefix.Contains false")
	}
}

func TestPrefix_Masked(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	pfx := netip.MustParsePrefix("192.168.1.1/24")
	masked := pfx.Masked()
	maskedAddr := masked.Addr()
	if maskedAddr.String(buf[:]) != "192.168.1.0" {
		t.Error("Prefix.Masked")
	}
}

func TestPrefix_Overlaps(t *testing.T) {
	a := netip.MustParsePrefix("192.168.0.0/16")
	b := netip.MustParsePrefix("192.168.1.0/24")
	if !a.Overlaps(b) {
		t.Error("Prefix.Overlaps true")
	}
	c := netip.MustParsePrefix("10.0.0.0/8")
	if a.Overlaps(c) {
		t.Error("Prefix.Overlaps false")
	}
}
