package main

import (
	"solod.dev/so/io"
	"solod.dev/so/net"
	"solod.dev/so/os"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// makeUnixDir creates a temporary directory to hold test socket files, built
// into buf. It returns the directory and false (after t.Fatal) on failure.
func makeUnixDir(t *testing.T, buf []byte) (string, bool) {
	dir, err := os.MkdirTemp(buf, "", "so-net-unix")
	if err != nil {
		t.Fatal("MkdirTemp failed")
		return "", false
	}
	return dir, true
}

// unixPath builds dir + "/" + name into buf and returns it.
func unixPath(buf []byte, dir, name string) string {
	b := buf[:0]
	b = append(b, dir...)
	b = append(b, '/')
	b = append(b, name...)
	return string(b)
}

func TestUnix_Resolve(t *testing.T) {
	addr, err := net.ResolveUnixAddr("unix", "/tmp/echo.sock")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if addr.Name != "/tmp/echo.sock" || addr.Net != "unix" {
		t.Error("unexpected ResolveUnixAddr result")
	}
	if addr.Network() != "unix" || addr.String() != "/tmp/echo.sock" {
		t.Error("unexpected UnixAddr Network/String")
	}

	gram, err := net.ResolveUnixAddr("unixgram", "/tmp/dg.sock")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if gram.Net != "unixgram" || gram.Network() != "unixgram" {
		t.Error("unexpected unixgram network")
	}

	// unixpacket is intentionally unsupported, as is any other network.
	if _, err := net.ResolveUnixAddr("unixpacket", "/tmp/x.sock"); err != net.ErrUnknownNetwork {
		t.Error("unixpacket should be unknown")
	}
	if _, err := net.ResolveUnixAddr("bogus", "/tmp/x.sock"); err != net.ErrUnknownNetwork {
		t.Error("bogus network should be unknown")
	}
}

func TestUnix_StreamDial(t *testing.T) {
	// A single-threaded loopback echo: the connect queues into the listener
	// backlog, so Accept does not block on another thread.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	laddr, err := net.ResolveUnixAddr("unix", unixPath(pathBuf[:], dir, "stream.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenUnix("unix", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if ln.Addr().Name != laddr.Name {
		t.Error("listener addr mismatch")
	}

	raddr := ln.Addr()
	client, err := net.DialUnix("unix", nil, &raddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if client.RemoteAddr().Name != raddr.Name {
		t.Error("client remote addr mismatch")
	}

	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if server.LocalAddr().Name != raddr.Name {
		t.Error("accepted local addr mismatch")
	}

	// Client writes, server echoes, client reads it back.
	if _, err := client.Write([]byte("ping")); err != nil {
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
	n, err = client.Read(got[:])
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if string(got[:n]) != "ping" {
		t.Error("echo mismatch")
	}

	client.Close()
	server.Close()
	ln.Close()
}

func TestUnix_StreamReadEOF(t *testing.T) {
	// Connect a pair, then close the server end; the client's next read must
	// report end of stream.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	laddr, err := net.ResolveUnixAddr("unix", unixPath(pathBuf[:], dir, "eof.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenUnix("unix", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	raddr := ln.Addr()
	client, err := net.DialUnix("unix", nil, &raddr)
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
	if _, err := client.Read(buf[:]); err != io.EOF {
		t.Error("expected EOF")
	}

	client.Close()
	ln.Close()
}

func TestUnix_DialRefused(t *testing.T) {
	// Dialing a path with no socket file (nothing listening) must fail.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	addr, err := net.ResolveUnixAddr("unix", unixPath(pathBuf[:], dir, "refused.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if _, err := net.DialUnix("unix", nil, &addr); err == nil {
		t.Error("expected dial to a missing socket to fail")
	}
}

func TestUnix_Datagram(t *testing.T) {
	// Two bound datagram sockets exchange messages in both directions, each
	// receiver checking the reported source path against the sender's address.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBufA [320]byte
	var pathBufB [320]byte
	addrA, err := net.ResolveUnixAddr("unixgram", unixPath(pathBufA[:], dir, "dga.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	a, err := net.ListenUnixgram("unixgram", &addrA)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	addrB, err := net.ResolveUnixAddr("unixgram", unixPath(pathBufB[:], dir, "dgb.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	b, err := net.ListenUnixgram("unixgram", &addrB)
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
	if r.Addr.Name != a.LocalAddr().Name {
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
	if r2.Addr.Name != b.LocalAddr().Name {
		t.Error("B->A source addr mismatch")
	}

	a.Close()
	b.Close()
}

func TestUnix_ReadDeadline(t *testing.T) {
	// A ReadFrom with a short deadline and no data must time out.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	laddr, err := net.ResolveUnixAddr("unixgram", unixPath(pathBuf[:], dir, "dl.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	conn, err := net.ListenUnixgram("unixgram", &laddr)
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
	}
}

func TestUnix_CloseErrors(t *testing.T) {
	// A double close, and any I/O after close, must report ErrClosed.
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	laddr, err := net.ResolveUnixAddr("unixgram", unixPath(pathBuf[:], dir, "close.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	conn, err := net.ListenUnixgram("unixgram", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err.Error())
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

func TestUnix_UnlinkOnClose(t *testing.T) {
	// Listening creates the socket file; Close must remove it. After Close, the
	// path is gone, so removing it again reports "not exist".
	var dirBuf [256]byte
	dir, ok := makeUnixDir(t, dirBuf[:])
	if !ok {
		return
	}
	defer os.Remove(dir)

	var pathBuf [320]byte
	laddr, err := net.ResolveUnixAddr("unix", unixPath(pathBuf[:], dir, "unlink.sock"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	ln, err := net.ListenUnix("unix", &laddr)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err.Error())
		return
	}

	if err := os.Remove(laddr.Name); err != os.ErrNotExist {
		t.Error("socket file should have been unlinked on Close")
	}
}

func TestUnix_InvalidConn(t *testing.T) {
	// DialUnix returns the zero UnixConn when it fails. A caller that ignores
	// the error gets an unopened connection, whose descriptor is 0 (standard
	// input). Every method must report the unopened connection.
	conn, err := net.DialUnix("bogus", nil, nil)
	if err != net.ErrUnknownNetwork {
		t.Fatal("expected ErrUnknownNetwork from DialUnix")
		return
	}
	var buf [16]byte
	var addr net.UnixAddr
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

	var nilConn *net.UnixConn
	if _, err := nilConn.Read(buf[:]); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on read from a nil conn")
	}
	if nilConn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of a nil conn")
	}

	// The same holds for a listener.
	ln, err := net.ListenUnix("bogus", nil)
	if err != net.ErrUnknownNetwork {
		t.Fatal("expected ErrUnknownNetwork from ListenUnix")
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

	var nilLn *net.UnixListener
	if _, err := nilLn.Accept(); err != net.ErrInvalid {
		t.Error("expected ErrInvalid on accept from a nil listener")
	}
	if nilLn.Close() != net.ErrInvalid {
		t.Error("expected ErrInvalid on close of a nil listener")
	}
}
