// Copyright ©2013 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat64

import (
	"math"

	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
)

type LQFactor struct {
	LQ    *Dense
	lDiag []float64
}

// LQ computes an LQ Decomposition for an m-by-n matrix a with m <= n by Householder
// reflections. The LQ decomposition is an m-by-n orthogonal matrix q and an m-by-m
// lower triangular matrix l so that a = l.q. LQ will panic with ErrShape if m > n.
//
// The LQ decomposition always exists, even if the matrix does not have full rank,
// so LQ will never fail unless m > n. The primary use of the LQ decomposition is
// in the least squares solution of non-square systems of simultaneous linear equations.
// This will fail if LQIsFullRank() returns false. The matrix a is overwritten by the
// decomposition.
func LQ(a *Dense) LQFactor {
	// Initialize.
	m, n := a.Dims()
	if m > n {
		panic(ErrShape)
	}

	lq := *a

	lDiag := make([]float64, m)
	projs := NewVector(m, nil)

	// Main loop.
	for k := 0; k < m; k++ {
		hh := lq.RawRowView(k)[k:]
		norm := blas64.Nrm2(len(hh), blas64.Vector{Inc: 1, Data: hh})
		lDiag[k] = norm

		if norm != 0 {
			hhNorm := (norm * math.Sqrt(1-hh[0]/norm))
			if hhNorm == 0 {
				hh[0] = 0
			} else {
				// Form k-th Householder vector.
				s := 1 / hhNorm
				hh[0] -= norm
				blas64.Scal(len(hh), s, blas64.Vector{Inc: 1, Data: hh})

				// Apply transformation to remaining columns.
				if k < m-1 {
					a = lq.View(k+1, k, m-k-1, n-k).(*Dense)
					projs = projs.ViewVec(0, m-k-1)
					projs.MulVec(a, false, NewVector(len(hh), hh))

					for j := 0; j < m-k-1; j++ {
						dst := a.RawRowView(j)
						blas64.Axpy(len(dst), -projs.at(j),
							blas64.Vector{Inc: 1, Data: hh},
							blas64.Vector{Inc: 1, Data: dst},
						)
					}
				}
			}
		}
	}
	*a = lq

	return LQFactor{a, lDiag}
}

// IsFullRank returns whether the L matrix and hence a has full rank.
func (f LQFactor) IsFullRank() bool {
	for _, v := range f.lDiag {
		if v == 0 {
			return false
		}
	}
	return true
}

// L returns the lower triangular factor for the LQ decomposition.
func (f LQFactor) L() *Dense {
	lq, lDiag := f.LQ, f.lDiag
	m, _ := lq.Dims()
	l := NewDense(m, m, nil)
	for i, v := range lDiag {
		for j := 0; j < m; j++ {
			if i < j {
				l.set(j, i, lq.at(j, i))
			} else if i == j {
				l.set(j, i, v)
			}
		}
	}
	return l
}

// replaces x with Q.x
func (f LQFactor) applyQTo(x *Dense, trans bool) {
	nh, nc := f.LQ.Dims()
	m, n := x.Dims()
	if m != nc {
		panic(ErrShape)
	}
	proj := make([]float64, n)

	if trans {
		for k := nh - 1; k >= 0; k-- {
			hh := f.LQ.RawRowView(k)[k:]

			sub := x.View(k, 0, m-k, n).(*Dense)

			blas64.Gemv(blas.Trans,
				1, sub.mat, blas64.Vector{Inc: 1, Data: hh},
				0, blas64.Vector{Inc: 1, Data: proj},
			)
			for i := k; i < m; i++ {
				row := x.RawRowView(i)
				blas64.Axpy(n, -hh[i-k],
					blas64.Vector{Inc: 1, Data: proj},
					blas64.Vector{Inc: 1, Data: row},
				)
			}
		}
	} else {
		for k := 0; k < nh; k++ {
			hh := f.LQ.RawRowView(k)[k:]

			sub := x.View(k, 0, m-k, n).(*Dense)

			blas64.Gemv(blas.Trans,
				1, sub.mat, blas64.Vector{Inc: 1, Data: hh},
				0, blas64.Vector{Inc: 1, Data: proj},
			)
			for i := k; i < m; i++ {
				row := x.RawRowView(i)
				blas64.Axpy(n, -hh[i-k],
					blas64.Vector{Inc: 1, Data: proj},
					blas64.Vector{Inc: 1, Data: row},
				)
			}
		}
	}
}

// Solve computes minimum norm least squares solution of a.x = b where b has as many rows as a.
// A matrix x is returned that minimizes the two norm of Q*R*X-B. Solve will panic
// if a is not full rank.
func (f LQFactor) Solve(b *Dense) (x *Dense) {
	lq := f.LQ
	lDiag := f.lDiag
	m, n := lq.Dims()
	bm, bn := b.Dims()
	if bm != m {
		panic(ErrShape)
	}
	if !f.IsFullRank() {
		panic(ErrSingular)
	}

	x = NewDense(n, bn, nil)
	x.Copy(b)

	tau := make([]float64, m)
	for i := range tau {
		tau[i] = lq.at(i, i)
		lq.set(i, i, lDiag[i])
	}
	lqT := blas64.Triangular{
		// N omitted since it is not used by Trsm.
		Stride: lq.mat.Stride,
		Data:   lq.mat.Data,
		Uplo:   blas.Lower,
		Diag:   blas.NonUnit,
	}
	x.mat.Rows = bm
	blas64.Trsm(blas.Left, blas.NoTrans, 1, lqT, x.mat)
	x.mat.Rows = n
	for i := range tau {
		lq.set(i, i, tau[i])
	}

	f.applyQTo(x, true)

	return x
}
