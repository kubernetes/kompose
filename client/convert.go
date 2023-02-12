package client

import (
	"fmt"

	"github.com/kubernetes/kompose/pkg/app"
	"github.com/kubernetes/kompose/pkg/kobject"
)

func (k *Kompose) Convert(options ConvertOptions) error {
	err := k.validateOptions(options)
	if err != nil {
		return err
	}

	kobjectConvertOptions := kobject.ConvertOptions{
		ToStdout:               options.ToStdout,
		Build:                  options.Build,
		PushImage:              options.PushImage,
		PushImageRegistry:      options.PushImageRegistry,
		GenerateJSON:           options.GenerateJson,
		OutFile:                options.OutFile,
		Replicas:               options.Replicas,
		Volumes:                options.VolumeType,
		PVCRequestSize:         options.PvcRequestSize,
		WithKomposeAnnotation:  options.WithKomposeAnnotations,
		CreateD:                k.createDeployment(options),
		CreateDS:               k.createDaemonSet(options),
		CreateRC:               k.createReplicationController(options),
		CreateDeploymentConfig: k.createDeploymentConfig(options),
		CreateChart:            k.createChart(options),
		InputFiles:             options.InputFiles,
		MultipleContainerMode:  k.multiContainerMode(options),
		ServiceGroupMode:       k.serviceGroupMode(options),
		ServiceGroupName:       k.serviceGroupName(options),
		SecretsAsFiles:         k.secretsAsFiles(options),
		InsecureRepository:     k.insecureRepository(options),
		BuildRepo:              k.buildRepo(options),
		BuildBranch:            k.buildBranch(options),
	}
	app.Convert(kobjectConvertOptions)
	return nil
}

func (k *Kompose) validateOptions(options ConvertOptions) error {
	build := options.Build
	if build != string(LOCAL) && build != string(BUILD_CONFIG) && build != string(NONE) {
		return fmt.Errorf(
			"unexpected Value for Build field. Possible values are: %v, %v, and %v", string(LOCAL), string(BUILD_CONFIG), string(NONE),
		)
	}

	volumeType := options.VolumeType
	if volumeType != string(PVC) && volumeType != string(EMPTYDIR) && volumeType != string(HOSTPATH) && volumeType != string(CONFIGMAP) {
		return fmt.Errorf(
			"unexpected Value for VolumeType field. Possible values are: %v, %v, %v, %v", string(PVC), string(EMPTYDIR), string(HOSTPATH), string(CONFIGMAP),
		)
	}

	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		kubernetesController := kubernetesProvider.Controller
		if kubernetesController != string(DEPLOYMENT) && kubernetesController != string(DAEMONSET) && kubernetesController != string(REPLICATION_CONTROLLER) {
			return fmt.Errorf(
				"unexpected Value for Kubernetes Controller field. Possible values are: %v, %v, and %v", string(DEPLOYMENT), string(DAEMONSET), string(REPLICATION_CONTROLLER),
			)
		}

		kubernetesServiceGroupMode := kubernetesProvider.ServiceGroupMode
		if kubernetesServiceGroupMode != string(LABEL) && kubernetesServiceGroupMode != string(VOLUME) {
			return fmt.Errorf(
				"unexpected Value for Kubernetes Service Groupe Mode field. Possible values are: %v and %v", string(LABEL), string(VOLUME),
			)
		}

		if build == string(BUILD_CONFIG) {
			return fmt.Errorf("the build value 'build-config' is only supported for Openshift provider")
		}
	}

	return nil
}

func (k *Kompose) createDeployment(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.Controller == string(DEPLOYMENT)
	}
	return false
}

func (k *Kompose) createDaemonSet(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.Controller == string(DAEMONSET)
	}
	return false
}

func (k *Kompose) createReplicationController(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.Controller == string(REPLICATION_CONTROLLER)
	}
	return false
}

func (k *Kompose) createChart(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.Chart
	}
	return false
}

func (k *Kompose) multiContainerMode(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.MultiContainerMode
	}
	return false
}

func (k *Kompose) serviceGroupMode(options ConvertOptions) string {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.ServiceGroupMode
	}
	return ""
}

func (k *Kompose) serviceGroupName(options ConvertOptions) string {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.ServiceGroupName
	}
	return ""
}

func (k *Kompose) secretsAsFiles(options ConvertOptions) bool {
	if kubernetesProvider, ok := options.Provider.(Kubernetes); ok {
		return kubernetesProvider.SecretsAsFiles
	}
	return false
}

func (k *Kompose) createDeploymentConfig(options ConvertOptions) bool {
	if _, ok := options.Provider.(Openshift); ok {
		return true
	}
	return false
}

func (k *Kompose) insecureRepository(options ConvertOptions) bool {
	if openshiftProvider, ok := options.Provider.(Openshift); ok {
		return openshiftProvider.InsecureRepository
	}
	return false
}

func (k *Kompose) buildRepo(options ConvertOptions) string {
	if openshiftProvider, ok := options.Provider.(Openshift); ok {
		return openshiftProvider.BuildRepo
	}
	return ""
}

func (k *Kompose) buildBranch(options ConvertOptions) string {
	if openshiftProvider, ok := options.Provider.(Openshift); ok {
		return openshiftProvider.BuildRepo
	}
	return ""
}
