package os_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestSeek(t *testing.T) {
	name := "test_seek.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)
	defer f.Close()

	f.Write([]byte("abcdef"))
	pos, err := f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal("Seek failed")
		return
	}
	if pos != 0 {
		t.Error("Seek: wrong position")
	}

	buf := make([]byte, 6)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatal("Read after Seek failed")
		return
	}
	if string(buf[:n]) != "abcdef" {
		t.Error("Seek: wrong data")
	}
}

func TestReadAt(t *testing.T) {
	name := "test_readat.txt"
	err := os.WriteFile(name, []byte("hello world"), 0o666)
	if err != nil {
		t.Fatal("WriteFile failed")
		return
	}
	defer os.Remove(name)

	f, err := os.Open(name)
	if err != nil {
		t.Fatal("Open failed")
		return
	}
	defer f.Close()

	buf := make([]byte, 5)
	n, err := f.ReadAt(buf, 6)
	if err != nil {
		t.Fatal("ReadAt failed")
		return
	}
	if n != 5 {
		t.Error("ReadAt: wrong count")
	}
	if string(buf[:n]) != "world" {
		t.Error("ReadAt: wrong data")
	}
}

func TestWriteAt(t *testing.T) {
	name := "test_writeat.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)
	defer f.Close()

	f.Write([]byte("hello world"))
	_, err = f.WriteAt([]byte("WORLD"), 6)
	if err != nil {
		t.Fatal("WriteAt failed")
		return
	}

	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "hello WORLD" {
		t.Error("WriteAt: wrong data")
	}
}
