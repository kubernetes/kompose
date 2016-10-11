// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import "math"

// Node is a graph node. It returns a graph-unique integer ID.
type Node interface {
	ID() int
}

// Edge is a graph edge. In directed graphs, the direction of the
// edge is given from -> to, otherwise the edge is semantically
// unordered.
type Edge interface {
	From() Node
	To() Node
}

// Graph is a generalized graph.
type Graph interface {
	// Has returns whether the node exists within the graph.
	Has(Node) bool

	// Nodes returns all the nodes in the graph.
	Nodes() []Node

	// From returns all nodes that can be reached directly
	// from the given node.
	From(Node) []Node

	// HasEdge returns whether an edge exists between
	// nodes x and y without considering direction.
	HasEdge(x, y Node) bool

	// Edge returns the edge from u to v if such an edge
	// exists and nil otherwise. The node v must be directly
	// reachable from u as defined by the From method.
	Edge(u, v Node) Edge
}

// Undirected is an undirected graph.
type Undirected interface {
	Graph

	// EdgeBetween returns the edge between nodes x and y.
	EdgeBetween(x, y Node) Edge
}

// Directed is a directed graph.
type Directed interface {
	Graph

	// HasEdgeFromTo returns whether an edge exists
	// in the graph from u to v.
	HasEdgeFromTo(u, v Node) bool

	// To returns all nodes that can reach directly
	// to the given node.
	To(Node) []Node
}

// Weighter defines graphs that can report edge weights.
type Weighter interface {
	// Weight returns the weight for the given edge.
	Weight(Edge) float64
}

// Mutable is an interface for generalized graph mutation.
type Mutable interface {
	// NewNodeID returns a new unique arbitrary ID.
	NewNodeID() int

	// Adds a node to the graph. AddNode panics if
	// the added node ID matches an existing node ID.
	AddNode(Node)

	// RemoveNode removes a node from the graph, as
	// well as any edges attached to it. If the node
	// is not in the graph it is a no-op.
	RemoveNode(Node)

	// SetEdge adds an edge from one node to another.
	// If the nodes do not exist, they are added.
	// SetEdge will panic if the IDs of the e.From
	// and e.To are equal.
	SetEdge(e Edge, cost float64)

	// RemoveEdge removes the given edge, leaving the
	// terminal nodes. If the edge does not exist it
	// is a no-op.
	RemoveEdge(Edge)
}

// MutableUndirected is an undirected graph that can be arbitrarily altered.
type MutableUndirected interface {
	Undirected
	Mutable
}

// MutableDirected is a directed graph that can be arbitrarily altered.
type MutableDirected interface {
	Directed
	Mutable
}

// WeightFunc is a mapping between an edge and an edge weight.
type WeightFunc func(Edge) float64

// UniformCost is a WeightFunc that returns an edge cost of 1 for a non-nil Edge
// and Inf for a nil Edge.
func UniformCost(e Edge) float64 {
	if e == nil {
		return math.Inf(1)
	}
	return 1
}

// CopyUndirected copies nodes and edges as undirected edges from the source to the
// destination without first clearing the destination. CopyUndirected will panic if
// a node ID in the source graph matches a node ID in the destination. If the source
// does not implement Weighter, UniformCost is used to define edge weights.
//
// Note that if the source is a directed graph and a fundamental cycle exists with
// two nodes where the edge weights differ, the resulting destination graph's edge
// weight between those nodes is undefined.
func CopyUndirected(dst MutableUndirected, src Graph) {
	var weight WeightFunc
	if g, ok := src.(Weighter); ok {
		weight = g.Weight
	} else {
		weight = UniformCost
	}

	nodes := src.Nodes()
	for _, n := range nodes {
		dst.AddNode(n)
	}
	for _, u := range nodes {
		for _, v := range src.From(u) {
			edge := src.Edge(u, v)
			dst.SetEdge(edge, weight(edge))
		}
	}
}

// CopyDirected copies nodes and edges as directed edges from the source to the
// destination without first clearing the destination. CopyDirected will panic if
// a node ID in the source graph matches a node ID in the destination. If the
// source is undirected both directions will be present in the destination after
// the copy is complete. If the source does not implement Weighter, UniformCost
// is used to define edge weights.
func CopyDirected(dst MutableDirected, src Graph) {
	var weight WeightFunc
	if g, ok := src.(Weighter); ok {
		weight = g.Weight
	} else {
		weight = UniformCost
	}

	nodes := src.Nodes()
	for _, n := range nodes {
		dst.AddNode(n)
	}
	for _, u := range nodes {
		for _, v := range src.From(u) {
			edge := src.Edge(u, v)
			dst.SetEdge(edge, weight(edge))
		}
	}
}
