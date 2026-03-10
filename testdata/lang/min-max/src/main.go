package main

func main() {
	{
		// 2 int args.
		a := 3
		b := 7
		if min(a, b) != 3 {
			panic("2 int args: min failed")
		}
		if max(a, b) != 7 {
			panic("2 int args: max failed")
		}
	}

	{
		// 3 int args.
		a := 5
		b := 2
		c := 8
		if min(a, b, c) != 2 {
			panic("3 int args: min failed")
		}
		if max(a, b, c) != 8 {
			panic("3 int args: max failed")
		}
	}

	{
		// float64 args.
		x := 1.5
		y := 2.5
		if min(x, y) != 1.5 {
			panic("float64 args: min failed")
		}
		if max(x, y) != 2.5 {
			panic("float64 args: max failed")
		}
	}

	{
		// string args.
		s1 := "apple"
		s2 := "banana"
		if min(s1, s2) != "apple" {
			panic("string args: min failed")
		}
		if max(s1, s2) != "banana" {
			panic("string args: max failed")
		}
	}
}
