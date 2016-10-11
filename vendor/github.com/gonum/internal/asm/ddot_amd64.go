// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !noasm

package asm

func DdotUnitary(x, y []float64) (sum float64)
func DdotInc(x, y []float64, n, incX, incY, ix, iy uintptr) (sum float64)
