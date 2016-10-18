// Generated code do not edit. Run `go generate`.

// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asm

func DsdotUnitary(x, y []float32) (sum float64) {
	for i, v := range x {
		sum += float64(y[i]) * float64(v)
	}
	return
}

func DsdotInc(x, y []float32, n, incX, incY, ix, iy uintptr) (sum float64) {
	for i := 0; i < int(n); i++ {
		sum += float64(y[iy]) * float64(x[ix])
		ix += incX
		iy += incY
	}
	return
}
