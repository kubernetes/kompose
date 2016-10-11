package client

import (
	"github.com/openshift/origin/pkg/image/api"
)

// ImageStreamTagsNamespacer has methods to work with ImageStreamTag resources in a namespace
type ImageStreamTagsNamespacer interface {
	ImageStreamTags(namespace string) ImageStreamTagInterface
}

// ImageStreamTagInterface exposes methods on ImageStreamTag resources.
type ImageStreamTagInterface interface {
	Get(name, tag string) (*api.ImageStreamTag, error)
	Create(tag *api.ImageStreamTag) (*api.ImageStreamTag, error)
	Update(tag *api.ImageStreamTag) (*api.ImageStreamTag, error)
	Delete(name, tag string) error
}

// imageStreamTags implements ImageStreamTagsNamespacer interface
type imageStreamTags struct {
	r  *Client
	ns string
}

// newImageStreamTags returns an imageStreamTags
func newImageStreamTags(c *Client, namespace string) *imageStreamTags {
	return &imageStreamTags{
		r:  c,
		ns: namespace,
	}
}

// Get finds the specified image by name of an image stream and tag.
func (c *imageStreamTags) Get(name, tag string) (result *api.ImageStreamTag, err error) {
	result = &api.ImageStreamTag{}
	err = c.r.Get().Namespace(c.ns).Resource("imageStreamTags").Name(api.JoinImageStreamTag(name, tag)).Do().Into(result)
	return
}

// Update updates an image stream tag (creating it if it does not exist).
func (c *imageStreamTags) Update(tag *api.ImageStreamTag) (result *api.ImageStreamTag, err error) {
	result = &api.ImageStreamTag{}
	err = c.r.Put().Namespace(c.ns).Resource("imageStreamTags").Name(tag.Name).Body(tag).Do().Into(result)
	return
}

func (c *imageStreamTags) Create(tag *api.ImageStreamTag) (result *api.ImageStreamTag, err error) {
	result = &api.ImageStreamTag{}
	err = c.r.Post().Namespace(c.ns).Resource("imageStreamTags").Body(tag).Do().Into(result)
	return
}

// Delete deletes the specified tag from the image stream.
func (c *imageStreamTags) Delete(name, tag string) error {
	return c.r.Delete().Namespace(c.ns).Resource("imageStreamTags").Name(api.JoinImageStreamTag(name, tag)).Do().Error()
}
