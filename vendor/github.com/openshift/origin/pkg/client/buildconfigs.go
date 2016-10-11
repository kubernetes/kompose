package client

import (
	"fmt"
	"io"
	"net/url"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	buildapi "github.com/openshift/origin/pkg/build/api"
)

// ErrTriggerIsNotAWebHook is returned when a webhook URL is requested for a trigger
// that is not a webhook type.
var ErrTriggerIsNotAWebHook = fmt.Errorf("the specified trigger is not a webhook")

// BuildConfigsNamespacer has methods to work with BuildConfig resources in a namespace
type BuildConfigsNamespacer interface {
	BuildConfigs(namespace string) BuildConfigInterface
}

// BuildConfigInterface exposes methods on BuildConfig resources
type BuildConfigInterface interface {
	List(opts kapi.ListOptions) (*buildapi.BuildConfigList, error)
	Get(name string) (*buildapi.BuildConfig, error)
	Create(config *buildapi.BuildConfig) (*buildapi.BuildConfig, error)
	Update(config *buildapi.BuildConfig) (*buildapi.BuildConfig, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)

	Instantiate(request *buildapi.BuildRequest) (result *buildapi.Build, err error)
	InstantiateBinary(request *buildapi.BinaryBuildRequestOptions, r io.Reader) (result *buildapi.Build, err error)

	WebHookURL(name string, trigger *buildapi.BuildTriggerPolicy) (*url.URL, error)
}

// buildConfigs implements BuildConfigsNamespacer interface
type buildConfigs struct {
	r  *Client
	ns string
}

// newBuildConfigs returns a buildConfigs
func newBuildConfigs(c *Client, namespace string) *buildConfigs {
	return &buildConfigs{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of buildconfigs that match the label and field selectors.
func (c *buildConfigs) List(opts kapi.ListOptions) (result *buildapi.BuildConfigList, err error) {
	result = &buildapi.BuildConfigList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("buildConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular buildconfig and error if one occurs.
func (c *buildConfigs) Get(name string) (result *buildapi.BuildConfig, err error) {
	result = &buildapi.BuildConfig{}
	err = c.r.Get().Namespace(c.ns).Resource("buildConfigs").Name(name).Do().Into(result)
	return
}

// WebHookURL returns the URL for the provided build config name and trigger policy, or ErrTriggerIsNotAWebHook
// if the trigger is not a webhook type.
func (c *buildConfigs) WebHookURL(name string, trigger *buildapi.BuildTriggerPolicy) (*url.URL, error) {
	switch {
	case trigger.GenericWebHook != nil:
		return c.r.Get().Namespace(c.ns).Resource("buildConfigs").Name(name).SubResource("webhooks").Suffix(trigger.GenericWebHook.Secret, "generic").URL(), nil
	case trigger.GitHubWebHook != nil:
		return c.r.Get().Namespace(c.ns).Resource("buildConfigs").Name(name).SubResource("webhooks").Suffix(trigger.GitHubWebHook.Secret, "github").URL(), nil
	default:
		return nil, ErrTriggerIsNotAWebHook
	}
}

// Create creates a new buildconfig. Returns the server's representation of the buildconfig and error if one occurs.
func (c *buildConfigs) Create(build *buildapi.BuildConfig) (result *buildapi.BuildConfig, err error) {
	result = &buildapi.BuildConfig{}
	err = c.r.Post().Namespace(c.ns).Resource("buildConfigs").Body(build).Do().Into(result)
	return
}

// Update updates the buildconfig on server. Returns the server's representation of the buildconfig and error if one occurs.
func (c *buildConfigs) Update(build *buildapi.BuildConfig) (result *buildapi.BuildConfig, err error) {
	result = &buildapi.BuildConfig{}
	err = c.r.Put().Namespace(c.ns).Resource("buildConfigs").Name(build.Name).Body(build).Do().Into(result)
	return
}

// Delete deletes a BuildConfig, returns error if one occurs.
func (c *buildConfigs) Delete(name string) error {
	return c.r.Delete().Namespace(c.ns).Resource("buildConfigs").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested buildConfigs.
func (c *buildConfigs) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("buildConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}

// Instantiate instantiates a new build from build config returning new object or an error
func (c *buildConfigs) Instantiate(request *buildapi.BuildRequest) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Post().Namespace(c.ns).Resource("buildConfigs").Name(request.Name).SubResource("instantiate").Body(request).Do().Into(result)
	return
}

// InstantiateBinary instantiates a new build from a build config, given a structured request and an input stream,
// and returns the created build or an error.
func (c *buildConfigs) InstantiateBinary(request *buildapi.BinaryBuildRequestOptions, r io.Reader) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("buildConfigs").
		Name(request.Name).
		SubResource("instantiatebinary").
		VersionedParams(request, kapi.ParameterCodec).
		Body(r).Do().Into(result)
	return
}
