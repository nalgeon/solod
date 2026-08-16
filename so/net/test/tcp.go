package net_test

import (
	"solod.dev/so/io"
	"solod.dev/so/net"
	"solod.dev/so/net/netip"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// The ResolveTCPAddr cases that need no name resolution: an IP literal host or
// an empty host, with a numeric port or a service name.
var tcpResolveCases = []resolveCase{
	// An IP literal needs no DNS. The bare network takes both families.
	{network: "tcp", address: "127.0.0.1:80", port: 80},
	{network: "tcp4", address: "127.0.0.1:80", port: 80},
	{network: "tcp", address: "[::1]:80", port: 80},
	{network: "tcp6", address: "[::1]:80", port: 80},

	// A named port resolves via the services database.
	{network: "tcp", address: "127.0.0.1:http", port: 80},

	// An empty host gives the unspecified address of the family.
	{network: "tcp", address: ":80", port: 80},

	// The port limits. An empty port means port 0.
	{network: "tcp", address: "127.0.0.1:", port: 0},
	{network: "tcp", address: "127.0.0.1:0", port: 0},
	{network: "tcp", address: "127.0.0.1:65535", port: 65535},

	// An unknown network.
	{network: "udp", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "tcp5", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "tcpx", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "tcp44", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "TCP", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "tc", address: "127.0.0.1:80", err: errUnknownNetwork},
	{network: "", address: "127.0.0.1:80", err: errUnknownNetwork},

	// A bad address text. The error comes from SplitHostPort.
	{network: "tcp", address: "127.0.0.1", err: errMissingPort},
	{network: "tcp", address: "::1", err: errTooManyColons},

	// A bad port.
	{network: "tcp", address: "127.0.0.1:65536", err: errInvalidPort},
	{network: "tcp", address: "127.0.0.1:99999", err: errInvalidPort},
	{network: "tcp", address: "127.0.0.1:-1", err: errInvalidPort},
	{network: "tcp", address: "127.0.0.1:nosuchservice", err: errInvalidPort},

	// An IP literal must match the family of the network.
	{network: "tcp4", address: "[::1]:80", err: errNoSuitableAddr},
	{network: "tcp6", address: "127.0.0.1:80", err: errNoSuitableAddr},
}

func TestTCP_Resolve(t *testing.T) {
	for _, tt := range tcpResolveCases {
		addr, err := net.ResolveTCPAddr(tt.network, tt.address)
		if errCode(err) != tt.err {
			t.Errorf("ResolveTCPAddr(%s, %s) error = %s, want %s",
				tt.network, tt.address, errName(errCode(err)), errName(tt.err))
			continue
		}
		if err != nil {
			if addr.IP.IsValid() || addr.Port != 0 {
				t.Errorf("ResolveTCPAddr(%s, %s) gives an address on failure",
					tt.network, tt.address)
			}
			continue
		}
		if addr.Port != tt.port {
			t.Errorf("ResolveTCPAddr(%s, %s) port = %d, want %d",
				tt.network, tt.address, addr.Port, tt.port)
		}
	}
}

func TestTCP_ResolveLiteral(t *testing.T) {
	// An IP literal is parsed directly, without the resolver.
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !addr.IP.Is4() || !addr.IP.Equal(netip.MustParseAddr("127.0.0.1")) {
		t.Error("unexpected IPv4 literal address")
	}

	addr, err = net.ResolveTCPAddr("tcp", "[::1]:80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !addr.IP.Is6() || !addr.IP.Equal(netip.MustParseAddr("::1")) {
		t.Error("unexpected IPv6 literal address")
	}
}

func TestTCP_ResolveEmptyHost(t *testing.T) {
	// An empty host gives the unspecified address of the network's family.
	addr, err := net.ResolveTCPAddr("tcp", ":80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !addr.IP.IsUnspecified() || !addr.IP.Is4() {
		t.Error("tcp with an empty host should give 0.0.0.0")
	}

	addr, err = net.ResolveTCPAddr("tcp6", ":80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !addr.IP.IsUnspecified() || !addr.IP.Is6() {
		t.Error("tcp6 with an empty host should give ::")
	}
}

func TestTCP_AddrText(t *testing.T) {
	var buf [netip.MaxAddrPortLen]byte

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if addr.Network() != "tcp" {
		t.Error("unexpected TCPAddr network")
	}
	if addr.String(buf[:]) != "127.0.0.1:80" {
		t.Error("unexpected TCPAddr text for an IPv4 address")
	}

	addr, err = net.ResolveTCPAddr("tcp", "[::1]:80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if addr.String(buf[:]) != "[::1]:80" {
		t.Error("unexpected TCPAddr text for an IPv6 address")
	}
}

func TestTCP_ResolveHostname(t *testing.T) {
	// "localhost" resolves via getaddrinfo (the system resolver), without any
	// external DNS. It must come back as a loopback address.
	addr, err := net.ResolveTCPAddr("tcp", "localhost:80")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if addr.Port != 80 {
		t.Error("unexpected port")
	}
	if !addr.IP.IsLoopback() {
		t.Error("localhost should resolve to a loopback address")
	}
}

func TestTCP_Listen(t *testing.T) {
	// Resolve an IP literal (no DNS).
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil || laddr.Port != 0 {
		t.Fatal("failed to resolve listen address")
		return
	}

	// Listen on an OS-assigned port.
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	if ln.Addr().Port == 0 {
		t.Error("listener port not assigned")
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
	}
}

func TestTCP_ListenAll(t *testing.T) {
	// A nil laddr binds the unspecified address (all interfaces), with an
	// OS-assigned port.
	ln, err := net.ListenTCP("tcp", nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if ln.Addr().Port == 0 {
		t.Error("listener port not assigned")
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
	}
}

func TestTCP_Dial(t *testing.T) {
	// A single-threaded loopback echo. Without goroutines this works because the
	// connect completes into the listener backlog and the small payload fits in
	// the kernel socket buffers, so no call blocks waiting on another thread.

	// Listen on an OS-assigned port (IP literal, no DNS).
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// Connect to the listener, binding to an explicit local address (an
	// ephemeral port on the loopback interface) to exercise bind-before-connect.
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", &laddr, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// Accept the queued connection.
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// The endpoints' addresses must line up: the client's remote address is the
	// listener, and the server's remote address is the client's local address.
	if conn.RemoteAddr().Port != raddr.Port {
		t.Error("client remote addr mismatch")
	}
	if conn.LocalAddr().Port == 0 || conn.LocalAddr().Port != server.RemoteAddr().Port {
		t.Error("local/remote addr mismatch")
	}

	// Client writes, server echoes, client reads it back.
	msg := []byte("ping")
	if _, err := conn.Write(msg); err != nil {
		t.Fatal(err.Error())
		return
	}

	var buf [256]byte
	n, err := server.Read(buf[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if _, err := server.Write(buf[:n]); err != nil {
		t.Fatal(err.Error())
		return
	}

	var got [256]byte
	n, err = conn.Read(got[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(got[:n]) != "ping" {
		t.Error("echo mismatch")
	}

	conn.Close()
	server.Close()
	ln.Close()
}

func TestTCP_DialRefused(t *testing.T) {
	// Bind a port, learn its address, then close the listener so nothing is
	// listening there. Dialing it must be refused.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}

	if _, err := net.DialTCP("tcp", nil, &raddr); err != net.ErrConnRefused {
		t.Error("expected connection refused")
	}
}

func TestTCP_ReadEOF(t *testing.T) {
	// Connect a pair, then close the server end. The client's next read must
	// report end of stream.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	if err := server.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf [16]byte
	if _, err := conn.Read(buf[:]); err != io.EOF {
		t.Error("expected EOF")
	}

	conn.Close()
	ln.Close()
}

func TestTCP_ReadDeadline(t *testing.T) {
	// Set up a connected pair, then read on the server side with no data sent.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// Nothing is written, so a read with a short deadline must time out.
	err = server.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf [16]byte
	if _, err := server.Read(buf[:]); err != net.ErrTimeout {
		t.Error("expected timeout")
	}

	conn.Close()
	server.Close()
	ln.Close()
}

func TestTCP_ClearDeadline(t *testing.T) {
	// After a read deadline fires, clearing it must leave the connection usable.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// Arm a short deadline and let it elapse with no data.
	err = server.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf [16]byte
	if _, err := server.Read(buf[:]); err != net.ErrTimeout {
		t.Error("expected timeout")
	}

	// Clearing the deadline must let a read of already-sent data succeed instead
	// of timing out. (Data is sent first because there is no second thread to
	// write during a blocking read.)
	if _, err = conn.Write([]byte("hi")); err != nil {
		t.Fatal(err.Error())
		return
	}
	if err := server.SetReadDeadline(time.Time{}); err != nil {
		t.Fatal(err.Error())
		return
	}
	n, err := server.Read(buf[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(buf[:n]) != "hi" {
		t.Error("read after clearing deadline failed")
	}

	conn.Close()
	server.Close()
	ln.Close()
}

func TestTCP_AcceptDeadline(t *testing.T) {
	// A listener with a short deadline and no incoming connection must time out.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	err = ln.SetDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if _, err := ln.Accept(); err != net.ErrTimeout {
		t.Error("expected timeout")
	}

	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
}

func TestTCP_CloseErrors(t *testing.T) {
	// A double close, and any I/O after close, must report ErrClosed on both
	// connections and listeners.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", &lnAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
	if err := conn.Close(); err != net.ErrClosed {
		t.Error("expected ErrClosed on double close")
	}
	var buf [16]byte
	if _, err := conn.Read(buf[:]); err != net.ErrClosed {
		t.Error("expected ErrClosed on read after close")
	}
	if _, err := conn.Write(buf[:]); err != net.ErrClosed {
		t.Error("expected ErrClosed on write after close")
	}

	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
	if err := ln.Close(); err != net.ErrClosed {
		t.Error("expected ErrClosed on double close (listener)")
	}
	if _, err := ln.Accept(); err != net.ErrClosed {
		t.Error("expected ErrClosed on accept after close")
	}

	server.Close()
}

func TestTCP_InvalidConn(t *testing.T) {
	// DialTCP returns the zero TCPConn when it fails. A caller that ignores
	// the error gets an unopened connection, whose descriptor is 0 (standard
	// input). Every method must report the unopened connection.
	conn, err := net.DialTCP("bogus", nil, nil)
	if err != net.ErrUnknownNetwork {
		t.Fatal("expected ErrUnknownNetwork from DialTCP")
		return
	}
	var buf [16]byte
	var zero time.Time
	if _, err := conn.Read(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on read from an unopened conn")
	}
	if _, err := conn.Write(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on write to an unopened conn")
	}
	if conn.SetDeadline(zero) != net.ErrInvalid {
		t.Error("expected ErrInvalid on SetDeadline of an unopened conn")
	}
	if conn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of an unopened conn")
	}

	var nilConn *net.TCPConn
	if _, err := nilConn.Read(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on read from a nil conn")
	}
	if nilConn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of a nil conn")
	}

	// The same holds for a listener.
	ln, err := net.ListenTCP("bogus", nil)
	if err != net.ErrUnknownNetwork {
		t.Fatal("expected ErrUnknownNetwork from ListenTCP")
		return
	}
	if _, err := ln.Accept(); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on accept from an unopened listener")
	}
	if ln.SetDeadline(zero) != net.ErrInvalid {
		t.Error("expected ErrInvalid on SetDeadline of an unopened listener")
	}
	if ln.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of an unopened listener")
	}

	var nilLn *net.TCPListener
	if _, err := nilLn.Accept(); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on accept from a nil listener")
	}
	if nilLn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of a nil listener")
	}
}
