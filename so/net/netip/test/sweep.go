// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"solod.dev/so/encoding/binary"
	"solod.dev/so/net/netip"
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

// addr4From returns the IPv4 address of v.
func addr4From(v uint32) netip.Addr {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return netip.AddrFrom4(b)
}

// addr6From returns the IPv6 address of the two halves.
func addr6From(hi, lo uint64) netip.Addr {
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], hi)
	binary.BigEndian.PutUint64(b[8:], lo)
	return netip.AddrFrom16(b)
}

// mask32 returns the mask of the first bits of a 32-bit address.
func mask32(bits int) uint32 {
	if bits == 0 {
		return 0
	}
	return ^uint32(0) << uint(32-bits)
}

// mask64hi returns the mask of the high half of a 128-bit address.
func mask64hi(bits int) uint64 {
	if bits == 0 {
		return 0
	}
	if bits >= 64 {
		return ^uint64(0)
	}
	return ^uint64(0) << uint(64-bits)
}

// mask64lo returns the mask of the low half of a 128-bit address.
func mask64lo(bits int) uint64 {
	if bits <= 64 {
		return 0
	}
	if bits >= 128 {
		return ^uint64(0)
	}
	return ^uint64(0) << uint(128-bits)
}

// flip32 returns v with bit i flipped. Bit 0 is the first bit of the address.
func flip32(v uint32, i int) uint32 {
	return v ^ (uint32(1) << uint(31-i))
}

// The addresses of the IPv4 sweeps.
var sweep4 = []uint32{
	0x00000000,
	0xffffffff,
	0xc0a80f0f, // 192.168.15.15
	0x64988c42, // 100.152.140.66
	0x01020304, // 1.2.3.4
	0x7f000001, // 127.0.0.1
	0xaaaaaaaa,
	0x55555555,
}

// A v6Addr is one address of the IPv6 sweeps.
type v6Addr struct {
	hi uint64
	lo uint64
}

// The addresses of the IPv6 sweeps.
var sweep6 = []v6Addr{
	{hi: 0x0000000000000000, lo: 0x0000000000000000},
	{hi: 0xffffffffffffffff, lo: 0xffffffffffffffff},
	{hi: 0x20010db812345678, lo: 0x9abcdef011223344},
	{hi: 0xfe80000000000000, lo: 0xdeadbeefdeadbeef},
	{hi: 0xaaaaaaaaaaaaaaaa, lo: 0x5555555555555555},
	{hi: 0x0000000000000000, lo: 0x0000000000000001},
}

func TestSweepPrefix4(t *testing.T) {
	// Mask every IPv4 address of the sweep with every bit count.
	for _, v := range sweep4 {
		ip := addr4From(v)
		for bits := 0; bits <= 32; bits++ {
			p, err := ip.Prefix(bits)
			if err != nil {
				t.Errorf("Prefix(%x, %d) failed: %s", v, bits, err.Error())
				continue
			}
			want := addr4From(v & mask32(bits))
			if !p.Addr().Equal(want) {
				t.Errorf("Prefix(%x, %d) is not %x", v, bits, v&mask32(bits))
			}
			if p.Bits() != bits {
				t.Errorf("Prefix(%x, %d).Bits() = %d", v, bits, p.Bits())
			}
			if !p.Contains(ip) {
				t.Errorf("Prefix(%x, %d) does not contain the address", v, bits)
			}
			if p.IsSingleIP() != (bits == 32) {
				t.Errorf("Prefix(%x, %d).IsSingleIP() = %t", v, bits, p.IsSingleIP())
			}
			// Masking an already masked prefix changes nothing.
			if !p.Masked().Equal(p) {
				t.Errorf("Prefix(%x, %d).Masked() != Prefix(%x, %d)", v, bits, v, bits)
			}
			// PrefixFrom and Masked give the same prefix.
			if !netip.PrefixFrom(ip, bits).Masked().Equal(p) {
				t.Errorf("PrefixFrom(%x, %d).Masked() != Prefix(%x, %d)", v, bits, v, bits)
			}
		}
	}
}

func TestSweepPrefix6(t *testing.T) {
	// Mask every IPv6 address of the sweep with every bit count.
	for _, v := range sweep6 {
		ip := addr6From(v.hi, v.lo)
		for bits := 0; bits <= 128; bits++ {
			p, err := ip.Prefix(bits)
			if err != nil {
				t.Errorf("Prefix(%x:%x, %d) failed: %s", v.hi, v.lo, bits, err.Error())
				continue
			}
			want := addr6From(v.hi&mask64hi(bits), v.lo&mask64lo(bits))
			if !p.Addr().Equal(want) {
				t.Errorf("Prefix(%x:%x, %d) is not %x:%x", v.hi, v.lo, bits,
					v.hi&mask64hi(bits), v.lo&mask64lo(bits))
			}
			if p.Bits() != bits {
				t.Errorf("Prefix(%x:%x, %d).Bits() = %d", v.hi, v.lo, bits, p.Bits())
			}
			if !p.Contains(ip) {
				t.Errorf("Prefix(%x:%x, %d) does not contain the address", v.hi, v.lo, bits)
			}
			if p.IsSingleIP() != (bits == 128) {
				t.Errorf("Prefix(%x:%x, %d).IsSingleIP() = %t",
					v.hi, v.lo, bits, p.IsSingleIP())
			}
			if !p.Masked().Equal(p) {
				t.Errorf("Prefix(%x:%x, %d).Masked() is a different prefix", v.hi, v.lo, bits)
			}
			if !netip.PrefixFrom(ip, bits).Masked().Equal(p) {
				t.Errorf("PrefixFrom(%x:%x, %d).Masked() is a different prefix",
					v.hi, v.lo, bits)
			}
		}
	}
}

func TestSweepContains4(t *testing.T) {
	// Flip one bit of an IPv4 address for every bit count.
	// A flipped bit inside the prefix puts the address out of the prefix.
	// A flipped bit after the prefix keeps the address inside.
	for _, v := range sweep4 {
		ip := addr4From(v)
		for bits := 1; bits <= 32; bits++ {
			p, err := ip.Prefix(bits)
			if err != nil {
				t.Errorf("Prefix(%x, %d) failed: %s", v, bits, err.Error())
				continue
			}
			out := addr4From(flip32(v, bits-1))
			if p.Contains(out) {
				t.Errorf("Prefix(%x, %d) contains an address of another prefix", v, bits)
			}
			if bits < 32 {
				in := addr4From(flip32(v, bits))
				if !p.Contains(in) {
					t.Errorf("Prefix(%x, %d) does not contain an address of the prefix",
						v, bits)
				}
			}
		}

		// A prefix of no bits contains every address of the family.
		p, err := ip.Prefix(0)
		if err != nil {
			t.Errorf("Prefix(%x, 0) failed: %s", v, err.Error())
			continue
		}
		for _, other := range sweep4 {
			if !p.Contains(addr4From(other)) {
				t.Errorf("Prefix(%x, 0) does not contain %x", v, other)
			}
		}
	}
}

func TestSweepContains6(t *testing.T) {
	// Flip one bit of an IPv6 address for every bit count.
	for _, v := range sweep6 {
		ip := addr6From(v.hi, v.lo)
		for bits := 1; bits <= 128; bits++ {
			p, err := ip.Prefix(bits)
			if err != nil {
				t.Errorf("Prefix(%x:%x, %d) failed: %s", v.hi, v.lo, bits, err.Error())
				continue
			}

			outHi, outLo := flip128(v.hi, v.lo, bits-1)
			if p.Contains(addr6From(outHi, outLo)) {
				t.Errorf("Prefix(%x:%x, %d) contains an address of another prefix",
					v.hi, v.lo, bits)
			}
			if bits < 128 {
				inHi, inLo := flip128(v.hi, v.lo, bits)
				if !p.Contains(addr6From(inHi, inLo)) {
					t.Errorf("Prefix(%x:%x, %d) does not contain an address of the prefix",
						v.hi, v.lo, bits)
				}
			}
		}

		p, err := ip.Prefix(0)
		if err != nil {
			t.Errorf("Prefix(%x:%x, 0) failed: %s", v.hi, v.lo, err.Error())
			continue
		}
		for _, other := range sweep6 {
			if !p.Contains(addr6From(other.hi, other.lo)) {
				t.Errorf("Prefix(%x:%x, 0) does not contain %x:%x",
					v.hi, v.lo, other.hi, other.lo)
			}
		}
	}
}

// flip128 returns the two halves with bit i flipped.
// Bit 0 is the first bit of the address.
func flip128(hi, lo uint64, i int) (uint64, uint64) {
	if i < 64 {
		return hi ^ (uint64(1) << uint(63-i)), lo
	}
	return hi, lo ^ (uint64(1) << uint(127-i))
}

func TestSweepNextPrev4(t *testing.T) {
	// Step an IPv4 address over every carry boundary.
	for k := range 32 {
		ones := uint32(1)<<uint(k) - 1 // k low bits set
		pow2 := uint32(1) << uint(k)   // the next value

		ip := addr4From(ones)
		want := addr4From(pow2)
		if !ip.Next().Equal(want) {
			t.Errorf("Next(%x) is not %x", ones, pow2)
		}
		if !want.Prev().Equal(ip) {
			t.Errorf("Prev(%x) is not %x", pow2, ones)
		}
	}

	// The last address has no next one, the first has no previous one.
	if addr4From(0xffffffff).Next().IsValid() {
		t.Error("Next(255.255.255.255) is valid, want the zero address")
	}
	if addr4From(0).Prev().IsValid() {
		t.Error("Prev(0.0.0.0) is valid, want the zero address")
	}
}

func TestSweepNextPrev6(t *testing.T) {
	// Step an IPv6 address over every carry boundary.
	for k := range 128 {
		onesHi, onesLo := lowOnes(k)
		pow2Hi, pow2Lo := bit128(k)

		ip := addr6From(onesHi, onesLo)
		want := addr6From(pow2Hi, pow2Lo)
		if !ip.Next().Equal(want) {
			t.Errorf("Next(%x:%x) is not %x:%x", onesHi, onesLo, pow2Hi, pow2Lo)
		}
		if !want.Prev().Equal(ip) {
			t.Errorf("Prev(%x:%x) is not %x:%x", pow2Hi, pow2Lo, onesHi, onesLo)
		}
	}

	if addr6From(^uint64(0), ^uint64(0)).Next().IsValid() {
		t.Error("Next(the last address) is valid, want the zero address")
	}
	if addr6From(0, 0).Prev().IsValid() {
		t.Error("Prev(::) is valid, want the zero address")
	}
}

// lowOnes returns the two halves of a 128-bit value with the k low bits set.
func lowOnes(k int) (uint64, uint64) {
	if k < 64 {
		return 0, uint64(1)<<uint(k) - 1
	}
	if k == 64 {
		return 0, ^uint64(0)
	}
	return uint64(1)<<uint(k-64) - 1, ^uint64(0)
}

// bit128 returns the two halves of a 128-bit value with only bit k set.
// Bit 0 is the last bit of the address.
func bit128(k int) (uint64, uint64) {
	if k < 64 {
		return 0, uint64(1) << uint(k)
	}
	return uint64(1) << uint(k-64), 0
}

func TestSweepAddr4Text(t *testing.T) {
	// Format and parse back every value of every octet.
	for pos := range 4 {
		for v := 0; v <= 255; v++ {
			var b [4]byte
			b[0] = 10
			b[1] = 200
			b[2] = 30
			b[3] = 255
			b[pos] = byte(v)
			ip := netip.AddrFrom4(b)

			var textBuf [bufLen]byte
			want := text4(textBuf[:0], b[0], b[1], b[2], b[3])

			var buf [bufLen]byte
			got := ip.String(buf[:])
			if got != want {
				t.Errorf("String() = %s, want %s", got, want)
				continue
			}

			back, err := netip.ParseAddr(got)
			if err != nil {
				t.Errorf("ParseAddr(%s) failed: %s", got, err.Error())
				continue
			}
			if !back.Equal(ip) {
				t.Errorf("ParseAddr(%s) is another address", got)
			}
			if !back.Is4() {
				t.Errorf("ParseAddr(%s) is not an IPv4 address", got)
			}
		}
	}
}

// text4 writes the dotted decimal form of the four bytes to dst.
func text4(dst []byte, a, b, c, d byte) string {
	dst = strconv.AppendUint(dst, uint64(a), 10)
	dst = append(dst, '.')
	dst = strconv.AppendUint(dst, uint64(b), 10)
	dst = append(dst, '.')
	dst = strconv.AppendUint(dst, uint64(c), 10)
	dst = append(dst, '.')
	dst = strconv.AppendUint(dst, uint64(d), 10)
	return string(dst)
}

// The addresses of the order sweep, in no particular order.
var orderCases = [...]string{
	"", // the zero address
	"0.0.0.0",
	"1.2.3.4",
	"1.2.3.5",
	"255.255.255.255",
	"::",
	"::1",
	"::ffff:1.2.3.4",
	"2001:db8::1",
	"fe80::1",
	"fe80::1%1",
	"fe80::1%2",
	"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
}

func TestSweepCompareOrder(t *testing.T) {
	// Check that Compare is a total order over the addresses of the sweep.
	var addrs [len(orderCases)]netip.Addr
	for i, s := range orderCases {
		if s != "" {
			addrs[i] = netip.MustParseAddr(s)
		}
	}

	for i, a := range addrs {
		for j, b := range addrs {
			ab := a.Compare(b)
			ba := b.Compare(a)

			// Compare is antisymmetric.
			if ab != -ba {
				t.Errorf("Compare(%s, %s) = %d, Compare(%s, %s) = %d",
					orderCases[i], orderCases[j], ab, orderCases[j], orderCases[i], ba)
			}
			// Only one address equals itself.
			if (ab == 0) != (i == j) {
				t.Errorf("Compare(%s, %s) = %d", orderCases[i], orderCases[j], ab)
			}
			// Equal and Less agree with Compare.
			if a.Equal(b) != (ab == 0) {
				t.Errorf("Equal(%s, %s) does not agree with Compare",
					orderCases[i], orderCases[j])
			}
			if a.Less(b) != (ab < 0) {
				t.Errorf("Less(%s, %s) does not agree with Compare",
					orderCases[i], orderCases[j])
			}

			// Compare is transitive.
			for k, c := range addrs {
				bc := b.Compare(c)
				ac := a.Compare(c)
				if ab < 0 && bc < 0 && ac >= 0 {
					t.Errorf("Compare is not transitive over %s, %s, %s",
						orderCases[i], orderCases[j], orderCases[k])
				}
			}
		}
	}
}
