package os_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestTempDir(t *testing.T) {
	td := os.TempDir()
	if len(td) == 0 {
		t.Error("TempDir: empty")
	}
}

func TestCreateTemp(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	f, err := os.CreateTemp(buf, "", "sotest")
	if err != nil {
		t.Fatal("CreateTemp failed")
		return
	}
	name := f.Name()
	if len(name) == 0 {
		t.Fatal("CreateTemp: empty name")
		return
	}
	defer os.Remove(name)

	// Name should contain the pattern prefix.
	if !strings.Contains(name, "sotest") {
		t.Error("CreateTemp: name missing pattern")
	}
	f.Write([]byte("temp data"))
	f.Close()

	// Verify the file exists.
	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile temp failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)
	if string(b) != "temp data" {
		t.Error("CreateTemp: wrong data")
	}
}

func TestCreateTemp_Dir(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	td := os.TempDir()
	f, err := os.CreateTemp(buf, td, "myprefix")
	if err != nil {
		t.Fatal("CreateTemp dir failed")
		return
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()

	if !strings.Contains(name, "myprefix") {
		t.Error("CreateTemp dir: missing pattern")
	}
	if !strings.HasPrefix(name, td) {
		t.Error("CreateTemp dir: wrong dir")
	}
}

func TestMkdirTemp(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	dir, err := os.MkdirTemp(buf, "", "sotest")
	if err != nil {
		t.Fatal("MkdirTemp failed")
		return
	}
	if len(dir) == 0 {
		t.Fatal("MkdirTemp: empty")
		return
	}
	defer os.Remove(dir)

	if !strings.Contains(dir, "sotest") {
		t.Error("MkdirTemp: name missing pattern")
	}

	// Verify it's a directory.
	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatal("Stat MkdirTemp failed")
		return
	}
	if !fi.IsDir() {
		t.Error("MkdirTemp: not a directory")
	}
}
