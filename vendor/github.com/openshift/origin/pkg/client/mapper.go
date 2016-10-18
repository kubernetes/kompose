package client

import (
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/util/sets"
)

// DefaultMultiRESTMapper returns the multi REST mapper with all OpenShift and
// Kubernetes objects already registered.
func DefaultMultiRESTMapper() meta.MultiRESTMapper {
	var restMapper meta.MultiRESTMapper
	seenGroups := sets.String{}
	for _, gv := range registered.EnabledVersions() {
		if seenGroups.Has(gv.Group) {
			continue
		}
		seenGroups.Insert(gv.Group)
		groupMeta, err := registered.Group(gv.Group)
		if err != nil {
			continue
		}
		restMapper = meta.MultiRESTMapper(append(restMapper, groupMeta.RESTMapper))
	}
	return restMapper
}
