// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import "github.com/gonum/blas"

// Dgeqr2 computes a QR factorization of the m×n matrix a.
//
// In a QR factorization, Q is an m×m orthonormal matrix, and R is an
// upper triangular m×n matrix.
//
// During Dgeqr2, a is modified to contain the information to construct Q and R.
// The upper triangle of a contains the matrix R. The lower triangular elements
// (not including the diagonal) contain the elementary reflectors. Tau is modified
// to contain the reflector scales. Tau must have length at least k = min(m,n), and
// this function will panic otherwise.
//
// The ith elementary reflector can be explicitly constructed by first extracting
// the
//  v[j] = 0           j < i
//  v[j] = i           j == i
//  v[j] = a[i*lda+j]  j > i
// and computing h_i = I - tau[i] * v * v^T.
//
// The orthonormal matrix Q can be constucted from a product of these elementary
// reflectors, Q = H_1*H_2 ... H_k, where k = min(m,n).
//
// Work is temporary storage of length at least n and this function will panic otherwise.
func (impl Implementation) Dgeqr2(m, n int, a []float64, lda int, tau, work []float64) {
	// TODO(btracey): This is oriented such that columns of a are eliminated.
	// This likely could be re-arranged to take better advantage of row-major
	// storage.
	checkMatrix(m, n, a, lda)
	if len(work) < n {
		panic(badWork)
	}
	k := min(m, n)
	if len(tau) < k {
		panic(badTau)
	}
	for i := 0; i < k; i++ {
		// Generate elementary reflector H(i).
		a[i*lda+i], tau[i] = impl.Dlarfg(m-i, a[i*lda+i], a[min((i+1), m-1)*lda+i:], lda)
		if i < n-1 {
			aii := a[i*lda+i]
			a[i*lda+i] = 1
			impl.Dlarf(blas.Left, m-i, n-i-1,
				a[i*lda+i:], lda,
				tau[i],
				a[i*lda+i+1:], lda,
				work)
			a[i*lda+i] = aii
		}
	}
}
