package graphview

import (
	"k8s.io/kubernetes/pkg/api/unversioned"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	"github.com/openshift/origin/pkg/api/kubegraph/analysis"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
)

type ReplicationController struct {
	RC *kubegraph.ReplicationControllerNode

	OwnedPods   []*kubegraph.PodNode
	CreatedPods []*kubegraph.PodNode

	ConflictingRCs        []*kubegraph.ReplicationControllerNode
	ConflictingRCIDToPods map[int][]*kubegraph.PodNode
}

// AllReplicationControllers returns all the ReplicationControllers that aren't in the excludes set and the set of covered NodeIDs
func AllReplicationControllers(g osgraph.Graph, excludeNodeIDs IntSet) ([]ReplicationController, IntSet) {
	covered := IntSet{}
	rcViews := []ReplicationController{}

	for _, uncastNode := range g.NodesByKind(kubegraph.ReplicationControllerNodeKind) {
		if excludeNodeIDs.Has(uncastNode.ID()) {
			continue
		}

		rcView, covers := NewReplicationController(g, uncastNode.(*kubegraph.ReplicationControllerNode))
		covered.Insert(covers.List()...)
		rcViews = append(rcViews, rcView)
	}

	return rcViews, covered
}

// MaxRecentContainerRestarts returns the maximum container restarts for all pods in
// replication controller.
func (rc *ReplicationController) MaxRecentContainerRestarts() int32 {
	var maxRestarts int32
	for _, pod := range rc.OwnedPods {
		for _, status := range pod.Status.ContainerStatuses {
			if status.RestartCount > maxRestarts && analysis.ContainerRestartedRecently(status, unversioned.Now()) {
				maxRestarts = status.RestartCount
			}
		}
	}
	return maxRestarts
}

// NewReplicationController returns the ReplicationController and a set of all the NodeIDs covered by the ReplicationController
func NewReplicationController(g osgraph.Graph, rcNode *kubegraph.ReplicationControllerNode) (ReplicationController, IntSet) {
	covered := IntSet{}
	covered.Insert(rcNode.ID())

	rcView := ReplicationController{}
	rcView.RC = rcNode
	rcView.ConflictingRCIDToPods = map[int][]*kubegraph.PodNode{}

	for _, uncastPodNode := range g.PredecessorNodesByEdgeKind(rcNode, kubeedges.ManagedByControllerEdgeKind) {
		podNode := uncastPodNode.(*kubegraph.PodNode)
		covered.Insert(podNode.ID())
		rcView.OwnedPods = append(rcView.OwnedPods, podNode)

		// check to see if this pod is managed by more than one RC
		uncastOwningRCs := g.SuccessorNodesByEdgeKind(podNode, kubeedges.ManagedByControllerEdgeKind)
		if len(uncastOwningRCs) > 1 {
			for _, uncastOwningRC := range uncastOwningRCs {
				if uncastOwningRC.ID() == rcNode.ID() {
					continue
				}

				conflictingRC := uncastOwningRC.(*kubegraph.ReplicationControllerNode)
				rcView.ConflictingRCs = append(rcView.ConflictingRCs, conflictingRC)

				conflictingPods, ok := rcView.ConflictingRCIDToPods[conflictingRC.ID()]
				if !ok {
					conflictingPods = []*kubegraph.PodNode{}
				}
				conflictingPods = append(conflictingPods, podNode)
				rcView.ConflictingRCIDToPods[conflictingRC.ID()] = conflictingPods
			}
		}
	}

	return rcView, covered
}

// MaxRecentContainerRestartsForRC returns the maximum container restarts in pods
// in the replication controller node for the last 10 minutes.
func MaxRecentContainerRestartsForRC(g osgraph.Graph, rcNode *kubegraph.ReplicationControllerNode) int32 {
	if rcNode == nil {
		return 0
	}
	rc, _ := NewReplicationController(g, rcNode)
	return rc.MaxRecentContainerRestarts()
}
