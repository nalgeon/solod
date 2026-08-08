package net

import (
	"solod.dev/so/c"
	"solod.dev/so/net/netip"
	"solod.dev/so/time"
)

// UDPAddr represents the address of a UDP endpoint.
type UDPAddr struct {
	IP   netip.Addr
	Port int
}

// Network returns the address's network name, "udp".
func (UDPAddr) Network() string {
	return "udp"
}

// String returns the address in "host:port" form (or "[host]:port" for IPv6),
// built into buf. buf must have at least netip.MaxAddrPortLen bytes of
// capacity; the returned string aliases buf.
func (a UDPAddr) String(buf []byte) string {
	return a.addrPort().String(buf)
}

// addrPort packs the address into a netip.AddrPort.
func (a UDPAddr) addrPort() netip.AddrPort {
	return netip.AddrPortFrom(a.IP, uint16(a.Port))
}

// family returns the address family (AF_INET or AF_INET6)
// for the address's IP. Returns AF_INET for an invalid IP.
func (a UDPAddr) family() c.Int {
	if a.IP.Is6() {
		return c_AF_INET6
	}
	return c_AF_INET
}

// udpAddrOf builds a UDPAddr from a netip.AddrPort.
func udpAddrOf(ap netip.AddrPort) UDPAddr {
	return UDPAddr{IP: ap.Addr(), Port: int(ap.Port())}
}

// ResolveUDPAddr returns the address of a UDP endpoint.
//
// Known networks are "udp", "udp4" (IPv4-only), and "udp6" (IPv6-only).
//
// The address has the form "host:port". An empty host means the unspecified
// address (0.0.0.0 or ::). An IP literal host is parsed directly; otherwise it
// is resolved via the system resolver, and a host name with several addresses
// resolves to its first. The port may be a decimal number or a service name
// (for example "domain").
//
// Examples:
//
//	ResolveUDPAddr("udp", "golang.org:53")
//	ResolveUDPAddr("udp", "192.0.2.1:53")
//	ResolveUDPAddr("udp", "localhost:domain")
//	ResolveUDPAddr("udp", ":53")
func ResolveUDPAddr(network, address string) (UDPAddr, error) {
	var aps [1]netip.AddrPort
	opts := resolveOpts{
		network:  network,
		proto:    "udp",
		socktype: c_SOCK_DGRAM,
		address:  address,
	}
	if _, err := resolveAddrs(opts, aps[:]); err != nil {
		return UDPAddr{}, err
	}
	return udpAddrOf(aps[0]), nil
}

// UDPConn is a UDP network connection. It is used both for connected sockets
// (from [DialUDP], which fix a peer and support [UDPConn.Read]/[UDPConn.Write])
// and for unconnected sockets (from [ListenUDP], which exchange datagrams with
// arbitrary peers via [UDPConn.ReadFrom]/[UDPConn.WriteTo]).
//
// The zero value is not usable. A UDPConn must not be copied after use
// (copies share the underlying socket descriptor). The methods return
// [ErrInvalid] for a nil pointer or an unopened connection, and
// [ErrClosed] for a closed one.
type UDPConn struct {
	fd        c.Int
	laddr     UDPAddr
	raddr     UDPAddr // valid only when connected
	connected bool
	state     connState
	// Read/write deadlines; the zero Time means no deadline (block forever).
	rdeadline time.Time
	wdeadline time.Time
}

// UDPRead is the result of [UDPConn.ReadFrom]:
// the byte count and the source address.
type UDPRead struct {
	N    int
	Addr UDPAddr
}

// DialUDP creates a connected UDP socket bound to a fixed peer raddr, on the
// named UDP network.
//
// Known networks are "udp", "udp4" (IPv4-only), and "udp6" (IPv6-only).
// Use [ResolveUDPAddr] to obtain raddr (and an optional laddr) from a
// "host:port" string.
//
// A connected socket sends every datagram to raddr and only accepts datagrams
// from it; use [UDPConn.Read] and [UDPConn.Write]. If laddr is nil, a local
// address is automatically chosen. A laddr with an invalid IP binds only its
// port, on the unspecified address of the remote's family.
func DialUDP(network string, laddr, raddr *UDPAddr) (UDPConn, error) {
	if familyFor(network, "udp") == afInvalid {
		return UDPConn{}, ErrUnknownNetwork
	}
	if raddr == nil {
		return UDPConn{}, ErrAddrNotAvail
	}

	var rstor sockaddr_storage
	rlen := rstor.fill(raddr.addrPort())
	if rlen == 0 {
		return UDPConn{}, ErrAddrNotAvail
	}

	fd := socket(raddr.family(), c_SOCK_DGRAM, 0)
	if fd < 0 {
		return UDPConn{}, mapError()
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
			return UDPConn{}, err
		}
	}

	// For UDP, connect just records the peer address and returns immediately;
	// it sends no packets and cannot be interrupted, so there is no EINTR retry.
	if connect(fd, rstor.sockAddr(), rlen) != 0 {
		err := mapError()
		fd_close(fd)
		return UDPConn{}, err
	}

	conn := UDPConn{fd: fd, raddr: *raddr, connected: true, state: stateOpen}
	conn.laddr = udpAddrOf(sockname(fd))
	return conn, nil
}

// ListenUDP creates an unconnected UDP socket bound to the local address laddr.
//
// Known networks are "udp", "udp4" (IPv4-only), and "udp6" (IPv6-only).
// Use [ResolveUDPAddr] to obtain laddr from a "host:port" string.
//
// A nil laddr, or an unspecified IP in laddr (0.0.0.0 or ::, as produced by an
// empty host), binds all interfaces of the network's address family. A zero
// Port lets the system pick a free port, which the returned connection's
// [UDPConn.LocalAddr] reports. The socket is unconnected: exchange datagrams
// with arbitrary peers via [UDPConn.ReadFrom] and [UDPConn.WriteTo].
func ListenUDP(network string, laddr *UDPAddr) (UDPConn, error) {
	family := familyFor(network, "udp")
	if family == afInvalid {
		return UDPConn{}, ErrUnknownNetwork
	}

	// A nil laddr, or one with an invalid IP, binds the unspecified address
	// (all interfaces) for the network's family, keeping any requested port.
	addr := UDPAddr{IP: unspecifiedIP(family)}
	if laddr != nil {
		addr.Port = laddr.Port
		if laddr.IP.IsValid() {
			addr.IP = laddr.IP
		}
	}

	var stor sockaddr_storage
	slen := stor.fill(addr.addrPort())
	if slen == 0 {
		return UDPConn{}, ErrAddrNotAvail
	}

	fd := socket(addr.family(), c_SOCK_DGRAM, 0)
	if fd < 0 {
		return UDPConn{}, mapError()
	}
	closeOnExec(fd)

	// SO_REUSEADDR is intentionally not set here. Unlike TCP (no TIME_WAIT to
	// work around), on Linux it would let a second socket bind the same unicast
	// address and port, making datagram delivery between them unpredictable.
	if bind(fd, stor.sockAddr(), slen) != 0 {
		err := mapError()
		fd_close(fd)
		return UDPConn{}, err
	}

	// Report the bound address; with port 0 the system assigns the real port.
	return UDPConn{fd: fd, laddr: udpAddrOf(sockname(fd)), state: stateOpen}, nil
}

// Read reads a datagram from a connected connection into b.
//
// Read requires a connection from [DialUDP]; on an unconnected socket it
// returns [ErrAddrNotAvail] (use [UDPConn.ReadFrom] instead). A zero-length
// datagram is valid and returns (0, nil); Read never returns io.EOF.
func (conn *UDPConn) Read(b []byte) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	if !conn.connected {
		return 0, ErrAddrNotAvail
	}
	if len(b) == 0 {
		return 0, nil
	}
	// Restart on EINTR: a read interrupted by a signal before any data was
	// transferred returns -1/EINTR, and is retried transparently. Unlike TCP,
	// n == 0 is an empty datagram, not end of stream, so it is not an error.
	for {
		if err := waitFD(conn.fd, c_POLLIN, conn.rdeadline); err != nil {
			return 0, err
		}
		n := fd_read(conn.fd, &b[0], uintptr(len(b)))
		if n >= 0 {
			return n, nil
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// Write writes the datagram in b to a connected connection.
//
// Write requires a connection from [DialUDP]; on an unconnected socket it
// returns [ErrAddrNotAvail] (use [UDPConn.WriteTo] instead).
func (conn *UDPConn) Write(b []byte) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	if !conn.connected {
		return 0, ErrAddrNotAvail
	}
	// One datagram is one write. Restart on EINTR (interrupted before sending);
	// otherwise return whatever the single send reported.
	for {
		if err := waitFD(conn.fd, c_POLLOUT, conn.wdeadline); err != nil {
			return 0, err
		}
		var p *byte
		if len(b) > 0 {
			p = &b[0]
		}
		n := fd_write(conn.fd, p, uintptr(len(b)))
		if n >= 0 {
			return n, nil
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// ReadFrom reads a datagram from the connection into b and returns the byte
// count together with the source address. The buffer should be large enough
// to hold the datagram; any excess is discarded.
//
// ReadFrom requires an unconnected socket from [ListenUDP]; on a connected
// socket it returns [ErrAddrNotAvail] (use [UDPConn.Read] instead). A
// zero-length datagram is valid and reported as N == 0; ReadFrom never
// returns io.EOF.
func (conn *UDPConn) ReadFrom(b []byte) (UDPRead, error) {
	if err := conn.checkValid(); err != nil {
		return UDPRead{}, err
	}
	if conn.connected {
		return UDPRead{}, ErrAddrNotAvail
	}
	if len(b) == 0 {
		return UDPRead{}, nil
	}
	// Restart on EINTR. slen is in/out, so reset it each try.
	var stor sockaddr_storage
	for {
		if err := waitFD(conn.fd, c_POLLIN, conn.rdeadline); err != nil {
			return UDPRead{}, err
		}
		slen := c.UInt(c.Sizeof[sockaddr_storage]())
		n := recvfrom(conn.fd, &b[0], uintptr(len(b)), 0, stor.sockAddr(), &slen)
		if n >= 0 {
			return UDPRead{N: n, Addr: udpAddrOf(stor.addrPort())}, nil
		}
		if errno != eINTR {
			return UDPRead{}, mapError()
		}
	}
}

// WriteTo writes the datagram in b to addr.
//
// WriteTo requires an unconnected socket from [ListenUDP]; on a connected
// socket it returns [ErrAddrNotAvail] (use [UDPConn.Write] instead).
func (conn *UDPConn) WriteTo(b []byte, addr *UDPAddr) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	if conn.connected {
		return 0, ErrAddrNotAvail
	}
	if addr == nil {
		return 0, ErrAddrNotAvail
	}
	var stor sockaddr_storage
	slen := stor.fill(addr.addrPort())
	if slen == 0 {
		return 0, ErrAddrNotAvail
	}
	// One datagram is one send. Restart on EINTR (interrupted before sending).
	for {
		if err := waitFD(conn.fd, c_POLLOUT, conn.wdeadline); err != nil {
			return 0, err
		}
		var p *byte
		if len(b) > 0 {
			p = &b[0]
		}
		n := sendto(conn.fd, p, uintptr(len(b)), 0, stor.sockAddr(), slen)
		if n >= 0 {
			return n, nil
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// Close closes the connection. Returns an error
// if it has already been called.
func (conn *UDPConn) Close() error {
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
func (conn *UDPConn) LocalAddr() UDPAddr {
	return conn.laddr
}

// RemoteAddr returns the remote network address. It is meaningful only for a
// connected connection (from [DialUDP]); for an unconnected socket it is the
// zero UDPAddr.
func (conn *UDPConn) RemoteAddr() UDPAddr {
	return conn.raddr
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future I/O,
// not just the immediately following call.
// After a deadline has been exceeded, the connection can be
// refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to a read or write method
// will return [ErrTimeout].
//
// A zero value for t means I/O operations will not time out.
func (conn *UDPConn) SetDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	conn.wdeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future read calls.
// A zero value for t means reads will not time out.
func (conn *UDPConn) SetReadDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future write calls.
// A zero value for t means writes will not time out.
func (conn *UDPConn) SetWriteDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.wdeadline = t
	return nil
}

// checkValid returns an error if the connection is not usable for I/O.
func (conn *UDPConn) checkValid() error {
	if conn == nil {
		return ErrInvalid
	}
	return checkState(conn.state)
}
