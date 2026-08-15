package fmt_test

import (
	"solod.dev/so/c"
	"solod.dev/so/fmt"
	"solod.dev/so/runtime"
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

func TestPrintNoArgs(t *testing.T) {
	n, err := fmt.Print()
	if err != nil {
		t.Fatal("Print failed")
		return
	}
	if n != 0 {
		t.Error("Print: wrong count")
	}
	n, err = fmt.Println()
	if err != nil {
		t.Fatal("Println failed")
		return
	}
	if n != 1 {
		t.Error("Println: wrong count")
	}
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
	buf := make([]byte, 32)
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

	// %d takes a So int. A narrower integer needs no conversion, because the
	// print family is nodecay and every scalar widens.
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

func TestFprintfLong(t *testing.T) {
	var sb strings.Builder
	defer sb.Free()

	n, err := fmt.Fprintf(&sb, "%2000d", 42)
	if err != nil {
		t.Fatal("Fprintf failed")
		return
	}
	if n != 2000 {
		t.Error("Fprintf: wrong count")
	}
	if len(sb.String()) != 2000 {
		t.Error("Fprintf: wrong length")
	}
}

func TestSscanf(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("Sscanf needs a hosted environment")
		return
	}
	var n1, n2 int32
	buf := make([]byte, 32)
	ptr := c.PtrAs[c.Char](&buf[0])
	n, err := fmt.Sscanf("5 1 gophers", "%d %d %s", &n1, &n2, ptr)
	if err != nil {
		t.Fatal("Sscanf failed")
		return
	}
	if n != 3 {
		t.Error("Sscanf: wrong count")
	}
	if n1 != 5 || n2 != 1 || c.String(ptr) != "gophers" {
		t.Error("Sscanf: wrong values")
	}
}

func TestFscanf(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("Fscanf needs a hosted environment")
		return
	}
	var n1, n2 int32
	buf := make([]byte, 32)
	ptr := c.PtrAs[c.Char](&buf[0])
	r := strings.NewReader("5 1 gophers")
	n, err := fmt.Fscanf(&r, "%d %d %s", &n1, &n2, ptr)
	if err != nil {
		t.Fatal("Fscanf failed")
		return
	}
	if n != 3 {
		t.Error("Fscanf: wrong count")
	}
	if n1 != 5 || n2 != 1 || c.String(ptr) != "gophers" {
		t.Error("Fscanf: wrong values")
	}
}
