package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	oauthapi "github.com/openshift/origin/pkg/oauth/api"
)

type OAuthClientsInterface interface {
	OAuthClients() OAuthClientInterface
}

type OAuthClientInterface interface {
	Create(obj *oauthapi.OAuthClient) (*oauthapi.OAuthClient, error)
	List(opts kapi.ListOptions) (*oauthapi.OAuthClientList, error)
	Get(name string) (*oauthapi.OAuthClient, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
	Update(client *oauthapi.OAuthClient) (*oauthapi.OAuthClient, error)
}

type oauthClients struct {
	r *Client
}

func newOAuthClients(c *Client) *oauthClients {
	return &oauthClients{
		r: c,
	}
}

func (c *oauthClients) Create(obj *oauthapi.OAuthClient) (result *oauthapi.OAuthClient, err error) {
	result = &oauthapi.OAuthClient{}
	err = c.r.Post().Resource("oAuthClients").Body(obj).Do().Into(result)
	return
}

func (c *oauthClients) List(opts kapi.ListOptions) (result *oauthapi.OAuthClientList, err error) {
	result = &oauthapi.OAuthClientList{}
	err = c.r.Get().Resource("oAuthClients").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

func (c *oauthClients) Get(name string) (result *oauthapi.OAuthClient, err error) {
	result = &oauthapi.OAuthClient{}
	err = c.r.Get().Resource("oAuthClients").Name(name).Do().Into(result)
	return
}

func (c *oauthClients) Delete(name string) (err error) {
	err = c.r.Delete().Resource("oAuthClients").Name(name).Do().Error()
	return
}

func (c *oauthClients) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().Prefix("watch").Resource("oAuthClients").VersionedParams(&opts, kapi.ParameterCodec).Watch()
}

func (c *oauthClients) Update(client *oauthapi.OAuthClient) (result *oauthapi.OAuthClient, err error) {
	result = &oauthapi.OAuthClient{}
	err = c.r.Put().Resource("oAuthClients").Name(client.Name).Body(client).Do().Into(result)
	return
}
