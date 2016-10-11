package graph

import (
	"github.com/gonum/graph"

	kapi "k8s.io/kubernetes/pkg/api"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	routegraph "github.com/openshift/origin/pkg/route/graph/nodes"
)

const (
	// ExposedThroughRouteEdgeKind is an edge from a route to any object that
	// is exposed through routes
	ExposedThroughRouteEdgeKind = "ExposedThroughRoute"
)

// AddRouteEdges adds an edge that connect a service to a route in the given graph
func AddRouteEdges(g osgraph.MutableUniqueGraph, node *routegraph.RouteNode) {
	syntheticService := &kapi.Service{}
	syntheticService.Namespace = node.Namespace
	syntheticService.Name = node.Spec.To.Name

	serviceNode := kubegraph.FindOrCreateSyntheticServiceNode(g, syntheticService)
	g.AddEdge(node, serviceNode, ExposedThroughRouteEdgeKind)

	for _, svc := range node.Spec.AlternateBackends {
		syntheticService := &kapi.Service{}
		syntheticService.Namespace = node.Namespace
		syntheticService.Name = svc.Name

		serviceNode := kubegraph.FindOrCreateSyntheticServiceNode(g, syntheticService)
		g.AddEdge(node, serviceNode, ExposedThroughRouteEdgeKind)
	}
}

// AddAllRouteEdges adds service edges to all route nodes in the given graph
func AddAllRouteEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if routeNode, ok := node.(*routegraph.RouteNode); ok {
			AddRouteEdges(g, routeNode)
		}
	}
}
