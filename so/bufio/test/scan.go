package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestScanner(t *testing.T) {
	sr := strings.NewReader("line1\nline2\n")
	s := bufio.NewScanner(mem.System, &sr)
	defer s.Free()

	count := 0
	for s.Scan() {
		if count == 0 && s.Text() != "line1" {
			t.Error("Scanner line 0 = " + s.Text() + ", want line1")
		}
		if count == 1 && s.Text() != "line2" {
			t.Error("Scanner line 1 = " + s.Text() + ", want line2")
		}
		count++
	}
	if count != 2 {
		t.Error("Scanner scanned wrong number of lines")
	}
}
