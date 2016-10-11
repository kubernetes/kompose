package api

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""
const FutureGroupName = "networking.openshift.io"

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
		&ClusterNetwork{},
		&ClusterNetworkList{},
		&HostSubnet{},
		&HostSubnetList{},
		&NetNamespace{},
		&NetNamespaceList{},
		&EgressNetworkPolicy{},
		&EgressNetworkPolicyList{},
	)
	return nil
}

func (obj *ClusterNetwork) GetObjectKind() unversioned.ObjectKind          { return &obj.TypeMeta }
func (obj *ClusterNetworkList) GetObjectKind() unversioned.ObjectKind      { return &obj.TypeMeta }
func (obj *HostSubnet) GetObjectKind() unversioned.ObjectKind              { return &obj.TypeMeta }
func (obj *HostSubnetList) GetObjectKind() unversioned.ObjectKind          { return &obj.TypeMeta }
func (obj *NetNamespace) GetObjectKind() unversioned.ObjectKind            { return &obj.TypeMeta }
func (obj *NetNamespaceList) GetObjectKind() unversioned.ObjectKind        { return &obj.TypeMeta }
func (obj *EgressNetworkPolicy) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *EgressNetworkPolicyList) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
