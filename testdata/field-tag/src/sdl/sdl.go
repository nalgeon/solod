package sdl

// The c tag override must apply in packages that import this one.
type CommonEvent struct {
	Type      uint32 `c:"type"`
	Timestamp uint64
}
