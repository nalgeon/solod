package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/testing"
)

func TestParse(t *testing.T) {
	flags := flag.NewFlagSet("example", flag.ContinueOnError)
	var b bool
	flags.BoolVar(&b, "b", false, "a boolean flag")
	var n int
	flags.IntVar(&n, "n", 0, "an int flag")
	var f float64
	flags.Float64Var(&f, "f", 0.0, "a float flag")
	var s string
	flags.StringVar(&s, "s", "default", "a string flag")

	err := flags.Parse([]string{"-b", "-n", "42", "-f", "3.14", "-s", "hello"})
	if err != nil {
		t.Fatal("Parse failed")
		return
	}

	if !b {
		t.Error("b != true")
	}
	if n != 42 {
		t.Error("n != 42")
	}
	if f != 3.14 {
		t.Error("f != 3.14")
	}
	if s != "hello" {
		t.Error("s != hello")
	}
}
