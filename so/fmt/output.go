package fmt

import (
	"os" // for testing

	"solod.dev/so/io"
)

// writeOut writes p to the standard output of the host, and returns the number
// of bytes written. A freestanding host has no standard output, so it drops
// the bytes and reports a full write.
//
//so:extern fmt_writeOut nodecay
func writeOut(p []byte) int {
	n, _ := os.Stdout.Write(p)
	return n
}

// stdoutWriter writes to the standard output of the host.
type stdoutWriter struct{}

// Write writes p to the standard output.
func (*stdoutWriter) Write(p []byte) (int, error) {
	n := writeOut(p)
	if n != len(p) {
		return n, ErrPrint
	}
	return n, nil
}

// Output is the destination of [Print], [Println] and [Printf]. Its default is
// the standard output of the host. A freestanding host has no standard output,
// so the default drops the bytes there. Assign another writer, a UART or a host
// import, to get the output back.
var Output io.Writer = &stdoutWriter{}
