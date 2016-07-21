package api

import "k8s.io/kubernetes/pkg/fields"

// TemplateToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func TemplateToSelectableFields(template *Template) fields.Set {
	return fields.Set{
		"metadata.name": template.Name,
	}
}
