package graphview

import (
	"fmt"
	"sort"

	kapi "k8s.io/kubernetes/pkg/api"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubeedges "github.com/openshift/origin/pkg/api/kubegraph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
	routeedges "github.com/openshift/origin/pkg/route/graph"
	routegraph "github.com/openshift/origin/pkg/route/graph/nodes"
)

// ServiceGroup is a service, the DeploymentConfigPipelines it covers, and lists of the other nodes that fulfill it
type ServiceGroup struct {
	Service *kubegraph.ServiceNode

	DeploymentConfigPipelines []DeploymentConfigPipeline
	ReplicationControllers    []ReplicationController
	PetSets                   []PetSet

	// TODO: this has to stop
	FulfillingPetSets []*kubegraph.PetSetNode
	FulfillingDCs     []*deploygraph.DeploymentConfigNode
	FulfillingRCs     []*kubegraph.ReplicationControllerNode
	FulfillingPods    []*kubegraph.PodNode

	ExposingRoutes []*routegraph.RouteNode
}

// AllServiceGroups returns all the ServiceGroups that aren't in the excludes set and the set of covered NodeIDs
func AllServiceGroups(g osgraph.Graph, excludeNodeIDs IntSet) ([]ServiceGroup, IntSet) {
	covered := IntSet{}
	services := []ServiceGroup{}

	for _, uncastNode := range g.NodesByKind(kubegraph.ServiceNodeKind) {
		if excludeNodeIDs.Has(uncastNode.ID()) {
			continue
		}

		service, covers := NewServiceGroup(g, uncastNode.(*kubegraph.ServiceNode))
		covered.Insert(covers.List()...)
		services = append(services, service)
	}

	sort.Sort(ServiceGroupByObjectMeta(services))
	return services, covered
}

// NewServiceGroup returns the ServiceGroup and a set of all the NodeIDs covered by the service
func NewServiceGroup(g osgraph.Graph, serviceNode *kubegraph.ServiceNode) (ServiceGroup, IntSet) {
	covered := IntSet{}
	covered.Insert(serviceNode.ID())

	service := ServiceGroup{}
	service.Service = serviceNode

	for _, uncastServiceFulfiller := range g.PredecessorNodesByEdgeKind(serviceNode, kubeedges.ExposedThroughServiceEdgeKind) {
		container := osgraph.GetTopLevelContainerNode(g, uncastServiceFulfiller)

		switch castContainer := container.(type) {
		case *deploygraph.DeploymentConfigNode:
			service.FulfillingDCs = append(service.FulfillingDCs, castContainer)
		case *kubegraph.ReplicationControllerNode:
			service.FulfillingRCs = append(service.FulfillingRCs, castContainer)
		case *kubegraph.PodNode:
			service.FulfillingPods = append(service.FulfillingPods, castContainer)
		case *kubegraph.PetSetNode:
			service.FulfillingPetSets = append(service.FulfillingPetSets, castContainer)
		default:
			utilruntime.HandleError(fmt.Errorf("unrecognized container: %v", castContainer))
		}
	}

	for _, uncastServiceFulfiller := range g.PredecessorNodesByEdgeKind(serviceNode, routeedges.ExposedThroughRouteEdgeKind) {
		container := osgraph.GetTopLevelContainerNode(g, uncastServiceFulfiller)

		switch castContainer := container.(type) {
		case *routegraph.RouteNode:
			service.ExposingRoutes = append(service.ExposingRoutes, castContainer)
		default:
			utilruntime.HandleError(fmt.Errorf("unrecognized container: %v", castContainer))
		}
	}

	// add the DCPipelines for all the DCs that fulfill the service
	for _, fulfillingDC := range service.FulfillingDCs {
		dcPipeline, dcCovers := NewDeploymentConfigPipeline(g, fulfillingDC)

		covered.Insert(dcCovers.List()...)
		service.DeploymentConfigPipelines = append(service.DeploymentConfigPipelines, dcPipeline)
	}

	for _, fulfillingRC := range service.FulfillingRCs {
		rcView, rcCovers := NewReplicationController(g, fulfillingRC)

		covered.Insert(rcCovers.List()...)
		service.ReplicationControllers = append(service.ReplicationControllers, rcView)
	}

	for _, fulfillingPetSet := range service.FulfillingPetSets {
		view, covers := NewPetSet(g, fulfillingPetSet)

		covered.Insert(covers.List()...)
		service.PetSets = append(service.PetSets, view)
	}

	for _, fulfillingPod := range service.FulfillingPods {
		_, podCovers := NewPod(g, fulfillingPod)
		covered.Insert(podCovers.List()...)
	}

	return service, covered
}

type ServiceGroupByObjectMeta []ServiceGroup

func (m ServiceGroupByObjectMeta) Len() int      { return len(m) }
func (m ServiceGroupByObjectMeta) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ServiceGroupByObjectMeta) Less(i, j int) bool {
	a, b := m[i], m[j]
	return CompareObjectMeta(&a.Service.Service.ObjectMeta, &b.Service.Service.ObjectMeta)
}

func CompareObjectMeta(a, b *kapi.ObjectMeta) bool {
	if a.Namespace == b.Namespace {
		return a.Name < b.Name
	}
	return a.Namespace < b.Namespace
}
