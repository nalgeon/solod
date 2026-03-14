package main

import (
	"github.com/nalgeon/solod/so/bytes"
	"github.com/nalgeon/solod/so/mem"
)

func isLatinLower(r rune) bool {
	return 'a' <= uint32(r) && uint32(r) <= 'z'
}

func isPunct(r rune) bool {
	return uint32(r) == ',' || uint32(r) == ';'
}

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
		// ContainsAny.
		b := []byte("I like seafood.")
		if !bytes.ContainsAny(b, "aei") {
			panic("ContainsAny failed")
		}
		if bytes.ContainsAny(b, "xyz") {
			panic("ContainsAny failed")
		}
	}
	{
		// ContainsRune.
		b := []byte("I like seafood.")
		if !bytes.ContainsRune(b, 'f') {
			panic("ContainsRune failed")
		}
		if bytes.ContainsRune(b, 'x') {
			panic("ContainsRune failed")
		}
	}
	{
		// ContainsFunc.
		if bytes.ContainsFunc([]byte("HELLO"), isLatinLower) {
			panic("ContainsFunc failed")
		}
		if !bytes.ContainsFunc([]byte("World"), isLatinLower) {
			panic("ContainsFunc failed")
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
		// CutPrefix.
		b := []byte("hello")
		after, found := bytes.CutPrefix(b, []byte("hel"))
		if string(after) != "lo" || !found {
			panic("CutPrefix failed")
		}
	}
	{
		// CutSuffix.
		b := []byte("hello")
		before, found := bytes.CutSuffix(b, []byte("lo"))
		if string(before) != "hel" || !found {
			panic("CutSuffix failed")
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
		// Fields.
		b := []byte("go is fun")
		fields := bytes.Fields(nil, b)
		if len(fields) != 3 {
			panic("Fields failed")
		}
		if string(fields[0]) != "go" || string(fields[1]) != "is" || string(fields[2]) != "fun" {
			panic("Fields failed")
		}
		mem.FreeSlice(nil, fields)
	}
	{
		// FieldsFunc.
		b := []byte("go,is;fun")
		fields := bytes.FieldsFunc(nil, b, isPunct)
		if len(fields) != 3 {
			panic("FieldsFunc failed")
		}
		if string(fields[0]) != "go" || string(fields[1]) != "is" || string(fields[2]) != "fun" {
			panic("FieldsFunc failed")
		}
		mem.FreeSlice(nil, fields)
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
		// Index, IndexByte, IndexAny, IndexRune.
		b := []byte("hello")
		if bytes.Index(b, []byte("l")) != 2 {
			panic("Index failed")
		}
		if bytes.IndexByte(b, 'e') != 1 {
			panic("Index failed")
		}
		if bytes.IndexAny(b, "aeiou") != 1 {
			panic("IndexAny failed")
		}
		if bytes.IndexRune(b, 'o') != 4 {
			panic("IndexRune failed")
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
		// LastIndex, LastIndexByte, LastIndexAny.
		b := []byte("hello")
		if bytes.LastIndex(b, []byte("l")) != 3 {
			panic("LastIndex failed")
		}
		if bytes.LastIndexByte(b, 'l') != 3 {
			panic("LastIndexByte failed")
		}
		if bytes.LastIndexAny(b, "al") != 3 {
			panic("LastIndexAny failed")
		}
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
		b := []byte("go")
		repeated := bytes.Repeat(nil, b, 3)
		if string(repeated) != "gogogo" {
			panic("Repeat failed")
		}
		mem.FreeSlice(nil, repeated)
	}
	{
		// Replace and ReplaceAll.
		b := []byte("hello")
		r1 := bytes.Replace(nil, b, []byte("l"), []byte("x"), 1)
		if string(r1) != "hexlo" {
			panic("Replace failed")
		}
		mem.FreeSlice(nil, r1)
		r2 := bytes.ReplaceAll(nil, b, []byte("l"), []byte("x"))
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
		// ToTitle.
		b := []byte("hello")
		titled := bytes.ToTitle(nil, b)
		if string(titled) != "HELLO" {
			panic("ToTitle failed")
		}
		mem.FreeSlice(nil, titled)
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
}
