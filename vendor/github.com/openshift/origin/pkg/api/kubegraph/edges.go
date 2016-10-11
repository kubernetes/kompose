package kubegraph

import (
	"strings"

	"github.com/gonum/graph"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
)

const (
	// ExposedThroughServiceEdgeKind goes from a PodTemplateSpec or a Pod to Service.  The head should make the service's selector.
	ExposedThroughServiceEdgeKind = "ExposedThroughService"
	// ManagedByControllerEdgeKind goes from Pod to controller when the Pod satisfies a controller's label selector
	ManagedByControllerEdgeKind = "ManagedByController"
	// MountedSecretEdgeKind goes from PodSpec to Secret indicating that is or will be a request to mount a volume with the Secret.
	MountedSecretEdgeKind = "MountedSecret"
	// MountableSecretEdgeKind goes from ServiceAccount to Secret indicating that the SA allows the Secret to be mounted
	MountableSecretEdgeKind = "MountableSecret"
	// ReferencedServiceAccountEdgeKind goes from PodSpec to ServiceAccount indicating that Pod is or will be running as the SA.
	ReferencedServiceAccountEdgeKind = "ReferencedServiceAccount"
	// ScalingEdgeKind goes from HorizontalPodAutoscaler to scaled objects indicating that the HPA scales the object
	ScalingEdgeKind = "Scaling"
)

// AddExposedPodTemplateSpecEdges ensures that a directed edge exists between a service and all the PodTemplateSpecs
// in the graph that match the service selector
func AddExposedPodTemplateSpecEdges(g osgraph.MutableUniqueGraph, node *kubegraph.ServiceNode) {
	if node.Service.Spec.Selector == nil {
		return
	}
	query := labels.SelectorFromSet(node.Service.Spec.Selector)
	for _, n := range g.(graph.Graph).Nodes() {
		switch target := n.(type) {
		case *kubegraph.PodTemplateSpecNode:
			if target.Namespace != node.Namespace {
				continue
			}

			if query.Matches(labels.Set(target.PodTemplateSpec.Labels)) {
				g.AddEdge(target, node, ExposedThroughServiceEdgeKind)
			}
		}
	}
}

// AddAllExposedPodTemplateSpecEdges calls AddExposedPodTemplateSpecEdges for every ServiceNode in the graph
func AddAllExposedPodTemplateSpecEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if serviceNode, ok := node.(*kubegraph.ServiceNode); ok {
			AddExposedPodTemplateSpecEdges(g, serviceNode)
		}
	}
}

// AddExposedPodEdges ensures that a directed edge exists between a service and all the pods
// in the graph that match the service selector
func AddExposedPodEdges(g osgraph.MutableUniqueGraph, node *kubegraph.ServiceNode) {
	if node.Service.Spec.Selector == nil {
		return
	}
	query := labels.SelectorFromSet(node.Service.Spec.Selector)
	for _, n := range g.(graph.Graph).Nodes() {
		switch target := n.(type) {
		case *kubegraph.PodNode:
			if target.Namespace != node.Namespace {
				continue
			}
			if query.Matches(labels.Set(target.Labels)) {
				g.AddEdge(target, node, ExposedThroughServiceEdgeKind)
			}
		}
	}
}

// AddAllExposedPodEdges calls AddExposedPodEdges for every ServiceNode in the graph
func AddAllExposedPodEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if serviceNode, ok := node.(*kubegraph.ServiceNode); ok {
			AddExposedPodEdges(g, serviceNode)
		}
	}
}

// AddManagedByControllerPodEdges ensures that a directed edge exists between a controller and all the pods
// in the graph that match the label selector
func AddManagedByControllerPodEdges(g osgraph.MutableUniqueGraph, to graph.Node, namespace string, selector map[string]string) {
	if selector == nil {
		return
	}
	query := labels.SelectorFromSet(selector)
	for _, n := range g.(graph.Graph).Nodes() {
		switch target := n.(type) {
		case *kubegraph.PodNode:
			if target.Namespace != namespace {
				continue
			}
			if query.Matches(labels.Set(target.Labels)) {
				g.AddEdge(target, to, ManagedByControllerEdgeKind)
			}
		}
	}
}

// AddAllManagedByControllerPodEdges calls AddManagedByControllerPodEdges for every node in the graph
// TODO: should do this through an interface (selects pods)
func AddAllManagedByControllerPodEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		switch cast := node.(type) {
		case *kubegraph.ReplicationControllerNode:
			AddManagedByControllerPodEdges(g, cast, cast.ReplicationController.Namespace, cast.ReplicationController.Spec.Selector)
		case *kubegraph.PetSetNode:
			// TODO: refactor to handle expanded selectors (along with ReplicaSets and Deployments)
			AddManagedByControllerPodEdges(g, cast, cast.PetSet.Namespace, cast.PetSet.Spec.Selector.MatchLabels)
		}
	}
}

func AddMountedSecretEdges(g osgraph.Graph, podSpec *kubegraph.PodSpecNode) {
	//pod specs are always contained.  We'll get the toplevel container so that we can pull a namespace from it
	containerNode := osgraph.GetTopLevelContainerNode(g, podSpec)
	containerObj := g.GraphDescriber.Object(containerNode)

	meta, err := kapi.ObjectMetaFor(containerObj.(runtime.Object))
	if err != nil {
		// this should never happen.  it means that a podSpec is owned by a top level container that is not a runtime.Object
		panic(err)
	}

	for _, volume := range podSpec.Volumes {
		source := volume.VolumeSource
		if source.Secret == nil {
			continue
		}

		// pod secrets must be in the same namespace
		syntheticSecret := &kapi.Secret{}
		syntheticSecret.Namespace = meta.Namespace
		syntheticSecret.Name = source.Secret.SecretName

		secretNode := kubegraph.FindOrCreateSyntheticSecretNode(g, syntheticSecret)
		g.AddEdge(podSpec, secretNode, MountedSecretEdgeKind)
	}
}

func AddAllMountedSecretEdges(g osgraph.Graph) {
	for _, node := range g.Nodes() {
		if podSpecNode, ok := node.(*kubegraph.PodSpecNode); ok {
			AddMountedSecretEdges(g, podSpecNode)
		}
	}
}

func AddMountableSecretEdges(g osgraph.Graph, saNode *kubegraph.ServiceAccountNode) {
	for _, mountableSecret := range saNode.ServiceAccount.Secrets {
		syntheticSecret := &kapi.Secret{}
		syntheticSecret.Namespace = saNode.ServiceAccount.Namespace
		syntheticSecret.Name = mountableSecret.Name

		secretNode := kubegraph.FindOrCreateSyntheticSecretNode(g, syntheticSecret)
		g.AddEdge(saNode, secretNode, MountableSecretEdgeKind)
	}
}

func AddAllMountableSecretEdges(g osgraph.Graph) {
	for _, node := range g.Nodes() {
		if saNode, ok := node.(*kubegraph.ServiceAccountNode); ok {
			AddMountableSecretEdges(g, saNode)
		}
	}
}

func AddRequestedServiceAccountEdges(g osgraph.Graph, podSpecNode *kubegraph.PodSpecNode) {
	//pod specs are always contained.  We'll get the toplevel container so that we can pull a namespace from it
	containerNode := osgraph.GetTopLevelContainerNode(g, podSpecNode)
	containerObj := g.GraphDescriber.Object(containerNode)

	meta, err := kapi.ObjectMetaFor(containerObj.(runtime.Object))
	if err != nil {
		panic(err)
	}

	// if no SA name is present, admission will set 'default'
	name := "default"
	if len(podSpecNode.ServiceAccountName) > 0 {
		name = podSpecNode.ServiceAccountName
	}

	syntheticSA := &kapi.ServiceAccount{}
	syntheticSA.Namespace = meta.Namespace
	syntheticSA.Name = name

	saNode := kubegraph.FindOrCreateSyntheticServiceAccountNode(g, syntheticSA)
	g.AddEdge(podSpecNode, saNode, ReferencedServiceAccountEdgeKind)
}

func AddAllRequestedServiceAccountEdges(g osgraph.Graph) {
	for _, node := range g.Nodes() {
		if podSpecNode, ok := node.(*kubegraph.PodSpecNode); ok {
			AddRequestedServiceAccountEdges(g, podSpecNode)
		}
	}
}

func AddHPAScaleRefEdges(g osgraph.Graph) {
	for _, node := range g.NodesByKind(kubegraph.HorizontalPodAutoscalerNodeKind) {
		hpaNode := node.(*kubegraph.HorizontalPodAutoscalerNode)

		syntheticMeta := kapi.ObjectMeta{
			Name:      hpaNode.HorizontalPodAutoscaler.Spec.ScaleTargetRef.Name,
			Namespace: hpaNode.HorizontalPodAutoscaler.Namespace,
		}

		var groupVersionResource unversioned.GroupVersionResource
		resource := strings.ToLower(hpaNode.HorizontalPodAutoscaler.Spec.ScaleTargetRef.Kind)
		if groupVersion, err := unversioned.ParseGroupVersion(hpaNode.HorizontalPodAutoscaler.Spec.ScaleTargetRef.APIVersion); err == nil {
			groupVersionResource = groupVersion.WithResource(resource)
		} else {
			groupVersionResource = unversioned.GroupVersionResource{Resource: resource}
		}

		groupVersionResource, err := registered.RESTMapper().ResourceFor(groupVersionResource)
		if err != nil {
			continue
		}

		var syntheticNode graph.Node
		switch groupVersionResource.GroupResource() {
		case kapi.Resource("replicationcontrollers"):
			syntheticNode = kubegraph.FindOrCreateSyntheticReplicationControllerNode(g, &kapi.ReplicationController{ObjectMeta: syntheticMeta})
		case deployapi.Resource("deploymentconfigs"):
			syntheticNode = deploygraph.FindOrCreateSyntheticDeploymentConfigNode(g, &deployapi.DeploymentConfig{ObjectMeta: syntheticMeta})
		default:
			continue
		}

		g.AddEdge(hpaNode, syntheticNode, ScalingEdgeKind)
	}
}
