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
	alloc := t.Allocator()
	buf := make([]byte, os.MaxPathLen)
	f, err := os.CreateTemp(buf, "", "sotest")
	if err != nil {
		t.Fatalf("CreateTemp: %s", errText(err))
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
		t.Errorf("CreateTemp = %s, want the sotest pattern in the name", name)
	}
	f.Write([]byte("temp data"))
	f.Close()

	// Verify the file exists.
	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)
	if string(b) != "temp data" {
		t.Errorf("ReadFile = %s, want temp data", string(b))
	}
}

func TestCreateTemp_Dir(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	td := os.TempDir()
	f, err := os.CreateTemp(buf, td, "myprefix")
	if err != nil {
		t.Fatalf("CreateTemp: %s", errText(err))
		return
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()

	if !strings.Contains(name, "myprefix") {
		t.Errorf("CreateTemp = %s, want the myprefix pattern in the name", name)
	}
	if !strings.HasPrefix(name, td) {
		t.Errorf("CreateTemp = %s, want the name in %s", name, td)
	}
}

func TestCreateTemp_Unique(t *testing.T) {
	// Each call gets a name of its own.
	// Each name is a view into its own buffer.
	buf1 := make([]byte, os.MaxPathLen)
	f1, err := os.CreateTemp(buf1, "", "sounique")
	if err != nil {
		t.Fatalf("CreateTemp 1: %s", errText(err))
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	buf2 := make([]byte, os.MaxPathLen)
	f2, err := os.CreateTemp(buf2, "", "sounique")
	if err != nil {
		t.Fatalf("CreateTemp 2: %s", errText(err))
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	if f1.Name() == f2.Name() {
		t.Errorf("two CreateTemp calls both gave %s, want two names", f1.Name())
	}
}

func TestCreateTemp_Mode(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	f, err := os.CreateTemp(buf, "", "somode")
	if err != nil {
		t.Fatalf("CreateTemp: %s", errText(err))
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// CreateTemp makes the file readable and writable by the owner only.
	fi, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("Perm = %o, want 600", int(fi.Mode().Perm()))
	}
}

func TestMkdirTemp(t *testing.T) {
	buf := make([]byte, os.MaxPathLen)
	dir, err := os.MkdirTemp(buf, "", "sotest")
	if err != nil {
		t.Fatalf("MkdirTemp: %s", errText(err))
		return
	}
	if len(dir) == 0 {
		t.Fatal("MkdirTemp: empty")
		return
	}
	defer os.Remove(dir)

	if !strings.Contains(dir, "sotest") {
		t.Errorf("MkdirTemp = %s, want the sotest pattern in the name", dir)
	}

	// Verify it's a directory.
	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if !fi.IsDir() {
		t.Errorf("MkdirTemp = %s, want a directory", dir)
	}
	// MkdirTemp makes the directory usable by the owner only.
	if fi.Mode().Perm() != 0o700 {
		t.Errorf("Perm = %o, want 700", int(fi.Mode().Perm()))
	}
}

func TestMkdirTemp_Dir(t *testing.T) {
	parent := os.TempDir() + "/so_os_mkdirtemp_parent"
	if err := os.Mkdir(parent, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(parent)

	buf := make([]byte, os.MaxPathLen)
	dir, err := os.MkdirTemp(buf, parent, "child")
	if err != nil {
		t.Fatalf("MkdirTemp: %s", errText(err))
		return
	}
	defer os.Remove(dir)

	if !strings.HasPrefix(dir, parent) {
		t.Errorf("MkdirTemp = %s, want the name in %s", dir, parent)
	}
}
