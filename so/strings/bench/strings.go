package strings_bench

import (
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

//so:volatile
var sink string

//so:volatile
var sinkInt int

func BenchmarkClone_So(b *testing.B) {
	a := b.Allocator()
	str := strings.Repeat(nil, "a", 1024)
	defer mem.FreeString(nil, str)
	for b.Loop() {
		s := strings.Clone(a, str)
		mem.FreeString(a, s)
	}
}

func BenchmarkCompare_So(b *testing.B) {
	str1 := strings.Repeat(nil, "01234567890αβγδεζ", 64)
	str2 := strings.Repeat(nil, "01234567890αβγδεζ", 64)
	defer mem.FreeString(nil, str1)
	defer mem.FreeString(nil, str2)
	for b.Loop() {
		sinkInt = strings.Compare(str1, str2)
	}
}

func BenchmarkFields_So(b *testing.B) {
	a := b.Allocator()
	str := strings.Repeat(nil, "01234567890αβ γδεζ", 16)
	defer mem.FreeString(nil, str)
	for b.Loop() {
		fields := strings.Fields(a, str)
		mem.FreeSlice(a, fields)
	}
}

func BenchmarkIndex_So(b *testing.B) {
	sb := strings.NewBuilder(nil)
	for range 64 {
		sb.WriteString("01234567890αβγδεζ")
	}
	sb.WriteRune('ω')
	str := sb.String() // 1025 chars, search for last
	defer mem.FreeString(nil, str)
	for b.Loop() {
		sinkInt = strings.Index(str, "ω")
	}
}

func BenchmarkIndexByte_So(b *testing.B) {
	sb := strings.NewBuilder(nil)
	for range 64 {
		sb.WriteString("01234567890αβγδεζ")
	}
	sb.WriteByte('X')
	str := sb.String() // 1025 chars, search for last
	defer mem.FreeString(nil, str)
	for b.Loop() {
		sinkInt = strings.Index(str, "X")
	}
}

func BenchmarkRepeat_So(b *testing.B) {
	a := b.Allocator()
	str := "0123456789abcdef"
	for b.Loop() {
		s := strings.Repeat(a, str, 64)
		mem.FreeString(a, s)
	}
}

func BenchmarkReplaceAll_So(b *testing.B) {
	a := b.Allocator()
	str := strings.Repeat(nil, "0123456789abcdef", 16)
	defer mem.FreeString(nil, str)
	for b.Loop() {
		s := strings.ReplaceAll(a, str, "a", "A")
		mem.FreeString(a, s)
	}
}

func BenchmarkSplit_So(b *testing.B) {
	a := b.Allocator()
	str := strings.Repeat(nil, "01234567890αβγδεζ", 16)
	defer mem.FreeString(nil, str)
	for b.Loop() {
		fields := strings.Split(a, str, "γ")
		mem.FreeSlice(a, fields)
	}
}

func BenchmarkTrim_So(b *testing.B) {
	sb := strings.NewBuilder(nil)
	sb.WriteString("ηθικλμνξοπρστυφχψω")
	for range 64 {
		sb.WriteString("01234567890αβγδεζ")
	}
	sb.WriteString("ηθικλμνξοπρστυφχψω")
	str := sb.String()
	defer mem.FreeString(nil, str)
	for b.Loop() {
		sink = strings.Trim(str, "ωψχφυτσρποξνμλκιθη")
	}
}

func BenchmarkTrimSuffix_So(b *testing.B) {
	str := strings.Repeat(nil, "01234567890αβγδεζ", 16)
	suffix := "01234567890αβγδεζ"
	defer mem.FreeString(nil, str)
	for b.Loop() {
		sink = strings.TrimSuffix(str, suffix)
	}
}

func BenchmarkToUpper_So(b *testing.B) {
	a := b.Allocator()
	str := strings.Repeat(nil, "01234567890αβγδεζ", 16)
	defer mem.FreeString(nil, str)
	for b.Loop() {
		s := strings.ToUpper(a, str)
		mem.FreeString(a, s)
	}
}
