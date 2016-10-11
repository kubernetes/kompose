package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	oauthapi "github.com/openshift/origin/pkg/oauth/api"
)

type OAuthClientAuthorizationsInterface interface {
	OAuthClientAuthorizations() OAuthClientAuthorizationInterface
}

type OAuthClientAuthorizationInterface interface {
	Create(obj *oauthapi.OAuthClientAuthorization) (*oauthapi.OAuthClientAuthorization, error)
	List(opts kapi.ListOptions) (*oauthapi.OAuthClientAuthorizationList, error)
	Get(name string) (*oauthapi.OAuthClientAuthorization, error)
	Update(obj *oauthapi.OAuthClientAuthorization) (*oauthapi.OAuthClientAuthorization, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
}

type oauthClientAuthorizations struct {
	r *Client
}

func newOAuthClientAuthorizations(c *Client) *oauthClientAuthorizations {
	return &oauthClientAuthorizations{
		r: c,
	}
}

func (c *oauthClientAuthorizations) Create(obj *oauthapi.OAuthClientAuthorization) (result *oauthapi.OAuthClientAuthorization, err error) {
	result = &oauthapi.OAuthClientAuthorization{}
	err = c.r.Post().Resource("oAuthClientAuthorizations").Body(obj).Do().Into(result)
	return
}

func (c *oauthClientAuthorizations) Update(obj *oauthapi.OAuthClientAuthorization) (result *oauthapi.OAuthClientAuthorization, err error) {
	result = &oauthapi.OAuthClientAuthorization{}
	err = c.r.Put().Resource("oAuthClientAuthorizations").Name(obj.Name).Body(obj).Do().Into(result)
	return
}

func (c *oauthClientAuthorizations) List(opts kapi.ListOptions) (result *oauthapi.OAuthClientAuthorizationList, err error) {
	result = &oauthapi.OAuthClientAuthorizationList{}
	err = c.r.Get().Resource("oAuthClientAuthorizations").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

func (c *oauthClientAuthorizations) Get(name string) (result *oauthapi.OAuthClientAuthorization, err error) {
	result = &oauthapi.OAuthClientAuthorization{}
	err = c.r.Get().Resource("oAuthClientAuthorizations").Name(name).Do().Into(result)
	return
}

func (c *oauthClientAuthorizations) Delete(name string) (err error) {
	err = c.r.Delete().Resource("oAuthClientAuthorizations").Name(name).Do().Error()
	return
}

func (c *oauthClientAuthorizations) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().Prefix("watch").Resource("oAuthClientAuthorizations").VersionedParams(&opts, kapi.ParameterCodec).Watch()
}
