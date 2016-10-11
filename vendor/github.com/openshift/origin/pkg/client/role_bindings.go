package client

import (
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// RoleBindingsNamespacer has methods to work with RoleBinding resources in a namespace
type RoleBindingsNamespacer interface {
	RoleBindings(namespace string) RoleBindingInterface
}

// RoleBindingInterface exposes methods on RoleBinding resources.
type RoleBindingInterface interface {
	List(opts kapi.ListOptions) (*authorizationapi.RoleBindingList, error)
	Get(name string) (*authorizationapi.RoleBinding, error)
	Create(roleBinding *authorizationapi.RoleBinding) (*authorizationapi.RoleBinding, error)
	Update(roleBinding *authorizationapi.RoleBinding) (*authorizationapi.RoleBinding, error)
	Delete(name string) error
}

// roleBindings implements RoleBindingsNamespacer interface
type roleBindings struct {
	r  *Client
	ns string
}

// newRoleBindings returns a roleBindings
func newRoleBindings(c *Client, namespace string) *roleBindings {
	return &roleBindings{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of roleBindings that match the label and field selectors.
func (c *roleBindings) List(opts kapi.ListOptions) (result *authorizationapi.RoleBindingList, err error) {
	result = &authorizationapi.RoleBindingList{}
	err = c.r.Get().Namespace(c.ns).Resource("roleBindings").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

// Get returns information about a particular roleBinding and error if one occurs.
func (c *roleBindings) Get(name string) (result *authorizationapi.RoleBinding, err error) {
	result = &authorizationapi.RoleBinding{}
	err = c.r.Get().Namespace(c.ns).Resource("roleBindings").Name(name).Do().Into(result)
	return
}

// Create creates new roleBinding. Returns the server's representation of the roleBinding and error if one occurs.
func (c *roleBindings) Create(roleBinding *authorizationapi.RoleBinding) (result *authorizationapi.RoleBinding, err error) {
	result = &authorizationapi.RoleBinding{}
	err = c.r.Post().Namespace(c.ns).Resource("roleBindings").Body(roleBinding).Do().Into(result)
	return
}

// Update updates the roleBinding on server. Returns the server's representation of the roleBinding and error if one occurs.
func (c *roleBindings) Update(roleBinding *authorizationapi.RoleBinding) (result *authorizationapi.RoleBinding, err error) {
	result = &authorizationapi.RoleBinding{}
	err = c.r.Put().Namespace(c.ns).Resource("roleBindings").Name(roleBinding.Name).Body(roleBinding).Do().Into(result)
	return
}

// Delete deletes a roleBinding, returns error if one occurs.
func (c *roleBindings) Delete(name string) (err error) {
	err = c.r.Delete().Namespace(c.ns).Resource("roleBindings").Name(name).Do().Error()
	return
}
