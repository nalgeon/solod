// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/encoding/binary"
	"solod.dev/so/net/netip"
)

// bufLen is the length of a text buffer. It holds the longest address text
// with room to spare.
const bufLen = 64

// addrParts returns the address as two 64-bit halves, the way the package
// stores it. The high half holds the first eight bytes.
func addrParts(ip netip.Addr) (uint64, uint64) {
	var a16 [16]byte
	a16 = ip.As16(a16)
	return binary.BigEndian.Uint64(a16[:8]), binary.BigEndian.Uint64(a16[8:])
}

// zoneOf returns the zone of ip as text. The text is a view into buf,
// so buf must outlive the result.
func zoneOf(ip netip.Addr, buf []byte) string {
	return ip.Zone(buf)
}

// wantErr maps an error code of a test table to the error itself.
// It returns nil for errAny.
func wantErr(code int) error {
	switch code {
	case errIP:
		return netip.ErrIP
	case errIPv4:
		return netip.ErrIPv4
	case errIPv6:
		return netip.ErrIPv6
	case errIPPort:
		return netip.ErrIPPort
	case errPort:
		return netip.ErrPort
	case errPrefix:
		return netip.ErrPrefix
	}
	return nil
}

// compareAddrs compares two addresses.
func compareAddrs(a, b any) int {
	pa := a.(*netip.Addr)
	pb := b.(*netip.Addr)
	return pa.Compare(*pb)
}

// compareAddrPorts compares two address-port pairs.
func compareAddrPorts(a, b any) int {
	pa := a.(*netip.AddrPort)
	pb := b.(*netip.AddrPort)
	return pa.Compare(*pb)
}

// comparePrefixes compares two prefixes.
func comparePrefixes(a, b any) int {
	pa := a.(*netip.Prefix)
	pb := b.(*netip.Prefix)
	return pa.Compare(*pb)
}
