package nodes

import (
	"github.com/gonum/graph"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildapi "github.com/openshift/origin/pkg/build/api"
)

// EnsureBuildConfigNode adds a graph node for the specific build config if it does not exist
func EnsureBuildConfigNode(g osgraph.MutableUniqueGraph, config *buildapi.BuildConfig) *BuildConfigNode {
	return osgraph.EnsureUnique(
		g,
		BuildConfigNodeName(config),
		func(node osgraph.Node) graph.Node {
			return &BuildConfigNode{
				Node:        node,
				BuildConfig: config,
			}
		},
	).(*BuildConfigNode)
}

// EnsureSourceRepositoryNode adds the specific BuildSource to the graph if it does not already exist.
func EnsureSourceRepositoryNode(g osgraph.MutableUniqueGraph, source buildapi.BuildSource) *SourceRepositoryNode {
	switch {
	case source.Git != nil:
	default:
		return nil
	}
	return osgraph.EnsureUnique(g,
		SourceRepositoryNodeName(source),
		func(node osgraph.Node) graph.Node {
			return &SourceRepositoryNode{node, source}
		},
	).(*SourceRepositoryNode)
}

// EnsureBuildNode adds a graph node for the build if it does not already exist.
func EnsureBuildNode(g osgraph.MutableUniqueGraph, build *buildapi.Build) *BuildNode {
	return osgraph.EnsureUnique(g,
		BuildNodeName(build),
		func(node osgraph.Node) graph.Node {
			return &BuildNode{node, build}
		},
	).(*BuildNode)
}
