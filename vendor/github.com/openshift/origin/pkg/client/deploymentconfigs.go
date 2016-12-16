package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	extensionsv1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

// DeploymentConfigsNamespacer has methods to work with DeploymentConfig resources in a namespace
type DeploymentConfigsNamespacer interface {
	DeploymentConfigs(namespace string) DeploymentConfigInterface
}

// DeploymentConfigInterface contains methods for working with DeploymentConfigs
type DeploymentConfigInterface interface {
	List(opts kapi.ListOptions) (*deployapi.DeploymentConfigList, error)
	Get(name string) (*deployapi.DeploymentConfig, error)
	Create(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Update(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
	Generate(name string) (*deployapi.DeploymentConfig, error)
	Rollback(config *deployapi.DeploymentConfigRollback) (*deployapi.DeploymentConfig, error)
	RollbackDeprecated(config *deployapi.DeploymentConfigRollback) (*deployapi.DeploymentConfig, error)
	GetScale(name string) (*extensions.Scale, error)
	UpdateScale(scale *extensions.Scale) (*extensions.Scale, error)
	UpdateStatus(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Instantiate(request *deployapi.DeploymentRequest) (*deployapi.DeploymentConfig, error)
}

// deploymentConfigs implements DeploymentConfigsNamespacer interface
type deploymentConfigs struct {
	r  *Client
	ns string
}

// newDeploymentConfigs returns a deploymentConfigs
func newDeploymentConfigs(c *Client, namespace string) *deploymentConfigs {
	return &deploymentConfigs{
		r:  c,
		ns: namespace,
	}
}

// List takes a label and field selectors, and returns the list of deploymentConfigs that match that selectors
func (c *deploymentConfigs) List(opts kapi.ListOptions) (result *deployapi.DeploymentConfigList, err error) {
	result = &deployapi.DeploymentConfigList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("deploymentConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular deploymentConfig
func (c *deploymentConfigs) Get(name string) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Get().Namespace(c.ns).Resource("deploymentConfigs").Name(name).Do().Into(result)
	return
}

// Create creates a new deploymentConfig
func (c *deploymentConfigs) Create(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().Namespace(c.ns).Resource("deploymentConfigs").Body(deploymentConfig).Do().Into(result)
	return
}

// Update updates an existing deploymentConfig
func (c *deploymentConfigs) Update(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(deploymentConfig.Name).Body(deploymentConfig).Do().Into(result)
	return
}

// Delete deletes an existing deploymentConfig.
func (c *deploymentConfigs) Delete(name string) error {
	return c.r.Delete().Namespace(c.ns).Resource("deploymentConfigs").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested deploymentConfigs.
func (c *deploymentConfigs) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("deploymentConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}

// Generate generates a new deploymentConfig for the given name.
func (c *deploymentConfigs) Generate(name string) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Get().Namespace(c.ns).Resource("generateDeploymentConfigs").Name(name).Do().Into(result)
	return
}

// Rollback rolls a deploymentConfig back to a previous configuration
func (c *deploymentConfigs) Rollback(config *deployapi.DeploymentConfigRollback) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("deploymentConfigs").
		Name(config.Name).
		SubResource("rollback").
		Body(config).
		Do().
		Into(result)
	return
}

// RollbackDeprecated rolls a deploymentConfig back to a previous configuration
func (c *deploymentConfigs) RollbackDeprecated(config *deployapi.DeploymentConfigRollback) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("deploymentConfigRollbacks").
		Body(config).
		Do().
		Into(result)
	return
}

// GetScale returns information about a particular deploymentConfig via its scale subresource
func (c *deploymentConfigs) GetScale(name string) (result *extensions.Scale, err error) {
	result = &extensions.Scale{}
	err = c.r.Get().Namespace(c.ns).Resource("deploymentConfigs").Name(name).SubResource("scale").Do().Into(result)
	return
}

// UpdateScale scales an existing deploymentConfig via its scale subresource
func (c *deploymentConfigs) UpdateScale(scale *extensions.Scale) (result *extensions.Scale, err error) {
	result = &extensions.Scale{}

	// TODO fix by making the client understand how to encode using different codecs for different resources
	encodedBytes, err := runtime.Encode(kapi.Codecs.LegacyCodec(extensionsv1beta1.SchemeGroupVersion), scale)
	if err != nil {
		return result, err
	}

	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(scale.Name).SubResource("scale").Body(encodedBytes).Do().Into(result)
	return
}

// UpdateStatus updates the status for an existing deploymentConfig.
func (c *deploymentConfigs) UpdateStatus(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(deploymentConfig.Name).SubResource("status").Body(deploymentConfig).Do().Into(result)
	return
}

// Instantiate instantiates a new build from build config returning new object or an error
func (c *deploymentConfigs) Instantiate(request *deployapi.DeploymentRequest) (*deployapi.DeploymentConfig, error) {
	result := &deployapi.DeploymentConfig{}
	resp := c.r.Post().Namespace(c.ns).Resource("deploymentConfigs").Name(request.Name).SubResource("instantiate").Body(request).Do()
	var statusCode int
	if resp.StatusCode(&statusCode); statusCode == 204 {
		return nil, nil
	}
	err := resp.Into(result)
	return result, err
}

type updateConfigFunc func(d *deployapi.DeploymentConfig)

// UpdateConfigWithRetries will try to update a deployment config and ignore any update conflicts.
func UpdateConfigWithRetries(dn DeploymentConfigsNamespacer, namespace, name string, applyUpdate updateConfigFunc) (*deployapi.DeploymentConfig, error) {
	var config *deployapi.DeploymentConfig

	resultErr := kclient.RetryOnConflict(kclient.DefaultBackoff, func() error {
		var err error
		config, err = dn.DeploymentConfigs(namespace).Get(name)
		if err != nil {
			return err
		}
		// Apply the update, then attempt to push it to the apiserver.
		applyUpdate(config)
		config, err = dn.DeploymentConfigs(namespace).Update(config)
		return err
	})
	return config, resultErr
}
