// Copyright ©2013 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat64

import (
	"math"

	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
)

var (
	vector *Vector

	_ Matrix = vector

	// _ Cloner      = vector
	// _ Viewer      = vector
	// _ Subvectorer = vector

	// _ Adder     = vector
	// _ Suber     = vector
	// _ Muler = vector
	// _ Dotter    = vector
	// _ ElemMuler = vector

	// _ Scaler  = vector
	// _ Applyer = vector

	// _ Normer = vector
	// _ Sumer  = vector

	// _ Stacker   = vector
	// _ Augmenter = vector

	// _ Equaler       = vector
	// _ ApproxEqualer = vector

	// _ RawMatrixLoader = vector
	// _ RawMatrixer     = vector
)

// Vector represents a column vector.
type Vector struct {
	mat blas64.Vector
	n   int
	// A BLAS vector can have a negative increment, but allowing this
	// in the mat64 type complicates a lot of code, and doesn't gain anything.
	// Vector must have positive increment in this package.
}

// NewVector creates a new Vector of length n. If len(data) == n, data is used
// as the backing data slice. If data == nil, a new slice is allocated. If
// neither of these is true, NewVector will panic.
func NewVector(n int, data []float64) *Vector {
	if len(data) != n && data != nil {
		panic(ErrShape)
	}
	if data == nil {
		data = make([]float64, n)
	}
	return &Vector{
		mat: blas64.Vector{
			Inc:  1,
			Data: data,
		},
		n: n,
	}
}

// ViewVec returns a sub-vector view of the receiver starting at element i and
// extending n columns. If i is out of range, or if n is zero or extend beyond the
// bounds of the Vector ViewVec will panic with ErrIndexOutOfRange. The returned
// Vector retains reference to the underlying vector.
func (v *Vector) ViewVec(i, n int) *Vector {
	if i+n > v.n {
		panic(ErrIndexOutOfRange)
	}
	return &Vector{
		n: n,
		mat: blas64.Vector{
			Inc:  v.mat.Inc,
			Data: v.mat.Data[i*v.mat.Inc:],
		},
	}
}

func (v *Vector) Dims() (r, c int) { return v.n, 1 }

// Len returns the length of the vector.
func (v *Vector) Len() int {
	return v.n
}

func (v *Vector) Reset() {
	v.mat.Data = v.mat.Data[:0]
	v.mat.Inc = 0
	v.n = 0
}

func (v *Vector) RawVector() blas64.Vector {
	return v.mat
}

// CopyVec makes a copy of elements of a into the receiver. It is similar to the
// built-in copy; it copies as much as the overlap between the two matrices and
// returns the number of rows and columns it copied.
func (v *Vector) CopyVec(a *Vector) (n int) {
	n = min(v.Len(), a.Len())
	blas64.Copy(n, a.mat, v.mat)
	return n
}

// AddVec adds a and b element-wise, placing the result in the receiver.
func (v *Vector) AddVec(a, b *Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	amat, bmat := a.RawVector(), b.RawVector()
	for i := 0; i < v.n; i++ {
		v.mat.Data[i*v.mat.Inc] = amat.Data[i*amat.Inc] + bmat.Data[i*bmat.Inc]
	}
}

// SubVec subtracts the vector b from a, placing the result in the receiver.
func (v *Vector) SubVec(a, b *Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	amat, bmat := a.RawVector(), b.RawVector()
	for i := 0; i < v.n; i++ {
		v.mat.Data[i*v.mat.Inc] = amat.Data[i*amat.Inc] - bmat.Data[i*bmat.Inc]
	}
}

// MulElemVec performs element-wise multiplication of a and b, placing the result
// in the receiver.
func (v *Vector) MulElemVec(a, b *Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	amat, bmat := a.RawVector(), b.RawVector()
	for i := 0; i < v.n; i++ {
		v.mat.Data[i*v.mat.Inc] = amat.Data[i*amat.Inc] * bmat.Data[i*bmat.Inc]
	}
}

// DivElemVec performs element-wise division of a by b, placing the result
// in the receiver.
func (v *Vector) DivElemVec(a, b *Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	amat, bmat := a.RawVector(), b.RawVector()
	for i := 0; i < v.n; i++ {
		v.mat.Data[i*v.mat.Inc] = amat.Data[i*amat.Inc] / bmat.Data[i*bmat.Inc]
	}
}

// MulVec computes a * b if trans == false and a^T * b if trans == true. The
// result is stored into the receiver. MulVec panics if the number of columns in
// a does not equal the number of rows in b.
func (v *Vector) MulVec(a Matrix, trans bool, b *Vector) {
	ar, ac := a.Dims()
	br := b.Len()
	if trans {
		if ar != br {
			panic(ErrShape)
		}
	} else {
		if ac != br {
			panic(ErrShape)
		}
	}

	var w Vector
	if v != a && v != b {
		w = *v
	}
	if w.n == 0 {
		if trans {
			w.mat.Data = use(w.mat.Data, ac)
		} else {
			w.mat.Data = use(w.mat.Data, ar)
		}

		w.mat.Inc = 1
		w.n = ar
		if trans {
			w.n = ac
		}
	} else {
		if trans {
			if ac != w.n {
				panic(ErrShape)
			}
		} else {
			if ar != w.n {
				panic(ErrShape)
			}
		}
	}

	switch a := a.(type) {
	case RawSymmetricer:
		amat := a.RawSymmetric()
		blas64.Symv(1, amat, b.mat, 0, w.mat)
	case RawTriangular:
		w.CopyVec(b)
		amat := a.RawTriangular()
		ta := blas.NoTrans
		if trans {
			ta = blas.Trans
		}
		blas64.Trmv(ta, amat, w.mat)
	case RawMatrixer:
		amat := a.RawMatrix()
		t := blas.NoTrans
		if trans {
			t = blas.Trans
		}
		blas64.Gemv(t, 1, amat, b.mat, 0, w.mat)
	case Vectorer:
		if trans {
			col := make([]float64, ar)
			for c := 0; c < ac; c++ {
				w.mat.Data[c*w.mat.Inc] = blas64.Dot(ar,
					blas64.Vector{Inc: 1, Data: a.Col(col, c)},
					b.mat,
				)
			}
		} else {
			row := make([]float64, ac)
			for r := 0; r < ar; r++ {
				w.mat.Data[r*w.mat.Inc] = blas64.Dot(ac,
					blas64.Vector{Inc: 1, Data: a.Row(row, r)},
					b.mat,
				)
			}
		}
	default:
		if trans {
			col := make([]float64, ar)
			for c := 0; c < ac; c++ {
				for i := range col {
					col[i] = a.At(i, c)
				}
				var f float64
				for i, e := range col {
					f += e * b.mat.Data[i*b.mat.Inc]
				}
				w.mat.Data[c*w.mat.Inc] = f
			}
		} else {
			row := make([]float64, ac)
			for r := 0; r < ar; r++ {
				for i := range row {
					row[i] = a.At(r, i)
				}
				var f float64
				for i, e := range row {
					f += e * b.mat.Data[i*b.mat.Inc]
				}
				w.mat.Data[r*w.mat.Inc] = f
			}
		}
	}
	*v = w
}

// Equals compares the vectors represented by b and the receiver and returns true
// if the vectors are element-wise equal.
func (v *Vector) EqualsVec(b *Vector) bool {
	n := v.Len()
	nb := b.Len()
	if n != nb {
		return false
	}
	for i := 0; i < n; i++ {
		if v.mat.Data[i*v.mat.Inc] != b.mat.Data[i*b.mat.Inc] {
			return false
		}
	}
	return true
}

// EqualsApproxVec compares the vectors represented by b and the receiver, with
// tolerance for element-wise equality specified by epsilon.
func (v *Vector) EqualsApproxVec(b *Vector, epsilon float64) bool {
	n := v.Len()
	nb := b.Len()
	if n != nb {
		return false
	}
	for i := 0; i < n; i++ {
		if math.Abs(v.mat.Data[i*v.mat.Inc]-b.mat.Data[i*b.mat.Inc]) > epsilon {
			return false
		}
	}
	return true
}

// reuseAs resizes an empty vector to a r×1 vector,
// or checks that a non-empty matrix is r×1.
func (v *Vector) reuseAs(r int) {
	if v.isZero() {
		v.mat = blas64.Vector{
			Inc:  1,
			Data: use(v.mat.Data, r),
		}
		v.n = r
		return
	}
	if r != v.n {
		panic(ErrShape)
	}
}

func (v *Vector) isZero() bool {
	// It must be the case that v.Dims() returns
	// zeros in this case. See comment in Reset().
	return v.mat.Inc == 0
}
