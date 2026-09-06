package main

type node struct {
	value int
	next  *node
}

func two() int { return 2 }

func incr(n *int) int {
	*n++
	return *n
}

func main() {
	testBasic()
	testClause()
}

func testBasic() {
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

	// The loop assigns to k2, so it keeps the value of the last iteration.
	var k2 int
	for k2 = range 3 {
	}
	if k2 != 2 {
		panic("want k2 == 2")
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

func testClause() {
	{
		// Init statements that go before the loop.
		for i, j := 0, 3; i < j; i++ {
			_ = i
		}
		src := [2]int{1, 2}
		for arr := src; false; {
			_ = arr
		}
		m := make(map[string]int, 2)
		for v, ok := m["k"]; false; {
			_, _ = v, ok
		}
		for m["k"] = 5; false; {
		}
		if m["k"] != 5 {
			panic("want m[\"k\"] == 5")
		}
		var a any
		for a = 1; true; {
			if a.(int) != 1 {
				panic("want a.(int) == 1")
			}
			break
		}
		for _ = 0; false; {
		}
	}
	{
		// Post statements that fit in the clause.
		sum := 0
		for i := 0; i < 9; i += 3 {
			sum += i
		}
		if sum != 9 {
			panic("want sum == 9")
		}
		bits := 0
		for i := 8; i > 0; i >>= 1 {
			bits++
		}
		if bits != 4 {
			panic("want bits == 4")
		}
		// A map variable emits so_Map*, a scalar.
		mm := make(map[string]int, 2)
		for i := 0; i < 1; mm = nil {
			i++
		}
		if mm != nil {
			panic("want mm == nil")
		}
		// The division keeps its zero divisor guard.
		n := 10
		for i := 0; i < 1; n /= two() {
			i++
		}
		if n != 5 {
			panic("want n == 5")
		}
		// A pointer walk and a discarded call.
		tail := node{value: 2}
		head := node{value: 1, next: &tail}
		count := 0
		for p := &head; p != nil; p = p.next {
			count++
		}
		if count != 2 {
			panic("want count == 2")
		}
		calls := 0
		for i := 0; i < 2; _ = incr(&calls) {
			i++
		}
		if calls != 2 {
			panic("want calls == 2")
		}
	}
}
