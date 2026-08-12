// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import (
	"solod.dev/so/bytealg"
	"solod.dev/so/encoding/binary"
	"solod.dev/so/strconv"
)

// Maximum length of an IP address string.
const (
	MaxZoneLen     = 10 // 4294967295
	MaxAddr4Len    = 15 // 255.255.255.255
	MaxAddr4In6Len = 29 // ::ffff:255.255.255.255%enp5s0
	MaxAddr6Len    = 46 // ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0
	MaxAddrLen     = MaxAddr6Len
)

// Largest value of an IPv6 address group.
const maxUint16 = 0xffff

// Addr represents an IPv4 or IPv6 address (with or without
// a scoped addressing zone).
//
// The zero Addr is not a valid IP address.
// Addr{} is distinct from both 0.0.0.0 and ::.
type Addr struct {
	// addr is the hi and lo bits of an IPv6 address. If bitlen==z4,
	// hi and lo contain the IPv4-mapped IPv6 address.
	//
	// hi and lo are constructed by interpreting a 16-byte IPv6
	// address as a big-endian 128-bit number. The most significant
	// bits of that number go into hi, the rest into lo.
	//
	// For example, 0011:2233:4455:6677:8899:aabb:ccdd:eeff is stored as:
	//  addr.hi = 0x0011223344556677
	//  addr.lo = 0x8899aabbccddeeff
	//
	// We store IPs like this, rather than as [16]byte, because it
	// turns most operations on IPs into arithmetic and bit-twiddling
	// operations on 64-bit registers, which is much faster than
	// bytewise processing.
	addr uint128

	bitlen  uint8  // 0 for zero Addr, 32 for IPv4, 128 for IPv6
	scopeID uint32 // optional IPv6 zone index, 0 if not set
}

const (
	z0 uint8 = 0   // zero Addr
	z4 uint8 = 32  // IPv4 Addr
	z6 uint8 = 128 // IPv6 Addr
)

// IPv6LinkLocalAllNodes returns the IPv6 link-local all nodes multicast
// address ff02::1.
func IPv6LinkLocalAllNodes() Addr {
	a16 := [16]byte{0: 0xff, 1: 0x02, 15: 0x01}
	return AddrFrom16(a16)
}

// IPv6LinkLocalAllRouters returns the IPv6 link-local all routers multicast
// address ff02::2.
func IPv6LinkLocalAllRouters() Addr {
	a16 := [16]byte{0: 0xff, 1: 0x02, 15: 0x02}
	return AddrFrom16(a16)
}

// IPv6Loopback returns the IPv6 loopback address ::1.
func IPv6Loopback() Addr {
	a16 := [16]byte{15: 0x01}
	return AddrFrom16(a16)
}

// IPv6Unspecified returns the IPv6 unspecified address "::".
func IPv6Unspecified() Addr {
	return Addr{bitlen: z6}
}

// IPv4Unspecified returns the IPv4 unspecified address "0.0.0.0".
func IPv4Unspecified() Addr {
	var a4 [4]byte
	return AddrFrom4(a4)
}

// AddrFrom4 returns the address of the IPv4 address given by the bytes in addr.
func AddrFrom4(addr [4]byte) Addr {
	lo := 0xffff00000000 | uint64(addr[0])<<24 | uint64(addr[1])<<16 | uint64(addr[2])<<8 | uint64(addr[3])
	return Addr{addr: uint128{0, lo}, bitlen: z4}
}

// AddrFrom16 returns the IPv6 address given by the bytes in addr.
// An IPv4-mapped IPv6 address is left as an IPv6 address.
// (Use Unmap to convert them if needed.)
func AddrFrom16(addr [16]byte) Addr {
	return Addr{
		addr: uint128{
			binary.BigEndian.Uint64(addr[:8]),
			binary.BigEndian.Uint64(addr[8:]),
		},
		bitlen: z6,
	}
}

// ParseAddr parses s as an IP address, returning the result. The string
// s can be in dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"),
// or IPv6 with a scoped addressing zone ("fe80::1cc0:3e8c:119f:c2e1%ens18").
func ParseAddr(s string) (Addr, error) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return parseIPv4(s)
		case ':':
			return parseIPv6(s)
		case '%':
			// Assume that this was trying to be an IPv6 address with
			// a zone specifier, but the address is missing.
			return Addr{}, ErrIPv6
		}
	}
	return Addr{}, ErrIP
}

// MustParseAddr calls [ParseAddr](s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParseAddr(s string) Addr {
	ip, err := ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return ip
}

func parseIPv4Fields(in string, off, end int, fields []uint8) error {
	var val, pos int
	var digLen int // number of digits in current octet
	s := in[off:end]
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			if digLen == 1 && val == 0 {
				return ErrIPv4
			}
			val = val*10 + int(s[i]) - '0'
			digLen++
			if val > 255 {
				return ErrIPv4
			}
		} else if s[i] == '.' {
			// .1.2.3
			// 1.2.3.
			// 1..2.3
			if i == 0 || i == len(s)-1 || s[i-1] == '.' {
				return ErrIPv4
			}
			// 1.2.3.4.5
			if pos == 3 {
				return ErrIPv4
			}
			fields[pos] = uint8(val)
			pos++
			val = 0
			digLen = 0
		} else {
			return ErrIPv4
		}
	}
	if pos < 3 {
		return ErrIPv4
	}
	fields[3] = uint8(val)
	return nil
}

// parseIPv4 parses s as an IPv4 address (in form "192.168.0.1").
func parseIPv4(s string) (Addr, error) {
	var fields [4]uint8
	err := parseIPv4Fields(s, 0, len(s), fields[:])
	if err != nil {
		return Addr{}, err
	}
	return AddrFrom4(fields), nil
}

// parseIPv6 parses s as an IPv6 address (in form "2001:db8::68").
func parseIPv6(in string) (Addr, error) {
	s := in

	// Split off the zone right from the start. Yes it's a second scan
	// of the string, but trying to handle it inline makes a bunch of
	// other inner loop conditionals more expensive, and it ends up
	// being slower.
	zone := ""
	i := bytealg.IndexByteString(s, '%')
	if i != -1 {
		zone = s[i+1:]
		s = s[:i]
		if zone == "" {
			// Not allowed to have an empty zone if explicitly specified.
			return Addr{}, ErrIPv6
		}
	}

	var ip [16]byte
	ellipsis := -1 // position of ellipsis in ip

	// Might have leading ellipsis
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// Might be only ellipsis
		if len(s) == 0 {
			return IPv6Unspecified().WithZone(zone), nil
		}
	}

	// Loop, parsing hex numbers followed by colon.
	i = 0
	for i < 16 {
		// Hex number. Similar to parseIPv4, inlining the hex number
		// parsing yields a significant performance increase.
		off := 0
		acc := uint32(0)
		for ; off < len(s); off++ {
			c := s[off]
			if c >= '0' && c <= '9' {
				acc = (acc << 4) + uint32(c-'0')
			} else if c >= 'a' && c <= 'f' {
				acc = (acc << 4) + uint32(c-'a'+10)
			} else if c >= 'A' && c <= 'F' {
				acc = (acc << 4) + uint32(c-'A'+10)
			} else {
				break
			}
			if off > 3 {
				//more than 4 digits in group, fail.
				return Addr{}, ErrIPv6
			}
			if acc > maxUint16 {
				// Overflow, fail.
				return Addr{}, ErrIPv6
			}
		}
		if off == 0 {
			// No digits found, fail.
			return Addr{}, ErrIPv6
		}

		// If followed by dot, might be in trailing IPv4.
		if off < len(s) && s[off] == '.' {
			if ellipsis < 0 && i != 12 {
				// Not the right place.
				return Addr{}, ErrIPv6
			}
			if i+4 > 16 {
				// Not enough room.
				return Addr{}, ErrIPv6
			}

			end := len(in)
			if len(zone) > 0 {
				end -= len(zone) + 1
			}
			err := parseIPv4Fields(in, end-len(s), end, ip[i:i+4])
			if err != nil {
				return Addr{}, err
			}
			s = ""
			i += 4
			break
		}

		// Save this 16-bit chunk.
		ip[i] = byte(acc >> 8)
		ip[i+1] = byte(acc)
		i += 2

		// Stop at end of string.
		s = s[off:]
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' {
			return Addr{}, ErrIPv6
		} else if len(s) == 1 {
			return Addr{}, ErrIPv6
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return Addr{}, ErrIPv6
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return Addr{}, ErrIPv6
	}

	// If didn't parse enough, expand ellipsis.
	if i < 16 {
		if ellipsis < 0 {
			return Addr{}, ErrIPv6
		}
		n := 16 - i
		for j := i - 1; j >= ellipsis; j-- {
			ip[j+n] = ip[j]
		}
		clear(ip[ellipsis : ellipsis+n])
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return Addr{}, ErrIPv6
	}
	return AddrFrom16(ip).WithZone(zone), nil
}

// AddrFromSlice parses the 4- or 16-byte byte slice as an IPv4 or IPv6 address.
// If slice's length is not 4 or 16, returns a zero Addr.
func AddrFromSlice(slice []byte) Addr {
	switch len(slice) {
	case 4:
		a4 := [4]byte(slice)
		return AddrFrom4(a4)
	case 16:
		a16 := [16]byte(slice)
		return AddrFrom16(a16)
	}
	return Addr{}
}

// v4 returns the i'th byte of ip. If ip is not an IPv4, v4 returns
// unspecified garbage.
func (ip Addr) v4(i uint8) uint8 {
	return uint8(ip.addr.lo >> ((3 - i) * 8))
}

// v6 returns the i'th byte of ip. If ip is an IPv4 address, this
// accesses the IPv4-mapped IPv6 address form of the IP.
func (ip Addr) v6(i uint8) uint8 {
	halves := [2]*uint64{&ip.addr.hi, &ip.addr.lo}
	return uint8(*(halves[(i/8)%2]) >> ((7 - i%8) * 8))
}

// v6u16 returns the i'th 16-bit word of ip. If ip is an IPv4 address,
// this accesses the IPv4-mapped IPv6 address form of the IP.
func (ip Addr) v6u16(i uint8) uint16 {
	halves := [2]*uint64{&ip.addr.hi, &ip.addr.lo}
	return uint16(*(halves[(i/4)%2]) >> ((3 - i%4) * 16))
}

// isZero reports whether ip is the zero value of the IP type.
// The zero value is not a valid IP address of any type.
//
// Note that "0.0.0.0" and "::" are not the zero value. Use IsUnspecified to
// check for these values instead.
func (ip Addr) isZero() bool {
	return ip.bitlen == z0
}

// IsValid reports whether the [Addr] is an initialized address (not the zero Addr).
//
// Note that "0.0.0.0" and "::" are both valid values.
func (ip Addr) IsValid() bool { return ip.bitlen != z0 }

// BitLen returns the number of bits in the IP address:
// 128 for IPv6, 32 for IPv4, and 0 for the zero [Addr].
//
// Note that IPv4-mapped IPv6 addresses are considered IPv6 addresses
// and therefore have bit length 128.
func (ip Addr) BitLen() int {
	return int(ip.bitlen)
}

// Zone returns ip's IPv6 scoped addressing zone, if any.
// buf length must be at least [MaxZoneLen].
func (ip Addr) Zone(buf []byte) string {
	if ip.scopeID == 0 {
		return ""
	}
	return strconv.FormatUint(buf, uint64(ip.scopeID), 10)
}

// Equal reports whether ip and ip2 are the same IP address.
func (ip Addr) Equal(ip2 Addr) bool {
	return ip.bitlen == ip2.bitlen && ip.addr.equal(ip2.addr) && ip.scopeID == ip2.scopeID
}

// Compare returns an integer comparing two IPs.
// The result will be 0 if ip == ip2, -1 if ip < ip2, and +1 if ip > ip2.
// The definition of "less than" is the same as the [Addr.Less] method.
func (ip Addr) Compare(ip2 Addr) int {
	f1, f2 := ip.BitLen(), ip2.BitLen()
	if f1 < f2 {
		return -1
	}
	if f1 > f2 {
		return 1
	}
	hi1, hi2 := ip.addr.hi, ip2.addr.hi
	if hi1 < hi2 {
		return -1
	}
	if hi1 > hi2 {
		return 1
	}
	lo1, lo2 := ip.addr.lo, ip2.addr.lo
	if lo1 < lo2 {
		return -1
	}
	if lo1 > lo2 {
		return 1
	}
	if ip.Is6() {
		za, zb := ip.scopeID, ip2.scopeID
		if za < zb {
			return -1
		}
		if za > zb {
			return 1
		}
	}
	return 0
}

// Less reports whether ip sorts before ip2.
// IP addresses sort first by length, then their address.
// IPv6 addresses with zones sort just after the same address without a zone.
func (ip Addr) Less(ip2 Addr) bool { return ip.Compare(ip2) == -1 }

// Is4 reports whether ip is an IPv4 address.
//
// It returns false for IPv4-mapped IPv6 addresses. See [Addr.Unmap].
func (ip Addr) Is4() bool {
	return ip.bitlen == z4
}

// Is4In6 reports whether ip is an "IPv4-mapped IPv6 address"
// as defined by RFC 4291.
// That is, it reports whether ip is in ::ffff:0:0/96.
func (ip Addr) Is4In6() bool {
	return ip.Is6() && ip.addr.hi == 0 && ip.addr.lo>>32 == 0xffff
}

// Is6 reports whether ip is an IPv6 address, including IPv4-mapped
// IPv6 addresses.
func (ip Addr) Is6() bool {
	return ip.bitlen != z0 && ip.bitlen != z4
}

// Unmap returns ip with any IPv4-mapped IPv6 address prefix removed.
//
// That is, if ip is an IPv6 address wrapping an IPv4 address, it
// returns the wrapped IPv4 address. Otherwise it returns ip unmodified.
func (ip Addr) Unmap() Addr {
	if ip.Is4In6() {
		ip.bitlen = z4
		ip.scopeID = 0
	}
	return ip
}

// WithZone returns an IP that's the same as ip but with the provided
// zone. If zone is empty, the zone is removed. If ip is an IPv4
// address, WithZone is a no-op and returns ip unchanged.
func (ip Addr) WithZone(zone string) Addr {
	if !ip.Is6() {
		return ip
	}
	if zone == "" {
		ip.bitlen = z6
		return ip
	}
	// Try parsing zone as a number first.
	scopeID, err := strconv.Atoi(zone)
	if err == nil {
		ip.scopeID = uint32(scopeID)
		return ip
	}
	// Not a number, so must be an interface name.
	ip.scopeID = uint32(if_nametoindex(zone))
	return ip
}

// withoutZone unconditionally strips the zone from ip.
// It's similar to WithZone, but small enough to be inlinable.
func (ip Addr) withoutZone() Addr {
	if !ip.Is6() {
		return ip
	}
	ip.bitlen = z6
	ip.scopeID = 0
	return ip
}

// hasZone reports whether ip has an IPv6 zone.
func (ip Addr) hasZone() bool {
	return ip.scopeID != 0
}

// IsLinkLocalUnicast reports whether ip is a link-local unicast address.
func (ip Addr) IsLinkLocalUnicast() bool {
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// Dynamic Configuration of IPv4 Link-Local Addresses
	// https://datatracker.ietf.org/doc/html/rfc3927#section-2.1
	if ip.Is4() {
		return ip.v4(0) == 169 && ip.v4(1) == 254
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.v6u16(0)&0xffc0 == 0xfe80
	}
	return false // zero value
}

// IsLoopback reports whether ip is a loopback address.
func (ip Addr) IsLoopback() bool {
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// Requirements for Internet Hosts -- Communication Layers (3.2.1.3 Addressing)
	// https://datatracker.ietf.org/doc/html/rfc1122#section-3.2.1.3
	if ip.Is4() {
		return ip.v4(0) == 127
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.addr.hi == 0 && ip.addr.lo == 1
	}
	return false // zero value
}

// IsMulticast reports whether ip is a multicast address.
func (ip Addr) IsMulticast() bool {
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// Host Extensions for IP Multicasting (4. HOST GROUP ADDRESSES)
	// https://datatracker.ietf.org/doc/html/rfc1112#section-4
	if ip.Is4() {
		return ip.v4(0)&0xf0 == 0xe0
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.addr.hi>>(64-8) == 0xff // ip.v6(0) == 0xff
	}
	return false // zero value
}

// IsInterfaceLocalMulticast reports whether ip is an IPv6 interface-local
// multicast address.
func (ip Addr) IsInterfaceLocalMulticast() bool {
	// IPv6 Addressing Architecture (2.7.1. Pre-Defined Multicast Addresses)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.7.1
	if ip.Is6() && !ip.Is4In6() {
		return ip.v6u16(0)&0xff0f == 0xff01
	}
	return false // zero value
}

// IsLinkLocalMulticast reports whether ip is a link-local multicast address.
func (ip Addr) IsLinkLocalMulticast() bool {
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// IPv4 Multicast Guidelines (4. Local Network Control Block (224.0.0/24))
	// https://datatracker.ietf.org/doc/html/rfc5771#section-4
	if ip.Is4() {
		return ip.v4(0) == 224 && ip.v4(1) == 0 && ip.v4(2) == 0
	}
	// IPv6 Addressing Architecture (2.7.1. Pre-Defined Multicast Addresses)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.7.1
	if ip.Is6() {
		return ip.v6u16(0)&0xff0f == 0xff02
	}
	return false // zero value
}

// IsGlobalUnicast reports whether ip is a global unicast address.
//
// It returns true for IPv6 addresses which fall outside of the current
// IANA-allocated 2000::/3 global unicast space, with the exception of the
// link-local address space. It also returns true even if ip is in the IPv4
// private address space or IPv6 unique local address space.
// It returns false for the zero [Addr].
//
// For reference, see RFC 1122, RFC 4291, and RFC 4632.
func (ip Addr) IsGlobalUnicast() bool {
	if ip.bitlen == z0 {
		// Invalid or zero-value.
		return false
	}

	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// Match package net's IsGlobalUnicast logic. Notably private IPv4 addresses
	// and ULA IPv6 addresses are still considered "global unicast".
	ipv4Unspecified := IPv4Unspecified()
	ipv4Broadcast := AddrFrom4([4]byte{255, 255, 255, 255})
	if ip.Is4() && (ip.Equal(ipv4Unspecified) || ip.Equal(ipv4Broadcast)) {
		return false
	}

	ipv6Unspecified := IPv6Unspecified()
	return !ip.Equal(ipv6Unspecified) &&
		!ip.IsLoopback() &&
		!ip.IsMulticast() &&
		!ip.IsLinkLocalUnicast()
}

// IsPrivate reports whether ip is a private address, according to RFC 1918
// (IPv4 addresses) and RFC 4193 (IPv6 addresses). That is, it reports whether
// ip is in 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, or fc00::/7.
func (ip Addr) IsPrivate() bool {
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	// Match the stdlib's IsPrivate logic.
	if ip.Is4() {
		// RFC 1918 allocates 10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16 as
		// private IPv4 address subnets.
		return ip.v4(0) == 10 ||
			(ip.v4(0) == 172 && ip.v4(1)&0xf0 == 16) ||
			(ip.v4(0) == 192 && ip.v4(1) == 168)
	}

	if ip.Is6() {
		// RFC 4193 allocates fc00::/7 as the unique local unicast IPv6 address
		// subnet.
		return ip.v6(0)&0xfe == 0xfc
	}

	return false // zero value
}

// IsUnspecified reports whether ip is an unspecified address, either the IPv4
// address "0.0.0.0" or the IPv6 address "::".
//
// Note that the zero [Addr] is not an unspecified address.
func (ip Addr) IsUnspecified() bool {
	ipv4Unspecified := IPv4Unspecified()
	ipv6Unspecified := IPv6Unspecified()
	return ip.Equal(ipv4Unspecified) || ip.Equal(ipv6Unspecified)
}

// Prefix keeps only the top b bits of IP, producing a Prefix
// of the specified length.
// If ip is a zero [Addr], Prefix always returns a zero Prefix and a nil error.
// Otherwise, if bits is less than zero or greater than ip.BitLen(),
// Prefix returns an error.
func (ip Addr) Prefix(b int) (Prefix, error) {
	if b < 0 {
		return Prefix{}, ErrNegativePrefix
	}
	effectiveBits := b
	switch ip.bitlen {
	case z0:
		return Prefix{}, nil
	case z4:
		if b > 32 {
			return Prefix{}, ErrLargePrefix
		}
		effectiveBits += 96
	default:
		if b > 128 {
			return Prefix{}, ErrLargePrefix
		}
	}
	ip.addr = ip.addr.and(mask6(effectiveBits))
	return PrefixFrom(ip, b), nil
}

// As16 returns the IP address in its 16-byte representation.
// IPv4 addresses are returned as IPv4-mapped IPv6 addresses.
// IPv6 addresses with zones are returned without their zone (use the
// [Addr.Zone] method to get it).
// The ip zero value returns all zeroes.
func (ip Addr) As16(a16 [16]byte) [16]byte {
	binary.BigEndian.PutUint64(a16[:8], ip.addr.hi)
	binary.BigEndian.PutUint64(a16[8:], ip.addr.lo)
	return a16
}

// As4 returns an IPv4 or IPv4-in-IPv6 address in its 4-byte representation.
// If ip is the zero [Addr] or an IPv6 address, As4 panics.
// Note that 0.0.0.0 is not the zero Addr.
func (ip Addr) As4(a4 [4]byte) [4]byte {
	if ip.bitlen == z4 || ip.Is4In6() {
		binary.BigEndian.PutUint32(a4[:], uint32(ip.addr.lo))
		return a4
	}
	if ip.bitlen == z0 {
		panic("As4 called on IP zero value")
	}
	panic("As4 called on IPv6 address")
}

// AsSlice returns an IPv4 or IPv6 address in its respective 4-byte or 16-byte representation.
func (ip Addr) AsSlice(b []byte) []byte {
	switch ip.bitlen {
	case z0:
		return nil
	case z4:
		ret := b[:4]
		binary.BigEndian.PutUint32(ret[:], uint32(ip.addr.lo))
		return ret[:]
	default:
		ret := b[:16]
		binary.BigEndian.PutUint64(ret[:8], ip.addr.hi)
		binary.BigEndian.PutUint64(ret[8:], ip.addr.lo)
		return ret[:]
	}
}

// Next returns the address following ip.
// If there is none, it returns the zero [Addr].
func (ip Addr) Next() Addr {
	ip.addr = ip.addr.addOne()
	if ip.Is4() {
		if uint32(ip.addr.lo) == 0 {
			// Overflowed.
			return Addr{}
		}
	} else {
		if ip.addr.isZero() {
			// Overflowed
			return Addr{}
		}
	}
	return ip
}

// Prev returns the IP before ip.
// If there is none, it returns the IP zero value.
func (ip Addr) Prev() Addr {
	if ip.Is4() {
		if uint32(ip.addr.lo) == 0 {
			return Addr{}
		}
	} else if ip.addr.isZero() {
		return Addr{}
	}
	ip.addr = ip.addr.subOne()
	return ip
}

// AppendText implements the [encoding.TextAppender] interface.
// Requires at least [MaxAddrLen] bytes of spare capacity in b.
// Always returns a nil error.
func (ip Addr) AppendText(b []byte) ([]byte, error) {
	return ip.appendTo(b), nil
}

// String returns the string form of the IP address ip.
// It returns one of 5 forms:
//
//   - "invalid IP", if ip is the zero [Addr]
//   - IPv4 dotted decimal ("192.0.2.1")
//   - IPv6 ("2001:db8::1")
//   - "::ffff:1.2.3.4" (if [Addr.Is4In6])
//   - IPv6 with zone ("fe80:db8::1%eth0")
//
// Note that unlike package net's IP.String method,
// IPv4-mapped IPv6 addresses format with a "::ffff:"
// prefix before the dotted quad.
//
// buf length must be at least [MaxAddr4Len] for IPv4 addresses
// and [MaxAddr6Len] for IPv6 addresses.
func (ip Addr) String(buf []byte) string {
	switch ip.bitlen {
	case z0:
		return "invalid IP"
	case z4:
		return ip.string4(buf)
	default:
		if ip.Is4In6() {
			return ip.string4In6(buf)
		}
		return ip.string6(buf)
	}
}

// digits is a string of the hex digits from 0 to f. It's used in
// appendDecimal and appendHex to format IP addresses.
const digits = "0123456789abcdef"

// appendDecimal appends the decimal string representation of x to b.
func appendDecimal(b []byte, x uint8) []byte {
	// Using this function rather than strconv.AppendUint makes IPv4
	// string building 2x faster.

	if x >= 100 {
		b = append(b, digits[x/100])
	}
	if x >= 10 {
		b = append(b, digits[x/10%10])
	}
	return append(b, digits[x%10])
}

// appendHex appends the hex string representation of x to b.
func appendHex(b []byte, x uint16) []byte {
	// Using this function rather than strconv.AppendUint makes IPv6
	// string building 2x faster.

	if x >= 0x1000 {
		b = append(b, digits[x>>12])
	}
	if x >= 0x100 {
		b = append(b, digits[x>>8&0xf])
	}
	if x >= 0x10 {
		b = append(b, digits[x>>4&0xf])
	}
	return append(b, digits[x&0xf])
}

func (ip Addr) string4(buf []byte) string {
	ret := ip.appendTo4(buf[:0])
	return string(ret)
}

// appendTo appends a text encoding of ip
// to b and returns the extended buffer.
func (ip Addr) appendTo(b []byte) []byte {
	switch ip.bitlen {
	case z0:
		return b
	case z4:
		return ip.appendTo4(b)
	default:
		if ip.Is4In6() {
			return ip.appendTo4In6(b)
		}
		return ip.appendTo6(b)
	}
}

func (ip Addr) appendTo4(ret []byte) []byte {
	ret = appendDecimal(ret, ip.v4(0))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(1))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(2))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(3))
	return ret
}

func (ip Addr) string4In6(buf []byte) string {
	ret := ip.appendTo4In6(buf[:0])
	return string(ret)
}

func (ip Addr) appendTo4In6(ret []byte) []byte {
	ret = append(ret, "::ffff:"...)
	ret = ip.Unmap().appendTo4(ret)
	if ip.scopeID != 0 {
		ret = append(ret, '%')
		ret = strconv.AppendUint(ret, uint64(ip.scopeID), 10)
	}
	return ret
}

// string6 formats ip in IPv6 textual representation. It follows the
// guidelines in section 4 of RFC 5952
// (https://tools.ietf.org/html/rfc5952#section-4): no unnecessary
// zeros, use :: to elide the longest run of zeros, and don't use ::
// to compact a single zero field.
func (ip Addr) string6(buf []byte) string {
	// Use a zone with a "plausibly long" name, so that most zone-ful
	// IP addresses won't require additional allocation.
	//
	// The compiler does a cool optimization here, where ret ends up
	// stack-allocated and so the only allocation this function does
	// is to construct the returned string. As such, it's okay to be a
	// bit greedy here, size-wise.
	ret := ip.appendTo6(buf[:0])
	return string(ret)
}

func (ip Addr) appendTo6(ret []byte) []byte {
	zeroStart, zeroEnd := uint8(255), uint8(255)
	for i := uint8(0); i < 8; i++ {
		j := i
		for j < 8 && ip.v6u16(j) == 0 {
			j++
		}
		if l := j - i; l >= 2 && l > zeroEnd-zeroStart {
			zeroStart, zeroEnd = i, j
		}
	}

	for i := uint8(0); i < 8; i++ {
		if i == zeroStart {
			ret = append(ret, ':', ':')
			i = zeroEnd
			if i >= 8 {
				break
			}
		} else if i > 0 {
			ret = append(ret, ':')
		}

		ret = appendHex(ret, ip.v6u16(i))
	}

	if ip.scopeID != 0 {
		ret = append(ret, '%')
		ret = strconv.AppendUint(ret, uint64(ip.scopeID), 10)
	}
	return ret
}
