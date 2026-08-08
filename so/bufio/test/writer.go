package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestWriter_WriteRune(t *testing.T) {
	var buf bytes.Buffer
	buf = bytes.NewBuffer(mem.System, nil)
	defer buf.Free()

	w := bufio.NewWriter(mem.System, &buf)
	defer w.Free()

	w.WriteRune('A')
	w.Flush()
	if buf.String() != "A" {
		t.Error("WriteRune = " + buf.String() + ", want A")
	}
}
