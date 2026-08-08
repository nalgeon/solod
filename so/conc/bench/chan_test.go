package conc_bench

import "testing"

func BenchmarkChanUncontended_Go(b *testing.B) {
	ch := make(chan int, chanCap)

	var v int
	for b.Loop() {
		for range chanCap {
			ch <- 0
		}
		for range chanCap {
			v = <-ch
		}
	}
	chanSink = v
}

func BenchmarkChanProdCons0_Go(b *testing.B) {
	benchChanProdCons_Go(b, 0)
}

func BenchmarkChanProdCons10_Go(b *testing.B) {
	benchChanProdCons_Go(b, 10)
}

func BenchmarkChanProdCons100_Go(b *testing.B) {
	benchChanProdCons_Go(b, 100)
}

// benchChanProdCons mirrors benchChanProdCons_So: one consumer goroutine drains
// the channel for the whole run while each timed iteration pushes chanBatch
// values through a channel of the given buffer size.
func benchChanProdCons_Go(b *testing.B, size int) {
	ch := make(chan int, size)
	done := make(chan int)
	go func() {
		sum := 0
		for v := range ch {
			sum += v
		}
		done <- sum
	}()

	for b.Loop() {
		for i := range chanBatch {
			ch <- i
		}
	}

	close(ch)
	chanSink = <-done
}
