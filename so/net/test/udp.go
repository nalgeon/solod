package main

import (
	"solod.dev/so/net"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

func TestUDP_ResolveAddr(t *testing.T) {
	// A named port resolves via the udp services database (no DNS for the host).
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:domain")
	if err != nil || addr.Port != 53 {
		t.Error("failed to resolve named UDP port")
	}

	// "localhost" resolves via getaddrinfo (the system resolver), without any
	// external DNS. It must come back as a loopback address.
	addr, err = net.ResolveUDPAddr("udp", "localhost:53")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if addr.Port != 53 {
		t.Error("unexpected port")
	}
	if !addr.IP.IsLoopback() {
		t.Error("localhost should resolve to a loopback address")
	}
}

func TestUDP_Listen(t *testing.T) {
	// Resolve an IP literal (no DNS) and listen on an OS-assigned port.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil || laddr.Port != 0 {
		t.Fatal("failed to resolve listen address")
		return
	}

	conn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if conn.LocalAddr().Port == 0 {
		t.Error("listener port not assigned")
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
}

func TestUDP_Dial(t *testing.T) {
	// A single-threaded loopback echo. Datagrams are buffered in the kernel, so
	// no call blocks waiting on another thread.

	// Server listens on an OS-assigned port.
	srvAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	server, err := net.ListenUDP("udp", &srvAddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// Client connects to the server.
	raddr := server.LocalAddr()
	client, err := net.DialUDP("udp", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// The client's remote address is the server.
	if client.RemoteAddr().Port != raddr.Port {
		t.Error("client remote addr mismatch")
	}

	// Client writes a datagram; the server receives it and learns the client's
	// address, then echoes it back via WriteTo.
	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err.Error())
		return
	}

	var buf [256]byte
	r, err := server.ReadFrom(buf[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if r.Addr.Port != client.LocalAddr().Port {
		t.Error("server learned wrong client addr")
	}
	if _, err := server.WriteTo(buf[:r.N], &r.Addr); err != nil {
		t.Fatal(err.Error())
		return
	}

	// Client reads the echo on its connected socket.
	var got [256]byte
	n, err := client.Read(got[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(got[:n]) != "ping" {
		t.Error("echo mismatch")
	}

	client.Close()
	server.Close()
}

func TestUDP_ReadFromWriteTo(t *testing.T) {
	// Two unconnected sockets exchange datagrams in both directions, with each
	// receiver checking the reported source address against the sender's local
	// address.
	addrA, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	a, err := net.ListenUDP("udp", &addrA)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	addrB, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	b, err := net.ListenUDP("udp", &addrB)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	// A -> B.
	bAddr := b.LocalAddr()
	if _, err := a.WriteTo([]byte("ping"), &bAddr); err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf [256]byte
	r, err := b.ReadFrom(buf[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(buf[:r.N]) != "ping" {
		t.Error("A->B payload mismatch")
	}
	if r.Addr.Port != a.LocalAddr().Port {
		t.Error("A->B source addr mismatch")
	}

	// B -> A, replying to the learned source address.
	if _, err := b.WriteTo([]byte("pong"), &r.Addr); err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf2 [256]byte
	r2, err := a.ReadFrom(buf2[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(buf2[:r2.N]) != "pong" {
		t.Error("B->A payload mismatch")
	}
	if r2.Addr.Port != b.LocalAddr().Port {
		t.Error("B->A source addr mismatch")
	}

	a.Close()
	b.Close()
}

func TestUDP_ReadDeadline(t *testing.T) {
	// A ReadFrom with a short deadline and no data must time out.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var buf [16]byte
	if _, err := conn.ReadFrom(buf[:]); err != net.ErrTimeout {
		t.Error("expected timeout")
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}
}

func TestUDP_CloseErrors(t *testing.T) {
	// A double close, and any I/O after close, must report ErrClosed.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", &laddr)
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
	if _, err := conn.ReadFrom(buf[:]); err != net.ErrClosed {
		t.Error("expected ErrClosed on ReadFrom after close")
	}
	if _, err := conn.WriteTo(buf[:], &laddr); err != net.ErrClosed {
		t.Error("expected ErrClosed on WriteTo after close")
	}
}

func TestUDP_InvalidConn(t *testing.T) {
	// ListenUDP returns the zero UDPConn when it fails. A caller that ignores
	// the error gets an unopened connection, whose descriptor is 0 (standard
	// input). Every method must report the unopened connection.
	conn, err := net.ListenUDP("bogus", nil)
	if err != net.ErrUnknownNetwork {
		t.Fatal("expected ErrUnknownNetwork from ListenUDP")
		return
	}
	var buf [16]byte
	var addr net.UDPAddr
	var zero time.Time
	if _, err := conn.Read(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on read from an unopened conn")
	}
	if _, err := conn.Write(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on write to an unopened conn")
	}
	if _, err := conn.ReadFrom(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on ReadFrom of an unopened conn")
	}
	if _, err := conn.WriteTo(buf[:], &addr); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on WriteTo of an unopened conn")
	}
	if conn.SetDeadline(zero) != net.ErrInvalid {
		t.Error("expected ErrInvalid on SetDeadline of an unopened conn")
	}
	if conn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of an unopened conn")
	}

	var nilConn *net.UDPConn
	if _, err := nilConn.ReadFrom(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on ReadFrom of a nil conn")
	}
	if nilConn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of a nil conn")
	}
}
