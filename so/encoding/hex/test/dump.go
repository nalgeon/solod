// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex_test

import (
	"solod.dev/so/encoding/hex"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// dumpSteps are the strides of the byte values of the dump sweep. They cover
// the printable range, the control range, and the high half of the byte range.
var dumpSteps = [4]int{1, 4, 17, 63}

// dumpInput fills buf with the 40 bytes of the golden dump and returns them.
func dumpInput(buf []byte) []byte {
	for i := range 40 {
		buf[i] = byte(i + 30)
	}
	return buf[:40]
}

// goldenDump returns the dump of dumpInput. The value is one literal per line,
// because a package level constant cannot hold a concatenation.
func goldenDump() string {
	return "00000000  1e 1f 20 21 22 23 24 25  26 27 28 29 2a 2b 2c 2d  |.. !\"#$%&'()*+,-|\n" +
		"00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|\n" +
		"00000020  3e 3f 40 41 42 43 44 45                           |>?@ABCDE|\n"
}

func TestDump(t *testing.T) {
	alloc := t.Allocator()
	var in [40]byte
	got := hex.Dump(alloc, dumpInput(in[:]))
	defer mem.FreeString(alloc, got)
	if got != goldenDump() {
		t.Errorf("Dump() = \n%s\nwant\n%s", got, goldenDump())
	}
}

func TestDumpEmpty(t *testing.T) {
	alloc := t.Allocator()
	if got := hex.Dump(alloc, nil); got != "" {
		t.Errorf("Dump(nil) = %s, want an empty string", got)
	}
}

func TestDumpBrute(t *testing.T) {
	var in [40]byte
	var outArr [256]byte
	buf := strings.FixedBuilder(outArr[:0])
	dumpBrute(&buf, dumpInput(in[:]))
	if buf.String() != goldenDump() {
		t.Errorf("dumpBrute() = \n%s\nwant\n%s", buf.String(), goldenDump())
	}
}

func TestDumper(t *testing.T) {
	// Write the golden input in every chunk size. The dump does not
	// depend on the sizes of the writes.
	var in [40]byte
	dumpInput(in[:])
	var outArr [256]byte

	for stride := 1; stride < len(in); stride++ {
		out := strings.FixedBuilder(outArr[:0])
		dumper := hex.NewDumper(&out)
		done := 0
		for done < len(in) {
			todo := min(done+stride, len(in))
			dumper.Write(in[done:todo])
			done = todo
		}
		dumper.Close()
		if out.String() != goldenDump() {
			t.Errorf("stride %d: the dumper wrote \n%s\nwant\n%s",
				stride, out.String(), goldenDump())
			return
		}
	}
}

func TestDumperDoubleClose(t *testing.T) {
	var outArr [128]byte
	out := strings.FixedBuilder(outArr[:0])
	dumper := hex.NewDumper(&out)

	dumper.Write([]byte("gopher"))
	if err := dumper.Close(); err != nil {
		t.Errorf("Close() = %s, want nil", errName(errCode(err)))
	}
	if err := dumper.Close(); err != nil {
		t.Errorf("Close() = %s, want nil", errName(errCode(err)))
	}
	n, err := dumper.Write([]byte("gopher"))
	if n != 0 || errCode(err) != errDumperClosed {
		t.Errorf("Write() = %d, %s, want 0, ErrDumperClosed", n, errName(errCode(err)))
	}
	if err := dumper.Close(); err != nil {
		t.Errorf("Close() = %s, want nil", errName(errCode(err)))
	}

	want := "00000000  67 6f 70 68 65 72                                 |gopher|\n"
	if out.String() != want {
		t.Errorf("the dumper wrote \n%s\nwant\n%s", out.String(), want)
	}
}

func TestDumperEarlyClose(t *testing.T) {
	var outArr [128]byte
	out := strings.FixedBuilder(outArr[:0])
	dumper := hex.NewDumper(&out)

	if err := dumper.Close(); err != nil {
		t.Errorf("Close() = %s, want nil", errName(errCode(err)))
	}
	dumper.Write([]byte("gopher"))
	if out.String() != "" {
		t.Errorf("the dumper wrote %s, want nothing", out.String())
	}
}

func TestDumperWriteErr(t *testing.T) {
	var w errWriter
	dumper := hex.NewDumper(&w)
	n, err := dumper.Write([]byte("gopher"))
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Write() = %d, %s, want 0, errWriteFailed", n, errName(errCode(err)))
	}
}

func TestDumperLateWriteErr(t *testing.T) {
	// Fail the underlying writer in the middle of a line.
	// The dumper writes the offset, then one write per byte, then the ASCII
	// column. A limit of 4 stops it inside the first line.
	w := lateErrWriter{left: 4}
	dumper := hex.NewDumper(&w)
	n, err := dumper.Write([]byte("gopher"))
	if n != 3 || errCode(err) != errWrite {
		t.Errorf("Write() = %d, %s, want 3, errWriteFailed", n, errName(errCode(err)))
	}
}

func TestDumperCloseErr(t *testing.T) {
	// Check that the dumper reports the error of the underlying writer at a close.
	// The writer accepts the offset, the six bytes, and fails at the padding
	// of the last line.
	w := lateErrWriter{left: 7}
	dumper := hex.NewDumper(&w)
	if _, err := dumper.Write([]byte("gopher")); err != nil {
		t.Errorf("Write() = %s, want nil", errName(errCode(err)))
		return
	}
	if err := dumper.Close(); errCode(err) != errWrite {
		t.Errorf("Close() = %s, want errWriteFailed", errName(errCode(err)))
	}
}

// dumpArenaSize holds the allocations of one step of the dump sweep. The sweep
// resets the arena after every step.
const dumpArenaSize = 2048

func TestDumpSweep(t *testing.T) {
	// Compare Dump against the reference dump for every length up to four
	// lines, over bytes that cover the printable range and both edges.
	var backing [dumpArenaSize]byte
	arena := mem.NewArena(backing[:])
	var in [64]byte
	var wantArr [512]byte

	for length := range len(in) + 1 {
		for _, step := range dumpSteps {
			for i := range length {
				in[i] = byte(i * step)
			}
			want := strings.FixedBuilder(wantArr[:0])
			dumpBrute(&want, in[:length])

			got := hex.Dump(&arena, in[:length])
			if got != want.String() {
				t.Errorf("length %d, step %d: Dump() = \n%s\nwant\n%s",
					length, step, got, want.String())
				return
			}
			arena.Reset()
		}
	}
}
