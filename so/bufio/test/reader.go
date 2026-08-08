package bufio_test

import (
	"solod.dev/so/bufio"
	"solod.dev/so/bytes"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

func TestReader_ReadString(t *testing.T) {
	var buf bytes.Buffer
	buf = bytes.NewBuffer(mem.System, nil)
	defer buf.Free()
	
	w := bufio.NewWriter(mem.System, &buf)
	w.WriteString("Hello, ")
	w.WriteString("World!")
	w.WriteByte('\n')
	w.Flush()
	w.Free()

	sr := strings.NewReader(buf.String())
	r := bufio.NewReader(mem.System, &sr)
	defer r.Free()

	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatal("ReadString failed")
		return
	}
	if line != "Hello, World!\n" {
		t.Error("ReadString = " + line + ", want Hello, World!")
	}
	mem.FreeString(nil, line)
}

func TestReader_ReadByte(t *testing.T) {
	sr := strings.NewReader("abc")
	r := bufio.NewReader(mem.System, &sr)
	defer r.Free()

	b, err := r.ReadByte()
	if err != nil || b != 'a' {
		t.Fatal("ReadByte failed")
		return
	}
	if err := r.UnreadByte(); err != nil {
		t.Fatal("UnreadByte failed")
		return
	}
	b, err = r.ReadByte()
	if err != nil || b != 'a' {
		t.Fatal("UnreadByte re-read failed")
		return
	}
}

func TestReader_Peek(t *testing.T) {
	sr := strings.NewReader("hello")
	r := bufio.NewReader(mem.System, &sr)
	defer r.Free()

	p, err := r.Peek(3)
	if err != nil {
		t.Fatal("Peek failed")
		return
	}
	if string(p) != "hel" {
		t.Error("Peek = " + string(p) + ", want hel")
	}
}
