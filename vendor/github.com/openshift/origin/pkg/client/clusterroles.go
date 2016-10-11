package client

import (
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// ClusterRolesInterface has methods to work with ClusterRoles resources in a namespace
type ClusterRolesInterface interface {
	ClusterRoles() ClusterRoleInterface
}

// ClusterRoleInterface exposes methods on ClusterRoles resources
type ClusterRoleInterface interface {
	List(opts kapi.ListOptions) (*authorizationapi.ClusterRoleList, error)
	Get(name string) (*authorizationapi.ClusterRole, error)
	Create(role *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, error)
	Update(role *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, error)
	Delete(name string) error
}

type clusterRoles struct {
	r *Client
}

// newClusterRoles returns a clusterRoles
func newClusterRoles(c *Client) *clusterRoles {
	return &clusterRoles{
		r: c,
	}
}

// List returns a list of clusterRoles that match the label and field selectors.
func (c *clusterRoles) List(opts kapi.ListOptions) (result *authorizationapi.ClusterRoleList, err error) {
	result = &authorizationapi.ClusterRoleList{}
	err = c.r.Get().Resource("clusterRoles").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

// Get returns information about a particular role and error if one occurs.
func (c *clusterRoles) Get(name string) (result *authorizationapi.ClusterRole, err error) {
	result = &authorizationapi.ClusterRole{}
	err = c.r.Get().Resource("clusterRoles").Name(name).Do().Into(result)
	return
}

// Create creates new role. Returns the server's representation of the role and error if one occurs.
func (c *clusterRoles) Create(role *authorizationapi.ClusterRole) (result *authorizationapi.ClusterRole, err error) {
	result = &authorizationapi.ClusterRole{}
	err = c.r.Post().Resource("clusterRoles").Body(role).Do().Into(result)
	return
}

// Update updates the roleBinding on server. Returns the server's representation of the roleBinding and error if one occurs.
func (c *clusterRoles) Update(role *authorizationapi.ClusterRole) (result *authorizationapi.ClusterRole, err error) {
	result = &authorizationapi.ClusterRole{}
	err = c.r.Put().Resource("clusterRoles").Name(role.Name).Body(role).Do().Into(result)
	return
}

// Delete deletes a role, returns error if one occurs.
func (c *clusterRoles) Delete(name string) (err error) {
	err = c.r.Delete().Resource("clusterRoles").Name(name).Do().Error()
	return
}
