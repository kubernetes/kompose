// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concrete

import (
	"math"

	"github.com/gonum/graph"
)

type nodeSorter []graph.Node

func (ns nodeSorter) Less(i, j int) bool {
	return ns[i].ID() < ns[j].ID()
}

func (ns nodeSorter) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns nodeSorter) Len() int {
	return len(ns)
}

// The math package only provides explicitly sized max
// values. This ensures we get the max for the actual
// type int.
const maxInt int = int(^uint(0) >> 1)

var inf = math.Inf(1)

func isSame(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
