package net

import (
	"solod.dev/so/c"
	"solod.dev/so/io"
	"solod.dev/so/time"
)

// UnixAddr represents the address of a Unix domain socket endpoint.
type UnixAddr struct {
	Name string // filesystem path of the socket
	Net  string // "unix" or "unixgram"
}

// Network returns the address's network name, "unix" or "unixgram".
// The zero UnixAddr reports "unix".
func (a UnixAddr) Network() string {
	if a.Net == "" {
		return "unix"
	}
	return a.Net
}

// String returns the socket path.
func (a UnixAddr) String() string {
	return a.Name
}

// ResolveUnixAddr returns the address of a Unix domain socket endpoint.
//
// Known networks are "unix" (stream) and "unixgram" (datagram). The address is
// a filesystem path; it is not resolved or validated against the filesystem,
// only carried through, so there is no DNS or service lookup as with TCP/UDP.
//
// Examples:
//
//	ResolveUnixAddr("unix", "/tmp/echo.sock")
//	ResolveUnixAddr("unixgram", "/tmp/dgram.sock")
func ResolveUnixAddr(network, address string) (UnixAddr, error) {
	if unixSocktype(network) == afInvalid {
		return UnixAddr{}, ErrUnknownNetwork
	}
	return UnixAddr{Name: address, Net: network}, nil
}

// UnixConn is a Unix domain socket connection. Like [UDPConn] it serves several
// roles: a connected socket (from [DialUnix], with [UnixConn.Read]/[UnixConn.Write])
// and, for datagrams, an unconnected socket (from [ListenUnixgram], exchanging
// datagrams with arbitrary peers via [UnixConn.ReadFrom]/[UnixConn.WriteTo]). An
// accepted stream connection (from [UnixListener.Accept]) is also a UnixConn.
//
// The zero value is not usable. A UnixConn must not be copied after use
// (copies share the underlying socket descriptor and the source-address buffer).
// The methods return [ErrInvalid] for a nil pointer or an unopened connection,
// and [ErrClosed] for a closed one.
type UnixConn struct {
	fd        c.Int
	laddr     UnixAddr
	raddr     UnixAddr // valid only when connected
	connected bool
	stream    bool // true for "unix" (stream), false for "unixgram" (datagram)
	state     connState
	path      string // bound socket file to unlink on Close; "" if none
	// rnamebuf backs the source path reported by ReadFrom; see that method.
	rnamebuf  [maxUnixPath]byte
	rdeadline time.Time
	wdeadline time.Time
}

// UnixRead is the result of [UnixConn.ReadFrom]:
// the byte count and the source address.
type UnixRead struct {
	N    int
	Addr UnixAddr
}

// DialUnix connects to raddr on the named Unix network.
//
// Known networks are "unix" (stream) and "unixgram" (datagram). Use
// [ResolveUnixAddr] to obtain raddr (and an optional laddr) from a path.
//
// If laddr is non-nil it is bound first; the bound socket file is removed on
// [UnixConn.Close]. For an unnamed local end pass a nil laddr. The returned
// connection is connected: use [UnixConn.Read] and [UnixConn.Write].
func DialUnix(network string, laddr, raddr *UnixAddr) (UnixConn, error) {
	socktype := unixSocktype(network)
	if socktype == afInvalid {
		return UnixConn{}, ErrUnknownNetwork
	}
	if raddr == nil {
		return UnixConn{}, ErrAddrNotAvail
	}

	var rstor sockaddr_storage
	rlen := rstor.fillUnix(raddr.Name)
	if rlen == 0 {
		return UnixConn{}, ErrAddrNotAvail
	}

	fd := socket(c_AF_UNIX, socktype, 0)
	if fd < 0 {
		return UnixConn{}, mapError()
	}
	closeOnExec(fd)

	// Optional local bind address (bind-before-connect).
	var path string
	if laddr != nil {
		var lstor sockaddr_storage
		llen := lstor.fillUnix(laddr.Name)
		if llen == 0 || bind(fd, lstor.sockAddr(), llen) != 0 {
			err := mapError()
			fd_close(fd)
			return UnixConn{}, err
		}
		path = laddr.Name
	}

	if connect(fd, rstor.sockAddr(), rlen) != 0 {
		err := mapError()
		fd_close(fd)
		return UnixConn{}, err
	}

	conn := UnixConn{fd: fd, connected: true, stream: network == "unix", path: path, state: stateOpen}
	conn.raddr = UnixAddr{Name: raddr.Name, Net: network}
	conn.laddr = UnixAddr{Net: network}
	if laddr != nil {
		conn.laddr.Name = laddr.Name
	}
	return conn, nil
}

// ListenUnixgram creates an unconnected Unix datagram socket bound to laddr.
//
// The only known network is "unixgram". laddr must be non-nil and name a path;
// that socket file is removed on [UnixConn.Close]. The socket is unconnected:
// exchange datagrams with arbitrary peers via [UnixConn.ReadFrom] and
// [UnixConn.WriteTo]. A peer must itself be bound (also via ListenUnixgram) to
// be addressable as a reply destination.
func ListenUnixgram(network string, laddr *UnixAddr) (UnixConn, error) {
	if network != "unixgram" {
		return UnixConn{}, ErrUnknownNetwork
	}
	if laddr == nil {
		return UnixConn{}, ErrAddrNotAvail
	}

	var stor sockaddr_storage
	slen := stor.fillUnix(laddr.Name)
	if slen == 0 {
		return UnixConn{}, ErrAddrNotAvail
	}

	fd := socket(c_AF_UNIX, c_SOCK_DGRAM, 0)
	if fd < 0 {
		return UnixConn{}, mapError()
	}
	closeOnExec(fd)

	if bind(fd, stor.sockAddr(), slen) != 0 {
		err := mapError()
		fd_close(fd)
		return UnixConn{}, err
	}

	conn := UnixConn{fd: fd, path: laddr.Name, state: stateOpen}
	conn.laddr = UnixAddr{Name: laddr.Name, Net: "unixgram"}
	return conn, nil
}

// Read reads data from a connected connection into b.
//
// Read requires a connection from [DialUnix] or [UnixListener.Accept]; on an
// unconnected socket it returns [ErrAddrNotAvail] (use [UnixConn.ReadFrom]).
// For a stream connection, Read returns 0, io.EOF at end of stream; for a
// connected datagram socket a zero-length datagram is valid and returns (0, nil).
func (conn *UnixConn) Read(b []byte) (int, error) {
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
			// For a stream this is end of stream; for a datagram it is an
			// empty datagram, which is valid and not an error.
			if conn.stream {
				return 0, io.EOF
			}
			return 0, nil
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// Write writes data to a connected connection.
//
// Write requires a connection from [DialUnix] or [UnixListener.Accept]; on an
// unconnected socket it returns [ErrAddrNotAvail] (use [UnixConn.WriteTo]).
// A stream connection writes all of b, looping as the send buffer drains; a
// connected datagram socket sends b as a single datagram.
func (conn *UnixConn) Write(b []byte) (int, error) {
	if err := conn.checkValid(); err != nil {
		return 0, err
	}
	if !conn.connected {
		return 0, ErrAddrNotAvail
	}
	if conn.stream {
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
	// Datagram: one write is one datagram. Restart on EINTR (interrupted before
	// sending); otherwise return whatever the single send reported.
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
// count together with the source address. The buffer should be large enough to
// hold the datagram; any excess is discarded.
//
// ReadFrom requires an unconnected socket from [ListenUnixgram]; on a connected
// socket it returns [ErrAddrNotAvail] (use [UnixConn.Read] instead). A
// zero-length datagram is valid and reported as N == 0.
//
// The returned Addr.Name is a view into a per-connection buffer and stays valid
// only until the next ReadFrom on this connection. An anonymous (unbound) peer
// has an empty Name.
func (conn *UnixConn) ReadFrom(b []byte) (UnixRead, error) {
	if err := conn.checkValid(); err != nil {
		return UnixRead{}, err
	}
	if conn.connected {
		return UnixRead{}, ErrAddrNotAvail
	}
	if len(b) == 0 {
		return UnixRead{}, nil
	}
	// Restart on EINTR. slen is in/out, so reset it each try.
	var stor sockaddr_storage
	for {
		if err := waitFD(conn.fd, c_POLLIN, conn.rdeadline); err != nil {
			return UnixRead{}, err
		}
		slen := c.UInt(c.Sizeof[sockaddr_storage]())
		n := recvfrom(conn.fd, &b[0], uintptr(len(b)), 0, stor.sockAddr(), &slen)
		if n >= 0 {
			// stor.unixName aliases the local stor; copy the path into the
			// connection-owned buffer so the returned view outlives this call.
			m := copy(conn.rnamebuf[:], stor.unixName())
			addr := UnixAddr{Name: string(conn.rnamebuf[:m]), Net: "unixgram"}
			return UnixRead{N: n, Addr: addr}, nil
		}
		if errno != eINTR {
			return UnixRead{}, mapError()
		}
	}
}

// WriteTo writes the datagram in b to addr.
//
// WriteTo requires an unconnected socket from [ListenUnixgram]; on a connected
// socket it returns [ErrAddrNotAvail] (use [UnixConn.Write] instead).
func (conn *UnixConn) WriteTo(b []byte, addr *UnixAddr) (int, error) {
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
	slen := stor.fillUnix(addr.Name)
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

// Close closes the connection, removing the bound socket file if this
// connection created one (from [ListenUnixgram] or a [DialUnix] with a local
// address). Returns an error if it has already been called. An unlink failure
// after a successful close is not surfaced; the close error takes precedence.
func (conn *UnixConn) Close() error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.state = stateClosed
	var err error
	if fd_close(conn.fd) != 0 {
		err = mapError()
	}
	if conn.path != "" {
		unlink(conn.path)
	}
	return err
}

// LocalAddr returns the local network address.
func (conn *UnixConn) LocalAddr() UnixAddr {
	return conn.laddr
}

// RemoteAddr returns the remote network address. It is meaningful only for a
// connected connection (from [DialUnix] or [UnixListener.Accept]); for an
// unconnected socket it is the zero UnixAddr.
func (conn *UnixConn) RemoteAddr() UnixAddr {
	return conn.raddr
}

// SetDeadline sets the read and write deadlines associated with the connection.
// It is equivalent to calling both SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations fail with
// [ErrTimeout] instead of blocking. A zero value for t means I/O operations
// will not time out.
func (conn *UnixConn) SetDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	conn.wdeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future read calls.
// A zero value for t means reads will not time out.
func (conn *UnixConn) SetReadDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.rdeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future write calls.
// A zero value for t means writes will not time out.
func (conn *UnixConn) SetWriteDeadline(t time.Time) error {
	if err := conn.checkValid(); err != nil {
		return err
	}
	conn.wdeadline = t
	return nil
}

// checkValid returns an error if the connection is not usable for I/O.
func (conn *UnixConn) checkValid() error {
	if conn == nil {
		return ErrInvalid
	}
	return checkState(conn.state)
}

// UnixListener is a Unix domain stream listener. The zero value
// is not usable; obtain one from [ListenUnix]. The methods return
// [ErrInvalid] for a nil pointer or an unopened listener, and
// [ErrClosed] for a closed one.
type UnixListener struct {
	fd    c.Int
	addr  UnixAddr
	path  string // bound socket file to unlink on Close
	state connState
	// Accept deadline; the zero Time means no deadline (block forever).
	deadline time.Time
}

// ListenUnix announces on the local Unix address laddr.
//
// The only known network is "unix" (stream). laddr must be non-nil and name a
// path; that socket file is created on bind and removed on [UnixListener.Close].
func ListenUnix(network string, laddr *UnixAddr) (UnixListener, error) {
	if network != "unix" {
		return UnixListener{}, ErrUnknownNetwork
	}
	if laddr == nil {
		return UnixListener{}, ErrAddrNotAvail
	}

	var stor sockaddr_storage
	slen := stor.fillUnix(laddr.Name)
	if slen == 0 {
		return UnixListener{}, ErrAddrNotAvail
	}

	fd := socket(c_AF_UNIX, c_SOCK_STREAM, 0)
	if fd < 0 {
		return UnixListener{}, mapError()
	}
	closeOnExec(fd)

	// No SO_REUSEADDR: a stale socket file is removed by unlinking on Close, and
	// binding over an existing path fails with EADDRINUSE, the desired signal
	// that something is already listening there.
	if bind(fd, stor.sockAddr(), slen) != 0 || listen(fd, listenBacklog) != 0 {
		err := mapError()
		fd_close(fd)
		return UnixListener{}, err
	}

	addr := UnixAddr{Name: laddr.Name, Net: "unix"}
	return UnixListener{fd: fd, addr: addr, path: laddr.Name, state: stateOpen}, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *UnixListener) Accept() (UnixConn, error) {
	if err := l.checkValid(); err != nil {
		return UnixConn{}, err
	}
	// Restart on EINTR: an accept interrupted by a signal returns -1/EINTR,
	// and is retried transparently. slen is in/out, so reset it each try.
	var stor sockaddr_storage
	for {
		if err := waitFD(l.fd, c_POLLIN, l.deadline); err != nil {
			return UnixConn{}, err
		}
		slen := c.UInt(c.Sizeof[sockaddr_storage]())
		fd := accept(l.fd, stor.sockAddr(), &slen)
		if fd >= 0 {
			closeOnExec(fd)
			// The accepted socket has no bound socket file of its own (so it does
			// not unlink on Close), and the peer is typically anonymous.
			conn := UnixConn{fd: fd, connected: true, stream: true, state: stateOpen}
			conn.laddr = UnixAddr{Name: l.addr.Name, Net: "unix"}
			conn.raddr = UnixAddr{Net: "unix"}
			return conn, nil
		}
		if errno != eINTR {
			return UnixConn{}, mapError()
		}
	}
}

// Close stops listening and removes the bound socket file.
// Already accepted connections are not closed. An unlink failure after a
// successful close is not surfaced; the close error takes precedence.
func (l *UnixListener) Close() error {
	if err := l.checkValid(); err != nil {
		return err
	}
	l.state = stateClosed
	var err error
	if fd_close(l.fd) != 0 {
		err = mapError()
	}
	if l.path != "" {
		unlink(l.path)
	}
	return err
}

// Addr returns the listener's network address.
func (l *UnixListener) Addr() UnixAddr {
	return l.addr
}

// SetDeadline sets the deadline for future Accept calls. An Accept that has no
// connection ready before t fails with [ErrTimeout]. The zero value for t
// clears the deadline (Accept blocks until a connection arrives).
func (l *UnixListener) SetDeadline(t time.Time) error {
	if err := l.checkValid(); err != nil {
		return err
	}
	l.deadline = t
	return nil
}

// checkValid returns an error if the listener is not usable.
func (l *UnixListener) checkValid() error {
	if l == nil {
		return ErrInvalid
	}
	return checkState(l.state)
}

// unixSocktype maps a Unix network name to a socket type (c_SOCK_STREAM for
// "unix", c_SOCK_DGRAM for "unixgram"), or afInvalid for any other network.
// Note the two socktype constants are both 0 in the Go stubs and only become
// distinct after transpiling; the network string, not this value, is what
// decides stream vs datagram behavior on the host.
func unixSocktype(network string) c.Int {
	if network == "unix" {
		return c_SOCK_STREAM
	}
	if network == "unixgram" {
		return c_SOCK_DGRAM
	}
	return afInvalid
}
