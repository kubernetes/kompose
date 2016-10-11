package client

import (
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// RolesNamespacer has methods to work with Role resources in a namespace
type RolesNamespacer interface {
	Roles(namespace string) RoleInterface
}

// RoleInterface exposes methods on Role resources.
type RoleInterface interface {
	List(opts kapi.ListOptions) (*authorizationapi.RoleList, error)
	Get(name string) (*authorizationapi.Role, error)
	Create(role *authorizationapi.Role) (*authorizationapi.Role, error)
	Update(role *authorizationapi.Role) (*authorizationapi.Role, error)
	Delete(name string) error
}

// roles implements RolesNamespacer interface
type roles struct {
	r  *Client
	ns string
}

// newRoles returns a roles
func newRoles(c *Client, namespace string) *roles {
	return &roles{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of roles that match the label and field selectors.
func (c *roles) List(opts kapi.ListOptions) (result *authorizationapi.RoleList, err error) {
	result = &authorizationapi.RoleList{}
	err = c.r.Get().Namespace(c.ns).Resource("roles").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

// Get returns information about a particular role and error if one occurs.
func (c *roles) Get(name string) (result *authorizationapi.Role, err error) {
	result = &authorizationapi.Role{}
	err = c.r.Get().Namespace(c.ns).Resource("roles").Name(name).Do().Into(result)
	return
}

// Create creates new role. Returns the server's representation of the role and error if one occurs.
func (c *roles) Create(role *authorizationapi.Role) (result *authorizationapi.Role, err error) {
	result = &authorizationapi.Role{}
	err = c.r.Post().Namespace(c.ns).Resource("roles").Body(role).Do().Into(result)
	return
}

// Update updates the role on server. Returns the server's representation of the role and error if one occurs.
func (c *roles) Update(role *authorizationapi.Role) (result *authorizationapi.Role, err error) {
	result = &authorizationapi.Role{}
	err = c.r.Put().Namespace(c.ns).Resource("roles").Name(role.Name).Body(role).Do().Into(result)
	return
}

// Delete deletes a role, returns error if one occurs.
func (c *roles) Delete(name string) (err error) {
	err = c.r.Delete().Namespace(c.ns).Resource("roles").Name(name).Do().Error()
	return
}
