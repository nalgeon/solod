package os_test

import (
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestOpenFile_Create(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_openfile.txt"
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer os.Remove(name)
	f.Write([]byte("openfile"))
	f.Close()

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "openfile" {
		t.Errorf("ReadFile = %s, want openfile", string(b))
	}
}

func TestOpenFile_RdOnly(t *testing.T) {
	name := os.TempDir() + "/so_os_openfile_rd.txt"
	if err := os.WriteFile(name, []byte("readonly"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer f.Close()

	buf := make([]byte, 16)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read: %s", errText(err))
		return
	}
	if string(buf[:n]) != "readonly" {
		t.Errorf("Read = %s, want readonly", string(buf[:n]))
	}
}

func TestOpenFile_RdWr(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_openfile_rdwr.txt"
	if err := os.WriteFile(name, []byte("abcdef"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// O_RDWR keeps the content and reads and writes it in place.
	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	buf := make([]byte, 3)
	if _, err := f.Read(buf); err != nil {
		t.Fatalf("Read: %s", errText(err))
		return
	}
	if string(buf) != "abc" {
		t.Errorf("Read = %s, want abc", string(buf))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek: %s", errText(err))
		return
	}
	if _, err := f.Write([]byte("XYZ")); err != nil {
		t.Fatalf("Write: %s", errText(err))
		return
	}
	f.Close()

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "XYZdef" {
		t.Errorf("ReadFile = %s, want XYZdef", string(b))
	}
}

func TestOpenFile_RdWrAppend(t *testing.T) {
	name := os.TempDir() + "/so_os_openfile_rdwr_append.txt"
	if err := os.WriteFile(name, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// O_RDWR with O_APPEND writes at the end and reads from anywhere.
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer f.Close()

	if _, err := f.Write([]byte("def")); err != nil {
		t.Fatalf("Write: %s", errText(err))
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek: %s", errText(err))
		return
	}
	buf := make([]byte, 6)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read: %s", errText(err))
		return
	}
	if string(buf[:n]) != "abcdef" {
		t.Errorf("Read = %s, want abcdef", string(buf[:n]))
	}
}

func TestOpenFile_Append(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_append.txt"
	if err := os.WriteFile(name, []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	f.Write([]byte(" world"))
	f.Close()

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "hello world" {
		t.Errorf("ReadFile = %s, want hello world", string(b))
	}
}

func TestOpenFile_Trunc(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_openfile_trunc.txt"
	if err := os.WriteFile(name, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// O_TRUNC drops the content of the file.
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	f.Write([]byte("hi"))
	f.Close()

	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "hi" {
		t.Errorf("ReadFile = %s, want hi", string(b))
	}
}

func TestOpenFile_Excl(t *testing.T) {
	name := os.TempDir() + "/so_os_openfile_excl.txt"
	flag := os.O_CREATE | os.O_EXCL | os.O_WRONLY

	// O_EXCL creates the file.
	f, err := os.OpenFile(name, flag, 0o600)
	if err != nil {
		t.Fatalf("OpenFile: %s", errText(err))
		return
	}
	defer os.Remove(name)
	f.Close()

	// O_EXCL fails when the file is already there.
	_, err = os.OpenFile(name, flag, 0o600)
	if err != os.ErrExist {
		t.Errorf("OpenFile with O_EXCL = %s, want %s", errText(err), os.ErrExist.Error())
	}
}

func TestFile_Name(t *testing.T) {
	name := os.TempDir() + "/so_os_filename.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if f.Name() != name {
		t.Errorf("Name = %s, want %s", f.Name(), name)
	}
	f.Close()
}

func TestLink(t *testing.T) {
	alloc := t.Allocator()
	target := os.TempDir() + "/so_os_link_target.txt"
	if err := os.WriteFile(target, []byte("linked"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(target)

	hard := os.TempDir() + "/so_os_hard_link.txt"
	if err := os.Link(target, hard); err != nil {
		t.Fatalf("Link: %s", errText(err))
		return
	}
	defer os.Remove(hard)

	// The link and the target are one file, not a copy.
	fi1, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat target: %s", errText(err))
		return
	}
	fi2, err := os.Stat(hard)
	if err != nil {
		t.Fatalf("Stat link: %s", errText(err))
		return
	}
	if !os.SameFile(fi1, fi2) {
		t.Error("SameFile of a hard link: want true")
	}

	b, err := os.ReadFile(alloc, hard)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "linked" {
		t.Errorf("ReadFile = %s, want linked", string(b))
	}
}

func TestLink_ErrExist(t *testing.T) {
	target := os.TempDir() + "/so_os_link_exist_target.txt"
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(target)

	link := os.TempDir() + "/so_os_link_exist.txt"
	if err := os.WriteFile(link, []byte("in the way"), 0o600); err != nil {
		t.Fatalf("WriteFile 2: %s", errText(err))
		return
	}
	defer os.Remove(link)

	if err := os.Link(target, link); err != os.ErrExist {
		t.Errorf("Link onto a file = %s, want %s", errText(err), os.ErrExist.Error())
	}
}

func TestSymlink(t *testing.T) {
	target := os.TempDir() + "/so_os_sym_target.txt"
	if err := os.WriteFile(target, []byte("sym"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(target)

	link := os.TempDir() + "/so_os_sym_link"
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink: %s", errText(err))
		return
	}
	defer os.Remove(link)

	var rlBuf [os.MaxPathLen]byte
	dest, err := os.Readlink(rlBuf[:], link)
	if err != nil {
		t.Fatalf("Readlink: %s", errText(err))
		return
	}
	if dest != target {
		t.Errorf("Readlink = %s, want %s", dest, target)
	}
}

func TestSymlink_Dangling(t *testing.T) {
	// A symbolic link to a missing file is still a link.
	link := os.TempDir() + "/so_os_sym_dangling"
	if err := os.Symlink("so_os_no_such_target.txt", link); err != nil {
		t.Fatalf("Symlink: %s", errText(err))
		return
	}
	defer os.Remove(link)

	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Lstat: %s", errText(err))
		return
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("Lstat mode: want ModeSymlink")
	}

	// Open follows the link and finds nothing.
	if _, err := os.Open(link); err != os.ErrNotExist {
		t.Errorf("Open of a dangling link = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
	if _, err := os.Stat(link); err != os.ErrNotExist {
		t.Errorf("Stat of a dangling link = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestReadlink_ShortBuf(t *testing.T) {
	link := os.TempDir() + "/so_os_readlink_short"
	if err := os.Symlink("abcdefgh.txt", link); err != nil {
		t.Fatalf("Symlink: %s", errText(err))
		return
	}
	defer os.Remove(link)

	// A short buffer gets the start of the destination, and no error.
	var buf [4]byte
	dest, err := os.Readlink(buf[:], link)
	if err != nil {
		t.Fatalf("Readlink: %s", errText(err))
		return
	}
	if dest != "abcd" {
		t.Errorf("Readlink into a 4-byte buffer = %s, want abcd", dest)
	}
}

func TestReadlink_NotLink(t *testing.T) {
	name := os.TempDir() + "/so_os_readlink_file.txt"
	if err := os.WriteFile(name, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	var buf [os.MaxPathLen]byte
	if _, err := os.Readlink(buf[:], name); err == nil {
		t.Error("Readlink of a regular file: want an error")
	}
}

func TestMkdir_Perm(t *testing.T) {
	dir := os.TempDir() + "/so_os_mkdir_perm"
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dir)

	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode().Perm() != 0o700 {
		t.Errorf("Perm = %o, want 700", int(fi.Mode().Perm()))
	}
}

func TestMkdirChdir(t *testing.T) {
	dir := os.TempDir() + "/so_os_mkdir_dir"
	err := os.Mkdir(dir, 0o700)
	if err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dir)

	// Get current dir.
	var wdBuf [os.MaxPathLen]byte
	origWd, err := os.Getwd(wdBuf[:])
	if err != nil {
		t.Fatalf("Getwd: %s", errText(err))
		return
	}

	// Change to new dir.
	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Chdir: %s", errText(err))
		return
	}
	defer os.Chdir(origWd) // Change back at the end.

	// Verify we moved.
	var wdBuf2 [os.MaxPathLen]byte
	newWd, err := os.Getwd(wdBuf2[:])
	if err != nil {
		t.Fatalf("Getwd after Chdir: %s", errText(err))
		return
	}
	if newWd == origWd {
		t.Errorf("Getwd after Chdir = %s, want another directory", newWd)
	}
}

func TestChdir_NotExist(t *testing.T) {
	err := os.Chdir(os.TempDir() + "/so_os_no_such_chdir")
	if err != os.ErrNotExist {
		t.Errorf("Chdir into a missing directory = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestChdir_NotDir(t *testing.T) {
	name := os.TempDir() + "/so_os_chdir_file.txt"
	if err := os.WriteFile(name, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if err := os.Chdir(name); err != os.ErrNotDir {
		t.Errorf("Chdir into a file = %s, want %s", errText(err), os.ErrNotDir.Error())
	}
}

func TestTruncate(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_truncate.txt"
	if err := os.WriteFile(name, []byte("abcdef"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	err := os.Truncate(name, 3)
	if err != nil {
		t.Fatalf("Truncate: %s", errText(err))
		return
	}
	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "abc" {
		t.Errorf("ReadFile = %s, want abc", string(b))
	}
}

func TestTruncate_Grow(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_truncate_grow.txt"
	if err := os.WriteFile(name, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// A size above the length of the file adds zero bytes.
	if err := os.Truncate(name, 6); err != nil {
		t.Fatalf("Truncate: %s", errText(err))
		return
	}
	b, err := os.ReadFile(alloc, name)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if len(b) != 6 {
		t.Fatalf("ReadFile length = %d, want 6", len(b))
		return
	}
	if string(b[:3]) != "abc" {
		t.Errorf("ReadFile = %s, want abc at the start", string(b[:3]))
	}
	if b[3] != 0 || b[4] != 0 || b[5] != 0 {
		t.Error("Truncate: want zero bytes at the end")
	}
}

func TestTruncate_Negative(t *testing.T) {
	name := os.TempDir() + "/so_os_truncate_neg.txt"
	if err := os.WriteFile(name, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if err := os.Truncate(name, -1); err == nil {
		t.Error("Truncate to a negative size: want an error")
	}
}

func TestChown(t *testing.T) {
	// Chown with -1, -1 (no change) - should succeed.
	name := os.TempDir() + "/so_os_chown.txt"
	if err := os.WriteFile(name, []byte("chown"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if err := os.Chown(name, -1, -1); err != nil {
		t.Errorf("Chown: %s", errText(err))
	}
}

func TestLchown(t *testing.T) {
	// Lchown with -1, -1 (no change) - should succeed.
	name := os.TempDir() + "/so_os_lchown.txt"
	if err := os.WriteFile(name, []byte("lchown"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if err := os.Lchown(name, -1, -1); err != nil {
		t.Errorf("Lchown: %s", errText(err))
	}
}

func TestRemove(t *testing.T) {
	name := os.TempDir() + "/so_os_remove.txt"
	err := os.WriteFile(name, []byte("tmp"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}

	err = os.Remove(name)
	if err != nil {
		t.Fatalf("Remove: %s", errText(err))
		return
	}

	_, err = os.Open(name)
	if err != os.ErrNotExist {
		t.Errorf("Open after Remove = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestRemove_NotExist(t *testing.T) {
	err := os.Remove(os.TempDir() + "/so_os_no_such_remove.txt")
	if err != os.ErrNotExist {
		t.Errorf("Remove of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestRemove_NotEmptyDir(t *testing.T) {
	dir := os.TempDir() + "/so_os_remove_dir"
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dir)
	if err := os.WriteFile(dir+"/file.txt", []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(dir + "/file.txt")

	// Remove takes an empty directory only.
	if err := os.Remove(dir); err == nil {
		t.Error("Remove of a directory with a file in it: want an error")
	}
}

func TestRename(t *testing.T) {
	alloc := t.Allocator()
	oldName := os.TempDir() + "/so_os_rename_old.txt"
	newName := os.TempDir() + "/so_os_rename_new.txt"
	if err := os.WriteFile(oldName, []byte("renamed"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	err := os.Rename(oldName, newName)
	if err != nil {
		os.Remove(oldName)
		t.Fatalf("Rename: %s", errText(err))
		return
	}
	defer os.Remove(newName)

	b, err := os.ReadFile(alloc, newName)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "renamed" {
		t.Errorf("ReadFile = %s, want renamed", string(b))
	}
	if _, err := os.Stat(oldName); err != os.ErrNotExist {
		t.Errorf("Stat of the old name = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestRename_Replace(t *testing.T) {
	alloc := t.Allocator()
	src := os.TempDir() + "/so_os_rename_src.txt"
	dst := os.TempDir() + "/so_os_rename_dst.txt"
	if err := os.WriteFile(src, []byte("source"), 0o600); err != nil {
		t.Fatalf("WriteFile src: %s", errText(err))
		return
	}
	if err := os.WriteFile(dst, []byte("destination"), 0o600); err != nil {
		os.Remove(src)
		t.Fatalf("WriteFile dst: %s", errText(err))
		return
	}
	defer os.Remove(dst)

	// Rename replaces a file that is already there.
	if err := os.Rename(src, dst); err != nil {
		os.Remove(src)
		t.Fatalf("Rename: %s", errText(err))
		return
	}

	b, err := os.ReadFile(alloc, dst)
	if err != nil {
		t.Fatalf("ReadFile: %s", errText(err))
		return
	}
	defer mem.FreeSlice(alloc, b)

	if string(b) != "source" {
		t.Errorf("ReadFile = %s, want source", string(b))
	}
}

func TestRename_NotExist(t *testing.T) {
	src := os.TempDir() + "/so_os_no_such_rename.txt"
	dst := os.TempDir() + "/so_os_rename_never.txt"
	if err := os.Rename(src, dst); err != os.ErrNotExist {
		t.Errorf("Rename of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestMkdir_ErrExist(t *testing.T) {
	// ErrExist - try to create dir that already exists.
	name := os.TempDir() + "/so_os_exist_dir"
	if err := os.Mkdir(name, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(name)

	err := os.Mkdir(name, 0o700)
	if err != os.ErrExist {
		t.Errorf("Mkdir of a directory that is there = %s, want %s", errText(err), os.ErrExist.Error())
	}
}

func TestOpen_ErrNotExist(t *testing.T) {
	_, err := os.Open(os.TempDir() + "/so_os_no_such_open.txt")
	if err != os.ErrNotExist {
		t.Errorf("Open of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestOpenFile_ErrNotExist(t *testing.T) {
	_, err := os.OpenFile(os.TempDir()+"/so_os_no_such_openfile.txt", os.O_RDONLY, 0)
	if err != os.ErrNotExist {
		t.Errorf("OpenFile of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestOpen_ErrPermission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("the root user ignores the permission bits")
		return
	}
	name := os.TempDir() + "/so_os_noperm.txt"
	if err := os.WriteFile(name, []byte("secret"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)
	if err := os.Chmod(name, 0); err != nil {
		t.Fatalf("Chmod: %s", errText(err))
		return
	}

	_, err := os.Open(name)
	if err != os.ErrPermission {
		t.Errorf("Open of a file with no permissions = %s, want %s", errText(err), os.ErrPermission.Error())
	}
}

func TestWriteFile_ErrIsDir(t *testing.T) {
	dir := os.TempDir() + "/so_os_isdir"
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dir)

	err := os.WriteFile(dir, []byte("x"), 0o600)
	if err != os.ErrIsDir {
		t.Errorf("WriteFile onto a directory = %s, want %s", errText(err), os.ErrIsDir.Error())
	}
}

func TestFile_NilInvalid(t *testing.T) {
	var f *os.File
	buf := make([]byte, 8)

	_, err := f.Read(buf)
	if err != os.ErrInvalid {
		t.Errorf("Read on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.Write(buf)
	if err != os.ErrInvalid {
		t.Errorf("Write on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.WriteString("data")
	if err != os.ErrInvalid {
		t.Errorf("WriteString on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrInvalid {
		t.Errorf("Seek on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.ReadAt(buf, 0)
	if err != os.ErrInvalid {
		t.Errorf("ReadAt on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.WriteAt(buf, 0)
	if err != os.ErrInvalid {
		t.Errorf("WriteAt on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	if err := f.Sync(); err != os.ErrInvalid {
		t.Errorf("Sync on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	if err := f.Close(); err != os.ErrInvalid {
		t.Errorf("Close on a nil file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
}

func TestFile_ClosedInvalid(t *testing.T) {
	name := os.TempDir() + "/so_os_closed.txt"
	f, err := os.Create(name)
	if err != nil {
		t.Fatalf("Create: %s", errText(err))
		return
	}
	defer os.Remove(name)
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %s", errText(err))
		return
	}
	buf := make([]byte, 8)

	// Every method must report the closed file instead of
	// using the stream that Close destroyed.
	_, err = f.Read(buf)
	if err != os.ErrClosed {
		t.Errorf("Read on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	_, err = f.Write(buf)
	if err != os.ErrClosed {
		t.Errorf("Write on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	_, err = f.WriteString("data")
	if err != os.ErrClosed {
		t.Errorf("WriteString on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrClosed {
		t.Errorf("Seek on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	_, err = f.ReadAt(buf, 0)
	if err != os.ErrClosed {
		t.Errorf("ReadAt on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	_, err = f.WriteAt(buf, 0)
	if err != os.ErrClosed {
		t.Errorf("WriteAt on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	if err := f.Sync(); err != os.ErrClosed {
		t.Errorf("Sync on a closed file = %s, want %s", errText(err), os.ErrClosed.Error())
	}
	if err := f.Close(); err != os.ErrClosed {
		t.Errorf("second Close = %s, want %s", errText(err), os.ErrClosed.Error())
	}
}

func TestFile_ZeroInvalid(t *testing.T) {
	// Open returns the zero File when it fails.
	// A caller that ignores the error gets an unopened file.
	f, err := os.Open(os.TempDir() + "/so_os_no_such_zero.txt")
	if err != os.ErrNotExist {
		t.Fatalf("Open of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
		return
	}
	buf := make([]byte, 8)

	_, err = f.Read(buf)
	if err != os.ErrInvalid {
		t.Errorf("Read on a zero file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.Write(buf)
	if err != os.ErrInvalid {
		t.Errorf("Write on a zero file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != os.ErrInvalid {
		t.Errorf("Seek on a zero file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	if err := f.Sync(); err != os.ErrInvalid {
		t.Errorf("Sync on a zero file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	if err := f.Close(); err != os.ErrInvalid {
		t.Errorf("Close on a zero file = %s, want %s", errText(err), os.ErrInvalid.Error())
	}
	if f.Name() != "" {
		t.Errorf("Name on a zero file = %s, want an empty name", f.Name())
	}
}
