package main

type box[T any] interface {
	get() T
}

type intBox struct{ v int }

func (b *intBox) get() int { return b.v }

func main() {
	var b box[int] = &intBox{42}
	_ = b.get()
}
