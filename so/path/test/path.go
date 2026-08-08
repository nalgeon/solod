package path_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/path"
	"solod.dev/so/testing"
)

func TestClean(t *testing.T) {
	cleaned := path.Clean(mem.System, "/opt/app/../config.json")
	defer mem.FreeString(mem.System, cleaned)
	if cleaned != "/opt/config.json" {
		t.Error("unexpected cleaned path: " + cleaned)
	}
}

func TestSplit(t *testing.T) {
	dir, file := path.Split("/opt/app/config.json")
	if dir != "/opt/app/" {
		t.Error("unexpected dir: " + dir)
	}
	if file != "config.json" {
		t.Error("unexpected file: " + file)
	}
}

func TestJoin(t *testing.T) {
	joined := path.Join(mem.System, "opt", "app", "config.json")
	defer mem.FreeString(mem.System, joined)
	if joined != "opt/app/config.json" {
		t.Error("unexpected path: " + joined)
	}
}

func TestIsAbs(t *testing.T) {
	if !path.IsAbs("/opt/app/config.json") {
		t.Error("want absolute")
	}
	if path.IsAbs("opt/app/config.json") {
		t.Error("want not absolute")
	}
}

func TestDir(t *testing.T) {
	dir := path.Dir(mem.System, "/opt/app/config.json")
	defer mem.FreeString(mem.System, dir)
	if dir != "/opt/app" {
		t.Error("unexpected dir: " + dir)
	}
}

func TestBase(t *testing.T) {
	base := path.Base("/opt/app/config.json")
	if base != "config.json" {
		t.Error("unexpected base: " + base)
	}
}

func TestExt(t *testing.T) {
	ext := path.Ext("/opt/app/config.json")
	if ext != ".json" {
		t.Error("unexpected ext: " + ext)
	}
}

func TestMatch(t *testing.T) {
	ok, err := path.Match("/opt/*/*.js?n", "/opt/app/config.json")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !ok {
		t.Error("want match")
	}
}
