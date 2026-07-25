package json

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strconv"
	"solod.dev/so/unicode/utf8"
)

// MaxTokenSize is the largest token (a string or a number) that a streaming
// decoder will hold by default. A larger one yields [ErrTooLong]; see
// [ReaderOptions] to raise or lower the limit.
const MaxTokenSize = 64 * 1024

// Kind describes the type of the current JSON token.
type Kind byte

const (
	KindInvalid Kind = 0
	KindNull    Kind = 'n'
	KindBool    Kind = 'b'
	KindNumber  Kind = '0'
	KindString  Kind = '"'
	KindObjBeg  Kind = '{'
	KindObjEnd  Kind = '}'
	KindArrBeg  Kind = '['
	KindArrEnd  Kind = ']'
)

// Decoder walks the tokens of a JSON document, validating the grammar as it
// goes. It holds the current token; query it with [Decoder.Kind] and the typed
// getters.
//
// Each getter reads one kind of token and nothing else: [Decoder.Int] on a
// string sets [ErrKind] rather than parsing it. Check [Decoder.Kind] first
// wherever the document may not have the shape the code expects.
//
// The byte source for a Decoder can be a complete in-memory document
// (see [NewDecoder]) or a stream pulled from an [io.Reader] (see [NewReader]).
type Decoder struct {
	b   buffer
	knd Kind
	tok []byte // raw bytes of the current string (unquoted) or number
	scr []byte // scratch for unescaping; see Decoder.scratch

	// Each nesting level is tracked as a frame: 't' (top), 'o' (object),
	// or 'a' (array), with a count of the tokens emitted so far. In an object,
	// even counts are key positions and odd counts are value positions; that
	// is how the mandatory ':' and ',' separators are enforced.
	stack [MaxDepth + 1]byte
	count [MaxDepth + 1]int
	depth int

	err error // sticky error (syntax, read, or buffer)
}

const (
	frameTop byte = 't'
	frameObj byte = 'o'
	frameArr byte = 'a'
)

// NewDecoder returns a [Decoder] over a complete in-memory JSON document. The
// document is read, never written, and never copied: doc may be a read-only
// mapping or a constant, and it is still intact once decoding is done.
//
// The only thing such a decoder allocates is the scratch buffer it unescapes
// strings into. A document with no escape sequences costs nothing; one with
// escapes costs a single buffer, sized to its longest escaped string.
//
// As with [NewReader], a token returned by [Decoder.Str] is only guaranteed to
// be valid until the next call to [Decoder.Next]: an escaped string is
// unescaped into the scratch buffer, which the next escaped string reuses.
//
// Call [Decoder.Free] when done using it to release its resources.
func NewDecoder(alloc mem.Allocator, doc []byte) Decoder {
	d := Decoder{b: fixedBuffer(alloc, doc)}
	d.stack[0] = frameTop
	return d
}

// NewReader returns a [Decoder] that streams JSON from r, using alloc to
// allocate its internal buffer. The document may be arbitrarily large;
// memory is bounded by the largest single token (a string or number), which
// must fit in [MaxTokenSize] or decoding fails with [ErrTooLong]. Use
// [NewReaderWith] to size the buffer or move that limit.
//
// A value returned by [Decoder.Str] is valid only until the next call to
// [Decoder.Next]. Match an object key before advancing to its value, and
// copy out any token that must outlive the loop:
//
//	for d.Next() && d.Kind() == KindString {
//		switch d.Str() { // the key, still valid here
//		case "name":
//			d.Next() // the key is gone; the value is current
//			n = copy(name[:], d.Str())
//		default:
//			d.Next()
//			d.Skip()
//		}
//	}
//
// Call [Decoder.Free] when done using it to release the internal buffer.
func NewReader(alloc mem.Allocator, r io.Reader) Decoder {
	return NewReaderWith(alloc, r, ReaderOptions{})
}

// ReaderOptions configures the buffer a streaming [Decoder] reassembles tokens in.
// The zero value is the default: a buffer that starts small and grows on
// demand, up to [MaxTokenSize].
type ReaderOptions struct {
	// BufSize is the initial size of the buffer in bytes, and thus the size of
	// the reads the decoder issues. The buffer grows past it only when a token
	// requires it. If 0, the buffer starts at 4096 bytes.
	BufSize int

	// MaxTokenSize is the largest token (a string or a number) the decoder will
	// hold, in bytes. A larger one yields [ErrTooLong]. If 0, the limit is
	// [MaxTokenSize].
	//
	// The buffer is never made smaller than BufSize to fit this limit, so the
	// effective maximum token size is max(BufSize, MaxTokenSize).
	MaxTokenSize int
}

// NewReaderWith is like [NewReader], but with the buffer configured by opts.
//
// Call [Decoder.Free] when done using it to release its resources.
func NewReaderWith(alloc mem.Allocator, r io.Reader, opts ReaderOptions) Decoder {
	bufSize := opts.BufSize
	if bufSize == 0 {
		bufSize = defaultBufSize
	}
	maxTok := opts.MaxTokenSize
	if maxTok == 0 {
		maxTok = MaxTokenSize
	}
	d := Decoder{b: streamBuffer(alloc, r, bufSize, maxTok)}
	d.stack[0] = frameTop
	return d
}

// Free releases the memory the decoder allocated.
func (d *Decoder) Free() {
	if d.scr != nil {
		mem.FreeSlice(d.b.alloc, d.scr)
		d.scr = nil
	}
	d.b.free()
	d.tok = nil
}

// Kind returns the kind of the current token.
func (d *Decoder) Kind() Kind { return d.knd }

// Depth returns the current nesting depth: 0 at the top level
// and increases with each nested object or array.
func (d *Decoder) Depth() int { return d.depth }

// Err returns the first error encountered, or nil.
// A read error or [ErrTooLong] outranks the scanners' [ErrSyntax].
func (d *Decoder) Err() error {
	// Preferring the buffer's error cannot hide a real syntax error: once the
	// buffer fails it stays failed, and a scanner only reports bad syntax about
	// a byte it actually read.
	if err := d.bufErr(); err != nil {
		return err
	}
	return d.err
}

// Str returns the current string or object key. The result is a view, not a
// copy, and is valid only until the next call to [Decoder.Next].
//
// The current token must be a string, otherwise Str returns "" and sets
// [ErrKind].
func (d *Decoder) Str() string {
	if !d.expect(KindString) {
		return ""
	}
	return string(d.tok)
}

// Int parses the current number as an int64.
//
// The current token must be a number, otherwise Int returns 0 and sets
// [ErrKind]. A number that is not an integer, or that does not fit in an
// int64, returns 0 and sets [ErrValue]; use [Decoder.Uint] for an integer
// above the int64 range.
func (d *Decoder) Int() int64 {
	if !d.expect(KindNumber) {
		return 0
	}
	n, err := strconv.ParseInt(string(d.tok), 10, 64)
	if err != nil {
		d.err = ErrValue
		return 0
	}
	return n
}

// Uint parses the current number as a uint64, which reaches the integers above
// the int64 range that [Decoder.Int] cannot represent.
//
// The current token must be a number, otherwise Uint returns 0 and sets
// [ErrKind]. A number that is not an integer, is negative, or does not fit in
// a uint64, returns 0 and sets [ErrValue].
func (d *Decoder) Uint() uint64 {
	if !d.expect(KindNumber) {
		return 0
	}
	n, err := strconv.ParseUint(string(d.tok), 10, 64)
	if err != nil {
		d.err = ErrValue
		return 0
	}
	return n
}

// Float parses the current number as a float64.
//
// The current token must be a number, otherwise Float returns 0 and sets [ErrKind].
// A number that does not fit in a float64 returns 0 and sets [ErrValue].
func (d *Decoder) Float() float64 {
	if !d.expect(KindNumber) {
		return 0
	}
	f, err := strconv.ParseFloat(string(d.tok), 64)
	if err != nil {
		d.err = ErrValue
		return 0
	}
	return f
}

// Bool returns the current boolean value.
//
// The current token must be a boolean, otherwise
// Bool returns false and sets [ErrKind].
func (d *Decoder) Bool() bool {
	if !d.expect(KindBool) {
		return false
	}
	return len(d.tok) == 4 // "true" vs "false"
}

// expect reports whether the current token is of kind k, setting [ErrKind] if
// it is not. The wrong kind is reported through the sticky error like any other
// decode failure: it is what valid JSON of an unexpected shape looks like.
//
// It also refuses once an error is set, so a getter cannot mask the first error
// or parse a token the scanner never finished.
func (d *Decoder) expect(k Kind) bool {
	if d.err != nil {
		return false
	}
	if d.knd != k {
		d.err = ErrKind
		return false
	}
	return true
}

// Skip consumes the current value. For a container token it advances to the
// matching close token.
func (d *Decoder) Skip() {
	depth := 0
	if d.knd == KindObjBeg || d.knd == KindArrBeg {
		depth = 1
	}
	for depth > 0 && d.Next() {
		switch d.knd {
		case KindObjBeg, KindArrBeg:
			depth++
		case KindObjEnd, KindArrEnd:
			depth--
		}
	}
}

// Next advances to the next token and reports whether one was found. It sets a
// sticky error (see [Decoder.Err]) and returns false on the first syntax error
// or once the single root value has been consumed.
func (d *Decoder) Next() bool {
	if d.err != nil {
		return false
	}
	d.skipWS()
	cur := d.b.peek()

	// After the root value, only trailing whitespace is allowed.
	if d.depth == 0 && d.count[0] > 0 {
		if cur >= 0 {
			d.err = ErrSyntax
		}
		return false
	}
	// The input ended before a complete root value: either a container
	// is still open, or there was no value at all (an empty document).
	// The consumed-root case returned above, so both are syntax errors.
	if cur < 0 {
		d.err = ErrSyntax
		return false
	}
	if !d.transition() {
		return false
	}

	cur = d.b.peek()
	if cur < 0 {
		d.err = ErrSyntax
		return false
	}

	switch {
	case cur == '{':
		// Count the container in the enclosing frame before pushing its own.
		d.count[d.depth]++
		if !d.push(frameObj) {
			return false
		}
		d.setContainer(KindObjBeg)
	case cur == '}':
		if d.stack[d.depth] != frameObj || d.count[d.depth]%2 != 0 {
			d.err = ErrSyntax
			return false
		}
		d.pop()
		d.setContainer(KindObjEnd)
	case cur == '[':
		d.count[d.depth]++
		if !d.push(frameArr) {
			return false
		}
		d.setContainer(KindArrBeg)
	case cur == ']':
		if d.stack[d.depth] != frameArr {
			d.err = ErrSyntax
			return false
		}
		d.pop()
		d.setContainer(KindArrEnd)
	case cur == '"':
		if !d.scanString() {
			return false
		}
		d.count[d.depth]++
	case cur == 't' || cur == 'f':
		if !d.scanBool(cur) {
			return false
		}
		d.count[d.depth]++
	case cur == 'n':
		if !d.scanLiteral("null", KindNull) {
			return false
		}
		d.count[d.depth]++
	case cur == '-' || (cur >= '0' && cur <= '9'):
		if !d.scanNumber() {
			return false
		}
		d.count[d.depth]++
	default:
		d.err = ErrSyntax
		return false
	}
	return true
}

// transition consumes the ',' and ':' separators required before the upcoming
// token and validates that the token is allowed in the current frame. On
// return the cursor points at the start of the token (a value, an object key,
// or a closing bracket).
func (d *Decoder) transition() bool {
	frame := d.stack[d.depth]
	ntok := d.count[d.depth]
	if frame == frameArr {
		if d.b.peek() == ']' {
			return true
		}
		if ntok > 0 {
			if !d.expectSep(',') {
				return false
			}
			cur := d.b.peek()
			if cur < 0 || cur == ']' {
				d.err = ErrSyntax // trailing comma
				return false
			}
		}
		return true
	}
	if frame == frameObj {
		if ntok%2 == 1 {
			return d.expectSep(':') // value position
		}
		if d.b.peek() == '}' {
			return true
		}
		if ntok > 0 {
			if !d.expectSep(',') {
				return false
			}
		}
		if d.b.peek() != '"' {
			d.err = ErrSyntax // object key must be a string
			return false
		}
		return true
	}
	return true // top level: a single root value
}

// expectSep consumes the separator byte ch and the whitespace after it,
// erroring if ch is absent or the input ends.
func (d *Decoder) expectSep(ch byte) bool {
	if d.b.peek() != int(ch) {
		d.err = ErrSyntax
		return false
	}
	d.b.advance()
	d.skipWS()
	if d.b.peek() < 0 {
		d.err = ErrSyntax
		return false
	}
	return true
}

// scanString captures the current string token. It brackets the bytes between
// the quotes with hold/take, then unescapes them if needed.
func (d *Decoder) scanString() bool {
	d.b.advance() // opening quote
	d.b.hold()
	escaped := false
	nonASCII := false
	for {
		cur := d.b.peek()
		if cur < 0 {
			d.err = ErrSyntax
			return false
		}
		if cur == '"' {
			break
		}
		if cur < 0x20 {
			d.err = ErrSyntax // control characters must be escaped
			return false
		}
		if cur == '\\' {
			escaped = true
			d.b.advance() // backslash
			if d.b.peek() < 0 {
				d.err = ErrSyntax
				return false
			}
			d.b.advance() // escaped char, so \" is not seen as the end
			continue
		}
		if cur >= utf8.RuneSelf {
			nonASCII = true // needs the UTF-8 check below
		}
		d.b.advance()
	}
	raw := d.b.take()
	d.b.advance() // closing quote
	d.knd = KindString

	d.tok = raw
	if escaped {
		// Unescape in place when the buffer owns the bytes. A fixed buffer
		// borrows the caller's document, which may be read-only and must stay
		// intact, so unescape into the scratch instead.
		dst := raw
		if !d.b.owned() {
			dst = d.scratch(len(raw))
		}
		n, ok := unescape(dst, raw)
		if !ok {
			d.err = ErrValue
			return false
		}
		d.tok = dst[:n]
	}
	// A JSON string must be valid UTF-8.
	if nonASCII && !utf8.Valid(d.tok) {
		d.err = ErrValue
		return false
	}
	return true
}

// scratch returns a decoder-owned buffer of at least n bytes to unescape into.
//
// Unescaping only shrinks a string, so the raw token length is an exact upper
// bound and the scratch never grows mid-token. It is reused across tokens and
// grows only for an escaped string longer than any before it. It is not
// allocated until the document turns out to contain an escape.
func (d *Decoder) scratch(n int) []byte {
	if len(d.scr) >= n {
		return d.scr
	}
	if d.scr != nil {
		mem.FreeSlice(d.b.alloc, d.scr)
		d.scr = nil
	}
	d.scr = mem.AllocSlice[byte](d.b.alloc, n, n)
	return d.scr
}

// scanNumber validates and captures a JSON number:
// -?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?
func (d *Decoder) scanNumber() bool {
	d.b.hold()
	cur := d.b.peek()
	if cur == '-' {
		d.b.advance()
		cur = d.b.peek()
	}
	// Integer part.
	if cur < '0' || cur > '9' {
		d.err = ErrSyntax
		return false
	}
	if cur == '0' {
		d.b.advance()
	} else {
		d.b.advance()
		for {
			cur = d.b.peek()
			if cur < '0' || cur > '9' {
				break
			}
			d.b.advance()
		}
	}
	// Fraction.
	cur = d.b.peek()
	if cur == '.' {
		d.b.advance()
		cur = d.b.peek()
		if cur < '0' || cur > '9' {
			d.err = ErrSyntax
			return false
		}
		for {
			cur = d.b.peek()
			if cur < '0' || cur > '9' {
				break
			}
			d.b.advance()
		}
	}
	// Exponent.
	cur = d.b.peek()
	if cur == 'e' || cur == 'E' {
		d.b.advance()
		cur = d.b.peek()
		if cur == '+' || cur == '-' {
			d.b.advance()
			cur = d.b.peek()
		}
		if cur < '0' || cur > '9' {
			d.err = ErrSyntax
			return false
		}
		for {
			cur = d.b.peek()
			if cur < '0' || cur > '9' {
				break
			}
			d.b.advance()
		}
	}
	// A number has no closing delimiter, so the loops above end at any
	// non-digit, including the -1 that peek returns on a buffer error. Check for
	// that error here, or a truncated number would look like a complete token.
	if err := d.bufErr(); err != nil {
		d.err = err
		return false
	}
	d.tok = d.b.take()
	d.knd = KindNumber
	return true
}

// bufErr returns the buffer's error if it is a genuine failure (a read error
// or [ErrTooLong]), and nil at a plain EOF.
func (d *Decoder) bufErr() error {
	if d.b.err != nil && d.b.err != io.EOF {
		return d.b.err
	}
	return nil
}

// scanBool captures the current boolean token.
func (d *Decoder) scanBool(c int) bool {
	if c == 't' {
		return d.scanLiteral("true", KindBool)
	}
	return d.scanLiteral("false", KindBool)
}

// scanLiteral captures the current literal token (null, true, or false).
func (d *Decoder) scanLiteral(lit string, k Kind) bool {
	d.b.hold()
	for i := 0; i < len(lit); i++ {
		if d.b.peek() != int(lit[i]) {
			d.err = ErrSyntax
			return false
		}
		d.b.advance()
	}
	d.tok = d.b.take()
	d.knd = k
	return true
}

// setContainer accepts a bracket token. A container carries no value, so it
// clears the token bytes. The getters reject a bracket on its kind alone, but
// leaving the previous token in place would keep stale bytes reachable.
func (d *Decoder) setContainer(k Kind) {
	d.knd = k
	d.tok = nil
	d.b.advance()
}

// push adds a new frame to the stack, reporting whether it fits in MaxDepth.
func (d *Decoder) push(frame byte) bool {
	if d.depth >= MaxDepth {
		d.err = ErrDepth
		return false
	}
	d.depth++
	d.stack[d.depth] = frame
	d.count[d.depth] = 0
	return true
}

// pop removes the top frame from the stack.
func (d *Decoder) pop() {
	if d.depth > 0 {
		d.depth--
	}
}

// skipWS consumes whitespace until the next non-space byte or EOF.
func (d *Decoder) skipWS() {
	for {
		cur := d.b.peek()
		if cur == ' ' || cur == '\t' || cur == '\n' || cur == '\r' {
			d.b.advance()
			continue
		}
		return
	}
}

// unescape writes src to dst, decoding escape sequences, and returns the length
// written. It reports false on an invalid escape.
//
// dst must be at least as long as src, and may be src itself: an escape is
// always longer than its decoded form (\n is 2 bytes for 1, a \uXXXX surrogate
// pair is 12 for at most 4), so the write cursor never overtakes the read one.
func unescape(dst, src []byte) (int, bool) {
	w := 0
	r := 0
	for r < len(src) {
		cur := src[r]
		if cur != '\\' {
			dst[w] = cur
			w++
			r++
			continue
		}
		r++ // consume backslash; scanString guarantees a following byte
		cur = src[r]
		switch cur {
		case '"', '\\', '/':
			dst[w] = cur
			w++
		case 'b':
			dst[w] = '\b'
			w++
		case 'f':
			dst[w] = '\f'
			w++
		case 'n':
			dst[w] = '\n'
			w++
		case 'r':
			dst[w] = '\r'
			w++
		case 't':
			dst[w] = '\t'
			w++
		case 'u':
			nw, nr := unescapeU(dst, src, w, r)
			if nr < 0 {
				return 0, false
			}
			w = nw
			r = nr
		default:
			return 0, false
		}
		r++
	}
	return w, true
}

// unescapeU decodes a \uXXXX escape (r points at the 'u' in src), writing the
// UTF-8 encoding at dst[w]. It combines a high/low surrogate pair. It returns
// the advanced write and read cursors (r left on the last consumed byte), or a
// read cursor of -1 on malformed hex or a surrogate that is not part of a pair.
func unescapeU(dst, src []byte, w, r int) (int, int) {
	hi, ok := hex4(src, r+1)
	if !ok {
		return w, -1
	}
	r += 4 // r now on the last hex digit
	rn := rune(hi)
	if hi >= 0xDC00 && hi <= 0xDFFF {
		return w, -1 // a low surrogate with no high one before it
	}
	if hi >= 0xD800 && hi <= 0xDBFF {
		// A high surrogate is half a code point; a low surrogate escape
		// (\uXXXX) must follow it.
		if r+2 >= len(src) || src[r+1] != '\\' || src[r+2] != 'u' {
			return w, -1
		}
		lo, ok := hex4(src, r+3)
		if !ok || lo < 0xDC00 || lo > 0xDFFF {
			return w, -1
		}
		rn = 0x10000 + (rune(hi-0xD800) << 10) + rune(lo-0xDC00)
		r += 6
	}
	n := utf8.EncodeRune(dst[w:], rn)
	return w + n, r
}

// hex4 decodes four hex digits starting at buf[i].
// Returns false for an invalid input.
func hex4(buf []byte, i int) (int, bool) {
	if i+4 > len(buf) {
		return 0, false
	}
	val := 0
	for j := range 4 {
		cur := buf[i+j]
		val <<= 4
		switch {
		case cur >= '0' && cur <= '9':
			val |= int(cur - '0')
		case cur >= 'a' && cur <= 'f':
			val |= int(cur-'a') + 10
		case cur >= 'A' && cur <= 'F':
			val |= int(cur-'A') + 10
		default:
			return 0, false
		}
	}
	return val, true
}
