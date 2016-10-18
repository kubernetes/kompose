// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"container/heap"

	"github.com/gonum/graph"
	"github.com/gonum/graph/internal"
)

// Heuristic returns an estimate of the cost of travelling between two nodes.
type Heuristic func(x, y graph.Node) float64

// HeuristicCoster wraps the HeuristicCost method. A graph implementing the
// interface provides a heuristic between any two given nodes.
type HeuristicCoster interface {
	HeuristicCost(x, y graph.Node) float64
}

// AStar finds the A*-shortest path from s to t in g using the heuristic h. The path and
// its cost are returned in a Shortest along with paths and costs to all nodes explored
// during the search. The number of expanded nodes is also returned. This value may help
// with heuristic tuning.
//
// The path will be the shortest path if the heuristic is admissible. A heuristic is
// admissible if for any node, n, in the graph, the heuristic estimate of the cost of
// the path from n to t is less than or equal to the true cost of that path.
//
// If h is nil, AStar will use the g.HeuristicCost method if g implements HeuristicCoster,
// falling back to NullHeuristic otherwise. If the graph does not implement graph.Weighter,
// graph.UniformCost is used. AStar will panic if g has an A*-reachable negative edge weight.
func AStar(s, t graph.Node, g graph.Graph, h Heuristic) (path Shortest, expanded int) {
	if !g.Has(s) || !g.Has(t) {
		return Shortest{from: s}, 0
	}
	var weight graph.WeightFunc
	if g, ok := g.(graph.Weighter); ok {
		weight = g.Weight
	} else {
		weight = graph.UniformCost
	}
	if h == nil {
		if g, ok := g.(HeuristicCoster); ok {
			h = g.HeuristicCost
		} else {
			h = NullHeuristic
		}
	}

	path = newShortestFrom(s, g.Nodes())
	tid := t.ID()

	visited := make(internal.IntSet)
	open := &aStarQueue{indexOf: make(map[int]int)}
	heap.Push(open, aStarNode{node: s, gscore: 0, fscore: h(s, t)})

	for open.Len() != 0 {
		u := heap.Pop(open).(aStarNode)
		uid := u.node.ID()
		i := path.indexOf[uid]
		expanded++

		if uid == tid {
			break
		}

		visited.Add(uid)
		for _, v := range g.From(u.node) {
			vid := v.ID()
			if visited.Has(vid) {
				continue
			}
			j := path.indexOf[vid]

			w := weight(g.Edge(u.node, v))
			if w < 0 {
				panic("A*: negative edge weight")
			}
			g := u.gscore + w
			if n, ok := open.node(vid); !ok {
				path.set(j, g, i)
				heap.Push(open, aStarNode{node: v, gscore: g, fscore: g + h(v, t)})
			} else if g < n.gscore {
				path.set(j, g, i)
				open.update(vid, g, g+h(v, t))
			}
		}
	}

	return path, expanded
}

// NullHeuristic is an admissible, consistent heuristic that will not speed up computation.
func NullHeuristic(_, _ graph.Node) float64 {
	return 0
}

// aStarNode adds A* accounting to a graph.Node.
type aStarNode struct {
	node   graph.Node
	gscore float64
	fscore float64
}

// aStarQueue is an A* priority queue.
type aStarQueue struct {
	indexOf map[int]int
	nodes   []aStarNode
}

func (q *aStarQueue) Less(i, j int) bool {
	return q.nodes[i].fscore < q.nodes[j].fscore
}

func (q *aStarQueue) Swap(i, j int) {
	q.indexOf[q.nodes[i].node.ID()] = j
	q.indexOf[q.nodes[j].node.ID()] = i
	q.nodes[i], q.nodes[j] = q.nodes[j], q.nodes[i]
}

func (q *aStarQueue) Len() int {
	return len(q.nodes)
}

func (q *aStarQueue) Push(x interface{}) {
	n := x.(aStarNode)
	q.indexOf[n.node.ID()] = len(q.nodes)
	q.nodes = append(q.nodes, n)
}

func (q *aStarQueue) Pop() interface{} {
	n := q.nodes[len(q.nodes)-1]
	q.nodes = q.nodes[:len(q.nodes)-1]
	delete(q.indexOf, n.node.ID())
	return n
}

func (q *aStarQueue) update(id int, g, f float64) {
	i, ok := q.indexOf[id]
	if !ok {
		return
	}
	q.nodes[i].gscore = g
	q.nodes[i].fscore = f
	heap.Fix(q, i)
}

func (q *aStarQueue) node(id int) (aStarNode, bool) {
	loc, ok := q.indexOf[id]
	if ok {
		return q.nodes[loc], true
	}
	return aStarNode{}, false
}
