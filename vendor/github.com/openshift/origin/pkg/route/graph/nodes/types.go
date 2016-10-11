package nodes

import (
	"reflect"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	routeapi "github.com/openshift/origin/pkg/route/api"
)

var (
	RouteNodeKind = reflect.TypeOf(routeapi.Route{}).Name()
)

func RouteNodeName(o *routeapi.Route) osgraph.UniqueName {
	return osgraph.GetUniqueRuntimeObjectNodeName(RouteNodeKind, o)
}

type RouteNode struct {
	osgraph.Node
	*routeapi.Route
}

func (n RouteNode) Object() interface{} {
	return n.Route
}

func (n RouteNode) String() string {
	return string(RouteNodeName(n.Route))
}

func (*RouteNode) Kind() string {
	return RouteNodeKind
}
