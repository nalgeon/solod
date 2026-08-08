package strings_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestClone(t *testing.T) {
	s := "hello"
	c := strings.Clone(mem.System, s)
	defer mem.FreeString(mem.System, c)
	if c != s {
		t.Error("Clone failed")
	}
}

func TestCompare(t *testing.T) {
	if strings.Compare("abc", "abb") <= 0 {
		t.Error("Compare failed")
	}
	if strings.Compare("abc", "abd") >= 0 {
		t.Error("Compare failed")
	}
	if strings.Compare("abc", "abc") != 0 {
		t.Error("Compare failed")
	}
}

func TestCount(t *testing.T) {
	n := strings.Count("hello world", "o")
	if n != 2 {
		t.Error("Count failed")
	}
	n = strings.Count("hello world", "")
	if n != 12 {
		t.Error("Count failed")
	}
}

func TestCut(t *testing.T) {
	before, after := strings.Cut("hello world", " ")
	if before != "hello" || after != "world" {
		t.Error("Cut failed")
	}
}

func TestCutPrefixSuffix(t *testing.T) {
	src := "hello world"
	if s, ok := strings.CutPrefix(src, "hello"); !ok || s != " world" {
		t.Error("CutPrefix failed")
	}
	if s, ok := strings.CutSuffix(src, "world"); !ok || s != "hello " {
		t.Error("CutSuffix failed")
	}
}

func TestIndex(t *testing.T) {
	idx := strings.Index("hello world", "o")
	if idx != 4 {
		t.Error("Index failed")
	}
	idx = strings.IndexAny("hello world", "ow")
	if idx != 4 {
		t.Error("IndexAny failed")
	}
	idx = strings.IndexByte("hello world", 'o')
	if idx != 4 {
		t.Error("IndexByte failed")
	}
}

func TestHasPrefixSuffix(t *testing.T) {
	src := "hello world"
	if !strings.HasPrefix(src, "hello") || strings.HasPrefix(src, "world") {
		t.Error("HasPrefix failed")
	}
	if !strings.HasSuffix(src, "world") || strings.HasSuffix(src, "hello") {
		t.Error("HasSuffix failed")
	}
}

func TestTrimPrefixSuffix(t *testing.T) {
	src := "hello world"
	if s := strings.TrimPrefix(src, "hello "); s != "world" {
		t.Error("TrimPrefix failed")
	}
	if s := strings.TrimSuffix(src, " world"); s != "hello" {
		t.Error("TrimSuffix failed")
	}
}

func TestRepeat(t *testing.T) {
	r := strings.Repeat(mem.System, "abc", 3)
	defer mem.FreeString(mem.System, r)
	if r != "abcabcabc" {
		t.Error("Repeat failed")
	}
}

func TestReplace(t *testing.T) {
	s := "hello world"
	r := strings.Replace(mem.System, s, "o", "0", 1)
	if r != "hell0 world" {
		t.Error("Replace failed")
	}
	mem.FreeString(mem.System, r)
	r = strings.ReplaceAll(mem.System, s, "o", "0")
	if r != "hell0 w0rld" {
		t.Error("ReplaceAll failed")
	}
	mem.FreeString(mem.System, r)
}

func TestSplitJoin(t *testing.T) {
	s := "a,b,c"
	parts := strings.Split(mem.System, s, ",")
	defer mem.FreeSlice(mem.System, parts)
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Error("Split failed")
	}
	j := strings.Join(mem.System, parts, ",")
	defer mem.FreeString(mem.System, j)
	if j != s {
		t.Error("Join failed")
	}
}

func TestToUpperLower(t *testing.T) {
	s := "Hello, 世界!"
	u := strings.ToUpper(mem.System, s)
	if u != "HELLO, 世界!" {
		t.Error("ToUpper failed")
	}
	mem.FreeString(mem.System, u)
	l := strings.ToLower(mem.System, s)
	if l != "hello, 世界!" {
		t.Error("ToLower failed")
	}
	mem.FreeString(mem.System, l)
}

func TestTrim(t *testing.T) {
	s := "  hello world  "
	trimmed := strings.TrimSpace(s)
	if trimmed != "hello world" {
		t.Error("TrimSpace failed")
	}
	trimmed = strings.Trim(s, " dh")
	if trimmed != "ello worl" {
		t.Error("Trim failed")
	}
}

func TestBuilder(t *testing.T) {
	b := strings.NewBuilder(mem.System)
	defer b.Free()
	b.WriteString("Hello")
	b.WriteByte(',')
	b.WriteRune(' ')
	b.WriteString("world")
	s := b.String()
	if s != "Hello, world" {
		t.Error("Builder failed")
	}
}

func TestReader(t *testing.T) {
	r := strings.NewReader("hello world")
	buf := make([]byte, 5)
	n, err := r.Read(buf)
	if err != nil || n != 5 || string(buf) != "hello" {
		t.Error("Reader failed")
	}
}
