package rand_test

import (
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
)

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

// checkValue checks one result against the wanted value and reports whether
// the check passed. The name and the index identify the failed check.
func checkValue(t *testing.T, name string, i int, got, want uint64) bool {
	if got == want {
		return true
	}
	var b1, b2 [16]byte
	t.Errorf("%s case %d = %s, want %s", name, i, hex(b1[:], got, 16), hex(b2[:], want, 16))
	return false
}

// checkBelow checks that one result is in the range [0,n) and reports whether
// the check passed. A negative result becomes a large uint64, so the check
// rejects it. The name and the index identify the failed check.
func checkBelow(t *testing.T, name string, i int, got, n uint64) bool {
	if got < n {
		return true
	}
	var b1, b2 [16]byte
	t.Errorf("%s case %d = %s, want below %s", name, i, hex(b1[:], got, 16), hex(b2[:], n, 16))
	return false
}

// checkBytes checks an encoded value against the wanted bytes. The name
// identifies the failed check. so/fmt has no %q verb, so the message gives the
// index and the value of the first byte that differs.
func checkBytes(t *testing.T, name string, got []byte, want string) {
	if string(got) == want {
		return
	}
	if len(got) != len(want) {
		t.Errorf("%s length = %d, want %d", name, len(got), len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			var b1, b2 [2]byte
			t.Errorf("%s byte %d = %s, want %s", name, i,
				hex(b1[:], uint64(got[i]), 2), hex(b2[:], uint64(want[i]), 2))
			return
		}
	}
}

// errText returns the message of an error, or "nil" for no error.
func errText(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}

// fixed is a Source that repeats one value. A test that uses it must avoid
// the rejection loop of Uint64N, which never ends for a constant source.
type fixed struct {
	val uint64
}

func (f *fixed) Uint64() uint64 { return f.val }

// pcgWant holds the first values of a PCG seeded with 1 and 2.
var pcgWant = []uint64{
	0xc4f5a58656eef510,
	0x9dcec3ad077dec6c,
	0xc8d04605312f8088,
	0xcbedc0dcb63ac19a,
	0x3bf98798cae97950,
	0xa8c6d7f8d485abc,
	0x7ffa3780429cd279,
	0x730ad2626b1c2f8e,
	0x21ff2330f4a0ad99,
	0x2f0901a1947094b0,
	0xa9735a3cfbe36cef,
	0x71ddb0a01a12c84a,
	0xf0e53e77a78453bb,
	0x1f173e9663be1e9d,
	0x657651da3ac4115e,
	0xc8987376b65a157b,
	0xbb17008f5fca28e7,
	0x8232bd645f29ed22,
	0x12be8f07ad14c539,
	0x54908a48e8e4736e,
}

// srcValues holds the source values of the derivation test.
var srcValues = []uint64{
	0,
	1,
	0x0123456789abcdef,
	0x8000000000000000,
	0xffffffffffffffff,
}

// The bounds of the range tests. Every list has powers of two and the values
// next to them. A power of two takes the mask path of the generator, and the
// value next to it takes the rejection path.
var (
	int32Bounds  = []int32{1, 2, 10, 32, 1 << 20, 1<<20 + 1, 1000000000, 1 << 30, 1<<31 - 2, 1<<31 - 1}
	uint32Bounds = []uint32{1, 2, 10, 32, 1 << 20, 1<<20 + 1, 1000000000, 1 << 31, 1<<32 - 2, 1<<32 - 1}
	int64Bounds  = []int64{1, 2, 10, 1 << 20, 1000000000, 1 << 30, 1<<31 - 1, 1000000000000000000, 1 << 60, 1<<63 - 2, 1<<63 - 1}
	uint64Bounds = []uint64{1, 2, 10, 1 << 20, 1000000000, 1 << 30, 1<<31 - 1, 1000000000000000000, 1 << 60, 1<<63 - 1, 1<<64 - 2, 1<<64 - 1}
)

// draws is the number of values that a range test takes for one bound.
const draws = 20

func TestPCG(t *testing.T) {
	p := rand.NewPCG(1, 2)
	for i, want := range pcgWant {
		if !checkValue(t, "PCG", i, p.Uint64(), want) {
			return
		}
	}
}

func TestPCGZero(t *testing.T) {
	// A zero PCG behaves the same way as NewPCG(0, 0).
	var p rand.PCG
	q := rand.NewPCG(0, 0)
	for i := 0; i < 10; i++ {
		if !checkValue(t, "zero PCG", i, p.Uint64(), q.Uint64()) {
			return
		}
	}
}

func TestPCGSeed(t *testing.T) {
	p := rand.NewPCG(3, 4)
	for i := 0; i < 5; i++ {
		p.Uint64()
	}
	// Seed restarts the stream of NewPCG(1, 2).
	p.Seed(1, 2)
	for i, want := range pcgWant {
		if !checkValue(t, "PCG after Seed", i, p.Uint64(), want) {
			return
		}
	}
}

func TestPCGMarshal(t *testing.T) {
	const (
		seed1      = 0x123456789abcdef0
		seed2      = 0xfedcba9876543210
		want       = "pcg:\x12\x34\x56\x78\x9a\xbc\xde\xf0\xfe\xdc\xba\x98\x76\x54\x32\x10"
		wantAppend = "\x00\x00\x00\x00" + want
	)
	var p rand.PCG
	p.Seed(seed1, seed2)

	buf := make([]byte, 20)
	data, err := p.MarshalBinary(buf)
	if err != nil {
		t.Errorf("MarshalBinary(): %s", err.Error())
		return
	}
	checkBytes(t, "MarshalBinary()", data, want)

	buf = make([]byte, 4, 32)
	data, err = p.AppendBinary(buf)
	if err != nil {
		t.Errorf("AppendBinary(): %s", err.Error())
		return
	}
	checkBytes(t, "AppendBinary()", data, wantAppend)

	var q rand.PCG
	if err := q.UnmarshalBinary([]byte(want)); err != nil {
		t.Errorf("UnmarshalBinary(): %s", err.Error())
		return
	}
	// The round trip gives a generator with the same stream.
	for i := 0; i < 10; i++ {
		if !checkValue(t, "PCG after round trip", i, q.Uint64(), p.Uint64()) {
			return
		}
	}
}

func TestPCGUnmarshalError(t *testing.T) {
	const good = "pcg:\x12\x34\x56\x78\x9a\xbc\xde\xf0\xfe\xdc\xba\x98\x76\x54\x32\x10"
	bad := []string{
		"",
		good[:19],
		good + "\x00",
		"PCG:" + good[4:],
	}
	p := rand.NewPCG(1, 2)
	for i, s := range bad {
		if err := p.UnmarshalBinary([]byte(s)); err != rand.ErrUnmarshalPCG {
			t.Errorf("UnmarshalBinary() case %d = %s, want ErrUnmarshalPCG", i, errText(err))
		}
	}
	// A rejected value leaves the generator alone.
	for i, want := range pcgWant {
		if !checkValue(t, "PCG after error", i, p.Uint64(), want) {
			return
		}
	}
}

func TestFromSource(t *testing.T) {
	// Every method takes one value from the source and cuts it to the type.
	for i, v := range srcValues {
		src := fixed{val: v}
		r := rand.New(&src)
		checkValue(t, "Uint64", i, r.Uint64(), v)
		checkValue(t, "Int64", i, uint64(r.Int64()), v&^(1<<63))
		checkValue(t, "Uint32", i, uint64(r.Uint32()), v>>32)
		checkValue(t, "Int32", i, uint64(r.Int32()), v>>33)
		checkValue(t, "Uint", i, uint64(r.Uint()), uint64(uint(v)))
		checkValue(t, "Int", i, uint64(r.Int()), uint64(uint(v)<<1>>1))
	}
}

func TestShared(t *testing.T) {
	// Two Rand values over one source take from one stream.
	p := rand.NewPCG(1, 2)
	r1 := rand.New(&p)
	r2 := rand.New(&p)
	for i, want := range pcgWant {
		var got uint64
		if i%2 == 0 {
			got = r1.Uint64()
		} else {
			got = r2.Uint64()
		}
		if !checkValue(t, "shared source", i, got, want) {
			return
		}
	}
}

func TestPowerOfTwo(t *testing.T) {
	// A bound that is a power of two masks the value of the source.
	const v = 0x0123456789abcdef
	for k := 1; k < 64; k++ {
		n := uint64(1) << uint(k)
		src := fixed{val: v}
		r := rand.New(&src)
		if !checkValue(t, "Uint64N", k, r.Uint64N(n), v&(n-1)) {
			return
		}
	}
}

func TestUint64N(t *testing.T) {
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range uint64Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "Uint64N", i, r.Uint64N(n), n) {
				return
			}
		}
	}
}

func TestInt64N(t *testing.T) {
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range int64Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "Int64N", i, uint64(r.Int64N(n)), uint64(n)) {
				return
			}
		}
	}
}

func TestUint32N(t *testing.T) {
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range uint32Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "Uint32N", i, uint64(r.Uint32N(n)), uint64(n)) {
				return
			}
		}
	}
}

func TestInt32N(t *testing.T) {
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range int32Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "Int32N", i, uint64(r.Int32N(n)), uint64(n)) {
				return
			}
		}
	}
}

func TestUintN(t *testing.T) {
	// uint is 32 bits on a freestanding target, so the test uses the uint32
	// bounds.
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range uint32Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "UintN", i, uint64(r.UintN(uint(n))), uint64(n)) {
				return
			}
		}
	}
}

func TestIntN(t *testing.T) {
	// int is 32 bits on a freestanding target, so the test uses the int32
	// bounds.
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	for i, n := range int32Bounds {
		for j := 0; j < draws; j++ {
			if !checkBelow(t, "IntN", i, uint64(r.IntN(int(n))), uint64(n)) {
				return
			}
		}
	}
}

func TestUniform(t *testing.T) {
	// A bound of 10 does not divide 2⁶⁴, so IntN rejects some values of the
	// source to keep the results uniform. The bound of the count is wide,
	// because the test checks the shape of the distribution, not the stream.
	const buckets = 10
	const total = 10000
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	var count [buckets]int
	for i := 0; i < total; i++ {
		count[r.IntN(buckets)]++
	}
	for i := 0; i < buckets; i++ {
		if count[i] < 800 || count[i] > 1200 {
			t.Errorf("bucket %d = %d, want near %d", i, count[i], total/buckets)
		}
	}
}

func TestFloat64(t *testing.T) {
	const total = 100000
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	sum := 0.0
	for i := 0; i < total; i++ {
		f := r.Float64()
		if f < 0 || f >= 1 {
			t.Errorf("Float64() case %d = %g, want [0,1)", i, f)
			return
		}
		sum += f
	}
	if mean := sum / total; mean < 0.49 || mean > 0.51 {
		t.Errorf("Float64() mean = %g, want near 0.5", mean)
	}
}

func TestFloat32(t *testing.T) {
	// The rounding of the division gave a result of 1 for some values of the
	// source. See https://go.dev/issue/6721.
	const total = 100000
	p := rand.NewPCG(1, 2)
	r := rand.New(&p)
	sum := 0.0
	for i := 0; i < total; i++ {
		f := r.Float32()
		if f < 0 || f >= 1 {
			t.Errorf("Float32() case %d = %g, want [0,1)", i, float64(f))
			return
		}
		sum += float64(f)
	}
	if mean := sum / total; mean < 0.49 || mean > 0.51 {
		t.Errorf("Float32() mean = %g, want near 0.5", mean)
	}
}

func TestGlobal(t *testing.T) {
	// The global generator takes a random seed, so the test checks the range
	// of the results.
	if rand.Int64() < 0 {
		t.Error("negative Int64()")
	}
	if rand.Int32() < 0 {
		t.Error("negative Int32()")
	}
	if rand.Int() < 0 {
		t.Error("negative Int()")
	}

	checkBelow(t, "Uint64N", 0, rand.Uint64N(1000), 1000)
	checkBelow(t, "Uint32N", 0, uint64(rand.Uint32N(1000)), 1000)
	checkBelow(t, "UintN", 0, uint64(rand.UintN(1000)), 1000)
	checkBelow(t, "Int64N", 0, uint64(rand.Int64N(1000)), 1000)
	checkBelow(t, "Int32N", 0, uint64(rand.Int32N(1000)), 1000)
	checkBelow(t, "IntN", 0, uint64(rand.IntN(1000)), 1000)

	if f := rand.Float64(); f < 0 || f >= 1 {
		t.Errorf("Float64() = %g, want [0,1)", f)
	}
	if f := rand.Float32(); f < 0 || f >= 1 {
		t.Errorf("Float32() = %g, want [0,1)", float64(f))
	}
}

func TestGlobalStream(t *testing.T) {
	// The global functions take from one stream, so two values of 64 bits are
	// almost never the same.
	if rand.Uint64() == rand.Uint64() {
		t.Error("two Uint64() in a row are the same")
	}
	if rand.Uint64() == rand.Uint64() {
		t.Error("two Uint64() in a row are the same")
	}
}
