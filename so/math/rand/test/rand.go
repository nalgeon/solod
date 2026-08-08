package rand_test

import (
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
)

func TestInt(t *testing.T) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	n1 := r.Int()
	if n1 < 0 {
		t.Error("negative Int()")
	}
	n2 := r.Int()
	if n2 < 0 {
		t.Error("negative Int()")
	}
	if n1 == n2 {
		t.Error("same Int() twice in a row")
	}
}

func TestFloat64(t *testing.T) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	f1 := r.Float64()
	if f1 < 0 || f1 >= 1 {
		t.Error("Float64() out of range")
	}
	f2 := r.Float64()
	if f2 < 0 || f2 >= 1 {
		t.Error("Float64() out of range")
	}
	if f1 == f2 {
		t.Error("same Float64() twice in a row")
	}
}

func TestGlobal(t *testing.T) {
	n1 := rand.IntN(100)
	if n1 < 0 || n1 >= 100 {
		t.Error("IntN() out of range")
	}
	n2 := rand.IntN(100)
	if n2 < 0 || n2 >= 100 {
		t.Error("IntN() out of range")
	}
}
