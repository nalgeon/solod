package os_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestWriteReadFile(t *testing.T) {
	name := "test_rw.txt"
	data := []byte("hello world")
	err := os.WriteFile(name, data, 0o666)
	if err != nil {
		t.Fatal("WriteFile failed")
		return
	}
	defer os.Remove(name)

	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != string(data) {
		t.Error("ReadFile: wrong data")
	}
}

func TestCreateWriteClose(t *testing.T) {
	name := "test_file.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)

	n, err := f.Write([]byte("abcdef"))
	if err != nil {
		t.Fatal("Write failed")
		return
	}
	if n != 6 {
		t.Error("Write: wrong count")
	}
	if err := f.Close(); err != nil {
		t.Error("Close failed")
	}
}

func TestOpenReadClose(t *testing.T) {
	name := "test_file.txt"
	data := []byte("abcdef")
	if err := os.WriteFile(name, data, 0o666); err != nil {
		t.Fatal("WriteFile failed")
		return
	}
	f, err := os.Open(name)
	if err != nil {
		t.Fatal("Open failed")
		return
	}
	defer os.Remove(name)

	buf := make([]byte, 10)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatal("Read failed")
		return
	}
	if n != 6 {
		t.Error("Read: wrong count")
	}
	if string(buf[:n]) != "abcdef" {
		t.Error("Read: wrong data")
	}
	if err := f.Close(); err != nil {
		t.Error("Close failed")
	}
}

func TestWriteString(t *testing.T) {
	name := "test_writestr.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)

	n, err := f.WriteString("hello")
	if err != nil {
		t.Fatal("WriteString failed")
		return
	}
	if n != 5 {
		t.Error("WriteString: wrong count")
	}
	f.Close()

	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "hello" {
		t.Error("WriteString: wrong data")
	}
}

func TestStdoutStderr(t *testing.T) {
	n, err := os.Stdout.WriteString("hello")
	if err != nil {
		t.Fatal("Stdout failed")
		return
	}
	if n != 5 {
		t.Error("Stdout: wrong count")
	}
	n, err = os.Stderr.WriteString("goodbye")
	if err != nil {
		t.Fatal("Stderr failed")
		return
	}
	if n != 7 {
		t.Error("Stderr: wrong count")
	}
	println()
}
