package bits_test

import (
	"solod.dev/so/math/bits"
	"solod.dev/so/testing"
)

const (
	_M   = ^uint(0)
	_M32 = ^uint32(0)
	_M64 = ^uint64(0)
)

// deBruijn64 is the constant that the package uses for its lookup tables. The
// tests use it as a source of well mixed bits.
const deBruijn64 = 0x03f79d71b4ca8b09

// entry holds the reference results for one uint8 value.
type entry struct {
	nlz int // the number of leading zero bits
	ntz int // the number of trailing zero bits
	pop int // the number of one bits
}

// tab holds the reference results for every uint8 value.
var tab [256]entry

func init() {
	tab[0] = entry{8, 8, 0}
	for i := 1; i < len(tab); i++ {
		x := i
		n := 0
		for x&0x80 == 0 {
			n++
			x <<= 1
		}
		tab[i].nlz = n

		x = i
		n = 0
		for x&1 == 0 {
			n++
			x >>= 1
		}
		tab[i].ntz = n

		x = i
		n = 0
		for x != 0 {
			n += x & 1
			x >>= 1
		}
		tab[i].pop = n
	}
}

// hexDigits holds the digits of a hexadecimal number.
const hexDigits = "0123456789abcdef"

// hex writes the low n digits of x into buf as hexadecimal and returns the
// result. The integer verbs of so/fmt take a word sized value, which is 32
// bits on a freestanding target, so a uint64 needs a conversion of its own.
// The result is a view of buf, so every value of one message needs a buffer
// of its own.
func hex(buf []byte, x uint64, n int) string {
	for i := n - 1; i >= 0; i-- {
		buf[i] = hexDigits[x&0xf]
		x >>= 4
	}
	return string(buf[:n])
}

// checkValue checks one result against the wanted value. The name and the
// index identify the failed check.
func checkValue(t *testing.T, name string, i int, got, want uint64) {
	if got == want {
		return
	}
	var b1, b2 [16]byte
	t.Errorf("%s case %d = %s, want %s", name, i, hex(b1[:], got, 16), hex(b2[:], want, 16))
}

// checkPair checks a result pair against the wanted pair. The name and the
// index identify the failed check.
func checkPair(t *testing.T, name string, i int, got1, got2, want1, want2 uint64) {
	if got1 == want1 && got2 == want2 {
		return
	}
	var b1, b2, b3, b4 [16]byte
	t.Errorf("%s case %d = %s:%s, want %s:%s", name, i,
		hex(b1[:], got1, 16), hex(b2[:], got2, 16),
		hex(b3[:], want1, 16), hex(b4[:], want2, 16))
}

func TestUintSize(t *testing.T) {
	if bits.UintSize != 32 && bits.UintSize != 64 {
		t.Errorf("UintSize = %d, want 32 or 64", bits.UintSize)
		return
	}
	if got := bits.Len(_M); got != bits.UintSize {
		t.Errorf("Len(^uint(0)) = %d, want %d", got, bits.UintSize)
	}
	if got := bits.OnesCount(_M); got != bits.UintSize {
		t.Errorf("OnesCount(^uint(0)) = %d, want %d", got, bits.UintSize)
	}
}

func TestLeadingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		nlz := tab[i].nlz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			var b [16]byte

			if x <= 0xff {
				want := nlz - k
				if x == 0 {
					want = 8
				}
				if got := bits.LeadingZeros8(uint8(x)); got != want {
					t.Errorf("LeadingZeros8(%s) = %d, want %d", hex(b[:], x, 2), got, want)
					return
				}
			}

			if x <= 0xffff {
				want := nlz - k + 8
				if x == 0 {
					want = 16
				}
				if got := bits.LeadingZeros16(uint16(x)); got != want {
					t.Errorf("LeadingZeros16(%s) = %d, want %d", hex(b[:], x, 4), got, want)
					return
				}
			}

			if x <= 0xffffffff {
				want := nlz - k + 24
				if x == 0 {
					want = 32
				}
				if got := bits.LeadingZeros32(uint32(x)); got != want {
					t.Errorf("LeadingZeros32(%s) = %d, want %d", hex(b[:], x, 8), got, want)
					return
				}
				if bits.UintSize == 32 {
					if got := bits.LeadingZeros(uint(x)); got != want {
						t.Errorf("LeadingZeros(%s) = %d, want %d", hex(b[:], x, 8), got, want)
						return
					}
				}
			}

			want := nlz - k + 56
			if x == 0 {
				want = 64
			}
			if got := bits.LeadingZeros64(x); got != want {
				t.Errorf("LeadingZeros64(%s) = %d, want %d", hex(b[:], x, 16), got, want)
				return
			}
			if bits.UintSize == 64 {
				if got := bits.LeadingZeros(uint(x)); got != want {
					t.Errorf("LeadingZeros(%s) = %d, want %d", hex(b[:], x, 16), got, want)
					return
				}
			}
		}
	}
}

func TestTrailingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		ntz := tab[i].ntz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			var b [16]byte

			if x <= 0xff {
				want := ntz + k
				if x == 0 {
					want = 8
				}
				if got := bits.TrailingZeros8(uint8(x)); got != want {
					t.Errorf("TrailingZeros8(%s) = %d, want %d", hex(b[:], x, 2), got, want)
					return
				}
			}

			if x <= 0xffff {
				want := ntz + k
				if x == 0 {
					want = 16
				}
				if got := bits.TrailingZeros16(uint16(x)); got != want {
					t.Errorf("TrailingZeros16(%s) = %d, want %d", hex(b[:], x, 4), got, want)
					return
				}
			}

			if x <= 0xffffffff {
				want := ntz + k
				if x == 0 {
					want = 32
				}
				if got := bits.TrailingZeros32(uint32(x)); got != want {
					t.Errorf("TrailingZeros32(%s) = %d, want %d", hex(b[:], x, 8), got, want)
					return
				}
				if bits.UintSize == 32 {
					if got := bits.TrailingZeros(uint(x)); got != want {
						t.Errorf("TrailingZeros(%s) = %d, want %d", hex(b[:], x, 8), got, want)
						return
					}
				}
			}

			want := ntz + k
			if x == 0 {
				want = 64
			}
			if got := bits.TrailingZeros64(x); got != want {
				t.Errorf("TrailingZeros64(%s) = %d, want %d", hex(b[:], x, 16), got, want)
				return
			}
			if bits.UintSize == 64 {
				if got := bits.TrailingZeros(uint(x)); got != want {
					t.Errorf("TrailingZeros(%s) = %d, want %d", hex(b[:], x, 16), got, want)
					return
				}
			}
		}
	}
}

func TestOnesCount(t *testing.T) {
	// A run of one bits that grows from the low end.
	var x uint64
	for i := 0; i <= 64; i++ {
		checkOnesCount(t, x, i)
		x = x<<1 | 1
	}

	// A run of one bits that leaves at the low end.
	for i := 64; i >= 0; i-- {
		checkOnesCount(t, x, i)
		x = x << 1
	}

	// Every uint8 value at every offset.
	for i := 0; i < 256; i++ {
		for k := 0; k < 64-8; k++ {
			checkOnesCount(t, uint64(i)<<uint(k), tab[i].pop)
		}
	}
}

// checkOnesCount checks every OnesCount function that x fits.
func checkOnesCount(t *testing.T, x uint64, want int) {
	var b [16]byte

	if x <= 0xff {
		if got := bits.OnesCount8(uint8(x)); got != want {
			t.Errorf("OnesCount8(%s) = %d, want %d", hex(b[:], x, 2), got, want)
			return
		}
	}

	if x <= 0xffff {
		if got := bits.OnesCount16(uint16(x)); got != want {
			t.Errorf("OnesCount16(%s) = %d, want %d", hex(b[:], x, 4), got, want)
			return
		}
	}

	if x <= 0xffffffff {
		if got := bits.OnesCount32(uint32(x)); got != want {
			t.Errorf("OnesCount32(%s) = %d, want %d", hex(b[:], x, 8), got, want)
			return
		}
		if bits.UintSize == 32 {
			if got := bits.OnesCount(uint(x)); got != want {
				t.Errorf("OnesCount(%s) = %d, want %d", hex(b[:], x, 8), got, want)
				return
			}
		}
	}

	if got := bits.OnesCount64(x); got != want {
		t.Errorf("OnesCount64(%s) = %d, want %d", hex(b[:], x, 16), got, want)
		return
	}
	if bits.UintSize == 64 {
		if got := bits.OnesCount(uint(x)); got != want {
			t.Errorf("OnesCount(%s) = %d, want %d", hex(b[:], x, 16), got, want)
		}
	}
}

// The reference rotations below never shift by the width of the type. Go gives
// zero for such a shift, but C does not define it.

func rotl8(x uint8, k uint) uint8 {
	s := k & 7
	if s == 0 {
		return x
	}
	return x<<s | x>>(8-s)
}

func rotl16(x uint16, k uint) uint16 {
	s := k & 15
	if s == 0 {
		return x
	}
	return x<<s | x>>(16-s)
}

func rotl32(x uint32, k uint) uint32 {
	s := k & 31
	if s == 0 {
		return x
	}
	return x<<s | x>>(32-s)
}

func rotl64(x uint64, k uint) uint64 {
	s := k & 63
	if s == 0 {
		return x
	}
	return x<<s | x>>(64-s)
}

func rotlUint(x uint, k uint) uint {
	s := k & uint(bits.UintSize-1)
	if s == 0 {
		return x
	}
	return x<<s | x>>(uint(bits.UintSize)-s)
}

func TestRotateLeft(t *testing.T) {
	// The conversion sets the type of m. The default type is int, which is 32
	// bits on a freestanding target and cannot hold the value.
	m := uint64(deBruijn64)
	for k := uint(0); k < 128; k++ {
		x8 := uint8(m)
		want8 := rotl8(x8, k)
		if got := bits.RotateLeft8(x8, int(k)); got != want8 {
			checkValue(t, "RotateLeft8", int(k), uint64(got), uint64(want8))
			return
		}
		if got := bits.RotateLeft8(want8, -int(k)); got != x8 {
			checkValue(t, "RotateLeft8 back", int(k), uint64(got), uint64(x8))
			return
		}

		x16 := uint16(m)
		want16 := rotl16(x16, k)
		if got := bits.RotateLeft16(x16, int(k)); got != want16 {
			checkValue(t, "RotateLeft16", int(k), uint64(got), uint64(want16))
			return
		}
		if got := bits.RotateLeft16(want16, -int(k)); got != x16 {
			checkValue(t, "RotateLeft16 back", int(k), uint64(got), uint64(x16))
			return
		}

		x32 := uint32(m)
		want32 := rotl32(x32, k)
		if got := bits.RotateLeft32(x32, int(k)); got != want32 {
			checkValue(t, "RotateLeft32", int(k), uint64(got), uint64(want32))
			return
		}
		if got := bits.RotateLeft32(want32, -int(k)); got != x32 {
			checkValue(t, "RotateLeft32 back", int(k), uint64(got), uint64(x32))
			return
		}

		x64 := m
		want64 := rotl64(x64, k)
		if got := bits.RotateLeft64(x64, int(k)); got != want64 {
			checkValue(t, "RotateLeft64", int(k), got, want64)
			return
		}
		if got := bits.RotateLeft64(want64, -int(k)); got != x64 {
			checkValue(t, "RotateLeft64 back", int(k), got, x64)
			return
		}

		x := uint(m)
		want := rotlUint(x, k)
		if got := bits.RotateLeft(x, int(k)); got != want {
			checkValue(t, "RotateLeft", int(k), uint64(got), uint64(want))
			return
		}
		if got := bits.RotateLeft(want, -int(k)); got != x {
			checkValue(t, "RotateLeft back", int(k), uint64(got), uint64(x))
			return
		}
	}
}

// revCase is a value and the value with the bits in reverse order.
type revCase struct {
	x, r uint64
}

var revTests = [...]revCase{
	{0, 0},
	{0x1, 0x8 << 60},
	{0x2, 0x4 << 60},
	{0x3, 0xc << 60},
	{0x4, 0x2 << 60},
	{0x5, 0xa << 60},
	{0x6, 0x6 << 60},
	{0x7, 0xe << 60},
	{0x8, 0x1 << 60},
	{0x9, 0x9 << 60},
	{0xa, 0x5 << 60},
	{0xb, 0xd << 60},
	{0xc, 0x3 << 60},
	{0xd, 0xb << 60},
	{0xe, 0x7 << 60},
	{0xf, 0xf << 60},
	{0x5686487, 0xe12616a000000000},
	{0x0123456789abcdef, 0xf7b3d591e6a2c480},
}

func TestReverse(t *testing.T) {
	// Every bit on its own.
	for i := uint(0); i < 64; i++ {
		checkReverse(t, int(i), uint64(1)<<i, uint64(1)<<(63-i))
	}
	// A few patterns, in both directions.
	for i, tt := range revTests {
		checkReverse(t, i, tt.x, tt.r)
		checkReverse(t, i, tt.r, tt.x)
	}
}

// checkReverse checks every Reverse function against the wanted result. The
// narrow functions take the low bits of x and give the high bits of want.
func checkReverse(t *testing.T, i int, x, want uint64) {
	x8 := uint8(x)
	want8 := uint8(want >> (64 - 8))
	if got := bits.Reverse8(x8); got != want8 {
		checkValue(t, "Reverse8", i, uint64(got), uint64(want8))
		return
	}

	x16 := uint16(x)
	want16 := uint16(want >> (64 - 16))
	if got := bits.Reverse16(x16); got != want16 {
		checkValue(t, "Reverse16", i, uint64(got), uint64(want16))
		return
	}

	x32 := uint32(x)
	want32 := uint32(want >> (64 - 32))
	if got := bits.Reverse32(x32); got != want32 {
		checkValue(t, "Reverse32", i, uint64(got), uint64(want32))
		return
	}
	if bits.UintSize == 32 {
		if got := bits.Reverse(uint(x32)); got != uint(want32) {
			checkValue(t, "Reverse", i, uint64(got), uint64(want32))
			return
		}
	}

	if got := bits.Reverse64(x); got != want {
		checkValue(t, "Reverse64", i, got, want)
		return
	}
	if bits.UintSize == 64 {
		if got := bits.Reverse(uint(x)); got != uint(want) {
			checkValue(t, "Reverse", i, uint64(got), want)
		}
	}
}

var revBytesTests = [...]revCase{
	{0, 0},
	{0x01, 0x01 << 56},
	{0x0123, 0x2301 << 48},
	{0x012345, 0x452301 << 40},
	{0x01234567, 0x67452301 << 32},
	{0x0123456789, 0x8967452301 << 24},
	{0x0123456789ab, 0xab8967452301 << 16},
	{0x0123456789abcd, 0xcdab8967452301 << 8},
	{0x0123456789abcdef, 0xefcdab8967452301},
}

func TestReverseBytes(t *testing.T) {
	for i, tt := range revBytesTests {
		checkReverseBytes(t, i, tt.x, tt.r)
		checkReverseBytes(t, i, tt.r, tt.x)
	}
}

// checkReverseBytes checks every ReverseBytes function against the wanted
// result. The narrow functions take the low bytes of x and give the high
// bytes of want.
func checkReverseBytes(t *testing.T, i int, x, want uint64) {
	x16 := uint16(x)
	want16 := uint16(want >> (64 - 16))
	if got := bits.ReverseBytes16(x16); got != want16 {
		checkValue(t, "ReverseBytes16", i, uint64(got), uint64(want16))
		return
	}

	x32 := uint32(x)
	want32 := uint32(want >> (64 - 32))
	if got := bits.ReverseBytes32(x32); got != want32 {
		checkValue(t, "ReverseBytes32", i, uint64(got), uint64(want32))
		return
	}
	if bits.UintSize == 32 {
		if got := bits.ReverseBytes(uint(x32)); got != uint(want32) {
			checkValue(t, "ReverseBytes", i, uint64(got), uint64(want32))
			return
		}
	}

	if got := bits.ReverseBytes64(x); got != want {
		checkValue(t, "ReverseBytes64", i, got, want)
		return
	}
	if bits.UintSize == 64 {
		if got := bits.ReverseBytes(uint(x)); got != uint(want) {
			checkValue(t, "ReverseBytes", i, uint64(got), want)
		}
	}
}

func TestLen(t *testing.T) {
	for i := 0; i < 256; i++ {
		length := 8 - tab[i].nlz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := 0
			if x != 0 {
				want = length + k
			}
			var b [16]byte

			if x <= 0xff {
				if got := bits.Len8(uint8(x)); got != want {
					t.Errorf("Len8(%s) = %d, want %d", hex(b[:], x, 2), got, want)
					return
				}
			}

			if x <= 0xffff {
				if got := bits.Len16(uint16(x)); got != want {
					t.Errorf("Len16(%s) = %d, want %d", hex(b[:], x, 4), got, want)
					return
				}
			}

			if x <= 0xffffffff {
				if got := bits.Len32(uint32(x)); got != want {
					t.Errorf("Len32(%s) = %d, want %d", hex(b[:], x, 8), got, want)
					return
				}
				if bits.UintSize == 32 {
					if got := bits.Len(uint(x)); got != want {
						t.Errorf("Len(%s) = %d, want %d", hex(b[:], x, 8), got, want)
						return
					}
				}
			}

			if got := bits.Len64(x); got != want {
				t.Errorf("Len64(%s) = %d, want %d", hex(b[:], x, 16), got, want)
				return
			}
			if bits.UintSize == 64 {
				if got := bits.Len(uint(x)); got != want {
					t.Errorf("Len(%s) = %d, want %d", hex(b[:], x, 16), got, want)
					return
				}
			}
		}
	}
}

// addSubCase holds x + y + c = z with the carry out cout.
type addSubCase struct {
	x, y, c, z, cout uint
}

var addSubTests = [...]addSubCase{
	{0, 0, 0, 0, 0},
	{0, 1, 0, 1, 0},
	{0, 0, 1, 1, 0},
	{0, 1, 1, 2, 0},
	{12345, 67890, 0, 80235, 0},
	{12345, 67890, 1, 80236, 0},
	{_M, 1, 0, 0, 1},
	{_M, 0, 1, 0, 1},
	{_M, 1, 1, 1, 1},
	{_M, _M, 0, _M - 1, 1},
	{_M, _M, 1, _M, 1},
}

func TestAddSubUint(t *testing.T) {
	for i, a := range addSubTests {
		z, cout := bits.Add(a.x, a.y, a.c)
		checkPair(t, "Add", i, uint64(z), uint64(cout), uint64(a.z), uint64(a.cout))
		z, cout = bits.Add(a.y, a.x, a.c)
		checkPair(t, "Add symmetric", i, uint64(z), uint64(cout), uint64(a.z), uint64(a.cout))

		y, cout := bits.Sub(a.z, a.x, a.c)
		checkPair(t, "Sub", i, uint64(y), uint64(cout), uint64(a.y), uint64(a.cout))
		x, cout := bits.Sub(a.z, a.y, a.c)
		checkPair(t, "Sub symmetric", i, uint64(x), uint64(cout), uint64(a.x), uint64(a.cout))
	}
}

type addSubCase32 struct {
	x, y, c, z, cout uint32
}

var addSubTests32 = [...]addSubCase32{
	{0, 0, 0, 0, 0},
	{0, 1, 0, 1, 0},
	{0, 0, 1, 1, 0},
	{0, 1, 1, 2, 0},
	{12345, 67890, 0, 80235, 0},
	{12345, 67890, 1, 80236, 0},
	{_M32, 1, 0, 0, 1},
	{_M32, 0, 1, 0, 1},
	{_M32, 1, 1, 1, 1},
	{_M32, _M32, 0, _M32 - 1, 1},
	{_M32, _M32, 1, _M32, 1},
}

func TestAddSubUint32(t *testing.T) {
	for i, a := range addSubTests32 {
		z, cout := bits.Add32(a.x, a.y, a.c)
		checkPair(t, "Add32", i, uint64(z), uint64(cout), uint64(a.z), uint64(a.cout))
		z, cout = bits.Add32(a.y, a.x, a.c)
		checkPair(t, "Add32 symmetric", i, uint64(z), uint64(cout), uint64(a.z), uint64(a.cout))

		y, cout := bits.Sub32(a.z, a.x, a.c)
		checkPair(t, "Sub32", i, uint64(y), uint64(cout), uint64(a.y), uint64(a.cout))
		x, cout := bits.Sub32(a.z, a.y, a.c)
		checkPair(t, "Sub32 symmetric", i, uint64(x), uint64(cout), uint64(a.x), uint64(a.cout))
	}
}

type addSubCase64 struct {
	x, y, c, z, cout uint64
}

var addSubTests64 = [...]addSubCase64{
	{0, 0, 0, 0, 0},
	{0, 1, 0, 1, 0},
	{0, 0, 1, 1, 0},
	{0, 1, 1, 2, 0},
	{12345, 67890, 0, 80235, 0},
	{12345, 67890, 1, 80236, 0},
	{_M64, 1, 0, 0, 1},
	{_M64, 0, 1, 0, 1},
	{_M64, 1, 1, 1, 1},
	{_M64, _M64, 0, _M64 - 1, 1},
	{_M64, _M64, 1, _M64, 1},
}

func TestAddSubUint64(t *testing.T) {
	for i, a := range addSubTests64 {
		z, cout := bits.Add64(a.x, a.y, a.c)
		checkPair(t, "Add64", i, z, cout, a.z, a.cout)
		z, cout = bits.Add64(a.y, a.x, a.c)
		checkPair(t, "Add64 symmetric", i, z, cout, a.z, a.cout)

		y, cout := bits.Sub64(a.z, a.x, a.c)
		checkPair(t, "Sub64", i, y, cout, a.y, a.cout)
		x, cout := bits.Sub64(a.z, a.y, a.c)
		checkPair(t, "Sub64 symmetric", i, x, cout, a.x, a.cout)
	}
}

// mulDivCase holds x * y = (hi, lo). The remainder r turns the product into
// a dividend that Div takes apart again.
type mulDivCase struct {
	x, y      uint
	hi, lo, r uint
}

var mulDivTests = [...]mulDivCase{
	{_M/2 + 1, 2, 1, 0, 1},
	{_M, _M, _M - 1, 1, 42},
}

func TestMulDiv(t *testing.T) {
	for i, a := range mulDivTests {
		hi, lo := bits.Mul(a.x, a.y)
		checkPair(t, "Mul", i, uint64(hi), uint64(lo), uint64(a.hi), uint64(a.lo))
		hi, lo = bits.Mul(a.y, a.x)
		checkPair(t, "Mul symmetric", i, uint64(hi), uint64(lo), uint64(a.hi), uint64(a.lo))

		q, r := bits.Div(a.hi, a.lo+a.r, a.y)
		checkPair(t, "Div", i, uint64(q), uint64(r), uint64(a.x), uint64(a.r))
		q, r = bits.Div(a.hi, a.lo+a.r, a.x)
		checkPair(t, "Div symmetric", i, uint64(q), uint64(r), uint64(a.y), uint64(a.r))
	}
}

type mulDivCase32 struct {
	x, y      uint32
	hi, lo, r uint32
}

var mulDivTests32 = [...]mulDivCase32{
	{1 << 31, 2, 1, 0, 1},
	{0xc47dfa8c, 50911, 0x98a4, 0x998587f4, 13},
	{_M32, _M32, _M32 - 1, 1, 42},
}

func TestMulDiv32(t *testing.T) {
	for i, a := range mulDivTests32 {
		hi, lo := bits.Mul32(a.x, a.y)
		checkPair(t, "Mul32", i, uint64(hi), uint64(lo), uint64(a.hi), uint64(a.lo))
		hi, lo = bits.Mul32(a.y, a.x)
		checkPair(t, "Mul32 symmetric", i, uint64(hi), uint64(lo), uint64(a.hi), uint64(a.lo))

		q, r := bits.Div32(a.hi, a.lo+a.r, a.y)
		checkPair(t, "Div32", i, uint64(q), uint64(r), uint64(a.x), uint64(a.r))
		q, r = bits.Div32(a.hi, a.lo+a.r, a.x)
		checkPair(t, "Div32 symmetric", i, uint64(q), uint64(r), uint64(a.y), uint64(a.r))
	}
}

type mulDivCase64 struct {
	x, y      uint64
	hi, lo, r uint64
}

var mulDivTests64 = [...]mulDivCase64{
	{1 << 63, 2, 1, 0, 1},
	{0x3626229738a3b9, 0xd8988a9f1cc4a61, 0x2dd0712657fe8, 0x9dd6a3364c358319, 13},
	{_M64, _M64, _M64 - 1, 1, 42},
}

func TestMulDiv64(t *testing.T) {
	for i, a := range mulDivTests64 {
		hi, lo := bits.Mul64(a.x, a.y)
		checkPair(t, "Mul64", i, hi, lo, a.hi, a.lo)
		hi, lo = bits.Mul64(a.y, a.x)
		checkPair(t, "Mul64 symmetric", i, hi, lo, a.hi, a.lo)

		q, r := bits.Div64(a.hi, a.lo+a.r, a.y)
		checkPair(t, "Div64", i, q, r, a.x, a.r)
		q, r = bits.Div64(a.hi, a.lo+a.r, a.x)
		checkPair(t, "Div64 symmetric", i, q, r, a.y, a.r)
	}
}

func TestRem32(t *testing.T) {
	// Check Rem32 against the remainder of Div32 for a dividend that
	// does not overflow the quotient.
	hi, lo, y := uint32(510510), uint32(9699690), uint32(510510+1) // hi < y
	for i := 0; i < 1000; i++ {
		r := bits.Rem32(hi, lo, y)
		_, r2 := bits.Div32(hi, lo, y)
		if r != r2 {
			checkValue(t, "Rem32", i, uint64(r), uint64(r2))
			return
		}
		y += 13
	}
}

func TestRem32Overflow(t *testing.T) {
	// Check Rem32 for a dividend that overflows the quotient.
	// Div32 panics on it, so the reference is Div64.
	hi, lo, y := uint32(510510), uint32(9699690), uint32(7) // y <= hi
	for i := 0; i < 1000; i++ {
		r := bits.Rem32(hi, lo, y)
		_, r2 := bits.Div64(0, uint64(hi)<<32|uint64(lo), uint64(y))
		if uint64(r) != r2 {
			checkValue(t, "Rem32 overflow", i, uint64(r), r2)
			return
		}
		y += 13
	}
}

func TestRem64(t *testing.T) {
	// Check Rem64 against the remainder of Div64 for a dividend that
	// does not overflow the quotient.
	hi, lo, y := uint64(510510), uint64(9699690), uint64(510510+1) // hi < y
	for i := 0; i < 1000; i++ {
		r := bits.Rem64(hi, lo, y)
		_, r2 := bits.Div64(hi, lo, y)
		if r != r2 {
			checkValue(t, "Rem64", i, r, r2)
			return
		}
		y += 13
	}
}

// rem64Case holds the remainder of (hi, lo) divided by y.
type rem64Case struct {
	hi, lo, y uint64
	rem       uint64
}

// The results come from Python 3: ((hi<<64)+lo) % y.
var rem64Tests = [...]rem64Case{
	{42, 1119, 42, 27},
	{42, 1119, 38, 9},
	{42, 1119, 26, 23},
	{469, 0, 467, 271},
	{469, 0, 113, 58},
	{111111, 111111, 1171, 803},
	{3968194946088682615, 3192705705065114702, 1000037, 56067},
}

func TestRem64Overflow(t *testing.T) {
	// Check Rem64 for a dividend that overflows the quotient.
	for i, rt := range rem64Tests {
		if rt.hi < rt.y {
			t.Errorf("rem64Tests[%d] does not overflow the quotient", i)
			return
		}
		checkValue(t, "Rem64 overflow", i, bits.Rem64(rt.hi, rt.lo, rt.y), rt.rem)
	}
}
