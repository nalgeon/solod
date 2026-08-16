// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/runtime"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
	"solod.dev/so/unicode/utf8"
)

// smallMaxTokenSize is much smaller than the default maximum token size.
// A small size makes the tests faster.
const smallMaxTokenSize = 256

var scanTests = []string{
	"",
	"a",
	"¼",
	"☹",
	"\x81",   // UTF-8 error
	"\uFFFD", // correctly encoded RuneError
	"abcdefgh",
	"abc def\n\t\tgh    ",
	"abc¼☹\x81\uFFFD日本語\x82abc",
}

func TestScanByte(t *testing.T) {
	alloc := t.Allocator()
	for n := range scanTests {
		test := scanTests[n]
		buf := strings.NewReader(test)
		s := bufio.NewScanner(alloc, &buf)
		s.Split(bufio.ScanBytes)
		i := 0
		for ; s.Scan(); i++ {
			b := s.Bytes()
			if len(b) != 1 || b[0] != test[i] {
				t.Errorf("#%d: %d: got %s, want %c", n, i, string(b), test[i])
			}
		}
		if i != len(test) {
			t.Errorf("#%d: termination expected at %d; got %d", n, len(test), i)
		}
		if err := s.Err(); err != nil {
			t.Errorf("#%d: Err() = %s, want nil", n, errName(errCode(err)))
		}
		s.Free()
	}
}

func TestScanRune(t *testing.T) {
	// Check that the rune splitter gives the same runes as a
	// range loop over the string.
	alloc := t.Allocator()
	for n := range scanTests {
		test := scanTests[n]
		buf := strings.NewReader(test)
		s := bufio.NewScanner(alloc, &buf)
		s.Split(bufio.ScanRunes)
		runeCount := 0
		// The range loop over the string gives the expected runes.
		for i, expect := range test {
			if !s.Scan() {
				break
			}
			runeCount++
			got, _ := utf8.DecodeRune(s.Bytes())
			if got != expect {
				t.Errorf("#%d: %d: got %c, want %c", n, i, got, expect)
			}
		}
		if s.Scan() {
			t.Errorf("#%d: scan ran too long, got %s", n, s.Text())
		}
		testRuneCount := utf8.RuneCountInString(test)
		if runeCount != testRuneCount {
			t.Errorf("#%d: termination expected at %d; got %d", n, testRuneCount, runeCount)
		}
		if err := s.Err(); err != nil {
			t.Errorf("#%d: Err() = %s, want nil", n, errName(errCode(err)))
		}
		s.Free()
	}
}

var wordScanTests = []string{
	"",
	" ",
	"\n",
	"a",
	" a ",
	"abc def",
	" abc def ",
	" abc\tdef\nghi\rjkl\fmno\vpqr\u0085stu\u00a0\n",
}

func TestScanWords(t *testing.T) {
	// Check that the word splitter gives the same data as strings.Fields.
	alloc := t.Allocator()
	for n := range wordScanTests {
		test := wordScanTests[n]
		buf := strings.NewReader(test)
		s := bufio.NewScanner(alloc, &buf)
		s.Split(bufio.ScanWords)
		words := strings.Fields(alloc, test)
		wordCount := 0
		for ; wordCount < len(words); wordCount++ {
			if !s.Scan() {
				break
			}
			got := s.Text()
			if got != words[wordCount] {
				t.Errorf("#%d: %d: got %s, want %s", n, wordCount, got, words[wordCount])
			}
		}
		if s.Scan() {
			t.Errorf("#%d: scan ran too long, got %s", n, s.Text())
		}
		if wordCount != len(words) {
			t.Errorf("#%d: termination expected at %d; got %d", n, len(words), wordCount)
		}
		if err := s.Err(); err != nil {
			t.Errorf("#%d: Err() = %s, want nil", n, errName(errCode(err)))
		}
		mem.FreeSlice(alloc, words)
		s.Free()
	}
}

// slowReader gives only a few bytes at a time. It tests the incremental
// reads of Scanner.Scan.
type slowReader struct {
	max int
	buf io.Reader
}

func (sr *slowReader) Read(p []byte) (int, error) {
	if len(p) > sr.max {
		p = p[0:sr.max]
	}
	return sr.buf.Read(p)
}

// genLine writes to buf a predictable but non-trivial line of text of
// length n, with the final newline and an occasional carriage return.
// If addNewline is false, it writes no \r and no \n.
func genLine(buf *bytes.Buffer, lineNum, n int, addNewline bool) {
	buf.Reset()
	doCR := lineNum%5 == 0
	if doCR {
		n--
	}
	for i := 0; i < n-1; i++ { // stop early for \n
		c := 'a' + byte(lineNum+i)
		if c == '\n' || c == '\r' { // do not confuse the test
			c = 'N'
		}
		buf.WriteByte(c)
	}
	if addNewline {
		if doCR {
			buf.WriteByte('\r')
		}
		buf.WriteByte('\n')
	}
}

func TestScanLongLines(t *testing.T) {
	// Check that the line splitter handles some carriage returns and no long lines.
	// The line lengths grow to smallMaxTokenSize and then fall back to 0,
	// so the whole text is a bit more than smallMaxTokenSize squared.
	const textSize = 68 * 1024
	alloc := t.Allocator()

	var tmpArr [smallMaxTokenSize + 8]byte
	tmp := bytes.NewBuffer(mem.NoAlloc, tmpArr[:0])
	buf := bytes.NewBuffer(alloc, mem.AllocSlice[byte](alloc, 0, textSize))

	lineNum := 0
	j := 0
	for range 2 * smallMaxTokenSize {
		genLine(&tmp, lineNum, j, true)
		if j < smallMaxTokenSize {
			j++
		} else {
			j--
		}
		buf.Write(tmp.Bytes())
		lineNum++
	}

	sr := slowReader{1, &buf}
	scanBuf := mem.AllocSlice[byte](alloc, smallMaxTokenSize, smallMaxTokenSize)
	s := bufio.NewScanner(alloc, &sr)
	s.Split(bufio.ScanLines)
	s.Buffer(scanBuf, smallMaxTokenSize)
	j = 0
	for lineNum := 0; s.Scan(); lineNum++ {
		genLine(&tmp, lineNum, j, false)
		if j < smallMaxTokenSize {
			j++
		} else {
			j--
		}
		line := tmp.String() // the string token, for variety
		if s.Text() != line {
			t.Errorf("%d: bad line: %d %d", lineNum, len(s.Bytes()), len(line))
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}

	s.Free()
	mem.FreeSlice(alloc, scanBuf)
	buf.Free()
}

func TestScanLineTooLong(t *testing.T) {
	// Check that the line splitter fails on a long line.
	// The line lengths grow to 2*smallMaxTokenSize, so the whole text is
	// about four times smallMaxTokenSize squared.
	const textSize = 132 * 1024
	alloc := t.Allocator()

	var tmpArr [2*smallMaxTokenSize + 8]byte
	tmp := bytes.NewBuffer(mem.NoAlloc, tmpArr[:0])
	buf := bytes.NewBuffer(alloc, mem.AllocSlice[byte](alloc, 0, textSize))

	lineNum := 0
	j := 0
	for range 2 * smallMaxTokenSize {
		genLine(&tmp, lineNum, j, true)
		j++
		buf.Write(tmp.Bytes())
		lineNum++
	}

	sr := slowReader{3, &buf}
	scanBuf := mem.AllocSlice[byte](alloc, smallMaxTokenSize, smallMaxTokenSize)
	s := bufio.NewScanner(alloc, &sr)
	s.Split(bufio.ScanLines)
	s.Buffer(scanBuf, smallMaxTokenSize)
	j = 0
	for lineNum := 0; s.Scan(); lineNum++ {
		genLine(&tmp, lineNum, j, false)
		if j < smallMaxTokenSize {
			j++
		} else {
			j--
		}
		line := tmp.Bytes()
		if !bytes.Equal(s.Bytes(), line) {
			t.Errorf("%d: bad line: %d %d", lineNum, len(s.Bytes()), len(line))
		}
	}
	if err := s.Err(); errCode(err) != errTooLong {
		t.Errorf("Err() = %s, want ErrTooLong", errName(errCode(err)))
	}

	s.Free()
	mem.FreeSlice(alloc, scanBuf)
	buf.Free()
}

// testNoNewline checks that the line splitter handles a final line
// without a newline.
func testNoNewline(t *testing.T, text string, lines []string) {
	buf := strings.NewReader(text)
	sr := slowReader{7, &buf}
	s := bufio.NewScanner(t.Allocator(), &sr)
	s.Split(bufio.ScanLines)
	for lineNum := 0; s.Scan(); lineNum++ {
		line := lines[lineNum]
		if s.Text() != line {
			t.Errorf("%d: bad line: got %s, want %s", lineNum, s.Text(), line)
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}
	s.Free()
}

func TestScanLineNoNewline(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz"
	lines := []string{
		"abcdefghijklmn",
		"opqrstuvwxyz",
	}
	testNoNewline(t, text, lines)
}

func TestScanLineReturnButNoNewline(t *testing.T) {
	// Check that the line splitter handles a final line
	// with a carriage return and no newline.
	const text = "abcdefghijklmn\nopqrstuvwxyz\r"
	lines := []string{
		"abcdefghijklmn",
		"opqrstuvwxyz",
	}
	testNoNewline(t, text, lines)
}

func TestScanLineEmptyFinalLine(t *testing.T) {
	// Check that the line splitter handles a final empty line.
	const text = "abcdefghijklmn\nopqrstuvwxyz\n\n"
	lines := []string{
		"abcdefghijklmn",
		"opqrstuvwxyz",
		"",
	}
	testNoNewline(t, text, lines)
}

func TestScanLineEmptyFinalLineWithCR(t *testing.T) {
	// Check that the line splitter handles a final empty line
	// with a carriage return and no newline.
	const text = "abcdefghijklmn\nopqrstuvwxyz\n\r"
	lines := []string{
		"abcdefghijklmn",
		"opqrstuvwxyz",
		"",
	}
	testNoNewline(t, text, lines)
}

// errorSplitCount counts the tokens of errorSplit.
var errorSplitCount int

// errorSplitAtEOF records a call of errorSplit at the end of the input.
// The test expects the error before the end of the input.
var errorSplitAtEOF bool

// errorSplitOK is the number of the tokens before the error.
const errorSplitOK = 7

// errorSplit gives a little data, then a predictable error.
func errorSplit(data []byte, atEOF bool) bufio.SplitResult {
	if atEOF {
		errorSplitAtEOF = true
		return bufio.SplitResult{Err: errSplitFailed}
	}
	if errorSplitCount >= errorSplitOK {
		return bufio.SplitResult{Err: errSplitFailed}
	}
	errorSplitCount++
	return bufio.SplitResult{Advance: 1, Token: data[0:1], HasToken: true}
}

func TestSplitError(t *testing.T) {
	errorSplitCount = 0
	errorSplitAtEOF = false

	const text = "abcdefghijklmnopqrstuvwxyz"
	buf := strings.NewReader(text)
	sr := slowReader{1, &buf}
	s := bufio.NewScanner(t.Allocator(), &sr)
	s.Split(errorSplit)
	i := 0
	for ; s.Scan(); i++ {
		if len(s.Bytes()) != 1 || text[i] != s.Bytes()[0] {
			t.Errorf("#%d: got %s, want %c", i, s.Text(), text[i])
		}
	}

	if errorSplitAtEOF {
		t.Error("split function did not get enough data")
	}
	if i != errorSplitOK {
		t.Errorf("unexpected termination; expected %d tokens got %d", errorSplitOK, i)
	}
	if err := s.Err(); errCode(err) != errSplit {
		t.Errorf("Err() = %s, want split error", errName(errCode(err)))
	}
	s.Free()
}

// errAtEOFSplit fails on the last token, after the scanner holds io.EOF.
func errAtEOFSplit(data []byte, atEOF bool) bufio.SplitResult {
	res := bufio.ScanWords(data, atEOF)
	if res.HasToken && len(res.Token) > 1 {
		res.Err = errSplitFailed
	}
	return res
}

func TestErrAtEOF(t *testing.T) {
	// Check that a split function error replaces an EOF.
	sr := strings.NewReader("1 2 33")
	s := bufio.NewScanner(t.Allocator(), &sr)
	s.Split(errAtEOFSplit)
	for s.Scan() {
	}
	if err := s.Err(); errCode(err) != errSplit {
		t.Errorf("Err() = %s, want split error", errName(errCode(err)))
	}
	s.Free()
}

// unexpectedEOFReader fails every read with io.ErrUnexpectedEOF.
type unexpectedEOFReader struct{}

func (*unexpectedEOFReader) Read(p []byte) (int, error) {
	_ = p
	return 0, io.ErrUnexpectedEOF
}

func TestNonEOFWithEmptyRead(t *testing.T) {
	// Check that the scanner keeps a read error that
	// is not an EOF (Go issue 5268).
	var r unexpectedEOFReader
	s := bufio.NewScanner(t.Allocator(), &r)
	for s.Scan() {
		t.Error("read should fail")
		break
	}
	if err := s.Err(); errCode(err) != errUnexpectedEOF {
		t.Errorf("Err() = %s, want ErrUnexpectedEOF", errName(errCode(err)))
	}
	s.Free()
}

// TestBadReader checks that Scan finishes with endless empty reads.
func TestBadReader(t *testing.T) {
	var r zeroReader
	s := bufio.NewScanner(t.Allocator(), &r)
	for s.Scan() {
		t.Error("read should fail")
		break
	}
	if err := s.Err(); errCode(err) != errNoProgress {
		t.Errorf("Err() = %s, want ErrNoProgress", errName(errCode(err)))
	}
	s.Free()
}

func TestScanWordsExcessiveWhiteSpace(t *testing.T) {
	// Check that the word splitter skips more
	// white space than the buffer holds.
	const word = "ipsum"
	alloc := t.Allocator()

	const spaceCount = 4 * smallMaxTokenSize
	text := mem.AllocSlice[byte](alloc, spaceCount+len(word), spaceCount+len(word))
	for i := range spaceCount {
		text[i] = ' '
	}
	copy(text[spaceCount:], []byte(word))

	sr := strings.NewReader(string(text))
	scanBuf := mem.AllocSlice[byte](alloc, smallMaxTokenSize, smallMaxTokenSize)
	s := bufio.NewScanner(alloc, &sr)
	s.Buffer(scanBuf, smallMaxTokenSize)
	s.Split(bufio.ScanWords)
	if !s.Scan() {
		t.Errorf("scan failed: %s", errName(errCode(s.Err())))
	} else if s.Text() != word {
		t.Errorf("unexpected token: %s", s.Text())
	}

	s.Free()
	mem.FreeSlice(alloc, scanBuf)
	mem.FreeSlice(alloc, text)
}

// commaSplit splits the input at every comma. The last token is empty if
// the input ends with a comma.
func commaSplit(data []byte, atEOF bool) bufio.SplitResult {
	_ = atEOF
	for i := range data {
		if data[i] == ',' {
			return bufio.SplitResult{Advance: i + 1, Token: data[:i], HasToken: true}
		}
	}
	return bufio.SplitResult{Token: data, HasToken: true, Err: bufio.ErrFinalToken}
}

// testEmptyTokens checks that the scanner finds the empty tokens,
// at the end of a line and at the end of the input (Go issue 8672).
func testEmptyTokens(t *testing.T, text string, values []string) {
	sr := strings.NewReader(text)
	s := bufio.NewScanner(t.Allocator(), &sr)
	s.Split(commaSplit)
	i := 0
	for ; s.Scan(); i++ {
		if i >= len(values) {
			t.Errorf("got %d fields, expected %d", i+1, len(values))
			break
		}
		if s.Text() != values[i] {
			t.Errorf("%d: got %s, want %s", i, s.Text(), values[i])
		}
	}
	if i != len(values) {
		t.Errorf("got %d fields, expected %d", i, len(values))
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}
	s.Free()
}

func TestEmptyTokens(t *testing.T) {
	testEmptyTokens(t, "1,2,3,", []string{"1", "2", "3", ""})
}

func TestWithNoEmptyTokens(t *testing.T) {
	testEmptyTokens(t, "1,2,3", []string{"1", "2", "3"})
}

func TestBlankLines(t *testing.T) {
	alloc := t.Allocator()
	text := strings.Repeat(alloc, "\n", 1000)
	sr := strings.NewReader(text)
	s := bufio.NewScanner(alloc, &sr)
	for count := 0; s.Scan(); count++ {
		if count > 2000 {
			t.Error("looping")
			break
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}
	s.Free()
	mem.FreeString(alloc, text)
}

// countdown is the number of the tokens that countdownSplit still gives.
var countdown int

// countdownSplit gives one byte at a time until the countdown ends.
func countdownSplit(data []byte, atEOF bool) bufio.SplitResult {
	_ = atEOF
	if countdown > 0 {
		countdown--
		return bufio.SplitResult{Advance: 1, Token: data[:1], HasToken: true}
	}
	return bufio.SplitResult{}
}

func TestEmptyLinesOK(t *testing.T) {
	// Check that the check for a loop at EOF accepts the empty tokens.
	const count = 10000
	alloc := t.Allocator()
	countdown = count

	text := strings.Repeat(alloc, "\n", count)
	sr := strings.NewReader(text)
	s := bufio.NewScanner(alloc, &sr)
	s.Split(countdownSplit)
	for s.Scan() {
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}
	if countdown != 0 {
		t.Errorf("stopped with %d left to process", countdown)
	}
	s.Free()
	mem.FreeString(alloc, text)
}

func TestHugeBuffer(t *testing.T) {
	// Check that the scanner reads a huge token with a big enough buffer.
	if !runtime.Hosted {
		t.Skip("the token needs more memory than the freestanding heap holds")
		return
	}
	const size = 2 * bufio.MaxScanTokenSize
	alloc := t.Allocator()

	text := mem.AllocSlice[byte](alloc, size+1, size+1)
	for i := range size {
		text[i] = 'x'
	}
	text[size] = '\n'
	want := string(text[:size])

	sr := strings.NewReader(string(text))
	s := bufio.NewScanner(alloc, &sr)
	var start [100]byte
	s.Buffer(start[:], 3*bufio.MaxScanTokenSize)
	for s.Scan() {
		if s.Text() != want {
			t.Errorf("scan got incorrect token of length %d", len(s.Text()))
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Err() = %s, want nil", errName(errCode(err)))
	}

	s.Free()
	mem.FreeSlice(alloc, text)
}

// negativeEOFReader gives an invalid -1 at the end, like a wrapper of the
// read system call.
type negativeEOFReader struct {
	n int
}

func (r *negativeEOFReader) Read(p []byte) (int, error) {
	if r.n > 0 {
		c := min(r.n, len(p))
		for i := 0; i < c; i++ {
			p[i] = 'a'
		}
		p[c-1] = '\n'
		r.n -= c
		return c, nil
	}
	return -1, io.EOF
}

func TestNegativeEOFReader(t *testing.T) {
	// Check that the scanner gives ErrBadReadCount on a reader that
	// gives a negative count of the read bytes (Go issue 38053).
	r := negativeEOFReader{n: 10}
	s := bufio.NewScanner(t.Allocator(), &r)
	c := 0
	for s.Scan() {
		c++
		if c > 1 {
			t.Error("read too many lines")
			break
		}
	}
	if err := s.Err(); errCode(err) != errBadReadCount {
		t.Errorf("Err() = %s, want ErrBadReadCount", errName(errCode(err)))
	}
	s.Free()
}

// largeReader gives an invalid count that is larger than the number of the
// requested bytes.
type largeReader struct{}

func (*largeReader) Read(p []byte) (int, error) {
	return len(p) + 1, nil
}

func TestLargeReader(t *testing.T) {
	// Check that the scanner gives ErrBadReadCount on a reader that gives
	// an impossibly large count of the read bytes (Go issue 38053).
	var r largeReader
	s := bufio.NewScanner(t.Allocator(), &r)
	for s.Scan() {
	}
	if err := s.Err(); errCode(err) != errBadReadCount {
		t.Errorf("Err() = %s, want ErrBadReadCount", errName(errCode(err)))
	}
	s.Free()
}
