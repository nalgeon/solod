package main

type Meters struct{ n int }
type Feet struct{ n int }

func main() {
	m := Meters{n: 1}
	f := Feet(m)
	_ = f
}
