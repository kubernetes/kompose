package api

import "k8s.io/kubernetes/pkg/fields"

func ClusterResourceQuotaToSelectableFields(quota *ClusterResourceQuota) fields.Set {
	return fields.Set{
		"metadata.name": quota.Name,
	}
}
