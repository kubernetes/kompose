package analysis

import (
	"fmt"
	"strconv"

	"github.com/gonum/graph"

	kapi "k8s.io/kubernetes/pkg/api"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	routeapi "github.com/openshift/origin/pkg/route/api"
	routeedges "github.com/openshift/origin/pkg/route/graph"
	routegraph "github.com/openshift/origin/pkg/route/graph/nodes"
)

const (
	// MissingRoutePortWarning is returned when a route has no route port specified
	// and the service it routes to has multiple ports.
	MissingRoutePortWarning = "MissingRoutePort"
	// WrongRoutePortWarning is returned when a route has a route port specified
	// but the service it points to has no such port (either as a named port or as
	// a target port).
	WrongRoutePortWarning = "WrongRoutePort"
	// MissingServiceWarning is returned when there is no service for the specific route.
	MissingServiceWarning = "MissingService"
	// MissingTLSTerminationTypeErr is returned when a route with a tls config doesn't
	// specify a tls termination type.
	MissingTLSTerminationTypeErr = "MissingTLSTermination"
	// PathBasedPassthroughErr is returned when a path based route is passthrough
	// terminated.
	PathBasedPassthroughErr = "PathBasedPassthrough"
	// MissingTLSTerminationTypeErr is returned when a route with a tls config doesn't
	// specify a tls termination type.
	RouteNotAdmittedTypeErr = "RouteNotAdmitted"
	// MissingRequiredRouterErr is returned when no router has been setup.
	MissingRequiredRouterErr = "MissingRequiredRouter"
)

// FindPortMappingIssues checks all routes and reports any issues related to their ports.
// Also non-existent services for routes are reported here.
func FindPortMappingIssues(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastRouteNode := range g.NodesByKind(routegraph.RouteNodeKind) {
		routeNode := uncastRouteNode.(*routegraph.RouteNode)
		marker := routePortMarker(g, f, routeNode)
		if marker != nil {
			markers = append(markers, *marker)
		}
	}

	return markers
}

func routePortMarker(g osgraph.Graph, f osgraph.Namer, routeNode *routegraph.RouteNode) *osgraph.Marker {
	for _, uncastServiceNode := range g.SuccessorNodesByEdgeKind(routeNode, routeedges.ExposedThroughRouteEdgeKind) {
		svcNode := uncastServiceNode.(*kubegraph.ServiceNode)

		if !svcNode.Found() {
			return &osgraph.Marker{
				Node:         routeNode,
				RelatedNodes: []graph.Node{svcNode},

				Severity: osgraph.WarningSeverity,
				Key:      MissingServiceWarning,
				Message: fmt.Sprintf("%s is supposed to route traffic to %s but %s doesn't exist.",
					f.ResourceName(routeNode), f.ResourceName(svcNode), f.ResourceName(svcNode)),
				// TODO: Suggest 'oc create service' once that's a thing.
				// See https://github.com/kubernetes/kubernetes/pull/19509
			}
		}

		if len(svcNode.Spec.Ports) > 1 && (routeNode.Spec.Port == nil || len(routeNode.Spec.Port.TargetPort.String()) == 0) {
			return &osgraph.Marker{
				Node:         routeNode,
				RelatedNodes: []graph.Node{svcNode},

				Severity: osgraph.WarningSeverity,
				Key:      MissingRoutePortWarning,
				Message: fmt.Sprintf("%s doesn't have a port specified and is routing traffic to %s which uses multiple ports.",
					f.ResourceName(routeNode), f.ResourceName(svcNode)),
			}
		}

		if routeNode.Spec.Port == nil {
			// If no port is specified, we don't need to analyze any further.
			return nil
		}

		routePortString := routeNode.Spec.Port.TargetPort.String()
		if routePort, err := strconv.Atoi(routePortString); err == nil {
			for _, port := range svcNode.Spec.Ports {
				if port.TargetPort.IntValue() == routePort {
					return nil
				}
			}

			// route has a numeric port, service has no port with that number as a targetPort.
			marker := &osgraph.Marker{
				Node:         routeNode,
				RelatedNodes: []graph.Node{svcNode},

				Severity: osgraph.WarningSeverity,
				Key:      WrongRoutePortWarning,
				Message: fmt.Sprintf("%s has a port specified (%d) but %s has no such targetPort.",
					f.ResourceName(routeNode), routePort, f.ResourceName(svcNode)),
			}
			if len(svcNode.Spec.Ports) == 1 {
				marker.Suggestion = osgraph.Suggestion(fmt.Sprintf("oc patch %s -p '{\"spec\":{\"port\":{\"targetPort\": %d}}}'", f.ResourceName(routeNode), svcNode.Spec.Ports[0].TargetPort.IntValue()))
			}

			return marker
		}

		for _, port := range svcNode.Spec.Ports {
			if port.Name == routePortString {
				return nil
			}
		}

		// route has a named port, service has no port with that name.
		marker := &osgraph.Marker{
			Node:         routeNode,
			RelatedNodes: []graph.Node{svcNode},

			Severity: osgraph.WarningSeverity,
			Key:      WrongRoutePortWarning,
			Message: fmt.Sprintf("%s has a named port specified (%q) but %s has no such named port.",
				f.ResourceName(routeNode), routePortString, f.ResourceName(svcNode)),
		}
		if len(svcNode.Spec.Ports) == 1 {
			marker.Suggestion = osgraph.Suggestion(fmt.Sprintf("oc patch %s -p '{\"spec\":{\"port\":{\"targetPort\": %d}}}'", f.ResourceName(routeNode), svcNode.Spec.Ports[0].TargetPort.IntValue()))
		}

		return marker
	}
	return nil
}

func FindMissingTLSTerminationType(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastRouteNode := range g.NodesByKind(routegraph.RouteNodeKind) {
		routeNode := uncastRouteNode.(*routegraph.RouteNode)

		if routeNode.Spec.TLS != nil && len(routeNode.Spec.TLS.Termination) == 0 {
			markers = append(markers, osgraph.Marker{
				Node: routeNode,

				Severity:   osgraph.ErrorSeverity,
				Key:        MissingTLSTerminationTypeErr,
				Message:    fmt.Sprintf("%s has a TLS configuration but no termination type specified.", f.ResourceName(routeNode)),
				Suggestion: osgraph.Suggestion(fmt.Sprintf("oc patch %s -p '{\"spec\":{\"tls\":{\"termination\":\"<type>\"}}}' (replace <type> with a valid termination type: edge, passthrough, reencrypt)", f.ResourceName(routeNode)))})
		}
	}

	return markers
}

// FindRouteAdmissionFailures creates markers for any routes that were rejected by their routers
func FindRouteAdmissionFailures(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastRouteNode := range g.NodesByKind(routegraph.RouteNodeKind) {
		routeNode := uncastRouteNode.(*routegraph.RouteNode)
	Route:
		for _, ingress := range routeNode.Status.Ingress {
			switch status, condition := routeapi.IngressConditionStatus(&ingress, routeapi.RouteAdmitted); status {
			case kapi.ConditionFalse:
				markers = append(markers, osgraph.Marker{
					Node: routeNode,

					Severity: osgraph.ErrorSeverity,
					Key:      RouteNotAdmittedTypeErr,
					Message:  fmt.Sprintf("%s was not accepted by router %q: %s (%s)", f.ResourceName(routeNode), ingress.RouterName, condition.Message, condition.Reason),
				})
				break Route
			}
		}
	}

	return markers
}

// FindMissingRouter creates markers for all routes in case there is no running router.
func FindMissingRouter(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastRouteNode := range g.NodesByKind(routegraph.RouteNodeKind) {
		routeNode := uncastRouteNode.(*routegraph.RouteNode)

		if len(routeNode.Route.Status.Ingress) == 0 {
			markers = append(markers, osgraph.Marker{
				Node: routeNode,

				Severity:   osgraph.ErrorSeverity,
				Key:        MissingRequiredRouterErr,
				Message:    fmt.Sprintf("%s is routing traffic to svc/%s, but either the administrator has not installed a router or the router is not selecting this route.", f.ResourceName(routeNode), routeNode.Spec.To.Name),
				Suggestion: osgraph.Suggestion("oc adm router -h"),
			})
		}
	}

	return markers
}

func FindPathBasedPassthroughRoutes(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastRouteNode := range g.NodesByKind(routegraph.RouteNodeKind) {
		routeNode := uncastRouteNode.(*routegraph.RouteNode)

		if len(routeNode.Spec.Path) > 0 && routeNode.Spec.TLS != nil && routeNode.Spec.TLS.Termination == routeapi.TLSTerminationPassthrough {
			markers = append(markers, osgraph.Marker{
				Node: routeNode,

				Severity:   osgraph.ErrorSeverity,
				Key:        PathBasedPassthroughErr,
				Message:    fmt.Sprintf("%s is path-based and uses passthrough termination, which is an invalid combination.", f.ResourceName(routeNode)),
				Suggestion: osgraph.Suggestion(fmt.Sprintf("1. use spec.tls.termination=edge or 2. use spec.tls.termination=reencrypt and specify spec.tls.destinationCACertificate or 3. remove spec.path")),
			})
		}
	}

	return markers
}
