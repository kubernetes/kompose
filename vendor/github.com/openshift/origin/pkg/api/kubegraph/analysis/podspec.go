package analysis

import (
	"fmt"

	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
)

const (
	UnmountableSecretWarning    = "UnmountableSecret"
	MissingSecretWarning        = "MissingSecret"
	MissingLivenessProbeWarning = "MissingLivenessProbe"
)

// FindUnmountableSecrets inspects all PodSpecs for any Secret reference that isn't listed as mountable by the referenced ServiceAccount
func FindUnmountableSecrets(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastPodSpecNode := range g.NodesByKind(kubegraph.PodSpecNodeKind) {
		podSpecNode := uncastPodSpecNode.(*kubegraph.PodSpecNode)
		unmountableSecrets := CheckForUnmountableSecrets(g, podSpecNode)

		topLevelNode := osgraph.GetTopLevelContainerNode(g, podSpecNode)
		topLevelString := f.ResourceName(topLevelNode)

		saString := "MISSING_SA"
		saNodes := g.SuccessorNodesByEdgeKind(podSpecNode, kubeedges.ReferencedServiceAccountEdgeKind)
		if len(saNodes) > 0 {
			saString = f.ResourceName(saNodes[0])
		}

		for _, unmountableSecret := range unmountableSecrets {
			markers = append(markers, osgraph.Marker{
				Node:         podSpecNode,
				RelatedNodes: []graph.Node{unmountableSecret},

				Severity: osgraph.WarningSeverity,
				Key:      UnmountableSecretWarning,
				Message: fmt.Sprintf("%s is attempting to mount a secret %s disallowed by %s",
					topLevelString, f.ResourceName(unmountableSecret), saString),
			})
		}
	}

	return markers
}

// FindMissingSecrets inspects all PodSpecs for any Secret reference that is a synthetic node (not a pre-existing node in the graph)
func FindMissingSecrets(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastPodSpecNode := range g.NodesByKind(kubegraph.PodSpecNodeKind) {
		podSpecNode := uncastPodSpecNode.(*kubegraph.PodSpecNode)
		missingSecrets := CheckMissingMountedSecrets(g, podSpecNode)

		topLevelNode := osgraph.GetTopLevelContainerNode(g, podSpecNode)
		topLevelString := f.ResourceName(topLevelNode)

		for _, missingSecret := range missingSecrets {
			markers = append(markers, osgraph.Marker{
				Node:         podSpecNode,
				RelatedNodes: []graph.Node{missingSecret},

				Severity: osgraph.WarningSeverity,
				Key:      UnmountableSecretWarning,
				Message: fmt.Sprintf("%s is attempting to mount a missing secret %s",
					topLevelString, f.ResourceName(missingSecret)),
			})
		}
	}

	return markers
}

// FindMissingLivenessProbes inspects all PodSpecs for missing liveness probes and generates a list of non-duplicate markers
func FindMissingLivenessProbes(g osgraph.Graph, f osgraph.Namer, setProbeCommand string) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastPodSpecNode := range g.NodesByKind(kubegraph.PodSpecNodeKind) {
		podSpecNode := uncastPodSpecNode.(*kubegraph.PodSpecNode)
		if hasLivenessProbe(podSpecNode) {
			continue
		}

		topLevelNode := osgraph.GetTopLevelContainerNode(g, podSpecNode)

		// skip any podSpec nodes that are managed by other nodes.
		// Liveness probes should only be applied to a controlling
		// podSpec node, and not to any of its children.
		if hasControllerRefEdge(g, topLevelNode) {
			continue
		}

		topLevelString := f.ResourceName(topLevelNode)
		markers = append(markers, osgraph.Marker{
			Node:         podSpecNode,
			RelatedNodes: []graph.Node{topLevelNode},

			Severity: osgraph.WarningSeverity,
			Key:      MissingLivenessProbeWarning,
			Message: fmt.Sprintf("%s has no liveness probe to verify pods are still running.",
				topLevelString),
			Suggestion: osgraph.Suggestion(fmt.Sprintf("%s %s --liveness ...", setProbeCommand, topLevelString)),
		})
	}

	return markers
}

// hasLivenessProbe iterates through all of the containers in a podSpecNode returning true
// if at least one container has a liveness probe, or false otherwise
func hasLivenessProbe(podSpecNode *kubegraph.PodSpecNode) bool {
	for _, container := range podSpecNode.PodSpec.Containers {
		if container.LivenessProbe != nil {
			return true
		}
	}
	return false
}

// hasControllerRefEdge returns true if a given node contains one or more "ManagedByController" outbound edges
func hasControllerRefEdge(g osgraph.Graph, node graph.Node) bool {
	managedEdges := g.OutboundEdges(node, kubeedges.ManagedByControllerEdgeKind)
	return len(managedEdges) > 0
}

// CheckForUnmountableSecrets checks to be sure that all the referenced secrets are mountable (by service account)
func CheckForUnmountableSecrets(g osgraph.Graph, podSpecNode *kubegraph.PodSpecNode) []*kubegraph.SecretNode {
	saNodes := g.SuccessorNodesByNodeAndEdgeKind(podSpecNode, kubegraph.ServiceAccountNodeKind, kubeedges.ReferencedServiceAccountEdgeKind)
	saMountableSecrets := []*kubegraph.SecretNode{}

	if len(saNodes) > 0 {
		saNode := saNodes[0].(*kubegraph.ServiceAccountNode)
		for _, secretNode := range g.SuccessorNodesByNodeAndEdgeKind(saNode, kubegraph.SecretNodeKind, kubeedges.MountableSecretEdgeKind) {
			saMountableSecrets = append(saMountableSecrets, secretNode.(*kubegraph.SecretNode))
		}
	}

	unmountableSecrets := []*kubegraph.SecretNode{}

	for _, uncastMountedSecretNode := range g.SuccessorNodesByNodeAndEdgeKind(podSpecNode, kubegraph.SecretNodeKind, kubeedges.MountedSecretEdgeKind) {
		mountedSecretNode := uncastMountedSecretNode.(*kubegraph.SecretNode)

		mountable := false
		for _, mountableSecretNode := range saMountableSecrets {
			if mountableSecretNode == mountedSecretNode {
				mountable = true
				break
			}
		}

		if !mountable {
			unmountableSecrets = append(unmountableSecrets, mountedSecretNode)
			continue
		}
	}

	return unmountableSecrets
}

// CheckMissingMountedSecrets checks to be sure that all the referenced secrets are present (not synthetic)
func CheckMissingMountedSecrets(g osgraph.Graph, podSpecNode *kubegraph.PodSpecNode) []*kubegraph.SecretNode {
	missingSecrets := []*kubegraph.SecretNode{}

	for _, uncastMountedSecretNode := range g.SuccessorNodesByNodeAndEdgeKind(podSpecNode, kubegraph.SecretNodeKind, kubeedges.MountedSecretEdgeKind) {
		mountedSecretNode := uncastMountedSecretNode.(*kubegraph.SecretNode)
		if !mountedSecretNode.Found() {
			missingSecrets = append(missingSecrets, mountedSecretNode)
		}
	}

	return missingSecrets
}
