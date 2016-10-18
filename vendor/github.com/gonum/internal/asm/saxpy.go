// Generated code do not edit. Run `go generate`.

// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asm

// The extra z parameter is needed because of floats.AddScaledTo
func SaxpyUnitary(alpha float32, x, y, z []float32) {
	for i, v := range x {
		z[i] = alpha*v + y[i]
	}
}

func SaxpyInc(alpha float32, x, y []float32, n, incX, incY, ix, iy uintptr) {
	for i := 0; i < int(n); i++ {
		y[iy] += alpha * x[ix]
		ix += incX
		iy += incY
	}
}
