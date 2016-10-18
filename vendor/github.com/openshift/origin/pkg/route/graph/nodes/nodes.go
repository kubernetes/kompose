package nodes

import (
	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	routeapi "github.com/openshift/origin/pkg/route/api"
)

// EnsureRouteNode adds a graph node for the specific route if it does not exist
func EnsureRouteNode(g osgraph.MutableUniqueGraph, route *routeapi.Route) *RouteNode {
	return osgraph.EnsureUnique(
		g,
		RouteNodeName(route),
		func(node osgraph.Node) graph.Node {
			return &RouteNode{
				Node:  node,
				Route: route,
			}
		},
	).(*RouteNode)
}
