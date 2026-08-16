package slog_test

import (
	"solod.dev/so/log/slog"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// The alphabet and the longest word of the sweeps. The alphabet holds the
// three characters that force a quote, a plain letter, and the two edge bytes.
const (
	sweepAlpha   = " =\"a\x00\xff"
	maxSweepWord = 4
)

// wordCount returns the number of words of alpha with the length n.
func wordCount(alpha string, n int) int {
	count := 1
	for range n {
		count *= len(alpha)
	}
	return count
}

// wordTotal returns the number of words of alpha with a length up to max.
func wordTotal(alpha string, max int) int {
	total := 0
	for n := 0; n <= max; n++ {
		total += wordCount(alpha, n)
	}
	return total
}

// wordAt writes the word number i of alpha into buf and returns the result.
// The shorter words come first, and every word with a length up to max appears
// once. The caller must keep i below wordTotal(alpha, max).
func wordAt(buf []byte, alpha string, max, i int) string {
	for n := 0; n <= max; n++ {
		count := wordCount(alpha, n)
		if i < count {
			for k := 0; k < n; k++ {
				buf[k] = alpha[i%len(alpha)]
				i /= len(alpha)
			}
			return string(buf[:n])
		}
		i -= count
	}
	return ""
}

// quoteBrute writes the text of a string value into b the simple way. The
// handler quotes an empty value, and a value with a space, a quote or an
// equals sign.
func quoteBrute(b *strings.Builder, s string) {
	quote := len(s) == 0
	for i := range len(s) {
		if s[i] == ' ' || s[i] == '"' || s[i] == '=' {
			quote = true
		}
	}
	if quote {
		b.WriteString("\"")
		b.WriteString(s)
		b.WriteString("\"")
		return
	}
	b.WriteString(s)
}

// formatBrute writes the text of a record with one string attr into buf the
// simple way. The result is a view into buf.
func formatBrute(buf []byte, msg, key, val string) string {
	b := strings.FixedBuilder(buf)
	b.WriteString(refText)
	b.WriteString(" INFO ")
	b.WriteString(msg)
	b.WriteString(" ")
	b.WriteString(key)
	b.WriteString("=")
	quoteBrute(&b, val)
	b.WriteString("\n")
	return b.String()
}

func TestSweepValue(t *testing.T) {
	var word [maxSweepWord]byte
	var buf, wbuf [bufLen]byte
	var attrs [1]slog.Attr

	total := wordTotal(sweepAlpha, maxSweepWord)
	for i := range total {
		val := wordAt(word[:], sweepAlpha, maxSweepWord, i)
		attrs[0] = slog.String("k", val)
		got := handle(buf[:], record("m", attrs[:]))
		want := formatBrute(wbuf[:], "m", "k", val)
		if got != want {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, want)
			return
		}
	}
}

func TestSweepKeyMessage(t *testing.T) {
	var word [maxSweepWord]byte
	var buf, wbuf [bufLen]byte
	var attrs [1]slog.Attr

	total := wordTotal(sweepAlpha, maxSweepWord)
	for i := range total {
		w := wordAt(word[:], sweepAlpha, maxSweepWord, i)
		attrs[0] = slog.String(w, "v")
		got := handle(buf[:], record(w, attrs[:]))
		want := formatBrute(wbuf[:], w, w, "v")
		if got != want {
			t.Errorf("#%d: Handle() wrote %s, want %s", i, got, want)
			return
		}
	}
}
