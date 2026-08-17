package main

func main() {
	// if-else
	if 7%2 == 0 {
		panic("want 7%2 != 0")
	} else {
		println("7 is odd")
	}

	// if without else
	if 8%2 == 0 || 7%2 == 0 {
		println("either 8 or 7 are even")
	}

	// if with a complex condition
	if 1 == 2-1 && (2 == 1+1 || 3 == 6/2) && !(4 != 2*2) {
		println("all conditions are true")
	}

	// if-elseif-else
	if 9%3 == 0 {
		println("9 is divisible by 3")
	} else if 9%2 == 0 {
		panic("want 9%2 != 0")
	} else {
		panic("want 9%3 == 0")
	}

	// if with init
	if num := 9; num < 0 {
		panic("want num >= 0")
	} else if num < 10 {
		println(num, "has 1 digit")
	} else {
		panic("want 0 <= num < 10")
	}

	// else-if init
	n := 0
	if n == 1 {
		panic("want n == 0")
	} else if m := n + 1; m == 1 {
		println("m == 1")
	} else {
		panic("want m == 1")
	}

	// else-if init that shadows an outer variable
	v := 100
	if v == 1 {
		panic("want v == 100")
	} else if v := n + 1; v == 1 {
		println("shadowed v == 1")
	} else {
		panic("want shadowed v == 1")
	}
	if v != 100 {
		panic("want outer v == 100")
	}

	// chained else-if with init
	if k := 1; k == 0 {
		panic("want k == 1")
	} else if j := k + 1; j == 0 {
		panic("want j == 2")
	} else if i := j + 1; i == 3 {
		println("i == 3")
	} else {
		panic("want i == 3")
	}
}
