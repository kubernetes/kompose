package client

import (
	userapi "github.com/openshift/origin/pkg/user/api"
)

// UserIdentityMappingsInterface has methods to work with UserIdentityMapping resources in a namespace
type UserIdentityMappingsInterface interface {
	UserIdentityMappings() UserIdentityMappingInterface
}

// UserIdentityMappingInterface exposes methods on UserIdentityMapping resources.
type UserIdentityMappingInterface interface {
	Get(string) (*userapi.UserIdentityMapping, error)
	Create(*userapi.UserIdentityMapping) (*userapi.UserIdentityMapping, error)
	Update(*userapi.UserIdentityMapping) (*userapi.UserIdentityMapping, error)
	Delete(string) error
}

// userIdentityMappings implements UserIdentityMappingsNamespacer interface
type userIdentityMappings struct {
	r *Client
}

// newUserIdentityMappings returns a userIdentityMappings
func newUserIdentityMappings(c *Client) *userIdentityMappings {
	return &userIdentityMappings{
		r: c,
	}
}

// Get returns information about a particular mapping or an error
func (c *userIdentityMappings) Get(name string) (result *userapi.UserIdentityMapping, err error) {
	result = &userapi.UserIdentityMapping{}
	err = c.r.Get().Resource("userIdentityMappings").Name(name).Do().Into(result)
	return
}

// Create creates a new mapping. Returns the server's representation of the mapping and error if one occurs.
func (c *userIdentityMappings) Create(mapping *userapi.UserIdentityMapping) (result *userapi.UserIdentityMapping, err error) {
	result = &userapi.UserIdentityMapping{}
	err = c.r.Post().Resource("userIdentityMappings").Body(mapping).Do().Into(result)
	return
}

// Update updates the mapping on server. Returns the server's representation of the mapping and error if one occurs.
func (c *userIdentityMappings) Update(mapping *userapi.UserIdentityMapping) (result *userapi.UserIdentityMapping, err error) {
	result = &userapi.UserIdentityMapping{}
	err = c.r.Put().Resource("userIdentityMappings").Name(mapping.Name).Body(mapping).Do().Into(result)
	return
}

// Delete deletes the mapping on server.
func (c *userIdentityMappings) Delete(name string) (err error) {
	err = c.r.Delete().Resource("userIdentityMappings").Name(name).Do().Error()
	return
}
