package fmt_test

import (
	"solod.dev/so/fmt"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestPrint(t *testing.T) {
	n, err := fmt.Print("hello", "world")
	if err != nil {
		t.Fatal("Print failed")
		return
	}
	if n != 11 {
		t.Error("Print: wrong count")
	}
	fmt.Print("\n")
}

func TestPrintln(t *testing.T) {
	n, err := fmt.Println("hello", "world")
	if err != nil {
		t.Fatal("Println failed")
		return
	}
	if n != 12 {
		t.Error("Println: wrong count")
	}
}

func TestPrintf(t *testing.T) {
	s := "world"
	d := 42
	n, err := fmt.Printf("s = %s, d = %d\n", s, d)
	if err != nil {
		t.Fatal("Printf failed")
		return
	}
	if n != 18 {
		t.Error("Printf: wrong count")
	}
}

func TestSprintf(t *testing.T) {
	buf := fmt.NewBuffer(32)
	s := "world"
	d := 42
	out := fmt.Sprintf(buf, "s = %s, d = %d", s, d)
	if out != "s = world, d = 42" {
		t.Error("Sprintf: wrong output")
	}
}

func TestFprintf(t *testing.T) {
	var sb strings.Builder
	defer sb.Free()

	var i int32 = 42
	s := "world"
	n, err := fmt.Fprintf(&sb, "hello %d %s", i, s)
	if err != nil {
		t.Fatal("Fprintf failed")
		return
	}
	if n != 14 {
		t.Error("Fprintf: wrong count")
	}
	if sb.String() != "hello 42 world" {
		t.Error("Fprintf: wrong output")
	}
}

func TestSscanf(t *testing.T) {
	var n1, n2 int32
	buf := fmt.NewBuffer(32)
	n, err := fmt.Sscanf("5 1 gophers", "%d %d %s", &n1, &n2, buf.Ptr)
	if err != nil {
		t.Fatal("Sscanf failed")
		return
	}
	if n != 3 {
		t.Error("Sscanf: wrong count")
	}
	if n1 != 5 || n2 != 1 || buf.String() != "gophers" {
		t.Error("Sscanf: wrong values")
	}
}

func TestFscanf(t *testing.T) {
	var n1, n2 int32
	buf := fmt.NewBuffer(32)
	r := strings.NewReader("5 1 gophers")
	n, err := fmt.Fscanf(&r, "%d %d %s", &n1, &n2, buf.Ptr)
	if err != nil {
		t.Fatal("Fscanf failed")
		return
	}
	if n != 3 {
		t.Error("Fscanf: wrong count")
	}
	if n1 != 5 || n2 != 1 || buf.String() != "gophers" {
		t.Error("Fscanf: wrong values")
	}
}
