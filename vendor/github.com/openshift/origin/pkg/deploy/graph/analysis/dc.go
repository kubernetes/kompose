package analysis

import (
	"fmt"

	"github.com/gonum/graph"

	kapi "k8s.io/kubernetes/pkg/api"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	buildedges "github.com/openshift/origin/pkg/build/graph"
	buildutil "github.com/openshift/origin/pkg/build/util"
	deployedges "github.com/openshift/origin/pkg/deploy/graph"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
	imageedges "github.com/openshift/origin/pkg/image/graph"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
)

const (
	MissingImageStreamErr        = "MissingImageStream"
	MissingImageStreamTagWarning = "MissingImageStreamTag"
	MissingReadinessProbeWarning = "MissingReadinessProbe"

	SingleHostVolumeWarning = "SingleHostVolume"
	MissingPVCWarning       = "MissingPersistentVolumeClaim"
)

// FindDeploymentConfigTriggerErrors checks for possible failures in deployment config
// image change triggers.
//
// Precedence of failures:
// 1. The image stream for the tag of interest does not exist.
// 2. The image stream tag does not exist.
func FindDeploymentConfigTriggerErrors(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastDcNode := range g.NodesByKind(deploygraph.DeploymentConfigNodeKind) {
		dcNode := uncastDcNode.(*deploygraph.DeploymentConfigNode)
		marker := ictMarker(g, f, dcNode)
		if marker != nil {
			markers = append(markers, *marker)
		}
	}

	return markers
}

// ictMarker inspects the image change triggers for the provided deploymentconfig and returns
// a marker in case of the following two scenarios:
//
// 1. The image stream pointed by the dc trigger doen not exist.
// 2. The image stream tag pointed by the dc trigger does not exist and there is no build in
// 	  flight that could push to the tag.
func ictMarker(g osgraph.Graph, f osgraph.Namer, dcNode *deploygraph.DeploymentConfigNode) *osgraph.Marker {
	for _, uncastIstNode := range g.PredecessorNodesByEdgeKind(dcNode, deployedges.TriggersDeploymentEdgeKind) {
		if istNode := uncastIstNode.(*imagegraph.ImageStreamTagNode); !istNode.Found() {
			// The image stream for the tag of interest does not exist.
			if isNode, exists := doesImageStreamExist(g, uncastIstNode); !exists {
				return &osgraph.Marker{
					Node:         dcNode,
					RelatedNodes: []graph.Node{uncastIstNode, isNode},

					Severity: osgraph.ErrorSeverity,
					Key:      MissingImageStreamErr,
					Message: fmt.Sprintf("The image trigger for %s will have no effect because %s does not exist.",
						f.ResourceName(dcNode), f.ResourceName(isNode)),
					// TODO: Suggest `oc create imagestream` once we have that.
				}
			}

			for _, bcNode := range buildedges.BuildConfigsForTag(g, istNode) {
				// Avoid warning for the dc image trigger in case there is a build in flight.
				if latestBuild := buildedges.GetLatestBuild(g, bcNode); latestBuild != nil && !buildutil.IsBuildComplete(latestBuild.Build) {
					return nil
				}
			}

			// The image stream tag of interest does not exist.
			return &osgraph.Marker{
				Node:         dcNode,
				RelatedNodes: []graph.Node{uncastIstNode},

				Severity: osgraph.WarningSeverity,
				Key:      MissingImageStreamTagWarning,
				Message: fmt.Sprintf("The image trigger for %s will have no effect until %s is imported or created by a build.",
					f.ResourceName(dcNode), f.ResourceName(istNode)),
			}
		}
	}
	return nil
}

func doesImageStreamExist(g osgraph.Graph, istag graph.Node) (graph.Node, bool) {
	for _, imagestream := range g.SuccessorNodesByEdgeKind(istag, imageedges.ReferencedImageStreamGraphEdgeKind) {
		return imagestream, imagestream.(*imagegraph.ImageStreamNode).Found()
	}
	for _, imagestream := range g.SuccessorNodesByEdgeKind(istag, imageedges.ReferencedImageStreamImageGraphEdgeKind) {
		return imagestream, imagestream.(*imagegraph.ImageStreamNode).Found()
	}
	return nil, false
}

// FindDeploymentConfigReadinessWarnings inspects deploymentconfigs and reports those that
// don't have readiness probes set up.
func FindDeploymentConfigReadinessWarnings(g osgraph.Graph, f osgraph.Namer, setProbeCommand string) []osgraph.Marker {
	markers := []osgraph.Marker{}

Node:
	for _, uncastDcNode := range g.NodesByKind(deploygraph.DeploymentConfigNodeKind) {
		dcNode := uncastDcNode.(*deploygraph.DeploymentConfigNode)
		if t := dcNode.DeploymentConfig.Spec.Template; t != nil && len(t.Spec.Containers) > 0 {
			for _, container := range t.Spec.Containers {
				if container.ReadinessProbe != nil {
					continue Node
				}
			}
			// All of the containers in the deployment config lack a readiness probe
			markers = append(markers, osgraph.Marker{
				Node:     uncastDcNode,
				Severity: osgraph.WarningSeverity,
				Key:      MissingReadinessProbeWarning,
				Message: fmt.Sprintf("%s has no readiness probe to verify pods are ready to accept traffic or ensure deployment is successful.",
					f.ResourceName(dcNode)),
				Suggestion: osgraph.Suggestion(fmt.Sprintf("%s %s --readiness ...", setProbeCommand, f.ResourceName(dcNode))),
			})
			continue Node
		}
	}

	return markers
}

func FindPersistentVolumeClaimWarnings(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastDcNode := range g.NodesByKind(deploygraph.DeploymentConfigNodeKind) {
		dcNode := uncastDcNode.(*deploygraph.DeploymentConfigNode)
		marker := pvcMarker(g, f, dcNode)
		if marker != nil {
			markers = append(markers, *marker)
		}
	}

	return markers
}

func pvcMarker(g osgraph.Graph, f osgraph.Namer, dcNode *deploygraph.DeploymentConfigNode) *osgraph.Marker {
	for _, uncastPvcNode := range g.SuccessorNodesByEdgeKind(dcNode, deployedges.VolumeClaimEdgeKind) {
		pvcNode := uncastPvcNode.(*kubegraph.PersistentVolumeClaimNode)

		if !pvcNode.Found() {
			return &osgraph.Marker{
				Node:         dcNode,
				RelatedNodes: []graph.Node{uncastPvcNode},

				Severity: osgraph.WarningSeverity,
				Key:      MissingPVCWarning,
				Message:  fmt.Sprintf("%s points to a missing persistent volume claim (%s).", f.ResourceName(dcNode), f.ResourceName(pvcNode)),
				// TODO: Suggestion: osgraph.Suggestion(fmt.Sprintf("oc create pvc ...")),
			}
		}

		dc := dcNode.DeploymentConfig
		rollingParams := dc.Spec.Strategy.RollingParams
		isBlockedBySize := dc.Spec.Replicas > 1
		isBlockedRolling := rollingParams != nil && rollingParams.MaxSurge.IntValue() > 0

		// If the claim is not RWO or deployments will not have more than a pod running at any time
		// then they should be fine.
		if !hasRWOAccess(pvcNode) || (!isBlockedRolling && !isBlockedBySize) {
			continue
		}

		// This shouldn't be an issue on single-host clusters but they are not the common case anyway.
		// If github.com/kubernetes/kubernetes/issues/26567 ever gets fixed upstream, then we can drop
		// this warning.
		return &osgraph.Marker{
			Node:         dcNode,
			RelatedNodes: []graph.Node{uncastPvcNode},

			Severity: osgraph.WarningSeverity,
			Key:      SingleHostVolumeWarning,
			Message:  fmt.Sprintf("%s references a volume which may only be used in a single pod at a time - this may lead to hung deployments", f.ResourceName(dcNode)),
		}
	}
	return nil
}

func hasRWOAccess(pvcNode *kubegraph.PersistentVolumeClaimNode) bool {
	for _, accessMode := range pvcNode.PersistentVolumeClaim.Spec.AccessModes {
		if accessMode == kapi.ReadWriteOnce {
			return true
		}
	}
	return false
}
