package maps_bench

import (
	"fmt"
	"testing"
)

func init() {
	for i := range nKeys {
		strKeys = append(strKeys, fmt.Sprintf("key-%d", i))
	}
}

func BenchmarkIntSet_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[int]int)
		for i := range nKeys {
			m[i] = i
		}
	}
}

func BenchmarkIntPre_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[int]int, nKeys)
		for i := range nKeys {
			m[i] = i
		}
	}
}

func BenchmarkIntGet_Go(b *testing.B) {
	b.ReportAllocs()
	m := make(map[int]int, nKeys)
	for i := range nKeys {
		m[i] = i
	}
	for b.Loop() {
		for i := range nKeys {
			sinkInt = m[i]
		}
	}
}

func BenchmarkIntDel_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[int]int, nKeys)
		for i := range nKeys {
			m[i] = i
		}
		for i := range nKeys {
			delete(m, i)
		}
	}
}

func BenchmarkStrSet_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[string]int)
		for i := range nKeys {
			m[strKeys[i]] = i
		}
	}
}

func BenchmarkStrPre_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[string]int, nKeys)
		for i := range nKeys {
			m[strKeys[i]] = i
		}
	}
}

func BenchmarkStrGet_Go(b *testing.B) {
	b.ReportAllocs()
	m := make(map[string]int, nKeys)
	for i := range nKeys {
		m[strKeys[i]] = i
	}
	for b.Loop() {
		for i := range nKeys {
			sinkInt = m[strKeys[i]]
		}
	}
}

func BenchmarkStrDel_Go(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := make(map[string]int, nKeys)
		for i := range nKeys {
			m[strKeys[i]] = i
		}
		for i := range nKeys {
			delete(m, strKeys[i])
		}
	}
}
