package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestArgs(t *testing.T) {
	// os.Args should be populated.
	if len(os.Args) == 0 {
		t.Fatal("os.Args: empty")
		return
	}
	// First arg (program name) should be non-empty.
	if len(os.Args[0]) == 0 || os.Args[0] == "" {
		t.Error("os.Args[0]: empty")
	}
}
