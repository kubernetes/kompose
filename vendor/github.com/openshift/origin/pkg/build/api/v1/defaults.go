package v1

import "k8s.io/kubernetes/pkg/runtime"

func SetDefaults_BuildConfigSpec(config *BuildConfigSpec) {
	if len(config.RunPolicy) == 0 {
		config.RunPolicy = BuildRunPolicySerial
	}
}

func SetDefaults_BuildSource(source *BuildSource) {
	if (source != nil) && (source.Type == BuildSourceBinary) && (source.Binary == nil) {
		source.Binary = &BinaryBuildSource{}
	}
}

func SetDefaults_BuildStrategy(strategy *BuildStrategy) {
	if (strategy != nil) && (strategy.Type == DockerBuildStrategyType) && (strategy.DockerStrategy == nil) {
		strategy.DockerStrategy = &DockerBuildStrategy{}
	}
}

func SetDefaults_SourceBuildStrategy(obj *SourceBuildStrategy) {
	if len(obj.From.Kind) == 0 {
		obj.From.Kind = "ImageStreamTag"
	}
}

func SetDefaults_DockerBuildStrategy(obj *DockerBuildStrategy) {
	if obj.From != nil && len(obj.From.Kind) == 0 {
		obj.From.Kind = "ImageStreamTag"
	}
}

func SetDefaults_CustomBuildStrategy(obj *CustomBuildStrategy) {
	if len(obj.From.Kind) == 0 {
		obj.From.Kind = "ImageStreamTag"
	}
}

func SetDefaults_BuildTriggerPolicy(obj *BuildTriggerPolicy) {
	if obj.Type == ImageChangeBuildTriggerType && obj.ImageChange == nil {
		obj.ImageChange = &ImageChangeTrigger{}
	}
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return scheme.AddDefaultingFuncs(
		SetDefaults_BuildConfigSpec,
		SetDefaults_BuildSource,
		SetDefaults_BuildStrategy,
		SetDefaults_SourceBuildStrategy,
		SetDefaults_DockerBuildStrategy,
		SetDefaults_CustomBuildStrategy,
		SetDefaults_BuildTriggerPolicy,
	)
}
