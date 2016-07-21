package api

import "k8s.io/kubernetes/pkg/fields"

// ClusterPolicyToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func ClusterPolicyToSelectableFields(policy *ClusterPolicy) fields.Set {
	return fields.Set{
		"metadata.name": policy.Name,
	}
}

// ClusterPolicyBindingToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func ClusterPolicyBindingToSelectableFields(policyBinding *ClusterPolicyBinding) fields.Set {
	return fields.Set{
		"metadata.name": policyBinding.Name,
	}
}

// PolicyToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func PolicyToSelectableFields(policy *Policy) fields.Set {
	return fields.Set{
		"metadata.name":      policy.Name,
		"metadata.namespace": policy.Namespace,
	}
}

// PolicyBindingToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func PolicyBindingToSelectableFields(policyBinding *PolicyBinding) fields.Set {
	return fields.Set{
		"metadata.name":       policyBinding.Name,
		"metadata.namespace":  policyBinding.Namespace,
		"policyRef.namespace": policyBinding.PolicyRef.Namespace,
	}
}

// RoleToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func RoleToSelectableFields(role *Role) fields.Set {
	return fields.Set{
		"metadata.name":      role.Name,
		"metadata.namespace": role.Namespace,
	}
}

// RoleBindingToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func RoleBindingToSelectableFields(roleBinding *RoleBinding) fields.Set {
	return fields.Set{
		"metadata.name":      roleBinding.Name,
		"metadata.namespace": roleBinding.Namespace,
	}
}
