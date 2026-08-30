// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import "solod.dev/so/c"

// Implementations to avoid importing other dependencies.

// package math

func float64frombits(b uint64) float64 { return c.Bitcast[float64](b) }
func float32frombits(b uint32) float32 { return c.Bitcast[float32](b) }
func float64bits(f float64) uint64     { return c.Bitcast[uint64](f) }
func float32bits(f float32) uint32     { return c.Bitcast[uint32](f) }

func floatInf(sign int) float64 {
	var v uint64
	if sign >= 0 {
		v = 0x7FF0000000000000
	} else {
		v = 0xFFF0000000000000
	}
	return float64frombits(v)
}

func floatNaN() float64 { return float64frombits(0x7FF8000000000001) }
