// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lapack

import "github.com/gonum/blas"

const None = 'N'

type Job byte

// CompSV determines if the singular values are to be computed in compact form.
type CompSV byte

const (
	Compact  CompSV = 'P'
	Explicit CompSV = 'I'
)

// Complex128 defines the public complex128 LAPACK API supported by gonum/lapack.
type Complex128 interface{}

// Float64 defines the public float64 LAPACK API supported by gonum/lapack.
type Float64 interface {
	Dpotrf(ul blas.Uplo, n int, a []float64, lda int) (ok bool)
}

// Direct specifies the direction of the multiplication for the Householder matrix.
type Direct byte

const (
	Forward  Direct = 'F' // Reflectors are right-multiplied, H_1 * H_2 * ... * H_k
	Backward Direct = 'B' // Reflectors are left-multiplied, H_k * ... * H_2 * H_1
)

// StoreV indicates the storage direction of elementary reflectors.
type StoreV byte

const (
	ColumnWise StoreV = 'C' // Reflector stored in a column of the matrix.
	RowWise    StoreV = 'R' // Reflector stored in a row of the matrix.
)

// MatrixNorm represents the kind of matrix norm to compute.
type MatrixNorm byte

const (
	MaxAbs       MatrixNorm = 'M' // max(abs(A(i,j)))  ('M')
	MaxColumnSum MatrixNorm = 'O' // Maximum column sum (one norm) ('1', 'O')
	MaxRowSum    MatrixNorm = 'I' // Maximum row sum (infinity norm) ('I', 'i')
	NormFrob     MatrixNorm = 'F' // Frobenium norm (sqrt of sum of squares) ('F', 'f', E, 'e')
)

// MatrixType represents the kind of matrix represented in the data.
type MatrixType byte

const (
	General MatrixType = 'G' // A dense matrix (like blas64.General).
)
