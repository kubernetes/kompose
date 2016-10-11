// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/internal"
)

// VertexOrdering returns the vertex ordering and the k-cores of
// the undirected graph g.
func VertexOrdering(g graph.Undirected) (order []graph.Node, cores [][]graph.Node) {
	nodes := g.Nodes()

	// The algorithm used here is essentially as described at
	// http://en.wikipedia.org/w/index.php?title=Degeneracy_%28graph_theory%29&oldid=640308710

	// Initialize an output list L.
	var l []graph.Node

	// Compute a number d_v for each vertex v in G,
	// the number of neighbors of v that are not already in L.
	// Initially, these numbers are just the degrees of the vertices.
	dv := make(map[int]int, len(nodes))
	var (
		maxDegree  int
		neighbours = make(map[int][]graph.Node)
	)
	for _, n := range nodes {
		adj := g.From(n)
		neighbours[n.ID()] = adj
		dv[n.ID()] = len(adj)
		if len(adj) > maxDegree {
			maxDegree = len(adj)
		}
	}

	// Initialize an array D such that D[i] contains a list of the
	// vertices v that are not already in L for which d_v = i.
	d := make([][]graph.Node, maxDegree+1)
	for _, n := range nodes {
		deg := dv[n.ID()]
		d[deg] = append(d[deg], n)
	}

	// Initialize k to 0.
	k := 0
	// Repeat n times:
	s := []int{0}
	for _ = range nodes { // TODO(kortschak): Remove blank assignment when go1.3.3 is no longer supported.
		// Scan the array cells D[0], D[1], ... until
		// finding an i for which D[i] is nonempty.
		var (
			i  int
			di []graph.Node
		)
		for i, di = range d {
			if len(di) != 0 {
				break
			}
		}

		// Set k to max(k,i).
		if i > k {
			k = i
			s = append(s, make([]int, k-len(s)+1)...)
		}

		// Select a vertex v from D[i]. Add v to the
		// beginning of L and remove it from D[i].
		var v graph.Node
		v, d[i] = di[len(di)-1], di[:len(di)-1]
		l = append(l, v)
		s[k]++
		delete(dv, v.ID())

		// For each neighbor w of v not already in L,
		// subtract one from d_w and move w to the
		// cell of D corresponding to the new value of d_w.
		for _, w := range neighbours[v.ID()] {
			dw, ok := dv[w.ID()]
			if !ok {
				continue
			}
			for i, n := range d[dw] {
				if n.ID() == w.ID() {
					d[dw][i], d[dw] = d[dw][len(d[dw])-1], d[dw][:len(d[dw])-1]
					dw--
					d[dw] = append(d[dw], w)
					break
				}
			}
			dv[w.ID()] = dw
		}
	}

	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	cores = make([][]graph.Node, len(s))
	offset := len(l)
	for i, n := range s {
		cores[i] = l[offset-n : offset]
		offset -= n
	}
	return l, cores
}

// BronKerbosch returns the set of maximal cliques of the undirected graph g.
func BronKerbosch(g graph.Undirected) [][]graph.Node {
	nodes := g.Nodes()

	// The algorithm used here is essentially BronKerbosch3 as described at
	// http://en.wikipedia.org/w/index.php?title=Bron%E2%80%93Kerbosch_algorithm&oldid=656805858

	p := make(internal.Set, len(nodes))
	for _, n := range nodes {
		p.Add(n)
	}
	x := make(internal.Set)
	var bk bronKerbosch
	order, _ := VertexOrdering(g)
	for _, v := range order {
		neighbours := g.From(v)
		nv := make(internal.Set, len(neighbours))
		for _, n := range neighbours {
			nv.Add(n)
		}
		bk.maximalCliquePivot(g, []graph.Node{v}, make(internal.Set).Intersect(p, nv), make(internal.Set).Intersect(x, nv))
		p.Remove(v)
		x.Add(v)
	}
	return bk
}

type bronKerbosch [][]graph.Node

func (bk *bronKerbosch) maximalCliquePivot(g graph.Undirected, r []graph.Node, p, x internal.Set) {
	if len(p) == 0 && len(x) == 0 {
		*bk = append(*bk, r)
		return
	}

	neighbours := bk.choosePivotFrom(g, p, x)
	nu := make(internal.Set, len(neighbours))
	for _, n := range neighbours {
		nu.Add(n)
	}
	for _, v := range p {
		if nu.Has(v) {
			continue
		}
		neighbours := g.From(v)
		nv := make(internal.Set, len(neighbours))
		for _, n := range neighbours {
			nv.Add(n)
		}

		var found bool
		for _, n := range r {
			if n.ID() == v.ID() {
				found = true
				break
			}
		}
		var sr []graph.Node
		if !found {
			sr = append(r[:len(r):len(r)], v)
		}

		bk.maximalCliquePivot(g, sr, make(internal.Set).Intersect(p, nv), make(internal.Set).Intersect(x, nv))
		p.Remove(v)
		x.Add(v)
	}
}

func (*bronKerbosch) choosePivotFrom(g graph.Undirected, p, x internal.Set) (neighbors []graph.Node) {
	// TODO(kortschak): Investigate the impact of pivot choice that maximises
	// |p ⋂ neighbours(u)| as a function of input size. Until then, leave as
	// compile time option.
	if !tomitaTanakaTakahashi {
		for _, n := range p {
			return g.From(n)
		}
		for _, n := range x {
			return g.From(n)
		}
		panic("bronKerbosch: empty set")
	}

	var (
		max   = -1
		pivot graph.Node
	)
	maxNeighbors := func(s internal.Set) {
	outer:
		for _, u := range s {
			nb := g.From(u)
			c := len(nb)
			if c <= max {
				continue
			}
			for n := range nb {
				if _, ok := p[n]; ok {
					continue
				}
				c--
				if c <= max {
					continue outer
				}
			}
			max = c
			pivot = u
			neighbors = nb
		}
	}
	maxNeighbors(p)
	maxNeighbors(x)
	if pivot == nil {
		panic("bronKerbosch: empty set")
	}
	return neighbors
}
