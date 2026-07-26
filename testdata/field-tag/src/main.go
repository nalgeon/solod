package main

import "example/sdl"

// Regular struct.
type Event struct {
	etype uint32 `c:"type"`
	data  int32
}

// Generic struct.
type Box[T any] struct {
	id uint32 `c:"ident"`
}

func main() {
	// Keyed literal, field assignment, and field access.
	e := Event{etype: 7, data: 42}
	e.etype = 8
	if e.etype != 8 {
		panic("unexpected Event.etype value")
	}

	// Positional literal.
	p := Event{9, 10}
	if p.etype != 9 {
		panic("unexpected Event.etype value")
	}

	// Instantiated generic field resolves the same override.
	b := Box[int]{id: 5}
	if b.id != 5 {
		panic("unexpected Box.id value")
	}

	// Override declared in an imported package still applies here.
	ce := sdl.CommonEvent{Type: 3}
	ce.Type = 4
	if ce.Type != 4 {
		panic("unexpected CommonEvent.Type value")
	}
}
