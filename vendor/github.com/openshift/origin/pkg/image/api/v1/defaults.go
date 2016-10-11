package v1

import (
	newer "github.com/openshift/origin/pkg/image/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

func SetDefaults_ImageImportSpec(obj *ImageImportSpec) {
	if obj.To == nil {
		if ref, err := newer.ParseDockerImageReference(obj.From.Name); err == nil {
			if len(ref.Tag) > 0 {
				obj.To = &v1.LocalObjectReference{Name: ref.Tag}
			}
		}
	}
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return scheme.AddDefaultingFuncs(
		SetDefaults_ImageImportSpec,
	)
}
