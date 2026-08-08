package os_test

import (
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestSetGetenv(t *testing.T) {
	err := os.Setenv("SO_TEST_KEY", "test_value")
	if err != nil {
		t.Fatal("Setenv failed")
		return
	}
	val := os.Getenv("SO_TEST_KEY")
	if val != "test_value" {
		t.Error("Getenv: wrong value")
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
		t.Error("LookupEnv: wrong value")
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
		t.Fatal("Unsetenv failed")
		return
	}
	val := os.Getenv("SO_TEST_UNSET")
	if val != "" {
		t.Error("Unsetenv: should be empty")
	}
}

func TestGetenv_PATH(t *testing.T) {
	// Getenv on PATH (should always be set).
	path := os.Getenv("PATH")
	if len(path) == 0 {
		t.Error("Getenv PATH: empty")
	}
}
