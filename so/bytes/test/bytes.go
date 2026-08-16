package bytes_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func toDot(r rune) rune {
	_ = r
	return '.'
}

func TestClone(t *testing.T) {
	clone := bytes.Clone(mem.System, []byte("abc"))
	if string(clone) != "abc" {
		t.Error("Clone(abc) != abc")
	}
	mem.FreeSlice(mem.System, clone)
}

func TestCompare(t *testing.T) {
	// The result is -1, 0 or +1, not the difference of the bytes.
	a := []byte("abc")
	if got := bytes.Compare(a, []byte("abc")); got != 0 {
		t.Errorf("Compare(abc, abc) = %d, want 0", got)
	}
	if got := bytes.Compare(a, []byte("xyz")); got != -1 {
		t.Errorf("Compare(abc, xyz) = %d, want -1", got)
	}
	if got := bytes.Compare([]byte("xyz"), a); got != +1 {
		t.Errorf("Compare(xyz, abc) = %d, want +1", got)
	}
}

func TestEqual(t *testing.T) {
	a := []byte("hello")
	if !bytes.Equal(a, []byte("hello")) {
		t.Error("Equal(hello, hello) = false")
	}
	if bytes.Equal(a, []byte("world")) {
		t.Error("Equal(hello, world) = true")
	}
}

func TestContains(t *testing.T) {
	b := []byte("seafood")
	if !bytes.Contains(b, []byte("foo")) {
		t.Error("Contains(seafood, foo) = false")
	}
	if bytes.Contains(b, []byte("bar")) {
		t.Error("Contains(seafood, bar) = true")
	}
}

func TestCount(t *testing.T) {
	b := []byte("cheese")
	if bytes.Count(b, []byte("e")) != 3 {
		t.Error("Count(cheese, e) != 3")
	}
	if bytes.Count(b, []byte("x")) != 0 {
		t.Error("Count(cheese, x) != 0")
	}
}

func TestCut(t *testing.T) {
	res := bytes.Cut([]byte("go is fun"), []byte(" is "))
	if string(res.Before) != "go" || string(res.After) != "fun" || !res.Found {
		t.Error("Cut(go is fun, ' is ') != (go, fun, true)")
	}
}

func TestHasPrefix(t *testing.T) {
	b := []byte("hello")
	if !bytes.HasPrefix(b, []byte("he")) {
		t.Error("HasPrefix(hello, he) = false")
	}
	if bytes.HasPrefix(b, []byte("lo")) {
		t.Error("HasPrefix(hello, lo) = true")
	}
}

func TestHasSuffix(t *testing.T) {
	b := []byte("hello")
	if !bytes.HasSuffix(b, []byte("lo")) {
		t.Error("HasSuffix(hello, lo) = false")
	}
	if bytes.HasSuffix(b, []byte("he")) {
		t.Error("HasSuffix(hello, he) = true")
	}
}

func TestIndex(t *testing.T) {
	b := []byte("hello")
	if bytes.Index(b, []byte("l")) != 2 {
		t.Error("Index(hello, l) != 2")
	}
	if bytes.IndexByte(b, 'e') != 1 {
		t.Error("IndexByte(hello, e) != 1")
	}
}

func TestJoin(t *testing.T) {
	parts := [][]byte{[]byte("go"), []byte("is"), []byte("fun")}
	joined := bytes.Join(mem.System, parts, []byte(" "))
	if string(joined) != "go is fun" {
		t.Error("Join(go is fun) failed")
	}
	mem.FreeSlice(mem.System, joined)
}

func TestMap(t *testing.T) {
	mapped := bytes.Map(mem.System, toDot, []byte("hello"))
	if string(mapped) != "....." {
		t.Error("Map(toDot, hello) != .....")
	}
	mem.FreeSlice(mem.System, mapped)
}

func TestRepeat(t *testing.T) {
	repeated := bytes.Repeat(mem.System, []byte("abc"), 3)
	if string(repeated) != "abcabcabc" {
		t.Error("Repeat(abc, 3) != abcabcabc")
	}
	mem.FreeSlice(mem.System, repeated)
}

func TestReplace(t *testing.T) {
	b := []byte("hello")
	r1 := bytes.Replace(mem.System, b, []byte("l"), []byte("x"), 1)
	if string(r1) != "hexlo" {
		t.Error("Replace(hello, l, x, 1) != hexlo")
	}
	mem.FreeSlice(mem.System, r1)
	r2 := bytes.Replace(mem.System, b, []byte("l"), []byte("x"), -1)
	if string(r2) != "hexxo" {
		t.Error("Replace(hello, l, x, -1) != hexxo")
	}
	mem.FreeSlice(mem.System, r2)
}

func TestRunes(t *testing.T) {
	runes := bytes.Runes(mem.System, []byte("fun"))
	defer mem.FreeSlice(mem.System, runes)
	if len(runes) != 3 {
		t.Fatal("Runes(fun) has wrong length")
		return
	}
	if runes[0] != 'f' || runes[1] != 'u' || runes[2] != 'n' {
		t.Error("Runes(fun) != [f u n]")
	}
}

func TestSplit(t *testing.T) {
	b := []byte("go is fun")
	s1 := bytes.Split(mem.System, b, []byte(" "))
	defer mem.FreeSlice(mem.System, s1)
	if len(s1) != 3 {
		t.Fatal("Split(go is fun) has wrong length")
		return
	}
	if string(s1[0]) != "go" || string(s1[1]) != "is" || string(s1[2]) != "fun" {
		t.Error("Split(go is fun) != [go is fun]")
	}

	s2 := bytes.SplitN(mem.System, b, []byte(" "), 2)
	defer mem.FreeSlice(mem.System, s2)
	if len(s2) != 2 {
		t.Fatal("SplitN(go is fun, 2) has wrong length")
		return
	}
	if string(s2[0]) != "go" || string(s2[1]) != "is fun" {
		t.Error("SplitN(go is fun, 2) != [go, is fun]")
	}
}

func TestTrim(t *testing.T) {
	b := []byte("  hello  ")
	if string(bytes.Trim(b, " ")) != "hello" {
		t.Error("Trim failed")
	}
	if string(bytes.TrimLeft(b, " ")) != "hello  " {
		t.Error("TrimLeft failed")
	}
	if string(bytes.TrimRight(b, " ")) != "  hello" {
		t.Error("TrimRight failed")
	}
}

func TestTrimPrefix(t *testing.T) {
	b := []byte("hello")
	if string(bytes.TrimPrefix(b, []byte("he"))) != "llo" {
		t.Error("TrimPrefix(hello, he) != llo")
	}
	if string(bytes.TrimSuffix(b, []byte("lo"))) != "hel" {
		t.Error("TrimSuffix(hello, lo) != hel")
	}
}

func TestToLower(t *testing.T) {
	lowered := bytes.ToLower(mem.System, []byte("Hello"))
	if string(lowered) != "hello" {
		t.Error("ToLower(Hello) != hello")
	}
	mem.FreeSlice(mem.System, lowered)
}

func TestToUpper(t *testing.T) {
	uppered := bytes.ToUpper(mem.System, []byte("Hello"))
	if string(uppered) != "HELLO" {
		t.Error("ToUpper(Hello) != HELLO")
	}
	mem.FreeSlice(mem.System, uppered)
}
