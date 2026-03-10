package main

func main() {
	i := 1
	for i <= 3 {
		println(i)
		i = i + 1
	}

	for j := 0; j < 3; j++ {
		println(j)
	}

	start := 5
	for start--; start >= 0; start-- {
		if start == 2 {
			break
		}
	}

	for start = 5; start >= 0; start-- {
	}

	for k := range 3 {
		println("range", k)
	}

	for range 3 {
	}

	for {
		println("loop")
		break
	}

	for n := range 6 {
		if n%2 == 0 {
			continue
		}
		println(n)
	}
}
