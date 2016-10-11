// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package blas64 provides a simple interface to the float64 BLAS API.
package blas64

import (
	"github.com/gonum/blas"
	"github.com/gonum/blas/native"
)

var blas64 blas.Float64 = native.Implementation{}

// Use sets the BLAS float64 implementation to be used by subsequent BLAS calls.
// The default implementation is native.Implementation.
func Use(b blas.Float64) {
	blas64 = b
}

// Implementation returns the current BLAS float64 implementation.
//
// Implementation allows direct calls to the current the BLAS float64 implementation
// giving finer control of parameters.
func Implementation() blas.Float64 {
	return blas64
}

// Vector represents a vector with an associated element increment.
type Vector struct {
	Inc  int
	Data []float64
}

// General represents a matrix using the conventional storage scheme.
type General struct {
	Rows, Cols int
	Stride     int
	Data       []float64
}

// Band represents a band matrix using the band storage scheme.
type Band struct {
	Rows, Cols int
	KL, KU     int
	Stride     int
	Data       []float64
}

// Triangular represents a triangular matrix using the conventional storage scheme.
type Triangular struct {
	N      int
	Stride int
	Data   []float64
	Uplo   blas.Uplo
	Diag   blas.Diag
}

// TriangularBand represents a triangular matrix using the band storage scheme.
type TriangularBand struct {
	N, K   int
	Stride int
	Data   []float64
	Uplo   blas.Uplo
	Diag   blas.Diag
}

// TriangularPacked represents a triangular matrix using the packed storage scheme.
type TriangularPacked struct {
	N    int
	Data []float64
	Uplo blas.Uplo
	Diag blas.Diag
}

// Symmetric represents a symmetric matrix using the conventional storage scheme.
type Symmetric struct {
	N      int
	Stride int
	Data   []float64
	Uplo   blas.Uplo
}

// SymmetricBand represents a symmetric matrix using the band storage scheme.
type SymmetricBand struct {
	N, K   int
	Stride int
	Data   []float64
	Uplo   blas.Uplo
}

// SymmetricPacked represents a symmetric matrix using the packed storage scheme.
type SymmetricPacked struct {
	N    int
	Data []float64
	Uplo blas.Uplo
}

// Level 1

const negInc = "blas64: negative vector increment"

// Dot computes the dot product of the two vectors
//  \sum_i x[i]*y[i]
func Dot(n int, x, y Vector) float64 {
	return blas64.Ddot(n, x.Data, x.Inc, y.Data, y.Inc)
}

// Nrm2 computes the Euclidean norm of a vector,
//  sqrt(\sum_i x[i] * x[i]).
//
// Nrm2 will panic if the vector increment is negative.
func Nrm2(n int, x Vector) float64 {
	if x.Inc < 0 {
		panic(negInc)
	}
	return blas64.Dnrm2(n, x.Data, x.Inc)
}

// Asum computes the sum of the absolute values of the elements of x.
//  \sum_i |x[i]|
//
// Asum will panic if the vector increment is negative.
func Asum(n int, x Vector) float64 {
	if x.Inc < 0 {
		panic(negInc)
	}
	return blas64.Dasum(n, x.Data, x.Inc)
}

// Iamax returns the index of the largest element of x. If there are multiple
// such indices the earliest is returned. Iamax returns -1 if n == 0.
//
// Iamax will panic if the vector increment is negative.
func Iamax(n int, x Vector) int {
	if x.Inc < 0 {
		panic(negInc)
	}
	return blas64.Idamax(n, x.Data, x.Inc)
}

// Swap exchanges the elements of two vectors.
//  x[i], y[i] = y[i], x[i] for all i
func Swap(n int, x, y Vector) {
	blas64.Dswap(n, x.Data, x.Inc, y.Data, y.Inc)
}

// Copy copies the elements of x into the elements of y.
//  y[i] = x[i] for all i
func Copy(n int, x, y Vector) {
	blas64.Dcopy(n, x.Data, x.Inc, y.Data, y.Inc)
}

// Axpy adds alpha times x to y
//  y[i] += alpha * x[i] for all i
func Axpy(n int, alpha float64, x, y Vector) {
	blas64.Daxpy(n, alpha, x.Data, x.Inc, y.Data, y.Inc)
}

// Rotg computes the plane rotation
//   _    _      _ _       _ _
//  | c  s |    | a |     | r |
//  | -s c |  * | b |   = | 0 |
//   ‾    ‾      ‾ ‾       ‾ ‾
// where
//  r = ±(a^2 + b^2)
//  c = a/r, the cosine of the plane rotation
//  s = b/r, the sine of the plane rotation
func Rotg(a, b float64) (c, s, r, z float64) {
	return blas64.Drotg(a, b)
}

// Rotmg computes the modified Givens rotation. See
// http://www.netlib.org/lapack/explore-html/df/deb/drotmg_8f.html
// for more details.
func Rotmg(d1, d2, b1, b2 float64) (p blas.DrotmParams, rd1, rd2, rb1 float64) {
	return blas64.Drotmg(d1, d2, b1, b2)
}

// Rot applies a plane transformation.
//  x[i] = c * x[i] + s * y[i]
//  y[i] = c * y[i] - s * x[i]
func Rot(n int, x, y Vector, c, s float64) {
	blas64.Drot(n, x.Data, x.Inc, y.Data, y.Inc, c, s)
}

// Rotm applies the modified Givens rotation to the 2×n matrix.
func Rotm(n int, x, y Vector, p blas.DrotmParams) {
	blas64.Drotm(n, x.Data, x.Inc, y.Data, y.Inc, p)
}

// Scal scales x by alpha.
//  x[i] *= alpha
//
// Scal will panic if the vector increment is negative
func Scal(n int, alpha float64, x Vector) {
	if x.Inc < 0 {
		panic(negInc)
	}
	blas64.Dscal(n, alpha, x.Data, x.Inc)
}

// Level 2

// Gemv computes
//  y = alpha * a * x + beta * y if tA = blas.NoTrans
//  y = alpha * A^T * x + beta * y if tA = blas.Trans or blas.ConjTrans
// where A is an m×n dense matrix, x and y are vectors, and alpha is a scalar.
func Gemv(tA blas.Transpose, alpha float64, a General, x Vector, beta float64, y Vector) {
	blas64.Dgemv(tA, a.Rows, a.Cols, alpha, a.Data, a.Stride, x.Data, x.Inc, beta, y.Data, y.Inc)
}

// Gbmv computes
//  y = alpha * A * x + beta * y if tA == blas.NoTrans
//  y = alpha * A^T * x + beta * y if tA == blas.Trans or blas.ConjTrans
// where a is an m×n band matrix kL subdiagonals and kU super-diagonals, and
// m and n refer to the size of the full dense matrix it represents.
// x and y are vectors, and alpha and beta are scalars.
func Gbmv(tA blas.Transpose, alpha float64, a Band, x Vector, beta float64, y Vector) {
	blas64.Dgbmv(tA, a.Rows, a.Cols, a.KL, a.KU, alpha, a.Data, a.Stride, x.Data, x.Inc, beta, y.Data, y.Inc)
}

// Trmv computes
//  x = A * x if tA == blas.NoTrans
//  x = A^T * x if tA == blas.Trans or blas.ConjTrans
// A is an n×n Triangular matrix and x is a vector.
func Trmv(tA blas.Transpose, a Triangular, x Vector) {
	blas64.Dtrmv(a.Uplo, tA, a.Diag, a.N, a.Data, a.Stride, x.Data, x.Inc)
}

// Tbmv computes
//  x = A * x if tA == blas.NoTrans
//  x = A^T * x if tA == blas.Trans or blas.ConjTrans
// where A is an n×n triangular banded matrix with k diagonals, and x is a vector.
func Tbmv(tA blas.Transpose, a TriangularBand, x Vector) {
	blas64.Dtbmv(a.Uplo, tA, a.Diag, a.N, a.K, a.Data, a.Stride, x.Data, x.Inc)
}

// Tpmv computes
//  x = A * x if tA == blas.NoTrans
//  x = A^T * x if tA == blas.Trans or blas.ConjTrans
// where A is an n×n unit triangular matrix in packed format, and x is a vector.
func Tpmv(tA blas.Transpose, a TriangularPacked, x Vector) {
	blas64.Dtpmv(a.Uplo, tA, a.Diag, a.N, a.Data, x.Data, x.Inc)
}

// Trsv solves
//  A * x = b if tA == blas.NoTrans
//  A^T * x = b if tA == blas.Trans or blas.ConjTrans
// A is an n×n triangular matrix and x is a vector.
// At entry to the function, x contains the values of b, and the result is
// stored in place into x.
//
// No test for singularity or near-singularity is included in this
// routine. Such tests must be performed before calling this routine.
func Trsv(tA blas.Transpose, a Triangular, x Vector) {
	blas64.Dtrsv(a.Uplo, tA, a.Diag, a.N, a.Data, a.Stride, x.Data, x.Inc)
}

// Tbsv solves
//  A * x = b
// where A is an n×n triangular banded matrix with k diagonals in packed format,
// and x is a vector.
// At entry to the function, x contains the values of b, and the result is
// stored in place into x.
//
// No test for singularity or near-singularity is included in this
// routine. Such tests must be performed before calling this routine.
func Tbsv(tA blas.Transpose, a TriangularBand, x Vector) {
	blas64.Dtbsv(a.Uplo, tA, a.Diag, a.N, a.K, a.Data, a.Stride, x.Data, x.Inc)
}

// Tpsv solves
//  A * x = b if tA == blas.NoTrans
//  A^T * x = b if tA == blas.Trans or blas.ConjTrans
// where A is an n×n triangular matrix in packed format and x is a vector.
// At entry to the function, x contains the values of b, and the result is
// stored in place into x.
//
// No test for singularity or near-singularity is included in this
// routine. Such tests must be performed before calling this routine.
func Tpsv(tA blas.Transpose, a TriangularPacked, x Vector) {
	blas64.Dtpsv(a.Uplo, tA, a.Diag, a.N, a.Data, x.Data, x.Inc)
}

// Symv computes
//    y = alpha * A * x + beta * y,
// where a is an n×n symmetric matrix, x and y are vectors, and alpha and
// beta are scalars.
func Symv(alpha float64, a Symmetric, x Vector, beta float64, y Vector) {
	blas64.Dsymv(a.Uplo, a.N, alpha, a.Data, a.Stride, x.Data, x.Inc, beta, y.Data, y.Inc)
}

// Sbmv performs
//  y = alpha * A * x + beta * y
// where A is an n×n symmetric banded matrix, x and y are vectors, and alpha
// and beta are scalars.
func Sbmv(alpha float64, a SymmetricBand, x Vector, beta float64, y Vector) {
	blas64.Dsbmv(a.Uplo, a.N, a.K, alpha, a.Data, a.Stride, x.Data, x.Inc, beta, y.Data, y.Inc)
}

// Spmv performs
//    y = alpha * A * x + beta * y,
// where A is an n×n symmetric matrix in packed format, x and y are vectors
// and alpha and beta are scalars.
func Spmv(alpha float64, a SymmetricPacked, x Vector, beta float64, y Vector) {
	blas64.Dspmv(a.Uplo, a.N, alpha, a.Data, x.Data, x.Inc, beta, y.Data, y.Inc)
}

// Ger performs the rank-one operation
//  A += alpha * x * y^T
// where A is an m×n dense matrix, x and y are vectors, and alpha is a scalar.
func Ger(alpha float64, x, y Vector, a General) {
	blas64.Dger(a.Rows, a.Cols, alpha, x.Data, x.Inc, y.Data, y.Inc, a.Data, a.Stride)
}

// Syr performs the rank-one update
//  a += alpha * x * x^T
// where a is an n×n symmetric matrix, and x is a vector.
func Syr(alpha float64, x Vector, a Symmetric) {
	blas64.Dsyr(a.Uplo, a.N, alpha, x.Data, x.Inc, a.Data, a.Stride)
}

// Spr computes the rank-one operation
//  a += alpha * x * x^T
// where a is an n×n symmetric matrix in packed format, x is a vector, and
// alpha is a scalar.
func Spr(alpha float64, x Vector, a SymmetricPacked) {
	blas64.Dspr(a.Uplo, a.N, alpha, x.Data, x.Inc, a.Data)
}

// Syr2 performs the symmetric rank-two update
//  A += alpha * x * y^T + alpha * y * x^T
// where A is a symmetric n×n matrix, x and y are vectors, and alpha is a scalar.
func Syr2(alpha float64, x, y Vector, a Symmetric) {
	blas64.Dsyr2(a.Uplo, a.N, alpha, x.Data, x.Inc, y.Data, y.Inc, a.Data, a.Stride)
}

// Spr2 performs the symmetric rank-2 update
//  a += alpha * x * y^T + alpha * y * x^T
// where a is an n×n symmetric matirx in packed format and x and y are vectors.
func Spr2(alpha float64, x, y Vector, a SymmetricPacked) {
	blas64.Dspr2(a.Uplo, a.N, alpha, x.Data, x.Inc, y.Data, y.Inc, a.Data)
}

// Level 3

// Gemm computes
//  C = beta * C + alpha * A * B.
// tA and tB specify whether A or B are transposed. A, B, and C are m×n dense
// matrices.
func Gemm(tA, tB blas.Transpose, alpha float64, a, b General, beta float64, c General) {
	var m, n, k int
	if tA == blas.NoTrans {
		m, k = a.Rows, a.Cols
	} else {
		m, k = a.Cols, a.Rows
	}
	if tB == blas.NoTrans {
		n = b.Cols
	} else {
		n = b.Rows
	}
	blas64.Dgemm(tA, tB, m, n, k, alpha, a.Data, a.Stride, b.Data, b.Stride, beta, c.Data, c.Stride)
}

// Symm performs one of
//  C = alpha * A * B + beta * C if side == blas.Left
//  C = alpha * B * A + beta * C if side == blas.Right
// where A is an n×n symmetric matrix, B and C are m×n matrices, and alpha
// is a scalar.
func Symm(s blas.Side, alpha float64, a Symmetric, b General, beta float64, c General) {
	var m, n int
	if s == blas.Left {
		m, n = a.N, b.Cols
	} else {
		m, n = b.Rows, a.N
	}
	blas64.Dsymm(s, a.Uplo, m, n, alpha, a.Data, a.Stride, b.Data, b.Stride, beta, c.Data, c.Stride)
}

// Syrk performs the symmetric rank-k operation
//  C = alpha * A * A^T + beta*C
// C is an n×n symmetric matrix. A is an n×k matrix if tA == blas.NoTrans, and
// a k×n matrix otherwise. alpha and beta are scalars.
func Syrk(t blas.Transpose, alpha float64, a General, beta float64, c Symmetric) {
	var n, k int
	if t == blas.NoTrans {
		n, k = a.Rows, a.Cols
	} else {
		n, k = a.Cols, a.Rows
	}
	blas64.Dsyrk(c.Uplo, t, n, k, alpha, a.Data, a.Stride, beta, c.Data, c.Stride)
}

// Syr2k performs the symmetric rank 2k operation
//  C = alpha * A * B^T + alpha * B * A^T + beta * C
// where C is an n×n symmetric matrix. A and B are n×k matrices if
// tA == NoTrans and k×n otherwise. alpha and beta are scalars.
func Syr2k(t blas.Transpose, alpha float64, a, b General, beta float64, c Symmetric) {
	var n, k int
	if t == blas.NoTrans {
		n, k = a.Rows, a.Cols
	} else {
		n, k = a.Cols, a.Rows
	}
	blas64.Dsyr2k(c.Uplo, t, n, k, alpha, a.Data, a.Stride, b.Data, b.Stride, beta, c.Data, c.Stride)
}

// Trmm performs
//  B = alpha * A * B if tA == blas.NoTrans and side == blas.Left
//  B = alpha * A^T * B if tA == blas.Trans or blas.ConjTrans, and side == blas.Left
//  B = alpha * B * A if tA == blas.NoTrans and side == blas.Right
//  B = alpha * B * A^T if tA == blas.Trans or blas.ConjTrans, and side == blas.Right
// where A is an n×n triangular matrix, and B is an m×n matrix.
func Trmm(s blas.Side, tA blas.Transpose, alpha float64, a Triangular, b General) {
	blas64.Dtrmm(s, a.Uplo, tA, a.Diag, b.Rows, b.Cols, alpha, a.Data, a.Stride, b.Data, b.Stride)
}

// Trsm solves
//  A * X = alpha * B if tA == blas.NoTrans side == blas.Left
//  A^T * X = alpha * B if tA == blas.Trans or blas.ConjTrans, and side == blas.Left
//  X * A = alpha * B if tA == blas.NoTrans side == blas.Right
//  X * A^T = alpha * B if tA == blas.Trans or blas.ConjTrans, and side == blas.Right
// where A is an n×n triangular matrix, x is an m×n matrix, and alpha is a
// scalar.
//
// At entry to the function, X contains the values of B, and the result is
// stored in place into X.
//
// No check is made that A is invertible.
func Trsm(s blas.Side, tA blas.Transpose, alpha float64, a Triangular, b General) {
	blas64.Dtrsm(s, a.Uplo, tA, a.Diag, b.Rows, b.Cols, alpha, a.Data, a.Stride, b.Data, b.Stride)
}
