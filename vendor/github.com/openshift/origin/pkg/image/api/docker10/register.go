package docker10

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: "1.0"}

func AddToScheme(scheme *runtime.Scheme) {
	addKnownTypes(scheme)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&DockerImage{},
	)
}

func (obj *DockerImage) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
