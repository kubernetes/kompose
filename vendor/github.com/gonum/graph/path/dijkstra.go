// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"container/heap"

	"github.com/gonum/graph"
)

// DijkstraFrom returns a shortest-path tree for a shortest path from u to all nodes in
// the graph g. If the graph does not implement graph.Weighter, graph.UniformCost is used.
// DijkstraFrom will panic if g has a u-reachable negative edge weight.
//
// The time complexity of DijkstrFrom is O(|E|+|V|.log|V|).
func DijkstraFrom(u graph.Node, g graph.Graph) Shortest {
	if !g.Has(u) {
		return Shortest{from: u}
	}
	var weight graph.WeightFunc
	if g, ok := g.(graph.Weighter); ok {
		weight = g.Weight
	} else {
		weight = graph.UniformCost
	}

	nodes := g.Nodes()
	path := newShortestFrom(u, nodes)

	// Dijkstra's algorithm here is implemented essentially as
	// described in Function B.2 in figure 6 of UTCS Technical
	// Report TR-07-54.
	//
	// http://www.cs.utexas.edu/ftp/techreports/tr07-54.pdf
	Q := priorityQueue{{node: u, dist: 0}}
	for Q.Len() != 0 {
		mid := heap.Pop(&Q).(distanceNode)
		k := path.indexOf[mid.node.ID()]
		if mid.dist < path.dist[k] {
			path.dist[k] = mid.dist
		}
		for _, v := range g.From(mid.node) {
			j := path.indexOf[v.ID()]
			w := weight(g.Edge(mid.node, v))
			if w < 0 {
				panic("dijkstra: negative edge weight")
			}
			joint := path.dist[k] + w
			if joint < path.dist[j] {
				heap.Push(&Q, distanceNode{node: v, dist: joint})
				path.set(j, joint, k)
			}
		}
	}

	return path
}

// DijkstraAllPaths returns a shortest-path tree for shortest paths in the graph g.
// If the graph does not implement graph.Weighter, graph.UniformCost is used.
// DijkstraAllPaths will panic if g has a negative edge weight.
//
// The time complexity of DijkstrAllPaths is O(|V|.|E|+|V|^2.log|V|).
func DijkstraAllPaths(g graph.Graph) (paths AllShortest) {
	paths = newAllShortest(g.Nodes(), false)
	dijkstraAllPaths(g, paths)
	return paths
}

// dijkstraAllPaths is the all-paths implementation of Dijkstra. It is shared
// between DijkstraAllPaths and JohnsonAllPaths to avoid repeated allocation
// of the nodes slice and the indexOf map. It returns nothing, but stores the
// result of the work in the paths parameter which is a reference type.
func dijkstraAllPaths(g graph.Graph, paths AllShortest) {
	var weight graph.WeightFunc
	if g, ok := g.(graph.Weighter); ok {
		weight = g.Weight
	} else {
		weight = graph.UniformCost
	}

	var Q priorityQueue
	for i, u := range paths.nodes {
		// Dijkstra's algorithm here is implemented essentially as
		// described in Function B.2 in figure 6 of UTCS Technical
		// Report TR-07-54 with the addition of handling multiple
		// co-equal paths.
		//
		// http://www.cs.utexas.edu/ftp/techreports/tr07-54.pdf

		// Q must be empty at this point.
		heap.Push(&Q, distanceNode{node: u, dist: 0})
		for Q.Len() != 0 {
			mid := heap.Pop(&Q).(distanceNode)
			k := paths.indexOf[mid.node.ID()]
			if mid.dist < paths.dist.At(i, k) {
				paths.dist.Set(i, k, mid.dist)
			}
			for _, v := range g.From(mid.node) {
				j := paths.indexOf[v.ID()]
				w := weight(g.Edge(mid.node, v))
				if w < 0 {
					panic("dijkstra: negative edge weight")
				}
				joint := paths.dist.At(i, k) + w
				if joint < paths.dist.At(i, j) {
					heap.Push(&Q, distanceNode{node: v, dist: joint})
					paths.set(i, j, joint, k)
				} else if joint == paths.dist.At(i, j) {
					paths.add(i, j, k)
				}
			}
		}
	}
}

type distanceNode struct {
	node graph.Node
	dist float64
}

// priorityQueue implements a no-dec priority queue.
type priorityQueue []distanceNode

func (q priorityQueue) Len() int            { return len(q) }
func (q priorityQueue) Less(i, j int) bool  { return q[i].dist < q[j].dist }
func (q priorityQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *priorityQueue) Push(n interface{}) { *q = append(*q, n.(distanceNode)) }
func (q *priorityQueue) Pop() interface{} {
	t := *q
	var n interface{}
	n, *q = t[len(t)-1], t[:len(t)-1]
	return n
}
