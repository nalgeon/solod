package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestGetpid(t *testing.T) {
	pid := os.Getpid()
	if pid <= 0 {
		t.Errorf("Getpid = %d, want a positive id", pid)
	}
}

func TestGetppid(t *testing.T) {
	ppid := os.Getppid()
	// The init process of a container has no parent, so the id is zero.
	// Every other process has a parent, and the parent of an orphan
	// is the init process.
	if os.Getpid() == 1 {
		if ppid != 0 {
			t.Errorf("Getppid of the init process = %d, want 0", ppid)
		}
		return
	}
	if ppid <= 0 {
		t.Errorf("Getppid = %d, want a positive id", ppid)
	}
}

func TestGetuid(t *testing.T) {
	uid := os.Getuid()
	if uid < 0 {
		t.Errorf("Getuid = %d, want zero or more", uid)
	}
}

func TestGeteuid(t *testing.T) {
	euid := os.Geteuid()
	if euid < 0 {
		t.Errorf("Geteuid = %d, want zero or more", euid)
	}
}

func TestGetgid(t *testing.T) {
	gid := os.Getgid()
	if gid < 0 {
		t.Errorf("Getgid = %d, want zero or more", gid)
	}
}

func TestGetegid(t *testing.T) {
	egid := os.Getegid()
	if egid < 0 {
		t.Errorf("Getegid = %d, want zero or more", egid)
	}
}

func TestGetwd(t *testing.T) {
	var wdBuf [os.MaxPathLen]byte
	wd, err := os.Getwd(wdBuf[:])
	if err != nil {
		t.Fatalf("Getwd: %s", errText(err))
		return
	}
	if len(wd) == 0 {
		t.Fatal("Getwd: empty")
		return
	}
	// Should start with '/'.
	if wd[0] != '/' {
		t.Errorf("Getwd = %s, want an absolute path", wd)
	}
}

func TestGetwd_ShortBuf(t *testing.T) {
	// The shortest path is "/", which needs a byte for the null terminator.
	var wdBuf [1]byte
	wd, err := os.Getwd(wdBuf[:])
	if err == nil {
		t.Errorf("Getwd into a short buffer = %s, want an error", wd)
	}
}

func TestHostname(t *testing.T) {
	var hostBuf [os.MaxNameLen]byte
	name, err := os.Hostname(hostBuf[:])
	if err != nil {
		t.Fatalf("Hostname: %s", errText(err))
		return
	}
	if len(name) == 0 {
		t.Error("Hostname: empty")
	}
}
