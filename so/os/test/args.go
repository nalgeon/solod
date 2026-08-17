package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestArgs(t *testing.T) {
	if len(os.Args) == 0 {
		t.Fatal("os.Args: empty")
		return
	}
	// The first argument is the program name.
	if os.Args[0] == "" {
		t.Error("os.Args[0]: empty")
	}
}
