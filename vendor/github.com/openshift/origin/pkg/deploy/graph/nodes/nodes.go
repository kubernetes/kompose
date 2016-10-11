package nodes

import (
	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	depoyapi "github.com/openshift/origin/pkg/deploy/api"
)

// EnsureDeploymentConfigNode adds the provided deployment config to the graph if it does not exist
func EnsureDeploymentConfigNode(g osgraph.MutableUniqueGraph, dc *depoyapi.DeploymentConfig) *DeploymentConfigNode {
	dcName := DeploymentConfigNodeName(dc)
	dcNode := osgraph.EnsureUnique(
		g,
		dcName,
		func(node osgraph.Node) graph.Node {
			return &DeploymentConfigNode{Node: node, DeploymentConfig: dc, IsFound: true}
		},
	).(*DeploymentConfigNode)

	if dc.Spec.Template != nil {
		podTemplateSpecNode := kubegraph.EnsurePodTemplateSpecNode(g, dc.Spec.Template, dc.Namespace, dcName)
		g.AddEdge(dcNode, podTemplateSpecNode, osgraph.ContainsEdgeKind)
	}

	return dcNode
}

func FindOrCreateSyntheticDeploymentConfigNode(g osgraph.MutableUniqueGraph, dc *depoyapi.DeploymentConfig) *DeploymentConfigNode {
	return osgraph.EnsureUnique(
		g,
		DeploymentConfigNodeName(dc),
		func(node osgraph.Node) graph.Node {
			return &DeploymentConfigNode{Node: node, DeploymentConfig: dc, IsFound: false}
		},
	).(*DeploymentConfigNode)
}
