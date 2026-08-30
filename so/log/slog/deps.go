// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog

import "solod.dev/so/c"

// Implementations to avoid importing other dependencies.

// package math

func float64bits(f float64) uint64     { return c.Bitcast[uint64](f) }
func float64frombits(b uint64) float64 { return c.Bitcast[float64](b) }
