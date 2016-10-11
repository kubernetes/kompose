package graphview

import (
	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
)

type PetSet struct {
	PetSet *kubegraph.PetSetNode

	OwnedPods   []*kubegraph.PodNode
	CreatedPods []*kubegraph.PodNode

	// TODO: handle conflicting once controller refs are present, not worth it yet
}

// AllPetSets returns all the PetSets that aren't in the excludes set and the set of covered NodeIDs
func AllPetSets(g osgraph.Graph, excludeNodeIDs IntSet) ([]PetSet, IntSet) {
	covered := IntSet{}
	views := []PetSet{}

	for _, uncastNode := range g.NodesByKind(kubegraph.PetSetNodeKind) {
		if excludeNodeIDs.Has(uncastNode.ID()) {
			continue
		}

		view, covers := NewPetSet(g, uncastNode.(*kubegraph.PetSetNode))
		covered.Insert(covers.List()...)
		views = append(views, view)
	}

	return views, covered
}

// NewPetSet returns the PetSet and a set of all the NodeIDs covered by the PetSet
func NewPetSet(g osgraph.Graph, node *kubegraph.PetSetNode) (PetSet, IntSet) {
	covered := IntSet{}
	covered.Insert(node.ID())

	view := PetSet{}
	view.PetSet = node

	for _, uncastPodNode := range g.PredecessorNodesByEdgeKind(node, kubeedges.ManagedByControllerEdgeKind) {
		podNode := uncastPodNode.(*kubegraph.PodNode)
		covered.Insert(podNode.ID())
		view.OwnedPods = append(view.OwnedPods, podNode)
	}

	return view, covered
}
