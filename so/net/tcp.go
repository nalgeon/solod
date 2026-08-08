package net

import (
	"solod.dev/so/c"
	"solod.dev/so/io"
	"solod.dev/so/net/netip"
	"solod.dev/so/time"
)

// listenBacklog is the default connection backlog for listeners.
const listenBacklog = 128

// TCPAddr represents the address of a TCP endpoint.
type TCPAddr struct {
	IP   netip.Addr
	Port int
}

// Network returns the address's network name, "tcp".
func (TCPAddr) Network() string {
	return "tcp"
}

// String returns the address in "host:port" form (or "[host]:port" for IPv6),
// built into buf. buf must have at least netip.MaxAddrPortLen bytes of
// capacity; the returned string aliases buf.
func (a TCPAddr) String(buf []byte) string {
	return a.addrPort().String(buf)
}

// addrPort packs the address into a netip.AddrPort.
func (a TCPAddr) addrPort() netip.AddrPort {
	return netip.AddrPortFrom(a.IP, uint16(a.Port))
}

// family returns the address family (AF_INET or AF_INET6)
// for the address's IP. Returns AF_INET for an invalid IP.
func (a TCPAddr) family() c.Int {
	if a.IP.Is6() {
		return c_AF_INET6
	}
	return c_AF_INET
}

// ResolveTCPAddr returns the address of a TCP endpoint.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
//
// The address has the form "host:port". An empty host means the unspecified
// address (0.0.0.0 or ::). An IP literal host is parsed directly; otherwise it
// is resolved via the system resolver, and a host name with several addresses
// resolves to its first. The port may be a decimal number or a service name
// (for example "http").
//
// Examples:
//
//	ResolveTCPAddr("tcp", "golang.org:443")
//	ResolveTCPAddr("tcp", "192.0.2.1:80")
//	ResolveTCPAddr("tcp", "localhost:http")
//	ResolveTCPAddr("tcp", ":80")
func ResolveTCPAddr(network, address string) (TCPAddr, error) {
	var aps [1]netip.AddrPort
	opts := resolveOpts{
		network:  network,
		proto:    "tcp",
		socktype: c_SOCK_STREAM,
		address:  address,
	}
	if _, err := resolveAddrs(opts, aps[:]); err != nil {
		return TCPAddr{}, err
	}
	return tcpAddrOf(aps[0]), nil
}

// TCPConn abstracts a TCP network connection.
//
// The zero value is not usable; obtain a TCPConn from [DialTCP] or
// [TCPListener.Accept]. A TCPConn must not be copied after use
// (copies share the underlying socket descriptor). The methods return
// [ErrInvalid] for a nil pointer or an unopened connection, and
// [ErrClosed] for a closed one.
type TCPConn struct {
	fd    c.Int
	laddr TCPAddr
	raddr TCPAddr
	state connState
	// Read/write deadlines; the zero Time means no deadline (block forever).
	rdeadline time.Time
	wdeadline time.Time
}

// DialTCP connects to raddr on the named TCP network.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
// Use [ResolveTCPAddr] to obtain raddr (and an optional laddr) from a
// "host:port" string.
//
// If laddr is nil, a local address is automatically chosen.
// A laddr with an invalid IP binds only its port, on the
// unspecified address of the remote's family.
func DialTCP(network string, laddr, raddr *TCPAddr) (TCPConn, error) {
	if familyFor(network, "tcp") == afInvalid {
		return TCPConn{}, ErrUnknownNetwork
	}
	if raddr == nil {
		return TCPConn{}, ErrAddrNotAvail
	}

	var rstor sockaddr_storage
	rlen := rstor.fill(raddr.addrPort())
	if rlen == 0 {
		return TCPConn{}, ErrAddrNotAvail
	}

	fd := socket(raddr.family(), c_SOCK_STREAM, 0)
	if fd < 0 {
		return TCPConn{}, mapError()
	}
	closeOnExec(fd)

	// Optional local bind address (bind-before-connect).
	if laddr != nil {
		local := *laddr
		if !local.IP.IsValid() {
			// Bind only the port, on the unspecified address of the remote's family.
			local.IP = unspecifiedIP(raddr.family())
		}
		var lstor sockaddr_storage
		llen := lstor.fill(local.addrPort())
		if llen == 0 || bind(fd, lstor.sockAddr(), llen) != 0 {
			err := mapError()
			fd_close(fd)
			return TCPConn{}, err
		}
	}

	// Unlike read/write/accept, connect is not restarted on EINTR: a blocking
	// connect interrupted by a signal keeps completing in the background, so
	// re-calling it would return EALREADY/EISCONN. Doing it right needs
	// poll(POLLOUT) plus getsockopt(SO_ERROR); not worth it for this rare case.
	if connect(fd, rstor.sockAddr(), rlen) != 0 {
		err := mapError()
		fd_close(fd)
		return TCPConn{}, err
	}

	conn := TCPConn{fd: fd, raddr: *raddr, state: stateOpen}
	conn.laddr = tcpAddrOf(sockname(fd))
	return conn, nil
}

// Read reads data from the connection into b.
// At end of stream, Read returns 0, io.EOF.
func (conn *TCPConn) Read(b []byte) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	// Restart on EINTR: a read interrupted by a signal before any data was
	// transferred returns -1/EINTR, and is retried transparently.
	for {
		if err := waitFD(conn.fd, c_POLLIN, conn.rdeadline); err != nil {
			return 0, err
		}
		n := fd_read(conn.fd, &b[0], uintptr(len(b)))
		if n > 0 {
			return n, nil
		}
		if n == 0 {
			return 0, io.EOF
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// Write writes len(b) bytes from b to the connection.
// Returns the number of bytes written and an error, if any.
func (conn *TCPConn) Write(b []byte) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	// Loop until all bytes are written: a single write may transfer fewer
	// bytes than requested when the socket send buffer fills up. A write
	// interrupted by a signal before any data was transferred returns
	// -1/EINTR and is restarted.
	total := 0
	for total < len(b) {
		if err := waitFD(conn.fd, c_POLLOUT, conn.wdeadline); err != nil {
			return total, err
		}
		n := fd_write(conn.fd, &b[total], uintptr(len(b)-total))
		if n < 0 {
			if errno == eINTR {
				continue
			}
			return total, mapError()
		}
		total += n
	}
	return total, nil
}

// Close closes the connection. Returns an error
// if it has already been called.
func (conn *TCPConn) Close() error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.state = stateClosed
	if fd_close(conn.fd) != 0 {
		return mapError()
	}
	return nil
}

// LocalAddr returns the local network address.
func (conn *TCPConn) LocalAddr() TCPAddr {
	return conn.laddr
}

// RemoteAddr returns the remote network address.
func (conn *TCPConn) RemoteAddr() TCPAddr {
	return conn.raddr
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future I/O,
// not just the immediately following call to Read or Write.
// After a deadline has been exceeded, the connection can be
// refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return [ErrTimeout].
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (conn *TCPConn) SetDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	conn.wdeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (conn *TCPConn) SetReadDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// A zero value for t means Write will not time out.
func (conn *TCPConn) SetWriteDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.wdeadline = t
	return nil
}

// checkValid returns an error if the connection is not usable for I/O.
func (conn *TCPConn) checkValid() error {
	if conn == nil {
		return ErrInvalid
	}
	return checkState(conn.state)
}

// TCPListener is a TCP network listener. The zero value
// is not usable; obtain one from [ListenTCP]. The methods return
// [ErrInvalid] for a nil pointer or an unopened listener, and
// [ErrClosed] for a closed one.
type TCPListener struct {
	fd    c.Int
	addr  TCPAddr
	state connState
	// Accept deadline; the zero Time means no deadline (block forever).
	deadline time.Time
}

// ListenTCP announces on the local TCP address laddr.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
// Use [ResolveTCPAddr] to obtain laddr from a "host:port" string.
//
// A nil laddr, or an unspecified IP in laddr (0.0.0.0 or ::, as produced by an
// empty host), listens on all interfaces of the network's address family:
// 0.0.0.0 (all IPv4 interfaces) for "tcp" or "tcp4", :: (all IPv6 interfaces)
// for "tcp6". A zero Port lets the system pick a free port, which the returned
// listener's [TCPListener.Addr] reports.
func ListenTCP(network string, laddr *TCPAddr) (TCPListener, error) {
	family := familyFor(network, "tcp")
	if family == afInvalid {
		return TCPListener{}, ErrUnknownNetwork
	}

	// A nil laddr, or one with an invalid IP, binds the unspecified address
	// (all interfaces) for the network's family, keeping any requested port.
	addr := TCPAddr{IP: unspecifiedIP(family)}
	if laddr != nil {
		addr.Port = laddr.Port
		if laddr.IP.IsValid() {
			addr.IP = laddr.IP
		}
	}

	var stor sockaddr_storage
	slen := stor.fill(addr.addrPort())
	if slen == 0 {
		return TCPListener{}, ErrAddrNotAvail
	}

	fd := socket(addr.family(), c_SOCK_STREAM, 0)
	if fd < 0 {
		return TCPListener{}, mapError()
	}
	closeOnExec(fd)

	one := c.Int(1)
	setsockopt(fd, c_SOL_SOCKET, c_SO_REUSEADDR, &one, c.UInt(c.Sizeof[c.Int]()))
	if bind(fd, stor.sockAddr(), slen) != 0 || listen(fd, listenBacklog) != 0 {
		err := mapError()
		fd_close(fd)
		return TCPListener{}, err
	}

	// Report the bound address; with port 0 the system assigns the real port.
	return TCPListener{fd: fd, addr: tcpAddrOf(sockname(fd)), state: stateOpen}, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *TCPListener) Accept() (TCPConn, error) {
	if err := l.checkValid(); err != nil {
		return TCPConn{}, err
	}
	// Restart on EINTR: an accept interrupted by a signal returns -1/EINTR,
	// and is retried transparently. slen is in/out, so reset it each try.
	var stor sockaddr_storage
	for {
		if err := waitFD(l.fd, c_POLLIN, l.deadline); err != nil {
			return TCPConn{}, err
		}
		slen := c.UInt(c.Sizeof[sockaddr_storage]())
		fd := accept(l.fd, stor.sockAddr(), &slen)
		if fd >= 0 {
			closeOnExec(fd)
			// Report the accepted socket's real local address; for a wildcard
			// listener this is the concrete interface, not l.addr's 0.0.0.0/::.
			conn := TCPConn{fd: fd, laddr: tcpAddrOf(sockname(fd)), state: stateOpen}
			conn.raddr = tcpAddrOf(stor.addrPort())
			return conn, nil
		}
		if errno != eINTR {
			return TCPConn{}, mapError()
		}
	}
}

// Close stops listening on the TCP address.
// Already accepted connections are not closed.
func (l *TCPListener) Close() error {
	if err := l.checkValid(); err != nil {
		return err
	}
	l.state = stateClosed
	if fd_close(l.fd) != 0 {
		return mapError()
	}
	return nil
}

// Addr returns the listener's network address.
func (l *TCPListener) Addr() TCPAddr {
	return l.addr
}

// SetDeadline sets the deadline for future Accept calls. An Accept that has no
// connection ready before t fails with [ErrTimeout]. The zero value for t
// clears the deadline (Accept blocks until a connection arrives).
func (l *TCPListener) SetDeadline(t time.Time) error {
	if err := l.checkValid(); err != nil {
		return err
	}
	l.deadline = t
	return nil
}

// tcpAddrOf builds a TCPAddr from a netip.AddrPort.
func tcpAddrOf(ap netip.AddrPort) TCPAddr {
	return TCPAddr{IP: ap.Addr(), Port: int(ap.Port())}
}

// checkValid returns an error if the listener is not usable.
func (l *TCPListener) checkValid() error {
	if l == nil {
		return ErrInvalid
	}
	return checkState(l.state)
}
