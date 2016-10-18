// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import "github.com/gonum/blas"

// Dlaset sets the off-diagonal elements of a to alpha, and the diagonal elements
// of a to beta. If uplo == blas.Upper, only the upper diagonal elements are set.
// If uplo == blas.Lower, only the lower diagonal elements are set. If uplo is
// otherwise, all of the elements of a are set.
func (impl Implementation) Dlaset(uplo blas.Uplo, m, n int, alpha, beta float64, a []float64, lda int) {
	checkMatrix(m, n, a, lda)
	if uplo == blas.Upper {
		for i := 0; i < m; i++ {
			for j := i + 1; j < n; j++ {
				a[i*lda+j] = alpha
			}
		}
	} else if uplo == blas.Lower {
		for i := 0; i < m; i++ {
			for j := 0; j < i; j++ {
				a[i*lda+j] = alpha
			}
		}
	} else {
		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				a[i*lda+j] = alpha
			}
		}
	}
	for i := 0; i < min(m, n); i++ {
		a[i*lda+i] = beta
	}
}
