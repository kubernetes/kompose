package client

import (
	imageapi "github.com/openshift/origin/pkg/image/api"
)

// ImageSignaturesInterfacer has methods to work with ImageSignature resource.
type ImageSignaturesInterfacer interface {
	ImageSignatures() ImageSignatureInterface
}

// ImageSignatureInterface exposes methods on ImageSignature virtual resource.
type ImageSignatureInterface interface {
	Create(signature *imageapi.ImageSignature) (*imageapi.ImageSignature, error)
	Delete(name string) error
}

// imageSignatures implements ImageSignatureInterface.
type imageSignatures struct {
	r *Client
}

// newImageSignatures returns imageSignatures
func newImageSignatures(c *Client) ImageSignatureInterface {
	return &imageSignatures{
		r: c,
	}
}

// Create creates a new ImageSignature. Returns the server's representation of the signature and error if one
// occurs.
func (c *imageSignatures) Create(signature *imageapi.ImageSignature) (result *imageapi.ImageSignature, err error) {
	result = &imageapi.ImageSignature{}
	err = c.r.Post().Resource("imageSignatures").Body(signature).Do().Into(result)
	return
}

// Delete deletes an ImageSignature, returns error if one occurs.
func (c *imageSignatures) Delete(name string) error {
	return c.r.Delete().Resource("imageSignatures").Name(name).Do().Error()
}
