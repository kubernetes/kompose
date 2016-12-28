package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: "v1"}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes, addConversionFuncs, addDefaultingFuncs)
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
