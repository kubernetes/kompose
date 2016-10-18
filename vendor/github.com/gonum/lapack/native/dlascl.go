// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"math"

	"github.com/gonum/lapack"
)

// Dlascl multiplies a rectangular matrix by a scalar.
func (impl Implementation) Dlascl(kind lapack.MatrixType, kl, ku int, cfrom, cto float64, m, n int, a []float64, lda int) {
	checkMatrix(m, n, a, lda)
	if cfrom == 0 {
		panic("dlascl: zero divisor")
	}
	if math.IsNaN(cfrom) || math.IsNaN(cto) {
		panic("dlascl: NaN scale factor")
	}
	if n == 0 || m == 0 {
		return
	}
	smlnum := dlamchS
	bignum := 1 / smlnum
	cfromc := cfrom
	ctoc := cto
	cfrom1 := cfromc * smlnum
	for {
		var done bool
		var mul, ctol float64
		if cfrom1 == cfromc {
			// cfromc is inf
			mul = ctoc / cfromc
			done = true
			ctol = ctoc
		} else {
			ctol = ctoc / bignum
			if ctol == ctoc {
				// ctoc is either 0 or inf.
				mul = ctoc
				done = true
				cfromc = 1
			} else if math.Abs(cfrom1) > math.Abs(ctoc) && ctoc != 0 {
				mul = smlnum
				done = false
				cfromc = cfrom1
			} else if math.Abs(ctol) > math.Abs(cfromc) {
				mul = bignum
				done = false
				ctoc = ctol
			} else {
				mul = ctoc / cfromc
				done = true
			}
		}
		switch kind {
		default:
			panic("lapack: not implemented")
		case lapack.General:
			for i := 0; i < m; i++ {
				for j := 0; j < n; j++ {
					a[i*lda+j] = a[i*lda+j] * mul
				}
			}
		}
		if done {
			break
		}
	}
}
