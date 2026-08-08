package main

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestOpenFile_Create(t *testing.T) {
	name := "test_openfile.txt"
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatal("OpenFile create failed")
		return
	}
	defer os.Remove(name)
	f.Write([]byte("openfile"))
	f.Close()

	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile after OpenFile failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "openfile" {
		t.Error("OpenFile: wrong data")
	}
}

func TestOpenFile_RdOnly(t *testing.T) {
	name := "test_openfile_rd.txt"
	os.WriteFile(name, []byte("readonly"), 0o666)
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		t.Fatal("OpenFile rdonly failed")
		return
	}
	defer f.Close()

	buf := make([]byte, 16)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatal("Read from rdonly failed")
		return
	}
	if string(buf[:n]) != "readonly" {
		t.Error("OpenFile rdonly: wrong data")
	}
}

func TestFile_Name(t *testing.T) {
	name := "test_filename.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)

	if f.Name() != name {
		t.Error("Name: wrong")
	}
	f.Close()
}

func TestLink(t *testing.T) {
	target := "test_link_target.txt"
	os.WriteFile(target, []byte("linked"), 0o666)
	defer os.Remove(target)

	// Hard link.
	hard := "test_hard_link.txt"
	err := os.Link(target, hard)
	if err != nil {
		t.Fatal("Link failed")
		return
	}
	defer os.Remove(hard)

	b, err := os.ReadFile(mem.System, hard)
	if err != nil {
		t.Fatal("ReadFile hard link failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "linked" {
		t.Error("Hard link: wrong data")
	}
}

func TestSymlink(t *testing.T) {
	target := "test_sym_target.txt"
	os.WriteFile(target, []byte("sym"), 0o666)
	defer os.Remove(target)

	link := "test_sym_link"
	err := os.Symlink(target, link)
	if err != nil {
		t.Fatal("Symlink failed")
		return
	}
	defer os.Remove(link)

	var rlBuf [os.MaxPathLen]byte
	dest, err := os.Readlink(rlBuf[:], link)
	if err != nil {
		t.Fatal("Readlink failed")
		return
	}
	if dest != target {
		t.Error("Readlink: wrong target")
	}
}

func TestMkdirChdir(t *testing.T) {
	dir := "test_mkdir_dir"
	err := os.Mkdir(dir, 0o755)
	if err != nil {
		t.Fatal("Mkdir failed")
		return
	}
	defer os.Remove(dir)

	// Get current dir.
	var wdBuf [os.MaxPathLen]byte
	origWd, err := os.Getwd(wdBuf[:])
	if err != nil {
		t.Fatal("Getwd failed")
		return
	}

	// Change to new dir.
	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Chdir failed")
		return
	}
	defer os.Chdir(origWd) // Change back at the end.

	// Verify we moved.
	var wdBuf2 [os.MaxPathLen]byte
	newWd, err := os.Getwd(wdBuf2[:])
	if err != nil {
		t.Fatal("Getwd after Chdir failed")
		return
	}
	if newWd == origWd {
		t.Error("Chdir: dir did not change")
	}
}

func TestTruncate(t *testing.T) {
	name := "test_truncate.txt"
	os.WriteFile(name, []byte("abcdef"), 0o666)
	defer os.Remove(name)

	err := os.Truncate(name, 3)
	if err != nil {
		t.Fatal("Truncate failed")
		return
	}
	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile after Truncate failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "abc" {
		t.Error("Truncate: wrong data")
	}
}

func TestOpenFile_Append(t *testing.T) {
	name := "test_append.txt"
	os.WriteFile(name, []byte("hello"), 0o666)
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal("OpenFile append failed")
		return
	}
	f.Write([]byte(" world"))
	f.Close()

	b, err := os.ReadFile(mem.System, name)
	if err != nil {
		t.Fatal("ReadFile after append failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "hello world" {
		t.Error("Append: wrong data")
	}
}

func TestChtimes(t *testing.T) {
	// Chtimes - just verify it doesn't error.
	name := "test_chtimes.txt"
	os.WriteFile(name, []byte("times"), 0o666)
	defer os.Remove(name)

	fi, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat for Chtimes failed")
		return
	}
	mt := fi.ModTime()
	err = os.Chtimes(name, mt, mt)
	if err != nil {
		t.Error("Chtimes failed")
	}
}

func TestChown(t *testing.T) {
	// Chown with -1, -1 (no change) - should succeed.
	name := "test_chown.txt"
	os.WriteFile(name, []byte("chown"), 0o666)
	defer os.Remove(name)

	err := os.Chown(name, -1, -1)
	if err != nil {
		t.Error("Chown failed")
	}
}

func TestLchown(t *testing.T) {
	// Lchown with -1, -1 (no change) - should succeed.
	name := "test_lchown.txt"
	os.WriteFile(name, []byte("lchown"), 0o666)
	defer os.Remove(name)

	err := os.Lchown(name, -1, -1)
	if err != nil {
		t.Error("Lchown failed")
	}
}

func TestRemove(t *testing.T) {
	name := "test_remove.txt"
	err := os.WriteFile(name, []byte("tmp"), 0o666)
	if err != nil {
		t.Fatal("WriteFile failed")
		return
	}

	err = os.Remove(name)
	if err != nil {
		t.Fatal("Remove failed")
		return
	}

	_, err = os.Open(name)
	if err == nil {
		t.Error("Open after Remove should fail")
	}
}

func TestRename(t *testing.T) {
	oldName := "test_old.txt"
	newName := "test_new.txt"
	os.WriteFile(oldName, []byte("renamed"), 0o666)
	err := os.Rename(oldName, newName)
	if err != nil {
		t.Fatal("Rename failed")
		return
	}
	defer os.Remove(newName)

	b, err := os.ReadFile(mem.System, newName)
	if err != nil {
		t.Fatal("ReadFile after Rename failed")
		return
	}
	defer mem.FreeSlice(mem.System, b)

	if string(b) != "renamed" {
		t.Error("Rename: wrong data")
	}
}

func TestMkdir_ErrExist(t *testing.T) {
	// ErrExist - try to create dir that already exists.
	name := "test_exist_dir"
	os.Mkdir(name, 0o755)
	err := os.Mkdir(name, 0o755)
	if err != os.ErrExist {
		t.Error("Mkdir existing: wrong error")
	}
	os.Remove(name)
}

func TestOpen_ErrNotExist(t *testing.T) {
	_, err := os.Open("nonexistent_file.txt")
	if err != os.ErrNotExist {
		t.Error("Open nonexistent: wrong error")
	}
}

func TestOpenFile_ErrNotExist(t *testing.T) {
	_, err := os.OpenFile("nonexistent_open.txt", os.O_RDONLY, 0)
	if err != os.ErrNotExist {
		t.Error("OpenFile nonexistent: wrong error")
	}
}

func TestFile_NilInvalid(t *testing.T) {
	var f *os.File
	buf := make([]byte, 8)

	_, err := f.Read(buf)
	if err != os.ErrInvalid {
		t.Error("Read on nil file: wrong error")
	}
	_, err = f.Write(buf)
	if err != os.ErrInvalid {
		t.Error("Write on nil file: wrong error")
	}
	_, err = f.WriteString("data")
	if err != os.ErrInvalid {
		t.Error("WriteString on nil file: wrong error")
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrInvalid {
		t.Error("Seek on nil file: wrong error")
	}
	_, err = f.ReadAt(buf, 0)
	if err != os.ErrInvalid {
		t.Error("ReadAt on nil file: wrong error")
	}
	_, err = f.WriteAt(buf, 0)
	if err != os.ErrInvalid {
		t.Error("WriteAt on nil file: wrong error")
	}
	if f.Close() != os.ErrInvalid {
		t.Error("Close on nil file: wrong error")
	}
}

func TestFile_ClosedInvalid(t *testing.T) {
	name := "test_closed_file.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatal("Create failed")
		return
	}
	defer os.Remove(name)
	if f.Close() != nil {
		t.Fatal("Close failed")
		return
	}
	buf := make([]byte, 8)

	// Every method must report the closed file instead of
	// using the stream that Close destroyed.
	_, err = f.Read(buf)
	if err != os.ErrClosed {
		t.Error("Read on closed file: wrong error")
	}
	_, err = f.Write(buf)
	if err != os.ErrClosed {
		t.Error("Write on closed file: wrong error")
	}
	_, err = f.WriteString("data")
	if err != os.ErrClosed {
		t.Error("WriteString on closed file: wrong error")
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrClosed {
		t.Error("Seek on closed file: wrong error")
	}
	_, err = f.ReadAt(buf, 0)
	if err != os.ErrClosed {
		t.Error("ReadAt on closed file: wrong error")
	}
	_, err = f.WriteAt(buf, 0)
	if err != os.ErrClosed {
		t.Error("WriteAt on closed file: wrong error")
	}
	if f.Close() != os.ErrClosed {
		t.Error("second Close: wrong error")
	}
}

func TestFile_ZeroInvalid(t *testing.T) {
	// Open returns the zero File when it fails.
	// A caller that ignores the error gets an unopened file.
	f, err := os.Open("nonexistent_zero_file.txt")
	if err != os.ErrNotExist {
		t.Fatal("Open nonexistent: wrong error")
		return
	}
	buf := make([]byte, 8)

	_, err = f.Read(buf)
	if err != os.ErrInvalid {
		t.Error("Read on zero file: wrong error")
	}
	_, err = f.Write(buf)
	if err != os.ErrInvalid {
		t.Error("Write on zero file: wrong error")
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrInvalid {
		t.Error("Seek on zero file: wrong error")
	}
	if f.Close() != os.ErrInvalid {
		t.Error("Close on zero file: wrong error")
	}
	if f.Name() != "" {
		t.Error("Name on zero file: want empty")
	}
}
