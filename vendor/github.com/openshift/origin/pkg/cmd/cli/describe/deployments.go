package describe

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	rcutils "k8s.io/kubernetes/pkg/controller/replication"
	kctl "k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/openshift/origin/pkg/api/graph"
	kubegraph "github.com/openshift/origin/pkg/api/kubegraph/nodes"
	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deployedges "github.com/openshift/origin/pkg/deploy/graph"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph/nodes"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

const (
	// maxDisplayDeployments is the number of deployments to show when describing
	// deployment configuration.
	maxDisplayDeployments = 3

	// maxDisplayDeploymentsEvents is the number of events to display when
	// describing the deployment configuration.
	// TODO: Make the estimation of this number more sophisticated and make this
	// number configurable via DescriberSettings
	maxDisplayDeploymentsEvents = 8
)

// DeploymentConfigDescriber generates information about a DeploymentConfig
type DeploymentConfigDescriber struct {
	osClient   client.Interface
	kubeClient kclient.Interface

	config *deployapi.DeploymentConfig
}

// NewDeploymentConfigDescriber returns a new DeploymentConfigDescriber
func NewDeploymentConfigDescriber(client client.Interface, kclient kclient.Interface, config *deployapi.DeploymentConfig) *DeploymentConfigDescriber {
	return &DeploymentConfigDescriber{
		osClient:   client,
		kubeClient: kclient,
		config:     config,
	}
}

// Describe returns the description of a DeploymentConfig
func (d *DeploymentConfigDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	var deploymentConfig *deployapi.DeploymentConfig
	if d.config != nil {
		// If a deployment config is already provided use that.
		// This is used by `oc rollback --dry-run`.
		deploymentConfig = d.config
	} else {
		var err error
		deploymentConfig, err = d.osClient.DeploymentConfigs(namespace).Get(name)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, deploymentConfig.ObjectMeta)
		var (
			deploymentsHistory   []kapi.ReplicationController
			activeDeploymentName string
		)

		if d.config == nil {
			if rcs, err := d.kubeClient.ReplicationControllers(namespace).List(kapi.ListOptions{LabelSelector: deployutil.ConfigSelector(deploymentConfig.Name)}); err == nil {
				deploymentsHistory = rcs.Items
			}
		}

		if deploymentConfig.Status.LatestVersion == 0 {
			formatString(out, "Latest Version", "Not deployed")
		} else {
			formatString(out, "Latest Version", strconv.FormatInt(deploymentConfig.Status.LatestVersion, 10))
		}

		printDeploymentConfigSpec(d.kubeClient, *deploymentConfig, out)
		fmt.Fprintln(out)

		latestDeploymentName := deployutil.LatestDeploymentNameForConfig(deploymentConfig)
		if activeDeployment := deployutil.ActiveDeployment(deploymentsHistory); activeDeployment != nil {
			activeDeploymentName = activeDeployment.Name
		}

		var deployment *kapi.ReplicationController
		isNotDeployed := len(deploymentsHistory) == 0
		for _, item := range deploymentsHistory {
			if item.Name == latestDeploymentName {
				deployment = &item
			}
		}
		if deployment == nil {
			isNotDeployed = true
		}

		if isNotDeployed {
			formatString(out, "Latest Deployment", "<none>")
		} else {
			header := fmt.Sprintf("Deployment #%d (latest)", deployutil.DeploymentVersionFor(deployment))
			// Show details if the current deployment is the active one or it is the
			// initial deployment.
			printDeploymentRc(deployment, d.kubeClient, out, header, (deployment.Name == activeDeploymentName) || len(deploymentsHistory) == 1)
		}

		// We don't show the deployment history when running `oc rollback --dry-run`.
		if d.config == nil && !isNotDeployed {
			sorted := deploymentsHistory
			sort.Sort(sort.Reverse(rcutils.OverlappingControllers(sorted)))
			counter := 1
			for _, item := range sorted {
				if item.Name != latestDeploymentName && deploymentConfig.Name == deployutil.DeploymentConfigNameFor(&item) {
					header := fmt.Sprintf("Deployment #%d", deployutil.DeploymentVersionFor(&item))
					printDeploymentRc(&item, d.kubeClient, out, header, item.Name == activeDeploymentName)
					counter++
				}
				if counter == maxDisplayDeployments {
					break
				}
			}
		}

		if settings.ShowEvents {
			// Events
			if events, err := d.kubeClient.Events(deploymentConfig.Namespace).Search(deploymentConfig); err == nil && events != nil {
				latestDeploymentEvents := &kapi.EventList{Items: []kapi.Event{}}
				for i := len(events.Items); i != 0 && i > len(events.Items)-maxDisplayDeploymentsEvents; i-- {
					latestDeploymentEvents.Items = append(latestDeploymentEvents.Items, events.Items[i-1])
				}
				fmt.Fprintln(out)
				kctl.DescribeEvents(latestDeploymentEvents, out)
			}
		}
		return nil
	})
}

func multilineStringArray(sep, indent string, args ...string) string {
	for i, s := range args {
		if strings.HasSuffix(s, "\n") {
			s = strings.TrimSuffix(s, "\n")
		}
		if strings.Contains(s, "\n") {
			s = "\n" + indent + strings.Join(strings.Split(s, "\n"), "\n"+indent)
		}
		args[i] = s
	}
	strings.TrimRight(args[len(args)-1], "\n ")
	return strings.Join(args, " ")
}

func printStrategy(strategy deployapi.DeploymentStrategy, indent string, w *tabwriter.Writer) {
	if strategy.CustomParams != nil {
		if len(strategy.CustomParams.Image) == 0 {
			fmt.Fprintf(w, "%sImage:\t%s\n", indent, "<default>")
		} else {
			fmt.Fprintf(w, "%sImage:\t%s\n", indent, strategy.CustomParams.Image)
		}

		if len(strategy.CustomParams.Environment) > 0 {
			fmt.Fprintf(w, "%sEnvironment:\t%s\n", indent, formatLabels(convertEnv(strategy.CustomParams.Environment)))
		}

		if len(strategy.CustomParams.Command) > 0 {
			fmt.Fprintf(w, "%sCommand:\t%v\n", indent, multilineStringArray(" ", "\t  ", strategy.CustomParams.Command...))
		}
	}

	if strategy.RecreateParams != nil {
		pre := strategy.RecreateParams.Pre
		mid := strategy.RecreateParams.Mid
		post := strategy.RecreateParams.Post
		if pre != nil {
			printHook("Pre-deployment", pre, indent, w)
		}
		if mid != nil {
			printHook("Mid-deployment", mid, indent, w)
		}
		if post != nil {
			printHook("Post-deployment", post, indent, w)
		}
	}

	if strategy.RollingParams != nil {
		pre := strategy.RollingParams.Pre
		post := strategy.RollingParams.Post
		if pre != nil {
			printHook("Pre-deployment", pre, indent, w)
		}
		if post != nil {
			printHook("Post-deployment", post, indent, w)
		}
	}
}

func printHook(prefix string, hook *deployapi.LifecycleHook, indent string, w io.Writer) {
	if hook.ExecNewPod != nil {
		fmt.Fprintf(w, "%s%s hook (pod type, failure policy: %s):\n", indent, prefix, hook.FailurePolicy)
		fmt.Fprintf(w, "%s  Container:\t%s\n", indent, hook.ExecNewPod.ContainerName)
		fmt.Fprintf(w, "%s  Command:\t%v\n", indent, multilineStringArray(" ", "\t  ", hook.ExecNewPod.Command...))
		if len(hook.ExecNewPod.Env) > 0 {
			fmt.Fprintf(w, "%s  Env:\t%s\n", indent, formatLabels(convertEnv(hook.ExecNewPod.Env)))
		}
	}
	if len(hook.TagImages) > 0 {
		fmt.Fprintf(w, "%s%s hook (tag images, failure policy: %s):\n", indent, prefix, hook.FailurePolicy)
		for _, image := range hook.TagImages {
			fmt.Fprintf(w, "%s  Tag:\tcontainer %s to %s %s %s\n", indent, image.ContainerName, image.To.Kind, image.To.Name, image.To.Namespace)
		}
	}
}

func printTriggers(triggers []deployapi.DeploymentTriggerPolicy, w *tabwriter.Writer) {
	if len(triggers) == 0 {
		formatString(w, "Triggers", "<none>")
		return
	}

	labels := []string{}

	for _, t := range triggers {
		switch t.Type {
		case deployapi.DeploymentTriggerOnConfigChange:
			labels = append(labels, "Config")
		case deployapi.DeploymentTriggerOnImageChange:
			if len(t.ImageChangeParams.From.Name) > 0 {
				name, tag, _ := imageapi.SplitImageStreamTag(t.ImageChangeParams.From.Name)
				labels = append(labels, fmt.Sprintf("Image(%s@%s, auto=%v)", name, tag, t.ImageChangeParams.Automatic))
			}
		}
	}

	desc := strings.Join(labels, ", ")
	formatString(w, "Triggers", desc)
}

func printDeploymentConfigSpec(kc kclient.Interface, dc deployapi.DeploymentConfig, w *tabwriter.Writer) error {
	spec := dc.Spec
	// Selector
	formatString(w, "Selector", formatLabels(spec.Selector))

	// Replicas
	test := ""
	if spec.Test {
		test = " (test, will be scaled down between deployments)"
	}
	formatString(w, "Replicas", fmt.Sprintf("%d%s", spec.Replicas, test))

	if spec.Paused {
		formatString(w, "Paused", "yes")
	}

	// Autoscaling info
	printAutoscalingInfo(deployapi.Resource("DeploymentConfig"), dc.Namespace, dc.Name, kc, w)

	// Triggers
	printTriggers(spec.Triggers, w)

	// Strategy
	formatString(w, "Strategy", spec.Strategy.Type)
	printStrategy(spec.Strategy, "  ", w)

	if dc.Spec.MinReadySeconds > 0 {
		formatString(w, "MinReadySeconds", fmt.Sprintf("%d", spec.MinReadySeconds))
	}

	// Pod template
	fmt.Fprintf(w, "Template:\n")
	kctl.DescribePodTemplate(spec.Template, w)

	return nil
}

// TODO: Move this upstream
func printAutoscalingInfo(res unversioned.GroupResource, namespace, name string, kclient kclient.Interface, w *tabwriter.Writer) {
	hpaList, err := kclient.Autoscaling().HorizontalPodAutoscalers(namespace).List(kapi.ListOptions{LabelSelector: labels.Everything()})
	if err != nil {
		return
	}

	scaledBy := []autoscaling.HorizontalPodAutoscaler{}
	for _, hpa := range hpaList.Items {
		if hpa.Spec.ScaleTargetRef.Name == name && hpa.Spec.ScaleTargetRef.Kind == res.String() {
			scaledBy = append(scaledBy, hpa)
		}
	}

	for _, hpa := range scaledBy {
		cpuUtil := ""
		if hpa.Spec.TargetCPUUtilizationPercentage != nil {
			cpuUtil = fmt.Sprintf(", triggered at %d%% CPU usage", *hpa.Spec.TargetCPUUtilizationPercentage)
		}
		fmt.Fprintf(w, "Autoscaling:\tbetween %d and %d replicas%s\n", *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas, cpuUtil)
		// TODO: Print a warning in case of multiple hpas.
		// Related oc status PR: https://github.com/openshift/origin/pull/7799
		break
	}
}

func printDeploymentRc(deployment *kapi.ReplicationController, kubeClient kclient.Interface, w io.Writer, header string, verbose bool) error {
	if len(header) > 0 {
		fmt.Fprintf(w, "%v:\n", header)
	}

	if verbose {
		fmt.Fprintf(w, "\tName:\t%s\n", deployment.Name)
	}
	timeAt := strings.ToLower(formatRelativeTime(deployment.CreationTimestamp.Time))
	fmt.Fprintf(w, "\tCreated:\t%s ago\n", timeAt)
	fmt.Fprintf(w, "\tStatus:\t%s\n", deployutil.DeploymentStatusFor(deployment))
	fmt.Fprintf(w, "\tReplicas:\t%d current / %d desired\n", deployment.Status.Replicas, deployment.Spec.Replicas)

	if verbose {
		fmt.Fprintf(w, "\tSelector:\t%s\n", formatLabels(deployment.Spec.Selector))
		fmt.Fprintf(w, "\tLabels:\t%s\n", formatLabels(deployment.Labels))
		running, waiting, succeeded, failed, err := getPodStatusForDeployment(deployment, kubeClient)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\tPods Status:\t%d Running / %d Waiting / %d Succeeded / %d Failed\n", running, waiting, succeeded, failed)
	}

	return nil
}

func getPodStatusForDeployment(deployment *kapi.ReplicationController, kubeClient kclient.Interface) (running, waiting, succeeded, failed int, err error) {
	rcPods, err := kubeClient.Pods(deployment.Namespace).List(kapi.ListOptions{LabelSelector: labels.Set(deployment.Spec.Selector).AsSelector()})
	if err != nil {
		return
	}
	for _, pod := range rcPods.Items {
		switch pod.Status.Phase {
		case kapi.PodRunning:
			running++
		case kapi.PodPending:
			waiting++
		case kapi.PodSucceeded:
			succeeded++
		case kapi.PodFailed:
			failed++
		}
	}
	return
}

type LatestDeploymentsDescriber struct {
	count      int
	osClient   client.Interface
	kubeClient kclient.Interface
}

// NewLatestDeploymentsDescriber lists the latest deployments limited to "count". In case count == -1, list back to the last successful.
func NewLatestDeploymentsDescriber(client client.Interface, kclient kclient.Interface, count int) *LatestDeploymentsDescriber {
	return &LatestDeploymentsDescriber{
		count:      count,
		osClient:   client,
		kubeClient: kclient,
	}
}

// Describe returns the description of the latest deployments for a config
func (d *LatestDeploymentsDescriber) Describe(namespace, name string) (string, error) {
	var f formatter

	config, err := d.osClient.DeploymentConfigs(namespace).Get(name)
	if err != nil {
		return "", err
	}

	var deployments []kapi.ReplicationController
	if d.count == -1 || d.count > 1 {
		list, err := d.kubeClient.ReplicationControllers(namespace).List(kapi.ListOptions{LabelSelector: deployutil.ConfigSelector(name)})
		if err != nil && !kerrors.IsNotFound(err) {
			return "", err
		}
		deployments = list.Items
	} else {
		deploymentName := deployutil.LatestDeploymentNameForConfig(config)
		deployment, err := d.kubeClient.ReplicationControllers(config.Namespace).Get(deploymentName)
		if err != nil && !kerrors.IsNotFound(err) {
			return "", err
		}
		if deployment != nil {
			deployments = []kapi.ReplicationController{*deployment}
		}
	}

	g := graph.New()
	dcNode := deploygraph.EnsureDeploymentConfigNode(g, config)
	for i := range deployments {
		kubegraph.EnsureReplicationControllerNode(g, &deployments[i])
	}
	deployedges.AddTriggerEdges(g, dcNode)
	deployedges.AddDeploymentEdges(g, dcNode)
	activeDeployment, inactiveDeployments := deployedges.RelevantDeployments(g, dcNode)

	return tabbedString(func(out *tabwriter.Writer) error {
		descriptions := describeDeployments(f, dcNode, activeDeployment, inactiveDeployments, nil, d.count)
		for i, description := range descriptions {
			descriptions[i] = fmt.Sprintf("%v %v", name, description)
		}
		printLines(out, "", 0, descriptions...)
		return nil
	})
}
