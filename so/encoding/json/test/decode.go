package json_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/encoding/json"
	"solod.dev/so/errors"
	"solod.dev/so/io"
	"solod.dev/so/math"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestDecodeScalars(t *testing.T) {
	src := `{"name":"Bob","age":42,"tall":true,"note":null}`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	if !dec.Next() || dec.Kind() != json.KindObjBeg {
		t.Fatal("expected object begin")
		return
	}

	var name string
	var age int64
	var tall bool
	sawNull := false
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "name":
			dec.Next()
			name = dec.Str()
		case "age":
			dec.Next()
			age = dec.Int()
		case "tall":
			dec.Next()
			tall = dec.Bool()
		case "note":
			dec.Next()
			sawNull = dec.Kind() == json.KindNull
		default:
			dec.Next()
			dec.Skip()
		}
	}

	if name != "Bob" {
		t.Error("name mismatch")
	}
	if age != 42 {
		t.Error("age mismatch")
	}
	if !tall {
		t.Error("tall mismatch")
	}
	if !sawNull {
		t.Error("note should be null")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeEscapes(t *testing.T) {
	src := `"a\"b\n\tcé"`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	if !dec.Next() || dec.Kind() != json.KindString {
		t.Fatal("expected string")
		return
	}
	if dec.Str() != "a\"b\n\tcé" {
		t.Error("escape decode mismatch")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeKeepsSource(t *testing.T) {
	// Unescaping goes to the decoder's scratch, never to the document, so the
	// document survives the decode. This is what lets the source be a string
	// literal, a read-only mapping, or a buffer the caller still needs: the
	// alternative, rewriting it in place, would corrupt all three (and fault on
	// the middle one) the moment a string carried an escape.
	src := `{"msg":"a\"b\n\tc"}`

	// The copy is what gives this test teeth. []byte(src) would be a view of
	// src's own bytes, so a decoder that rewrote it would rewrite both sides of
	// the comparison below and still look innocent. A separate buffer records
	// the write. (That the literal-backed decoders in the other tests do not
	// fault is the other half of the guarantee.)
	doc := make([]byte, len(src))
	copy(doc, src)

	dec := json.NewDecoder(mem.System, doc)
	defer dec.Free()
	for dec.Next() {
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
		return
	}
	if string(doc) != src {
		t.Error("the decoder rewrote the document")
	}

	// Decoding it a second time yields the same thing, rather than tripping
	// over the leftovers of the first pass.
	dec2 := json.NewDecoder(mem.System, doc)
	defer dec2.Free()
	dec2.Next() // {
	dec2.Next() // the key
	dec2.Next() // the value
	if dec2.Str() != "a\"b\n\tc" {
		t.Error("the document did not decode the same way twice")
	}
	if dec2.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeInvalidUTF8(t *testing.T) {
	// Raw bytes that do not represent a character are not a valid token.
	cases := []string{
		"\"\xff\"",         // a byte that starts no sequence
		"\"\xc3\"",         // a 2-byte sequence cut short
		"\"\xed\xa0\x80\"", // a surrogate half, spelled out in UTF-8
		"{\"\xff\":1}",     // an object key is a string too
		"[1,\"a\xffz\",2]", // ... and so is an array element
	}
	for _, src := range cases {
		dec := json.NewDecoder(mem.System, []byte(src))
		for dec.Next() {
		}
		if dec.Err() != json.ErrValue {
			t.Error("invalid UTF-8 should yield ErrValue")
		}
		dec.Free()
	}

	// Valid multi-byte runes decode untouched.
	src := `"héllo 😀"`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()
	dec.Next()
	if dec.Str() != "héllo 😀" {
		t.Error("valid UTF-8 should decode as it is")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeSurrogates(t *testing.T) {
	// A high surrogate and a low one together spell a single code point
	// outside the BMP: U+1F600.
	hi, lo := `\ud83d`, `\ude00`
	src := `"` + hi + lo + `"`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()
	dec.Next()
	if dec.Str() != "😀" {
		t.Error("a surrogate pair should decode to one rune")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}

	// Half a pair names no character at all, so it is rejected rather than
	// silently becoming U+FFFD, which would rewrite the caller's data.
	cases := []string{
		`"\ud83d"`,       // a high surrogate with nothing after it
		`"\ud83dx"`,      // ... or with a plain character after it
		`"\ud83d\ud83d"`, // ... or with a second high one instead of a low one
		`"\ude00"`,       // a low surrogate with no high one before it
	}
	for _, src := range cases {
		dec := json.NewDecoder(mem.System, []byte(src))
		for dec.Next() {
		}
		if dec.Err() != json.ErrValue {
			t.Error("an unpaired surrogate should yield ErrValue")
		}
		dec.Free()
	}
}

func TestDecodeSkip(t *testing.T) {
	src := `{"a":{"nested":[1,2,3]},"b":7}`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	dec.Next() // {
	var b int64
	for dec.Next() && dec.Kind() == json.KindString {
		isB := dec.Str() == "b"
		dec.Next()
		if isB {
			b = dec.Int()
		} else {
			dec.Skip()
		}
	}
	if b != 7 {
		t.Error("skip did not land on b")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeNestingDepth(t *testing.T) {
	// Depth counts the open containers. A container's opening bracket is reported
	// at its own depth; its closing bracket at the enclosing depth. So the same
	// pair of brackets is seen at different depths (e.g. the inner object below is
	// entered at 3 and left at 2).
	src := `{"a":[1,{"b":2}],"c":3}`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	kinds := []json.Kind{
		json.KindObjBeg, // {
		json.KindString, // "a"
		json.KindArrBeg, // [
		json.KindNumber, // 1
		json.KindObjBeg, // {
		json.KindString, // "b"
		json.KindNumber, // 2
		json.KindObjEnd, // }
		json.KindArrEnd, // ]
		json.KindString, // "c"
		json.KindNumber, // 3
		json.KindObjEnd, // }
	}
	depths := []int{1, 1, 2, 2, 3, 3, 3, 2, 1, 1, 1, 0}

	for i := range kinds {
		if !dec.Next() {
			t.Fatal("stream ended early")
			return
		}
		if dec.Kind() != kinds[i] {
			t.Error("kind mismatch")
		}
		if dec.Depth() != depths[i] {
			t.Error("depth mismatch")
		}
	}
	if dec.Next() {
		t.Error("expected end of stream")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeInvalid(t *testing.T) {
	cases := []string{
		`[1 2]`,    // missing comma
		`{"a" 1}`,  // missing colon
		`{"a":1,}`, // trailing comma in object
		`[1,]`,     // trailing comma in array
		`{1:2}`,    // non-string key
		`[1,2}`,    // mismatched brackets
		`{"a":1]`,  // mismatched brackets
		`01`,       // leading zero
		`1.`,       // no fraction digits
		`1e`,       // no exponent digits
		`-`,        // lone minus
		`{"a":}`,   // missing value
		`42 43`,    // trailing data
		`[1,2,3`,   // unterminated array
		`{"a":1`,   // unterminated object
		``,         // empty document
		"  \n\t",   // whitespace only
	}
	for _, src := range cases {
		dec := json.NewDecoder(mem.System, []byte(src))
		for dec.Next() {
		}
		if dec.Err() == nil {
			t.Error("expected error for invalid input")
		}
		dec.Free()
	}
}

// brackets builds "[[[...]]]", n levels deep. Free it with mem.FreeSlice.
func brackets(n int) []byte {
	doc := mem.AllocSlice[byte](mem.System, 2*n, 2*n)
	for i := range n {
		doc[i] = '['
		doc[2*n-1-i] = ']'
	}
	return doc
}

func TestDecodeDepth(t *testing.T) {
	// The nesting limit is exact: MaxDepth levels decode, one more does not.
	// Next fails on the token that overflows rather than accepting a token it
	// did not push.
	doc := brackets(json.MaxDepth)
	defer mem.FreeSlice(mem.System, doc)
	dec := json.NewDecoder(mem.System, doc)
	defer dec.Free()
	ntok := 0
	for dec.Next() {
		ntok++
	}
	if dec.Err() != nil {
		t.Error("MaxDepth levels should decode")
	}
	if ntok != 2*json.MaxDepth {
		t.Error("token count mismatch")
	}

	doc2 := brackets(json.MaxDepth + 1)
	defer mem.FreeSlice(mem.System, doc2)
	dec2 := json.NewDecoder(mem.System, doc2)
	defer dec2.Free()
	ntok2 := 0
	for dec2.Next() {
		ntok2++
	}
	if dec2.Err() != json.ErrDepth {
		t.Error("expected ErrDepth")
	}
	if ntok2 != json.MaxDepth {
		t.Error("Next accepted a token past MaxDepth")
	}
}

func TestDecodeContainerToken(t *testing.T) {
	// A container carries no value, so a getter must not pick up the token
	// before it.
	src := `[1,[2]]`
	dec := json.NewDecoder(mem.System, []byte(src))
	defer dec.Free()

	dec.Next() // [
	dec.Next() // 1
	if dec.Int() != 1 {
		t.Error("expected 1")
	}
	dec.Next() // [ - the nested array, not the 1 before it
	if dec.Kind() != json.KindArrBeg {
		t.Error("expected array begin")
		return
	}
	if dec.Int() != 0 {
		t.Error("container should carry no token")
	}
	if dec.Err() != json.ErrKind {
		t.Error("a getter on a container should yield ErrKind")
	}
}

func TestDecodeWrongKind(t *testing.T) {
	// A getter reads one kind of token and nothing else. Asking for the wrong
	// one is how a document that is valid JSON of an unexpected shape shows up,
	// so it is reported like any other decode failure rather than returning
	// something plausible.
	cases := []string{
		`"42"`, // a number field that arrived as a string
		`true`, // ... or as a bool: Int must not read the literal bytes
		`null`, //
		`[1]`,  // a container carries no value at all
	}
	for _, src := range cases {
		dec := json.NewDecoder(mem.System, []byte(src))
		dec.Next()
		if dec.Int() != 0 {
			t.Error("Int on a non-number should return 0")
		}
		if dec.Err() != json.ErrKind {
			t.Error("Int on a non-number should yield ErrKind")
		}
		dec.Free()
	}

	// Str must not hand back the raw bytes of a literal or a number.
	dec := json.NewDecoder(mem.System, []byte(`true`))
	defer dec.Free()
	dec.Next()
	if dec.Str() != "" {
		t.Error("Str on a bool should return the empty string")
	}
	if dec.Err() != json.ErrKind {
		t.Error("Str on a bool should yield ErrKind")
	}

	// Bool is false for a non-bool, as before, but now says so.
	dec2 := json.NewDecoder(mem.System, []byte(`12`))
	defer dec2.Free()
	dec2.Next()
	if dec2.Bool() {
		t.Error("Bool on a number should return false")
	}
	if dec2.Err() != json.ErrKind {
		t.Error("Bool on a number should yield ErrKind")
	}
}

func TestDecodeGetterKeepsFirstError(t *testing.T) {
	// A getter is a read: it must not overwrite the error that already stopped
	// the decode, nor parse a token the scanner never finished.
	dec := json.NewDecoder(mem.System, []byte(`[1,]`)) // trailing comma
	defer dec.Free()
	for dec.Next() {
	}
	if dec.Err() != json.ErrSyntax {
		t.Error("expected ErrSyntax")
		return
	}
	if dec.Int() != 0 {
		t.Error("a getter on a failed decode should return the zero value")
	}
	if dec.Err() != json.ErrSyntax {
		t.Error("a getter must not mask the first error")
	}
}

func TestDecodeNumberNotInt(t *testing.T) {
	// 1.5 is a number JSON allows but an int64 cannot hold. The kind is right,
	// so this an ErrValue and not ErrKind.
	dec := json.NewDecoder(mem.System, []byte(`1.5`))
	defer dec.Free()
	dec.Next()
	if dec.Float() != 1.5 {
		t.Error("Float should read 1.5")
	}
	if dec.Err() != nil {
		t.Error("1.5 is a valid float")
	}
	if dec.Int() != 0 {
		t.Error("Int on 1.5 should return 0")
	}
	if dec.Err() != json.ErrValue {
		t.Error("Int on 1.5 should set ErrValue")
	}
}

func TestDecodeUint(t *testing.T) {
	// MaxUint64 is a number JSON allows but an int64 cannot hold, which is the
	// whole reason Uint exists. Int reports it as an ErrValue; Uint reads it.
	dec := json.NewDecoder(mem.System, []byte(`18446744073709551615`))
	defer dec.Free()
	dec.Next()
	if dec.Uint() != math.MaxUint64 {
		t.Error("Uint should read MaxUint64")
	}
	if dec.Err() != nil {
		t.Error("MaxUint64 is a valid uint64")
	}

	// An out-of-range number reads as zero, not as the nearest uint64. A caller
	// that has not checked Err yet must not mistake a clamped value for a real one.
	dec2 := json.NewDecoder(mem.System, []byte(`18446744073709551616`)) // MaxUint64+1
	defer dec2.Free()
	dec2.Next()
	if dec2.Uint() != 0 {
		t.Error("Uint past MaxUint64 should return 0")
	}
	if dec2.Err() != json.ErrValue {
		t.Error("Uint past MaxUint64 should set ErrValue")
	}

	// A uint64 cannot be negative, and the kind is still right, so this is an
	// ErrValue like any other number the getter cannot represent.
	dec3 := json.NewDecoder(mem.System, []byte(`-1`))
	defer dec3.Free()
	dec3.Next()
	if dec3.Uint() != 0 {
		t.Error("Uint on a negative number should return 0")
	}
	if dec3.Err() != json.ErrValue {
		t.Error("Uint on a negative number should set ErrValue")
	}

	// Nor can it be fractional.
	dec4 := json.NewDecoder(mem.System, []byte(`1.5`))
	defer dec4.Free()
	dec4.Next()
	if dec4.Uint() != 0 {
		t.Error("Uint on a non-integer should return 0")
	}
	if dec4.Err() != json.ErrValue {
		t.Error("Uint on a non-integer should set ErrValue")
	}

	// A number that arrived as a string is the wrong kind, not a bad value.
	dec5 := json.NewDecoder(mem.System, []byte(`"7"`))
	defer dec5.Free()
	dec5.Next()
	if dec5.Uint() != 0 {
		t.Error("Uint on a string should return 0")
	}
	if dec5.Err() != json.ErrKind {
		t.Error("Uint on a string should set ErrKind")
	}
}

func TestDecodeFloatOutOfRange(t *testing.T) {
	// 1e400 is valid JSON with no float64 representation. Float must not hand
	// back +Inf. The kind is right, so this is an ErrValue.
	dec := json.NewDecoder(mem.System, []byte(`1e400`))
	defer dec.Free()
	dec.Next()
	if dec.Float() != 0 {
		t.Error("Float past MaxFloat64 should return 0")
	}
	if dec.Err() != json.ErrValue {
		t.Error("Float past MaxFloat64 should set ErrValue")
	}
}

func TestDecodeFreeClearsToken(t *testing.T) {
	// Free releases the memory the current token points into, so it must drop
	// the token as well. Otherwise Str would hand back a view of freed memory.
	dec := json.NewDecoder(mem.System, []byte(`"a\nb"`))
	if !dec.Next() || dec.Str() != "a\nb" {
		t.Fatal("decode failed")
		return
	}
	dec.Free()
	if dec.Str() != "" {
		t.Error("Str after Free should return the empty string")
	}
}

func TestDecodeIntOutOfRange(t *testing.T) {
	// MaxInt64+1 does not fit an int64. Int must not hand back the nearest one.
	dec := json.NewDecoder(mem.System, []byte(`9223372036854775808`))
	defer dec.Free()
	dec.Next()
	if dec.Int() != 0 {
		t.Error("Int past MaxInt64 should return 0, not a clamped value")
	}
	if dec.Err() != json.ErrValue {
		t.Error("Int past MaxInt64 should set ErrValue")
	}
}

func TestDecodeStream(t *testing.T) {
	src := `{"name":"Bob","nums":[1,2,3],"tall":true}`
	br := bytes.NewReader([]byte(src))
	dec := json.NewReader(mem.System, &br)
	defer dec.Free()

	if !dec.Next() || dec.Kind() != json.KindObjBeg {
		t.Fatal("expected object begin")
		return
	}

	var nameLen int
	var nameBuf [8]byte
	var sum int64
	var tall bool
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "name":
			dec.Next()
			nameLen = copy(nameBuf[:], dec.Str())
		case "nums":
			dec.Next()
			for dec.Next() && dec.Kind() == json.KindNumber {
				sum += dec.Int()
			}
		case "tall":
			dec.Next()
			tall = dec.Bool()
		default:
			dec.Next()
			dec.Skip()
		}
	}

	if string(nameBuf[:nameLen]) != "Bob" {
		t.Error("name mismatch")
	}
	if sum != 6 {
		t.Error("nums sum mismatch")
	}
	if !tall {
		t.Error("tall mismatch")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

// slowReader hands out one byte per Read, forcing the decoder's buffer to
// compact and refill repeatedly so the mark-preservation invariant is
// exercised across token boundaries.
type slowReader struct {
	data []byte
	pos  int
}

func (r *slowReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

func TestDecodeStreamChunked(t *testing.T) {
	// One byte per Read: every peek that needs data triggers a refill,
	// so the buffer compacts constantly.
	src := `{"msg":"a\"b\n\tcé","n":-12.5e2,"ok":false}`
	rd := slowReader{data: []byte(src)}
	dec := json.NewReader(mem.System, &rd)
	defer dec.Free()

	var msgLen int
	var msgBuf [16]byte
	var n float64
	var ok bool

	dec.Next() // {
	for dec.Next() && dec.Kind() == json.KindString {
		switch dec.Str() {
		case "msg":
			dec.Next()
			msgLen = copy(msgBuf[:], dec.Str())
		case "n":
			dec.Next()
			n = dec.Float()
		case "ok":
			dec.Next()
			ok = dec.Bool()
		default:
			dec.Next()
			dec.Skip()
		}
	}

	if string(msgBuf[:msgLen]) != "a\"b\n\tcé" {
		t.Error("msg mismatch")
	}
	if n != -1250.0 {
		t.Error("n mismatch")
	}
	if ok {
		t.Error("ok should be false")
	}
	if dec.Err() != nil {
		t.Error("unexpected decode error")
	}
}

func TestDecodeStreamInvalid(t *testing.T) {
	// The streaming path must validate exactly like the in-memory one.
	rd := slowReader{data: []byte(`{"a":1,}`)}
	dec := json.NewReader(mem.System, &rd)
	defer dec.Free()
	for dec.Next() {
	}
	if dec.Err() == nil {
		t.Error("expected error for trailing comma")
	}
}

// errReader hands out one byte per Read, then fails. It stands in for a
// connection that drops in the middle of a token.
type errReader struct {
	data []byte
	pos  int
}

var errBroken = errors.New("broken pipe")

func (r *errReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, errBroken
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

func TestDecodeTooLong(t *testing.T) {
	// A string token larger than MaxTokenSize must report ErrTooLong, not a
	// syntax error: the document is well-formed, it just does not fit. The same
	// document decodes in TestDecodeMaxToken, which differs only in the options.
	size := json.MaxTokenSize + 1024
	doc := oversizedDoc(size)
	defer mem.FreeSlice(mem.System, doc)

	br := bytes.NewReader(doc)
	dec := json.NewReader(mem.System, &br)
	defer dec.Free()
	for dec.Next() {
	}
	if dec.Err() != json.ErrTooLong {
		t.Error("oversized token should yield ErrTooLong")
	}
}

func TestDecodeMaxToken(t *testing.T) {
	// The document that TestDecodeTooLong rejects decodes fine once the caller
	// raises the limit. A single JSON string carrying a base64 blob or a
	// certificate runs past MaxTokenSize easily, and Str hands back a
	// contiguous view, so that token has to fit somewhere.
	//
	// Note that BufSize is left alone: raising the ceiling does not prepay for
	// it. The buffer still starts at the default and grows into the big token
	// only because this document actually carries one.
	size := json.MaxTokenSize + 1024
	doc := oversizedDoc(size)
	defer mem.FreeSlice(mem.System, doc)

	br := bytes.NewReader(doc)
	opts := json.ReaderOptions{MaxTokenSize: size}
	dec := json.NewReaderWith(mem.System, &br, opts)
	defer dec.Free()
	strLen := 0
	for dec.Next() {
		if dec.Kind() == json.KindString {
			strLen = len(dec.Str())
		}
	}
	if dec.Err() != nil {
		t.Error("an oversized token should decode once MaxTokenSize allows it")
	}
	if strLen != size-4 { // the document without its brackets and quotes
		t.Error("string length mismatch")
	}

	// The limit can be lowered as well as raised, to cap what an untrusted
	// document can make the decoder hold. BufSize goes down with it: a buffer
	// larger than the limit would raise it back (see TestDecodeBufSizeWins).
	br2 := bytes.NewReader([]byte(`["abcdefghijklmnopqrstuvwxyz"]`))
	opts2 := json.ReaderOptions{BufSize: 16, MaxTokenSize: 16}
	dec2 := json.NewReaderWith(mem.System, &br2, opts2)
	defer dec2.Free()
	for dec2.Next() {
	}
	if dec2.Err() != json.ErrTooLong {
		t.Error("a token larger than MaxTokenSize should yield ErrTooLong")
	}
}

func TestDecodeTokenSizeExact(t *testing.T) {
	size := 64
	doc := oversizedDoc(size)
	defer mem.FreeSlice(mem.System, doc)
	tokLen := size - 4

	// A token of exactly MaxTokenSize fits.
	br := bytes.NewReader(doc)
	opts := json.ReaderOptions{BufSize: tokLen, MaxTokenSize: tokLen}
	dec := json.NewReaderWith(mem.System, &br, opts)
	defer dec.Free()
	strLen := 0
	for dec.Next() {
		if dec.Kind() == json.KindString {
			strLen = len(dec.Str())
		}
	}
	if dec.Err() != nil {
		t.Error("a token of exactly MaxTokenSize should decode")
	}
	if strLen != tokLen {
		t.Error("string length mismatch")
	}

	// One byte past it does not.
	br2 := bytes.NewReader(doc)
	opts2 := json.ReaderOptions{BufSize: tokLen - 1, MaxTokenSize: tokLen - 1}
	dec2 := json.NewReaderWith(mem.System, &br2, opts2)
	defer dec2.Free()
	for dec2.Next() {
	}
	if dec2.Err() != json.ErrTooLong {
		t.Error("a token one byte over MaxTokenSize should yield ErrTooLong")
	}
}

func TestDecodeNumberTruncated(t *testing.T) {
	// A number is the one token with no closing delimiter: the scanner stops at
	// any non-digit, and a buffer that has filled up looks the same to it. The
	// digits it did capture must not be handed back as a complete number.
	doc := make([]byte, 40)
	for i := range doc {
		doc[i] = '1'
	}

	rd := slowReader{data: doc}
	opts := json.ReaderOptions{BufSize: 20, MaxTokenSize: 20}
	dec := json.NewReaderWith(mem.System, &rd, opts)
	defer dec.Free()
	if dec.Next() {
		t.Error("Next accepted a truncated number")
	}
	if dec.Err() != json.ErrTooLong {
		t.Error("a number over MaxTokenSize should yield ErrTooLong")
	}
}

func TestDecodeBufSize(t *testing.T) {
	// A caller who knows its token size can allocate the buffer once, up front,
	// by starting it at the limit. The growth loop then never runs, which is
	// what an arena allocator wants: it cannot reuse the buffers that growing
	// leaves behind.
	//
	// The decode must come out the same as it does through the growth path in
	// TestDecodeMaxToken - the buffer's size is a memory decision, not a
	// semantic one.
	size := json.MaxTokenSize + 1024
	doc := oversizedDoc(size)
	defer mem.FreeSlice(mem.System, doc)

	br := bytes.NewReader(doc)
	opts := json.ReaderOptions{BufSize: size, MaxTokenSize: size}
	dec := json.NewReaderWith(mem.System, &br, opts)
	defer dec.Free()
	strLen := 0
	for dec.Next() {
		if dec.Kind() == json.KindString {
			strLen = len(dec.Str())
		}
	}
	if dec.Err() != nil {
		t.Error("a buffer sized to the token up front should decode it")
	}
	if strLen != size-4 {
		t.Error("string length mismatch")
	}

	// A tiny BufSize is legal: it just means more growing. It is not a limit,
	// so a token far larger than it still decodes.
	br2 := bytes.NewReader([]byte(`["abcdefghijklmnopqrstuvwxyz"]`))
	opts2 := json.ReaderOptions{BufSize: 1}
	dec2 := json.NewReaderWith(mem.System, &br2, opts2)
	defer dec2.Free()
	str := ""
	for dec2.Next() {
		if dec2.Kind() == json.KindString {
			str = dec2.Str()
		}
	}
	if dec2.Err() != nil {
		t.Error("a small BufSize should grow, not fail")
	}
	if str != "abcdefghijklmnopqrstuvwxyz" {
		t.Error("string value mismatch")
	}
}

// roomReader is a reader over an in-memory document that remembers how much room
// the decoder offered it on the first Read, which is the decoder's buffer size.
type roomReader struct {
	data  []byte
	pos   int
	first int
}

func (r *roomReader) Read(p []byte) (int, error) {
	if r.first == 0 {
		r.first = len(p)
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func TestDecodeBufSizeWins(t *testing.T) {
	// BufSize is not lowered to fit MaxTokenSize, so a small token limit does not
	// shrink the reads. The effective limit is the larger of the two: this 508-byte
	// token is over MaxTokenSize but well under BufSize, so the buffer holds it and
	// it decodes. Same rule as bufio.Scanner.Buffer. Lower both options to cap what
	// an untrusted document can make the decoder hold.
	bufSize := 4096
	doc := oversizedDoc(512)
	defer mem.FreeSlice(mem.System, doc)

	rd := roomReader{data: doc}
	opts := json.ReaderOptions{BufSize: bufSize, MaxTokenSize: 64}
	dec := json.NewReaderWith(mem.System, &rd, opts)
	defer dec.Free()
	strLen := 0
	for dec.Next() {
		if dec.Kind() == json.KindString {
			strLen = len(dec.Str())
		}
	}
	if dec.Err() != nil {
		t.Error("a token under BufSize should decode whatever MaxTokenSize says")
	}
	if strLen != 508 { // the document without its brackets and quotes
		t.Error("string length mismatch")
	}
	if rd.first != bufSize {
		t.Error("MaxTokenSize should not shrink the read buffer")
	}
}

func TestDecodeReadError(t *testing.T) {
	// A read failure mid-token must surface the reader's error rather than be
	// misreported as a syntax error.
	rd := errReader{data: []byte(`{"abc":"defg`)}
	dec := json.NewReader(mem.System, &rd)
	defer dec.Free()
	for dec.Next() {
	}
	if dec.Err() != errBroken {
		t.Error("read failure in a string should surface the reader error")
	}

	// Same for a number, which ends at any non-digit: the truncated token must
	// not be handed to the caller as a valid value.
	rd2 := errReader{data: []byte(`[12345`)}
	dec2 := json.NewReader(mem.System, &rd2)
	defer dec2.Free()
	for dec2.Next() {
		if dec2.Kind() == json.KindNumber && dec2.Int() != 12345 {
			t.Error("decoder emitted a truncated number")
		}
	}
	if dec2.Err() != errBroken {
		t.Error("read failure in a number should surface the reader error")
	}
}

// oversizedDoc builds ["xxx...x"], size bytes long, holding a single string
// token of size-4 bytes (the document without its brackets and quotes). At the
// sizes its callers pass, that token runs past the default MaxTokenSize.
// Free it with mem.FreeSlice.
func oversizedDoc(size int) []byte {
	doc := mem.AllocSlice[byte](mem.System, size, size)
	for i := range doc {
		doc[i] = 'x'
	}
	doc[0] = '['
	doc[1] = '"'
	doc[size-2] = '"'
	doc[size-1] = ']'
	return doc
}
