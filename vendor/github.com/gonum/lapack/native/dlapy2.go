// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import "math"

// Dlapy2 is the LAPACK version of math.Hypot.
func (Implementation) Dlapy2(x, y float64) float64 {
	return math.Hypot(x, y)
}
