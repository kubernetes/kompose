package api

import "k8s.io/kubernetes/pkg/fields"

// RouteToSelectableFields returns a label set that represents the object
func RouteToSelectableFields(route *Route) fields.Set {
	return fields.Set{
		"metadata.name":      route.Name,
		"metadata.namespace": route.Namespace,
		"spec.path":          route.Spec.Path,
		"spec.host":          route.Spec.Host,
		"spec.to.name":       route.Spec.To.Name,
	}
}
