package api

import "k8s.io/kubernetes/pkg/fields"

// ClusterNetworkToSelectableFields returns a label set that represents the object
func ClusterNetworkToSelectableFields(network *ClusterNetwork) fields.Set {
	return fields.Set{
		"metadata.name": network.Name,
	}
}

// HostSubnetToSelectableFields returns a label set that represents the object
func HostSubnetToSelectableFields(obj *HostSubnet) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
	}
}

// NetNamespaceToSelectableFields returns a label set that represents the object
func NetNamespaceToSelectableFields(obj *NetNamespace) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
	}
}
