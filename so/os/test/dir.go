package os_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestReadDir(t *testing.T) {
	// ReadDir on a directory with known contents.
	dirName := "test_readdir"

	os.Mkdir(dirName, 0o755)
	defer os.Remove(dirName)
	os.WriteFile(dirName+"/aaa.txt", []byte("hello"), 0o666)
	defer os.Remove(dirName + "/aaa.txt")
	os.WriteFile(dirName+"/bbb.txt", []byte("world"), 0o666)
	defer os.Remove(dirName + "/bbb.txt")
	os.Mkdir(dirName+"/subdir", 0o755)
	defer os.Remove(dirName + "/subdir")

	entries, err := os.ReadDir(mem.System, dirName)
	if err != nil {
		t.Fatal("ReadDir failed")
		return
	}
	defer os.FreeDirEntry(mem.System, entries)

	if len(entries) != 3 {
		t.Fatal("ReadDir: wrong count")
		return
	}

	entry := entries[0]
	if entry.Name != "aaa.txt" || entry.IsDir {
		t.Error("ReadDir: want 1st = aaa.txt")
	}
	entry = entries[1]
	if entry.Name != "bbb.txt" || entry.IsDir {
		t.Error("ReadDir: want 2nd = bbb.txt")
	}
	entry = entries[2]
	if entry.Name != "subdir" || !entry.IsDir {
		t.Error("ReadDir: want 3rd = subdir")
	}
	if entry.Type&os.ModeDir == 0 {
		t.Error("ReadDir: subdir should have ModeDir")
	}
}

func TestReadDir_NotExist(t *testing.T) {
	// ReadDir on nonexistent directory.
	_, err := os.ReadDir(mem.System, "nonexistent_dir_xyz")
	if err != os.ErrNotExist {
		t.Error("ReadDir nonexistent: wrong error")
	}
}
