package graph

import (
	"github.com/gonum/graph"
	kapi "k8s.io/kubernetes/pkg/api"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildgraph "github.com/openshift/origin/pkg/build/graph/nodes"
	buildutil "github.com/openshift/origin/pkg/build/util"
	imageapi "github.com/openshift/origin/pkg/image/api"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
)

const (
	// BuildTriggerImageEdgeKind is an edge from an ImageStream to a BuildConfig that
	// represents a trigger connection. Changes to the ImageStream will trigger a new build
	// from the BuildConfig.
	BuildTriggerImageEdgeKind = "BuildTriggerImage"

	// BuildInputImageEdgeKind is  an edge from an ImageStream to a BuildConfig, where the
	// ImageStream is the source image for the build (builder in S2I builds, FROM in Docker builds,
	// custom builder in Custom builds). The same ImageStream can also have a trigger
	// relationship with the BuildConfig, but not necessarily.
	BuildInputImageEdgeKind = "BuildInputImage"

	// BuildOutputEdgeKind is an edge from a BuildConfig to an ImageStream. The ImageStream will hold
	// the ouptut of the Builds created with that BuildConfig.
	BuildOutputEdgeKind = "BuildOutput"

	// BuildInputEdgeKind is an edge from a source repository to a BuildConfig. The source repository is the
	// input source for the build.
	BuildInputEdgeKind = "BuildInput"

	// BuildEdgeKind goes from a BuildConfigNode to a BuildNode and indicates that the buildConfig owns the build
	BuildEdgeKind = "Build"
)

// AddBuildEdges adds edges that connect a BuildConfig to Builds to the given graph
func AddBuildEdges(g osgraph.MutableUniqueGraph, node *buildgraph.BuildConfigNode) {
	for _, n := range g.(graph.Graph).Nodes() {
		if buildNode, ok := n.(*buildgraph.BuildNode); ok {
			if buildNode.Build.Namespace != node.BuildConfig.Namespace {
				continue
			}
			if belongsToBuildConfig(node.BuildConfig, buildNode.Build) {
				g.AddEdge(node, buildNode, BuildEdgeKind)
			}
		}
	}
}

// AddAllBuildEdges adds build edges to all BuildConfig nodes in the given graph
func AddAllBuildEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if bcNode, ok := node.(*buildgraph.BuildConfigNode); ok {
			AddBuildEdges(g, bcNode)
		}
	}
}

func imageRefNode(g osgraph.MutableUniqueGraph, ref *kapi.ObjectReference, bc *buildapi.BuildConfig) graph.Node {
	if ref == nil {
		return nil
	}
	switch ref.Kind {
	case "DockerImage":
		if ref, err := imageapi.ParseDockerImageReference(ref.Name); err == nil {
			tag := ref.Tag
			ref.Tag = ""
			return imagegraph.EnsureDockerRepositoryNode(g, ref.String(), tag)
		}
	case "ImageStream":
		return imagegraph.FindOrCreateSyntheticImageStreamTagNode(g, imagegraph.MakeImageStreamTagObjectMeta(defaultNamespace(ref.Namespace, bc.Namespace), ref.Name, imageapi.DefaultImageTag))
	case "ImageStreamTag":
		return imagegraph.FindOrCreateSyntheticImageStreamTagNode(g, imagegraph.MakeImageStreamTagObjectMeta2(defaultNamespace(ref.Namespace, bc.Namespace), ref.Name))
	case "ImageStreamImage":
		return imagegraph.FindOrCreateSyntheticImageStreamImageNode(g, imagegraph.MakeImageStreamImageObjectMeta(defaultNamespace(ref.Namespace, bc.Namespace), ref.Name))
	}
	return nil
}

// AddOutputEdges links the build config to its output image node.
func AddOutputEdges(g osgraph.MutableUniqueGraph, node *buildgraph.BuildConfigNode) {
	if node.BuildConfig.Spec.Output.To == nil {
		return
	}
	out := imageRefNode(g, node.BuildConfig.Spec.Output.To, node.BuildConfig)
	g.AddEdge(node, out, BuildOutputEdgeKind)
}

// AddInputEdges links the build config to its input image and source nodes.
func AddInputEdges(g osgraph.MutableUniqueGraph, node *buildgraph.BuildConfigNode) {
	if in := buildgraph.EnsureSourceRepositoryNode(g, node.BuildConfig.Spec.Source); in != nil {
		g.AddEdge(in, node, BuildInputEdgeKind)
	}
	inputImage := buildutil.GetInputReference(node.BuildConfig.Spec.Strategy)
	if input := imageRefNode(g, inputImage, node.BuildConfig); input != nil {
		g.AddEdge(input, node, BuildInputImageEdgeKind)
	}
}

// AddTriggerEdges links the build config to its trigger input image nodes.
func AddTriggerEdges(g osgraph.MutableUniqueGraph, node *buildgraph.BuildConfigNode) {
	for _, trigger := range node.BuildConfig.Spec.Triggers {
		if trigger.Type != buildapi.ImageChangeBuildTriggerType {
			continue
		}
		from := trigger.ImageChange.From
		if trigger.ImageChange.From == nil {
			from = buildutil.GetInputReference(node.BuildConfig.Spec.Strategy)
		}
		triggerNode := imageRefNode(g, from, node.BuildConfig)
		g.AddEdge(triggerNode, node, BuildTriggerImageEdgeKind)
	}
}

// AddInputOutputEdges links the build config to other nodes for the images and source repositories it depends on.
func AddInputOutputEdges(g osgraph.MutableUniqueGraph, node *buildgraph.BuildConfigNode) *buildgraph.BuildConfigNode {
	AddInputEdges(g, node)
	AddTriggerEdges(g, node)
	AddOutputEdges(g, node)
	return node
}

// AddAllInputOutputEdges adds input and output edges for all BuildConfigs in the given graph
func AddAllInputOutputEdges(g osgraph.MutableUniqueGraph) {
	for _, node := range g.(graph.Graph).Nodes() {
		if bcNode, ok := node.(*buildgraph.BuildConfigNode); ok {
			AddInputOutputEdges(g, bcNode)
		}
	}
}
