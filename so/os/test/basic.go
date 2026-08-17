package os_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestWriteReadFile(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_rw.txt"
	data := []byte("hello world")
	err := os.WriteFile(name, data, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != string(data) {
		t.Errorf("ReadFile = %s, want %s", string(b), string(data))
	}
}

func TestWriteReadFile_Empty(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_rw_empty.txt"
	err := os.WriteFile(name, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if len(b) != 0 {
		t.Errorf("ReadFile = %s, want empty", string(b))
	}
}

func TestWriteReadFile_Large(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_rw_large.txt"
	// The read buffer of io.ReadAll starts at 512 bytes,
	// so a larger file makes it grow.
	var data [4096]byte
	for i := range data {
		data[i] = byte('a' + i%26)
	}
	err := os.WriteFile(name, data[:], 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if len(b) != len(data) {
		t.Errorf("ReadFile length = %d, want %d", len(b), len(data))
		return
	}
	if !bytes.Equal(b, data[:]) {
		t.Error("ReadFile: wrong data")
	}
}

func TestWriteFile_Perm(t *testing.T) {
	name := os.TempDir() + "/so_os_wf_perm.txt"
	err := os.WriteFile(name, []byte("perm"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("Perm = %o, want 600", int(fi.Mode().Perm()))
	}
}

func TestWriteFile_Truncate(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_wf_trunc.txt"
	err := os.WriteFile(name, []byte("first write"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// The second write truncates the file and keeps the permissions.
	if err := os.Chmod(name, 0o640); err != nil {
		t.Fatalf("Chmod: %s", errText(err))
		return
	}
	if err := os.WriteFile(name, []byte("second"), 0o600); err != nil {
		t.Fatalf("WriteFile again: %s", errText(err))
		return
	}

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "second" {
		t.Errorf("ReadFile = %s, want second", string(b))
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode().Perm() != 0o640 {
		t.Errorf("Perm = %o, want 640", int(fi.Mode().Perm()))
	}
}

func TestCreateWriteClose(t *testing.T) {
	name := os.TempDir() + "/so_os_create.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)

	n, err := f.Write([]byte("abcdef"))
	if err != nil {
		t.Fatalf("Write: %s", errText(err))
		return
	}
	if n != 6 {
		t.Errorf("Write = %d, want 6", n)
	}
	if err := f.Close(); err != nil {
		t.Errorf("Close: %s", errText(err))
	}
}

func TestOpenReadClose(t *testing.T) {
	name := os.TempDir() + "/so_os_open.txt"
	data := []byte("abcdef")
	if err := os.WriteFile(name, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Open: %s", errText(err))
		return
	}

	buf := make([]byte, 10)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read: %s", errText(err))
		return
	}
	if n != 6 {
		t.Errorf("Read = %d, want 6", n)
	}
	if string(buf[:n]) != "abcdef" {
		t.Errorf("Read data = %s, want abcdef", string(buf[:n]))
	}
	if err := f.Close(); err != nil {
		t.Errorf("Close: %s", errText(err))
	}
}

func TestRead_EOF(t *testing.T) {
	name := os.TempDir() + "/so_os_read_eof.txt"
	if err := os.WriteFile(name, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Open: %s", errText(err))
		return
	}
	defer f.Close()

	// The first read gets the whole file.
	buf := make([]byte, 3)
	n, err := f.Read(buf)
	if n != 3 || err != nil {
		t.Errorf("first Read = %d, %s, want 3, nil", n, errText(err))
	}
	// The read at the end of the file gets nothing.
	n, err = f.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("Read at EOF = %d, %s, want 0, %s", n, errText(err), io.EOF.Error())
	}
}

func TestReadWrite_Empty(t *testing.T) {
	name := os.TempDir() + "/so_os_empty_arg.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)
	defer f.Close()

	n, err := f.Write([]byte{})
	if n != 0 || err != nil {
		t.Errorf("Write(empty) = %d, %s, want 0, nil", n, errText(err))
	}
	n, err = f.Read([]byte{})
	if n != 0 || err != nil {
		t.Errorf("Read(empty) = %d, %s, want 0, nil", n, errText(err))
	}
}

func TestWrite_ReadOnly(t *testing.T) {
	name := os.TempDir() + "/so_os_write_ro.txt"
	if err := os.WriteFile(name, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Open: %s", errText(err))
		return
	}
	defer f.Close()

	// A write to a file open for reading fails and writes nothing.
	n, err := f.Write([]byte("xyz"))
	if err == nil {
		t.Error("Write to a read-only file: want an error")
	}
	if n == 3 {
		t.Error("Write to a read-only file: want a short count")
	}
}

func TestWriteString(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_writestr.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)

	n, err := f.WriteString("hello")
	if err != nil {
		t.Fatalf("WriteString: %s", errText(err))
		return
	}
	if n != 5 {
		t.Errorf("WriteString = %d, want 5", n)
	}
	f.Close()

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "hello" {
		t.Errorf("ReadFile = %s, want hello", string(b))
	}
}

func TestSync(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_sync.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)
	defer f.Close()

	// The write waits in the buffer of the stream. Sync writes it out,
	// so a reader of the file sees it before Close.
	if _, err := f.WriteString("synced"); err != nil {
		t.Fatalf("WriteString: %s", errText(err))
		return
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("Sync: %s", errText(err))
		return
	}

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "synced" {
		t.Errorf("ReadFile after Sync = %s, want synced", string(b))
	}
}

func TestStdoutStderr(t *testing.T) {
	n, err := os.Stdout.WriteString("hello")
	if err != nil {
		t.Fatalf("Stdout: %s", errText(err))
		return
	}
	if n != 5 {
		t.Errorf("Stdout = %d, want 5", n)
	}
	n, err = os.Stderr.WriteString("goodbye")
	if err != nil {
		t.Fatalf("Stderr: %s", errText(err))
		return
	}
	if n != 7 {
		t.Errorf("Stderr = %d, want 7", n)
	}
	println()
}
