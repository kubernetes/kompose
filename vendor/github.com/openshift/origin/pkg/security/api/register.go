package api

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""

// TODO no one likes the name security because it so broad as to be meaningless.
// const FutureGroupName = "security.openshift.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) unversioned.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns back a Group qualified GroupResource
func Resource(resource string) unversioned.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&PodSecurityPolicySubjectReview{},
		&PodSecurityPolicySelfSubjectReview{},
		&PodSecurityPolicyReview{},
	)
	return nil
}

func (obj *PodSecurityPolicySubjectReview) GetObjectKind() unversioned.ObjectKind {
	return &obj.TypeMeta
}
func (obj *PodSecurityPolicySelfSubjectReview) GetObjectKind() unversioned.ObjectKind {
	return &obj.TypeMeta
}
func (obj *PodSecurityPolicyReview) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
