package graph

import (
	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
	imageapi "github.com/openshift/origin/pkg/image/api"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
)

const (
	// TriggersDeploymentEdgeKind points from DeploymentConfigs to ImageStreamTags that trigger the deployment
	TriggersDeploymentEdgeKind = "TriggersDeployment"
	// UsedInDeploymentEdgeKind points from DeploymentConfigs to DockerImageReferences that are used in the deployment
	UsedInDeploymentEdgeKind = "UsedInDeployment"
	// DeploymentEdgeKind points from DeploymentConfigs to the ReplicationControllers that are fulfilling the deployment
	DeploymentEdgeKind = "Deployment"
)

// AddTriggerEdges creates edges that point to named Docker image repositories for each image used in the deployment.
func AddTriggerEdges(g osgraph.MutableUniqueGraph, node *deploygraph.DeploymentConfigNode) *deploygraph.DeploymentConfigNode {
	podTemplate := node.DeploymentConfig.Spec.Template
	if podTemplate == nil {
		return node
	}

	deployapi.EachTemplateImage(
		&podTemplate.Spec,
		deployapi.DeploymentConfigHasTrigger(node.DeploymentConfig),
		func(image deployapi.TemplateImage, err error) {
			if err != nil {
				return
			}
			if image.From != nil {
				if len(image.From.Name) == 0 {
					return
				}
				name, tag, _ := imageapi.SplitImageStreamTag(image.From.Name)
				in := imagegraph.FindOrCreateSyntheticImageStreamTagNode(g, imagegraph.MakeImageStreamTagObjectMeta(image.From.Namespace, name, tag))
				g.AddEdge(in, node, TriggersDeploymentEdgeKind)
				return
			}

			tag := image.Ref.Tag
			image.Ref.Tag = ""
			in := imagegraph.EnsureDockerRepositoryNode(g, image.Ref.String(), tag)
			g.AddEdge(in, node, UsedInDeploymentEdgeKind)
		})

	return node
}

func AddAllTriggerEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if dcNode, ok := node.(*deploygraph.DeploymentConfigNode); ok {
			AddTriggerEdges(g, dcNode)
		}
	}
}

func AddDeploymentEdges(g osgraph.MutableUniqueGraph, node *deploygraph.DeploymentConfigNode) *deploygraph.DeploymentConfigNode {
	for _, n := range g.(graph.Graph).Nodes() {
		if rcNode, ok := n.(*kubegraph.ReplicationControllerNode); ok {
			if rcNode.ReplicationController.Namespace != node.DeploymentConfig.Namespace {
				continue
			}
			if BelongsToDeploymentConfig(node.DeploymentConfig, rcNode.ReplicationController) {
				g.AddEdge(node, rcNode, DeploymentEdgeKind)
			}
		}
	}

	return node
}

func AddAllDeploymentEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if dcNode, ok := node.(*deploygraph.DeploymentConfigNode); ok {
			AddDeploymentEdges(g, dcNode)
		}
	}
}
