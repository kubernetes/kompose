package analysis

import (
	"fmt"

	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
)

const (
	UnmountableSecretWarning = "UnmountableSecret"
	MissingSecretWarning     = "MissingSecret"
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
