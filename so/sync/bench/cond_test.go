package sync_bench

import (
	"sync"
	"testing"
)

func BenchmarkCond1_Go(b *testing.B)  { benchmarkCondGo(b, 1) }
func BenchmarkCond2_Go(b *testing.B)  { benchmarkCondGo(b, 2) }
func BenchmarkCond4_Go(b *testing.B)  { benchmarkCondGo(b, 4) }
func BenchmarkCond8_Go(b *testing.B)  { benchmarkCondGo(b, 8) }
func BenchmarkCond16_Go(b *testing.B) { benchmarkCondGo(b, 16) }
func BenchmarkCond32_Go(b *testing.B) { benchmarkCondGo(b, 32) }

// benchmarkCondGo mirrors benchmarkCond on the So side: waiters+1 persistent
// pool workers rendezvous on a Cond for numLoops rounds per iteration.
func benchmarkCondGo(b *testing.B, waiters int) {
	var mu sync.Mutex
	c := sync.NewCond(&mu)
	id := 0

	task := func() {
		for range numLoops {
			mu.Lock()
			if id == -1 {
				mu.Unlock()
				break
			}
			id++
			if id == waiters+1 {
				id = 0
				c.Broadcast()
			} else {
				c.Wait()
			}
			mu.Unlock()
		}
		mu.Lock()
		id = -1
		c.Broadcast()
		mu.Unlock()
	}

	p := newPool(waiters + 1)
	defer p.Free()

	for b.Loop() {
		id = 0
		for range waiters + 1 {
			p.Go(task)
		}
		p.Wait()
	}
}
