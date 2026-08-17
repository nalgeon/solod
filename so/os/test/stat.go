package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

func TestStat_File(t *testing.T) {
	name := os.TempDir() + "/so_os_stat.txt"
	if err := os.WriteFile(name, []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Name() != "so_os_stat.txt" {
		t.Errorf("Name = %s, want so_os_stat.txt", fi.Name())
	}
	if fi.Size() != 5 {
		t.Errorf("Size = %d, want 5", fi.Size())
	}
	if !fi.Mode().IsRegular() {
		t.Error("Mode: want a regular file")
	}
	if fi.IsDir() {
		t.Error("IsDir: want false")
	}
}

func TestStat_Dir(t *testing.T) {
	name := os.TempDir() + "/so_os_stat_dir"
	if err := os.Mkdir(name, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(name)

	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Name() != "so_os_stat_dir" {
		t.Errorf("Name = %s, want so_os_stat_dir", fi.Name())
	}
	if !fi.IsDir() {
		t.Error("IsDir: want true")
	}
	if fi.Mode().IsRegular() {
		t.Error("Mode: want a directory")
	}
	if fi.Mode().Perm() != 0o700 {
		t.Errorf("Perm = %o, want 700", int(fi.Mode().Perm()))
	}
}

func TestLstat_Symlink(t *testing.T) {
	target := os.TempDir() + "/so_os_lstat_target.txt"
	link := os.TempDir() + "/so_os_lstat_link"
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(target)
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink: %s", errText(err))
		return
	}
	defer os.Remove(link)

	// Lstat returns info about the link itself.
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Lstat: %s", errText(err))
		return
	}
	if fi.Name() != "so_os_lstat_link" {
		t.Errorf("Name = %s, want so_os_lstat_link", fi.Name())
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("Lstat mode: want ModeSymlink")
	}

	// Stat follows the link.
	fi2, err := os.Stat(link)
	if err != nil {
		t.Fatalf("Stat through the link: %s", errText(err))
		return
	}
	if fi2.Size() != 6 {
		t.Errorf("Size through the link = %d, want 6", fi2.Size())
	}
	if fi2.Mode()&os.ModeSymlink != 0 {
		t.Error("Stat mode: want no ModeSymlink")
	}
}

func TestSameFile(t *testing.T) {
	name := os.TempDir() + "/so_os_samefile.txt"
	if err := os.WriteFile(name, []byte("same"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	fi1, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	fi2, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat again: %s", errText(err))
		return
	}
	if !os.SameFile(fi1, fi2) {
		t.Error("SameFile of one file: want true")
	}

	name2 := os.TempDir() + "/so_os_samefile2.txt"
	if err := os.WriteFile(name2, []byte("other"), 0o600); err != nil {
		t.Fatalf("WriteFile 2: %s", errText(err))
		return
	}
	defer os.Remove(name2)

	fi3, err := os.Stat(name2)
	if err != nil {
		t.Fatalf("Stat 2: %s", errText(err))
		return
	}
	if os.SameFile(fi1, fi3) {
		t.Error("SameFile of two files: want false")
	}
}

func TestStat_NotExist(t *testing.T) {
	_, err := os.Stat(os.TempDir() + "/so_os_no_such_stat.txt")
	if err != os.ErrNotExist {
		t.Errorf("Stat of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestChmod(t *testing.T) {
	name := os.TempDir() + "/so_os_chmod.txt"
	if err := os.WriteFile(name, []byte("chmod"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	if err := os.Chmod(name, 0o644); err != nil {
		t.Fatalf("Chmod: %s", errText(err))
		return
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode().Perm() != 0o644 {
		t.Errorf("Perm = %o, want 644", int(fi.Mode().Perm()))
	}
}

func TestChmod_Sticky(t *testing.T) {
	name := os.TempDir() + "/so_os_chmod_sticky"
	if err := os.Mkdir(name, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// The sticky bit goes to the directory and comes back through Stat.
	if err := os.Chmod(name, 0o700|os.ModeSticky); err != nil {
		t.Fatalf("Chmod: %s", errText(err))
		return
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.Mode()&os.ModeSticky == 0 {
		t.Error("Mode: want ModeSticky")
	}
	if !fi.Mode().IsDir() {
		t.Error("Mode: want ModeDir")
	}
	if fi.Mode().Perm() != 0o700 {
		t.Errorf("Perm = %o, want 700", int(fi.Mode().Perm()))
	}
}

// The setuid and setgid bits have no test here. A system can drop them
// without an error from Chmod, so only the mode conversion is testable.
// See the Go unit tests in so/os/fmode_test.go.

func TestChmod_NotExist(t *testing.T) {
	err := os.Chmod(os.TempDir()+"/so_os_no_such_chmod.txt", 0o600)
	if err != os.ErrNotExist {
		t.Errorf("Chmod of a missing file = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestChtimes(t *testing.T) {
	name := os.TempDir() + "/so_os_chtimes.txt"
	if err := os.WriteFile(name, []byte("times"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	// 2001-09-09 01:46:40 UTC and a day later.
	mtime := time.Unix(1000000000, 0)
	atime := time.Unix(1000086400, 0)
	if err := os.Chtimes(name, atime, mtime); err != nil {
		t.Fatalf("Chtimes: %s", errText(err))
		return
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.ModTime().Unix() != 1000000000 {
		t.Errorf("ModTime = %d, want 1000000000", fi.ModTime().Unix())
	}
}

func TestChtimes_Zero(t *testing.T) {
	name := os.TempDir() + "/so_os_chtimes_zero.txt"
	if err := os.WriteFile(name, []byte("times"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	mtime := time.Unix(1000000000, 0)
	if err := os.Chtimes(name, mtime, mtime); err != nil {
		t.Fatalf("Chtimes: %s", errText(err))
		return
	}

	// A zero time leaves the file time as it is.
	var zero time.Time
	if err := os.Chtimes(name, zero, zero); err != nil {
		t.Fatalf("Chtimes with a zero time: %s", errText(err))
		return
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatalf("Stat: %s", errText(err))
		return
	}
	if fi.ModTime().Unix() != 1000000000 {
		t.Errorf("ModTime = %d, want 1000000000", fi.ModTime().Unix())
	}
}
