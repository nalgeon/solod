package os_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/os"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestSetGetenv(t *testing.T) {
	err := os.Setenv("SO_TEST_KEY", "test_value")
	if err != nil {
		t.Fatalf("Setenv: %s", errText(err))
		return
	}
	val := os.Getenv("SO_TEST_KEY")
	if val != "test_value" {
		t.Errorf("Getenv = %s, want test_value", val)
	}
}

func TestSetenv_Overwrite(t *testing.T) {
	os.Setenv("SO_TEST_OVER", "one")
	err := os.Setenv("SO_TEST_OVER", "two")
	if err != nil {
		t.Fatalf("Setenv: %s", errText(err))
		return
	}
	val := os.Getenv("SO_TEST_OVER")
	if val != "two" {
		t.Errorf("Getenv after the second Setenv = %s, want two", val)
	}
}

func TestSetenv_EmptyKey(t *testing.T) {
	// An empty name is not a variable name.
	if err := os.Setenv("", "value"); err == nil {
		t.Error("Setenv with an empty key: want an error")
	}
}

func TestLookupEnv_Present(t *testing.T) {
	os.Setenv("SO_TEST_LOOKUP", "found")
	val, ok := os.LookupEnv("SO_TEST_LOOKUP")
	if !ok {
		t.Fatal("LookupEnv: should be present")
		return
	}
	if val != "found" {
		t.Errorf("LookupEnv = %s, want found", val)
	}
}

func TestLookupEnv_Empty(t *testing.T) {
	// A variable set to an empty value is present.
	// Only the boolean tells it apart from a missing variable.
	os.Setenv("SO_TEST_EMPTY", "")
	val, ok := os.LookupEnv("SO_TEST_EMPTY")
	if !ok {
		t.Error("LookupEnv of an empty value: should be present")
	}
	if val != "" {
		t.Errorf("LookupEnv = %s, want an empty value", val)
	}
	if os.Getenv("SO_TEST_EMPTY") != "" {
		t.Error("Getenv of an empty value: want an empty value")
	}
}

func TestLookupEnv_Absent(t *testing.T) {
	_, ok := os.LookupEnv("SO_TEST_NONEXISTENT_VAR_XYZ")
	if ok {
		t.Error("LookupEnv: should not be present")
	}
}

func TestUnsetenv(t *testing.T) {
	os.Setenv("SO_TEST_UNSET", "bye")
	err := os.Unsetenv("SO_TEST_UNSET")
	if err != nil {
		t.Fatalf("Unsetenv: %s", errText(err))
		return
	}
	val := os.Getenv("SO_TEST_UNSET")
	if val != "" {
		t.Errorf("Getenv after Unsetenv = %s, want an empty value", val)
	}
	if _, ok := os.LookupEnv("SO_TEST_UNSET"); ok {
		t.Error("LookupEnv after Unsetenv: should not be present")
	}
}

func TestGetenv_PATH(t *testing.T) {
	// Getenv on PATH (should always be set).
	path := os.Getenv("PATH")
	if len(path) == 0 {
		t.Error("Getenv PATH: empty")
	}
}

func TestTempDir_Env(t *testing.T) {
	alloc := t.Allocator()

	// Getenv gives a view into the environment, and Setenv can replace it,
	// so the old value needs a copy of its own.
	old, had := os.LookupEnv("TMPDIR")
	saved := strings.Clone(alloc, old)
	defer mem.FreeString(alloc, saved)

	// TempDir gives $TMPDIR when it is set.
	os.Setenv("TMPDIR", "/so-test-tmpdir")
	if got := os.TempDir(); got != "/so-test-tmpdir" {
		t.Errorf("TempDir with TMPDIR = %s, want /so-test-tmpdir", got)
	}

	// TempDir gives /tmp when TMPDIR is empty or unset.
	os.Setenv("TMPDIR", "")
	if got := os.TempDir(); got != "/tmp" {
		t.Errorf("TempDir with an empty TMPDIR = %s, want /tmp", got)
	}
	os.Unsetenv("TMPDIR")
	if got := os.TempDir(); got != "/tmp" {
		t.Errorf("TempDir without TMPDIR = %s, want /tmp", got)
	}

	if had {
		os.Setenv("TMPDIR", saved)
	}
}
