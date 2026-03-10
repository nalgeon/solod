package main

func main() {
	if 7%2 == 0 {
		panic("want 7%2 != 0")
	} else {
		println("7 is odd")
	}

	if 8%2 == 0 || 7%2 == 0 {
		println("either 8 or 7 are even")
	}

	if 1 == 2-1 && (2 == 1+1 || 3 == 6/2) && !(4 != 2*2) {
		println("all conditions are true")
	}

	if 9%3 == 0 {
		println("9 is divisible by 3")
	} else if 9%2 == 0 {
		panic("want 9%2 != 0")
	} else {
		panic("want 9%3 == 0")
	}

	if num := 9; num < 0 {
		panic("want num >= 0")
	} else if num < 10 {
		println(num, "has 1 digit")
	} else {
		panic("want 0 <= num < 10")
	}
}
