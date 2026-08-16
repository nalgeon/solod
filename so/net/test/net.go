package net_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/net"
	"solod.dev/so/testing"
)

// The error a bad address text must give.
// A test table names one of these codes.
const (
	errNone = iota
	errAddrNotAvail
	errInvalidPort
	errMissingBracket
	errMissingPort
	errNoSuitableAddr
	errTooManyColons
	errUnexpectedBracket
	errUnknownNetwork
	errOther // an error this file does not name
)

// errCode returns the code of err.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	if err == net.ErrAddrNotAvail {
		return errAddrNotAvail
	}
	if err == net.ErrInvalidPort {
		return errInvalidPort
	}
	if err == net.ErrMissingBracket {
		return errMissingBracket
	}
	if err == net.ErrMissingPort {
		return errMissingPort
	}
	if err == net.ErrNoSuitableAddr {
		return errNoSuitableAddr
	}
	if err == net.ErrTooManyColons {
		return errTooManyColons
	}
	if err == net.ErrUnexpectedBracket {
		return errUnexpectedBracket
	}
	if err == net.ErrUnknownNetwork {
		return errUnknownNetwork
	}
	return errOther
}

// errName returns the name of an error code.
func errName(code int) string {
	switch code {
	case errNone:
		return "nil"
	case errAddrNotAvail:
		return "ErrAddrNotAvail"
	case errInvalidPort:
		return "ErrInvalidPort"
	case errMissingBracket:
		return "ErrMissingBracket"
	case errMissingPort:
		return "ErrMissingPort"
	case errNoSuitableAddr:
		return "ErrNoSuitableAddr"
	case errTooManyColons:
		return "ErrTooManyColons"
	case errUnexpectedBracket:
		return "ErrUnexpectedBracket"
	case errUnknownNetwork:
		return "ErrUnknownNetwork"
	}
	return "another error"
}

// A splitCase is one SplitHostPort case. A case that must fail names the error
// code and leaves host and port empty, because a failed split must report both
// parts as empty.
type splitCase struct {
	in   string // the address text
	host string // the host part
	port string // the port part
	err  int    // the expected error code
}

var splitCases = []splitCase{
	// Host name.
	{in: "localhost:http", host: "localhost", port: "http"},
	{in: "localhost:80", host: "localhost", port: "80"},

	// Host name with a zone identifier.
	{in: "localhost%lo0:http", host: "localhost%lo0", port: "http"},
	{in: "localhost%lo0:80", host: "localhost%lo0", port: "80"},
	{in: "[localhost%lo0]:http", host: "localhost%lo0", port: "http"},
	{in: "[localhost%lo0]:80", host: "localhost%lo0", port: "80"},

	// IP literal.
	{in: "127.0.0.1:http", host: "127.0.0.1", port: "http"},
	{in: "127.0.0.1:80", host: "127.0.0.1", port: "80"},
	{in: "[::1]:http", host: "::1", port: "http"},
	{in: "[::1]:80", host: "::1", port: "80"},

	// IP literal with a zone identifier.
	{in: "[::1%lo0]:http", host: "::1%lo0", port: "http"},
	{in: "[::1%lo0]:80", host: "::1%lo0", port: "80"},

	// Wildcard for the host name.
	{in: ":http", host: "", port: "http"},
	{in: ":80", host: "", port: "80"},

	// Wildcard for the service name or the port number.
	{in: "golang.org:", host: "golang.org", port: ""},
	{in: "127.0.0.1:", host: "127.0.0.1", port: ""},
	{in: "[::1]:", host: "::1", port: ""},

	// Empty host and empty port.
	{in: ":", host: "", port: ""},
	{in: "[]:", host: "", port: ""},

	// Opaque service name.
	{in: "golang.org:https%foo", host: "golang.org", port: "https%foo"},

	// A missing port.
	{in: "", err: errMissingPort},
	{in: "golang.org", err: errMissingPort},
	{in: "127.0.0.1", err: errMissingPort},
	{in: "[::1]", err: errMissingPort},
	{in: "[fe80::1%lo0]", err: errMissingPort},
	{in: "[localhost%lo0]", err: errMissingPort},
	{in: "localhost%lo0", err: errMissingPort},
	{in: "[foo:bar]", err: errMissingPort},
	{in: "[foo:bar]baz", err: errMissingPort},
	{in: "[foo]bar:baz", err: errMissingPort},

	// Too many colons.
	{in: "::1", err: errTooManyColons},
	{in: "fe80::1%lo0", err: errTooManyColons},
	{in: "fe80::1%lo0:80", err: errTooManyColons},
	{in: "[foo]:[bar]:baz", err: errTooManyColons},

	// An unexpected bracket.
	{in: "[foo]:[bar]baz", err: errUnexpectedBracket},
	{in: "foo[bar]:baz", err: errUnexpectedBracket},
	{in: "foo]bar:baz", err: errUnexpectedBracket},

	// A missing closing bracket.
	{in: "[foo:80", err: errMissingBracket},
	{in: "[:80", err: errMissingBracket},
}

func TestSplitHostPort(t *testing.T) {
	for _, tt := range splitCases {
		hp, err := net.SplitHostPort(tt.in)
		if errCode(err) != tt.err {
			t.Errorf("SplitHostPort(%s) error = %s, want %s",
				tt.in, errName(errCode(err)), errName(tt.err))
			continue
		}
		if hp.Host != tt.host || hp.Port != tt.port {
			t.Errorf("SplitHostPort(%s) = %s, %s; want %s, %s",
				tt.in, hp.Host, hp.Port, tt.host, tt.port)
		}
	}
}

// A resolveCase is one ResolveTCPAddr or ResolveUDPAddr case.
// The tables are in tcp.go and udp.go.
type resolveCase struct {
	network string // the network name
	address string // the address text
	err     int    // the expected error code
	port    int    // the expected port, for a case that must pass
}

// A joinCase is one JoinHostPort case.
type joinCase struct {
	host string
	port string
	want string
}

var joinCases = []joinCase{
	// Host name.
	{host: "localhost", port: "http", want: "localhost:http"},
	{host: "localhost", port: "80", want: "localhost:80"},

	// Host name with a zone identifier.
	{host: "localhost%lo0", port: "http", want: "localhost%lo0:http"},
	{host: "localhost%lo0", port: "80", want: "localhost%lo0:80"},

	// IP literal.
	{host: "127.0.0.1", port: "http", want: "127.0.0.1:http"},
	{host: "127.0.0.1", port: "80", want: "127.0.0.1:80"},
	{host: "::1", port: "http", want: "[::1]:http"},
	{host: "::1", port: "80", want: "[::1]:80"},

	// IP literal with a zone identifier.
	{host: "::1%lo0", port: "http", want: "[::1%lo0]:http"},
	{host: "::1%lo0", port: "80", want: "[::1%lo0]:80"},

	// Wildcard for the host name.
	{host: "", port: "http", want: ":http"},
	{host: "", port: "80", want: ":80"},

	// Wildcard for the service name or the port number.
	{host: "golang.org", port: "", want: "golang.org:"},
	{host: "127.0.0.1", port: "", want: "127.0.0.1:"},
	{host: "::1", port: "", want: "[::1]:"},

	// Both parts empty.
	{host: "", port: "", want: ":"},
	{host: ":", port: "", want: "[:]:"},

	// Opaque service name.
	{host: "golang.org", port: "https%foo", want: "golang.org:https%foo"},
}

func TestJoinHostPort(t *testing.T) {
	for _, tt := range joinCases {
		var buf [64]byte
		if got := net.JoinHostPort(buf[:], tt.host, tt.port); got != tt.want {
			t.Errorf("JoinHostPort(%s, %s) = %s; want %s", tt.host, tt.port, got, tt.want)
		}
	}
}

func TestJoinHostPortExactBuffer(t *testing.T) {
	for _, tt := range joinCases {
		n := len(tt.host) + len(tt.port) + 4
		buf := mem.AllocSlice[byte](t.Allocator(), n, n)
		if got := net.JoinHostPort(buf, tt.host, tt.port); got != tt.want {
			t.Errorf("JoinHostPort(%s, %s) = %s; want %s", tt.host, tt.port, got, tt.want)
		}
		mem.FreeSlice(t.Allocator(), buf)
	}
}

// idx returns the index of the first b in s, or -1.
func idx(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// hasAny reports whether s holds any byte of set.
func hasAny(s, set string) bool {
	for i := 0; i < len(s); i++ {
		if idx(set, s[i]) >= 0 {
			return true
		}
	}
	return false
}

// A splitWant is the result wantSplit expects of SplitHostPort.
type splitWant struct {
	host string
	port string
	ok   bool
}

// wantSplit states the grammar SplitHostPort accepts, without following how
// the function is written. A valid text has one of two shapes:
//
//	host ":" port      - the text holds no bracket and one colon
//	"[" host "]:" port - host holds no bracket, port holds no colon
//
// wantSplit reports whether s has one of the two shapes, and gives the parts.
func wantSplit(s string) splitWant {
	if len(s) > 0 && s[0] == '[' {
		end := idx(s, ']')
		if end < 0 || end+1 >= len(s) || s[end+1] != ':' {
			return splitWant{}
		}
		host, port := s[1:end], s[end+2:]
		if hasAny(host, "[]") || hasAny(port, ":[]") {
			return splitWant{}
		}
		return splitWant{host: host, port: port, ok: true}
	}
	i := idx(s, ':')
	if i < 0 || hasAny(s, "[]") || idx(s[i+1:], ':') >= 0 {
		return splitWant{}
	}
	return splitWant{host: s[:i], port: s[i+1:], ok: true}
}

// The bytes a swept address text is built from, and the longest swept text.
// The alphabet holds every byte SplitHostPort gives a meaning to, plus a
// plain letter.
const (
	sweepAlphabet = "[]:%a"
	sweepMaxLen   = 6
)

func TestSplitHostPortSweep(t *testing.T) {
	// Runs SplitHostPort over every text up to sweepMaxLen bytes of sweepAlphabet,
	// and compares the result against wantSplit. Every text the function accepts
	// must also survive a JoinHostPort round trip.
	var buf [sweepMaxLen]byte
	var joined [2*sweepMaxLen + 4]byte
	nok, nbad := 0, 0
	for n := 0; n <= sweepMaxLen; n++ {
		total := 1
		for range n {
			total *= len(sweepAlphabet)
		}
		for code := 0; code < total; code++ {
			v := code
			for i := 0; i < n; i++ {
				buf[i] = sweepAlphabet[v%len(sweepAlphabet)]
				v /= len(sweepAlphabet)
			}
			s := string(buf[:n])
			want := wantSplit(s)

			hp, err := net.SplitHostPort(s)
			if want.ok != (err == nil) {
				t.Errorf("SplitHostPort(%s) error = %s, want ok = %t",
					s, errName(errCode(err)), want.ok)
				continue
			}
			if err != nil {
				nbad++
				if errCode(err) == errOther {
					t.Errorf("SplitHostPort(%s) gives an unnamed error", s)
				}
				if hp.Host != "" || hp.Port != "" {
					t.Errorf("SplitHostPort(%s) = %s, %s on failure; want two empty parts",
						s, hp.Host, hp.Port)
				}
				continue
			}
			nok++
			if hp.Host != want.host || hp.Port != want.port {
				t.Errorf("SplitHostPort(%s) = %s, %s; want %s, %s",
					s, hp.Host, hp.Port, want.host, want.port)
				continue
			}

			// The two parts must join back into a text that splits the same way.
			j := net.JoinHostPort(joined[:], hp.Host, hp.Port)
			back, err := net.SplitHostPort(j)
			if err != nil {
				t.Errorf("SplitHostPort(%s) after join of %s: %s", j, s, errName(errCode(err)))
				continue
			}
			if back.Host != hp.Host || back.Port != hp.Port {
				t.Errorf("join and split of %s = %s, %s; want %s, %s",
					s, back.Host, back.Port, hp.Host, hp.Port)
			}
		}
	}
	// The sweep must reach both answers. A reference that always says no
	// would otherwise pass without checking a single valid text.
	if nok == 0 || nbad == 0 {
		t.Errorf("sweep accepted %d texts and rejected %d; want both above zero", nok, nbad)
	}
}
