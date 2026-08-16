package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/math"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// partSep separates the wanted subslices of a split case. So has no nested
// composite literals in a table, so a case joins the wanted subslices into
// one string. No case holds the byte 0x02.
const partSep = "\x02"

// partAt returns the subslice with the index i of a joined want string.
func partAt(want string, i int) string {
	start := 0
	for range i {
		for start < len(want) && want[start] != partSep[0] {
			start++
		}
		start++
	}
	end := start
	for end < len(want) && want[end] != partSep[0] {
		end++
	}
	if start > len(want) {
		return ""
	}
	return want[start:end]
}

// A splitCase is a test case of Split and SplitN.
type splitCase struct {
	s     string
	sep   string
	n     int
	want  string // the wanted subslices, joined with partSep
	parts int    // the number of wanted subslices
}

var splitCases = []splitCase{
	{"", "", -1, "", 0},
	{"abcd", "a", 0, "", 0},
	{"abcd", "", 2, "a\x02bcd", 2},
	{"abcd", "a", -1, "\x02bcd", 2},
	{"abcd", "z", -1, "abcd", 1},
	{"abcd", "", -1, "a\x02b\x02c\x02d", 4},
	{"1,2,3,4", ",", -1, "1\x022\x023\x024", 4},
	{"1....2....3....4", "...", -1, "1\x02.2\x02.3\x02.4", 4},
	{"☺☻☹", "☹", -1, "☺☻\x02", 2},
	{"☺☻☹", "~", -1, "☺☻☹", 1},
	{"☺☻☹", "", -1, "☺\x02☻\x02☹", 3},
	{"1 2 3 4", " ", 3, "1\x022\x023 4", 3},
	{"1 2", " ", 3, "1\x022", 2},
	{"123", "", 2, "1\x0223", 2},
	{"123", "", 17, "1\x022\x023", 3},
	{"bT", "T", math.MaxInt / 4, "b\x02", 2},
	// An invalid byte is one subslice of its own.
	{"\xff-\xff", "", -1, "\xff\x02-\x02\xff", 3},
	{"\xff-\xff", "-", -1, "\xff\x02\xff", 2},
}

// splitDiff compares the subslices with the joined want string. It returns
// the index of the first wrong subslice, or -1 if every subslice is correct.
func splitDiff(got [][]byte, want string, parts int) int {
	if len(got) != parts {
		return len(got)
	}
	for i, p := range got {
		if string(p) != partAt(want, i) {
			return i
		}
	}
	return -1
}

func TestSplitN(t *testing.T) {
	alloc := t.Allocator()
	for i, tc := range splitCases {
		got := bytes.SplitN(alloc, []byte(tc.s), []byte(tc.sep), tc.n)
		if bad := splitDiff(got, tc.want, tc.parts); bad >= 0 {
			var d [64]byte
			if bad >= len(got) {
				t.Errorf("SplitN() case %d gave %d subslices, want %d",
					i, len(got), tc.parts)
			} else {
				t.Errorf("SplitN() case %d subslice %d = %s, want %s",
					i, bad, dump(d[:], got[bad]), partAt(tc.want, bad))
			}
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestSplit(t *testing.T) {
	// Split gives the same subslices as SplitN with a negative count.
	alloc := t.Allocator()
	for i, tc := range splitCases {
		if tc.n >= 0 {
			continue
		}
		got := bytes.Split(alloc, []byte(tc.s), []byte(tc.sep))
		if bad := splitDiff(got, tc.want, tc.parts); bad >= 0 {
			var d [64]byte
			if bad >= len(got) {
				t.Errorf("Split() case %d gave %d subslices, want %d",
					i, len(got), tc.parts)
			} else {
				t.Errorf("Split() case %d subslice %d = %s, want %s",
					i, bad, dump(d[:], got[bad]), partAt(tc.want, bad))
			}
		}
		mem.FreeSlice(alloc, got)
	}
}

func TestSplitCount(t *testing.T) {
	// A positive count gives at most n subslices, and the last subslice holds
	// the rest of s.
	alloc := t.Allocator()
	const s = "a,b,c,d"
	for n := 1; n <= 6; n++ {
		got := bytes.SplitN(alloc, []byte(s), []byte(","), n)
		want := min(n, 4)
		if len(got) != want {
			t.Errorf("SplitN(%s, comma, %d) gave %d subslices, want %d",
				s, n, len(got), want)
		}
		mem.FreeSlice(alloc, got)
	}
	// A count of 0 gives no subslices.
	got := bytes.SplitN(alloc, []byte(s), []byte(","), 0)
	if len(got) != 0 {
		t.Errorf("SplitN(%s, comma, 0) gave %d subslices, want 0", s, len(got))
	}
	mem.FreeSlice(alloc, got)
}

func TestSplitView(t *testing.T) {
	// The subslices are views into s, not copies.
	var buf [7]byte
	s := buf[:]
	copy(s, "a,b,c,d")
	alloc := t.Allocator()
	got := bytes.Split(alloc, s, []byte(","))
	defer mem.FreeSlice(alloc, got)
	if len(got) != 4 {
		t.Errorf("Split() gave %d subslices, want 4", len(got))
		return
	}
	got[1][0] = 'B'
	if string(s) != "a,B,c,d" {
		t.Errorf("Split copies the bytes: %s", string(s))
	}
}

func TestSplitCap(t *testing.T) {
	// A subslice has the capacity of its length, so an append to a subslice
	// does not change the next subslice.
	alloc := t.Allocator()
	got := bytes.Split(alloc, []byte("a,b,c"), []byte(","))
	defer mem.FreeSlice(alloc, got)
	for i, p := range got {
		if cap(p) != len(p) {
			t.Errorf("subslice %d has cap %d, want %d", i, cap(p), len(p))
		}
	}
}

func TestSplitJoin(t *testing.T) {
	// Join undoes Split.
	alloc := t.Allocator()
	for i, tc := range splitCases {
		if tc.n != -1 || tc.parts == 0 {
			continue
		}
		parts := bytes.Split(alloc, []byte(tc.s), []byte(tc.sep))
		joined := bytes.Join(alloc, parts, []byte(tc.sep))
		if string(joined) != tc.s {
			var d [64]byte
			t.Errorf("Join(Split()) case %d = %s, want %s",
				i, dump(d[:], joined), tc.s)
		}
		mem.FreeSlice(alloc, joined)
		mem.FreeSlice(alloc, parts)
	}
}

// The sweep of TestSplitSweep. Split and Join allocate, and the freestanding
// heap is a bump allocator that reclaims nothing before the end of a test, so
// the split sweep is shorter than the search sweeps of bytes.go.
var splitSweeps = []sweep{
	{"ab", 5, 2},
	{"\x00a\xff", 3, 2},
}

// splitMaxWord is the length of the longest word of the split sweeps.
const splitMaxWord = 5

func TestSplitSweep(t *testing.T) {
	// Split gives one more subslice than the number of separators, no
	// subslice holds the separator, and Join undoes the split.
	alloc := t.Allocator()
	for _, sw := range splitSweeps {
		for wi := 0; wi < wordTotal(sw.alpha, sw.maxWord); wi++ {
			var sbuf [splitMaxWord]byte
			s := wordAt(sbuf[:], sw.alpha, sw.maxWord, wi)
			for pi := 1; pi < wordTotal(sw.alpha, sw.maxSep); pi++ {
				var pbuf [splitMaxWord]byte
				sep := wordAt(pbuf[:], sw.alpha, sw.maxSep, pi)
				parts := bytes.Split(alloc, s, sep)
				want := countBrute(s, sep) + 1
				if len(parts) != want {
					var d1, d2 [2 * splitMaxWord]byte
					t.Errorf("Split(%s, %s) gave %d subslices, want %d",
						dump(d1[:], s), dump(d2[:], sep), len(parts), want)
					mem.FreeSlice(alloc, parts)
					return
				}
				for i, p := range parts {
					if bytes.Index(p, sep) >= 0 {
						var d1, d2 [2 * splitMaxWord]byte
						t.Errorf("Split(%s, %s) subslice %d holds the separator",
							dump(d1[:], s), dump(d2[:], sep), i)
						mem.FreeSlice(alloc, parts)
						return
					}
				}
				joined := bytes.Join(alloc, parts, sep)
				if !bytes.Equal(joined, s) {
					var d1, d2, d3 [2 * splitMaxWord]byte
					t.Errorf("Join(Split(%s, %s)) = %s",
						dump(d1[:], s), dump(d2[:], sep), dump(d3[:], joined))
				}
				mem.FreeSlice(alloc, joined)
				mem.FreeSlice(alloc, parts)
			}
		}
	}
}

func TestSplitEmptySep(t *testing.T) {
	// An empty separator gives one subslice per code point. An invalid byte
	// is one subslice of its own.
	alloc := t.Allocator()
	const s = "a☺b\xffc"
	parts := bytes.Split(alloc, []byte(s), nil)
	defer mem.FreeSlice(alloc, parts)
	if len(parts) != 5 {
		t.Errorf("Split(%s, empty) gave %d subslices, want 5", s, len(parts))
		return
	}
	for i, p := range parts {
		r, size := utf8.DecodeRune(p)
		if size != len(p) {
			var d [16]byte
			t.Errorf("subslice %d = %s holds %d bytes of a code point, want %d",
				i, dump(d[:], p), len(p), size)
		}
		if r == utf8.RuneError && size == 1 && i != 3 {
			var d [16]byte
			t.Errorf("subslice %d = %s is invalid", i, dump(d[:], p))
		}
	}
}
