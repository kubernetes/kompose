package util

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kdeplutil "k8s.io/kubernetes/pkg/controller/deployment/util"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/openshift/origin/pkg/util/namer"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	// Reasons for deployment config conditions:
	//
	// ReplicationControllerUpdatedReason is added in a deployment config when one of its replication
	// controllers is updated as part of the rollout process.
	ReplicationControllerUpdatedReason = "ReplicationControllerUpdated"
	// FailedRcCreateReason is added in a deployment config when it cannot create a new replication
	// controller.
	FailedRcCreateReason = "ReplicationControllerCreateError"
	// NewReplicationControllerReason is added in a deployment config when it creates a new replication
	// controller.
	NewReplicationControllerReason = "NewReplicationControllerCreated"
	// NewRcAvailableReason is added in a deployment config when its newest replication controller is made
	// available ie. the number of new pods that have passed readiness checks and run for at least
	// minReadySeconds is at least the minimum available pods that need to run for the deployment config.
	NewRcAvailableReason = "NewReplicationControllerAvailable"
	// TimedOutReason is added in a deployment config when its newest replication controller fails to show
	// any progress within the given deadline (progressDeadlineSeconds).
	TimedOutReason = "ProgressDeadlineExceeded"
	// PausedDeployReason is added in a deployment config when it is paused. Lack of progress shouldn't be
	// estimated once a deployment config is paused.
	PausedDeployReason = "DeploymentConfigPaused"
	// ResumedDeployReason is added in a deployment config when it is resumed. Useful for not failing accidentally
	// deployment configs that paused amidst a rollout.
	ResumedDeployReason = "DeploymentConfigResumed"
)

// NewDeploymentCondition creates a new deployment condition.
func NewDeploymentCondition(condType deployapi.DeploymentConditionType, status api.ConditionStatus, reason, message string) *deployapi.DeploymentCondition {
	return &deployapi.DeploymentCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: unversioned.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetDeploymentCondition returns the condition with the provided type.
func GetDeploymentCondition(status deployapi.DeploymentConfigStatus, condType deployapi.DeploymentConditionType) *deployapi.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetDeploymentCondition updates the deployment to include the provided condition. If the condition that
// we are about to add already exists and has the same status and reason then we are not going to update.
func SetDeploymentCondition(status *deployapi.DeploymentConfigStatus, condition deployapi.DeploymentCondition) {
	currentCond := GetDeploymentCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}
	// Preserve lastTransitionTime if we are not switching between statuses of a condition.
	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

// RemoveDeploymentCondition removes the deployment condition with the provided type.
func RemoveDeploymentCondition(status *deployapi.DeploymentConfigStatus, condType deployapi.DeploymentConditionType) {
	status.Conditions = filterOutCondition(status.Conditions, condType)
}

// filterOutCondition returns a new slice of deployment conditions without conditions with the provided type.
func filterOutCondition(conditions []deployapi.DeploymentCondition, condType deployapi.DeploymentConditionType) []deployapi.DeploymentCondition {
	var newConditions []deployapi.DeploymentCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

// LatestDeploymentNameForConfig returns a stable identifier for config based on its version.
func LatestDeploymentNameForConfig(config *deployapi.DeploymentConfig) string {
	return fmt.Sprintf("%s-%d", config.Name, config.Status.LatestVersion)
}

// LatestDeploymentInfo returns info about the latest deployment for a config,
// or nil if there is no latest deployment. The latest deployment is not
// always the same as the active deployment.
func LatestDeploymentInfo(config *deployapi.DeploymentConfig, deployments []api.ReplicationController) (bool, *api.ReplicationController) {
	if config.Status.LatestVersion == 0 || len(deployments) == 0 {
		return false, nil
	}
	sort.Sort(ByLatestVersionDesc(deployments))
	candidate := &deployments[0]
	return DeploymentVersionFor(candidate) == config.Status.LatestVersion, candidate
}

// ActiveDeployment returns the latest complete deployment, or nil if there is
// no such deployment. The active deployment is not always the same as the
// latest deployment.
func ActiveDeployment(input []api.ReplicationController) *api.ReplicationController {
	var activeDeployment *api.ReplicationController
	var lastCompleteDeploymentVersion int64 = 0
	for i := range input {
		deployment := &input[i]
		deploymentVersion := DeploymentVersionFor(deployment)
		if IsCompleteDeployment(deployment) && deploymentVersion > lastCompleteDeploymentVersion {
			activeDeployment = deployment
			lastCompleteDeploymentVersion = deploymentVersion
		}
	}
	return activeDeployment
}

// DeployerPodSuffix is the suffix added to pods created from a deployment
const DeployerPodSuffix = "deploy"

// DeployerPodNameForDeployment returns the name of a pod for a given deployment
func DeployerPodNameForDeployment(deployment string) string {
	return namer.GetPodName(deployment, DeployerPodSuffix)
}

// LabelForDeployment builds a string identifier for a Deployment.
func LabelForDeployment(deployment *api.ReplicationController) string {
	return fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
}

// LabelForDeploymentConfig builds a string identifier for a DeploymentConfig.
func LabelForDeploymentConfig(config *deployapi.DeploymentConfig) string {
	return fmt.Sprintf("%s/%s", config.Namespace, config.Name)
}

// DeploymentNameForConfigVersion returns the name of the version-th deployment
// for the config that has the provided name
func DeploymentNameForConfigVersion(name string, version int64) string {
	return fmt.Sprintf("%s-%d", name, version)
}

// ConfigSelector returns a label Selector which can be used to find all
// deployments for a DeploymentConfig.
//
// TODO: Using the annotation constant for now since the value is correct
// but we could consider adding a new constant to the public types.
func ConfigSelector(name string) labels.Selector {
	return labels.Set{deployapi.DeploymentConfigAnnotation: name}.AsSelector()
}

// DeployerPodSelector returns a label Selector which can be used to find all
// deployer pods associated with a deployment with name.
func DeployerPodSelector(name string) labels.Selector {
	return labels.Set{deployapi.DeployerPodForDeploymentLabel: name}.AsSelector()
}

// AnyDeployerPodSelector returns a label Selector which can be used to find
// all deployer pods across all deployments, including hook and custom
// deployer pods.
func AnyDeployerPodSelector() labels.Selector {
	sel, _ := labels.Parse(deployapi.DeployerPodForDeploymentLabel)
	return sel
}

// HasChangeTrigger returns whether the provided deployment configuration has
// a config change trigger or not
func HasChangeTrigger(config *deployapi.DeploymentConfig) bool {
	for _, trigger := range config.Spec.Triggers {
		if trigger.Type == deployapi.DeploymentTriggerOnConfigChange {
			return true
		}
	}
	return false
}

// HasImageChangeTrigger returns whether the provided deployment configuration has
// an image change trigger or not.
func HasImageChangeTrigger(config *deployapi.DeploymentConfig) bool {
	for _, trigger := range config.Spec.Triggers {
		if trigger.Type == deployapi.DeploymentTriggerOnImageChange {
			return true
		}
	}
	return false
}

func DeploymentConfigDeepCopy(dc *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error) {
	objCopy, err := api.Scheme.DeepCopy(dc)
	if err != nil {
		return nil, err
	}
	copied, ok := objCopy.(*deployapi.DeploymentConfig)
	if !ok {
		return nil, fmt.Errorf("expected DeploymentConfig, got %#v", objCopy)
	}
	return copied, nil
}

func DeploymentDeepCopy(rc *api.ReplicationController) (*api.ReplicationController, error) {
	objCopy, err := api.Scheme.DeepCopy(rc)
	if err != nil {
		return nil, err
	}
	copied, ok := objCopy.(*api.ReplicationController)
	if !ok {
		return nil, fmt.Errorf("expected ReplicationController, got %#v", objCopy)
	}
	return copied, nil
}

// DecodeDeploymentConfig decodes a DeploymentConfig from controller using codec. An error is returned
// if the controller doesn't contain an encoded config.
func DecodeDeploymentConfig(controller *api.ReplicationController, decoder runtime.Decoder) (*deployapi.DeploymentConfig, error) {
	encodedConfig := []byte(EncodedDeploymentConfigFor(controller))
	decoded, err := runtime.Decode(decoder, encodedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode DeploymentConfig from controller: %v", err)
	}
	config, ok := decoded.(*deployapi.DeploymentConfig)
	if !ok {
		return nil, fmt.Errorf("decoded object from controller is not a DeploymentConfig")
	}
	return config, nil
}

// EncodeDeploymentConfig encodes config as a string using codec.
func EncodeDeploymentConfig(config *deployapi.DeploymentConfig, codec runtime.Codec) (string, error) {
	bytes, err := runtime.Encode(codec, config)
	if err != nil {
		return "", err
	}
	return string(bytes[:]), nil
}

// MakeDeployment creates a deployment represented as a ReplicationController and based on the given
// DeploymentConfig. The controller replica count will be zero.
func MakeDeployment(config *deployapi.DeploymentConfig, codec runtime.Codec) (*api.ReplicationController, error) {
	var err error
	var encodedConfig string

	if encodedConfig, err = EncodeDeploymentConfig(config, codec); err != nil {
		return nil, err
	}

	deploymentName := LatestDeploymentNameForConfig(config)

	podSpec := api.PodSpec{}
	if err := api.Scheme.Convert(&config.Spec.Template.Spec, &podSpec, nil); err != nil {
		return nil, fmt.Errorf("couldn't clone podSpec: %v", err)
	}

	controllerLabels := make(labels.Set)
	for k, v := range config.Labels {
		controllerLabels[k] = v
	}
	// Correlate the deployment with the config.
	// TODO: Using the annotation constant for now since the value is correct
	// but we could consider adding a new constant to the public types.
	controllerLabels[deployapi.DeploymentConfigAnnotation] = config.Name

	// Ensure that pods created by this deployment controller can be safely associated back
	// to the controller, and that multiple deployment controllers for the same config don't
	// manipulate each others' pods.
	selector := map[string]string{}
	for k, v := range config.Spec.Selector {
		selector[k] = v
	}
	selector[deployapi.DeploymentConfigLabel] = config.Name
	selector[deployapi.DeploymentLabel] = deploymentName

	podLabels := make(labels.Set)
	for k, v := range config.Spec.Template.Labels {
		podLabels[k] = v
	}
	podLabels[deployapi.DeploymentConfigLabel] = config.Name
	podLabels[deployapi.DeploymentLabel] = deploymentName

	podAnnotations := make(labels.Set)
	for k, v := range config.Spec.Template.Annotations {
		podAnnotations[k] = v
	}
	podAnnotations[deployapi.DeploymentAnnotation] = deploymentName
	podAnnotations[deployapi.DeploymentConfigAnnotation] = config.Name
	podAnnotations[deployapi.DeploymentVersionAnnotation] = strconv.FormatInt(config.Status.LatestVersion, 10)

	deployment := &api.ReplicationController{
		ObjectMeta: api.ObjectMeta{
			Name:      deploymentName,
			Namespace: config.Namespace,
			Annotations: map[string]string{
				deployapi.DeploymentConfigAnnotation:        config.Name,
				deployapi.DeploymentStatusAnnotation:        string(deployapi.DeploymentStatusNew),
				deployapi.DeploymentEncodedConfigAnnotation: encodedConfig,
				deployapi.DeploymentVersionAnnotation:       strconv.FormatInt(config.Status.LatestVersion, 10),
				// This is the target replica count for the new deployment.
				deployapi.DesiredReplicasAnnotation:    strconv.Itoa(int(config.Spec.Replicas)),
				deployapi.DeploymentReplicasAnnotation: strconv.Itoa(0),
			},
			Labels: controllerLabels,
		},
		Spec: api.ReplicationControllerSpec{
			// The deployment should be inactive initially
			Replicas: 0,
			Selector: selector,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels:      podLabels,
					Annotations: podAnnotations,
				},
				Spec: podSpec,
			},
		},
	}
	if config.Status.Details != nil && len(config.Status.Details.Message) > 0 {
		deployment.Annotations[deployapi.DeploymentStatusReasonAnnotation] = config.Status.Details.Message
	}
	if value, ok := config.Annotations[deployapi.DeploymentIgnorePodAnnotation]; ok {
		deployment.Annotations[deployapi.DeploymentIgnorePodAnnotation] = value
	}

	return deployment, nil
}

// GetReplicaCountForDeployments returns the sum of all replicas for the
// given deployments.
func GetReplicaCountForDeployments(deployments []api.ReplicationController) int32 {
	totalReplicaCount := int32(0)
	for _, deployment := range deployments {
		totalReplicaCount += deployment.Spec.Replicas
	}
	return totalReplicaCount
}

// GetStatusReplicaCountForDeployments returns the sum of the replicas reported in the
// status of the given deployments.
func GetStatusReplicaCountForDeployments(deployments []api.ReplicationController) int32 {
	totalReplicaCount := int32(0)
	for _, deployment := range deployments {
		totalReplicaCount += deployment.Status.Replicas
	}
	return totalReplicaCount
}

// GetAvailablePods returns all the available pods from the provided pod list.
func GetAvailablePods(pods []*api.Pod, minReadySeconds int32) int32 {
	available := int32(0)
	for i := range pods {
		pod := pods[i]
		if kdeplutil.IsPodAvailable(pod, minReadySeconds, time.Now()) {
			available++
		}
	}
	return available
}

func DeploymentConfigNameFor(obj runtime.Object) string {
	return annotationFor(obj, deployapi.DeploymentConfigAnnotation)
}

func DeploymentNameFor(obj runtime.Object) string {
	return annotationFor(obj, deployapi.DeploymentAnnotation)
}

func DeployerPodNameFor(obj runtime.Object) string {
	return annotationFor(obj, deployapi.DeploymentPodAnnotation)
}

func DeploymentStatusFor(obj runtime.Object) deployapi.DeploymentStatus {
	return deployapi.DeploymentStatus(annotationFor(obj, deployapi.DeploymentStatusAnnotation))
}

func DeploymentStatusReasonFor(obj runtime.Object) string {
	return annotationFor(obj, deployapi.DeploymentStatusReasonAnnotation)
}

func DeploymentDesiredReplicas(obj runtime.Object) (int32, bool) {
	return int32AnnotationFor(obj, deployapi.DesiredReplicasAnnotation)
}

func DeploymentReplicas(obj runtime.Object) (int32, bool) {
	return int32AnnotationFor(obj, deployapi.DeploymentReplicasAnnotation)
}

func EncodedDeploymentConfigFor(obj runtime.Object) string {
	return annotationFor(obj, deployapi.DeploymentEncodedConfigAnnotation)
}

func DeploymentVersionFor(obj runtime.Object) int64 {
	v, err := strconv.ParseInt(annotationFor(obj, deployapi.DeploymentVersionAnnotation), 10, 64)
	if err != nil {
		return -1
	}
	return v
}

func IsDeploymentCancelled(deployment *api.ReplicationController) bool {
	value := annotationFor(deployment, deployapi.DeploymentCancelledAnnotation)
	return strings.EqualFold(value, deployapi.DeploymentCancelledAnnotationValue)
}

// HasSynced checks if the provided deployment config has been noticed by the deployment
// config controller.
func HasSynced(dc *deployapi.DeploymentConfig, generation int64) bool {
	return dc.Status.ObservedGeneration >= generation
}

// IsOwnedByConfig checks whether the provided replication controller is part of a
// deployment configuration.
// TODO: Switch to use owner references once we got those working.
func IsOwnedByConfig(deployment *api.ReplicationController) bool {
	_, ok := deployment.Annotations[deployapi.DeploymentConfigAnnotation]
	return ok
}

// IsTerminatedDeployment returns true if the passed deployment has terminated (either
// complete or failed).
func IsTerminatedDeployment(deployment *api.ReplicationController) bool {
	return IsCompleteDeployment(deployment) || IsFailedDeployment(deployment)
}

// IsCompleteDeployment returns true if the passed deployment failed.
func IsCompleteDeployment(deployment *api.ReplicationController) bool {
	current := DeploymentStatusFor(deployment)
	return current == deployapi.DeploymentStatusComplete
}

// IsFailedDeployment returns true if the passed deployment failed.
func IsFailedDeployment(deployment *api.ReplicationController) bool {
	current := DeploymentStatusFor(deployment)
	return current == deployapi.DeploymentStatusFailed
}

// CanTransitionPhase returns whether it is allowed to go from the current to the next phase.
func CanTransitionPhase(current, next deployapi.DeploymentStatus) bool {
	switch current {
	case deployapi.DeploymentStatusNew:
		switch next {
		case deployapi.DeploymentStatusPending,
			deployapi.DeploymentStatusRunning,
			deployapi.DeploymentStatusFailed,
			deployapi.DeploymentStatusComplete:
			return true
		}
	case deployapi.DeploymentStatusPending:
		switch next {
		case deployapi.DeploymentStatusRunning,
			deployapi.DeploymentStatusFailed,
			deployapi.DeploymentStatusComplete:
			return true
		}
	case deployapi.DeploymentStatusRunning:
		switch next {
		case deployapi.DeploymentStatusFailed, deployapi.DeploymentStatusComplete:
			return true
		}
	}
	return false
}

// IsRollingConfig returns true if the strategy type is a rolling update.
func IsRollingConfig(config *deployapi.DeploymentConfig) bool {
	return config.Spec.Strategy.Type == deployapi.DeploymentStrategyTypeRolling
}

// IsProgressing expects a state deployment config and its updated status in order to
// determine if there is any progress.
func IsProgressing(config deployapi.DeploymentConfig, newStatus deployapi.DeploymentConfigStatus) bool {
	oldStatusOldReplicas := config.Status.Replicas - config.Status.UpdatedReplicas
	newStatusOldReplicas := newStatus.Replicas - newStatus.UpdatedReplicas

	return (newStatus.UpdatedReplicas > config.Status.UpdatedReplicas) || (newStatusOldReplicas < oldStatusOldReplicas)
}

// MaxUnavailable returns the maximum unavailable pods a rolling deployment config can take.
func MaxUnavailable(config deployapi.DeploymentConfig) int32 {
	if !IsRollingConfig(&config) {
		return int32(0)
	}
	// Error caught by validation
	_, maxUnavailable, _ := kdeplutil.ResolveFenceposts(&config.Spec.Strategy.RollingParams.MaxSurge, &config.Spec.Strategy.RollingParams.MaxUnavailable, config.Spec.Replicas)
	return maxUnavailable
}

// MaxSurge returns the maximum surge pods a rolling deployment config can take.
func MaxSurge(config deployapi.DeploymentConfig) int32 {
	if !IsRollingConfig(&config) {
		return int32(0)
	}
	// Error caught by validation
	maxSurge, _, _ := kdeplutil.ResolveFenceposts(&config.Spec.Strategy.RollingParams.MaxSurge, &config.Spec.Strategy.RollingParams.MaxUnavailable, config.Spec.Replicas)
	return maxSurge
}

// annotationFor returns the annotation with key for obj.
func annotationFor(obj runtime.Object, key string) string {
	meta, err := api.ObjectMetaFor(obj)
	if err != nil {
		return ""
	}
	return meta.Annotations[key]
}

func int32AnnotationFor(obj runtime.Object, key string) (int32, bool) {
	s := annotationFor(obj, key)
	if len(s) == 0 {
		return 0, false
	}
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(i), true
}

// DeploymentsForCleanup determines which deployments for a configuration are relevant for the
// revision history limit quota
func DeploymentsForCleanup(configuration *deployapi.DeploymentConfig, deployments []api.ReplicationController) []api.ReplicationController {
	// if the past deployment quota has been exceeded, we need to prune the oldest deployments
	// until we are not exceeding the quota any longer, so we sort oldest first
	sort.Sort(ByLatestVersionAsc(deployments))

	relevantDeployments := []api.ReplicationController{}
	activeDeployment := ActiveDeployment(deployments)
	if activeDeployment == nil {
		// if cleanup policy is set but no successful deployments have happened, there will be
		// no active deployment. We can consider all of the deployments in this case except for
		// the latest one
		for i := range deployments {
			deployment := &deployments[i]
			if DeploymentVersionFor(deployment) != configuration.Status.LatestVersion {
				relevantDeployments = append(relevantDeployments, *deployment)
			}
		}
	} else {
		// if there is an active deployment, we need to filter out any deployments that we don't
		// care about, namely the active deployment and any newer deployments
		for i := range deployments {
			deployment := &deployments[i]
			if deployment != activeDeployment && DeploymentVersionFor(deployment) < DeploymentVersionFor(activeDeployment) {
				relevantDeployments = append(relevantDeployments, *deployment)
			}
		}
	}

	return relevantDeployments
}

// WaitForRunningDeployerPod waits a given period of time until the deployer pod
// for given replication controller is not running.
func WaitForRunningDeployerPod(podClient kclient.PodsNamespacer, rc *api.ReplicationController, timeout time.Duration) error {
	podName := DeployerPodNameForDeployment(rc.Name)
	canGetLogs := func(p *api.Pod) bool {
		return api.PodSucceeded == p.Status.Phase || api.PodFailed == p.Status.Phase || api.PodRunning == p.Status.Phase
	}
	pod, err := podClient.Pods(rc.Namespace).Get(podName)
	if err == nil && canGetLogs(pod) {
		return nil
	}
	watcher, err := podClient.Pods(rc.Namespace).Watch(
		api.ListOptions{
			FieldSelector: fields.Set{"metadata.name": podName}.AsSelector(),
		},
	)
	if err != nil {
		return err
	}

	defer watcher.Stop()
	if _, err := watch.Until(timeout, watcher, func(e watch.Event) (bool, error) {
		if e.Type == watch.Error {
			return false, fmt.Errorf("encountered error while watching for pod: %v", e.Object)
		}
		obj, isPod := e.Object.(*api.Pod)
		if !isPod {
			return false, errors.New("received unknown object while watching for pods")
		}
		return canGetLogs(obj), nil
	}); err != nil {
		return err
	}
	return nil
}

// ByLatestVersionAsc sorts deployments by LatestVersion ascending.
type ByLatestVersionAsc []api.ReplicationController

func (d ByLatestVersionAsc) Len() int      { return len(d) }
func (d ByLatestVersionAsc) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d ByLatestVersionAsc) Less(i, j int) bool {
	return DeploymentVersionFor(&d[i]) < DeploymentVersionFor(&d[j])
}

// ByLatestVersionDesc sorts deployments by LatestVersion descending.
type ByLatestVersionDesc []api.ReplicationController

func (d ByLatestVersionDesc) Len() int      { return len(d) }
func (d ByLatestVersionDesc) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d ByLatestVersionDesc) Less(i, j int) bool {
	return DeploymentVersionFor(&d[j]) < DeploymentVersionFor(&d[i])
}

// ByMostRecent sorts deployments by most recently created.
type ByMostRecent []*api.ReplicationController

func (s ByMostRecent) Len() int      { return len(s) }
func (s ByMostRecent) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByMostRecent) Less(i, j int) bool {
	return !s[i].CreationTimestamp.Before(s[j].CreationTimestamp)
}
