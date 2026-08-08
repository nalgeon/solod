package json_bench

import (
	"solod.dev/so/encoding/json"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

//so:volatile
var sink int64

//so:volatile
var sinkF float64

// benchDoc is a small document covering every token kind: ints, a float, a
// bool, null, an escaped string, an array of strings, and a nested object with
// an array of ints. The same literal drives the Go decode benchmark, so both
// sides scan identical bytes.
const benchDoc = `{"id":12345,"name":"widget","active":true,"ratio":0.75,` +
	`"note":"a\"b\tc","tags":["alpha","beta","gamma"],` +
	`"nested":{"x":1,"y":2,"z":[10,20,30]},"missing":null,"count":42}`

// uniDoc is a single string holding one astral rune spelled as a UTF-16
// surrogate pair, from Go's BenchmarkUnicodeDecoder. The escaped form exercises
// the surrogate-decoding path, not the raw-UTF-8 one.
const uniDoc = `"\uD83D\uDE01"`

func BenchmarkDecode_So(b *testing.B) {
	doc := []byte(benchDoc)
	b.SetBytes(int64(len(doc)))
	alloc := b.Allocator()
	for b.Loop() {
		d := json.NewDecoder(alloc, doc)
		for d.Next() {
			switch d.Kind() {
			case json.KindString:
				sink += int64(len(d.Str()))
			case json.KindNumber:
				sinkF += d.Float()
			case json.KindBool:
				if d.Bool() {
					sink++
				}
			}
		}
		d.Free()
	}
}

func BenchmarkDecodeUnicode_So(b *testing.B) {
	doc := []byte(uniDoc)
	b.SetBytes(int64(len(doc)))
	alloc := b.Allocator()
	for b.Loop() {
		d := json.NewDecoder(alloc, doc)
		d.Next()
		sink += int64(len(d.Str()))
		d.Free()
	}
}

func BenchmarkEncode_So(b *testing.B) {
	alloc := b.Allocator()
	out := mem.AllocSlice[byte](alloc, 512, 512)
	sb := strings.FixedBuilder(out)
	emitDoc(&sb) // one warm-up to size SetBytes off the real output
	b.SetBytes(int64(sb.Len()))
	for b.Loop() {
		sb.Reset()
		emitDoc(&sb)
		sink += int64(sb.Len())
	}
	mem.FreeSlice(alloc, out)
}

// emitDoc writes the document the encode benchmark measures, the same shape as
// benchDoc: every token kind, exercising the whole encoder.
func emitDoc(w *strings.Builder) {
	enc := json.NewEncoder(w)
	enc.BeginObject()
	enc.Str("id")
	enc.Int(12345)
	enc.Str("name")
	enc.Str("widget")
	enc.Str("active")
	enc.Bool(true)
	enc.Str("ratio")
	enc.Float(0.75)
	enc.Str("note")
	enc.Str("a\"b\tc")
	enc.Str("tags")
	enc.BeginArray()
	enc.Str("alpha")
	enc.Str("beta")
	enc.Str("gamma")
	enc.EndArray()
	enc.Str("nested")
	enc.BeginObject()
	enc.Str("x")
	enc.Int(1)
	enc.Str("y")
	enc.Int(2)
	enc.EndObject()
	enc.Str("missing")
	enc.Null()
	enc.Str("count")
	enc.Int(42)
	enc.EndObject()
	enc.Flush()
}
