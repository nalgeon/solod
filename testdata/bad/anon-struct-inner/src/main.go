package main

func main() {
	var s struct{ inner struct{ f int } }
	s.inner.f = 10
}
