package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	projectapi "github.com/openshift/origin/pkg/project/api"
)

// ProjectRequestsInterface has methods to work with ProjectRequest resources in a namespace
type ProjectRequestsInterface interface {
	ProjectRequests() ProjectRequestInterface
}

// ProjectRequestInterface exposes methods on projectRequest resources.
type ProjectRequestInterface interface {
	Create(p *projectapi.ProjectRequest) (*projectapi.Project, error)
	List(opts kapi.ListOptions) (*unversioned.Status, error)
}

type projectRequests struct {
	r *Client
}

// newUsers returns a users
func newProjectRequests(c *Client) *projectRequests {
	return &projectRequests{
		r: c,
	}
}

// Create creates a new Project
func (c *projectRequests) Create(p *projectapi.ProjectRequest) (result *projectapi.Project, err error) {
	result = &projectapi.Project{}
	err = c.r.Post().Resource("projectRequests").Body(p).Do().Into(result)
	return
}

// List returns a status object indicating that a user can call the Create or an error indicating why not
func (c *projectRequests) List(opts kapi.ListOptions) (result *unversioned.Status, err error) {
	result = &unversioned.Status{}
	err = c.r.Get().Resource("projectRequests").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return result, err
}
