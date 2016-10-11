package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	userapi "github.com/openshift/origin/pkg/user/api"
)

// GroupsInterface has methods to work with Group resources
type GroupsInterface interface {
	Groups() GroupInterface
}

// GroupInterface exposes methods on group resources.
type GroupInterface interface {
	List(opts kapi.ListOptions) (*userapi.GroupList, error)
	Get(name string) (*userapi.Group, error)
	Create(group *userapi.Group) (*userapi.Group, error)
	Update(group *userapi.Group) (*userapi.Group, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
}

// groups implements GroupInterface interface
type groups struct {
	r *Client
}

// newGroups returns a groups
func newGroups(c *Client) *groups {
	return &groups{
		r: c,
	}
}

// List returns a list of groups that match the label and field selectors.
func (c *groups) List(opts kapi.ListOptions) (result *userapi.GroupList, err error) {
	result = &userapi.GroupList{}
	err = c.r.Get().
		Resource("groups").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular group or an error
func (c *groups) Get(name string) (result *userapi.Group, err error) {
	result = &userapi.Group{}
	err = c.r.Get().Resource("groups").Name(name).Do().Into(result)
	return
}

// Create creates a new group. Returns the server's representation of the group and error if one occurs.
func (c *groups) Create(group *userapi.Group) (result *userapi.Group, err error) {
	result = &userapi.Group{}
	err = c.r.Post().Resource("groups").Body(group).Do().Into(result)
	return
}

// Update updates the group on server. Returns the server's representation of the group and error if one occurs.
func (c *groups) Update(group *userapi.Group) (result *userapi.Group, err error) {
	result = &userapi.Group{}
	err = c.r.Put().Resource("groups").Name(group.Name).Body(group).Do().Into(result)
	return
}

// Delete takes the name of the groups, and returns an error if one occurs during deletion of the groups
func (c *groups) Delete(name string) error {
	return c.r.Delete().Resource("groups").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested groups.
func (c *groups) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Resource("groups").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}
