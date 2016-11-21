package analysis

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/util/sets"

	graphapi "github.com/gonum/graph"
	"github.com/gonum/graph/path"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	"github.com/openshift/origin/pkg/api/kubegraph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	kubenodes "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph"
	deploynodes "github.com/openshift/origin/pkg/deploy/graph/nodes"
)

const (
	// HPAMissingScaleRefError denotes an error where a Horizontal Pod Autoscaler does not have a reference to an object to scale
	HPAMissingScaleRefError = "HPAMissingScaleRef"
	// HPAMissingCPUTargetError denotes an error where a Horizontal Pod Autoscaler does not have a CPU target to scale by.
	// Currently, the only supported scale metric is CPU utilization, so without this metric an HPA is useless.
	HPAMissingCPUTargetError = "HPAMissingCPUTarget"
	// HPAOverlappingScaleRefWarning denotes a warning where a Horizontal Pod Autoscaler scales an object that is scaled by some other object as well
	HPAOverlappingScaleRefWarning = "HPAOverlappingScaleRef"
)

// FindHPASpecsMissingCPUTargets scans the graph in search of HorizontalPodAutoscalers that are missing a CPU utilization target.
// As of right now, the only metric that HPAs can use to scale pods is the CPU utilization, so if a HPA is missing this target it
// is effectively useless.
func FindHPASpecsMissingCPUTargets(graph osgraph.Graph, namer osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastNode := range graph.NodesByKind(kubenodes.HorizontalPodAutoscalerNodeKind) {
		node := uncastNode.(*kubenodes.HorizontalPodAutoscalerNode)

		if node.HorizontalPodAutoscaler.Spec.TargetCPUUtilizationPercentage == nil {
			markers = append(markers, osgraph.Marker{
				Node:       node,
				Severity:   osgraph.ErrorSeverity,
				Key:        HPAMissingCPUTargetError,
				Message:    fmt.Sprintf("%s is missing a CPU utilization target", namer.ResourceName(node)),
				Suggestion: osgraph.Suggestion(fmt.Sprintf(`oc patch %s -p '{"spec":{"targetCPUUtilizationPercentage": 80}}'`, namer.ResourceName(node))),
			})
		}
	}

	return markers
}

// FindHPASpecsMissingScaleRefs finds all Horizontal Pod Autoscalers whose scale reference points to an object that doesn't exist
// or that the client does not have the permission to see.
func FindHPASpecsMissingScaleRefs(graph osgraph.Graph, namer osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastNode := range graph.NodesByKind(kubenodes.HorizontalPodAutoscalerNodeKind) {
		node := uncastNode.(*kubenodes.HorizontalPodAutoscalerNode)

		scaledObjects := graph.SuccessorNodesByEdgeKind(
			uncastNode,
			kubegraph.ScalingEdgeKind,
		)

		if len(scaledObjects) < 1 {
			markers = append(markers, createMissingScaleRefMarker(node, nil, namer))
			continue
		}

		for _, scaleRef := range scaledObjects {
			if existenceChecker, ok := scaleRef.(osgraph.ExistenceChecker); ok && !existenceChecker.Found() {
				// if this node is synthetic, we can't be sure that the HPA is scaling something that actually exists
				markers = append(markers, createMissingScaleRefMarker(node, scaleRef, namer))
			}
		}
	}

	return markers
}

func createMissingScaleRefMarker(hpaNode *kubenodes.HorizontalPodAutoscalerNode, scaleRef graphapi.Node, namer osgraph.Namer) osgraph.Marker {
	return osgraph.Marker{
		Node:         hpaNode,
		Severity:     osgraph.ErrorSeverity,
		RelatedNodes: []graphapi.Node{scaleRef},
		Key:          HPAMissingScaleRefError,
		Message: fmt.Sprintf("%s is attempting to scale %s/%s, which doesn't exist",
			namer.ResourceName(hpaNode),
			hpaNode.HorizontalPodAutoscaler.Spec.ScaleTargetRef.Kind,
			hpaNode.HorizontalPodAutoscaler.Spec.ScaleTargetRef.Name,
		),
	}
}

// FindOverlappingHPAs scans the graph in search of HorizontalPodAutoscalers that are attempting to scale the same set of pods.
// This can occur in two ways:
//   - 1. label selectors for two ReplicationControllers/DeploymentConfigs/etc overlap
//   - 2. multiple HorizontalPodAutoscalers are attempting to scale the same ReplicationController/DeploymentConfig/etc
// Case 1 is handled by deconflicting the area of influence of ReplicationControllers/DeploymentConfigs/etc, and therefore we
// can assume that it will be handled before this step. Therefore, we are only concerned with finding HPAs that are trying to
// scale the same resources.
//
// The algorithm that is used to implement this check is described as follows:
//  - create a sub-graph containing only HPA nodes and other nodes that can be scaled, as well as any scaling edges or other
//    edges used to connect between objects that can be scaled
//  - for every resulting edge in the new sub-graph, create an edge in the reverse direction
//  - find the shortest paths between all HPA nodes in the graph
//  - shortest paths connecting two horizontal pod autoscalers are used to create markers for the graph
func FindOverlappingHPAs(graph osgraph.Graph, namer osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	nodeFilter := osgraph.NodesOfKind(
		kubenodes.HorizontalPodAutoscalerNodeKind,
		kubenodes.ReplicationControllerNodeKind,
		deploynodes.DeploymentConfigNodeKind,
	)
	edgeFilter := osgraph.EdgesOfKind(
		kubegraph.ScalingEdgeKind,
		deploygraph.DeploymentEdgeKind,
		kubeedges.ManagedByControllerEdgeKind,
	)

	hpaSubGraph := graph.Subgraph(nodeFilter, edgeFilter)
	for _, edge := range hpaSubGraph.Edges() {
		osgraph.AddReversedEdge(hpaSubGraph, edge.From(), edge.To(), sets.NewString())
	}

	hpaNodes := hpaSubGraph.NodesByKind(kubenodes.HorizontalPodAutoscalerNodeKind)

	for _, firstHPA := range hpaNodes {
		// we can use Dijkstra's algorithm as we know we do not have any negative edge weights
		shortestPaths := path.DijkstraFrom(firstHPA, hpaSubGraph)

		for _, secondHPA := range hpaNodes {
			if firstHPA == secondHPA {
				continue
			}

			shortestPath, _ := shortestPaths.To(secondHPA)

			if shortestPath == nil {
				// if two HPAs have no path between them, no error exists
				continue
			}

			markers = append(markers, osgraph.Marker{
				Node:         firstHPA,
				Severity:     osgraph.WarningSeverity,
				RelatedNodes: shortestPath[1:],
				Key:          HPAOverlappingScaleRefWarning,
				Message: fmt.Sprintf("%s and %s overlap because they both attempt to scale %s",
					namer.ResourceName(firstHPA), namer.ResourceName(secondHPA), nameList(shortestPath[1:len(shortestPath)-1], namer)),
			})
		}
	}

	return markers
}

// nameList outputs a nicely-formatted list of names:
//  - given nodes ['a', 'b', 'c'], this will return "one of a, b, or c"
//  - given nodes ['a', 'b'], this will return "a or b"
//  - given nodes ['a'], this will return "a"
func nameList(nodes []graphapi.Node, namer osgraph.Namer) string {
	names := []string{}

	for _, node := range nodes {
		names = append(names, namer.ResourceName(node))
	}

	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " or " + names[1]
	default:
		return "one of " + strings.Join(names[:len(names)-1], ", ") + ", or " + names[len(names)-1]
	}
}
