// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"math"

	"github.com/gonum/lapack"
)

// Implementation is the native Go implementation of LAPACK routines. It
// is built on top of calls to the return of blas64.Implementation(), so while
// this code is in pure Go, the underlying BLAS implementation may not be.
type Implementation struct{}

var _ lapack.Float64 = Implementation{}

const (
	badDirect     = "lapack: bad direct"
	badLdA        = "lapack: index of a out of range"
	badSide       = "lapack: bad side"
	badStore      = "lapack: bad store"
	badTau        = "lapack: tau has insufficient length"
	badTrans      = "lapack: bad trans"
	badUplo       = "lapack: illegal triangle"
	badWork       = "lapack: insufficient working memory"
	badWorkStride = "lapack: insufficient working array stride"
	negDimension  = "lapack: negative matrix dimension"
	nLT0          = "lapack: n < 0"
	shortWork     = "lapack: working array shorter than declared"
)

// checkMatrix verifies the parameters of a matrix input.
func checkMatrix(m, n int, a []float64, lda int) {
	if m < 0 {
		panic("lapack: has negative number of rows")
	}
	if m < 0 {
		panic("lapack: has negative number of columns")
	}
	if lda < n {
		panic("lapack: stride less than number of columns")
	}
	if len(a) < (m-1)*lda+n {
		panic("lapack: insufficient matrix slice length")
	}
}

func checkVector(n int, v []float64, inc int) {
	if n < 0 {
		panic("lapack: negative matrix length")
	}
	if (inc > 0 && (n-1)*inc >= len(v)) || (inc < 0 && (1-n)*inc >= len(v)) {
		panic("lapack: insufficient vector slice length")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// dlamch is a function in fortran, but since go forces IEEE-754 these are all
// fixed values. Probably a way to get them as constants.
// TODO(btracey): Is there a better way to find the smallest number such that 1+E > 1

var dlamchE, dlamchS, dlamchP float64

func init() {
	onePlusEps := math.Nextafter(1, math.Inf(1))
	eps := (math.Nextafter(1, math.Inf(1)) - 1) * 0.5
	dlamchE = eps
	sfmin := math.SmallestNonzeroFloat64
	small := 1 / math.MaxFloat64
	if small >= sfmin {
		sfmin = small * onePlusEps
	}
	dlamchS = sfmin
	radix := 2.0
	dlamchP = radix * eps
}
