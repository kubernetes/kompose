package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/openshift/origin/pkg/image/api/docker10"
	"github.com/openshift/origin/pkg/image/api/dockerpre012"
)

const GroupName = ""

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: "v1"}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes, addConversionFuncs, addDefaultingFuncs, docker10.AddToScheme, dockerpre012.AddToScheme)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Image{},
		&ImageList{},
		&ImageSignature{},
		&ImageStream{},
		&ImageStreamList{},
		&ImageStreamMapping{},
		&ImageStreamTag{},
		&ImageStreamTagList{},
		&ImageStreamImage{},
		&ImageStreamImport{},
	)
	return nil
}

func (obj *Image) GetObjectKind() unversioned.ObjectKind              { return &obj.TypeMeta }
func (obj *ImageList) GetObjectKind() unversioned.ObjectKind          { return &obj.TypeMeta }
func (obj *ImageSignature) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *ImageStream) GetObjectKind() unversioned.ObjectKind        { return &obj.TypeMeta }
func (obj *ImageStreamList) GetObjectKind() unversioned.ObjectKind    { return &obj.TypeMeta }
func (obj *ImageStreamMapping) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *ImageStreamTag) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *ImageStreamTagList) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *ImageStreamImage) GetObjectKind() unversioned.ObjectKind   { return &obj.TypeMeta }
func (obj *ImageStreamImport) GetObjectKind() unversioned.ObjectKind  { return &obj.TypeMeta }
