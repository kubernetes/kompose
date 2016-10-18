// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"fmt"
	"sort"

	"github.com/gonum/graph"
	"github.com/gonum/graph/internal"
)

// Unorderable is an error containing sets of unorderable graph.Nodes.
type Unorderable [][]graph.Node

// Error satisfies the error interface.
func (e Unorderable) Error() string {
	const maxNodes = 10
	var n int
	for _, c := range e {
		n += len(c)
	}
	if n > maxNodes {
		// Don't return errors that are too long.
		return fmt.Sprintf("topo: no topological ordering: %d nodes in %d cyclic components", n, len(e))
	}
	return fmt.Sprintf("topo: no topological ordering: cyclic components: %v", [][]graph.Node(e))
}

// Sort performs a topological sort of the directed graph g returning the 'from' to 'to'
// sort order. If a topological ordering is not possible, an Unorderable error is returned
// listing cyclic components in g with each cyclic component's members sorted by ID. When
// an Unorderable error is returned, each cyclic component's topological position within
// the sorted nodes is marked with a nil graph.Node.
func Sort(g graph.Directed) (sorted []graph.Node, err error) {
	sccs := TarjanSCC(g)
	sorted = make([]graph.Node, 0, len(sccs))
	var sc Unorderable
	for _, s := range sccs {
		if len(s) != 1 {
			sort.Sort(byID(s))
			sc = append(sc, s)
			sorted = append(sorted, nil)
			continue
		}
		sorted = append(sorted, s[0])
	}
	if sc != nil {
		for i, j := 0, len(sc)-1; i < j; i, j = i+1, j-1 {
			sc[i], sc[j] = sc[j], sc[i]
		}
		err = sc
	}
	reverse(sorted)
	return sorted, err
}

func reverse(p []graph.Node) {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

// TarjanSCC returns the strongly connected components of the graph g using Tarjan's algorithm.
//
// A strongly connected component of a graph is a set of vertices where it's possible to reach any
// vertex in the set from any other (meaning there's a cycle between them.)
//
// Generally speaking, a directed graph where the number of strongly connected components is equal
// to the number of nodes is acyclic, unless you count reflexive edges as a cycle (which requires
// only a little extra testing.)
//
func TarjanSCC(g graph.Directed) [][]graph.Node {
	nodes := g.Nodes()
	t := tarjan{
		succ: g.From,

		indexTable: make(map[int]int, len(nodes)),
		lowLink:    make(map[int]int, len(nodes)),
		onStack:    make(internal.IntSet, len(nodes)),
	}
	for _, v := range nodes {
		if t.indexTable[v.ID()] == 0 {
			t.strongconnect(v)
		}
	}
	return t.sccs
}

// tarjan implements Tarjan's strongly connected component finding
// algorithm. The implementation is from the pseudocode at
//
// http://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm?oldid=642744644
//
type tarjan struct {
	succ func(graph.Node) []graph.Node

	index      int
	indexTable map[int]int
	lowLink    map[int]int
	onStack    internal.IntSet

	stack []graph.Node

	sccs [][]graph.Node
}

// strongconnect is the strongconnect function described in the
// wikipedia article.
func (t *tarjan) strongconnect(v graph.Node) {
	vID := v.ID()

	// Set the depth index for v to the smallest unused index.
	t.index++
	t.indexTable[vID] = t.index
	t.lowLink[vID] = t.index
	t.stack = append(t.stack, v)
	t.onStack.Add(vID)

	// Consider successors of v.
	for _, w := range t.succ(v) {
		wID := w.ID()
		if t.indexTable[wID] == 0 {
			// Successor w has not yet been visited; recur on it.
			t.strongconnect(w)
			t.lowLink[vID] = min(t.lowLink[vID], t.lowLink[wID])
		} else if t.onStack.Has(wID) {
			// Successor w is in stack s and hence in the current SCC.
			t.lowLink[vID] = min(t.lowLink[vID], t.indexTable[wID])
		}
	}

	// If v is a root node, pop the stack and generate an SCC.
	if t.lowLink[vID] == t.indexTable[vID] {
		// Start a new strongly connected component.
		var (
			scc []graph.Node
			w   graph.Node
		)
		for {
			w, t.stack = t.stack[len(t.stack)-1], t.stack[:len(t.stack)-1]
			t.onStack.Remove(w.ID())
			// Add w to current strongly connected component.
			scc = append(scc, w)
			if w.ID() == vID {
				break
			}
		}
		// Output the current strongly connected component.
		t.sccs = append(t.sccs, scc)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
