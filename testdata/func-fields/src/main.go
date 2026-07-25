package main

type Movie struct {
	year     int
	ratingFn func(m Movie) int
	updateFn func(m *Movie)
	notifyFn func()
}

func freshness(m Movie) int {
	return m.year - 1970
}

// Global function variable.
var Global1 func()
var Global2 func() = nil
var rate func(m Movie) int = freshness

// A named function type can be used as a function argument or return value.
type RatingFn func(m Movie) int
type UpdateFn func(m *Movie)

func getRatingFn() RatingFn {
	return freshness
}

// Anonymous function types can be passed as arguments.
func rateMovie(m Movie, f func(m Movie) int) int {
	return f(m)
}

// Returning anonymous function types is not supported.
// func getRatingFn() func(m Movie) int {
// 	return freshness
// }

func main() {
	{
		// Function struct field.
		m1 := Movie{year: 2020, ratingFn: freshness}
		r1 := m1.ratingFn(m1)
		if r1 != 50 {
			panic("unexpected r1")
		}

		m2 := Movie{year: 1995, ratingFn: freshness}
		r2 := m2.ratingFn(m2)
		if r2 != 25 {
			panic("unexpected r2")
		}
	}
	{
		// Function variable.
		fn1 := freshness
		m := Movie{year: 2020}
		r3 := fn1(m)
		if r3 != 50 {
			panic("unexpected r3")
		}

		var fn2 RatingFn = freshness
		r4 := fn2(m)
		if r4 != 50 {
			panic("unexpected r4")
		}

		// Anonymous function type variable.
		var fn3 func(m Movie) int = freshness
		r4b := fn3(m)
		if r4b != 50 {
			panic("unexpected r4b")
		}

		r5 := rate(m)
		if r5 != 50 {
			panic("unexpected r5")
		}

		// Function variable without an initializer is nil.
		var fn4 func(m Movie) int
		if fn4 != nil {
			panic("unexpected fn4")
		}
		fn4 = freshness
		if fn4(m) != 50 {
			panic("unexpected fn4 result")
		}
	}
	{
		// Function argument.
		m := Movie{year: 2020}
		r5 := rateMovie(m, freshness)
		if r5 != 50 {
			panic("unexpected r5")
		}
	}
	{
		// Function return value.
		m := Movie{year: 2020}
		r6 := getRatingFn()(m)
		if r6 != 50 {
			panic("unexpected r6")
		}
	}
}
