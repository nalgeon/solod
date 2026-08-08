package json_test

import (
	"solod.dev/so/encoding/json"
	"solod.dev/so/io"
	"solod.dev/so/math"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestEncodeObject(t *testing.T) {
	out := make([]byte, 256)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	enc.BeginObject()
	enc.Str("name")
	enc.Str("Bob")
	enc.Str("age")
	enc.Int(42)
	enc.Str("tall")
	enc.Bool(true)
	enc.EndObject()
	enc.Flush()

	if sb.String() != `{"name":"Bob","age":42,"tall":true}` {
		t.Error("object encode mismatch")
	}
	if enc.Err() != nil {
		t.Error("unexpected encode error")
	}
}

func TestEncodeNested(t *testing.T) {
	out := make([]byte, 256)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	enc.BeginObject()
	enc.Str("nums")
	enc.BeginArray()
	enc.Int(1)
	enc.Int(2)
	enc.Int(3)
	enc.EndArray()
	enc.EndObject()
	enc.Flush()

	if sb.String() != `{"nums":[1,2,3]}` {
		t.Error("nested encode mismatch")
	}
}

func TestEncodeEscapes(t *testing.T) {
	out := make([]byte, 256)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	enc.Str("a\"b\n\tc")
	enc.Flush()

	if sb.String() != `"a\"b\n\tc"` {
		t.Error("escape encode mismatch")
	}
}

func TestEncodeInvalidUTF8(t *testing.T) {
	out := make([]byte, 64)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	// Raw bytes that do not represent a character are not a valid token.
	bad := []byte{'a', 0xff, 'b', 0xc3, 'c'} // lone 0xff, then a truncated 2-byte seq
	enc.Str(string(bad))
	if enc.Err() != json.ErrValue {
		t.Error("invalid UTF-8 should yield ErrValue")
	}
	enc.Flush()
	if sb.String() != "" {
		t.Error("a rejected string should write nothing, not even the quote")
	}

	// Valid multi-byte runes must survive untouched.
	out2 := make([]byte, 64)
	sb2 := strings.FixedBuilder(out2)
	enc2 := json.NewEncoder(&sb2)
	enc2.Str("héllo 😀")
	enc2.Flush()
	if sb2.String() != `"héllo 😀"` {
		t.Error("valid UTF-8 should pass through")
	}
	if enc2.Err() != nil {
		t.Error("unexpected encode error")
	}
}

func TestEncodeNonFinite(t *testing.T) {
	// JSON cannot spell NaN or an infinity, so neither is written.
	out := make([]byte, 64)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)
	enc.Float(math.NaN())
	enc.Flush()
	if enc.Err() != json.ErrNonFinite {
		t.Error("NaN should yield ErrNonFinite")
	}
	if sb.String() != "" {
		t.Error("NaN should write nothing")
	}

	out2 := make([]byte, 64)
	sb2 := strings.FixedBuilder(out2)
	enc2 := json.NewEncoder(&sb2)
	enc2.BeginArray()
	enc2.Float(math.Inf(1))
	if enc2.Err() != json.ErrNonFinite {
		t.Error("Inf should yield ErrNonFinite")
	}
}

func TestEncodeSingleRoot(t *testing.T) {
	// A JSON document holds exactly one root value, so a second one is rejected
	// rather than concatenated (which would have written "12").
	out := make([]byte, 64)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)
	enc.Int(1)
	enc.Int(2)
	enc.Flush()
	if enc.Err() != json.ErrSyntax {
		t.Error("a second root value should be rejected")
	}
	if sb.String() != "1" {
		t.Error("the rejected root value should write nothing")
	}

	// Anything after a completed root container is a second root too.
	out2 := make([]byte, 64)
	sb2 := strings.FixedBuilder(out2)
	enc2 := json.NewEncoder(&sb2)
	enc2.BeginArray()
	enc2.Int(1)
	enc2.EndArray()
	enc2.Str("extra")
	enc2.Flush()
	if enc2.Err() != json.ErrSyntax {
		t.Error("a value after the root container should be rejected")
	}
	if sb2.String() != "[1]" {
		t.Error("the rejected value should write nothing")
	}
}

func TestEncodeInvalid(t *testing.T) {
	// A key with no value.
	out := make([]byte, 64)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)
	enc.BeginObject()
	enc.Str("key")
	enc.EndObject()
	if enc.Err() != json.ErrSyntax {
		t.Error("key without a value should be rejected")
	}

	// A closer that does not match the open container.
	out2 := make([]byte, 64)
	sb2 := strings.FixedBuilder(out2)
	enc2 := json.NewEncoder(&sb2)
	enc2.BeginObject()
	enc2.EndArray()
	if enc2.Err() != json.ErrSyntax {
		t.Error("mismatched closer should be rejected")
	}

	// A closer with nothing open.
	out3 := make([]byte, 64)
	sb3 := strings.FixedBuilder(out3)
	enc3 := json.NewEncoder(&sb3)
	enc3.EndObject()
	if enc3.Err() != json.ErrSyntax {
		t.Error("stray closer should be rejected")
	}

	// An object key that is not a string.
	out4 := make([]byte, 64)
	sb4 := strings.FixedBuilder(out4)
	enc4 := json.NewEncoder(&sb4)
	enc4.BeginObject()
	enc4.Int(1)
	enc4.Flush()
	if enc4.Err() != json.ErrSyntax {
		t.Error("non-string object key should be rejected")
	}
	if sb4.String() != "{" {
		t.Error("rejected key should write nothing")
	}
}

func TestEncodeDepth(t *testing.T) {
	// Nesting past MaxDepth fails on the token that overflows, and fails before
	// anything is written: neither the bracket nor the separator in front of it
	// reaches the output, so the drained bytes hold everything the caller wrote
	// successfully and nothing it did not.
	out := make([]byte, json.MaxDepth+1)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	for range json.MaxDepth {
		enc.BeginArray()
	}
	enc.Int(1)
	if enc.Err() != nil {
		t.Error("MaxDepth levels should be allowed")
		return
	}

	enc.BeginArray() // one level too many
	if enc.Err() != json.ErrDepth {
		t.Error("expected ErrDepth")
	}
	enc.Flush()
	if len(sb.String()) != json.MaxDepth+1 { // MaxDepth brackets and the 1
		t.Error("the overflowing bracket should write nothing")
	}
}

func TestEncodeIncomplete(t *testing.T) {
	// A closing token the caller never wrote is the one mistake with no point at
	// which to report it, so Flush is where an incomplete document is caught.
	out := make([]byte, 64)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)
	enc.BeginObject()
	enc.Str("a")
	enc.Int(1)
	enc.Flush() // the caller forgot EndObject
	if enc.Err() != json.ErrSyntax {
		t.Error("an unclosed container should be rejected")
	}
	// The bytes written so far are still handed over: withholding them would
	// only make the mistake harder to see.
	if sb.String() != `{"a":1` {
		t.Error("Flush should drain what was written")
	}

	// An empty document is not a document either.
	out2 := make([]byte, 64)
	sb2 := strings.FixedBuilder(out2)
	enc2 := json.NewEncoder(&sb2)
	enc2.Flush()
	if enc2.Err() != json.ErrSyntax {
		t.Error("an empty document should be rejected")
	}

	// A complete document is not flagged.
	out3 := make([]byte, 64)
	sb3 := strings.FixedBuilder(out3)
	enc3 := json.NewEncoder(&sb3)
	enc3.BeginArray()
	enc3.Int(1)
	enc3.EndArray()
	enc3.Flush()
	if enc3.Err() != nil {
		t.Error("a complete document should flush cleanly")
	}
}

func TestEncodeBuffered(t *testing.T) {
	// A document larger than the output buffer must drain as it fills, so the
	// result is the same as one that fits. Encode, then read it back.
	out := make([]byte, 4096)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	enc.BeginArray()
	for i := range 500 {
		enc.Int(int64(i))
	}
	enc.EndArray()
	enc.Flush()
	if enc.Err() != nil {
		t.Error("unexpected encode error")
	}

	doc := []byte(sb.String())
	if len(doc) <= 512 {
		t.Error("test should produce more than one buffer's worth")
		return
	}

	dec := json.NewDecoder(mem.System, doc)
	defer dec.Free()
	var sum, count int64
	for dec.Next() {
		if dec.Kind() == json.KindNumber {
			sum += dec.Int()
			count++
		}
	}
	if dec.Err() != nil {
		t.Error("the buffered document should decode cleanly")
	}
	if count != 500 {
		t.Error("element count mismatch")
	}
	if sum != 124750 { // 0+1+...+499
		t.Error("element sum mismatch")
	}
}

// shortWriter takes the first limit bytes it is ever handed, drops the rest,
// and reports success either way. It stands in for a writer that has filled up
// (a full disk, a fixed buffer) without saying so.
type shortWriter struct {
	buf   []byte
	limit int
	n     int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	room := min(w.limit-w.n, len(p))
	copy(w.buf[w.n:], p[:room])
	w.n += room
	return room, nil // fewer bytes than asked for, no error
}

func TestEncodeShortWrite(t *testing.T) {
	out := make([]byte, 4096)
	w := shortWriter{buf: out, limit: 8}
	enc := json.NewEncoder(&w)

	enc.BeginArray()
	for i := range 500 { // more than one bufferful, so the encoder has to drain
		enc.Int(int64(i))
	}
	enc.EndArray()
	enc.Flush()

	if enc.Err() != io.ErrShortWrite {
		t.Error("a short write should yield ErrShortWrite")
	}
}

func TestEncodeLongToken(t *testing.T) {
	// A single token longer than the output buffer is written across drains.
	out := make([]byte, 4096)
	sb := strings.FixedBuilder(out)
	enc := json.NewEncoder(&sb)

	big := make([]byte, 2000)
	for i := range big {
		big[i] = 'x'
	}
	enc.Str(string(big))
	enc.Flush()

	if enc.Err() != nil {
		t.Error("unexpected encode error")
	}
	if len(sb.String()) != 2002 { // 2000 plus the two quotes
		t.Error("long token was truncated")
	}
}
