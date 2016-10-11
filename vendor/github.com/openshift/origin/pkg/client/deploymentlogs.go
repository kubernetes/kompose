package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"

	"github.com/openshift/origin/pkg/deploy/api"
)

// DeploymentLogsNamespacer has methods to work with DeploymentLogs resources in a namespace
type DeploymentLogsNamespacer interface {
	DeploymentLogs(namespace string) DeploymentLogInterface
}

// DeploymentLogInterface exposes methods on DeploymentLogs resources.
type DeploymentLogInterface interface {
	Get(name string, opts api.DeploymentLogOptions) *restclient.Request
}

// deploymentLogs implements DeploymentLogsNamespacer interface
type deploymentLogs struct {
	r  *Client
	ns string
}

// newDeploymentLogs returns a deploymentLogs
func newDeploymentLogs(c *Client, namespace string) *deploymentLogs {
	return &deploymentLogs{
		r:  c,
		ns: namespace,
	}
}

// Get gets the deploymentlogs and return a deploymentLog request
func (c *deploymentLogs) Get(name string, opts api.DeploymentLogOptions) *restclient.Request {
	return c.r.Get().Namespace(c.ns).Resource("deploymentConfigs").Name(name).SubResource("log").VersionedParams(&opts, kapi.ParameterCodec)
}
