package json_bench

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

// The Go benchmarks mirror the Solod ones over the same input (benchDoc, uniDoc,
// defined in json.go).
//
// Decode is a fair, token-level comparison: encoding/json's Token stream is
// the direct analogue of Solod's Next/Kind/getters.
//
// Encode is not: Go's Encoder streams to an io.Writer like Solod's, but it has
// no token-level API, so it marshals the whole value by reflection (the same
// work as Marshal for a single value). [BenchmarkEncode_Go] therefore measures
// a different kind of work and the numbers are indicative at best.

func BenchmarkDecode_Go(b *testing.B) {
	doc := []byte(benchDoc)
	b.SetBytes(int64(len(doc)))
	for b.Loop() {
		dec := json.NewDecoder(bytes.NewReader(doc))
		for {
			t, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			switch v := t.(type) {
			case string:
				sink += int64(len(v))
			case float64:
				sinkF += v
			case bool:
				if v {
					sink++
				}
			}
		}
	}
}

func BenchmarkDecodeUnicode_Go(b *testing.B) {
	doc := []byte(uniDoc)
	b.SetBytes(int64(len(doc)))
	for b.Loop() {
		dec := json.NewDecoder(bytes.NewReader(doc))
		t, err := dec.Token()
		if err != nil {
			b.Fatal(err)
		}
		s, _ := t.(string)
		sink += int64(len(s))
	}
}

// goRec is Marshaled to the same shape emitDoc writes on the Solod side.
type goRec struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Ratio  float64  `json:"ratio"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
	Nested struct {
		X int64 `json:"x"`
		Y int64 `json:"y"`
	} `json:"nested"`
	Missing *int  `json:"missing"`
	Count   int64 `json:"count"`
}

func BenchmarkEncode_Go(b *testing.B) {
	v := goRec{
		ID:     12345,
		Name:   "widget",
		Active: true,
		Ratio:  0.75,
		Note:   "a\"b\tc",
		Tags:   []string{"alpha", "beta", "gamma"},
		Count:  42,
	}
	v.Nested.X = 1
	v.Nested.Y = 2

	// Size SetBytes off the actual output, including the newline Encode adds.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(buf.Len()))

	// Reuse one Encoder writing to Discard, as Go's own BenchmarkCodeEncoder
	// does. The Solod side builds a fresh Encoder per document (a value type, so no
	// allocation), which its single-root design requires.
	enc := json.NewEncoder(io.Discard)
	for b.Loop() {
		if err := enc.Encode(v); err != nil {
			b.Fatal(err)
		}
		sink++
	}
}
