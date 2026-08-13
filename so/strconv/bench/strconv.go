package strconv_bench

import (
	"solod.dev/so/fmt"
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

//so:volatile
var sinkInt int

//so:volatile
var sinkFloat float64

func BenchmarkAtof64_Decimal_So(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("33909", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Float_So(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("339.7784", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Exp_So(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("-5.09e75", 64)
		sinkFloat += f
	}
}

func BenchmarkAtof64_Big_So(b *testing.B) {
	for b.Loop() {
		f, _ := strconv.ParseFloat("1844674407370955", 64)
		sinkFloat += f
	}
}

func BenchmarkParseInt_7bit_So(b *testing.B) {
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%d", 1<<7-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_26bit_So(b *testing.B) {
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%d", 1<<26-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_31bit_So(b *testing.B) {
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%d", 1<<31-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_56bit_So(b *testing.B) {
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%d", 1<<56-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkParseInt_62bit_So(b *testing.B) {
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%d", 1<<62-1)
	for b.Loop() {
		n, _ := strconv.ParseInt(s, 10, 64)
		sinkInt += int(n)
	}
}

func BenchmarkFormatFloat_Decimal_So(b *testing.B) {
	buf := make([]byte, 64)
	for b.Loop() {
		s := strconv.FormatFloat(buf, 33909, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Float_So(b *testing.B) {
	buf := make([]byte, 64)
	for b.Loop() {
		s := strconv.FormatFloat(buf, 339.7784, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Exp_So(b *testing.B) {
	buf := make([]byte, 64)
	for b.Loop() {
		s := strconv.FormatFloat(buf, -5.09e75, 'e', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatFloat_Big_So(b *testing.B) {
	buf := make([]byte, 64)
	for b.Loop() {
		s := strconv.FormatFloat(buf, 1844674407370955, 'f', -1, 64)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_7bit_So(b *testing.B) {
	buf := make([]byte, 64)
	n := 1<<7 - 1
	for b.Loop() {
		s := strconv.FormatInt(buf, int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_26bit_So(b *testing.B) {
	buf := make([]byte, 64)
	n := 1<<26 - 1
	for b.Loop() {
		s := strconv.FormatInt(buf, int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_31bit_So(b *testing.B) {
	buf := make([]byte, 64)
	n := 1<<31 - 1
	for b.Loop() {
		s := strconv.FormatInt(buf, int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_56bit_So(b *testing.B) {
	buf := make([]byte, 64)
	n := 1<<56 - 1
	for b.Loop() {
		s := strconv.FormatInt(buf, int64(n), 10)
		sinkInt += len(s)
	}
}

func BenchmarkFormatInt_62bit_So(b *testing.B) {
	buf := make([]byte, 64)
	n := 1<<62 - 1
	for b.Loop() {
		s := strconv.FormatInt(buf, int64(n), 10)
		sinkInt += len(s)
	}
}
