package client

import (
	imageapi "github.com/openshift/origin/pkg/image/api"
)

// ImageStreamMappingsNamespacer has methods to work with ImageStreamMapping resources in a namespace
type ImageStreamMappingsNamespacer interface {
	ImageStreamMappings(namespace string) ImageStreamMappingInterface
}

// ImageStreamMappingInterface exposes methods on ImageStreamMapping resources.
type ImageStreamMappingInterface interface {
	Create(mapping *imageapi.ImageStreamMapping) error
}

// imageStreamMappings implements ImageStreamMappingsNamespacer interface
type imageStreamMappings struct {
	r  *Client
	ns string
}

// newImageStreamMappings returns an imageStreamMappings
func newImageStreamMappings(c *Client, namespace string) *imageStreamMappings {
	return &imageStreamMappings{
		r:  c,
		ns: namespace,
	}
}

// Create creates a new image stream mapping on the server. Returns error if one occurs.
func (c *imageStreamMappings) Create(mapping *imageapi.ImageStreamMapping) error {
	return c.r.Post().Namespace(c.ns).Resource("imageStreamMappings").Body(mapping).Do().Error()
}
