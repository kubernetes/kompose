package nodes

import (
	"reflect"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

var (
	DeploymentConfigNodeKind = reflect.TypeOf(deployapi.DeploymentConfig{}).Name()
)

func DeploymentConfigNodeName(o *deployapi.DeploymentConfig) osgraph.UniqueName {
	return osgraph.GetUniqueRuntimeObjectNodeName(DeploymentConfigNodeKind, o)
}

type DeploymentConfigNode struct {
	osgraph.Node
	DeploymentConfig *deployapi.DeploymentConfig

	IsFound bool
}

func (n DeploymentConfigNode) Found() bool {
	return n.IsFound
}

func (n DeploymentConfigNode) Object() interface{} {
	return n.DeploymentConfig
}

func (n DeploymentConfigNode) String() string {
	return string(DeploymentConfigNodeName(n.DeploymentConfig))
}

func (*DeploymentConfigNode) Kind() string {
	return DeploymentConfigNodeKind
}
