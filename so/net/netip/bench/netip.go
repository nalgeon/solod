package netip_bench

import (
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
)

//so:volatile
var sinkIP netip.Addr

//so:volatile
var sinkStr string

const (
	v4      = "192.168.1.1"
	v6      = "fd7a:115c:a1e0:ab12:4843:cd96:626b:430b"
	v6e     = "fd7a:115c::626b:430b"
	v6_v4   = "::ffff:192.168.140.255"
	v6_zone = "1:2::ffff:192.168.140.255%eth1"
)

func BenchmarkParse_v4_So(b *testing.B) {
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v4)
	}
}

func BenchmarkParse_v6_So(b *testing.B) {
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6)
	}
}

func BenchmarkParse_v6e_So(b *testing.B) {
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6e)
	}
}

func BenchmarkParse_v6_v4_So(b *testing.B) {
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6_v4)
	}
}

func BenchmarkParse_v6_zone_So(b *testing.B) {
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6_zone)
	}
}

func BenchmarkString_v4_So(b *testing.B) {
	ip := netip.MustParseAddr(v4)
	buf := make([]byte, netip.MaxAddrLen)
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_So(b *testing.B) {
	ip := netip.MustParseAddr(v6)
	buf := make([]byte, netip.MaxAddrLen)
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6e_So(b *testing.B) {
	ip := netip.MustParseAddr(v6e)
	buf := make([]byte, netip.MaxAddrLen)
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_v4_So(b *testing.B) {
	ip := netip.MustParseAddr(v6_v4)
	buf := make([]byte, netip.MaxAddrLen)
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_zone_So(b *testing.B) {
	ip := netip.MustParseAddr(v6_zone)
	buf := make([]byte, netip.MaxAddrLen)
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}
