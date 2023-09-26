package client

type ConvertBuild string

const (
	LOCAL        ConvertBuild = "local"
	BUILD_CONFIG ConvertBuild = "build-config"
	NONE         ConvertBuild = "none"
)

type KubernetesController string

const (
	DEPLOYMENT             KubernetesController = "deployment"
	DAEMONSET              KubernetesController = "daemonSet"
	REPLICATION_CONTROLLER KubernetesController = "replicationController"
)

type ServiceGroupMode string

const (
	LABEL  ServiceGroupMode = "label"
	VOLUME ServiceGroupMode = "volume"
)

type VolumeType string

const (
	PVC       = "persistentVolumeClaim"
	EMPTYDIR  = "emptyDir"
	HOSTPATH  = "hostPath"
	CONFIGMAP = "configMap"
)

type ConvertOptions struct {
	Build                  *string
	PushImage              bool
	PushImageRegistry      string
	GenerateJson           bool
	ToStdout               bool
	OutFile                string
	Replicas               *int
	VolumeType             *string
	PvcRequestSize         string
	WithKomposeAnnotations *bool
	InputFiles             []string
	Profiles               []string
	Provider
	GenerateNetworkPolicies bool
}

type Provider interface{}

type Kubernetes struct {
	Provider
	Chart              bool
	Controller         *string
	MultiContainerMode bool
	ServiceGroupMode   *string
	ServiceGroupName   string
	SecretsAsFiles     bool
}

type Openshift struct {
	Provider
	DeploymentConfig   bool
	InsecureRepository bool
	BuildRepo          string
	BuildBranch        string
}
