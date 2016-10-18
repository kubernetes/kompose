// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import "github.com/gonum/graph"

// FloydWarshall returns a shortest-path tree for the graph g or false indicating
// that a negative cycle exists in the graph. If the graph does not implement
// graph.Weighter, graph.UniformCost is used.
//
// The time complexity of FloydWarshall is O(|V|^3).
func FloydWarshall(g graph.Graph) (paths AllShortest, ok bool) {
	var weight graph.WeightFunc
	if g, ok := g.(graph.Weighter); ok {
		weight = g.Weight
	} else {
		weight = graph.UniformCost
	}

	nodes := g.Nodes()
	paths = newAllShortest(nodes, true)
	for i, u := range nodes {
		paths.dist.Set(i, i, 0)
		for _, v := range g.From(u) {
			j := paths.indexOf[v.ID()]
			paths.set(i, j, weight(g.Edge(u, v)), j)
		}
	}

	for k := range nodes {
		for i := range nodes {
			for j := range nodes {
				ij := paths.dist.At(i, j)
				joint := paths.dist.At(i, k) + paths.dist.At(k, j)
				if ij > joint {
					paths.set(i, j, joint, paths.at(i, k)...)
				} else if ij-joint == 0 {
					paths.add(i, j, paths.at(i, k)...)
				}
			}
		}
	}

	ok = true
	for i := range nodes {
		if paths.dist.At(i, i) < 0 {
			ok = false
			break
		}
	}

	return paths, ok
}
