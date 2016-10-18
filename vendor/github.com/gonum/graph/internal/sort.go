// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

// BySliceValues implements the sort.Interface sorting a slice of
// []int lexically by the values of the []int.
type BySliceValues [][]int

func (c BySliceValues) Len() int { return len(c) }
func (c BySliceValues) Less(i, j int) bool {
	a, b := c[i], c[j]
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for k, v := range a[:l] {
		if v < b[k] {
			return true
		}
		if v > b[k] {
			return false
		}
	}
	return len(a) < len(b)
}
func (c BySliceValues) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
