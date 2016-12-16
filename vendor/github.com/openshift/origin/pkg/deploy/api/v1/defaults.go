package v1

import (
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

// Keep this in sync with pkg/api/serialization_test.go#defaultHookContainerName
func defaultHookContainerName(hook *LifecycleHook, containerName string) {
	if hook == nil {
		return
	}
	for i := range hook.TagImages {
		if len(hook.TagImages[i].ContainerName) == 0 {
			hook.TagImages[i].ContainerName = containerName
		}
	}
	if hook.ExecNewPod != nil {
		if len(hook.ExecNewPod.ContainerName) == 0 {
			hook.ExecNewPod.ContainerName = containerName
		}
	}
}

func SetDefaults_DeploymentConfigSpec(obj *DeploymentConfigSpec) {
	if obj.Triggers == nil {
		obj.Triggers = []DeploymentTriggerPolicy{
			{Type: DeploymentTriggerOnConfigChange},
		}
	}
	if len(obj.Selector) == 0 && obj.Template != nil {
		obj.Selector = obj.Template.Labels
	}

	// if you only specify a single container, default the TagImages hook to the container name
	if obj.Template != nil && len(obj.Template.Spec.Containers) == 1 {
		containerName := obj.Template.Spec.Containers[0].Name
		if p := obj.Strategy.RecreateParams; p != nil {
			defaultHookContainerName(p.Pre, containerName)
			defaultHookContainerName(p.Mid, containerName)
			defaultHookContainerName(p.Post, containerName)
		}
		if p := obj.Strategy.RollingParams; p != nil {
			defaultHookContainerName(p.Pre, containerName)
			defaultHookContainerName(p.Post, containerName)
		}
	}
}

func SetDefaults_DeploymentStrategy(obj *DeploymentStrategy) {
	if len(obj.Type) == 0 {
		obj.Type = DeploymentStrategyTypeRolling
	}

	if obj.Type == DeploymentStrategyTypeRolling && obj.RollingParams == nil {
		obj.RollingParams = &RollingDeploymentStrategyParams{
			IntervalSeconds:     mkintp(deployapi.DefaultRollingIntervalSeconds),
			UpdatePeriodSeconds: mkintp(deployapi.DefaultRollingUpdatePeriodSeconds),
			TimeoutSeconds:      mkintp(deployapi.DefaultRollingTimeoutSeconds),
		}
	}
	if obj.Type == DeploymentStrategyTypeRecreate && obj.RecreateParams == nil {
		obj.RecreateParams = &RecreateDeploymentStrategyParams{}
	}
}

func SetDefaults_RecreateDeploymentStrategyParams(obj *RecreateDeploymentStrategyParams) {
	if obj.TimeoutSeconds == nil {
		obj.TimeoutSeconds = mkintp(deployapi.DefaultRollingTimeoutSeconds)
	}
}

func SetDefaults_RollingDeploymentStrategyParams(obj *RollingDeploymentStrategyParams) {
	if obj.IntervalSeconds == nil {
		obj.IntervalSeconds = mkintp(deployapi.DefaultRollingIntervalSeconds)
	}

	if obj.UpdatePeriodSeconds == nil {
		obj.UpdatePeriodSeconds = mkintp(deployapi.DefaultRollingUpdatePeriodSeconds)
	}

	if obj.TimeoutSeconds == nil {
		obj.TimeoutSeconds = mkintp(deployapi.DefaultRollingTimeoutSeconds)
	}

	if obj.MaxUnavailable == nil && obj.MaxSurge == nil {
		maxUnavailable := intstr.FromString("25%")
		obj.MaxUnavailable = &maxUnavailable

		maxSurge := intstr.FromString("25%")
		obj.MaxSurge = &maxSurge
	}

	if obj.MaxUnavailable == nil && obj.MaxSurge != nil &&
		(*obj.MaxSurge == intstr.FromInt(0) || *obj.MaxSurge == intstr.FromString("0%")) {
		maxUnavailable := intstr.FromString("25%")
		obj.MaxUnavailable = &maxUnavailable
	}

	if obj.MaxSurge == nil && obj.MaxUnavailable != nil &&
		(*obj.MaxUnavailable == intstr.FromInt(0) || *obj.MaxUnavailable == intstr.FromString("0%")) {
		maxSurge := intstr.FromString("25%")
		obj.MaxSurge = &maxSurge
	}
}

func SetDefaults_DeploymentConfig(obj *DeploymentConfig) {
	for _, t := range obj.Spec.Triggers {
		if t.ImageChangeParams != nil {
			if len(t.ImageChangeParams.From.Kind) == 0 {
				t.ImageChangeParams.From.Kind = "ImageStreamTag"
			}
			if len(t.ImageChangeParams.From.Namespace) == 0 {
				t.ImageChangeParams.From.Namespace = obj.Namespace
			}
		}
	}
}

func mkintp(i int64) *int64 {
	return &i
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return scheme.AddDefaultingFuncs(
		SetDefaults_DeploymentConfigSpec,
		SetDefaults_DeploymentStrategy,
		SetDefaults_RecreateDeploymentStrategyParams,
		SetDefaults_RollingDeploymentStrategyParams,
		SetDefaults_DeploymentConfig,
	)
}
