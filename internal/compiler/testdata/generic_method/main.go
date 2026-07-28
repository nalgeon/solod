package main

// Phantom type param: the struct itself is fine.
type Box[T any] struct {
	id uint32
}

// The signature never mentions T, but call sites still pass it,
// so a regular C function cannot implement this.
func (b *Box[T]) Get() uint32 {
	return b.id
}

func main() {
	b := Box[int]{id: 5}
	_ = b.Get()
}
