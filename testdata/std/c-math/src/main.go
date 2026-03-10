package main

import "github.com/nalgeon/solod/so/c/math"

func main() {
	pi := math.Pi
	_ = pi

	x := math.Sqrt(16.0)
	_ = x

	y := math.Pow(2.0, 10.0)
	_ = y

	z := math.Abs(-3.14)
	_ = z

	f := math.Floor(2.7)
	_ = f

	c := math.Ceil(2.3)
	_ = c

	r := math.Round(2.5)
	_ = r

	s := math.Sin(math.Pi)
	_ = s

	a := math.Atan2(1.0, 1.0)
	_ = a

	m := math.Fmin(3.0, 5.0)
	_ = m

	lg := math.Log(math.E)
	_ = lg

	fm := math.Fmod(5.5, 2.0)
	_ = fm
}
