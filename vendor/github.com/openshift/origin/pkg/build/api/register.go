package api

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""
const FutureGroupName = "build.openshift.io"

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
		&Build{},
		&BuildList{},
		&BuildConfig{},
		&BuildConfigList{},
		&BuildLog{},
		&BuildRequest{},
		&BuildLogOptions{},
		&BinaryBuildRequestOptions{},
	)
	return nil
}

func (obj *Build) GetObjectKind() unversioned.ObjectKind                     { return &obj.TypeMeta }
func (obj *BuildList) GetObjectKind() unversioned.ObjectKind                 { return &obj.TypeMeta }
func (obj *BuildConfig) GetObjectKind() unversioned.ObjectKind               { return &obj.TypeMeta }
func (obj *BuildConfigList) GetObjectKind() unversioned.ObjectKind           { return &obj.TypeMeta }
func (obj *BuildLog) GetObjectKind() unversioned.ObjectKind                  { return &obj.TypeMeta }
func (obj *BuildRequest) GetObjectKind() unversioned.ObjectKind              { return &obj.TypeMeta }
func (obj *BuildLogOptions) GetObjectKind() unversioned.ObjectKind           { return &obj.TypeMeta }
func (obj *BinaryBuildRequestOptions) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
