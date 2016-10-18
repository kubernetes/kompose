package client

import (
	"github.com/openshift/origin/pkg/image/api"
)

// ImageStreamImagesNamespacer has methods to work with ImageStreamImage resources in a namespace
type ImageStreamImagesNamespacer interface {
	ImageStreamImages(namespace string) ImageStreamImageInterface
}

// ImageStreamImageInterface exposes methods on ImageStreamImage resources.
type ImageStreamImageInterface interface {
	Get(name, id string) (*api.ImageStreamImage, error)
}

// imageStreamImages implements ImageStreamImagesNamespacer interface
type imageStreamImages struct {
	r  *Client
	ns string
}

// newImageStreamImages returns an imageStreamImages
func newImageStreamImages(c *Client, namespace string) *imageStreamImages {
	return &imageStreamImages{
		r:  c,
		ns: namespace,
	}
}

// Get finds the specified image by name of an image repository and id.
func (c *imageStreamImages) Get(name, id string) (result *api.ImageStreamImage, err error) {
	result = &api.ImageStreamImage{}
	err = c.r.Get().Namespace(c.ns).Resource("imageStreamImages").Name(api.MakeImageStreamImageName(name, id)).Do().Into(result)
	return
}
