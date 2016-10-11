// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"github.com/gonum/internal/asm"
)

// Ddot computes the dot product of the two vectors
//  \sum_i x[i]*y[i]
func (Implementation) Ddot(n int, x []float64, incX int, y []float64, incY int) float64 {
	if n < 0 {
		panic(negativeN)
	}
	if incX == 0 {
		panic(zeroIncX)
	}
	if incY == 0 {
		panic(zeroIncY)
	}
	if incX == 1 && incY == 1 {
		if len(x) < n {
			panic(badLenX)
		}
		if len(y) < n {
			panic(badLenY)
		}
		return asm.DdotUnitary(x[:n], y)
	}
	var ix, iy int
	if incX < 0 {
		ix = (-n + 1) * incX
	}
	if incY < 0 {
		iy = (-n + 1) * incY
	}
	if ix >= len(x) || ix+(n-1)*incX >= len(x) {
		panic(badLenX)
	}
	if iy >= len(y) || iy+(n-1)*incY >= len(y) {
		panic(badLenY)
	}
	return asm.DdotInc(x, y, uintptr(n), uintptr(incX), uintptr(incY), uintptr(ix), uintptr(iy))
}
