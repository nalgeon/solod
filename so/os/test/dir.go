package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestReadDir(t *testing.T) {
	alloc := t.Allocator()
	dirName := os.TempDir() + "/so_os_readdir"

	// The entries are created in reverse order,
	// so the test also checks the sorting.
	os.Mkdir(dirName, 0o700)
	defer os.Remove(dirName)
	os.Mkdir(dirName+"/subdir", 0o700)
	defer os.Remove(dirName + "/subdir")
	os.WriteFile(dirName+"/bbb.txt", []byte("world"), 0o600)
	defer os.Remove(dirName + "/bbb.txt")
	os.WriteFile(dirName+"/aaa.txt", []byte("hello"), 0o600)
	defer os.Remove(dirName + "/aaa.txt")

	entries, err := os.ReadDir(alloc, dirName)
	if err != nil {
		t.Fatalf("ReadDir: %s", errText(err))
		return
	}
	defer os.FreeDirEntry(alloc, entries)

	if len(entries) != 3 {
		t.Fatalf("ReadDir = %d entries, want 3", len(entries))
		return
	}

	entry := entries[0]
	if entry.Name != "aaa.txt" || entry.IsDir {
		t.Errorf("1st entry = %s, want aaa.txt", entry.Name)
	}
	entry = entries[1]
	if entry.Name != "bbb.txt" || entry.IsDir {
		t.Errorf("2nd entry = %s, want bbb.txt", entry.Name)
	}
	entry = entries[2]
	if entry.Name != "subdir" || !entry.IsDir {
		t.Errorf("3rd entry = %s, want subdir", entry.Name)
	}
	if entry.Type&os.ModeDir == 0 {
		t.Error("subdir entry: want ModeDir")
	}
}

func TestReadDir_Empty(t *testing.T) {
	alloc := t.Allocator()
	dirName := os.TempDir() + "/so_os_readdir_empty"
	if err := os.Mkdir(dirName, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dirName)

	// A directory with no entries gives an empty result,
	// because ReadDir drops "." and "..".
	entries, err := os.ReadDir(alloc, dirName)
	if err != nil {
		t.Fatalf("ReadDir: %s", errText(err))
		return
	}
	defer os.FreeDirEntry(alloc, entries)

	if len(entries) != 0 {
		t.Errorf("ReadDir = %d entries, want 0", len(entries))
	}
}

func TestReadDir_Symlink(t *testing.T) {
	alloc := t.Allocator()
	dirName := os.TempDir() + "/so_os_readdir_link"
	if err := os.Mkdir(dirName, 0o700); err != nil {
		t.Fatalf("Mkdir: %s", errText(err))
		return
	}
	defer os.Remove(dirName)
	os.WriteFile(dirName+"/target.txt", []byte("target"), 0o600)
	defer os.Remove(dirName + "/target.txt")
	os.Symlink("target.txt", dirName+"/link")
	defer os.Remove(dirName + "/link")

	entries, err := os.ReadDir(alloc, dirName)
	if err != nil {
		t.Fatalf("ReadDir: %s", errText(err))
		return
	}
	defer os.FreeDirEntry(alloc, entries)

	if len(entries) != 2 {
		t.Fatalf("ReadDir = %d entries, want 2", len(entries))
		return
	}
	// The link entry describes the link, not the target.
	link := entries[0]
	if link.Name != "link" {
		t.Fatalf("1st entry = %s, want link", link.Name)
		return
	}
	if link.Type&os.ModeSymlink == 0 {
		t.Error("link entry: want ModeSymlink")
	}
	if link.IsDir {
		t.Error("link entry: want IsDir false")
	}
}

func TestReadDir_NotExist(t *testing.T) {
	alloc := t.Allocator()
	_, err := os.ReadDir(alloc, os.TempDir()+"/so_os_no_such_dir")
	if err != os.ErrNotExist {
		t.Errorf("ReadDir of a missing directory = %s, want %s", errText(err), os.ErrNotExist.Error())
	}
}

func TestReadDir_NotDir(t *testing.T) {
	alloc := t.Allocator()
	name := os.TempDir() + "/so_os_readdir_notdir.txt"
	if err := os.WriteFile(name, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", errText(err))
		return
	}
	defer os.Remove(name)

	_, err := os.ReadDir(alloc, name)
	if err != os.ErrNotDir {
		t.Errorf("ReadDir of a file = %s, want %s", errText(err), os.ErrNotDir.Error())
	}
}
