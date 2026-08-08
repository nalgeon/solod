package mem

import (
	"testing"

	"github.com/nalgeon/be"
)

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

func TestArenaOffset(t *testing.T) {
	t.Run("alignment padding", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)

		_, err := a.Alloc(1, 1)
		be.Err(t, err, nil)
		be.Equal(t, a.offset, 1)

		// 1 -> aligned to 8, then +8.
		_, err = a.Alloc(8, 8)
		be.Err(t, err, nil)
		be.Equal(t, a.lastStart, 8)
		be.Equal(t, a.offset, 16)
	})
	t.Run("exact fit", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)
		_, err := a.Alloc(256, 1)
		be.Err(t, err, nil)
		be.Equal(t, a.offset, 256)
	})
	t.Run("failed alloc keeps the offset", func(t *testing.T) {
		buf := make([]byte, 16)
		a := NewArena(buf)
		_, err := a.Alloc(32, 8)
		be.Err(t, err, ErrOutOfMemory)
		be.Equal(t, a.offset, 0)
	})
	t.Run("reset", func(t *testing.T) {
		buf := make([]byte, 256)
		a := NewArena(buf)

		_, err := a.Alloc(128, 8)
		be.Err(t, err, nil)
		be.Equal(t, a.offset, 128)

		a.Reset()
		be.Equal(t, a.offset, 0)
		be.Equal(t, a.lastStart, 0)
	})
}
