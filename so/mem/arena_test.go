package mem

import "testing"

func TestArenaPanic(t *testing.T) {
	t.Run("Alloc invalid size", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Alloc(0, 8)
	})
	t.Run("Alloc invalid alignment", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Alloc(16, 3)
	})
	t.Run("Realloc invalid size", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Realloc(nil, 0, 16, 8)
	})
	t.Run("Realloc invalid alignment", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Realloc(nil, 16, 32, 3)
	})
}

// block is a live arena allocation the fuzzer keeps track of.
type block struct {
	start int  // offset in the buffer
	size  int  // size in bytes
	fill  byte // the byte the fuzzer wrote over the whole block
}

func (b block) end() int {
	return b.start + b.size
}

func FuzzArena(f *testing.F) {
	f.Add([]byte{0, 0, 0, 4, 0, 8})
	f.Add([]byte{0, 1, 1, 0, 0, 2})
	f.Add([]byte{0, 3, 2, 40, 2, 9, 3, 0})
	f.Add([]byte{0, 4, 0, 4, 1, 1, 0, 4})

	const bufSize = 128

	f.Fuzz(func(t *testing.T, ops []byte) {
		if len(ops) > 64 {
			ops = ops[:64]
		}
		buf := make([]byte, bufSize)
		a := NewArena(buf)

		// offsetOf returns the position of ptr in the buffer.
		offsetOf := func(t *testing.T, ptr any) int {
			p := ptr.(*byte)
			for i := range buf {
				if &buf[i] == p {
					return i
				}
			}
			t.Fatal("the returned pointer is outside the buffer")
			return -1
		}

		var live []block
		var fill byte

		// check verifies the invariants of a block the arena returned,
		// then writes a new pattern over it.
		check := func(t *testing.T, name string, start int, size int, align int, keep int) {
			t.Helper()
			if start+size > bufSize {
				t.Fatalf("%s: block [%d,%d) is outside the buffer", name, start, start+size)
			}
			if align > 0 && start%align != 0 {
				t.Fatalf("%s: block at %d is not aligned to %d", name, start, align)
			}
			for _, b := range live {
				if start < b.end() && b.start < start+size {
					t.Fatalf("%s: block [%d,%d) overlaps the live block [%d,%d)",
						name, start, start+size, b.start, b.end())
				}
			}
			// The arena zeroes every byte it does not preserve.
			for i := keep; i < size; i++ {
				if buf[start+i] != 0 {
					t.Fatalf("%s: byte %d of the block at %d is not zeroed", name, i, start)
				}
			}
			fill++
			for i := range size {
				buf[start+i] = fill
			}
			live = append(live, block{start: start, size: size, fill: fill})
		}

		for i := 0; i+1 < len(ops); i += 2 {
			op := ops[i] % 4
			size := 1 + int(ops[i]/4)%17
			align := 1 << (int(ops[i+1]) % 5)

			switch {
			case op == 3:
				a.Reset()
				live = live[:0]

			case op == 0 || len(live) == 0:
				ptr, err := a.Alloc(size, align)
				if err != nil {
					continue
				}
				check(t, "Alloc", offsetOf(t, ptr), size, align, 0)

			case op == 1:
				idx := int(ops[i+1]) % len(live)
				b := live[idx]
				a.Free(&buf[b.start], b.size, align)
				live = append(live[:idx], live[idx+1:]...)

			default:
				idx := int(ops[i+1]) % len(live)
				b := live[idx]
				live = append(live[:idx], live[idx+1:]...)
				ptr, err := a.Realloc(&buf[b.start], b.size, size, align)
				if err != nil {
					// A failed Realloc keeps the block as it was.
					live = append(live, b)
					continue
				}
				start := offsetOf(t, ptr)
				keep := min(b.size, size)
				for j := range keep {
					if buf[start+j] != b.fill {
						t.Fatalf("Realloc: byte %d of the block at %d is not preserved", j, start)
					}
				}
				// Realloc aligns a block only when it moves the block.
				newAlign := align
				if start == b.start {
					newAlign = 0
				}
				check(t, "Realloc", start, size, newAlign, keep)
			}
		}
	})
}
