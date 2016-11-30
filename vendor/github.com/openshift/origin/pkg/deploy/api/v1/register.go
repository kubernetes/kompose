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
		&DeploymentConfig{},
		&DeploymentConfigList{},
		&DeploymentConfigRollback{},
		&DeploymentRequest{},
		&DeploymentLog{},
		&DeploymentLogOptions{},
	)
	return nil
}

func (obj *DeploymentConfig) GetObjectKind() unversioned.ObjectKind         { return &obj.TypeMeta }
func (obj *DeploymentConfigList) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *DeploymentConfigRollback) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *DeploymentRequest) GetObjectKind() unversioned.ObjectKind        { return &obj.TypeMeta }
func (obj *DeploymentLog) GetObjectKind() unversioned.ObjectKind            { return &obj.TypeMeta }
func (obj *DeploymentLogOptions) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
