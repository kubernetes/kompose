package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
)

// policies

func ToPolicyList(in *ClusterPolicyList) *PolicyList {
	ret := &PolicyList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToPolicy(&curr))
	}

	return ret
}

func ToPolicy(in *ClusterPolicy) *Policy {
	if in == nil {
		return nil
	}

	ret := &Policy{}
	ret.ObjectMeta = in.ObjectMeta
	ret.LastModified = in.LastModified
	ret.Roles = ToRoleMap(in.Roles)

	return ret
}

func ToRoleMap(in map[string]*ClusterRole) map[string]*Role {
	ret := map[string]*Role{}
	for key, role := range in {
		ret[key] = ToRole(role)
	}

	return ret
}

func ToRoleList(in *ClusterRoleList) *RoleList {
	ret := &RoleList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToRole(&curr))
	}

	return ret
}

func ToRole(in *ClusterRole) *Role {
	if in == nil {
		return nil
	}

	ret := &Role{}
	ret.ObjectMeta = in.ObjectMeta
	ret.Rules = in.Rules

	return ret
}

func ToClusterPolicyList(in *PolicyList) *ClusterPolicyList {
	ret := &ClusterPolicyList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToClusterPolicy(&curr))
	}

	return ret
}

func ToClusterPolicy(in *Policy) *ClusterPolicy {
	if in == nil {
		return nil
	}

	ret := &ClusterPolicy{}
	ret.ObjectMeta = in.ObjectMeta
	ret.LastModified = in.LastModified
	ret.Roles = ToClusterRoleMap(in.Roles)

	return ret
}

func ToClusterRoleMap(in map[string]*Role) map[string]*ClusterRole {
	ret := map[string]*ClusterRole{}
	for key, role := range in {
		ret[key] = ToClusterRole(role)
	}

	return ret
}

func ToClusterRoleList(in *RoleList) *ClusterRoleList {
	ret := &ClusterRoleList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToClusterRole(&curr))
	}

	return ret
}

func ToClusterRole(in *Role) *ClusterRole {
	if in == nil {
		return nil
	}

	ret := &ClusterRole{}
	ret.ObjectMeta = in.ObjectMeta
	ret.Rules = in.Rules

	return ret
}

// policy bindings

func ToPolicyBindingList(in *ClusterPolicyBindingList) *PolicyBindingList {
	ret := &PolicyBindingList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToPolicyBinding(&curr))
	}

	return ret
}

func ToPolicyBinding(in *ClusterPolicyBinding) *PolicyBinding {
	if in == nil {
		return nil
	}

	ret := &PolicyBinding{}
	ret.ObjectMeta = in.ObjectMeta
	ret.LastModified = in.LastModified
	ret.PolicyRef = ToPolicyRef(in.PolicyRef)
	ret.RoleBindings = ToRoleBindingMap(in.RoleBindings)

	return ret
}

func ToPolicyRef(in kapi.ObjectReference) kapi.ObjectReference {
	ret := kapi.ObjectReference{}

	ret.Name = in.Name
	return ret
}

func ToRoleBindingMap(in map[string]*ClusterRoleBinding) map[string]*RoleBinding {
	ret := map[string]*RoleBinding{}
	for key, RoleBinding := range in {
		ret[key] = ToRoleBinding(RoleBinding)
	}

	return ret
}

func ToRoleBindingList(in *ClusterRoleBindingList) *RoleBindingList {
	ret := &RoleBindingList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToRoleBinding(&curr))
	}

	return ret
}

func ToRoleBinding(in *ClusterRoleBinding) *RoleBinding {
	if in == nil {
		return nil
	}

	ret := &RoleBinding{}
	ret.ObjectMeta = in.ObjectMeta
	ret.Subjects = in.Subjects
	ret.RoleRef = ToRoleRef(in.RoleRef)
	return ret
}

func ToRoleRef(in kapi.ObjectReference) kapi.ObjectReference {
	ret := kapi.ObjectReference{}

	ret.Name = in.Name
	return ret
}

func ToClusterPolicyBindingList(in *PolicyBindingList) *ClusterPolicyBindingList {
	ret := &ClusterPolicyBindingList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToClusterPolicyBinding(&curr))
	}

	return ret
}

func ToClusterPolicyBinding(in *PolicyBinding) *ClusterPolicyBinding {
	if in == nil {
		return nil
	}

	ret := &ClusterPolicyBinding{}
	ret.ObjectMeta = in.ObjectMeta
	ret.LastModified = in.LastModified
	ret.PolicyRef = ToClusterPolicyRef(in.PolicyRef)
	ret.RoleBindings = ToClusterRoleBindingMap(in.RoleBindings)

	return ret
}

func ToClusterPolicyRef(in kapi.ObjectReference) kapi.ObjectReference {
	ret := kapi.ObjectReference{}

	ret.Name = in.Name
	return ret
}

func ToClusterRoleBindingMap(in map[string]*RoleBinding) map[string]*ClusterRoleBinding {
	ret := map[string]*ClusterRoleBinding{}
	for key, RoleBinding := range in {
		ret[key] = ToClusterRoleBinding(RoleBinding)
	}

	return ret
}

func ToClusterRoleBindingList(in *RoleBindingList) *ClusterRoleBindingList {
	ret := &ClusterRoleBindingList{}
	for _, curr := range in.Items {
		ret.Items = append(ret.Items, *ToClusterRoleBinding(&curr))
	}

	return ret
}

func ToClusterRoleBinding(in *RoleBinding) *ClusterRoleBinding {
	if in == nil {
		return nil
	}

	ret := &ClusterRoleBinding{}
	ret.ObjectMeta = in.ObjectMeta
	ret.Subjects = in.Subjects
	ret.RoleRef = ToClusterRoleRef(in.RoleRef)

	return ret
}

func ToClusterRoleRef(in kapi.ObjectReference) kapi.ObjectReference {
	ret := kapi.ObjectReference{}

	ret.Name = in.Name
	return ret
}
