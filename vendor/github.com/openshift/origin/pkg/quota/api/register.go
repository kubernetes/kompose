package api

import (
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""

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
		&ClusterResourceQuota{},
		&ClusterResourceQuotaList{},
		&AppliedClusterResourceQuota{},
		&AppliedClusterResourceQuotaList{},
	)
	return nil
}

func (obj *AppliedClusterResourceQuotaList) GetObjectKind() unversioned.ObjectKind {
	return &obj.TypeMeta
}
func (obj *AppliedClusterResourceQuota) GetObjectKind() unversioned.ObjectKind {
	return &obj.TypeMeta
}

func (obj *ClusterResourceQuotaList) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *ClusterResourceQuota) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *ClusterResourceQuota) GetObjectMeta() meta.Object                { return &obj.ObjectMeta }
