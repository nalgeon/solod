package json

import (
	"solod.dev/so/io"
	"solod.dev/so/math"
	"solod.dev/so/strconv"
	"solod.dev/so/unicode/utf8"
)

const (
	hexDigits = "0123456789abcdef"

	// outSize is the size of the Encoder's output buffer. Tokens are assembled
	// here and handed to the writer in batches; without it a short document
	// would cost a Write call per bracket, quote, separator and escape.
	outSize = 512
)

// Encoder writes JSON tokens to an [io.Writer], inserting commas and colons
// automatically based on the current container nesting. Inside an object the
// tokens alternate key, value, key, value.
//
// Like the [Decoder], the Encoder is strict about the call sequence it is
// given. A stray or mismatched closing token, an object key that is not a
// string, a key without a value, a non-finite number, a string that is not
// UTF-8, or a second root value sets a sticky error (see [Encoder.Err]), after
// which nothing more is written.
//
// The Encoder buffers its output, so call [Encoder.Flush]
// once the root value is complete.
type Encoder struct {
	w       io.Writer
	count   [MaxDepth]int  // tokens written at each nesting level
	isObj   [MaxDepth]bool // whether each level is an object (vs an array)
	depth   int
	hasRoot bool // the single root value a document holds has been written
	err     error

	numBuf [32]byte      // number formatting scratch; enough for any float64 or int64
	outBuf [outSize]byte // pending output
	outN   int           // bytes pending in outBuf
}

// NewEncoder returns an Encoder that writes to w. The Encoder buffers its
// output; call [Encoder.Flush] when the document is complete.
func NewEncoder(w io.Writer) Encoder {
	return Encoder{w: w}
}

// BeginObject writes '{'.
func (e *Encoder) BeginObject() {
	if !e.allowValue() || !e.allowNest() {
		return
	}
	e.sep()
	e.writeByte('{')
	e.push(true)
}

// EndObject writes '}'.
func (e *Encoder) EndObject() {
	if e.err != nil {
		return
	}
	if e.depth == 0 || !e.isObj[e.depth-1] || e.count[e.depth-1]%2 != 0 {
		e.err = ErrSyntax // no open object, or a key without a value
		return
	}
	e.pop()
	e.writeByte('}')
}

// BeginArray writes '['.
func (e *Encoder) BeginArray() {
	if !e.allowValue() || !e.allowNest() {
		return
	}
	e.sep()
	e.writeByte('[')
	e.push(false)
}

// EndArray writes ']'.
func (e *Encoder) EndArray() {
	if e.err != nil {
		return
	}
	if e.depth == 0 || e.isObj[e.depth-1] {
		e.err = ErrSyntax // no open array
		return
	}
	e.pop()
	e.writeByte(']')
}

// Str writes a quoted, escaped string. It is also used for object keys.
//
// If the string is not valid UTF-8, Str sets [ErrValue] and writes nothing.
func (e *Encoder) Str(s string) {
	if !e.allowToken() {
		return
	}
	esc, ok := needEscValidUTF(s)
	if !ok {
		e.err = ErrValue
		return
	}
	e.sep()
	e.writeByte('"')
	if esc {
		e.writeEscaped(s)
	} else {
		e.write(s) // nothing to escape, so it goes out as one copy
	}
	e.writeByte('"')
}

// Int writes an integer.
func (e *Encoder) Int(n int64) {
	if !e.allowValue() {
		return
	}
	e.sep()
	e.write(strconv.FormatInt(e.numBuf[:], n, 10))
}

// Float writes a floating-point number using the shortest round-trip form.
// JSON cannot spell NaN or an infinity, so a non-finite f is not written
// and yields [ErrNonFinite].
func (e *Encoder) Float(f float64) {
	if !e.allowValue() {
		return
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		e.err = ErrNonFinite
		return
	}
	e.sep()
	e.write(strconv.FormatFloat(e.numBuf[:], f, 'g', -1, 64))
}

// Bool writes true or false.
func (e *Encoder) Bool(b bool) {
	if !e.allowValue() {
		return
	}
	e.sep()
	if b {
		e.write("true")
	} else {
		e.write("false")
	}
}

// Null writes null.
func (e *Encoder) Null() {
	if !e.allowValue() {
		return
	}
	e.sep()
	e.write("null")
}

// Flush writes the buffered bytes to the underlying writer.
// Call it only once, after the root value is complete.
//
// Flush sets a sticky [ErrSyntax] error if the document is incomplete.
func (e *Encoder) Flush() {
	// Drain first, and drain even on a sticky error: the caller learns the
	// document is bad from Err, and withholding the bytes already written
	// would only make the mistake harder to see.
	e.drain()
	if e.err == nil && (e.depth != 0 || !e.hasRoot) {
		e.err = ErrSyntax
	}
}

// Err returns the first write error encountered, or nil.
func (e *Encoder) Err() error { return e.err }

// sep writes the separator that precedes the next token, if any, and bumps the
// token count for the current level.
func (e *Encoder) sep() {
	if e.depth == 0 {
		e.hasRoot = true // the root value; nothing may follow it
		return
	}
	n := e.count[e.depth-1]
	e.count[e.depth-1] = n + 1
	if n == 0 {
		return
	}
	// In an object, odd token indexes are values (preceded by ':'); everything
	// else, including all array elements, is comma-separated.
	if e.isObj[e.depth-1] && n%2 == 1 {
		e.writeByte(':')
	} else {
		e.writeByte(',')
	}
}

// allowToken reports whether any token may be written at the current position,
// setting an error if not. A JSON document holds a single root value, so once
// it is written nothing may follow it at the top level.
func (e *Encoder) allowToken() bool {
	if e.err != nil {
		return false
	}
	if e.depth == 0 && e.hasRoot {
		e.err = ErrSyntax // a document has a single root value
		return false
	}
	return true
}

// allowValue reports whether a value may be written at the current position,
// setting an error if not. Inside an object, even positions are keys, so only
// [Encoder.Str] may appear there.
func (e *Encoder) allowValue() bool {
	if !e.allowToken() {
		return false
	}
	if e.depth > 0 && e.isObj[e.depth-1] && e.count[e.depth-1]%2 == 0 {
		e.err = ErrSyntax // an object key must be a string
		return false
	}
	return true
}

// allowNest reports whether another nesting level may be opened, setting
// [ErrDepth] if not. It runs before the bracket is written, so an overflow
// does not emit a container that can never be closed.
func (e *Encoder) allowNest() bool {
	if e.depth >= MaxDepth {
		e.err = ErrDepth
		return false
	}
	return true
}

// push adds a new nesting level (an object or an array).
// The caller must have checked allowNest.
func (e *Encoder) push(obj bool) {
	e.count[e.depth] = 0
	e.isObj[e.depth] = obj
	e.depth++
}

// pop removes the current nesting level.
func (e *Encoder) pop() {
	if e.depth > 0 {
		e.depth--
	}
}

// writeEscaped writes s with the mandatory JSON escapes applied.
// Bytes that need no escaping are written in runs.
func (e *Encoder) writeEscaped(s string) {
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Encoder.Str has already checked that s is UTF-8, so every byte above ASCII
		// belongs to a well-formed rune, needs no escaping, and can be copied as it is.
		if c >= utf8.RuneSelf || (c != '"' && c != '\\' && c >= 0x20) {
			continue
		}
		e.writeRange(s, start, i)
		switch c {
		case '"':
			e.write("\\\"")
		case '\\':
			e.write("\\\\")
		case '\n':
			e.write("\\n")
		case '\r':
			e.write("\\r")
		case '\t':
			e.write("\\t")
		case '\b':
			e.write("\\b")
		case '\f':
			e.write("\\f")
		default:
			e.writeU(c)
		}
		start = i + 1
	}
	e.writeRange(s, start, len(s))
}

// writeU writes a control byte as a \u00XX escape.
func (e *Encoder) writeU(c byte) {
	var b [6]byte
	b[0], b[1], b[2], b[3] = '\\', 'u', '0', '0'
	b[4] = hexDigits[c>>4]
	b[5] = hexDigits[c&0x0f]
	e.write(string(b[:]))
}

// writeRange writes s[start:end] to the output buffer.
func (e *Encoder) writeRange(s string, start, end int) {
	if start < end {
		e.write(s[start:end])
	}
}

// write appends s to the output buffer, draining it whenever it fills. A token
// longer than the buffer (a long string) is written across several drains.
func (e *Encoder) write(s string) {
	if e.err != nil {
		return
	}
	// Fast path: the fragment already fits in the buffer, which is common for
	// short structural pieces and keys that the encoder emits.
	if e.outN+len(s) <= outSize {
		e.outN += copy(e.outBuf[e.outN:], s)
		return
	}
	for e.err == nil && len(s) > 0 {
		if e.outN == outSize {
			e.drain()
			continue
		}
		n := copy(e.outBuf[e.outN:], s)
		e.outN += n
		s = s[n:]
	}
}

// writeByte adds a single byte, flushing first if the buffer is full.
// It skips the copy loop and memmove for one-byte structural tokens
// (braces, brackets, quotes, separators).
func (e *Encoder) writeByte(c byte) {
	if e.err != nil {
		return
	}
	if e.outN == outSize {
		e.drain()
		if e.err != nil {
			return
		}
	}
	e.outBuf[e.outN] = c
	e.outN++
}

// drain hands the pending bytes to the writer.
func (e *Encoder) drain() {
	if e.outN == 0 {
		return
	}
	n, err := e.w.Write(e.outBuf[:e.outN])
	if n < e.outN && err == nil {
		err = io.ErrShortWrite
	}
	e.outN = 0 // written or lost; either way they are no longer pending
	if err != nil && e.err == nil {
		e.err = err
	}
}

// needEscValidUTF walks s once, reporting whether it needs
// any escaping and whether it is valid UTF-8.
func needEscValidUTF(s string) (bool, bool) {
	// Both checks share a pass because they partition the string:
	// only a byte below RuneSelf can need an escape, and only
	// a rune above it can be malformed.
	esc := false
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			if c == '"' || c == '\\' || c < 0x20 {
				esc = true
			}
			i++
			continue
		}
		// Above ASCII, only a whole well-formed rune advances the cursor by its
		// own width; a malformed one makes DecodeRuneInString eat a single byte.
		_, size := utf8.DecodeRuneInString(s[i:])
		if size <= 1 {
			return false, false
		}
		i += size
	}
	return esc, true
}
