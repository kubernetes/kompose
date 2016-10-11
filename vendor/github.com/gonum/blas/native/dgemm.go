// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/gonum/blas"
	"github.com/gonum/internal/asm"
)

// Dgemm computes
//  C = beta * C + alpha * A * B.
// tA and tB specify whether A or B are transposed. A, B, and C are m×n dense
// matrices.
func (Implementation) Dgemm(tA, tB blas.Transpose, m, n, k int, alpha float64, a []float64, lda int, b []float64, ldb int, beta float64, c []float64, ldc int) {
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if tB != blas.NoTrans && tB != blas.Trans && tB != blas.ConjTrans {
		panic(badTranspose)
	}

	var amat, bmat, cmat general64
	if tA != blas.NoTrans {
		amat = general64{
			data:   a,
			rows:   k,
			cols:   m,
			stride: lda,
		}
	} else {
		amat = general64{
			data:   a,
			rows:   m,
			cols:   k,
			stride: lda,
		}
	}
	err := amat.check('a')
	if err != nil {
		panic(err.Error())
	}
	if tB != blas.NoTrans {
		bmat = general64{
			data:   b,
			rows:   n,
			cols:   k,
			stride: ldb,
		}
	} else {
		bmat = general64{
			data:   b,
			rows:   k,
			cols:   n,
			stride: ldb,
		}
	}

	err = bmat.check('b')
	if err != nil {
		panic(err.Error())
	}
	cmat = general64{
		data:   c,
		rows:   m,
		cols:   n,
		stride: ldc,
	}
	err = cmat.check('c')
	if err != nil {
		panic(err.Error())
	}

	// scale c
	if beta != 1 {
		if beta == 0 {
			for i := 0; i < m; i++ {
				ctmp := cmat.data[i*cmat.stride : i*cmat.stride+cmat.cols]
				for j := range ctmp {
					ctmp[j] = 0
				}
			}
		} else {
			for i := 0; i < m; i++ {
				ctmp := cmat.data[i*cmat.stride : i*cmat.stride+cmat.cols]
				for j := range ctmp {
					ctmp[j] *= beta
				}
			}
		}
	}

	dgemmParallel(tA, tB, amat, bmat, cmat, alpha)
}

func dgemmParallel(tA, tB blas.Transpose, a, b, c general64, alpha float64) {
	// dgemmParallel computes a parallel matrix multiplication by partitioning
	// a and b into sub-blocks, and updating c with the multiplication of the sub-block
	// In all cases,
	// A = [ 	A_11	A_12 ... 	A_1j
	//			A_21	A_22 ...	A_2j
	//				...
	//			A_i1	A_i2 ...	A_ij]
	//
	// and same for B. All of the submatrix sizes are blockSize*blockSize except
	// at the edges.
	// In all cases, there is one dimension for each matrix along which
	// C must be updated sequentially.
	// Cij = \sum_k Aik Bki,	(A * B)
	// Cij = \sum_k Aki Bkj,	(A^T * B)
	// Cij = \sum_k Aik Bjk,	(A * B^T)
	// Cij = \sum_k Aki Bjk,	(A^T * B^T)
	//
	// This code computes one {i, j} block sequentially along the k dimension,
	// and computes all of the {i, j} blocks concurrently. This
	// partitioning allows Cij to be updated in-place without race-conditions.
	// Instead of launching a goroutine for each possible concurrent computation,
	// a number of worker goroutines are created and channels are used to pass
	// available and completed cases.
	//
	// http://alexkr.com/docs/matrixmult.pdf is a good reference on matrix-matrix
	// multiplies, though this code does not copy matrices to attempt to eliminate
	// cache misses.

	aTrans := tA == blas.Trans || tA == blas.ConjTrans
	bTrans := tB == blas.Trans || tB == blas.ConjTrans

	maxKLen, parBlocks := computeNumBlocks64(a, b, aTrans, bTrans)
	if parBlocks < minParBlock {
		// The matrix multiplication is small in the dimensions where it can be
		// computed concurrently. Just do it in serial.
		dgemmSerial(tA, tB, a, b, c, alpha)
		return
	}

	nWorkers := runtime.GOMAXPROCS(0)
	if parBlocks < nWorkers {
		nWorkers = parBlocks
	}
	// There is a tradeoff between the workers having to wait for work
	// and a large buffer making operations slow.
	buf := buffMul * nWorkers
	if buf > parBlocks {
		buf = parBlocks
	}

	sendChan := make(chan subMul, buf)

	// Launch workers. A worker receives an {i, j} submatrix of c, and computes
	// A_ik B_ki (or the transposed version) storing the result in c_ij. When the
	// channel is finally closed, it signals to the waitgroup that it has finished
	// computing.
	var wg sync.WaitGroup
	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Make local copies of otherwise global variables to reduce shared memory.
			// This has a noticable effect on benchmarks in some cases.
			alpha := alpha
			aTrans := aTrans
			bTrans := bTrans
			crows := c.rows
			ccols := c.cols
			for sub := range sendChan {
				i := sub.i
				j := sub.j
				leni := blockSize
				if i+leni > crows {
					leni = crows - i
				}
				lenj := blockSize
				if j+lenj > ccols {
					lenj = ccols - j
				}
				cSub := c.view(i, j, leni, lenj)

				// Compute A_ik B_kj for all k
				for k := 0; k < maxKLen; k += blockSize {
					lenk := blockSize
					if k+lenk > maxKLen {
						lenk = maxKLen - k
					}
					var aSub, bSub general64
					if aTrans {
						aSub = a.view(k, i, lenk, leni)
					} else {
						aSub = a.view(i, k, leni, lenk)
					}
					if bTrans {
						bSub = b.view(j, k, lenj, lenk)
					} else {
						bSub = b.view(k, j, lenk, lenj)
					}

					dgemmSerial(tA, tB, aSub, bSub, cSub, alpha)
				}
			}
		}()
	}

	// Send out all of the {i, j} subblocks for computation.
	for i := 0; i < c.rows; i += blockSize {
		for j := 0; j < c.cols; j += blockSize {
			sendChan <- subMul{
				i: i,
				j: j,
			}
		}
	}
	close(sendChan)
	wg.Wait()
}

// computeNumBlocks says how many blocks there are to compute. maxKLen says the length of the
// k dimension, parBlocks is the number of blocks that could be computed in parallel
// (the submatrices in i and j). expect is the full number of blocks that will be computed.
func computeNumBlocks64(a, b general64, aTrans, bTrans bool) (maxKLen, parBlocks int) {
	aRowBlocks := a.rows / blockSize
	if a.rows%blockSize != 0 {
		aRowBlocks++
	}
	aColBlocks := a.cols / blockSize
	if a.cols%blockSize != 0 {
		aColBlocks++
	}
	bRowBlocks := b.rows / blockSize
	if b.rows%blockSize != 0 {
		bRowBlocks++
	}
	bColBlocks := b.cols / blockSize
	if b.cols%blockSize != 0 {
		bColBlocks++
	}

	switch {
	case !aTrans && !bTrans:
		// Cij = \sum_k Aik Bki
		maxKLen = a.cols
		parBlocks = aRowBlocks * bColBlocks
	case aTrans && !bTrans:
		// Cij = \sum_k Aki Bkj
		maxKLen = a.rows
		parBlocks = aColBlocks * bColBlocks
	case !aTrans && bTrans:
		// Cij = \sum_k Aik Bjk
		maxKLen = a.cols
		parBlocks = aRowBlocks * bRowBlocks
	case aTrans && bTrans:
		// Cij = \sum_k Aki Bjk
		maxKLen = a.rows
		parBlocks = aColBlocks * bRowBlocks
	}
	return
}

// dgemmSerial is serial matrix multiply
func dgemmSerial(tA, tB blas.Transpose, a, b, c general64, alpha float64) {
	switch {
	case tA == blas.NoTrans && tB == blas.NoTrans:
		dgemmSerialNotNot(a, b, c, alpha)
		return
	case tA != blas.NoTrans && tB == blas.NoTrans:
		dgemmSerialTransNot(a, b, c, alpha)
		return
	case tA == blas.NoTrans && tB != blas.NoTrans:
		dgemmSerialNotTrans(a, b, c, alpha)
		return
	case tA != blas.NoTrans && tB != blas.NoTrans:
		dgemmSerialTransTrans(a, b, c, alpha)
		return
	default:
		panic("unreachable")
	}
}

// dgemmSerial where neither a nor b are transposed
func dgemmSerialNotNot(a, b, c general64, alpha float64) {
	if debug {
		if a.cols != b.rows {
			panic("inner dimension mismatch")
		}
		if a.rows != c.rows {
			panic("outer dimension mismatch")
		}
		if b.cols != c.cols {
			panic("outer dimension mismatch")
		}
	}

	// This style is used instead of the literal [i*stride +j]) is used because
	// approximately 5 times faster as of go 1.3.
	for i := 0; i < a.rows; i++ {
		ctmp := c.data[i*c.stride : i*c.stride+c.cols]
		for l, v := range a.data[i*a.stride : i*a.stride+a.cols] {
			tmp := alpha * v
			if tmp != 0 {
				asm.DaxpyUnitary(tmp, b.data[l*b.stride:l*b.stride+b.cols], ctmp, ctmp)
			}
		}
	}
}

// dgemmSerial where neither a is transposed and b is not
func dgemmSerialTransNot(a, b, c general64, alpha float64) {
	if debug {
		if a.rows != b.rows {
			fmt.Println(a.rows, b.rows)
			panic("inner dimension mismatch")
		}
		if a.cols != c.rows {
			panic("outer dimension mismatch")
		}
		if b.cols != c.cols {
			panic("outer dimension mismatch")
		}
	}

	// This style is used instead of the literal [i*stride +j]) is used because
	// approximately 5 times faster as of go 1.3.
	for l := 0; l < a.rows; l++ {
		btmp := b.data[l*b.stride : l*b.stride+b.cols]
		for i, v := range a.data[l*a.stride : l*a.stride+a.cols] {
			tmp := alpha * v
			ctmp := c.data[i*c.stride : i*c.stride+c.cols]
			if tmp != 0 {
				asm.DaxpyUnitary(tmp, btmp, ctmp, ctmp)
			}
		}
	}
}

// dgemmSerial where neither a is not transposed and b is
func dgemmSerialNotTrans(a, b, c general64, alpha float64) {
	if debug {
		if a.cols != b.cols {
			panic("inner dimension mismatch")
		}
		if a.rows != c.rows {
			panic("outer dimension mismatch")
		}
		if b.rows != c.cols {
			panic("outer dimension mismatch")
		}
	}

	// This style is used instead of the literal [i*stride +j]) is used because
	// approximately 5 times faster as of go 1.3.
	for i := 0; i < a.rows; i++ {
		atmp := a.data[i*a.stride : i*a.stride+a.cols]
		ctmp := c.data[i*c.stride : i*c.stride+c.cols]
		for j := 0; j < b.rows; j++ {
			ctmp[j] += alpha * asm.DdotUnitary(atmp, b.data[j*b.stride:j*b.stride+b.cols])
		}
	}

}

// dgemmSerial where both are transposed
func dgemmSerialTransTrans(a, b, c general64, alpha float64) {
	if debug {
		if a.rows != b.cols {
			panic("inner dimension mismatch")
		}
		if a.cols != c.rows {
			panic("outer dimension mismatch")
		}
		if b.rows != c.cols {
			panic("outer dimension mismatch")
		}
	}

	// This style is used instead of the literal [i*stride +j]) is used because
	// approximately 5 times faster as of go 1.3.
	for l := 0; l < a.rows; l++ {
		for i, v := range a.data[l*a.stride : l*a.stride+a.cols] {
			ctmp := c.data[i*c.stride : i*c.stride+c.cols]
			if v != 0 {
				tmp := alpha * v
				if tmp != 0 {
					asm.DaxpyInc(tmp, b.data[l:], ctmp, uintptr(b.rows), uintptr(b.stride), 1, 0, 0)
				}
			}
		}
	}
}
