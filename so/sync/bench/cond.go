package sync_bench

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/testing"
)

// condState is the shared rendezvous state for the Cond benchmark: a mutex,
// the condition variable it guards, the number of waiters, and the id counter
// the workers use to elect a broadcaster each round.
type condState struct {
	mu      sync.Mutex
	c       sync.Cond
	waiters int
	id      int
}

// BenchmarkCondN_So measures a Cond rendezvous with N waiters: each round one
// of the N+1 workers broadcasts and the rest wait, so every round wakes them all.
func BenchmarkCond1_So(b *testing.B)  { benchmarkCond(b, 1) }
func BenchmarkCond2_So(b *testing.B)  { benchmarkCond(b, 2) }
func BenchmarkCond4_So(b *testing.B)  { benchmarkCond(b, 4) }
func BenchmarkCond8_So(b *testing.B)  { benchmarkCond(b, 8) }
func BenchmarkCond16_So(b *testing.B) { benchmarkCond(b, 16) }
func BenchmarkCond32_So(b *testing.B) { benchmarkCond(b, 32) }

// benchmarkCond runs waiters+1 persistent pool threads that rendezvous
// on a Cond for numLoops rounds per benchmark iteration.
func benchmarkCond(b *testing.B, waiters int) {
	var st condState
	st.waiters = waiters
	st.mu.Init()
	defer st.mu.Free()
	st.c.Init(&st.mu)
	defer st.c.Free()

	opts := conc.PoolOptions{NumThreads: waiters + 1}
	p := conc.NewPool(mem.System, opts)
	defer p.Free()

	for b.Loop() {
		st.id = 0
		for range waiters + 1 {
			p.Go(condPingPong, &st)
		}
		p.Wait()
	}
}

// condPingPong runs numLoops rendezvous rounds on the shared Cond. In each round
// the last of the waiters+1 workers to arrive broadcasts and the rest wait, so
// every round wakes all workers once. Setting id to -1 releases stragglers.
func condPingPong(arg any) {
	st := arg.(*condState)
	for range numLoops {
		st.mu.Lock()
		if st.id == -1 {
			st.mu.Unlock()
			break
		}
		st.id++
		if st.id == st.waiters+1 {
			st.id = 0
			st.c.Broadcast()
		} else {
			st.c.Wait()
		}
		st.mu.Unlock()
	}
	st.mu.Lock()
	st.id = -1
	st.c.Broadcast()
	st.mu.Unlock()
}
