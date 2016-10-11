// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lapack64 provides a set of convenient wrapper functions for LAPACK
// calls, as specified in the netlib standard (www.netlib.org).
//
// The native Go routines are used by default, and the Use function can be used
// to set an alternate implementation.
//
// If the type of matrix (General, Symmetric, etc.) is known and fixed, it is
// used in the wrapper signature. In many cases, however, the type of the matrix
// changes during the call to the routine, for example the matrix is symmetric on
// entry and is triangular on exit. In these cases the correct types should be checked
// in the documentation.
//
// The full set of Lapack functions is very large, and it is not clear that a
// full implementation is desirable, let alone feasible. Please open up an issue
// if there is a specific function you need and/or are willing to implement.
package lapack64

import (
	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
	"github.com/gonum/lapack"
	"github.com/gonum/lapack/native"
)

var lapack64 lapack.Float64 = native.Implementation{}

// Use sets the LAPACK float64 implementation to be used by subsequent BLAS calls.
// The default implementation is native.Implementation.
func Use(l lapack.Float64) {
	lapack64 = l
}

// Potrf computes the cholesky factorization of a.
//  A = U^T * U if ul == blas.Upper
//  A = L * L^T if ul == blas.Lower
// The underlying data between the input matrix and output matrix is shared.
func Potrf(a blas64.Symmetric) (t blas64.Triangular, ok bool) {
	ok = lapack64.Dpotrf(a.Uplo, a.N, a.Data, a.Stride)
	t.Uplo = a.Uplo
	t.N = a.N
	t.Data = a.Data
	t.Stride = a.Stride
	t.Diag = blas.NonUnit
	return
}
