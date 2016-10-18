// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package graph implements functions and interfaces to deal with formal discrete graphs. It aims to
be first and foremost flexible, with speed as a strong second priority.

In this package, graphs are taken to be directed, and undirected graphs are considered to be a
special case of directed graphs that happen to have reciprocal edges. Graphs are, by default,
unweighted, but functions that require weighted edges have several methods of dealing with this.
In order of precedence:

1. These functions have an argument called Cost (and in some cases, HeuristicCost). If this is
present, it will always be used to determine the cost between two nodes.

2. These functions will check if your graph implements the Coster (and/or HeuristicCoster)
interface. If this is present, and the Cost (or HeuristicCost) argument is nil, these functions
will be used.

3. Finally, if no user data is supplied, it will use the functions UniformCost (always returns 1)
and/or NulLHeuristic (always returns 0).

For information on the specification for Cost functions, please see the Coster interface.

Finally, although the functions take in a Graph -- they will always use the correct behavior.
If your graph implements DirectedGraph, it will use Successors and To where applicable,
if undirected, it will use From instead. If it implements neither, it will scan the edge list
for successors and predecessors where applicable. (This is slow, you should always implement either
Directed or Undirected)

This package will never modify a graph that is not Mutable (and the interface does not allow it to
do so). However, return values are free to be modified, so never pass a reference to your own edge
list or node list.  It also guarantees that any nodes passed back to the user will be the same
nodes returned to it -- that is, it will never take a Node's ID and then wrap the ID in a new
struct and return that. You'll always get back your original data.
*/
package graph
