package client

import (
	"errors"

	kapi "k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/watch"

	imageapi "github.com/openshift/origin/pkg/image/api"
	quotautil "github.com/openshift/origin/pkg/quota/util"
)

var ErrImageStreamImportUnsupported = errors.New("the server does not support directly importing images - create an image stream with tags or the dockerImageRepository field set")

// ImageStreamsNamespacer has methods to work with ImageStream resources in a namespace
type ImageStreamsNamespacer interface {
	ImageStreams(namespace string) ImageStreamInterface
}

// ImageStreamInterface exposes methods on ImageStream resources.
type ImageStreamInterface interface {
	List(opts kapi.ListOptions) (*imageapi.ImageStreamList, error)
	Get(name string) (*imageapi.ImageStream, error)
	Create(stream *imageapi.ImageStream) (*imageapi.ImageStream, error)
	Update(stream *imageapi.ImageStream) (*imageapi.ImageStream, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
	UpdateStatus(stream *imageapi.ImageStream) (*imageapi.ImageStream, error)
	Import(isi *imageapi.ImageStreamImport) (*imageapi.ImageStreamImport, error)
}

// ImageStreamNamespaceGetter exposes methods to get ImageStreams by Namespace
type ImageStreamNamespaceGetter interface {
	GetByNamespace(namespace, name string) (*imageapi.ImageStream, error)
}

// imageStreams implements ImageStreamsNamespacer interface
type imageStreams struct {
	r  *Client
	ns string
}

// newImageStreams returns an imageStreams
func newImageStreams(c *Client, namespace string) *imageStreams {
	return &imageStreams{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of image streams that match the label and field selectors.
func (c *imageStreams) List(opts kapi.ListOptions) (result *imageapi.ImageStreamList, err error) {
	result = &imageapi.ImageStreamList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("imageStreams").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular image stream and error if one occurs.
func (c *imageStreams) Get(name string) (result *imageapi.ImageStream, err error) {
	result = &imageapi.ImageStream{}
	err = c.r.Get().Namespace(c.ns).Resource("imageStreams").Name(name).Do().Into(result)
	return
}

// Create create a new image stream. Returns the server's representation of the image stream and error if one occurs.
func (c *imageStreams) Create(stream *imageapi.ImageStream) (result *imageapi.ImageStream, err error) {
	result = &imageapi.ImageStream{}
	err = c.r.Post().Namespace(c.ns).Resource("imageStreams").Body(stream).Do().Into(result)
	return
}

// Update updates the image stream on the server. Returns the server's representation of the image stream and error if one occurs.
func (c *imageStreams) Update(stream *imageapi.ImageStream) (result *imageapi.ImageStream, err error) {
	result = &imageapi.ImageStream{}
	err = c.r.Put().Namespace(c.ns).Resource("imageStreams").Name(stream.Name).Body(stream).Do().Into(result)
	return
}

// Delete deletes an image stream, returns error if one occurs.
func (c *imageStreams) Delete(name string) (err error) {
	err = c.r.Delete().Namespace(c.ns).Resource("imageStreams").Name(name).Do().Error()
	return
}

// Watch returns a watch.Interface that watches the requested image streams.
func (c *imageStreams) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("imageStreams").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}

// UpdateStatus updates the image stream's status. Returns the server's representation of the image stream, and an error, if it occurs.
func (c *imageStreams) UpdateStatus(stream *imageapi.ImageStream) (result *imageapi.ImageStream, err error) {
	result = &imageapi.ImageStream{}
	err = c.r.Put().Namespace(c.ns).Resource("imageStreams").Name(stream.Name).SubResource("status").Body(stream).Do().Into(result)
	return
}

// Import makes a call to the server to retrieve information about the requested images or to perform an import. ImageStreamImport
// will be returned if no actual import was requested (the to fields were not set), or an ImageStream if import was requested.
func (c *imageStreams) Import(isi *imageapi.ImageStreamImport) (*imageapi.ImageStreamImport, error) {
	result := &imageapi.ImageStreamImport{}
	if err := c.r.Post().Namespace(c.ns).Resource("imageStreamImports").Body(isi).Do().Into(result); err != nil {
		return nil, transformUnsupported(err)
	}
	return result, nil
}

// transformUnsupported converts specific error conditions to unsupported
func transformUnsupported(err error) error {
	if err == nil {
		return nil
	}
	if apierrs.IsNotFound(err) {
		status, ok := err.(apierrs.APIStatus)
		if !ok {
			return ErrImageStreamImportUnsupported
		}
		if status.Status().Details == nil || status.Status().Details.Kind == "" {
			return ErrImageStreamImportUnsupported
		}
	}
	// The ImageStreamImport resource exists in v1.1.1 of origin but is not yet
	// enabled by policy. A create request will return a Forbidden(403) error.
	// We want to return ErrImageStreamImportUnsupported to allow fallback behavior
	// in clients.
	if apierrs.IsForbidden(err) && !quotautil.IsErrorQuotaExceeded(err) {
		return ErrImageStreamImportUnsupported
	}
	return err
}
