package describe

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	units "github.com/docker/go-units"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kctl "k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	imageapi "github.com/openshift/origin/pkg/image/api"
	oauthapi "github.com/openshift/origin/pkg/oauth/api"
	projectapi "github.com/openshift/origin/pkg/project/api"
	quotaapi "github.com/openshift/origin/pkg/quota/api"
	routeapi "github.com/openshift/origin/pkg/route/api"
	sdnapi "github.com/openshift/origin/pkg/sdn/api"
	templateapi "github.com/openshift/origin/pkg/template/api"
	userapi "github.com/openshift/origin/pkg/user/api"
)

func describerMap(c *client.Client, kclient kclient.Interface, host string) map[unversioned.GroupKind]kctl.Describer {
	m := map[unversioned.GroupKind]kctl.Describer{
		buildapi.Kind("Build"):                        &BuildDescriber{c, kclient},
		buildapi.Kind("BuildConfig"):                  &BuildConfigDescriber{c, kclient, host},
		deployapi.Kind("DeploymentConfig"):            &DeploymentConfigDescriber{c, kclient, nil},
		authorizationapi.Kind("Identity"):             &IdentityDescriber{c},
		imageapi.Kind("Image"):                        &ImageDescriber{c},
		imageapi.Kind("ImageStream"):                  &ImageStreamDescriber{c},
		imageapi.Kind("ImageStreamTag"):               &ImageStreamTagDescriber{c},
		imageapi.Kind("ImageStreamImage"):             &ImageStreamImageDescriber{c},
		routeapi.Kind("Route"):                        &RouteDescriber{c, kclient},
		projectapi.Kind("Project"):                    &ProjectDescriber{c, kclient},
		templateapi.Kind("Template"):                  &TemplateDescriber{c, meta.NewAccessor(), kapi.Scheme, nil},
		authorizationapi.Kind("Policy"):               &PolicyDescriber{c},
		authorizationapi.Kind("PolicyBinding"):        &PolicyBindingDescriber{c},
		authorizationapi.Kind("RoleBinding"):          &RoleBindingDescriber{c},
		authorizationapi.Kind("Role"):                 &RoleDescriber{c},
		authorizationapi.Kind("ClusterPolicy"):        &ClusterPolicyDescriber{c},
		authorizationapi.Kind("ClusterPolicyBinding"): &ClusterPolicyBindingDescriber{c},
		authorizationapi.Kind("ClusterRoleBinding"):   &ClusterRoleBindingDescriber{c},
		authorizationapi.Kind("ClusterRole"):          &ClusterRoleDescriber{c},
		oauthapi.Kind("OAuthAccessToken"):             &OAuthAccessTokenDescriber{c},
		userapi.Kind("User"):                          &UserDescriber{c},
		userapi.Kind("Group"):                         &GroupDescriber{c.Groups()},
		userapi.Kind("UserIdentityMapping"):           &UserIdentityMappingDescriber{c},
		quotaapi.Kind("ClusterResourceQuota"):         &ClusterQuotaDescriber{c},
		quotaapi.Kind("AppliedClusterResourceQuota"):  &AppliedClusterQuotaDescriber{c},
		sdnapi.Kind("ClusterNetwork"):                 &ClusterNetworkDescriber{c},
		sdnapi.Kind("HostSubnet"):                     &HostSubnetDescriber{c},
		sdnapi.Kind("NetNamespace"):                   &NetNamespaceDescriber{c},
		sdnapi.Kind("EgressNetworkPolicy"):            &EgressNetworkPolicyDescriber{c},
	}
	return m
}

// DescribableResources lists all of the resource types we can describe
func DescribableResources() []string {
	// Include describable resources in kubernetes
	keys := kctl.DescribableResources()

	for k := range describerMap(nil, nil, "") {
		resource := strings.ToLower(k.Kind)
		keys = append(keys, resource)
	}
	return keys
}

// DescriberFor returns a describer for a given kind of resource
func DescriberFor(kind unversioned.GroupKind, c *client.Client, kclient kclient.Interface, host string) (kctl.Describer, bool) {
	f, ok := describerMap(c, kclient, host)[kind]
	if ok {
		return f, true
	}
	return nil, false
}

// BuildDescriber generates information about a build
type BuildDescriber struct {
	osClient   client.Interface
	kubeClient kclient.Interface
}

// Describe returns the description of a build
func (d *BuildDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.osClient.Builds(namespace)
	build, err := c.Get(name)
	if err != nil {
		return "", err
	}
	events, _ := d.kubeClient.Events(namespace).Search(build)
	if events == nil {
		events = &kapi.EventList{}
	}
	// get also pod events and merge it all into one list for describe
	if pod, err := d.kubeClient.Pods(namespace).Get(buildapi.GetBuildPodName(build)); err == nil {
		if podEvents, _ := d.kubeClient.Events(namespace).Search(pod); podEvents != nil {
			events.Items = append(events.Items, podEvents.Items...)
		}
	}
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, build.ObjectMeta)

		fmt.Fprintln(out, "")

		status := bold(build.Status.Phase)
		if build.Status.Message != "" {
			status += " (" + build.Status.Message + ")"
		}
		formatString(out, "Status", status)

		if build.Status.StartTimestamp != nil && !build.Status.StartTimestamp.IsZero() {
			formatString(out, "Started", build.Status.StartTimestamp.Time.Format(time.RFC1123))
		}

		// Create the time object with second-level precision so we don't get
		// output like "duration: 1.2724395728934s"
		formatString(out, "Duration", describeBuildDuration(build))

		if build.Status.Config != nil {
			formatString(out, "Build Config", build.Status.Config.Name)
		}
		formatString(out, "Build Pod", buildapi.GetBuildPodName(build))

		describeCommonSpec(build.Spec.CommonSpec, out)
		describeBuildTriggerCauses(build.Spec.TriggeredBy, out)

		if settings.ShowEvents {
			kctl.DescribeEvents(events, out)
		}

		return nil
	})
}

func describeBuildDuration(build *buildapi.Build) string {
	t := unversioned.Now().Rfc3339Copy()
	if build.Status.StartTimestamp == nil &&
		build.Status.CompletionTimestamp != nil &&
		(build.Status.Phase == buildapi.BuildPhaseCancelled ||
			build.Status.Phase == buildapi.BuildPhaseFailed ||
			build.Status.Phase == buildapi.BuildPhaseError) {
		// time a build waited for its pod before ultimately being cancelled before that pod was created
		return fmt.Sprintf("waited for %s", build.Status.CompletionTimestamp.Rfc3339Copy().Time.Sub(build.CreationTimestamp.Rfc3339Copy().Time))
	} else if build.Status.StartTimestamp == nil && build.Status.Phase != buildapi.BuildPhaseCancelled {
		// time a new build has been waiting for its pod to be created so it can run
		return fmt.Sprintf("waiting for %v", t.Sub(build.CreationTimestamp.Rfc3339Copy().Time))
	} else if build.Status.StartTimestamp != nil && build.Status.CompletionTimestamp == nil {
		// time a still running build has been running in a pod
		return fmt.Sprintf("running for %v", build.Status.Duration)
	}
	return fmt.Sprintf("%v", build.Status.Duration)
}

// BuildConfigDescriber generates information about a buildConfig
type BuildConfigDescriber struct {
	client.Interface
	kubeClient kclient.Interface
	host       string
}

func nameAndNamespace(ns, name string) string {
	if len(ns) != 0 {
		return fmt.Sprintf("%s/%s", ns, name)
	}
	return name
}

func describeCommonSpec(p buildapi.CommonSpec, out *tabwriter.Writer) {
	formatString(out, "\nStrategy", buildapi.StrategyType(p.Strategy))
	noneType := true
	if p.Source.Git != nil {
		noneType = false
		formatString(out, "URL", p.Source.Git.URI)
		if len(p.Source.Git.Ref) > 0 {
			formatString(out, "Ref", p.Source.Git.Ref)
		}
		if len(p.Source.ContextDir) > 0 {
			formatString(out, "ContextDir", p.Source.ContextDir)
		}
		if p.Source.SourceSecret != nil {
			formatString(out, "Source Secret", p.Source.SourceSecret.Name)
		}
		squashGitInfo(p.Revision, out)
	}
	if p.Source.Dockerfile != nil {
		if len(strings.TrimSpace(*p.Source.Dockerfile)) == 0 {
			formatString(out, "Dockerfile", "")
		} else {
			fmt.Fprintf(out, "Dockerfile:\n")
			for _, s := range strings.Split(*p.Source.Dockerfile, "\n") {
				fmt.Fprintf(out, "  %s\n", s)
			}
		}
	}
	switch {
	case p.Strategy.DockerStrategy != nil:
		describeDockerStrategy(p.Strategy.DockerStrategy, out)
	case p.Strategy.SourceStrategy != nil:
		describeSourceStrategy(p.Strategy.SourceStrategy, out)
	case p.Strategy.CustomStrategy != nil:
		describeCustomStrategy(p.Strategy.CustomStrategy, out)
	case p.Strategy.JenkinsPipelineStrategy != nil:
		describeJenkinsPipelineStrategy(p.Strategy.JenkinsPipelineStrategy, out)
	}

	if p.Output.To != nil {
		formatString(out, "Output to", fmt.Sprintf("%s %s", p.Output.To.Kind, nameAndNamespace(p.Output.To.Namespace, p.Output.To.Name)))
	}

	if p.Source.Binary != nil {
		noneType = false
		if len(p.Source.Binary.AsFile) > 0 {
			formatString(out, "Binary", fmt.Sprintf("provided as file %q on build", p.Source.Binary.AsFile))
		} else {
			formatString(out, "Binary", "provided on build")
		}
	}

	if len(p.Source.Secrets) > 0 {
		result := []string{}
		for _, s := range p.Source.Secrets {
			result = append(result, fmt.Sprintf("%s->%s", s.Secret.Name, filepath.Clean(s.DestinationDir)))
		}
		formatString(out, "Build Secrets", strings.Join(result, ","))
	}
	if len(p.Source.Images) == 1 && len(p.Source.Images[0].Paths) == 1 {
		noneType = false
		image := p.Source.Images[0]
		path := image.Paths[0]
		formatString(out, "Image Source", fmt.Sprintf("copies %s from %s to %s", path.SourcePath, nameAndNamespace(image.From.Namespace, image.From.Name), path.DestinationDir))
	} else {
		for _, image := range p.Source.Images {
			noneType = false
			formatString(out, "Image Source", fmt.Sprintf("%s", nameAndNamespace(image.From.Namespace, image.From.Name)))
			for _, path := range image.Paths {
				fmt.Fprintf(out, "\t- %s -> %s\n", path.SourcePath, path.DestinationDir)
			}
		}
	}

	if noneType {
		formatString(out, "Empty Source", "no input source provided")
	}

	describePostCommitHook(p.PostCommit, out)

	if p.Output.PushSecret != nil {
		formatString(out, "Push Secret", p.Output.PushSecret.Name)
	}

	if p.CompletionDeadlineSeconds != nil {
		formatString(out, "Fail Build After", time.Duration(*p.CompletionDeadlineSeconds)*time.Second)
	}
}

func describePostCommitHook(hook buildapi.BuildPostCommitSpec, out *tabwriter.Writer) {
	command := hook.Command
	args := hook.Args
	script := hook.Script
	if len(command) == 0 && len(args) == 0 && len(script) == 0 {
		// Post commit hook is not set, nothing to do.
		return
	}
	if len(script) != 0 {
		command = []string{"/bin/sh", "-ic"}
		if len(args) > 0 {
			args = append([]string{script, command[0]}, args...)
		} else {
			args = []string{script}
		}
	}
	if len(command) == 0 {
		command = []string{"<image-entrypoint>"}
	}
	all := append(command, args...)
	for i, v := range all {
		all[i] = fmt.Sprintf("%q", v)
	}
	formatString(out, "Post Commit Hook", fmt.Sprintf("[%s]", strings.Join(all, ", ")))
}

func describeSourceStrategy(s *buildapi.SourceBuildStrategy, out *tabwriter.Writer) {
	if len(s.From.Name) != 0 {
		formatString(out, "From Image", fmt.Sprintf("%s %s", s.From.Kind, nameAndNamespace(s.From.Namespace, s.From.Name)))
	}
	if len(s.Scripts) != 0 {
		formatString(out, "Scripts", s.Scripts)
	}
	if s.PullSecret != nil {
		formatString(out, "Pull Secret Name", s.PullSecret.Name)
	}
	if s.Incremental != nil && *s.Incremental {
		formatString(out, "Incremental Build", "yes")
	}
	if s.ForcePull {
		formatString(out, "Force Pull", "yes")
	}
}

func describeDockerStrategy(s *buildapi.DockerBuildStrategy, out *tabwriter.Writer) {
	if s.From != nil && len(s.From.Name) != 0 {
		formatString(out, "From Image", fmt.Sprintf("%s %s", s.From.Kind, nameAndNamespace(s.From.Namespace, s.From.Name)))
	}
	if len(s.DockerfilePath) != 0 {
		formatString(out, "Dockerfile Path", s.DockerfilePath)
	}
	if s.PullSecret != nil {
		formatString(out, "Pull Secret Name", s.PullSecret.Name)
	}
	if s.NoCache {
		formatString(out, "No Cache", "true")
	}
	if s.ForcePull {
		formatString(out, "Force Pull", "true")
	}
}

func describeCustomStrategy(s *buildapi.CustomBuildStrategy, out *tabwriter.Writer) {
	if len(s.From.Name) != 0 {
		formatString(out, "Image Reference", fmt.Sprintf("%s %s", s.From.Kind, nameAndNamespace(s.From.Namespace, s.From.Name)))
	}
	if s.ExposeDockerSocket {
		formatString(out, "Expose Docker Socket", "yes")
	}
	if s.ForcePull {
		formatString(out, "Force Pull", "yes")
	}
	if s.PullSecret != nil {
		formatString(out, "Pull Secret Name", s.PullSecret.Name)
	}
	for i, env := range s.Env {
		if i == 0 {
			formatString(out, "Environment", formatEnv(env))
		} else {
			formatString(out, "", formatEnv(env))
		}
	}
}

func describeJenkinsPipelineStrategy(s *buildapi.JenkinsPipelineBuildStrategy, out *tabwriter.Writer) {
	if len(s.JenkinsfilePath) != 0 {
		formatString(out, "Jenkinsfile path", s.JenkinsfilePath)
	}
	if len(s.Jenkinsfile) != 0 {
		fmt.Fprintf(out, "Jenkinsfile contents:\n")
		for _, s := range strings.Split(s.Jenkinsfile, "\n") {
			fmt.Fprintf(out, "  %s\n", s)
		}
	}
	if len(s.Jenkinsfile) == 0 && len(s.JenkinsfilePath) == 0 {
		formatString(out, "Jenkinsfile", "from source repository root")
	}
}

// DescribeTriggers generates information about the triggers associated with a
// buildconfig
func (d *BuildConfigDescriber) DescribeTriggers(bc *buildapi.BuildConfig, out *tabwriter.Writer) {
	describeBuildTriggers(bc.Spec.Triggers, bc.Name, bc.Namespace, out, d)
}

func describeBuildTriggers(triggers []buildapi.BuildTriggerPolicy, name, namespace string, w *tabwriter.Writer, d *BuildConfigDescriber) {
	if len(triggers) == 0 {
		formatString(w, "Triggered by", "<none>")
		return
	}

	labels := []string{}

	for _, t := range triggers {
		switch t.Type {
		case buildapi.GitHubWebHookBuildTriggerType, buildapi.GenericWebHookBuildTriggerType:
			continue
		case buildapi.ConfigChangeBuildTriggerType:
			labels = append(labels, "Config")
		case buildapi.ImageChangeBuildTriggerType:
			if t.ImageChange != nil && t.ImageChange.From != nil && len(t.ImageChange.From.Name) > 0 {
				labels = append(labels, fmt.Sprintf("Image(%s %s)", t.ImageChange.From.Kind, t.ImageChange.From.Name))
			} else {
				labels = append(labels, string(t.Type))
			}
		case "":
			labels = append(labels, "<unknown>")
		default:
			labels = append(labels, string(t.Type))
		}
	}

	desc := strings.Join(labels, ", ")
	formatString(w, "Triggered by", desc)

	webHooks := webHooksDescribe(triggers, name, namespace, d.Interface)
	for webHookType, webHookDesc := range webHooks {
		fmt.Fprintf(w, "Webhook %s:\n", strings.Title(webHookType))
		for _, trigger := range webHookDesc {
			fmt.Fprintf(w, "\tURL:\t%s\n", trigger.URL)
			if webHookType == string(buildapi.GenericWebHookBuildTriggerType) && trigger.AllowEnv != nil {
				fmt.Fprintf(w, fmt.Sprintf("\t%s:\t%v\n", "AllowEnv", *trigger.AllowEnv))
			}
		}
	}
}

// Describe returns the description of a buildConfig
func (d *BuildConfigDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.BuildConfigs(namespace)
	buildConfig, err := c.Get(name)
	if err != nil {
		return "", err
	}
	buildList, err := d.Builds(namespace).List(kapi.ListOptions{})
	if err != nil {
		return "", err
	}
	buildList.Items = buildapi.FilterBuilds(buildList.Items, buildapi.ByBuildConfigPredicate(name))

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, buildConfig.ObjectMeta)
		if buildConfig.Status.LastVersion == 0 {
			formatString(out, "Latest Version", "Never built")
		} else {
			formatString(out, "Latest Version", strconv.FormatInt(buildConfig.Status.LastVersion, 10))
		}
		describeCommonSpec(buildConfig.Spec.CommonSpec, out)
		formatString(out, "\nBuild Run Policy", string(buildConfig.Spec.RunPolicy))
		d.DescribeTriggers(buildConfig, out)

		if len(buildList.Items) > 0 {
			fmt.Fprintf(out, "\nBuild\tStatus\tDuration\tCreation Time\n")

			builds := buildList.Items
			sort.Sort(sort.Reverse(buildapi.BuildSliceByCreationTimestamp(builds)))

			for i, build := range builds {
				fmt.Fprintf(out, "%s \t%s \t%v \t%v\n",
					build.Name,
					strings.ToLower(string(build.Status.Phase)),
					describeBuildDuration(&build),
					build.CreationTimestamp.Rfc3339Copy().Time)
				// only print the 10 most recent builds.
				if i == 9 {
					break
				}
			}
		}

		if settings.ShowEvents {
			events, _ := d.kubeClient.Events(namespace).Search(buildConfig)
			if events != nil {
				fmt.Fprint(out, "\n")
				kctl.DescribeEvents(events, out)
			}
		}
		return nil
	})
}

// OAuthAccessTokenDescriber generates information about an OAuth Acess Token (OAuth)
type OAuthAccessTokenDescriber struct {
	client.Interface
}

func (d *OAuthAccessTokenDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.OAuthAccessTokens()
	oAuthAccessToken, err := c.Get(name)
	if err != nil {
		return "", err
	}

	var timeCreated time.Time = oAuthAccessToken.ObjectMeta.CreationTimestamp.Time
	var timeExpired time.Time = timeCreated.Add(time.Duration(oAuthAccessToken.ExpiresIn) * time.Second)

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, oAuthAccessToken.ObjectMeta)
		formatString(out, "Scopes", oAuthAccessToken.Scopes)
		formatString(out, "Expires In", formatToHumanDuration(timeExpired.Sub(time.Now())))
		formatString(out, "User Name", oAuthAccessToken.UserName)
		formatString(out, "User UID", oAuthAccessToken.UserUID)
		formatString(out, "Client Name", oAuthAccessToken.ClientName)

		return nil
	})
}

// ImageDescriber generates information about a Image
type ImageDescriber struct {
	client.Interface
}

// Describe returns the description of an image
func (d *ImageDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.Images()
	image, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return describeImage(image, "")
}

func describeImage(image *imageapi.Image, imageName string) (string, error) {
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, image.ObjectMeta)
		formatString(out, "Docker Image", image.DockerImageReference)
		if len(imageName) > 0 {
			formatString(out, "Image Name", imageName)
		}
		switch l := len(image.DockerImageLayers); l {
		case 0:
			// legacy case, server does not know individual layers
			formatString(out, "Layer Size", units.HumanSize(float64(image.DockerImageMetadata.Size)))
		case 1:
			formatString(out, "Image Size", units.HumanSize(float64(image.DockerImageMetadata.Size)))
		default:
			info := []string{}
			if image.DockerImageLayers[0].LayerSize > 0 {
				info = append(info, fmt.Sprintf("first layer %s", units.HumanSize(float64(image.DockerImageLayers[0].LayerSize))))
			}
			for i := l - 1; i > 0; i-- {
				if image.DockerImageLayers[i].LayerSize == 0 {
					continue
				}
				info = append(info, fmt.Sprintf("last binary layer %s", units.HumanSize(float64(image.DockerImageLayers[i].LayerSize))))
				break
			}
			if len(info) > 0 {
				formatString(out, "Image Size", fmt.Sprintf("%s (%s)", units.HumanSize(float64(image.DockerImageMetadata.Size)), strings.Join(info, ", ")))
			} else {
				formatString(out, "Image Size", units.HumanSize(float64(image.DockerImageMetadata.Size)))
			}
		}
		//formatString(out, "Parent Image", image.DockerImageMetadata.Parent)
		formatString(out, "Image Created", fmt.Sprintf("%s ago", formatRelativeTime(image.DockerImageMetadata.Created.Time)))
		formatString(out, "Author", image.DockerImageMetadata.Author)
		formatString(out, "Arch", image.DockerImageMetadata.Architecture)
		describeDockerImage(out, image.DockerImageMetadata.Config)
		return nil
	})
}

func describeDockerImage(out *tabwriter.Writer, image *imageapi.DockerConfig) {
	if image == nil {
		return
	}
	hasCommand := false
	if len(image.Entrypoint) > 0 {
		hasCommand = true
		formatString(out, "Entrypoint", strings.Join(image.Entrypoint, " "))
	}
	if len(image.Cmd) > 0 {
		hasCommand = true
		formatString(out, "Command", strings.Join(image.Cmd, " "))
	}
	if !hasCommand {
		formatString(out, "Command", "")
	}
	formatString(out, "Working Dir", image.WorkingDir)
	formatString(out, "User", image.User)
	ports := sets.NewString()
	for k := range image.ExposedPorts {
		ports.Insert(k)
	}
	formatString(out, "Exposes Ports", strings.Join(ports.List(), ", "))
	formatMapStringString(out, "Docker Labels", image.Labels)
	for i, env := range image.Env {
		if i == 0 {
			formatString(out, "Environment", env)
		} else {
			fmt.Fprintf(out, "\t%s\n", env)
		}
	}
	volumes := sets.NewString()
	for k := range image.Volumes {
		volumes.Insert(k)
	}
	for i, volume := range volumes.List() {
		if i == 0 {
			formatString(out, "Volumes", volume)
		} else {
			fmt.Fprintf(out, "\t%s\n", volume)
		}
	}
}

// ImageStreamTagDescriber generates information about a ImageStreamTag (Image).
type ImageStreamTagDescriber struct {
	client.Interface
}

// Describe returns the description of an imageStreamTag
func (d *ImageStreamTagDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ImageStreamTags(namespace)
	repo, tag, err := imageapi.ParseImageStreamTagName(name)
	if err != nil {
		return "", err
	}
	if len(tag) == 0 {
		// TODO use repo's preferred default, when that's coded
		tag = imageapi.DefaultImageTag
	}
	imageStreamTag, err := c.Get(repo, tag)
	if err != nil {
		return "", err
	}

	return describeImage(&imageStreamTag.Image, imageStreamTag.Image.Name)
}

// ImageStreamImageDescriber generates information about a ImageStreamImage (Image).
type ImageStreamImageDescriber struct {
	client.Interface
}

// Describe returns the description of an imageStreamImage
func (d *ImageStreamImageDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ImageStreamImages(namespace)
	repo, id, err := imageapi.ParseImageStreamImageName(name)
	if err != nil {
		return "", err
	}
	imageStreamImage, err := c.Get(repo, id)
	if err != nil {
		return "", err
	}

	return describeImage(&imageStreamImage.Image, imageStreamImage.Image.Name)
}

// ImageStreamDescriber generates information about a ImageStream (Image).
type ImageStreamDescriber struct {
	client.Interface
}

// Describe returns the description of an imageStream
func (d *ImageStreamDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ImageStreams(namespace)
	imageStream, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, imageStream.ObjectMeta)
		formatString(out, "Docker Pull Spec", imageStream.Status.DockerImageRepository)
		formatImageStreamTags(out, imageStream)
		return nil
	})
}

// RouteDescriber generates information about a Route
type RouteDescriber struct {
	client.Interface
	kubeClient kclient.Interface
}

type routeEndpointInfo struct {
	*kapi.Endpoints
	Err error
}

// Describe returns the description of a route
func (d *RouteDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.Routes(namespace)
	route, err := c.Get(name)
	if err != nil {
		return "", err
	}

	backends := append([]routeapi.RouteTargetReference{route.Spec.To}, route.Spec.AlternateBackends...)
	totalWeight := int32(0)
	endpoints := make(map[string]routeEndpointInfo)
	for _, backend := range backends {
		if backend.Weight != nil {
			totalWeight += *backend.Weight
		}
		ep, endpointsErr := d.kubeClient.Endpoints(namespace).Get(backend.Name)
		endpoints[backend.Name] = routeEndpointInfo{ep, endpointsErr}
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, route.ObjectMeta)
		if len(route.Spec.Host) > 0 {
			formatString(out, "Requested Host", route.Spec.Host)
			for _, ingress := range route.Status.Ingress {
				if route.Spec.Host != ingress.Host {
					continue
				}
				switch status, condition := routeapi.IngressConditionStatus(&ingress, routeapi.RouteAdmitted); status {
				case kapi.ConditionTrue:
					fmt.Fprintf(out, "\t  exposed on router %s %s ago\n", ingress.RouterName, strings.ToLower(formatRelativeTime(condition.LastTransitionTime.Time)))
				case kapi.ConditionFalse:
					fmt.Fprintf(out, "\t  rejected by router %s: %s (%s ago)\n", ingress.RouterName, condition.Reason, strings.ToLower(formatRelativeTime(condition.LastTransitionTime.Time)))
					if len(condition.Message) > 0 {
						fmt.Fprintf(out, "\t    %s\n", condition.Message)
					}
				}
			}
		} else {
			formatString(out, "Requested Host", "<auto>")
		}

		for _, ingress := range route.Status.Ingress {
			if route.Spec.Host == ingress.Host {
				continue
			}
			switch status, condition := routeapi.IngressConditionStatus(&ingress, routeapi.RouteAdmitted); status {
			case kapi.ConditionTrue:
				fmt.Fprintf(out, "\t%s exposed on router %s %s ago\n", ingress.Host, ingress.RouterName, strings.ToLower(formatRelativeTime(condition.LastTransitionTime.Time)))
			case kapi.ConditionFalse:
				fmt.Fprintf(out, "\trejected by router %s: %s (%s ago)\n", ingress.RouterName, condition.Reason, strings.ToLower(formatRelativeTime(condition.LastTransitionTime.Time)))
				if len(condition.Message) > 0 {
					fmt.Fprintf(out, "\t  %s\n", condition.Message)
				}
			}
		}
		formatString(out, "Path", route.Spec.Path)

		tlsTerm := ""
		insecurePolicy := ""
		if route.Spec.TLS != nil {
			tlsTerm = string(route.Spec.TLS.Termination)
			insecurePolicy = string(route.Spec.TLS.InsecureEdgeTerminationPolicy)
		}
		formatString(out, "TLS Termination", tlsTerm)
		formatString(out, "Insecure Policy", insecurePolicy)
		if route.Spec.Port != nil {
			formatString(out, "Endpoint Port", route.Spec.Port.TargetPort.String())
		} else {
			formatString(out, "Endpoint Port", "<all endpoint ports>")
		}

		for _, backend := range backends {
			fmt.Fprintln(out)
			formatString(out, "Service", backend.Name)
			weight := int32(0)
			if backend.Weight != nil {
				weight = *backend.Weight
			}
			if weight > 0 {
				fmt.Fprintf(out, "Weight:\t%d (%d%%)\n", weight, weight*100/totalWeight)
			} else {
				formatString(out, "Weight", "0")
			}

			info := endpoints[backend.Name]
			if info.Err != nil {
				formatString(out, "Endpoints", fmt.Sprintf("<error: %v>", info.Err))
				continue
			}
			endpoints := info.Endpoints
			if len(endpoints.Subsets) == 0 {
				formatString(out, "Endpoints", "<none>")
				continue
			}

			list := []string{}
			max := 3
			count := 0
			for i := range endpoints.Subsets {
				ss := &endpoints.Subsets[i]
				for p := range ss.Ports {
					for a := range ss.Addresses {
						if len(list) < max {
							list = append(list, fmt.Sprintf("%s:%d", ss.Addresses[a].IP, ss.Ports[p].Port))
						}
						count++
					}
				}
			}
			ends := strings.Join(list, ", ")
			if count > max {
				ends += fmt.Sprintf(" + %d more...", count-max)
			}
			formatString(out, "Endpoints", ends)
		}
		return nil
	})
}

// ProjectDescriber generates information about a Project
type ProjectDescriber struct {
	osClient   client.Interface
	kubeClient kclient.Interface
}

// Describe returns the description of a project
func (d *ProjectDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	projectsClient := d.osClient.Projects()
	project, err := projectsClient.Get(name)
	if err != nil {
		return "", err
	}
	resourceQuotasClient := d.kubeClient.ResourceQuotas(name)
	resourceQuotaList, err := resourceQuotasClient.List(kapi.ListOptions{})
	if err != nil {
		return "", err
	}
	limitRangesClient := d.kubeClient.LimitRanges(name)
	limitRangeList, err := limitRangesClient.List(kapi.ListOptions{})
	if err != nil {
		return "", err
	}

	nodeSelector := ""
	if len(project.ObjectMeta.Annotations) > 0 {
		if ns, ok := project.ObjectMeta.Annotations[projectapi.ProjectNodeSelector]; ok {
			nodeSelector = ns
		}
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, project.ObjectMeta)
		formatString(out, "Display Name", project.Annotations[projectapi.ProjectDisplayName])
		formatString(out, "Description", project.Annotations[projectapi.ProjectDescription])
		formatString(out, "Status", project.Status.Phase)
		formatString(out, "Node Selector", nodeSelector)
		if len(resourceQuotaList.Items) == 0 {
			formatString(out, "Quota", "")
		} else {
			fmt.Fprintf(out, "Quota:\n")
			for i := range resourceQuotaList.Items {
				resourceQuota := &resourceQuotaList.Items[i]
				fmt.Fprintf(out, "\tName:\t%s\n", resourceQuota.Name)
				fmt.Fprintf(out, "\tResource\tUsed\tHard\n")
				fmt.Fprintf(out, "\t--------\t----\t----\n")

				resources := []kapi.ResourceName{}
				for resource := range resourceQuota.Status.Hard {
					resources = append(resources, resource)
				}
				sort.Sort(kctl.SortableResourceNames(resources))

				msg := "\t%v\t%v\t%v\n"
				for i := range resources {
					resource := resources[i]
					hardQuantity := resourceQuota.Status.Hard[resource]
					usedQuantity := resourceQuota.Status.Used[resource]
					fmt.Fprintf(out, msg, resource, usedQuantity.String(), hardQuantity.String())
				}
			}
		}
		if len(limitRangeList.Items) == 0 {
			formatString(out, "Resource limits", "")
		} else {
			fmt.Fprintf(out, "Resource limits:\n")
			for i := range limitRangeList.Items {
				limitRange := &limitRangeList.Items[i]
				fmt.Fprintf(out, "\tName:\t%s\n", limitRange.Name)
				fmt.Fprintf(out, "\tType\tResource\tMin\tMax\tDefault\n")
				fmt.Fprintf(out, "\t----\t--------\t---\t---\t---\n")
				for i := range limitRange.Spec.Limits {
					item := limitRange.Spec.Limits[i]
					maxResources := item.Max
					minResources := item.Min
					defaultResources := item.Default

					set := map[kapi.ResourceName]bool{}
					for k := range maxResources {
						set[k] = true
					}
					for k := range minResources {
						set[k] = true
					}
					for k := range defaultResources {
						set[k] = true
					}

					for k := range set {
						// if no value is set, we output -
						maxValue := "-"
						minValue := "-"
						defaultValue := "-"

						maxQuantity, maxQuantityFound := maxResources[k]
						if maxQuantityFound {
							maxValue = maxQuantity.String()
						}

						minQuantity, minQuantityFound := minResources[k]
						if minQuantityFound {
							minValue = minQuantity.String()
						}

						defaultQuantity, defaultQuantityFound := defaultResources[k]
						if defaultQuantityFound {
							defaultValue = defaultQuantity.String()
						}

						msg := "\t%v\t%v\t%v\t%v\t%v\n"
						fmt.Fprintf(out, msg, item.Type, k, minValue, maxValue, defaultValue)
					}
				}
			}
		}
		return nil
	})
}

// TemplateDescriber generates information about a template
type TemplateDescriber struct {
	client.Interface
	meta.MetadataAccessor
	runtime.ObjectTyper
	kctl.ObjectDescriber
}

// DescribeMessage prints the message that will be parameter substituted and displayed to the
// user when this template is processed.
func (d *TemplateDescriber) DescribeMessage(msg string, out *tabwriter.Writer) {
	if len(msg) == 0 {
		msg = "<none>"
	}
	formatString(out, "Message", msg)
}

// DescribeParameters prints out information about the parameters of a template
func (d *TemplateDescriber) DescribeParameters(params []templateapi.Parameter, out *tabwriter.Writer) {
	formatString(out, "Parameters", " ")
	indent := "    "
	for _, p := range params {
		formatString(out, indent+"Name", p.Name)
		if len(p.DisplayName) > 0 {
			formatString(out, indent+"Display Name", p.DisplayName)
		}
		if len(p.Description) > 0 {
			formatString(out, indent+"Description", p.Description)
		}
		formatString(out, indent+"Required", p.Required)
		if len(p.Generate) == 0 {
			formatString(out, indent+"Value", p.Value)
			continue
		}
		if len(p.Value) > 0 {
			formatString(out, indent+"Value", p.Value)
			formatString(out, indent+"Generated (ignored)", p.Generate)
			formatString(out, indent+"From", p.From)
		} else {
			formatString(out, indent+"Generated", p.Generate)
			formatString(out, indent+"From", p.From)
		}
		out.Write([]byte("\n"))
	}
}

// describeObjects prints out information about the objects of a template
func (d *TemplateDescriber) describeObjects(objects []runtime.Object, out *tabwriter.Writer) {
	formatString(out, "Objects", " ")
	indent := "    "
	for _, obj := range objects {
		if d.ObjectDescriber != nil {
			output, err := d.DescribeObject(obj)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				continue
			}
			fmt.Fprint(out, output)
			fmt.Fprint(out, "\n")
			continue
		}

		meta := kapi.ObjectMeta{}
		meta.Name, _ = d.MetadataAccessor.Name(obj)
		gvk, _, err := d.ObjectTyper.ObjectKinds(obj)
		if err != nil {
			fmt.Fprintf(out, fmt.Sprintf("%s%s\t%s\n", indent, "<unknown>", meta.Name))
			continue
		}
		fmt.Fprintf(out, fmt.Sprintf("%s%s\t%s\n", indent, gvk[0].Kind, meta.Name))
		//meta.Annotations, _ = d.MetadataAccessor.Annotations(obj)
		//meta.Labels, _ = d.MetadataAccessor.Labels(obj)
		/*if len(meta.Labels) > 0 {
			formatString(out, indent+"Labels", formatLabels(meta.Labels))
		}
		formatAnnotations(out, meta, indent)*/
	}
}

// Describe returns the description of a template
func (d *TemplateDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.Templates(namespace)
	template, err := c.Get(name)
	if err != nil {
		return "", err
	}
	return d.DescribeTemplate(template)
}

func (d *TemplateDescriber) DescribeTemplate(template *templateapi.Template) (string, error) {
	// TODO: write error?
	_ = runtime.DecodeList(template.Objects, kapi.Codecs.UniversalDecoder(), runtime.UnstructuredJSONScheme)

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, template.ObjectMeta)
		out.Write([]byte("\n"))
		out.Flush()
		d.DescribeParameters(template.Parameters, out)
		out.Write([]byte("\n"))
		formatString(out, "Object Labels", formatLabels(template.ObjectLabels))
		out.Write([]byte("\n"))
		d.DescribeMessage(template.Message, out)
		out.Write([]byte("\n"))
		out.Flush()
		d.describeObjects(template.Objects, out)
		return nil
	})
}

// IdentityDescriber generates information about a user
type IdentityDescriber struct {
	client.Interface
}

// Describe returns the description of an identity
func (d *IdentityDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	userClient := d.Users()
	identityClient := d.Identities()

	identity, err := identityClient.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, identity.ObjectMeta)

		if len(identity.User.Name) == 0 {
			formatString(out, "User Name", identity.User.Name)
			formatString(out, "User UID", identity.User.UID)
		} else {
			resolvedUser, err := userClient.Get(identity.User.Name)

			nameValue := identity.User.Name
			uidValue := string(identity.User.UID)

			if kerrs.IsNotFound(err) {
				nameValue += fmt.Sprintf(" (Error: User does not exist)")
			} else if err != nil {
				nameValue += fmt.Sprintf(" (Error: User lookup failed)")
			} else {
				if !sets.NewString(resolvedUser.Identities...).Has(name) {
					nameValue += fmt.Sprintf(" (Error: User identities do not include %s)", name)
				}
				if resolvedUser.UID != identity.User.UID {
					uidValue += fmt.Sprintf(" (Error: Actual user UID is %s)", string(resolvedUser.UID))
				}
			}

			formatString(out, "User Name", nameValue)
			formatString(out, "User UID", uidValue)
		}
		return nil
	})

}

// UserIdentityMappingDescriber generates information about a user
type UserIdentityMappingDescriber struct {
	client.Interface
}

// Describe returns the description of a userIdentity
func (d *UserIdentityMappingDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.UserIdentityMappings()

	mapping, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, mapping.ObjectMeta)
		formatString(out, "Identity", mapping.Identity.Name)
		formatString(out, "User Name", mapping.User.Name)
		formatString(out, "User UID", mapping.User.UID)
		return nil
	})
}

// UserDescriber generates information about a user
type UserDescriber struct {
	client.Interface
}

// Describe returns the description of a user
func (d *UserDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	userClient := d.Users()
	identityClient := d.Identities()

	user, err := userClient.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, user.ObjectMeta)
		if len(user.FullName) > 0 {
			formatString(out, "Full Name", user.FullName)
		}

		if len(user.Identities) == 0 {
			formatString(out, "Identities", "<none>")
		} else {
			for i, identity := range user.Identities {
				resolvedIdentity, err := identityClient.Get(identity)

				value := identity
				if kerrs.IsNotFound(err) {
					value += fmt.Sprintf(" (Error: Identity does not exist)")
				} else if err != nil {
					value += fmt.Sprintf(" (Error: Identity lookup failed)")
				} else if resolvedIdentity.User.Name != name {
					value += fmt.Sprintf(" (Error: Identity maps to user name '%s')", resolvedIdentity.User.Name)
				} else if resolvedIdentity.User.UID != user.UID {
					value += fmt.Sprintf(" (Error: Identity maps to user UID '%s')", resolvedIdentity.User.UID)
				}

				if i == 0 {
					formatString(out, "Identities", value)
				} else {
					fmt.Fprintf(out, "           \t%s\n", value)
				}
			}
		}
		return nil
	})
}

// GroupDescriber generates information about a group
type GroupDescriber struct {
	c client.GroupInterface
}

// Describe returns the description of a group
func (d *GroupDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	group, err := d.c.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, group.ObjectMeta)

		if len(group.Users) == 0 {
			formatString(out, "Users", "<none>")
		} else {
			for i, user := range group.Users {
				if i == 0 {
					formatString(out, "Users", user)
				} else {
					fmt.Fprintf(out, "           \t%s\n", user)
				}
			}
		}
		return nil
	})
}

// policy describers

// PolicyDescriber generates information about a Project
type PolicyDescriber struct {
	client.Interface
}

// Describe returns the description of a policy
// TODO make something a lot prettier
func (d *PolicyDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.Policies(namespace)
	policy, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribePolicy(policy)
}

func DescribePolicy(policy *authorizationapi.Policy) (string, error) {
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, policy.ObjectMeta)
		formatString(out, "Last Modified", policy.LastModified)

		// using .List() here because I always want the sorted order that it provides
		for _, key := range sets.StringKeySet(policy.Roles).List() {
			role := policy.Roles[key]
			fmt.Fprint(out, key+"\t"+PolicyRuleHeadings+"\n")
			for _, rule := range role.Rules {
				DescribePolicyRule(out, rule, "\t")
			}
		}

		return nil
	})
}

const PolicyRuleHeadings = "Verbs\tNon-Resource URLs\tExtension\tResource Names\tAPI Groups\tResources"

func DescribePolicyRule(out *tabwriter.Writer, rule authorizationapi.PolicyRule, indent string) {
	extensionString := ""
	if rule.AttributeRestrictions != nil {
		extensionString = fmt.Sprintf("%#v", rule.AttributeRestrictions)

		buffer := new(bytes.Buffer)

		printer := NewHumanReadablePrinter(kctl.PrintOptions{NoHeaders: true})
		if err := printer.PrintObj(rule.AttributeRestrictions, buffer); err == nil {
			extensionString = strings.TrimSpace(buffer.String())
		}
	}

	fmt.Fprintf(out, indent+"%v\t%v\t%v\t%v\t%v\t%v\n",
		rule.Verbs.List(),
		rule.NonResourceURLs.List(),
		extensionString,
		rule.ResourceNames.List(),
		rule.APIGroups,
		rule.Resources.List(),
	)
}

// RoleDescriber generates information about a Project
type RoleDescriber struct {
	client.Interface
}

// Describe returns the description of a role
func (d *RoleDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.Roles(namespace)
	role, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribeRole(role)
}

func DescribeRole(role *authorizationapi.Role) (string, error) {
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, role.ObjectMeta)

		fmt.Fprint(out, PolicyRuleHeadings+"\n")
		for _, rule := range role.Rules {
			DescribePolicyRule(out, rule, "")

		}

		return nil
	})
}

// PolicyBindingDescriber generates information about a Project
type PolicyBindingDescriber struct {
	client.Interface
}

// Describe returns the description of a policyBinding
func (d *PolicyBindingDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.PolicyBindings(namespace)
	policyBinding, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribePolicyBinding(policyBinding)
}

func DescribePolicyBinding(policyBinding *authorizationapi.PolicyBinding) (string, error) {

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, policyBinding.ObjectMeta)
		formatString(out, "Last Modified", policyBinding.LastModified)
		formatString(out, "Policy", policyBinding.PolicyRef.Namespace)

		// using .List() here because I always want the sorted order that it provides
		for _, key := range sets.StringKeySet(policyBinding.RoleBindings).List() {
			roleBinding := policyBinding.RoleBindings[key]
			users, groups, sas, others := authorizationapi.SubjectsStrings(roleBinding.Namespace, roleBinding.Subjects)

			formatString(out, "RoleBinding["+key+"]", " ")
			formatString(out, "\tRole", roleBinding.RoleRef.Name)
			formatString(out, "\tUsers", strings.Join(users, ", "))
			formatString(out, "\tGroups", strings.Join(groups, ", "))
			formatString(out, "\tServiceAccounts", strings.Join(sas, ", "))
			formatString(out, "\tSubjects", strings.Join(others, ", "))
		}

		return nil
	})
}

// RoleBindingDescriber generates information about a Project
type RoleBindingDescriber struct {
	client.Interface
}

// Describe returns the description of a roleBinding
func (d *RoleBindingDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.RoleBindings(namespace)
	roleBinding, err := c.Get(name)
	if err != nil {
		return "", err
	}

	var role *authorizationapi.Role
	if len(roleBinding.RoleRef.Namespace) == 0 {
		var clusterRole *authorizationapi.ClusterRole
		clusterRole, err = d.ClusterRoles().Get(roleBinding.RoleRef.Name)
		role = authorizationapi.ToRole(clusterRole)
	} else {
		role, err = d.Roles(roleBinding.RoleRef.Namespace).Get(roleBinding.RoleRef.Name)
	}

	return DescribeRoleBinding(roleBinding, role, err)
}

// DescribeRoleBinding prints out information about a role binding and its associated role
func DescribeRoleBinding(roleBinding *authorizationapi.RoleBinding, role *authorizationapi.Role, err error) (string, error) {
	users, groups, sas, others := authorizationapi.SubjectsStrings(roleBinding.Namespace, roleBinding.Subjects)

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, roleBinding.ObjectMeta)

		formatString(out, "Role", roleBinding.RoleRef.Namespace+"/"+roleBinding.RoleRef.Name)
		formatString(out, "Users", strings.Join(users, ", "))
		formatString(out, "Groups", strings.Join(groups, ", "))
		formatString(out, "ServiceAccounts", strings.Join(sas, ", "))
		formatString(out, "Subjects", strings.Join(others, ", "))

		switch {
		case err != nil:
			formatString(out, "Policy Rules", fmt.Sprintf("error: %v", err))

		case role != nil:
			fmt.Fprint(out, PolicyRuleHeadings+"\n")
			for _, rule := range role.Rules {
				DescribePolicyRule(out, rule, "")
			}

		default:
			formatString(out, "Policy Rules", "<none>")
		}

		return nil
	})
}

// ClusterPolicyDescriber generates information about a Project
type ClusterPolicyDescriber struct {
	client.Interface
}

// Describe returns the description of a policy
// TODO make something a lot prettier
func (d *ClusterPolicyDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ClusterPolicies()
	policy, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribePolicy(authorizationapi.ToPolicy(policy))
}

type ClusterRoleDescriber struct {
	client.Interface
}

// Describe returns the description of a role
func (d *ClusterRoleDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ClusterRoles()
	role, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribeRole(authorizationapi.ToRole(role))
}

// ClusterPolicyBindingDescriber generates information about a Project
type ClusterPolicyBindingDescriber struct {
	client.Interface
}

// Describe returns the description of a policyBinding
func (d *ClusterPolicyBindingDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ClusterPolicyBindings()
	policyBinding, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return DescribePolicyBinding(authorizationapi.ToPolicyBinding(policyBinding))
}

// ClusterRoleBindingDescriber generates information about a Project
type ClusterRoleBindingDescriber struct {
	client.Interface
}

// Describe returns the description of a roleBinding
func (d *ClusterRoleBindingDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.ClusterRoleBindings()
	roleBinding, err := c.Get(name)
	if err != nil {
		return "", err
	}

	role, err := d.ClusterRoles().Get(roleBinding.RoleRef.Name)
	return DescribeRoleBinding(authorizationapi.ToRoleBinding(roleBinding), authorizationapi.ToRole(role), err)
}

func describeBuildTriggerCauses(causes []buildapi.BuildTriggerCause, out *tabwriter.Writer) {
	if causes == nil {
		formatString(out, "\nBuild trigger cause", "<unknown>")
	}

	for _, cause := range causes {
		formatString(out, "\nBuild trigger cause", cause.Message)

		switch {
		case cause.GitHubWebHook != nil:
			squashGitInfo(cause.GitHubWebHook.Revision, out)
			formatString(out, "Secret", cause.GitHubWebHook.Secret)

		case cause.GenericWebHook != nil:
			squashGitInfo(cause.GenericWebHook.Revision, out)
			formatString(out, "Secret", cause.GenericWebHook.Secret)

		case cause.ImageChangeBuild != nil:
			formatString(out, "Image ID", cause.ImageChangeBuild.ImageID)
			formatString(out, "Image Name/Kind", fmt.Sprintf("%s / %s", cause.ImageChangeBuild.FromRef.Name, cause.ImageChangeBuild.FromRef.Kind))
		}
	}
	fmt.Fprintf(out, "\n")
}

func squashGitInfo(sourceRevision *buildapi.SourceRevision, out *tabwriter.Writer) {
	if sourceRevision != nil && sourceRevision.Git != nil {
		rev := sourceRevision.Git
		var commit string
		if len(rev.Commit) > 7 {
			commit = rev.Commit[:7]
		} else {
			commit = rev.Commit
		}
		formatString(out, "Commit", fmt.Sprintf("%s (%s)", commit, rev.Message))
		hasAuthor := len(rev.Author.Name) != 0
		hasCommitter := len(rev.Committer.Name) != 0
		if hasAuthor && hasCommitter {
			if rev.Author.Name == rev.Committer.Name {
				formatString(out, "Author/Committer", rev.Author.Name)
			} else {
				formatString(out, "Author/Committer", fmt.Sprintf("%s / %s", rev.Author.Name, rev.Committer.Name))
			}
		} else if hasAuthor {
			formatString(out, "Author", rev.Author.Name)
		} else if hasCommitter {
			formatString(out, "Committer", rev.Committer.Name)
		}
	}
}

type ClusterQuotaDescriber struct {
	client.Interface
}

func (d *ClusterQuotaDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	quota, err := d.ClusterResourceQuotas().Get(name)
	if err != nil {
		return "", err
	}
	return DescribeClusterQuota(quota)
}

func DescribeClusterQuota(quota *quotaapi.ClusterResourceQuota) (string, error) {
	labelSelector, err := unversioned.LabelSelectorAsSelector(quota.Spec.Selector.LabelSelector)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, quota.ObjectMeta)
		fmt.Fprintf(out, "Label Selector: %s\n", labelSelector)
		fmt.Fprintf(out, "AnnotationSelector: %s\n", quota.Spec.Selector.AnnotationSelector)
		if len(quota.Spec.Quota.Scopes) > 0 {
			scopes := []string{}
			for _, scope := range quota.Spec.Quota.Scopes {
				scopes = append(scopes, string(scope))
			}
			sort.Strings(scopes)
			fmt.Fprintf(out, "Scopes:\t%s\n", strings.Join(scopes, ", "))
		}
		fmt.Fprintf(out, "Resource\tUsed\tHard\n")
		fmt.Fprintf(out, "--------\t----\t----\n")

		resources := []kapi.ResourceName{}
		for resource := range quota.Status.Total.Hard {
			resources = append(resources, resource)
		}
		sort.Sort(kctl.SortableResourceNames(resources))

		msg := "%v\t%v\t%v\n"
		for i := range resources {
			resource := resources[i]
			hardQuantity := quota.Status.Total.Hard[resource]
			usedQuantity := quota.Status.Total.Used[resource]
			fmt.Fprintf(out, msg, resource, usedQuantity.String(), hardQuantity.String())
		}
		return nil
	})
}

type AppliedClusterQuotaDescriber struct {
	client.Interface
}

func (d *AppliedClusterQuotaDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	quota, err := d.AppliedClusterResourceQuotas(namespace).Get(name)
	if err != nil {
		return "", err
	}
	return DescribeClusterQuota(quotaapi.ConvertAppliedClusterResourceQuotaToClusterResourceQuota(quota))
}

type ClusterNetworkDescriber struct {
	client.Interface
}

// Describe returns the description of a ClusterNetwork
func (d *ClusterNetworkDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	cn, err := d.ClusterNetwork().Get(name)
	if err != nil {
		return "", err
	}
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, cn.ObjectMeta)
		formatString(out, "Cluster Network", cn.Network)
		formatString(out, "Host Subnet Length", cn.HostSubnetLength)
		formatString(out, "Service Network", cn.ServiceNetwork)
		formatString(out, "Plugin Name", cn.PluginName)
		return nil
	})
}

type HostSubnetDescriber struct {
	client.Interface
}

// Describe returns the description of a HostSubnet
func (d *HostSubnetDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	hs, err := d.HostSubnets().Get(name)
	if err != nil {
		return "", err
	}
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, hs.ObjectMeta)
		formatString(out, "Node", hs.Host)
		formatString(out, "Node IP", hs.HostIP)
		formatString(out, "Pod Subnet", hs.Subnet)
		return nil
	})
}

type NetNamespaceDescriber struct {
	client.Interface
}

// Describe returns the description of a NetNamespace
func (d *NetNamespaceDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	netns, err := d.NetNamespaces().Get(name)
	if err != nil {
		return "", err
	}
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, netns.ObjectMeta)
		formatString(out, "Name", netns.NetName)
		formatString(out, "ID", netns.NetID)
		return nil
	})
}

type EgressNetworkPolicyDescriber struct {
	osClient client.Interface
}

// Describe returns the description of an EgressNetworkPolicy
func (d *EgressNetworkPolicyDescriber) Describe(namespace, name string, settings kctl.DescriberSettings) (string, error) {
	c := d.osClient.EgressNetworkPolicies(namespace)
	policy, err := c.Get(name)
	if err != nil {
		return "", err
	}
	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, policy.ObjectMeta)
		for _, rule := range policy.Spec.Egress {
			fmt.Fprintf(out, "Rule:\t%s to %s\n", rule.Type, rule.To.CIDRSelector)
		}
		return nil
	})
}
