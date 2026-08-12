// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog

import "unsafe"

// Implementations to avoid importing other dependencies.

// package math

func float64bits(f float64) uint64     { return *(*uint64)(unsafe.Pointer(&f)) }
func float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }
