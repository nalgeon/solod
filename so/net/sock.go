package net

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/net/netip"
	"solod.dev/so/strconv"
	"solod.dev/so/time"
)

// connState tells whether a connection or listener is usable. A socket cannot
// use its descriptor for this: descriptor 0 is standard input, so the zero
// value of a connection would look like a socket that is ready for I/O.
//
//so:promote
type connState uint8

const (
	stateUnopened connState = iota // the zero value: never opened
	stateOpen                      // ready for I/O
	stateClosed                    // closed by Close
)

// checkState returns an error if the socket state does not allow I/O.
// A closed socket gives ErrClosed. An unopened socket gives ErrInvalid.
func checkState(state connState) error {
	if state == stateOpen {
		return nil
	}
	if state == stateClosed {
		return ErrClosed
	}
	return ErrInvalid
}

// resolveOpts holds the inputs for resolveAddrs.
type resolveOpts struct {
	network  string
	proto    string
	socktype c.Int
	address  string
}

// resolveAddrs resolves address into endpoints for the given protocol (proto
// "tcp" or "udp", with the matching socktype c_SOCK_STREAM or c_SOCK_DGRAM),
// storing up to len(dst) of them as netip.AddrPort and returning how many were
// stored (at least one on success). An empty host or IP literal yields a single
// address; a host name may yield several, in the order they should be tried.
//
// It is the shared body behind [ResolveTCPAddr] and [ResolveUDPAddr]; the only
// per-protocol inputs are proto and socktype.
func resolveAddrs(opts resolveOpts, dst []netip.AddrPort) (int, error) {
	family := familyFor(opts.network, opts.proto)
	if family == afInvalid {
		return 0, ErrUnknownNetwork
	}
	hp, err := SplitHostPort(opts.address)
	if err != nil {
		return 0, err
	}
	port, ok := lookupPort(hp.Port, opts.proto)
	if !ok {
		return 0, ErrInvalidPort
	}

	// Empty host: the unspecified address for the family.
	if len(hp.Host) == 0 {
		dst[0] = netip.AddrPortFrom(unspecifiedIP(family), uint16(port))
		return 1, nil
	}

	// IP literal: no resolution needed, but it must match the network's family
	// ("tcp4"/"udp4" rejects an IPv6 literal and vice versa).
	if ip, perr := netip.ParseAddr(hp.Host); perr == nil {
		if !familyMatch(family, ip) {
			return 0, ErrNoSuitableAddr
		}
		dst[0] = netip.AddrPortFrom(ip, uint16(port))
		return 1, nil
	}

	// Host name: resolve via getaddrinfo and decode each result.
	hints := addrinfo{ai_family: family, ai_socktype: opts.socktype}
	var ai *addrinfo
	if getaddrinfo(hp.Host, nil, &hints, &ai) != 0 || ai == nil {
		return 0, ErrNoSuchHost
	}
	n := 0
	for p := ai; p != nil && n < len(dst); p = p.ai_next {
		var stor sockaddr_storage
		mem.Copy(&stor, p.ai_addr, int(p.ai_addrlen))
		ap := stor.addrPort()
		if !ap.Addr().IsValid() {
			continue
		}
		dst[n] = netip.AddrPortFrom(ap.Addr(), uint16(port))
		n++
	}
	freeaddrinfo(ai)
	if n == 0 {
		return 0, ErrNoSuchHost
	}
	return n, nil
}

// sockname returns the local address of fd, or the zero AddrPort on error.
func sockname(fd c.Int) netip.AddrPort {
	var stor sockaddr_storage
	slen := c.UInt(c.Sizeof[sockaddr_storage]())
	if getsockname(fd, stor.sockAddr(), &slen) != 0 {
		return netip.AddrPort{}
	}
	return stor.addrPort()
}

// waitFD blocks until fd is ready for the requested events (c_POLLIN for a
// read or accept, c_POLLOUT for a write) or the deadline passes, in which case
// it returns ErrTimeout. A zero deadline means no deadline: waitFD returns nil
// immediately and lets the following syscall block indefinitely.
func waitFD(fd c.Int, events c.Short, deadline time.Time) error {
	if deadline.IsZero() {
		return nil
	}
	for {
		d := time.Until(deadline)
		if d <= 0 {
			return ErrTimeout
		}
		// poll is used rather than SO_RCVTIMEO/SO_SNDTIMEO because, although those
		// timeouts bound read and write on every platform, accept() ignores them on
		// macOS/BSD, so a listener deadline there would never fire. poll bounds all
		// three uniformly.
		pfd := pollfd{fd: fd, events: events}
		n := poll(&pfd, 1, pollTimeout(d))
		if n > 0 {
			return nil
		}
		// n == 0 means this poll slice elapsed: loop to recheck the deadline
		// (it may not have arrived yet if pollTimeout clamped a long wait).
		// EINTR likewise retries. Any other error is reported.
		if n < 0 && errno != eINTR {
			return mapError()
		}
	}
}

// pollTimeout converts a positive duration to a poll timeout in milliseconds,
// rounding sub-millisecond values up to 1 (never 0, which would make poll
// return immediately) and clamping to the int range so a very distant deadline
// cannot overflow into a negative (block-forever) value.
func pollTimeout(d time.Duration) c.Int {
	ms := d.Milliseconds()
	if ms < 1 {
		return 1
	}
	const maxMS = 1<<31 - 1
	if ms > maxMS {
		return maxMS
	}
	return c.Int(ms)
}

// closeOnExec sets the close-on-exec flag on fd so the socket is not inherited
// by child processes spawned via exec. It is best-effort: errors are ignored,
// and there is a small window between creating the fd and this call in which a
// concurrent exec could still inherit it (the atomic SOCK_CLOEXEC/accept4 forms
// are Linux-only, so this portable fallback is used everywhere).
func closeOnExec(fd c.Int) {
	fcntl(fd, c_F_SETFD, c_FD_CLOEXEC)
}

// lookupPort resolves a port string to a numeric port. An empty string means
// port 0. Otherwise the string may be a decimal number or a service name from
// the services database (for example "http") looked up for the given protocol
// ("tcp" or "udp"). Returns the port and true on success, or false if the
// number is out of range or the service name is unknown.
func lookupPort(port, proto string) (int, bool) {
	if port == "" {
		return 0, true
	}
	n, err := strconv.Atoi(port)
	if err == nil {
		return n, n >= 0 && n <= 65535
	}
	se := getservbyname(port, proto)
	if se == nil {
		return 0, false
	}
	return int(ntohs(uint16(se.s_port))), true
}

// unspecifiedIP returns the unspecified address (0.0.0.0 or ::) for the family.
// It is used when the host part is empty.
func unspecifiedIP(family c.Int) netip.Addr {
	if family == c_AF_INET6 {
		return netip.IPv6Unspecified()
	}
	return netip.IPv4Unspecified()
}

// afInvalid is returned by familyFor for an unsupported network. Real address
// families are non-negative, so -1 is a safe sentinel.
const afInvalid = -1

// familyFor maps a network name to an address family, or afInvalid if the
// network is not a supported network for proto. For proto "tcp" the networks
// are "tcp", "tcp4", "tcp6"; for "udp" they are "udp", "udp4", "udp6". The
// bare proto means AF_UNSPEC, a "4" suffix AF_INET, a "6" suffix AF_INET6.
func familyFor(network, proto string) c.Int {
	if network == proto {
		return c_AF_UNSPEC
	}
	// Match proto followed by a single version digit, without allocating.
	if len(network) != len(proto)+1 {
		return afInvalid
	}
	for i := 0; i < len(proto); i++ {
		if network[i] != proto[i] {
			return afInvalid
		}
	}
	switch network[len(proto)] {
	case '4':
		return c_AF_INET
	case '6':
		return c_AF_INET6
	}
	return afInvalid
}

// familyMatch reports whether ip is usable on the given address family.
// AF_UNSPEC (the bare "tcp"/"udp" network) accepts any IP; AF_INET requires an
// IPv4 address and AF_INET6 an IPv6 one.
func familyMatch(family c.Int, ip netip.Addr) bool {
	if family == c_AF_INET {
		return ip.Is4()
	}
	if family == c_AF_INET6 {
		return ip.Is6()
	}
	return true
}
