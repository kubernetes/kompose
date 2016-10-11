package client

import (
	templateapi "github.com/openshift/origin/pkg/template/api"
)

// TemplateConfigNamespacer has methods to work with Image resources in a namespace
// TODO: Rename to ProcessedTemplates
type TemplateConfigsNamespacer interface {
	TemplateConfigs(namespace string) TemplateConfigInterface
}

// TemplateConfigInterface exposes methods on Image resources.
type TemplateConfigInterface interface {
	Create(t *templateapi.Template) (*templateapi.Template, error)
}

// templateConfigs implements TemplateConfigsNamespacer interface
type templateConfigs struct {
	r  *Client
	ns string
}

// newTemplateConfigs returns an TemplateConfigInterface
func newTemplateConfigs(c *Client, namespace string) TemplateConfigInterface {
	return &templateConfigs{
		r:  c,
		ns: namespace,
	}
}

// Create process the Template and returns its current state
func (c *templateConfigs) Create(in *templateapi.Template) (*templateapi.Template, error) {
	template := &templateapi.Template{}
	err := c.r.Post().Namespace(c.ns).Resource("processedTemplates").Body(in).Do().Into(template)
	return template, err
}
