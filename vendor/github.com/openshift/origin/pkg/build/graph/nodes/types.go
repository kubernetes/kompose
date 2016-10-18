package nodes

import (
	"fmt"
	"reflect"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildapi "github.com/openshift/origin/pkg/build/api"
)

var (
	BuildConfigNodeKind = reflect.TypeOf(buildapi.BuildConfig{}).Name()
	BuildNodeKind       = reflect.TypeOf(buildapi.Build{}).Name()

	// non-api types
	SourceRepositoryNodeKind = reflect.TypeOf(buildapi.BuildSource{}).Name()
)

func BuildConfigNodeName(o *buildapi.BuildConfig) osgraph.UniqueName {
	return osgraph.GetUniqueRuntimeObjectNodeName(BuildConfigNodeKind, o)
}

type BuildConfigNode struct {
	osgraph.Node
	BuildConfig *buildapi.BuildConfig
}

func (n BuildConfigNode) Object() interface{} {
	return n.BuildConfig
}

func (n BuildConfigNode) String() string {
	return string(BuildConfigNodeName(n.BuildConfig))
}

func (n BuildConfigNode) UniqueName() osgraph.UniqueName {
	return BuildConfigNodeName(n.BuildConfig)
}

func (*BuildConfigNode) Kind() string {
	return BuildConfigNodeKind
}

func SourceRepositoryNodeName(source buildapi.BuildSource) osgraph.UniqueName {
	switch {
	case source.Git != nil:
		sourceType, uri, ref := "git", source.Git.URI, source.Git.Ref
		return osgraph.UniqueName(fmt.Sprintf("%s|%s|%s#%s", SourceRepositoryNodeKind, sourceType, uri, ref))
	default:
		panic(fmt.Sprintf("invalid build source: %v", source))
	}
}

type SourceRepositoryNode struct {
	osgraph.Node
	Source buildapi.BuildSource
}

func (n SourceRepositoryNode) String() string {
	return string(SourceRepositoryNodeName(n.Source))
}

func (SourceRepositoryNode) Kind() string {
	return SourceRepositoryNodeKind
}

func BuildNodeName(o *buildapi.Build) osgraph.UniqueName {
	return osgraph.GetUniqueRuntimeObjectNodeName(BuildNodeKind, o)
}

type BuildNode struct {
	osgraph.Node
	Build *buildapi.Build
}

func (n BuildNode) Object() interface{} {
	return n.Build
}

func (n BuildNode) String() string {
	return string(BuildNodeName(n.Build))
}

func (n BuildNode) UniqueName() osgraph.UniqueName {
	return BuildNodeName(n.Build)
}

func (*BuildNode) Kind() string {
	return BuildNodeKind
}
