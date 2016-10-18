// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"github.com/gonum/blas"
	"github.com/gonum/lapack"
)

// Dormlq multiplies the matrix c by the othogonal matrix q defined by the
// slices a and tau. A and tau are as returned from Dgelqf.
//  C = Q * C    if side == blas.Left and trans == blas.NoTrans
//  C = Q^T * C  if side == blas.Left and trans == blas.Trans
//  C = C * Q    if side == blas.Right and trans == blas.NoTrans
//  C = C * Q^T  if side == blas.Right and trans == blas.Trans
// If side == blas.Left, a is a matrix of side k×m, and if side == blas.Right
// a is of size k×n. This uses a blocked algorithm.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= m if side == blas.Left and lwork >= n if side == blas.Right,
// and this function will panic otherwise.
// Dormlq uses a block algorithm, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Dormlq,
// the optimal work length will be stored into work[0].
//
// Tau contains the householder scales and must have length at least k, and
// this function will panic otherwise.
func (impl Implementation) Dormlq(side blas.Side, trans blas.Transpose, m, n, k int, a []float64, lda int, tau, c []float64, ldc int, work []float64, lwork int) {
	if side != blas.Left && side != blas.Right {
		panic(badSide)
	}
	if trans != blas.Trans && trans != blas.NoTrans {
		panic(badTrans)
	}
	left := side == blas.Left
	notran := trans == blas.NoTrans
	if left {
		checkMatrix(k, m, a, lda)
	} else {
		checkMatrix(k, n, a, lda)
	}
	checkMatrix(m, n, c, ldc)
	if len(tau) < k {
		panic(badTau)
	}

	const nbmax = 64
	nw := n
	if !left {
		nw = m
	}
	opts := string(side) + string(trans)
	nb := min(nbmax, impl.Ilaenv(1, "DORMLQ", opts, m, n, k, -1))
	lworkopt := max(1, nw) * nb
	if lwork == -1 {
		work[0] = float64(lworkopt)
		return
	}
	if left {
		if lwork < n {
			panic(badWork)
		}
	} else {
		if lwork < m {
			panic(badWork)
		}
	}

	if m == 0 || n == 0 || k == 0 {
		return
	}
	nbmin := 2

	ldwork := nb
	if nb > 1 && nb < k {
		iws := nw * nb
		if lwork < iws {
			nb = lwork / nw
			nbmin = max(2, impl.Ilaenv(2, "DORMLQ", opts, m, n, k, -1))
		}
	}
	if nb < nbmin || nb >= k {
		// Call unblocked code
		impl.Dorml2(side, trans, m, n, k, a, lda, tau, c, ldc, work)
		return
	}
	ldt := nb
	t := make([]float64, nb*ldt)

	transt := blas.NoTrans
	if notran {
		transt = blas.Trans
	}

	switch {
	case left && notran:
		for i := 0; i < k; i += nb {
			ib := min(nb, k-i)
			impl.Dlarft(lapack.Forward, lapack.RowWise, m-i, ib,
				a[i*lda+i:], lda,
				tau[i:],
				t, ldt)
			impl.Dlarfb(side, transt, lapack.Forward, lapack.RowWise, m-i, n, ib,
				a[i*lda+i:], lda,
				t, ldt,
				c[i*ldc:], ldc,
				work, ldwork)
		}
		return
	case left && !notran:
		for i := ((k - 1) / nb) * nb; i >= 0; i -= nb {
			ib := min(nb, k-i)
			impl.Dlarft(lapack.Forward, lapack.RowWise, m-i, ib,
				a[i*lda+i:], lda,
				tau[i:],
				t, ldt)
			impl.Dlarfb(side, transt, lapack.Forward, lapack.RowWise, m-i, n, ib,
				a[i*lda+i:], lda,
				t, ldt,
				c[i*ldc:], ldc,
				work, ldwork)
		}
		return
	case !left && notran:
		for i := ((k - 1) / nb) * nb; i >= 0; i -= nb {
			ib := min(nb, k-i)
			impl.Dlarft(lapack.Forward, lapack.RowWise, n-i, ib,
				a[i*lda+i:], lda,
				tau[i:],
				t, ldt)
			impl.Dlarfb(side, transt, lapack.Forward, lapack.RowWise, m, n-i, ib,
				a[i*lda+i:], lda,
				t, ldt,
				c[i:], ldc,
				work, ldwork)
		}
		return
	case !left && !notran:
		for i := 0; i < k; i += nb {
			ib := min(nb, k-i)
			impl.Dlarft(lapack.Forward, lapack.RowWise, n-i, ib,
				a[i*lda+i:], lda,
				tau[i:],
				t, ldt)
			impl.Dlarfb(side, transt, lapack.Forward, lapack.RowWise, m, n-i, ib,
				a[i*lda+i:], lda,
				t, ldt,
				c[i:], ldc,
				work, ldwork)
		}
		return
	}
}
