package client

import (
	oauthapi "github.com/openshift/origin/pkg/oauth/api"
)

type OAuthAuthorizeTokensInterface interface {
	OAuthAuthorizeTokens() OAuthAuthorizeTokenInterface
}

type OAuthAuthorizeTokenInterface interface {
	Create(token *oauthapi.OAuthAuthorizeToken) (*oauthapi.OAuthAuthorizeToken, error)
	Delete(name string) error
}

type oauthAuthorizeTokenInterface struct {
	r *Client
}

func newOAuthAuthorizeTokens(c *Client) *oauthAuthorizeTokenInterface {
	return &oauthAuthorizeTokenInterface{
		r: c,
	}
}

func (c *oauthAuthorizeTokenInterface) Delete(name string) (err error) {
	err = c.r.Delete().Resource("oAuthAuthorizeTokens").Name(name).Do().Error()
	return
}

func (c *oauthAuthorizeTokenInterface) Create(token *oauthapi.OAuthAuthorizeToken) (result *oauthapi.OAuthAuthorizeToken, err error) {
	result = &oauthapi.OAuthAuthorizeToken{}
	err = c.r.Post().Resource("oAuthAuthorizeTokens").Body(token).Do().Into(result)
	return
}
