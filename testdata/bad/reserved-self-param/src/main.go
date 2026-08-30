package main

type Rect struct {
	W, H int
}

func (r *Rect) Scale(self int) {
	r.W *= self
}

func main() {
	r := Rect{2, 3}
	r.Scale(2)
}
