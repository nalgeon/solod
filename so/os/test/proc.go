package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestGetpid(t *testing.T) {
	pid := os.Getpid()
	if pid <= 0 {
		t.Error("Getpid: invalid")
	}
}

func TestGetppid(t *testing.T) {
	ppid := os.Getppid()
	if ppid < 0 {
		t.Error("Getppid: invalid")
	}
}

func TestGetuid(t *testing.T) {
	uid := os.Getuid()
	if uid < 0 {
		t.Error("Getuid: invalid")
	}
}

func TestGeteuid(t *testing.T) {
	euid := os.Geteuid()
	if euid < 0 {
		t.Error("Geteuid: invalid")
	}
}

func TestGetgid(t *testing.T) {
	gid := os.Getgid()
	if gid < 0 {
		t.Error("Getgid: invalid")
	}
}

func TestGetegid(t *testing.T) {
	egid := os.Getegid()
	if egid < 0 {
		t.Error("Getegid: invalid")
	}
}

func TestGetwd(t *testing.T) {
	var wdBuf [os.MaxPathLen]byte
	wd, err := os.Getwd(wdBuf[:])
	if err != nil {
		t.Fatal("Getwd failed")
		return
	}
	if len(wd) == 0 {
		t.Fatal("Getwd: empty")
		return
	}
	// Should start with '/'.
	if wd[0] != '/' {
		t.Error("Getwd: not absolute")
	}
}

func TestHostname(t *testing.T) {
	var hostBuf [os.MaxNameLen]byte
	name, err := os.Hostname(hostBuf[:])
	if err != nil {
		t.Fatal("Hostname failed")
		return
	}
	if len(name) == 0 {
		t.Error("Hostname: empty")
	}
}
