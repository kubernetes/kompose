package graphview

import (
	"sort"

	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildedges "github.com/openshift/origin/pkg/build/graph"
	buildgraph "github.com/openshift/origin/pkg/build/graph/nodes"
	imageedges "github.com/openshift/origin/pkg/image/graph"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
)

// ImagePipeline represents a build, its output, and any inputs. The input
// to a build may be another ImagePipeline.
type ImagePipeline struct {
	Image               ImageTagLocation
	DestinationResolved bool
	ScheduledImport     bool

	Build *buildgraph.BuildConfigNode

	LastSuccessfulBuild   *buildgraph.BuildNode
	LastUnsuccessfulBuild *buildgraph.BuildNode
	ActiveBuilds          []*buildgraph.BuildNode

	// If set, the base image used by the build
	BaseImage ImageTagLocation
	// if set, the build config names that produces the base image
	BaseBuilds []string
	// If set, the source repository that inputs to the build
	Source SourceLocation
}

// ImageTagLocation identifies the source or destination of an image. Represents
// both a tag in a Docker image repository, as well as a tag in an OpenShift image stream.
type ImageTagLocation interface {
	ID() int
	ImageSpec() string
	ImageTag() string
}

// SourceLocation identifies a repository that is an input to a build.
type SourceLocation interface {
	ID() int
}

func AllImagePipelinesFromBuildConfig(g osgraph.Graph, excludeNodeIDs IntSet) ([]ImagePipeline, IntSet) {
	covered := IntSet{}
	pipelines := []ImagePipeline{}

	for _, uncastNode := range g.NodesByKind(buildgraph.BuildConfigNodeKind) {
		if excludeNodeIDs.Has(uncastNode.ID()) {
			continue
		}

		pipeline, covers := NewImagePipelineFromBuildConfigNode(g, uncastNode.(*buildgraph.BuildConfigNode))
		covered.Insert(covers.List()...)
		pipelines = append(pipelines, pipeline)
	}

	sort.Sort(SortedImagePipelines(pipelines))

	outputImageToBCMap := make(map[string][]string)
	for _, pipeline := range pipelines {
		// note, bc does not have to have an output image
		if pipeline.Image != nil {
			bcs, ok := outputImageToBCMap[pipeline.Image.ImageSpec()]
			if !ok {
				bcs = []string{}
			}
			bcs = append(bcs, pipeline.Build.BuildConfig.Name)
			outputImageToBCMap[pipeline.Image.ImageSpec()] = bcs
		}
	}

	if len(outputImageToBCMap) > 0 {
		for i, pipeline := range pipelines {
			// note, bc does not have to have an input strategy image
			if pipeline.BaseImage != nil {
				baseBCs, ok := outputImageToBCMap[pipeline.BaseImage.ImageSpec()]
				if ok && len(baseBCs) > 0 {
					pipelines[i].BaseBuilds = baseBCs
				}
			}
		}
	}

	return pipelines, covered
}

// NewImagePipeline attempts to locate a build flow from the provided node. If no such
// build flow can be located, false is returned.
func NewImagePipelineFromBuildConfigNode(g osgraph.Graph, bcNode *buildgraph.BuildConfigNode) (ImagePipeline, IntSet) {
	covered := IntSet{}
	covered.Insert(bcNode.ID())

	flow := ImagePipeline{}

	base, src, coveredInputs, scheduled, _ := findBuildInputs(g, bcNode)
	covered.Insert(coveredInputs.List()...)
	flow.BaseImage = base
	flow.Source = src
	flow.Build = bcNode
	flow.ScheduledImport = scheduled
	flow.LastSuccessfulBuild, flow.LastUnsuccessfulBuild, flow.ActiveBuilds = buildedges.RelevantBuilds(g, flow.Build)
	flow.Image = findBuildOutput(g, bcNode)

	// we should have at most one
	for _, buildOutputNode := range g.SuccessorNodesByEdgeKind(bcNode, buildedges.BuildOutputEdgeKind) {
		// this will handle the imagestream tag case
		for _, input := range g.SuccessorNodesByEdgeKind(buildOutputNode, imageedges.ReferencedImageStreamGraphEdgeKind) {
			imageStreamNode := input.(*imagegraph.ImageStreamNode)

			flow.DestinationResolved = (len(imageStreamNode.Status.DockerImageRepository) != 0)
		}
		// this will handle the imagestream image case
		for _, input := range g.SuccessorNodesByEdgeKind(buildOutputNode, imageedges.ReferencedImageStreamImageGraphEdgeKind) {
			imageStreamNode := input.(*imagegraph.ImageStreamNode)

			flow.DestinationResolved = (len(imageStreamNode.Status.DockerImageRepository) != 0)
		}

		// TODO handle the DockerImage case
	}

	return flow, covered
}

// NewImagePipelineFromImageTagLocation returns the ImagePipeline and all the nodes contributing to it
func NewImagePipelineFromImageTagLocation(g osgraph.Graph, node graph.Node, imageTagLocation ImageTagLocation) (ImagePipeline, IntSet) {
	covered := IntSet{}
	covered.Insert(node.ID())

	flow := ImagePipeline{}
	flow.Image = imageTagLocation

	for _, input := range g.PredecessorNodesByEdgeKind(node, buildedges.BuildOutputEdgeKind) {
		covered.Insert(input.ID())
		build := input.(*buildgraph.BuildConfigNode)
		if flow.Build != nil {
			// report this as an error (unexpected duplicate input build)
		}
		if build.BuildConfig == nil {
			// report this as as a missing build / broken link
			break
		}

		base, src, coveredInputs, scheduled, _ := findBuildInputs(g, build)
		covered.Insert(coveredInputs.List()...)
		flow.BaseImage = base
		flow.Source = src
		flow.Build = build
		flow.ScheduledImport = scheduled
		flow.LastSuccessfulBuild, flow.LastUnsuccessfulBuild, flow.ActiveBuilds = buildedges.RelevantBuilds(g, flow.Build)
	}

	for _, input := range g.SuccessorNodesByEdgeKind(node, imageedges.ReferencedImageStreamGraphEdgeKind) {
		covered.Insert(input.ID())
		imageStreamNode := input.(*imagegraph.ImageStreamNode)

		flow.DestinationResolved = (len(imageStreamNode.Status.DockerImageRepository) != 0)
	}
	for _, input := range g.SuccessorNodesByEdgeKind(node, imageedges.ReferencedImageStreamImageGraphEdgeKind) {
		covered.Insert(input.ID())
		imageStreamNode := input.(*imagegraph.ImageStreamNode)

		flow.DestinationResolved = (len(imageStreamNode.Status.DockerImageRepository) != 0)
	}

	return flow, covered
}

func findBuildInputs(g osgraph.Graph, bcNode *buildgraph.BuildConfigNode) (base ImageTagLocation, source SourceLocation, covered IntSet, scheduled bool, err error) {
	covered = IntSet{}

	// find inputs to the build
	for _, input := range g.PredecessorNodesByEdgeKind(bcNode, buildedges.BuildInputEdgeKind) {
		if source != nil {
			// report this as an error (unexpected duplicate source)
		}
		covered.Insert(input.ID())
		source = input.(SourceLocation)
	}
	for _, input := range g.PredecessorNodesByEdgeKind(bcNode, buildedges.BuildInputImageEdgeKind) {
		if base != nil {
			// report this as an error (unexpected duplicate input build)
		}
		covered.Insert(input.ID())
		base = input.(ImageTagLocation)
		scheduled = imageStreamTagScheduled(g, input, base)
	}

	return
}

func findBuildOutput(g osgraph.Graph, bcNode *buildgraph.BuildConfigNode) (result ImageTagLocation) {
	for _, output := range g.SuccessorNodesByEdgeKind(bcNode, buildedges.BuildOutputEdgeKind) {
		result = output.(ImageTagLocation)
		return
	}
	return
}

func imageStreamTagScheduled(g osgraph.Graph, input graph.Node, base ImageTagLocation) (scheduled bool) {
	for _, uncastImageStreamNode := range g.SuccessorNodesByEdgeKind(input, imageedges.ReferencedImageStreamGraphEdgeKind) {
		imageStreamNode := uncastImageStreamNode.(*imagegraph.ImageStreamNode)
		if imageStreamNode.ImageStream != nil {
			if tag, ok := imageStreamNode.ImageStream.Spec.Tags[base.ImageTag()]; ok {
				scheduled = tag.ImportPolicy.Scheduled
				return
			}
		}
	}
	return
}

type SortedImagePipelines []ImagePipeline

func (m SortedImagePipelines) Len() int      { return len(m) }
func (m SortedImagePipelines) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m SortedImagePipelines) Less(i, j int) bool {
	return CompareImagePipeline(&m[i], &m[j])
}

func CompareImagePipeline(a, b *ImagePipeline) bool {
	switch {
	case a.Build != nil && b.Build != nil && a.Build.BuildConfig != nil && b.Build.BuildConfig != nil:
		return CompareObjectMeta(&a.Build.BuildConfig.ObjectMeta, &b.Build.BuildConfig.ObjectMeta)
	case a.Build != nil && a.Build.BuildConfig != nil:
		return true
	case b.Build != nil && b.Build.BuildConfig != nil:
		return false
	}
	if a.Image == nil || b.Image == nil {
		return true
	}
	return a.Image.ImageSpec() < b.Image.ImageSpec()
}
