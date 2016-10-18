// Copyright ©2013 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Based on the LUDecomposition class from Jama 1.0.3.

package mat64

import (
	"math"
)

type LUFactors struct {
	LU    *Dense
	Pivot []int
	Sign  int
}

// LU performs an LU Decomposition for an m-by-n matrix a.
//
// If m >= n, the LU decomposition is an m-by-n unit lower triangular matrix L,
// an n-by-n upper triangular matrix U, and a permutation vector piv of length m
// so that A(piv,:) = L*U.
//
// If m < n, then L is m-by-m and U is m-by-n.
//
// The LU decompostion with pivoting always exists, even if the matrix is
// singular, so LU will never fail. The primary use of the LU decomposition
// is in the solution of square systems of simultaneous linear equations. This
// will fail if IsSingular() returns true.
func LU(a *Dense) LUFactors {
	// Use a "left-looking", dot-product, Crout/Doolittle algorithm.
	m, n := a.Dims()
	lu := a

	piv := make([]int, m)
	for i := range piv {
		piv[i] = i
	}
	sign := 1

	// Outer loop.
	luColj := make([]float64, m)
	for j := 0; j < n; j++ {

		// Make a copy of the j-th column to localize references.
		for i := 0; i < m; i++ {
			luColj[i] = lu.at(i, j)
		}

		// Apply previous transformations.
		for i := 0; i < m; i++ {
			luRowi := lu.RawRowView(i)

			// Most of the time is spent in the following dot product.
			kmax := min(i, j)
			var s float64
			for k, v := range luRowi[:kmax] {
				s += v * luColj[k]
			}

			luColj[i] -= s
			luRowi[j] = luColj[i]
		}

		// Find pivot and exchange if necessary.
		p := j
		for i := j + 1; i < m; i++ {
			if math.Abs(luColj[i]) > math.Abs(luColj[p]) {
				p = i
			}
		}
		if p != j {
			for k := 0; k < n; k++ {
				t := lu.at(p, k)
				lu.set(p, k, lu.at(j, k))
				lu.set(j, k, t)
			}
			piv[p], piv[j] = piv[j], piv[p]
			sign = -sign
		}

		// Compute multipliers.
		if j < m && lu.at(j, j) != 0 {
			for i := j + 1; i < m; i++ {
				lu.set(i, j, lu.at(i, j)/lu.at(j, j))
			}
		}
	}

	return LUFactors{lu, piv, sign}
}

// IsSingular returns whether the the upper triangular factor and hence a is
// singular.
func (f LUFactors) IsSingular() bool {
	lu := f.LU
	_, n := lu.Dims()
	for j := 0; j < n; j++ {
		if lu.at(j, j) == 0 {
			return true
		}
	}
	return false
}

// L returns the lower triangular factor of the LU decomposition.
func (f LUFactors) L() *Dense {
	lu := f.LU
	m, n := lu.Dims()
	if m < n {
		n = m
	}
	l := NewDense(m, n, nil)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if i > j {
				l.set(i, j, lu.at(i, j))
			} else if i == j {
				l.set(i, j, 1)
			}
		}
	}
	return l
}

// U returns the upper triangular factor of the LU decomposition.
func (f LUFactors) U() *Dense {
	lu := f.LU
	m, n := lu.Dims()
	u := NewDense(m, n, nil)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if i <= j {
				u.set(i, j, lu.at(i, j))
			}
		}
	}
	return u
}

// Det returns the determinant of matrix a decomposed into lu. The matrix
// a must have been square.
func (f LUFactors) Det() float64 {
	lu, sign := f.LU, f.Sign
	m, n := lu.Dims()
	if m != n {
		panic(ErrSquare)
	}
	d := float64(sign)
	for j := 0; j < n; j++ {
		d *= lu.at(j, j)
	}
	return d
}

// Solve computes a solution of a.x = b where b has as many rows as a. A matrix x
// is returned that minimizes the two norm of L*U*X - B(piv,:). Solve will panic
// if a is singular. The matrix b is overwritten during the call.
func (f LUFactors) Solve(b *Dense) (x *Dense) {
	lu, piv := f.LU, f.Pivot
	m, n := lu.Dims()
	bm, bn := b.Dims()
	if bm != m {
		panic(ErrShape)
	}
	if f.IsSingular() {
		panic(ErrSingular)
	}

	// Copy right hand side with pivoting
	nx := bn
	x = pivotRows(b, piv)

	// Solve L*Y = B(piv,:)
	for k := 0; k < n; k++ {
		for i := k + 1; i < n; i++ {
			for j := 0; j < nx; j++ {
				x.set(i, j, x.at(i, j)-x.at(k, j)*lu.at(i, k))
			}
		}
	}

	// Solve U*X = Y;
	for k := n - 1; k >= 0; k-- {
		for j := 0; j < nx; j++ {
			x.set(k, j, x.at(k, j)/lu.at(k, k))
		}
		for i := 0; i < k; i++ {
			for j := 0; j < nx; j++ {
				x.set(i, j, x.at(i, j)-x.at(k, j)*lu.at(i, k))
			}
		}
	}

	return x
}

func pivotRows(a *Dense, piv []int) *Dense {
	visit := make([]bool, len(piv))
	_, n := a.Dims()
	tmpRow := make([]float64, n)
	for to, from := range piv {
		for to != from && !visit[from] {
			visit[from], visit[to] = true, true
			a.Row(tmpRow, from)
			a.SetRow(from, a.rowView(to))
			a.SetRow(to, tmpRow)
			to, from = from, piv[from]
		}
	}
	return a
}
