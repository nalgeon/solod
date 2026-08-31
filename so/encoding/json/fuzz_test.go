package json

import (
	stdjson "encoding/json"
	"testing"
)

// FuzzDecode checks the Decoder against Go's encoding/json, and the streaming
// Decoder against the in-memory one.
//
// The two decoders are not held to the same standard. Go's is the authority on
// what valid JSON is, so anything this package accepts it must accept too. The
// reverse does not hold: this package is deliberately stricter about values
// (it rejects non-UTF-8 and unpaired surrogates where Go substitutes U+FFFD)
// and it caps nesting at MaxDepth. So a Go-valid document may still be turned
// away, but only with ErrValue or ErrDepth.
//
// The two Solod decoders, on the other hand, must agree exactly. They share every
// scanner and differ only in where the bytes come from, so any disagreement is
// a bug in the buffer's refill, compaction, or growth - the code a fixed
// document never reaches. Driving it across buffer and chunk sizes is what
// puts a token boundary at every offset in the buffer.
func FuzzDecode(f *testing.F) {
	seeds := []string{
		`{"a":1}`, `[1,2,3]`, `"x"`, `null`, `true`, `-0.5e+3`, `{}`, `[]`,
		`{"a":{"b":[1,{"c":null}]}}`, `"é😀"`, `"\\\"\/\b\f\n\r\t"`,
		` { "k" : [ ] } `, `01`, `[1,]`, `{"a":}`, `"\ud800"`, `"\udc00x"`,
		`"😀"`, `1e309`, `-`, `{"a" 1}`, `[[[[[[]]]]]]`, "\"\x00\"",
		"\"\xff\"", `18446744073709551615`, "{}\n{}",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, doc []byte) {
		if len(doc) > 4096 {
			return
		}
		fixed, err := tokens(doc, false, 0, 0)

		if err == nil && !stdjson.Valid(doc) {
			t.Fatalf("accepted invalid JSON %q: %v", doc, fixed)
		}
		if err != nil && stdjson.Valid(doc) && err != ErrValue && err != ErrDepth {
			t.Fatalf("rejected valid JSON %q: %v", doc, err)
		}

		for _, bufSize := range []int{0, 1, 16, 17, 64} {
			for _, chunk := range []int{1, 2, 3, 7, 64, 4096} {
				got, gotErr := tokens(doc, true, bufSize, chunk)
				if gotErr != err {
					t.Fatalf("stream/fixed err mismatch on %q (buf=%d chunk=%d):\nstream %v\n fixed %v",
						doc, bufSize, chunk, gotErr, err)
				}
				if err == nil && !eq(got, fixed) {
					t.Fatalf("stream/fixed token mismatch on %q (buf=%d chunk=%d):\nstream %v\n fixed %v",
						doc, bufSize, chunk, got, fixed)
				}
			}
		}
	})
}

// FuzzEncode drives the Encoder with a sequence of calls taken from the fuzz
// input. Most sequences are nonsense the Encoder rejects, which is the point:
// whatever it does accept must be valid JSON, and must decode back to the very
// tokens it was handed.
func FuzzEncode(f *testing.F) {
	f.Add([]byte{0, 4, 5, 1})       // {"k":-42}
	f.Add([]byte{2, 0, 1, 7, 6, 3}) // [{},null,true]
	f.Add([]byte{1})                // a stray '}'
	f.Add([]byte{4, 4})             // a second root value
	f.Add([]byte{0, 5})             // a number where a key belongs

	// The string carries every escape class at once: a quote, a control byte,
	// a byte the encoder spells as \u00XX, and a multi-byte rune.
	const str = "k\"\n\x01é"

	f.Fuzz(func(t *testing.T, ops []byte) {
		if len(ops) > 256 {
			return
		}
		w := &writer{}
		e := NewEncoder(w)
		var want []string
		for _, op := range ops {
			switch op % 8 {
			case 0:
				e.BeginObject()
				want = append(want, "{")
			case 1:
				e.EndObject()
				want = append(want, "}")
			case 2:
				e.BeginArray()
				want = append(want, "[")
			case 3:
				e.EndArray()
				want = append(want, "]")
			case 4:
				e.Str(str)
				want = append(want, "s:"+str)
			case 5:
				e.Int(-42)
				want = append(want, "n:-42")
			case 6:
				e.Bool(true)
				want = append(want, "b:true")
			case 7:
				e.Null()
				want = append(want, "null")
			}
		}
		e.Flush()
		if e.Err() != nil {
			return // the call sequence was not a document; nothing to check
		}
		if !stdjson.Valid(w.b) {
			t.Fatalf("encoded invalid JSON from %v: %q", ops, w.b)
		}
		got, err := tokens(w.b, false, 0, 0)
		if err != nil {
			t.Fatalf("cannot decode own output %q: %v", w.b, err)
		}
		if !eq(got, want) {
			t.Fatalf("roundtrip mismatch for %q:\ngot  %v\nwant %v", w.b, got, want)
		}
	})
}
