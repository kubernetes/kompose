package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
)

// ImageStreamSecretsNamespacer has methods to work with ImageStreamSecret resources in a namespace
type ImageStreamSecretsNamespacer interface {
	ImageStreamSecrets(namespace string) ImageStreamSecretInterface
}

// ImageStreamSecretInterface exposes methods on ImageStreamSecret resources.
type ImageStreamSecretInterface interface {
	// Secrets retrieves the secrets for a named image stream with the provided list options.
	Secrets(name string, options kapi.ListOptions) (*kapi.SecretList, error)
}

// imageStreamSecrets implements ImageStreamSecretsNamespacer interface
type imageStreamSecrets struct {
	r  *Client
	ns string
}

// newImageStreamSecrets returns an imageStreamSecrets
func newImageStreamSecrets(c *Client, namespace string) *imageStreamSecrets {
	return &imageStreamSecrets{
		r:  c,
		ns: namespace,
	}
}

// GetSecrets returns a list of secrets for the named image stream
func (c *imageStreamSecrets) Secrets(name string, options kapi.ListOptions) (result *kapi.SecretList, err error) {
	result = &kapi.SecretList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("imageStreams").
		Name(name).
		SubResource("secrets").
		VersionedParams(&options, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}
