package strconv_bench

import (
	"fmt"
	"strconv"
	"testing"
)

func BenchmarkAtof64_Decimal_Go(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("33909", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Float_Go(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("339.7784", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Exp_Go(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("-5.09e75", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Big_Go(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("1844674407370955", 64)
		sinkFloat += f
	}
}

func BenchmarkParseInt_7bit_Go(b *testing.B) {
	s := fmt.Sprintf("%d", 1<<7-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_26bit_Go(b *testing.B) {
	s := fmt.Sprintf("%d", 1<<26-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_31bit_Go(b *testing.B) {
	s := fmt.Sprintf("%d", 1<<31-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_56bit_Go(b *testing.B) {
	s := fmt.Sprintf("%d", 1<<56-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_62bit_Go(b *testing.B) {
	s := fmt.Sprintf("%d", 1<<62-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkFormatFloat_Decimal_Go(b *testing.B) {
	for b.Loop() {
		s := strconv.FormatFloat(33909, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Float_Go(b *testing.B) {
	for b.Loop() {
		s := strconv.FormatFloat(339.7784, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Exp_Go(b *testing.B) {
	for b.Loop() {
		s := strconv.FormatFloat(-5.09e75, 'e', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Big_Go(b *testing.B) {
	for b.Loop() {
		s := strconv.FormatFloat(1844674407370955, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_7bit_Go(b *testing.B) {
	n := 1<<7 - 1
	for b.Loop() {
		s := strconv.FormatInt(int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_26bit_Go(b *testing.B) {
	n := 1<<26 - 1
	for b.Loop() {
		s := strconv.FormatInt(int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_31bit_Go(b *testing.B) {
	n := 1<<31 - 1
	for b.Loop() {
		s := strconv.FormatInt(int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_56bit_Go(b *testing.B) {
	n := 1<<56 - 1
	for b.Loop() {
		s := strconv.FormatInt(int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_62bit_Go(b *testing.B) {
	n := 1<<62 - 1
	for b.Loop() {
		s := strconv.FormatInt(int64(n), 10)
		sinkInt += len(s)
	}
}
