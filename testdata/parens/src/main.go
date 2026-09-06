package main

type point struct {
	x, y int
}

type shape interface {
	area() int
}

type rect struct {
	width, height int
}

func (r *rect) area() int {
	return r.width * r.height
}

func pair() (int, int) {
	return 7, 8
}

func main() {
	{
		// Parentheses around a blank identifier.
		(_) = 0
		(_) = 0
	}
	{
		// Parentheses around a variable, a dereference, an index, and a field.
		n := 0
		(n) = 1
		if n != 1 {
			panic("want n == 1")
		}
		p := &n
		(*p) = 2
		if n != 2 {
			panic("want n == 2")
		}
		s := []int{0}
		(s[0]) = 3
		if s[0] != 3 {
			panic("want s[0] == 3")
		}
		pt := point{}
		(pt.x) = 4
		if pt.x != 4 {
			panic("want pt.x == 4")
		}
	}
	{
		// Parentheses around a map index.
		m := make(map[string]int, 4)
		(m["k"]) = 5
		if m["k"] != 5 {
			panic("want m[\"k\"] == 5")
		}
	}
	{
		// Parentheses around a target of a multiple assignment.
		n := 0
		(_), n = 1, 2
		if n != 2 {
			panic("want n == 2")
		}
		var a, b int
		(a), (b) = 3, 4
		if a != 3 || b != 4 {
			panic("want a == 3 and b == 4")
		}
	}
	{
		// Parentheses around a target of a multi-return assignment.
		var n int
		(_), n = pair()
		if n != 8 {
			panic("want n == 8")
		}
	}
	{
		// Parentheses around a target of a comma-ok assignment.
		m := make(map[string]int, 4)
		m["k"] = 9
		var n int
		var ok bool
		(n), ok = m["k"]
		if n != 9 || !ok {
			panic("want n == 9 and ok")
		}
		(_), ok = m["none"]
		if ok {
			panic("want !ok")
		}
		var sh shape = &rect{width: 2, height: 3}
		_, (ok) = sh.(*rect)
		if !ok {
			panic("want ok")
		}
	}
	{
		// Parentheses around a range key and a range value.
		var i, v int
		sum := 0
		sl := []int{10, 20, 30}
		for (i), (v) = range sl {
			sum += i * v
		}
		if sum != 80 {
			panic("want sum == 80")
		}
		arr := [2]int{1, 2}
		total := 0
		for (i), v = range arr {
			total += i + v
		}
		if total != 4 {
			panic("want total == 4")
		}
		count := 0
		for (i) = range 3 {
			count += i
		}
		if count != 3 {
			panic("want count == 3")
		}
		var r rune
		runes := 0
		for (i), (r) = range "ab" {
			runes += i + int(r)
		}
		if runes != 'a'+'b'+1 {
			panic("want runes == 'a'+'b'+1")
		}
		m := make(map[string]int, 4)
		m["k"] = 11
		var k string
		for (k) = range m {
			if k != "k" {
				panic("want k == \"k\"")
			}
		}
		if k != "k" {
			panic("want k set after the loop")
		}
	}
}
