// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !noasm

package asm

// The extra z parameter is needed because of floats.AddScaledTo
func DaxpyUnitary(alpha float64, x, y, z []float64)

func DaxpyInc(alpha float64, x, y []float64, n, incX, incY, ix, iy uintptr)
