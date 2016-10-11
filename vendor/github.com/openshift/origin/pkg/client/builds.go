package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	buildapi "github.com/openshift/origin/pkg/build/api"
)

// BuildsNamespacer has methods to work with Build resources in a namespace
type BuildsNamespacer interface {
	Builds(namespace string) BuildInterface
}

// BuildInterface exposes methods on Build resources.
type BuildInterface interface {
	List(opts kapi.ListOptions) (*buildapi.BuildList, error)
	Get(name string) (*buildapi.Build, error)
	Create(build *buildapi.Build) (*buildapi.Build, error)
	Update(build *buildapi.Build) (*buildapi.Build, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
	Clone(request *buildapi.BuildRequest) (*buildapi.Build, error)
	UpdateDetails(build *buildapi.Build) (*buildapi.Build, error)
}

// builds implements BuildsNamespacer interface
type builds struct {
	r  *Client
	ns string
}

// newBuilds returns a builds
func newBuilds(c *Client, namespace string) *builds {
	return &builds{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of builds that match the label and field selectors.
func (c *builds) List(opts kapi.ListOptions) (result *buildapi.BuildList, err error) {
	result = &buildapi.BuildList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("builds").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular build and error if one occurs.
func (c *builds) Get(name string) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Get().Namespace(c.ns).Resource("builds").Name(name).Do().Into(result)
	return
}

// Create creates new build. Returns the server's representation of the build and error if one occurs.
func (c *builds) Create(build *buildapi.Build) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Post().Namespace(c.ns).Resource("builds").Body(build).Do().Into(result)
	return
}

// Update updates the build on server. Returns the server's representation of the build and error if one occurs.
func (c *builds) Update(build *buildapi.Build) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Put().Namespace(c.ns).Resource("builds").Name(build.Name).Body(build).Do().Into(result)
	return
}

// Delete deletes a build, returns error if one occurs.
func (c *builds) Delete(name string) (err error) {
	err = c.r.Delete().Namespace(c.ns).Resource("builds").Name(name).Do().Error()
	return
}

// Watch returns a watch.Interface that watches the requested builds
func (c *builds) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("builds").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}

// Clone creates a clone of a build returning new object or an error
func (c *builds) Clone(request *buildapi.BuildRequest) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Post().Namespace(c.ns).Resource("builds").Name(request.Name).SubResource("clone").Body(request).Do().Into(result)
	return
}

// UpdateDetails updates the build details for a given build.
// Currently only the Spec.Revision is allowed to be updated.
// Returns the server's representation of the build and error if one occurs.
func (c *builds) UpdateDetails(build *buildapi.Build) (result *buildapi.Build, err error) {
	result = &buildapi.Build{}
	err = c.r.Put().Namespace(c.ns).Resource("builds").Name(build.Name).SubResource("details").Body(build).Do().Into(result)
	return
}
