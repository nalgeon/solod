package mem

import "testing"

func TestSystemAllocatorPanic(t *testing.T) {
	a := SystemAllocator{}

	t.Run("Alloc invalid size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Alloc(0, 8)
	})
	t.Run("Alloc invalid alignment", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Alloc(16, 3)
	})
	t.Run("Realloc invalid size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Realloc(nil, 0, 16, 8)
	})
	t.Run("Realloc invalid alignment", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("want panic")
			}
		}()
		_, _ = a.Realloc(nil, 16, 32, 3)
	})
}
