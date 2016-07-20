package api

import "k8s.io/kubernetes/pkg/fields"

// ImageToSelectableFields returns a label set that represents the object.
func ImageToSelectableFields(image *Image) fields.Set {
	return fields.Set{
		"metadata.name":      image.Name,
		"metadata.namespace": image.Namespace,
	}
}

// ImageStreamToSelectableFields returns a label set that represents the object.
func ImageStreamToSelectableFields(ir *ImageStream) fields.Set {
	return fields.Set{
		"metadata.name":                ir.Name,
		"metadata.namespace":           ir.Namespace,
		"spec.dockerImageRepository":   ir.Spec.DockerImageRepository,
		"status.dockerImageRepository": ir.Status.DockerImageRepository,
	}
}
