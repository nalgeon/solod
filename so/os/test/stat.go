package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestStat_File(t *testing.T) {
	name := "test_stat.txt"
	os.WriteFile(name, []byte("hello"), 0o666)
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat failed")
		return
	}
	if fi.Name() != "test_stat.txt" {
		t.Error("Stat: wrong name")
	}
	if fi.Size() != 5 {
		t.Error("Stat: wrong size")
	}
	if !fi.Mode().IsRegular() {
		t.Error("Stat: not regular")
	}
	if fi.IsDir() {
		t.Error("Stat: should not be dir")
	}
	os.Remove(name)
}

func TestStat_Dir(t *testing.T) {
	name := "test_stat_dir"
	os.Mkdir(name, 0o755)
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat dir failed")
		return
	}
	if fi.Name() != "test_stat_dir" {
		t.Error("Stat dir: wrong name")
	}
	if !fi.IsDir() {
		t.Error("Stat dir: should be dir")
	}
	if fi.Mode().IsRegular() {
		t.Error("Stat dir: should not be regular")
	}
	os.Remove(name)
}

func TestLstat_Symlink(t *testing.T) {
	target := "test_lstat_target.txt"
	link := "test_lstat_link"
	os.WriteFile(target, []byte("target"), 0o666)
	os.Symlink(target, link)

	// Lstat returns info about the link itself.
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal("Lstat failed")
		return
	}
	if fi.Name() != "test_lstat_link" {
		t.Error("Lstat: wrong name")
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("Lstat: should be symlink")
	}

	// Stat follows the link.
	fi2, err := os.Stat(link)
	if err != nil {
		t.Fatal("Stat through link failed")
		return
	}
	if fi2.Size() != 6 {
		t.Error("Stat through link: wrong size")
	}
	if fi2.Mode()&os.ModeSymlink != 0 {
		t.Error("Stat through link: should not be symlink")
	}

	os.Remove(link)
	os.Remove(target)
}

func TestSameFile(t *testing.T) {
	name := "test_samefile.txt"
	os.WriteFile(name, []byte("same"), 0o666)

	fi1, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat 1 failed")
		return
	}
	fi2, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat 2 failed")
		return
	}
	if !os.SameFile(fi1, fi2) {
		t.Error("SameFile: should be same")
	}

	name2 := "test_samefile2.txt"
	os.WriteFile(name2, []byte("other"), 0o666)

	fi3, err := os.Stat(name2)
	if err != nil {
		t.Fatal("Stat 3 failed")
		return
	}
	if os.SameFile(fi1, fi3) {
		t.Error("SameFile: should be different")
	}

	os.Remove(name2)
	os.Remove(name)
}

func TestStat_NotExist(t *testing.T) {
	_, err := os.Stat("nonexistent_stat.txt")
	if err != os.ErrNotExist {
		t.Error("Stat nonexistent: wrong error")
	}
}

func TestChmod(t *testing.T) {
	name := "test_chmod.txt"
	os.WriteFile(name, []byte("chmod"), 0o666)
	err := os.Chmod(name, 0o644)
	if err != nil {
		t.Fatal("Chmod failed")
		return
	}
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatal("Stat after Chmod failed")
		return
	}
	if fi.Mode().Perm() != 0o644 {
		t.Error("Chmod: wrong perm")
	}
	os.Remove(name)
}
