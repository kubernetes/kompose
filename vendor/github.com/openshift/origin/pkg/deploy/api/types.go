package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// These constants represent defaults used in the deployment process.
const (
	// DefaultRollingTimeoutSeconds is the default TimeoutSeconds for RollingDeploymentStrategyParams.
	DefaultRollingTimeoutSeconds int64 = 10 * 60
	// DefaultRollingIntervalSeconds is the default IntervalSeconds for RollingDeploymentStrategyParams.
	DefaultRollingIntervalSeconds int64 = 1
	// DefaultRollingUpdatePeriodSeconds is the default PeriodSeconds for RollingDeploymentStrategyParams.
	DefaultRollingUpdatePeriodSeconds int64 = 1
	// MaxDeploymentDurationSeconds represents the maximum duration that a deployment is allowed to run.
	// This is set as the default value for ActiveDeadlineSeconds for the deployer pod.
	// Currently set to 6 hours.
	MaxDeploymentDurationSeconds int64 = 21600
)

// These constants represent keys used for correlating objects related to deployments.
const (
	// DeploymentConfigAnnotation is an annotation name used to correlate a deployment with the
	// DeploymentConfig on which the deployment is based.
	DeploymentConfigAnnotation = "openshift.io/deployment-config.name"
	// DeploymentAnnotation is an annotation on a deployer Pod. The annotation value is the name
	// of the deployment (a ReplicationController) on which the deployer Pod acts.
	DeploymentAnnotation = "openshift.io/deployment.name"
	// DeploymentPodAnnotation is an annotation on a deployment (a ReplicationController). The
	// annotation value is the name of the deployer Pod which will act upon the ReplicationController
	// to implement the deployment behavior.
	DeploymentPodAnnotation = "openshift.io/deployer-pod.name"
	// DeploymentIgnorePodAnnotation is an annotation on a deployment config that will bypass creating
	// a deployment pod with the deployment. The caller is responsible for setting the deployment
	// status and running the deployment process.
	DeploymentIgnorePodAnnotation = "deploy.openshift.io/deployer-pod.ignore"
	// DeploymentPodTypeLabel is a label with which contains a type of deployment pod.
	DeploymentPodTypeLabel = "openshift.io/deployer-pod.type"
	// DeployerPodForDeploymentLabel is a label which groups pods related to a
	// deployment. The value is a deployment name. The deployer pod and hook pods
	// created by the internal strategies will have this label. Custom
	// strategies can apply this label to any pods they create, enabling
	// platform-provided cancellation and garbage collection support.
	DeployerPodForDeploymentLabel = "openshift.io/deployer-pod-for.name"
	// DeploymentStatusAnnotation is an annotation name used to retrieve the DeploymentPhase of
	// a deployment.
	DeploymentStatusAnnotation = "openshift.io/deployment.phase"
	// DeploymentEncodedConfigAnnotation is an annotation name used to retrieve specific encoded
	// DeploymentConfig on which a given deployment is based.
	DeploymentEncodedConfigAnnotation = "openshift.io/encoded-deployment-config"
	// DeploymentVersionAnnotation is an annotation on a deployment (a ReplicationController). The
	// annotation value is the LatestVersion value of the DeploymentConfig which was the basis for
	// the deployment.
	DeploymentVersionAnnotation = "openshift.io/deployment-config.latest-version"
	// DeploymentLabel is the name of a label used to correlate a deployment with the Pod created
	// to execute the deployment logic.
	// TODO: This is a workaround for upstream's lack of annotation support on PodTemplate. Once
	// annotations are available on PodTemplate, audit this constant with the goal of removing it.
	DeploymentLabel = "deployment"
	// DeploymentConfigLabel is the name of a label used to correlate a deployment with the
	// DeploymentConfigs on which the deployment is based.
	DeploymentConfigLabel = "deploymentconfig"
	// DesiredReplicasAnnotation represents the desired number of replicas for a
	// new deployment.
	// TODO: This should be made public upstream.
	DesiredReplicasAnnotation = "kubectl.kubernetes.io/desired-replicas"
	// DeploymentStatusReasonAnnotation represents the reason for deployment being in a given state
	// Used for specifying the reason for cancellation or failure of a deployment
	DeploymentStatusReasonAnnotation = "openshift.io/deployment.status-reason"
	// DeploymentCancelledAnnotation indicates that the deployment has been cancelled
	// The annotation value does not matter and its mere presence indicates cancellation
	DeploymentCancelledAnnotation = "openshift.io/deployment.cancelled"
	// DeploymentReplicasAnnotation is for internal use only and is for
	// detecting external modifications to deployment replica counts.
	DeploymentReplicasAnnotation = "openshift.io/deployment.replicas"
	// PostHookPodSuffix is the suffix added to all pre hook pods
	PreHookPodSuffix = "hook-pre"
	// PostHookPodSuffix is the suffix added to all mid hook pods
	MidHookPodSuffix = "hook-mid"
	// PostHookPodSuffix is the suffix added to all post hook pods
	PostHookPodSuffix = "hook-post"
)

// These constants represent values used in deployment annotations.
const (
	// DeploymentCancelledAnnotationValue represents the value for the DeploymentCancelledAnnotation
	// annotation that signifies that the deployment should be cancelled
	DeploymentCancelledAnnotationValue = "true"
)

// These constants represent the various reasons for cancelling a deployment
// or for a deployment being placed in a failed state
const (
	DeploymentCancelledByUser                 = "cancelled by the user"
	DeploymentCancelledNewerDeploymentExists  = "newer deployment was found running"
	DeploymentFailedUnrelatedDeploymentExists = "unrelated pod with the same name as this deployment is already running"
	DeploymentFailedDeployerPodNoLongerExists = "deployer pod no longer exists"
)

// DeploymentStatus describes the possible states a deployment can be in.
type DeploymentStatus string

const (
	// DeploymentStatusNew means the deployment has been accepted but not yet acted upon.
	DeploymentStatusNew DeploymentStatus = "New"
	// DeploymentStatusPending means the deployment been handed over to a deployment strategy,
	// but the strategy has not yet declared the deployment to be running.
	DeploymentStatusPending DeploymentStatus = "Pending"
	// DeploymentStatusRunning means the deployment strategy has reported the deployment as
	// being in-progress.
	DeploymentStatusRunning DeploymentStatus = "Running"
	// DeploymentStatusComplete means the deployment finished without an error.
	DeploymentStatusComplete DeploymentStatus = "Complete"
	// DeploymentStatusFailed means the deployment finished with an error.
	DeploymentStatusFailed DeploymentStatus = "Failed"
)

// +genclient=true

// DeploymentConfig represents a configuration for a single deployment (represented as a
// ReplicationController). It also contains details about changes which resulted in the current
// state of the DeploymentConfig. Each change to the DeploymentConfig which should result in
// a new deployment results in an increment of LatestVersion.
type DeploymentConfig struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Spec represents a desired deployment state and how to deploy to it.
	Spec DeploymentConfigSpec

	// Status represents the current deployment state.
	Status DeploymentConfigStatus
}

// DeploymentConfigSpec represents the desired state of the deployment.
type DeploymentConfigSpec struct {
	// Strategy describes how a deployment is executed.
	Strategy DeploymentStrategy

	// MinReadySeconds is the minimum number of seconds for which a newly created pod should
	// be ready without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	MinReadySeconds int32

	// Triggers determine how updates to a DeploymentConfig result in new deployments. If no triggers
	// are defined, a new deployment can only occur as a result of an explicit client update to the
	// DeploymentConfig with a new LatestVersion.
	Triggers []DeploymentTriggerPolicy

	// Replicas is the number of desired replicas.
	Replicas int32

	// RevisionHistoryLimit is the number of old ReplicationControllers to retain to allow for rollbacks.
	// This field is a pointer to allow for differentiation between an explicit zero and not specified.
	RevisionHistoryLimit *int32

	// Test ensures that this deployment config will have zero replicas except while a deployment is running. This allows the
	// deployment config to be used as a continuous deployment test - triggering on images, running the deployment, and then succeeding
	// or failing. Post strategy hooks and After actions can be used to integrate successful deployment with an action.
	Test bool

	// Paused indicates that the deployment config is paused resulting in no new deployments on template
	// changes or changes in the template caused by other triggers.
	Paused bool

	// Selector is a label query over pods that should match the Replicas count.
	Selector map[string]string

	// Template is the object that describes the pod that will be created if
	// insufficient replicas are detected.
	Template *kapi.PodTemplateSpec
}

// DeploymentStrategy describes how to perform a deployment.
type DeploymentStrategy struct {
	// Type is the name of a deployment strategy.
	Type DeploymentStrategyType

	// CustomParams are the input to the Custom deployment strategy, and may also
	// be specified for the Recreate and Rolling strategies to customize the execution
	// process that runs the deployment.
	CustomParams *CustomDeploymentStrategyParams
	// RecreateParams are the input to the Recreate deployment strategy.
	RecreateParams *RecreateDeploymentStrategyParams
	// RollingParams are the input to the Rolling deployment strategy.
	RollingParams *RollingDeploymentStrategyParams

	// Resources contains resource requirements to execute the deployment and any hooks.
	Resources kapi.ResourceRequirements
	// Labels is a set of key, value pairs added to custom deployer and lifecycle pre/post hook pods.
	Labels map[string]string
	// Annotations is a set of key, value pairs added to custom deployer and lifecycle pre/post hook pods.
	Annotations map[string]string
}

// DeploymentStrategyType refers to a specific DeploymentStrategy implementation.
type DeploymentStrategyType string

const (
	// DeploymentStrategyTypeRecreate is a simple strategy suitable as a default.
	DeploymentStrategyTypeRecreate DeploymentStrategyType = "Recreate"
	// DeploymentStrategyTypeCustom is a user defined strategy.
	DeploymentStrategyTypeCustom DeploymentStrategyType = "Custom"
	// DeploymentStrategyTypeRolling uses the Kubernetes RollingUpdater.
	DeploymentStrategyTypeRolling DeploymentStrategyType = "Rolling"
)

// CustomDeploymentStrategyParams are the input to the Custom deployment strategy.
type CustomDeploymentStrategyParams struct {
	// Image specifies a Docker image which can carry out a deployment.
	Image string
	// Environment holds the environment which will be given to the container for Image.
	Environment []kapi.EnvVar
	// Command is optional and overrides CMD in the container Image.
	Command []string
}

// RecreateDeploymentStrategyParams are the input to the Recreate deployment
// strategy.
type RecreateDeploymentStrategyParams struct {
	// TimeoutSeconds is the time to wait for updates before giving up. If the
	// value is nil, a default will be used.
	TimeoutSeconds *int64
	// Pre is a lifecycle hook which is executed before the strategy manipulates
	// the deployment. All LifecycleHookFailurePolicy values are supported.
	Pre *LifecycleHook
	// Mid is a lifecycle hook which is executed while the deployment is scaled down to zero before the first new
	// pod is created. All LifecycleHookFailurePolicy values are supported.
	Mid *LifecycleHook
	// Post is a lifecycle hook which is executed after the strategy has
	// finished all deployment logic. All LifecycleHookFailurePolicy values are supported.
	Post *LifecycleHook
}

// RollingDeploymentStrategyParams are the input to the Rolling deployment
// strategy.
type RollingDeploymentStrategyParams struct {
	// UpdatePeriodSeconds is the time to wait between individual pod updates.
	// If the value is nil, a default will be used.
	UpdatePeriodSeconds *int64
	// IntervalSeconds is the time to wait between polling deployment status
	// after update. If the value is nil, a default will be used.
	IntervalSeconds *int64
	// TimeoutSeconds is the time to wait for updates before giving up. If the
	// value is nil, a default will be used.
	TimeoutSeconds *int64
	// MaxUnavailable is the maximum number of pods that can be unavailable
	// during the update. Value can be an absolute number (ex: 5) or a
	// percentage of total pods at the start of update (ex: 10%). Absolute
	// number is calculated from percentage by rounding up.
	//
	// This cannot be 0 if MaxSurge is 0. By default, 25% is used.
	//
	// Example: when this is set to 30%, the old RC can be scaled down by 30%
	// immediately when the rolling update starts. Once new pods are ready, old
	// RC can be scaled down further, followed by scaling up the new RC,
	// ensuring that at least 70% of original number of pods are available at
	// all times during the update.
	MaxUnavailable intstr.IntOrString
	// MaxSurge is the maximum number of pods that can be scheduled above the
	// original number of pods. Value can be an absolute number (ex: 5) or a
	// percentage of total pods at the start of the update (ex: 10%). Absolute
	// number is calculated from percentage by rounding up.
	//
	// This cannot be 0 if MaxUnavailable is 0. By default, 25% is used.
	//
	// Example: when this is set to 30%, the new RC can be scaled up by 30%
	// immediately when the rolling update starts. Once old pods have been
	// killed, new RC can be scaled up further, ensuring that total number of
	// pods running at any time during the update is atmost 130% of original
	// pods.
	MaxSurge intstr.IntOrString
	// Pre is a lifecycle hook which is executed before the deployment process
	// begins. All LifecycleHookFailurePolicy values are supported.
	Pre *LifecycleHook
	// Post is a lifecycle hook which is executed after the strategy has
	// finished all deployment logic. All LifecycleHookFailurePolicy values
	// are supported.
	Post *LifecycleHook
}

// LifecycleHook defines a specific deployment lifecycle action. Only one type of action may be specified at any time.
type LifecycleHook struct {
	// FailurePolicy specifies what action to take if the hook fails.
	FailurePolicy LifecycleHookFailurePolicy

	// ExecNewPod specifies the options for a lifecycle hook backed by a pod.
	ExecNewPod *ExecNewPodHook

	// TagImages instructs the deployer to tag the current image referenced under a container onto an image stream tag.
	TagImages []TagImageHook
}

// LifecycleHookFailurePolicy describes possibles actions to take if a hook fails.
type LifecycleHookFailurePolicy string

const (
	// LifecycleHookFailurePolicyRetry means retry the hook until it succeeds.
	LifecycleHookFailurePolicyRetry LifecycleHookFailurePolicy = "Retry"
	// LifecycleHookFailurePolicyAbort means abort the deployment.
	LifecycleHookFailurePolicyAbort LifecycleHookFailurePolicy = "Abort"
	// LifecycleHookFailurePolicyIgnore means ignore failure and continue the deployment.
	LifecycleHookFailurePolicyIgnore LifecycleHookFailurePolicy = "Ignore"
)

// ExecNewPodHook is a hook implementation which runs a command in a new pod
// based on the specified container which is assumed to be part of the
// deployment template.
type ExecNewPodHook struct {
	// Command is the action command and its arguments.
	Command []string
	// Env is a set of environment variables to supply to the hook pod's container.
	Env []kapi.EnvVar
	// ContainerName is the name of a container in the deployment pod template
	// whose Docker image will be used for the hook pod's container.
	ContainerName string
	// Volumes is a list of named volumes from the pod template which should be
	// copied to the hook pod. Volumes names not found in pod spec are ignored.
	// An empty list means no volumes will be copied.
	Volumes []string
}

// TagImageHook is a request to tag the image in a particular container onto an ImageStreamTag.
type TagImageHook struct {
	// ContainerName is the name of a container in the deployment config whose image value will be used as the source of the tag. If there is only a single
	// container this value will be defaulted to the name of that container.
	ContainerName string
	// To is the target ImageStreamTag to set the container's image onto.
	To kapi.ObjectReference
}

// DeploymentTriggerPolicy describes a policy for a single trigger that results in a new deployment.
type DeploymentTriggerPolicy struct {
	// Type of the trigger
	Type DeploymentTriggerType
	// ImageChangeParams represents the parameters for the ImageChange trigger.
	ImageChangeParams *DeploymentTriggerImageChangeParams
}

// DeploymentTriggerType refers to a specific DeploymentTriggerPolicy implementation.
type DeploymentTriggerType string

const (
	// DeploymentTriggerManual is a placeholder implementation which does nothing.
	DeploymentTriggerManual DeploymentTriggerType = "Manual"
	// DeploymentTriggerOnImageChange will create new deployments in response to updated tags from
	// a Docker image repository.
	DeploymentTriggerOnImageChange DeploymentTriggerType = "ImageChange"
	// DeploymentTriggerOnConfigChange will create new deployments in response to changes to
	// the ControllerTemplate of a DeploymentConfig.
	DeploymentTriggerOnConfigChange DeploymentTriggerType = "ConfigChange"
)

// DeploymentTriggerImageChangeParams represents the parameters to the ImageChange trigger.
type DeploymentTriggerImageChangeParams struct {
	// Automatic means that the detection of a new tag value should result in an image update
	// inside the pod template.
	Automatic bool
	// ContainerNames is used to restrict tag updates to the specified set of container names in a pod.
	ContainerNames []string
	// From is a reference to an image stream tag to watch for changes. From.Name is the only
	// required subfield - if From.Namespace is blank, the namespace of the current deployment
	// trigger will be used.
	From kapi.ObjectReference
	// LastTriggeredImage is the last image to be triggered.
	LastTriggeredImage string
}

// DeploymentConfigStatus represents the current deployment state.
type DeploymentConfigStatus struct {
	// LatestVersion is used to determine whether the current deployment associated with a deployment
	// config is out of sync.
	LatestVersion int64
	// ObservedGeneration is the most recent generation observed by the deployment config controller.
	ObservedGeneration int64
	// Replicas is the total number of pods targeted by this deployment config.
	Replicas int32
	// UpdatedReplicas is the total number of non-terminated pods targeted by this deployment config
	// that have the desired template spec.
	UpdatedReplicas int32
	// AvailableReplicas is the total number of available pods targeted by this deployment config.
	AvailableReplicas int32
	// UnavailableReplicas is the total number of unavailable pods targeted by this deployment config.
	UnavailableReplicas int32
	// Details are the reasons for the update to this deployment config.
	// This could be based on a change made by the user or caused by an automatic trigger
	Details *DeploymentDetails
	// Conditions represents the latest available observations of a deployment config's current state.
	Conditions []DeploymentCondition
}

// DeploymentDetails captures information about the causes of a deployment.
type DeploymentDetails struct {
	// Message is the user specified change message, if this deployment was triggered manually by the user
	Message string
	// Causes are extended data associated with all the causes for creating a new deployment
	Causes []DeploymentCause
}

// DeploymentCause captures information about a particular cause of a deployment.
type DeploymentCause struct {
	// Type is the type of the trigger that resulted in the creation of a new deployment
	Type DeploymentTriggerType
	// ImageTrigger contains the image trigger details, if this trigger was fired based on an image change
	ImageTrigger *DeploymentCauseImageTrigger
}

// DeploymentCauseImageTrigger contains information about a deployment caused by an image trigger
type DeploymentCauseImageTrigger struct {
	// From is a reference to the changed object which triggered a deployment. The field may have
	// the kinds DockerImage, ImageStreamTag, or ImageStreamImage.
	From kapi.ObjectReference
}

type DeploymentConditionType string

// These are valid conditions of a deployment config.
const (
	// DeploymentAvailable means the deployment config is available, ie. at least the minimum available
	// replicas required are up and running for at least minReadySeconds.
	DeploymentAvailable DeploymentConditionType = "Available"
	// DeploymentProgressing means the deployment config is progressing. Progress for a deployment
	// config is considered when a new replica set is created or adopted, and when new pods scale up or
	// old pods scale down. Progress is not estimated for paused deployment configs, when the deployment
	// config needs to rollback, or when progressDeadlineSeconds is not specified.
	DeploymentProgressing DeploymentConditionType = "Progressing"
	// DeploymentReplicaFailure is added in a deployment config when one of its pods
	// fails to be created or deleted.
	DeploymentReplicaFailure DeploymentConditionType = "ReplicaFailure"
)

// DeploymentCondition describes the state of a deployment config at a certain point.
type DeploymentCondition struct {
	// Type of deployment condition.
	Type DeploymentConditionType
	// Status of the condition, one of True, False, Unknown.
	Status kapi.ConditionStatus
	// The last time the condition transitioned from one status to another.
	LastTransitionTime unversioned.Time
	// The reason for the condition's last transition.
	Reason string
	// A human readable message indicating details about the transition.
	Message string
}

// DeploymentConfigList is a collection of deployment configs.
type DeploymentConfigList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	// Items is a list of deployment configs
	Items []DeploymentConfig
}

// DeploymentConfigRollback provides the input to rollback generation.
type DeploymentConfigRollback struct {
	unversioned.TypeMeta
	// Name of the deployment config that will be rolled back.
	Name string
	// UpdatedAnnotations is a set of new annotations that will be added in the deployment config.
	UpdatedAnnotations map[string]string
	// Spec defines the options to rollback generation.
	Spec DeploymentConfigRollbackSpec
}

// DeploymentConfigRollbackSpec represents the options for rollback generation.
type DeploymentConfigRollbackSpec struct {
	// From points to a ReplicationController which is a deployment.
	From kapi.ObjectReference
	// Revision to rollback to. If set to 0, rollback to the last revision.
	Revision int64
	// IncludeTriggers specifies whether to include config Triggers.
	IncludeTriggers bool
	// IncludeTemplate specifies whether to include the PodTemplateSpec.
	IncludeTemplate bool
	// IncludeReplicationMeta specifies whether to include the replica count and selector.
	IncludeReplicationMeta bool
	// IncludeStrategy specifies whether to include the deployment Strategy.
	IncludeStrategy bool
}

// DeploymentRequest is a request to a deployment config for a new deployment.
type DeploymentRequest struct {
	unversioned.TypeMeta
	// Name of the deployment config for requesting a new deployment.
	Name string
	// Latest will update the deployment config with the latest state from all triggers.
	Latest bool
	// Force will try to force a new deployment to run. If the deployment config is paused,
	// then setting this to true will return an Invalid error.
	Force bool
}

// DeploymentLog represents the logs for a deployment
type DeploymentLog struct {
	unversioned.TypeMeta
}

// DeploymentLogOptions is the REST options for a deployment log
type DeploymentLogOptions struct {
	unversioned.TypeMeta

	// Container for which to return logs
	Container string
	// Follow if true indicates that the deployment log should be streamed until
	// the deployment terminates.
	Follow bool
	// If true, return previous deployment logs
	Previous bool
	// A relative time in seconds before the current time from which to show logs. If this value
	// precedes the time a pod was started, only logs since the pod start will be returned.
	// If this value is in the future, no logs will be returned.
	// Only one of sinceSeconds or sinceTime may be specified.
	SinceSeconds *int64
	// An RFC3339 timestamp from which to show logs. If this value
	// precedes the time a pod was started, only logs since the pod start will be returned.
	// If this value is in the future, no logs will be returned.
	// Only one of sinceSeconds or sinceTime may be specified.
	SinceTime *unversioned.Time
	// If true, add an RFC3339 or RFC3339Nano timestamp at the beginning of every line
	// of log output.
	Timestamps bool
	// If set, the number of lines from the end of the logs to show. If not specified,
	// logs are shown from the creation of the container or sinceSeconds or sinceTime
	TailLines *int64
	// If set, the number of bytes to read from the server before terminating the
	// log output. This may not display a complete final line of logging, and may return
	// slightly more or slightly less than the specified limit.
	LimitBytes *int64

	// NoWait if true causes the call to return immediately even if the deployment
	// is not available yet. Otherwise the server will wait until the deployment has started.
	NoWait bool

	// Version of the deployment for which to view logs.
	Version *int64
}
