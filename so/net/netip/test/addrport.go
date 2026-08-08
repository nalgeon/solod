package netip_test

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
)

func TestAddrPort_IPv4(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ap, err := netip.ParseAddrPort("192.168.1.1:8080")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	addr := ap.Addr()
	if addr.String(buf[:]) != "192.168.1.1" {
		t.Error("AddrPort.Addr")
	}
	if ap.Port() != 8080 {
		t.Error("AddrPort.Port")
	}
	if ap.String(buf[:]) != "192.168.1.1:8080" {
		t.Error("AddrPort.String v4")
	}
}

func TestAddrPort_IPv6(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte
	ap := netip.MustParseAddrPort("[::1]:443")
	if ap.String(buf[:]) != "[::1]:443" {
		t.Error("AddrPort.String v6")
	}
}
