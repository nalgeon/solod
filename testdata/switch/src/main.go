package main

var calls int

// inc counts its calls, so a switch on it
// shows how many times the tag is evaluated.
func inc() int {
	calls++
	return calls
}

func name() string {
	calls++
	return "hello"
}

func twice(v int) int {
	return v * 2
}

type array [3]int

type Shape interface {
	Area() int
}

type Square struct {
	side int
}

func (s *Square) Area() int {
	return s.side * s.side
}

func main() {
	{
		// Empty switch statement.
		switch {
		}
	}
	{
		// Empty switch statement with a tag.
		i := 1
		switch i {
		}
	}
	{
		// Switch on int with cases and default.
		i := 2
		switch i {
		case 1:
			panic("unexpected i == 1")
		case 2:
			// ok
		case 3:
			panic("unexpected i == 3")
		default:
			panic("unexpected default")
		}
	}
	{
		// Tagless switch (bool conditions).
		x := 10
		switch {
		case x > 100:
			panic("unexpected x > 100")
		case x > 0:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Multiple values per case.
		y := 3
		switch y {
		case 1, 2, 3:
			if y != 3 {
				panic("want y == 3")
			}
		case 4, 5, 6:
			panic("unexpected y == 4, 5, 6")
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch with init statement.
		switch n := 42; n {
		case 42:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch on string.
		s := "hello"
		switch s {
		case "hello":
			// ok
		case "bye":
			panic("unexpected s == bye")
		default:
			panic("unexpected default")
		}
	}
	{
		// Cases without default. There is no default to catch a missed
		// case, so a flag records that the switch selected one.
		z := 5
		matched := false
		switch z {
		case 1:
			panic("unexpected z == 1")
		case 5:
			matched = true
		}
		if !matched {
			panic("want z == 5")
		}
	}
	{
		// The tag is evaluated only once.
		calls = 0
		switch inc() {
		case 2:
			panic("unexpected inc() == 2")
		case 3:
			panic("unexpected inc() == 3")
		default:
			if calls != 1 {
				panic("want 1 inc() call")
			}
		}
	}
	{
		// Same for a string tag.
		calls = 0
		switch name() {
		case "bye":
			panic("unexpected name() == bye")
		case "hello":
			if calls != 1 {
				panic("want 1 name() call")
			}
		default:
			panic("unexpected default")
		}
	}
	{
		// Init statement and a tag to evaluate.
		calls = 0
		switch base := 10; base + inc() {
		case 11:
			if calls != 1 {
				panic("want 1 inc() call")
			}
		default:
			panic("unexpected default")
		}
	}
	{
		// No cases to compare against, but the tag still runs.
		calls = 0
		switch inc() {
		default:
			if calls != 1 {
				panic("want 1 inc() call")
			}
		}
	}
	{
		// A case expression may convert and use builtins.
		s := "abc"
		var n int32 = 3
		switch len(s) {
		case int(n):
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Case expressions run in order, as in Go, so they may call functions.
		calls = 0
		switch {
		case inc() > 5:
			panic("unexpected inc() > 5")
		case inc() > 0:
			if calls != 2 {
				panic("want 2 inc() calls")
			}
		default:
			panic("unexpected default")
		}
	}
	{
		// The same for a tagged switch: the tag is already in a temporary,
		// so a case expression may change it.
		calls = 0
		switch calls {
		case inc():
			panic("unexpected calls == inc()")
		case 0:
			if calls != 1 {
				panic("want 1 inc() call")
			}
		default:
			panic("unexpected default")
		}
	}
	{
		// Once a case matches, the later case expressions do not run,
		// neither in the same case nor in the following ones.
		calls = 0
		switch 1 {
		case 1, inc():
			// ok
		case inc():
			panic("unexpected second case")
		default:
			panic("unexpected default")
		}
		if calls != 0 {
			panic("want no inc() calls")
		}
	}
	{
		// Case expressions that C groups differently than Go.
		a, b := 1, 2
		ok := true
		switch ok {
		case a < b && b < 3:
			// ok
		default:
			panic("unexpected default")
		}

		switch ok {
		case a == b:
			panic("unexpected a == b")
		default:
			// ok
		}

		switch 3 {
		case a + b:
			// ok
		default:
			panic("unexpected default")
		}

		switch 3 {
		case a | b:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// A binary case expression on a string tag compares with a call.
		s1, s2 := "he", "llo"
		switch "hello" {
		case s1 + s2:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Nested switches use separate tag temporaries.
		x, y := 1, 2
		switch x {
		case 1:
			switch y {
			case 2:
				// ok
			default:
				panic("unexpected inner default")
			}
		default:
			panic("unexpected outer default")
		}
	}
	{
		// A default-only body gets its own scope, and the tag still runs.
		calls = 0
		v := 1
		switch inc() {
		default:
			v := 2
			if v != 2 {
				panic("want inner v == 2")
			}
		}
		if calls != 1 {
			panic("want 1 inc() call")
		}
		if v != 1 {
			panic("want outer v == 1")
		}
	}
	{
		// A break in a case body is not supported, but a labeled break is.
		steps := 0
	done:
		switch 1 {
		case 1:
			steps++
			break done
		default:
			panic("unexpected default")
		}
		if steps != 1 {
			panic("want 1 step")
		}
	}
	{
		// A break inside a loop in a case body leaves the loop.
		n := 0
		switch 1 {
		case 1:
			for range 5 {
				n++
				if n == 2 {
					break
				}
			}
		default:
			panic("unexpected default")
		}
		if n != 2 {
			panic("want n == 2")
		}
	}
	{
		// Switch on a slice compares to nil.
		var s []int
		switch s {
		case nil:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch on a map compares to nil.
		var m map[string]int
		switch m {
		case nil:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch on an interface compares to nil.
		var sh Shape = &Square{2}
		switch sh {
		case nil:
			panic("unexpected sh == nil")
		default:
			if sh.Area() != 4 {
				panic("want sh.Area() == 4")
			}
		}
	}
	{
		// Pointer to array. A pointer to an unnamed array type
		// is not supported, so the case uses a named type.
		a := array{1, 2, 3}
		p := &a
		switch p {
		case nil:
			panic("unexpected p == nil")
		case &a:
			// ok
		default:
			panic("unexpected default")
		}
	}
	{
		// Function pointer.
		var fn func(int) int = twice
		switch fn {
		case nil:
			panic("unexpected fn == nil")
		default:
			if fn(2) != 4 {
				panic("want fn(2) == 4")
			}
		}
	}
}
