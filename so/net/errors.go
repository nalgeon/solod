package net

import "solod.dev/so/errors"

// Errors that can be returned by functions in this package.
var (
	// ErrAddrInUse indicates the local address is already in use.
	ErrAddrInUse = errors.New("net: address already in use")
	// ErrAddrNotAvail indicates the requested address is not available.
	ErrAddrNotAvail = errors.New("net: cannot assign requested address")
	// ErrBrokenPipe indicates a write to a connection whose peer has
	// already closed its end (the local send saw EPIPE).
	ErrBrokenPipe = errors.New("net: broken pipe")
	// ErrClosed is returned by I/O methods on a connection or listener
	// that has already been closed.
	ErrClosed = errors.New("net: use of closed network connection")
	// ErrConnAborted indicates the connection was aborted locally.
	ErrConnAborted = errors.New("net: connection aborted")
	// ErrConnRefused indicates the remote host refused the connection.
	ErrConnRefused = errors.New("net: connection refused")
	// ErrConnReset indicates the connection was reset by the peer.
	ErrConnReset = errors.New("net: connection reset by peer")
	// ErrInvalid is returned by methods on a connection or listener that was
	// never opened, for example the zero value or a nil pointer.
	ErrInvalid = errors.New("net: invalid argument")
	// ErrInvalidPort indicates the port is not a valid number in 0..65535.
	ErrInvalidPort = errors.New("net: invalid port")
	// ErrMissingBracket indicates an IPv6 literal is missing its closing ']'.
	ErrMissingBracket = errors.New("net: missing ']' in address")
	// ErrMissingPort indicates an address is missing the port.
	ErrMissingPort = errors.New("net: missing port in address")
	// ErrMsgTooLong indicates a datagram was larger
	// than the maximum allowed size.
	ErrMsgTooLong = errors.New("net: message too long")
	// ErrNoSuchHost indicates the host could not be resolved.
	ErrNoSuchHost = errors.New("net: no such host")
	// ErrNoSuitableAddr indicates the address does not match the network's
	// family, for example an IPv6 literal with the "tcp4" network.
	ErrNoSuitableAddr = errors.New("net: no suitable address found")
	// ErrTimeout indicates the operation timed out.
	ErrTimeout = errors.New("net: operation timed out")
	// ErrTooManyColons indicates an unbracketed address has too many colons.
	ErrTooManyColons = errors.New("net: too many colons in address")
	// ErrUnexpectedBracket indicates an unexpected bracket in the address.
	ErrUnexpectedBracket = errors.New("net: unexpected '[' or ']' in address")
	// ErrUnknownNetwork is returned when the network argument is not one of
	// "tcp", "tcp4", or "tcp6".
	ErrUnknownNetwork = errors.New("net: unknown network")
	// ErrUnreachable indicates the network or host is unreachable.
	ErrUnreachable = errors.New("net: network or host is unreachable")

	// ErrIO is a generic network error returned when the cause does not
	// match any of the other, more specific errors.
	ErrIO = errors.New("net: i/o error")
)

// mapError maps the current errno to a sentinel error.
func mapError() error {
	// Uses an if-chain rather than a switch because the errno constants are
	// all zero in the Go stubs (resolved to real values only after transpiling),
	// and a switch with duplicate case constants would not compile.
	if errno == eADDRINUSE {
		return ErrAddrInUse
	}
	if errno == eADDRNOTAVAIL {
		return ErrAddrNotAvail
	}
	if errno == ePIPE {
		return ErrBrokenPipe
	}
	if errno == eCONNREFUSED {
		return ErrConnRefused
	}
	if errno == eCONNRESET {
		return ErrConnReset
	}
	if errno == eCONNABORTED {
		return ErrConnAborted
	}
	if errno == eMSGSIZE {
		return ErrMsgTooLong
	}
	if errno == eTIMEDOUT {
		return ErrTimeout
	}
	if errno == eNETUNREACH || errno == eHOSTUNREACH {
		return ErrUnreachable
	}
	return ErrIO
}
