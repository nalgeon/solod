package fmt_test

import (
	"solod.dev/so/c"
	"solod.dev/so/errors"
	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/runtime"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestPrint(t *testing.T) {
	sb := strings.NewBuilder(t.Allocator())
	defer sb.Free()
	old := captureOutput(&sb)
	defer restoreOutput(old)

	n, err := fmt.Print("hello", "world")
	if err != nil {
		t.Fatal("Print failed")
		return
	}
	if n != 11 {
		t.Error("Print: wrong count")
	}
	check(t, "Print", sb.String(), "hello world")
}

func TestPrintNoArgs(t *testing.T) {
	sb := strings.NewBuilder(t.Allocator())
	defer sb.Free()
	old := captureOutput(&sb)
	defer restoreOutput(old)

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
	check(t, "Print and Println", sb.String(), "\n")
}

func TestPrintln(t *testing.T) {
	sb := strings.NewBuilder(t.Allocator())
	defer sb.Free()
	old := captureOutput(&sb)
	defer restoreOutput(old)

	n, err := fmt.Println("hello", "world")
	if err != nil {
		t.Fatal("Println failed")
		return
	}
	if n != 12 {
		t.Error("Println: wrong count")
	}
	check(t, "Println", sb.String(), "hello world\n")
}

func TestPrintf(t *testing.T) {
	sb := strings.NewBuilder(t.Allocator())
	defer sb.Free()
	old := captureOutput(&sb)
	defer restoreOutput(old)

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
	check(t, "Printf", sb.String(), "s = world, d = 42\n")
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

func TestSprintfSmallBuffer(t *testing.T) {
	// A buffer with no room takes no bytes at all.
	var none []byte
	check(t, "Sprintf into a nil buffer", fmt.Sprintf(none, "%d", 42), "")
	check(t, "Sprintf into an empty buffer", fmt.Sprintf(make([]byte, 0), "%d", 42), "")

	// Sprintf counts bytes, not runes, so a truncation can cut a rune in two.
	buf := make([]byte, 4)
	out := fmt.Sprintf(buf, "%s", "世界")
	if len(out) != 4 {
		t.Error("Sprintf: wrong truncated length")
	}
	if out[:3] != "世" {
		t.Error("Sprintf: wrong truncated output")
	}
}

func TestFprintf(t *testing.T) {
	sb := strings.NewBuilder(t.Allocator())
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
	sb := strings.NewBuilder(t.Allocator())
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

func TestFprintfError(t *testing.T) {
	// A writer that fails in the middle. Fprintf keeps the first error, drops
	// every byte after it, and counts the bytes the writer took.
	w := errWriter{free: 300}
	n, err := fmt.Fprintf(&w, "%2000d", 42)
	if err != errFull {
		t.Error("Fprintf: wrong error for a writer that fails")
	}
	if n != 300 {
		t.Error("Fprintf: wrong count for a writer that fails")
	}

	// A writer that takes nothing.
	full := errWriter{free: 0}
	n, err = fmt.Fprintf(&full, "hello")
	if err != errFull {
		t.Error("Fprintf: wrong error for a full writer")
	}
	if n != 0 {
		t.Error("Fprintf: wrong count for a full writer")
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

func TestSscanfError(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("Sscanf needs a hosted environment")
		return
	}
	var v int32

	// An input with nothing to read gives a negative count and [fmt.ErrScan].
	n, err := fmt.Sscanf("", "%d", &v)
	if err != fmt.ErrScan {
		t.Error("Sscanf: no error for an empty input")
	}
	if n >= 0 {
		t.Error("Sscanf: wrong count for an empty input")
	}

	// An input that does not match the format gives no error and no value.
	n, err = fmt.Sscanf("abc", "%d", &v)
	if err != nil {
		t.Error("Sscanf: error for an input that does not match")
	}
	if n != 0 {
		t.Error("Sscanf: wrong count for an input that does not match")
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

func TestFscanfEmpty(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("Fscanf needs a hosted environment")
		return
	}
	// A reader with no bytes fails the read, so Fscanf reports the read error
	// and scans nothing.
	var v int32
	r := strings.NewReader("")
	n, err := fmt.Fscanf(&r, "%d", &v)
	if err == nil {
		t.Error("Fscanf: no error for an empty reader")
	}
	if n != 0 {
		t.Error("Fscanf: wrong count for an empty reader")
	}
}

// captureOutput points [fmt.Output] at sb and returns the writer it replaces.
func captureOutput(sb *strings.Builder) io.Writer {
	old := fmt.Output
	fmt.Output = sb
	return old
}

// restoreOutput points [fmt.Output] back at w. The test runner keeps the
// writer it started with, so a redirected Output does not hide a test result.
func restoreOutput(w io.Writer) {
	fmt.Output = w
}

// errFull is the error of a writer with no room left.
var errFull = errors.New("fmt_test: writer full")

// errWriter takes free bytes and fails every write after them.
type errWriter struct {
	free int
}

func (w *errWriter) Write(p []byte) (int, error) {
	if w.free <= 0 {
		return 0, errFull
	}
	if len(p) > w.free {
		n := w.free
		w.free = 0
		return n, errFull
	}
	w.free -= len(p)
	return len(p), nil
}
