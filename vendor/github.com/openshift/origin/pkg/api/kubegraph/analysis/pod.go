package analysis

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
)

const (
	CrashLoopingPodError = "CrashLoopingPod"
	RestartingPodWarning = "RestartingPod"

	RestartThreshold = 5
	// TODO: if you change this, you must change the messages below.
	RestartRecentDuration = 10 * time.Minute
)

// exposed for testing
var nowFn = unversioned.Now

// FindRestartingPods inspects all Pods to see if they've restarted more than the threshold. logsCommandName is the name of
// the command that should be invoked to see pod logs. securityPolicyCommandPattern is a format string accepting two replacement
// variables for fmt.Sprintf - 1, the namespace of the current pod, 2 the service account of the pod.
func FindRestartingPods(g osgraph.Graph, f osgraph.Namer, logsCommandName, securityPolicyCommandPattern string) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastPodNode := range g.NodesByKind(kubegraph.PodNodeKind) {
		podNode := uncastPodNode.(*kubegraph.PodNode)
		pod, ok := podNode.Object().(*kapi.Pod)
		if !ok {
			continue
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerString := ""
			if len(pod.Spec.Containers) > 1 {
				containerString = fmt.Sprintf("container %q in ", containerStatus.Name)
			}
			switch {
			case containerCrashLoopBackOff(containerStatus):
				var suggestion string
				switch {
				case containerIsNonRoot(pod, containerStatus.Name):
					suggestion = heredoc.Docf(`
						The container is starting and exiting repeatedly. This usually means the container is unable
						to start, misconfigured, or limited by security restrictions. Check the container logs with

						  %s %s -c %s

						Current security policy prevents your containers from being run as the root user. Some images
						may fail expecting to be able to change ownership or permissions on directories. Your admin
						can grant you access to run containers that need to run as the root user with this command:

						  %s
						`, logsCommandName, pod.Name, containerStatus.Name, fmt.Sprintf(securityPolicyCommandPattern, pod.Namespace, pod.Spec.ServiceAccountName))
				default:
					suggestion = heredoc.Docf(`
						The container is starting and exiting repeatedly. This usually means the container is unable
						to start, misconfigured, or limited by security restrictions. Check the container logs with

						  %s %s -c %s
						`, logsCommandName, pod.Name, containerStatus.Name)
				}
				markers = append(markers, osgraph.Marker{
					Node: podNode,

					Severity: osgraph.ErrorSeverity,
					Key:      CrashLoopingPodError,
					Message: fmt.Sprintf("%s%s is crash-looping", containerString,
						f.ResourceName(podNode)),
					Suggestion: osgraph.Suggestion(suggestion),
				})
			case ContainerRestartedRecently(containerStatus, nowFn()):
				markers = append(markers, osgraph.Marker{
					Node: podNode,

					Severity: osgraph.WarningSeverity,
					Key:      RestartingPodWarning,
					Message: fmt.Sprintf("%s%s has restarted within the last 10 minutes", containerString,
						f.ResourceName(podNode)),
				})
			case containerRestartedFrequently(containerStatus):
				markers = append(markers, osgraph.Marker{
					Node: podNode,

					Severity: osgraph.WarningSeverity,
					Key:      RestartingPodWarning,
					Message: fmt.Sprintf("%s%s has restarted %d times", containerString,
						f.ResourceName(podNode), containerStatus.RestartCount),
				})
			}
		}
	}

	return markers
}

func containerIsNonRoot(pod *kapi.Pod, container string) bool {
	for _, c := range pod.Spec.Containers {
		if c.Name != container || c.SecurityContext == nil {
			continue
		}
		switch {
		case c.SecurityContext.RunAsUser != nil && *c.SecurityContext.RunAsUser != 0:
			//c.SecurityContext.RunAsNonRoot != nil && *c.SecurityContext.RunAsNonRoot,
			return true
		}
	}
	return false
}

func containerCrashLoopBackOff(status kapi.ContainerStatus) bool {
	return status.State.Waiting != nil && status.State.Waiting.Reason == "CrashLoopBackOff"
}

func ContainerRestartedRecently(status kapi.ContainerStatus, now unversioned.Time) bool {
	if status.RestartCount == 0 {
		return false
	}
	if status.LastTerminationState.Terminated != nil && now.Sub(status.LastTerminationState.Terminated.FinishedAt.Time) < RestartRecentDuration {
		return true
	}
	return false
}

func containerRestartedFrequently(status kapi.ContainerStatus) bool {
	return status.RestartCount > RestartThreshold
}
