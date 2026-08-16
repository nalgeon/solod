// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import "solod.dev/so/unicode/utf8"

// maxElem is the number of elements of the longest path of a test.
const maxElem = 32

// elemSep separates the elements of a Join case. No case holds the byte 0x01.
const elemSep byte = 1

// elemAt returns the element with the index i of a joined element string.
func elemAt(elems string, i int) string {
	start := 0
	for range i {
		for start < len(elems) && elems[start] != elemSep {
			start++
		}
		start++
	}
	if start > len(elems) {
		return ""
	}
	end := start
	for end < len(elems) && elems[end] != elemSep {
		end++
	}
	return elems[start:end]
}

// The reference implementations. Every sweep checks the package against the
// simplest code that gives the wanted result.

// cleanBrute returns the shortest path equal to path. It puts the elements
// that survive on a stack, so it shares no code with the package. The result
// is a view into dst, which must hold len(path) bytes.
func cleanBrute(dst []byte, path string) string {
	if path == "" {
		return "."
	}
	rooted := path[0] == '/'

	// starts and ends hold the bounds of the elements that survive.
	var starts, ends [maxElem]int
	n := 0
	for i := 0; i < len(path); {
		if path[i] == '/' {
			i++
			continue
		}
		start := i
		for i < len(path) && path[i] != '/' {
			i++
		}
		elem := path[start:i]
		if elem == "." {
			continue
		}
		if elem == ".." {
			// A parent element removes the element before it.
			// It cannot remove another parent element.
			if n > 0 && path[starts[n-1]:ends[n-1]] != ".." {
				n--
				continue
			}
			// A rooted path has no parent above the root.
			if rooted {
				continue
			}
		}
		starts[n] = start
		ends[n] = i
		n++
	}

	w := 0
	if rooted {
		dst[w] = '/'
		w++
	}
	for k := range n {
		if k > 0 {
			dst[w] = '/'
			w++
		}
		for i := starts[k]; i < ends[k]; i++ {
			dst[w] = path[i]
			w++
		}
	}
	if w == 0 {
		return "."
	}
	return string(dst[:w])
}

// baseBrute returns the last element of path. The result is a view into path.
func baseBrute(path string) string {
	if path == "" {
		return "."
	}
	end := len(path)
	for end > 0 && path[end-1] == '/' {
		end--
	}
	if end == 0 {
		return "/"
	}
	start := end
	for start > 0 && path[start-1] != '/' {
		start--
	}
	return path[start:end]
}

// dirBrute returns all but the last element of path. The result is a view
// into dst, which must hold len(path) bytes.
func dirBrute(dst []byte, path string) string {
	end := len(path)
	for end > 0 && path[end-1] != '/' {
		end--
	}
	return cleanBrute(dst, path[:end])
}

// extBrute returns the extension of path. The result is a view into path.
func extBrute(path string) string {
	start := len(path)
	for start > 0 && path[start-1] != '/' {
		start--
	}
	last := path[start:]
	for i := len(last) - 1; i >= 0; i-- {
		if last[i] == '.' {
			return last[i:]
		}
	}
	return ""
}

// joinBrute joins the elements with a slash and cleans the result. It writes
// the joined path into buf and returns a view into dst. Both buffers must
// hold the length of the elements plus the number of the elements.
func joinBrute(dst, buf []byte, elems []string) string {
	w := 0
	for _, e := range elems {
		if e == "" {
			continue
		}
		if w > 0 {
			buf[w] = '/'
			w++
		}
		for i := 0; i < len(e); i++ {
			buf[w] = e[i]
			w++
		}
	}
	if w == 0 {
		return ""
	}
	return cleanBrute(dst, string(buf[:w]))
}

// matchBrute reports whether name matches pattern. It tries every prefix of
// the name at a star, so it shares no code with the package. The pattern must
// hold no character class and no escape.
func matchBrute(pattern, name string) bool {
	if pattern == "" {
		return name == ""
	}
	switch pattern[0] {
	case '*':
		// A star matches any run of bytes without a slash.
		for i := 0; ; i++ {
			if matchBrute(pattern[1:], name[i:]) {
				return true
			}
			if i == len(name) || name[i] == '/' {
				return false
			}
		}
	case '?':
		// A question mark matches one code point, but not a slash.
		if name == "" || name[0] == '/' {
			return false
		}
		_, n := utf8.DecodeRuneInString(name)
		return matchBrute(pattern[1:], name[n:])
	}
	if name == "" || name[0] != pattern[0] {
		return false
	}
	return matchBrute(pattern[1:], name[1:])
}

// The word sweeps enumerate every word of an alphabet up to a length.

// The alphabet of the path sweep. The letters build every path shape: a name,
// a dot, and a separator. Two dots build a parent element.
const pathAlpha = "a./"

// maxPathWord is the length of the longest path of the sweep.
const maxPathWord = 7

// maxJoinWord is the length of the longest element of the join sweep.
const maxJoinWord = 3

// The alphabets of the match sweep.
const (
	patternAlpha = "ab*?/"
	nameAlpha    = "ab/"
)

// The lengths of the longest pattern and the longest name of the match sweep.
const (
	maxPattern = 4
	maxName    = 4
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
