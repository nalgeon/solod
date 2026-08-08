package netip_bench

import (
	"testing"

	"solod.dev/so/net/netip"
)

func BenchmarkParse_v4_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v4)
	}
}

func BenchmarkParse_v6_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6)
	}
}

func BenchmarkParse_v6e_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6e)
	}
}

func BenchmarkParse_v6_v4_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6_v4)
	}
}

func BenchmarkParse_v6_zone_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkIP, _ = netip.ParseAddr(v6_zone)
	}
}

func BenchmarkString_v4_Go(b *testing.B) {
	ip := netip.MustParseAddr(v4)
	buf := make([]byte, netip.MaxAddrLen)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_Go(b *testing.B) {
	ip := netip.MustParseAddr(v6)
	buf := make([]byte, netip.MaxAddrLen)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6e_Go(b *testing.B) {
	ip := netip.MustParseAddr(v6e)
	buf := make([]byte, netip.MaxAddrLen)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_v4_Go(b *testing.B) {
	ip := netip.MustParseAddr(v6_v4)
	buf := make([]byte, netip.MaxAddrLen)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}

func BenchmarkString_v6_zone_Go(b *testing.B) {
	ip := netip.MustParseAddr(v6_zone)
	buf := make([]byte, netip.MaxAddrLen)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = ip.String(buf)
	}
}
