// This case checks the promise of doc/freestanding.md: every package the
// document lists must compile, link and run with no C standard library.
//
// The compiler emits every function of an imported package, and bare mode links
// with --no-gc-sections, so one libc call anywhere in alloc listed package makes the
// link fail. Run it with `make run-case name=freestanding mode=bare`.
package main

import (
	"unsafe"

	"solod.dev/so/bufio"
	"solod.dev/so/bytealg"
	"solod.dev/so/bytes"
	"solod.dev/so/c"
	"solod.dev/so/cmp"
	"solod.dev/so/encoding/binary"
	"solod.dev/so/encoding/hex"
	"solod.dev/so/encoding/json"
	"solod.dev/so/errors"
	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/maps"
	"solod.dev/so/math/bits"
	"solod.dev/so/math/rand"
	"solod.dev/so/mem"
	"solod.dev/so/net/netip"
	"solod.dev/so/path"
	"solod.dev/so/runtime"
	"solod.dev/so/slices"
	"solod.dev/so/strconv"
	"solod.dev/so/strings"
	"solod.dev/so/time"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

var ErrCheck = errors.New("check failed")

func check(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func main() {
	heap := make([]byte, 4096)
	arena := mem.NewArena(heap)
	var alloc mem.Allocator = &arena

	{
		// bufio, bytes, io
		src := bytes.NewReader([]byte("hello"))
		br := bufio.NewReaderSize(alloc, &src, 64)
		b, err := br.ReadByte()
		check(err == nil, "bufio: read failed")
		check(b == 'h', "bufio: wrong byte")
		br.Free()

		w := bytes.NewBuffer(alloc, nil)
		n, err := io.WriteString(&w, "hello")
		check(err == nil, "io: write failed")
		check(n == 5, "io: wrong count")
		check(string(w.Bytes()) == "hello", "io: wrong text")
		w.Free()

		check(bytes.Equal([]byte("ab"), []byte("ab")), "bytes: not equal")
	}
	{
		// bytealg, strings, unicode, unicode/utf8
		check(bytealg.IndexByteString("hello", 'e') == 1, "bytealg: wrong index")
		check(strings.Contains("hello", "ell"), "strings: no substring")
		check(unicode.IsDigit('5'), "unicode: not alloc digit")
		check(utf8.RuneCountInString("héllo") == 5, "utf8: wrong rune count")
	}
	{
		// c, unsafe
		check(c.Sizeof[int32]() == 4, "c: wrong size")
		check(unsafe.Sizeof(int32(0)) == 4, "unsafe: wrong size")
	}
	{
		// cmp, slices
		check(cmp.Compare(1, 2) < 0, "cmp: wrong order")
		nums := []int{3, 1, 2}
		slices.Sort(nums)
		check(nums[0] == 1, "slices: not sorted")
		check(slices.Contains(nums, 3), "slices: no value")
	}
	{
		// encoding/binary, encoding/hex
		var le binary.LE
		word := []byte{0, 0, 0, 0}
		le.PutUint32(word, 0x01020304)
		check(le.Uint32(word) == 0x01020304, "binary: wrong value")

		dst := make([]byte, hex.EncodedLen(len(word)))
		hex.Encode(dst, word)
		check(string(dst) == "04030201", "hex: wrong encoding")
	}
	{
		// encoding/json
		out := make([]byte, 64)
		sb := strings.FixedBuilder(out)
		enc := json.NewEncoder(&sb)
		enc.BeginObject()
		enc.Str("pi")
		enc.Float(3.5)
		enc.EndObject()
		enc.Flush()
		check(enc.Err() == nil, "json: encode failed")
		check(sb.String() == `{"pi":3.5}`, "json: wrong text")
	}
	{
		// errors
		var err = ErrCheck
		check(err.Error() == "check failed", "errors: wrong text")
	}
	{
		// fmt. The print family works. Print, Println and Printf drop the
		// bytes, because a freestanding host has no standard output. The scan
		// family panics.
		buf := make([]byte, 32)
		text := fmt.Sprintf(buf, "n=%d", 42)
		check(text == "n=42", "fmt: wrong text")

		out := make([]byte, 32)
		sb := strings.FixedBuilder(out)
		cnt, err := fmt.Fprintf(&sb, "%s=%d", "n", 42)
		check(err == nil, "fmt: write failed")
		check(cnt == 4, "fmt: wrong write count")
		check(sb.String() == "n=42", "fmt: wrong output")

		cnt, err = fmt.Printf("%d\n", 42)
		check(err == nil, "fmt: print failed")
		check(cnt == 3, "fmt: wrong print count")
	}
	{
		// maps
		m := maps.New[string, int](alloc, 8)
		m.Set("one", 1)
		check(m.Len() == 1, "maps: wrong length")
		check(m.Get("one") == 1, "maps: wrong value")
		m.Free()
	}
	{
		// math/bits, math/rand, runtime
		check(bits.Len(255) == 8, "bits: wrong length")
		check(rand.Uint64() != rand.Uint64(), "rand: repeated value")
		check(runtime.NumCPU() >= 1, "runtime: no CPU")
		check(runtime.Seed() != 0, "runtime: zero seed")
	}
	{
		// mem
		p := mem.Alloc[int](alloc)
		*p = 42
		check(*p == 42, "mem: wrong value")
		mem.Free(alloc, p)
	}
	{
		// net/netip. A numeric zone works in freestanding mode,
		// but an interface name resolves to no zone.
		ip, err := netip.ParseAddr("fe80::1%2")
		check(err == nil, "netip: parse failed")
		buf := make([]byte, netip.MaxZoneLen)
		check(ip.Zone(buf) == "2", "netip: wrong zone")
		named := ip.WithZone("eth0")
		check(named.Zone(buf) == "", "netip: name resolved to a zone")
	}
	{
		// path
		check(path.Base("/dir/file.txt") == "file.txt", "path: wrong base")
		check(path.Ext("/dir/file.txt") == ".txt", "path: wrong extension")
	}
	{
		// strconv
		buf := make([]byte, 32)
		check(strconv.Itoa(buf, -42) == "-42", "strconv: wrong text")
		n, err := strconv.Atoi("42")
		check(err == nil, "strconv: parse failed")
		check(n == 42, "strconv: wrong number")
	}
	{
		// time. Now, Since and Until are missing in freestanding mode,
		// and Format only supports named layouts.
		ts := time.Date(2026, time.August, 11, 12, 0, 0, 0, time.UTC)
		check(ts.Year() == 2026, "time: wrong year")
		buf := make([]byte, time.RFC3339Len+1)
		got := ts.Format(buf, time.RFC3339, time.UTC)
		check(got == "2026-08-11T12:00:00Z", "time: wrong text")
	}

	println("ok")
}
