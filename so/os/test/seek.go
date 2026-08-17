package os_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestSeek(t *testing.T) {
	name := os.TempDir() + "/so_os_seek.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)
	defer f.Close()

	f.Write([]byte("abcdef"))
	pos, err := f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek: %s", errText(err))
		return
	}
	if pos != 0 {
		t.Errorf("Seek(0, SeekStart) = %d, want 0", pos)
	}

	buf := make([]byte, 6)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read after Seek: %s", errText(err))
		return
	}
	if string(buf[:n]) != "abcdef" {
		t.Errorf("Read after Seek = %s, want abcdef", string(buf[:n]))
	}
}

func TestSeek_Whence(t *testing.T) {
	name := os.TempDir() + "/so_os_seek_whence.txt"
	if err := os.WriteFile(name, []byte("abcdef"), 0o600); err != nil {
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

	// Every whence gives the new offset from the start of the file.
	pos, err := f.Seek(2, io.SeekStart)
	if pos != 2 || err != nil {
		t.Errorf("Seek(2, SeekStart) = %d, %s, want 2, nil", pos, errText(err))
	}
	pos, err = f.Seek(1, io.SeekCurrent)
	if pos != 3 || err != nil {
		t.Errorf("Seek(1, SeekCurrent) = %d, %s, want 3, nil", pos, errText(err))
	}
	pos, err = f.Seek(-2, io.SeekEnd)
	if pos != 4 || err != nil {
		t.Errorf("Seek(-2, SeekEnd) = %d, %s, want 4, nil", pos, errText(err))
	}
	pos, err = f.Seek(0, io.SeekEnd)
	if pos != 6 || err != nil {
		t.Errorf("Seek(0, SeekEnd) = %d, %s, want 6, nil", pos, errText(err))
	}

	// An offset before the start of the file fails.
	if _, err := f.Seek(-1, io.SeekStart); err == nil {
		t.Error("Seek(-1, SeekStart): want an error")
	}
}

func TestReadAt(t *testing.T) {
	name := os.TempDir() + "/so_os_readat.txt"
	err := os.WriteFile(name, []byte("hello world"), 0o600)
	if err != nil {
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

	buf := make([]byte, 5)
	n, err := f.ReadAt(buf, 6)
	if err != nil {
		t.Fatalf("ReadAt: %s", errText(err))
		return
	}
	if n != 5 {
		t.Errorf("ReadAt = %d, want 5", n)
	}
	if string(buf[:n]) != "world" {
		t.Errorf("ReadAt data = %s, want world", string(buf[:n]))
	}
}

func TestReadAt_Short(t *testing.T) {
	name := os.TempDir() + "/so_os_readat_short.txt"
	err := os.WriteFile(name, []byte("hello world"), 0o600)
	if err != nil {
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

	// A read that stops at the end of the file reports io.EOF.
	buf := make([]byte, 8)
	n, err := f.ReadAt(buf, 6)
	if n != 5 || err != io.EOF {
		t.Errorf("ReadAt short = %d, %s, want 5, %s", n, errText(err), io.EOF.Error())
	}
	if string(buf[:n]) != "world" {
		t.Errorf("ReadAt short data = %s, want world", string(buf[:n]))
	}

	// A read past the end of the file gets nothing.
	n, err = f.ReadAt(buf, 100)
	if n != 0 || err != io.EOF {
		t.Errorf("ReadAt past EOF = %d, %s, want 0, %s", n, errText(err), io.EOF.Error())
	}
}

func TestReadWriteAt_NegOffset(t *testing.T) {
	name := os.TempDir() + "/so_os_at_negoff.txt"
	err := os.WriteFile(name, []byte("hello"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := f.ReadAt(buf, -1); err != io.ErrOffset {
		t.Errorf("ReadAt(-1) = %s, want %s", errText(err), io.ErrOffset.Error())
	}
	if _, err := f.WriteAt(buf, -1); err != io.ErrOffset {
		t.Errorf("WriteAt(-1) = %s, want %s", errText(err), io.ErrOffset.Error())
	}
}

func TestReadWriteAt_KeepOffset(t *testing.T) {
	name := os.TempDir() + "/so_os_at_offset.txt"
	err := os.WriteFile(name, []byte("hello world"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer f.Close()

	// ReadAt and WriteAt keep the offset of the next Read and Write.
	buf := make([]byte, 5)
	if _, err := f.Read(buf); err != nil {
		t.Fatalf("Read: %s", errText(err))
		return
	}
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatalf("ReadAt: %s", errText(err))
		return
	}
	pos, err := f.Seek(0, io.SeekCurrent)
	if pos != 5 || err != nil {
		t.Errorf("offset after ReadAt = %d, %s, want 5, nil", pos, errText(err))
	}

	if _, err := f.WriteAt([]byte("HELLO"), 0); err != nil {
		t.Fatalf("WriteAt: %s", errText(err))
		return
	}
	pos, err = f.Seek(0, io.SeekCurrent)
	if pos != 5 || err != nil {
		t.Errorf("offset after WriteAt = %d, %s, want 5, nil", pos, errText(err))
	}
}

func TestWriteAt(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_writeat.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)
	defer f.Close()

	f.Write([]byte("hello world"))
	_, err = f.WriteAt([]byte("WORLD"), 6)
	if err != nil {
		t.Fatalf("WriteAt: %s", errText(err))
		return
	}

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "hello WORLD" {
		t.Errorf("ReadFile = %s, want hello WORLD", string(b))
	}
}
