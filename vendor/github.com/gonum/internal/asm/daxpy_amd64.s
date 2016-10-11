// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Some of the loop unrolling code is copied from:
// http://golang.org/src/math/big/arith_amd64.s
// which is distributed under these terms:
//
// Copyright (c) 2012 The Go Authors. All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
// 
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//+build !noasm

// TODO(fhs): use textflag.h after we drop Go 1.3 support
//#include "textflag.h"
// Don't insert stack check preamble.
#define NOSPLIT	4


// func DaxpyUnitary(alpha float64, x, y, z []float64)
// This function assumes len(y) >= len(x).
TEXT ·DaxpyUnitary(SB),NOSPLIT,$0
	MOVHPD alpha+0(FP), X7
	MOVLPD alpha+0(FP), X7
	MOVQ x_len+16(FP), DI	// n = len(x)
	MOVQ x+8(FP), R8
	MOVQ y+32(FP), R9
	MOVQ z+56(FP), R10
	
	MOVQ $0, SI				// i = 0
	SUBQ $2, DI				// n -= 2
	JL V1					// if n < 0 goto V1

U1:	// n >= 0
	// y[i] += alpha * x[i] unrolled 2x.
	MOVUPD 0(R8)(SI*8), X0
	MOVUPD 0(R9)(SI*8), X1
	MULPD X7, X0
	ADDPD X0, X1
	MOVUPD X1, 0(R10)(SI*8)
	
	ADDQ $2, SI				// i += 2
	SUBQ $2, DI				// n -= 2
	JGE U1					// if n >= 0 goto U1

V1:
	ADDQ $2, DI				// n += 2
	JLE E1					// if n <= 0 goto E1
	
	// y[i] += alpha * x[i] for last iteration if n is odd.
	MOVSD 0(R8)(SI*8), X0
	MOVSD 0(R9)(SI*8), X1
	MULSD X7, X0
	ADDSD X0, X1
	MOVSD X1, 0(R10)(SI*8)

E1:
	RET


// func DaxpyInc(alpha float64, x, y []float64, n, incX, incY, ix, iy uintptr)
TEXT ·DaxpyInc(SB),NOSPLIT,$0
	MOVHPD alpha+0(FP), X7
	MOVLPD alpha+0(FP), X7
	MOVQ x+8(FP), R8
	MOVQ y+32(FP), R9
	MOVQ n+56(FP), CX
	MOVQ incX+64(FP), R11
	MOVQ incY+72(FP), R12
	MOVQ ix+80(FP), SI
	MOVQ iy+88(FP), DI

	MOVQ SI, AX				// nextX = ix
	MOVQ DI, BX				// nextY = iy
	ADDQ R11, AX			// nextX += incX
	ADDQ R12, BX			// nextY += incX
	SHLQ $1, R11			// indX *= 2
	SHLQ $1, R12			// indY *= 2
	
	SUBQ $2, CX				// n -= 2
	JL V2					// if n < 0 goto V2

U2:	// n >= 0
	// y[i] += alpha * x[i] unrolled 2x.
	MOVHPD 0(R8)(SI*8), X0
	MOVHPD 0(R9)(DI*8), X1
	MOVLPD 0(R8)(AX*8), X0
	MOVLPD 0(R9)(BX*8), X1
	
	MULPD X7, X0
	ADDPD X0, X1
	MOVHPD X1, 0(R9)(DI*8)
	MOVLPD X1, 0(R9)(BX*8)

	ADDQ R11, SI			// ix += incX
	ADDQ R12, DI			// iy += incY
	ADDQ R11, AX			// nextX += incX
	ADDQ R12, BX			// nextY += incY

	SUBQ $2, CX				// n -= 2
	JGE U2					// if n >= 0 goto U2

V2:
	ADDQ $2, CX				// n += 2
	JLE E2					// if n <= 0 goto E2
	
	// y[i] += alpha * x[i] for the last iteration if n is odd.
	MOVSD 0(R8)(SI*8), X0
	MOVSD 0(R9)(DI*8), X1
	MULSD X7, X0
	ADDSD X0, X1
	MOVSD X1, 0(R9)(DI*8)

E2:
	RET
