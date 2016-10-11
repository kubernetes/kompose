// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concrete

import (
	"fmt"

	"github.com/gonum/graph"
)

// A simple int alias.
type Node int

func (n Node) ID() int {
	return int(n)
}

// Just a collection of two nodes
type Edge struct {
	F, T graph.Node
}

func (e Edge) From() graph.Node {
	return e.F
}

func (e Edge) To() graph.Node {
	return e.T
}

type WeightedEdge struct {
	graph.Edge
	Cost float64
}

// A GonumGraph is a very generalized graph that can handle an arbitrary number of vertices and
// edges -- as well as act as either directed or undirected.
//
// Internally, it uses a map of successors AND predecessors, to speed up some operations (such as
// getting all successors/predecessors). It also speeds up things like adding edges (assuming both
// edges exist).
//
// However, its generality is also its weakness (and partially a flaw in needing to satisfy
// MutableGraph). For most purposes, creating your own graph is probably better. For instance,
// see TileGraph for an example of an immutable 2D grid of tiles that also implements the Graph
// interface, but would be more suitable if all you needed was a simple undirected 2D grid.
type Graph struct {
	neighbors map[int]map[int]WeightedEdge
	nodeMap   map[int]graph.Node

	// Node add/remove convenience vars
	maxID   int
	freeMap map[int]struct{}
}

func NewGraph() *Graph {
	return &Graph{
		neighbors: make(map[int]map[int]WeightedEdge),
		nodeMap:   make(map[int]graph.Node),
		maxID:     0,
		freeMap:   make(map[int]struct{}),
	}
}

func (g *Graph) NewNodeID() int {
	if g.maxID != maxInt {
		g.maxID++
		return g.maxID
	}

	// Implicitly checks if len(g.freeMap) == 0
	for id := range g.freeMap {
		return id
	}

	// I cannot foresee this ever happening, but just in case, we check.
	if len(g.nodeMap) == maxInt {
		panic("cannot allocate node: graph too large")
	}

	for i := 0; i < maxInt; i++ {
		if _, ok := g.nodeMap[i]; !ok {
			return i
		}
	}

	// Should not happen.
	panic("cannot allocate node id: no free id found")
}

func (g *Graph) AddNode(n graph.Node) {
	if _, exists := g.nodeMap[n.ID()]; exists {
		panic(fmt.Sprintf("concrete: node ID collision: %d", n.ID()))
	}
	g.nodeMap[n.ID()] = n
	g.neighbors[n.ID()] = make(map[int]WeightedEdge)

	delete(g.freeMap, n.ID())
	g.maxID = max(g.maxID, n.ID())
}

func (g *Graph) SetEdge(e graph.Edge, cost float64) {
	var (
		from = e.From()
		fid  = from.ID()
		to   = e.To()
		tid  = to.ID()
	)

	if fid == tid {
		panic("concrete: adding self edge")
	}

	if !g.Has(from) {
		g.AddNode(from)
	}

	if !g.Has(to) {
		g.AddNode(to)
	}

	g.neighbors[fid][tid] = WeightedEdge{Edge: e, Cost: cost}
	g.neighbors[tid][fid] = WeightedEdge{Edge: e, Cost: cost}
}

func (g *Graph) RemoveNode(n graph.Node) {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return
	}
	delete(g.nodeMap, n.ID())

	for neigh := range g.neighbors[n.ID()] {
		delete(g.neighbors[neigh], n.ID())
	}
	delete(g.neighbors, n.ID())

	if g.maxID != 0 && n.ID() == g.maxID {
		g.maxID--
	}
	g.freeMap[n.ID()] = struct{}{}
}

func (g *Graph) RemoveEdge(e graph.Edge) {
	from, to := e.From(), e.To()
	if _, ok := g.nodeMap[from.ID()]; !ok {
		return
	} else if _, ok := g.nodeMap[to.ID()]; !ok {
		return
	}

	delete(g.neighbors[from.ID()], to.ID())
	delete(g.neighbors[to.ID()], from.ID())
}

func (g *Graph) EmptyGraph() {
	g.neighbors = make(map[int]map[int]WeightedEdge)
	g.nodeMap = make(map[int]graph.Node)
}

/* Graph implementation */

func (g *Graph) From(n graph.Node) []graph.Node {
	if !g.Has(n) {
		return nil
	}

	neighbors := make([]graph.Node, len(g.neighbors[n.ID()]))
	i := 0
	for id := range g.neighbors[n.ID()] {
		neighbors[i] = g.nodeMap[id]
		i++
	}

	return neighbors
}

func (g *Graph) HasEdge(n, neigh graph.Node) bool {
	_, ok := g.neighbors[n.ID()][neigh.ID()]
	return ok
}

func (g *Graph) Edge(u, v graph.Node) graph.Edge {
	return g.EdgeBetween(u, v)
}

func (g *Graph) EdgeBetween(u, v graph.Node) graph.Edge {
	// We don't need to check if neigh exists because
	// it's implicit in the neighbors access.
	if !g.Has(u) {
		return nil
	}

	return g.neighbors[u.ID()][v.ID()].Edge
}

func (g *Graph) Node(id int) graph.Node {
	return g.nodeMap[id]
}

func (g *Graph) Has(n graph.Node) bool {
	_, ok := g.nodeMap[n.ID()]

	return ok
}

func (g *Graph) Nodes() []graph.Node {
	nodes := make([]graph.Node, len(g.nodeMap))
	i := 0
	for _, n := range g.nodeMap {
		nodes[i] = n
		i++
	}

	return nodes
}

func (g *Graph) Weight(e graph.Edge) float64 {
	if n, ok := g.neighbors[e.From().ID()]; ok {
		if we, ok := n[e.To().ID()]; ok {
			return we.Cost
		}
	}
	return inf
}

func (g *Graph) Edges() []graph.Edge {
	m := make(map[WeightedEdge]struct{})
	toReturn := make([]graph.Edge, 0)

	for _, neighs := range g.neighbors {
		for _, we := range neighs {
			if _, ok := m[we]; !ok {
				m[we] = struct{}{}
				toReturn = append(toReturn, we.Edge)
			}
		}
	}

	return toReturn
}

func (g *Graph) Degree(n graph.Node) int {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return 0
	}

	return len(g.neighbors[n.ID()])
}
