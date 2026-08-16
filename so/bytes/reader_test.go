// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"sync"
	"testing"
)

func TestReaderAtConcurrent(t *testing.T) {
	// A test for the race detector: ReadAt must change no state.
	r := NewReader([]byte("0123456789"))
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var buf [1]byte
			r.ReadAt(buf[:], int64(i))
		}(i)
	}
	wg.Wait()
}

func TestEmptyReaderConcurrent(t *testing.T) {
	// A test for the race detector: a Read that gives no bytes is safe from
	// several goroutines.
	r := NewReader([]byte{})
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			var buf [1]byte
			r.Read(buf[:])
		}()
		go func() {
			defer wg.Done()
			r.Read(nil)
		}()
	}
	wg.Wait()
}
