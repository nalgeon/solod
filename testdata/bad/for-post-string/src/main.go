package main

func main() {
	s := "x"
	for i := 0; i < 1; s += "y" {
		i++
	}
	println(s)
}
