package main

type Movie struct {
	year     int
	ratingFn func(m Movie) int
}

func freshness(m Movie) int {
	return m.year - 1970
}

// Must define a named function type to use it
// as function argument or return value.
type RatingFn func(m Movie) int

func getRatingFn() RatingFn {
	return freshness
}

func rateMovie(m Movie, f RatingFn) int {
	return f(m)
}

// Returning anonymous function types is not supported.
// func getRatingFn() func(m Movie) int {
// 	return freshness
// }

// Passing anonymous function types is not supported.
// func rateMovie(m Movie, f func(m Movie) int) int {
// 	return f(m)
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
