package kubernetes

import (
	"errors"
	"fmt"

	"k8s.io/kubernetes/pkg/util/sets"
)

// Graph datastructure node
type Node struct {
	name string
	deps []string
}

// create new node of graph
func NewNode(name string, deps ...string) *Node {
	return &Node{
		name: name,
		deps: deps,
	}
}

type Graph []*Node

// Displays the dependency graph
func displayGraph(graph Graph) {
	for _, node := range graph {
		for _, dep := range node.deps {
			fmt.Printf("%s -> %s\n", node.name, dep)
		}
	}
}

// Resolves the dependency graph, feed it a graph that is unresolvedGraph
// get back a graph that has dependencies resolved
func resolveGraph(graph Graph) (Graph, error) {
	// A map containing the node names and the actual node object
	nodeNames := make(map[string]*Node)

	// A map containing the nodes and their dependencies
	nodeDependencies := make(map[string]sets.String)

	// Populate the maps
	for _, node := range graph {
		nodeNames[node.name] = node

		dependencySet := sets.NewString()
		for _, dep := range node.deps {
			dependencySet.Insert(dep)
		}
		nodeDependencies[node.name] = dependencySet
	}

	// Iteratively find and remove nodes from the graph which have no dependencies.
	// If at some point there are still nodes in the graph and we cannot find
	// nodes without dependencies, that means we have a circular dependency
	var resolved Graph
	for len(nodeDependencies) != 0 {
		// Get all nodes from the graph which have no dependencies
		readySet := sets.NewString()
		for name, deps := range nodeDependencies {
			if deps.Len() == 0 {
				readySet.Insert(name)
			}
		}

		// If there aren't any ready nodes, then we have a cicular dependency
		if readySet.Len() == 0 {
			var g Graph
			for name := range nodeDependencies {
				g = append(g, nodeNames[name])
			}

			return g, errors.New("Circular dependency found")
		}

		// Remove the ready nodes and add them to the resolved graph
		for _, name := range readySet.List() {
			delete(nodeDependencies, name)
			resolved = append(resolved, nodeNames[name])
		}

		// Also make sure to remove the ready nodes from the
		// remaining node dependencies as well
		for name, deps := range nodeDependencies {
			diff := deps.Difference(readySet)
			nodeDependencies[name] = diff
		}
	}
	return resolved, nil
}

// This function will create a graph from data fed it in the form of
// map[svc] = [svc1, svc2, svc3]
// so svc depends on svc1, svc2 and svc3
// returns a resolved graph of dependencies
func FindDependency(dependency map[string]sets.String) Graph {
	var unresolvedGraph Graph

	for name, deps := range dependency {
		unresolvedGraph = append(unresolvedGraph, NewNode(name, deps.List()...))
	}
	resolved, _ := resolveGraph(unresolvedGraph)
	return resolved
}

// Normally resolved graph has things like
// svc -> (depends on) svc1, svc2, etc.
// but we want them to be colocated as [svc, svc1, svc2]
func CalculateColocation(resolved Graph) []sets.String {
	var colocation []sets.String
	for _, node := range resolved {
		dependencies := sets.NewString(node.deps...)
		dependencies.Insert(node.name)

		index := searchSVC(colocation, dependencies)
		if index == -1 {
			colocation = append(colocation, dependencies)
		} else {
			colocation[index].Insert(node.deps...)
			colocation[index].Insert(node.name)
		}
	}
	return colocation
}

// this function will search in colocation array if the query has intersection
// of even one service with any of the already added services in colocation
// colocation = [[svc1, svc2], [svc3, svc4]]
// query1 = [svc1, svcbar] will return non-negative number, in above case '0' (index of colocation)
// query2 = [svcfoo] will return -1, because it is not in colocation
func searchSVC(colocation []sets.String, query sets.String) int {
	for i, v := range colocation {
		inter := v.Intersection(query)
		if inter.Len() != 0 {
			return i
		}
	}
	return -1
}
