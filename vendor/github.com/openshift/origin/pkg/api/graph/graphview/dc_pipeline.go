package graphview

import (
	"sort"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	deployedges "github.com/openshift/origin/pkg/deploy/graph"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
)

type DeploymentConfigPipeline struct {
	Deployment *deploygraph.DeploymentConfigNode

	ActiveDeployment    *kubegraph.ReplicationControllerNode
	InactiveDeployments []*kubegraph.ReplicationControllerNode

	Images []ImagePipeline
}

// AllDeploymentConfigPipelines returns all the DCPipelines that aren't in the excludes set and the set of covered NodeIDs
func AllDeploymentConfigPipelines(g osgraph.Graph, excludeNodeIDs IntSet) ([]DeploymentConfigPipeline, IntSet) {
	covered := IntSet{}
	dcPipelines := []DeploymentConfigPipeline{}

	for _, uncastNode := range g.NodesByKind(deploygraph.DeploymentConfigNodeKind) {
		if excludeNodeIDs.Has(uncastNode.ID()) {
			continue
		}

		pipeline, covers := NewDeploymentConfigPipeline(g, uncastNode.(*deploygraph.DeploymentConfigNode))
		covered.Insert(covers.List()...)
		dcPipelines = append(dcPipelines, pipeline)
	}

	sort.Sort(SortedDeploymentConfigPipeline(dcPipelines))
	return dcPipelines, covered
}

// NewDeploymentConfigPipeline returns the DeploymentConfigPipeline and a set of all the NodeIDs covered by the DeploymentConfigPipeline
func NewDeploymentConfigPipeline(g osgraph.Graph, dcNode *deploygraph.DeploymentConfigNode) (DeploymentConfigPipeline, IntSet) {
	covered := IntSet{}
	covered.Insert(dcNode.ID())

	dcPipeline := DeploymentConfigPipeline{}
	dcPipeline.Deployment = dcNode

	// for everything that can trigger a deployment, create an image pipeline and add it to the list
	for _, istNode := range g.PredecessorNodesByEdgeKind(dcNode, deployedges.TriggersDeploymentEdgeKind) {
		imagePipeline, covers := NewImagePipelineFromImageTagLocation(g, istNode, istNode.(ImageTagLocation))

		covered.Insert(covers.List()...)
		dcPipeline.Images = append(dcPipeline.Images, imagePipeline)
	}

	// for image that we use, create an image pipeline and add it to the list
	for _, tagNode := range g.PredecessorNodesByEdgeKind(dcNode, deployedges.UsedInDeploymentEdgeKind) {
		imagePipeline, covers := NewImagePipelineFromImageTagLocation(g, tagNode, tagNode.(ImageTagLocation))

		covered.Insert(covers.List()...)
		dcPipeline.Images = append(dcPipeline.Images, imagePipeline)
	}

	dcPipeline.ActiveDeployment, dcPipeline.InactiveDeployments = deployedges.RelevantDeployments(g, dcNode)
	for _, rc := range dcPipeline.InactiveDeployments {
		_, covers := NewReplicationController(g, rc)
		covered.Insert(covers.List()...)
	}

	if dcPipeline.ActiveDeployment != nil {
		_, covers := NewReplicationController(g, dcPipeline.ActiveDeployment)
		covered.Insert(covers.List()...)
	}

	return dcPipeline, covered
}

type SortedDeploymentConfigPipeline []DeploymentConfigPipeline

func (m SortedDeploymentConfigPipeline) Len() int      { return len(m) }
func (m SortedDeploymentConfigPipeline) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m SortedDeploymentConfigPipeline) Less(i, j int) bool {
	return CompareObjectMeta(&m[i].Deployment.DeploymentConfig.ObjectMeta, &m[j].Deployment.DeploymentConfig.ObjectMeta)
}
