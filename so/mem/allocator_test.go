package mem

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestSystemAllocator(t *testing.T) {
	a := SystemAllocator{}

	t.Run("Alloc", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var size, align = 16, 8

			p, err := a.Alloc(size, align)
			be.Err(t, err, nil)
			defer a.Dealloc(p, size, align)

			b := p.([]byte)
			be.Equal(t, len(b), size)
		})
		t.Run("invalid size", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("want panic")
				}
			}()
			_, _ = a.Alloc(0, 8)
		})
		t.Run("invalid alignment", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("want panic")
				}
			}()
			_, _ = a.Alloc(16, 3)
		})
		t.Run("out of memory", func(t *testing.T) {
			_, err := a.Alloc(maxAllocSize+1, 8)
			be.Err(t, err, ErrOutOfMemory)
		})
	})

	t.Run("Realloc", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var size, align = 16, 8

			p, err := a.Alloc(size, align)
			be.Err(t, err, nil)

			newSize := 32
			p2, err := a.Realloc(p, size, newSize, align)
			be.Err(t, err, nil)

			b := p2.([]byte)
			be.Equal(t, len(b), newSize)

			a.Dealloc(p2, newSize, align)
		})
		t.Run("invalid size", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("want panic")
				}
			}()
			_, _ = a.Realloc(nil, 0, 16, 8)
		})
		t.Run("invalid alignment", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("want panic")
				}
			}()
			_, _ = a.Realloc(nil, 16, 32, 3)
		})
		t.Run("out of memory", func(t *testing.T) {
			_, err := a.Realloc(nil, 16, maxAllocSize+1, 8)
			be.Err(t, err, ErrOutOfMemory)
		})
	})
}
