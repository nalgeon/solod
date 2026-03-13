package main

func main() {
	fails := 0

	for i := range 10 {
		if i%2 == 0 {
			goto next
		}
	next:
		fails++
		if fails > 2 {
			goto fallback
		}
	}

fallback:
	if fails != 3 {
		panic("want fails == 3")
	}
}
