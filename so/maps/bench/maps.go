package maps_bench

import (
	"solod.dev/so/fmt"
	"solod.dev/so/maps"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

//so:volatile
var sinkInt int

//so:volatile
var sinkBool bool

// nKeys is the number of map keys to use in benchmarks.
const nKeys = 1024

// strKeys holds pre-generated string keys for string benchmarks.
var strKeys []string

// ensureStrKeys lazily populates strKeys on first use. so bench provides no
// per-run setup hook, so the string benchmarks initialize it themselves.
func ensureStrKeys() {
	if strKeys != nil {
		return
	}
	strKeys = mem.AllocSlice[string](nil, nKeys, nKeys)
	buf := make([]byte, 32)
	for i := range nKeys {
		strKeys[i] = strings.Clone(nil, fmt.Sprintf(buf, "key-%d", i))
	}
}

func BenchmarkIntSet_So(b *testing.B) {
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[int, int](a, 0)
		for i := range nKeys {
			m.Set(i, i)
		}
		m.Free()
	}
}

func BenchmarkIntPre_So(b *testing.B) {
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[int, int](a, nKeys)
		for i := range nKeys {
			m.Set(i, i)
		}
		m.Free()
	}
}

func BenchmarkIntGet_So(b *testing.B) {
	m := maps.New[int, int](nil, nKeys)
	for i := range nKeys {
		m.Set(i, i)
	}
	defer m.Free()
	for b.Loop() {
		for i := range nKeys {
			sinkInt = m.Get(i)
		}
	}
}

func BenchmarkIntDel_So(b *testing.B) {
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[int, int](a, nKeys)
		for i := range nKeys {
			m.Set(i, i)
		}
		for i := range nKeys {
			m.Delete(i)
		}
		m.Free()
	}
}

func BenchmarkStrSet_So(b *testing.B) {
	ensureStrKeys()
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[string, int](a, 0)
		for i := range nKeys {
			m.Set(strKeys[i], i)
		}
		m.Free()
	}
}

func BenchmarkStrPre_So(b *testing.B) {
	ensureStrKeys()
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[string, int](a, nKeys)
		for i := range nKeys {
			m.Set(strKeys[i], i)
		}
		m.Free()
	}
}

func BenchmarkStrGet_So(b *testing.B) {
	ensureStrKeys()
	m := maps.New[string, int](nil, nKeys)
	for i := range nKeys {
		m.Set(strKeys[i], i)
	}
	defer m.Free()
	for b.Loop() {
		for i := range nKeys {
			sinkInt = m.Get(strKeys[i])
		}
	}
}

func BenchmarkStrDel_So(b *testing.B) {
	ensureStrKeys()
	a := b.Allocator()
	for b.Loop() {
		m := maps.New[string, int](a, nKeys)
		for i := range nKeys {
			m.Set(strKeys[i], i)
		}
		for i := range nKeys {
			m.Delete(strKeys[i])
		}
		m.Free()
	}
}

// The Builtin* benchmarks measure the language builtin map (make(map)) under
// So, as a counterpart to the so/maps package benchmarks above. The builtin map
// allocates on the stack (alloca), so it is only freed when the enclosing
// function returns; the Set benchmarks delegate to a helper called each
// iteration so the map does not accumulate across the loop.

func BenchmarkBuiltinIntSet_So(b *testing.B) {
	for b.Loop() {
		builtinIntSet(nKeys)
	}
}

func builtinIntSet(nKeys int) {
	m := make(map[int]int, nKeys)
	for i := range nKeys {
		m[i] = i
	}
	sinkInt = m[0]
}

func BenchmarkBuiltinIntGet_So(b *testing.B) {
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

func BenchmarkBuiltinStrSet_So(b *testing.B) {
	ensureStrKeys()
	for b.Loop() {
		builtinStrSet(nKeys)
	}
}

func builtinStrSet(nKeys int) {
	m := make(map[string]int, nKeys)
	for i := range nKeys {
		m[strKeys[i]] = i
	}
	sinkInt = m[strKeys[0]]
}

func BenchmarkBuiltinStrGet_So(b *testing.B) {
	ensureStrKeys()
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
