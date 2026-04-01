package main

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
)

func toDot(r rune) rune {
	_ = r
	return '.'
}

func main() {
	{
		// Clone.
		b := []byte("abc")
		clone := bytes.Clone(nil, b)
		if string(clone) != "abc" {
			panic("Clone failed")
		}
		mem.FreeSlice(nil, clone)
	}
	{
		// Compare and Equal.
		a := []byte("abc")
		b := []byte("abc")
		c := []byte("xyz")
		if bytes.Compare(a, b) != 0 {
			panic("Compare failed: a != b")
		}
		if bytes.Compare(a, c) >= 0 {
			panic("Compare failed: a >= c")
		}
		if bytes.Compare(c, a) <= 0 {
			panic("Compare failed: c <= a")
		}
		if !bytes.Equal(a, b) {
			panic("Equal failed: a != b")
		}
		if bytes.Equal(a, c) {
			panic("Equal failed: a == c")
		}
	}
	{
		// Contains.
		b := []byte("seafood")
		if !bytes.Contains(b, []byte("foo")) {
			panic("Contains failed")
		}
		if bytes.Contains(b, []byte("bar")) {
			panic("Contains failed")
		}
	}
	{
		// Count.
		b := []byte("cheese")
		if bytes.Count(b, []byte("e")) != 3 {
			panic("Count failed")
		}
		if bytes.Count(b, []byte("x")) != 0 {
			panic("Count failed")
		}
	}
	{
		// Cut.
		b := []byte("go is fun")
		res := bytes.Cut(b, []byte(" is "))
		if string(res.Before) != "go" || string(res.After) != "fun" || !res.Found {
			panic("Cut failed")
		}
	}
	{
		// Equal.
		b := []byte("hello")
		if !bytes.Equal(b, []byte("hello")) {
			panic("Equal failed")
		}
		if bytes.Equal(b, []byte("world")) {
			panic("Equal failed")
		}
	}
	{
		// HasPrefix and HasSuffix.
		b := []byte("hello")
		if !bytes.HasPrefix(b, []byte("he")) {
			panic("HasPrefix failed")
		}
		if bytes.HasPrefix(b, []byte("lo")) {
			panic("HasPrefix failed")
		}
		if !bytes.HasSuffix(b, []byte("lo")) {
			panic("HasSuffix failed")
		}
		if bytes.HasSuffix(b, []byte("he")) {
			panic("HasSuffix failed")
		}
	}
	{
		// Index, IndexByte.
		b := []byte("hello")
		if bytes.Index(b, []byte("l")) != 2 {
			panic("Index failed")
		}
		if bytes.IndexByte(b, 'e') != 1 {
			panic("Index failed")
		}
	}
	{
		// Join.
		b1 := []byte("go")
		b2 := []byte("is")
		b3 := []byte("fun")
		joined := bytes.Join(nil, [][]byte{b1, b2, b3}, []byte(" "))
		if string(joined) != "go is fun" {
			panic("Join failed")
		}
		mem.FreeSlice(nil, joined)
	}
	{
		// Map.
		b := []byte("hello")
		mapped := bytes.Map(nil, toDot, b)
		if string(mapped) != "....." {
			panic("Map failed")
		}
		mem.FreeSlice(nil, mapped)
	}
	{
		// Repeat.
		b := []byte("abc")
		repeated := bytes.Repeat(nil, b, 3)
		if string(repeated) != "abcabcabc" {
			panic("Repeat failed")
		}
		mem.FreeSlice(nil, repeated)
	}
	{
		// Replace.
		b := []byte("hello")
		r1 := bytes.Replace(nil, b, []byte("l"), []byte("x"), 1)
		if string(r1) != "hexlo" {
			panic("Replace failed")
		}
		mem.FreeSlice(nil, r1)
		r2 := bytes.Replace(nil, b, []byte("l"), []byte("x"), -1)
		if string(r2) != "hexxo" {
			panic("ReplaceAll failed")
		}
		mem.FreeSlice(nil, r2)
	}
	{
		// Runes.
		b := []byte("fun")
		runes := bytes.Runes(nil, b)
		if len(runes) != 3 {
			panic("Runes failed")
		}
		if runes[0] != 'f' || runes[1] != 'u' || runes[2] != 'n' {
			panic("Runes failed")
		}
		mem.FreeSlice(nil, runes)
	}
	{
		// Split and SplitN.
		b := []byte("go is fun")
		s1 := bytes.Split(nil, b, []byte(" "))
		if len(s1) != 3 {
			panic("Split failed")
		}
		if string(s1[0]) != "go" || string(s1[1]) != "is" || string(s1[2]) != "fun" {
			panic("Split failed")
		}
		mem.FreeSlice(nil, s1)
		s2 := bytes.SplitN(nil, b, []byte(" "), 2)
		if len(s2) != 2 {
			panic("SplitN failed")
		}
		if string(s2[0]) != "go" || string(s2[1]) != "is fun" {
			panic("SplitN failed")
		}
		mem.FreeSlice(nil, s2)
	}
	{
		// Trim, TrimLeft, TrimRight.
		b := []byte("  hello  ")
		if string(bytes.Trim(b, " ")) != "hello" {
			panic("Trim failed")
		}
		if string(bytes.TrimLeft(b, " ")) != "hello  " {
			panic("TrimLeft failed")
		}
		if string(bytes.TrimRight(b, " ")) != "  hello" {
			panic("TrimRight failed")
		}
	}
	{
		// TrimPrefix and TrimSuffix.
		b := []byte("hello")
		if string(bytes.TrimPrefix(b, []byte("he"))) != "llo" {
			panic("TrimPrefix failed")
		}
		if string(bytes.TrimSuffix(b, []byte("lo"))) != "hel" {
			panic("TrimSuffix failed")
		}
	}
	{
		// ToLower and ToUpper.
		b := []byte("Hello")
		lowered := bytes.ToLower(nil, b)
		if string(lowered) != "hello" {
			panic("ToLower failed")
		}
		mem.FreeSlice(nil, lowered)
		uppered := bytes.ToUpper(nil, b)
		if string(uppered) != "HELLO" {
			panic("ToUpper failed")
		}
		mem.FreeSlice(nil, uppered)
	}
	{
		// Buffer (stack-allocated).
		buf := bytes.NewBuffer(nil, []byte("hello world"))
		if buf.String() != "hello world" {
			panic("Buffer Write failed")
		}
		rdbuf := make([]byte, 5)
		n, err := buf.Read(rdbuf)
		if n != 5 || string(rdbuf) != "hello" || err != nil {
			panic("Buffer Read failed")
		}
		if buf.String() != " world" {
			panic("Buffer Read did not advance the buffer")
		}
	}
	{
		// Buffer (heap-allocated).
		buf := bytes.NewBuffer(nil, nil)
		buf.WriteString("hello")
		buf.WriteString(" world")
		if buf.String() != "hello world" {
			panic("Buffer Write failed")
		}
		buf.Grow(64)
		if buf.Cap() < 64 {
			panic("Buffer Grow failed")
		}
		rdbuf := make([]byte, 5)
		n, err := buf.Read(rdbuf)
		if n != 5 || string(rdbuf) != "hello" || err != nil {
			panic("Buffer Read failed")
		}
		if buf.String() != " world" {
			panic("Buffer Read did not advance the buffer")
		}
		buf.Free()
	}
	{
		// Reader.
		s := "hello world"
		r := bytes.NewReader([]byte(s))
		if r.Len() != len(s) {
			panic("Reader Len failed")
		}
		b, err := io.ReadAll(nil, &r)
		if err != nil {
			panic(err)
		}
		if string(b) != s {
			panic("Reader Read failed")
		}
		if r.Len() != 0 {
			panic("Reader Len failed")
		}
		mem.FreeSlice(nil, b)
	}
}
