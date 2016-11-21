package clientcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/emicklei/go-restful/swagger"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/typed/discovery"
	"k8s.io/kubernetes/pkg/client/typed/dynamic"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/homedir"

	"github.com/openshift/origin/pkg/api/latest"
	"github.com/openshift/origin/pkg/api/restmapper"
	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	authorizationreaper "github.com/openshift/origin/pkg/authorization/reaper"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildcmd "github.com/openshift/origin/pkg/build/cmd"
	buildutil "github.com/openshift/origin/pkg/build/util"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/describe"
	"github.com/openshift/origin/pkg/cmd/util"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploycmd "github.com/openshift/origin/pkg/deploy/cmd"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
	imageapi "github.com/openshift/origin/pkg/image/api"
	imageutil "github.com/openshift/origin/pkg/image/util"
	routegen "github.com/openshift/origin/pkg/route/generator"
	userapi "github.com/openshift/origin/pkg/user/api"
	authenticationreaper "github.com/openshift/origin/pkg/user/reaper"
)

// New creates a default Factory for commands that should share identical server
// connection behavior. Most commands should use this method to get a factory.
func New(flags *pflag.FlagSet) *Factory {
	// TODO: there should be two client configs, one for OpenShift, and one for Kubernetes
	clientConfig := DefaultClientConfig(flags)
	clientConfig = defaultingClientConfig{clientConfig}
	f := NewFactory(clientConfig)
	f.BindFlags(flags)

	return f
}

// defaultingClientConfig detects whether the provided config is the default, and if
// so returns an error that indicates the user should set up their config.
type defaultingClientConfig struct {
	nested kclientcmd.ClientConfig
}

// RawConfig calls the nested method
func (c defaultingClientConfig) RawConfig() (kclientcmdapi.Config, error) {
	return c.nested.RawConfig()
}

// Namespace calls the nested method, and if an empty config error is returned
// it checks for the same default as kubectl - the value of POD_NAMESPACE or
// "default".
func (c defaultingClientConfig) Namespace() (string, bool, error) {
	namespace, ok, err := c.nested.Namespace()
	if err == nil {
		return namespace, ok, nil
	}
	if !kclientcmd.IsEmptyConfig(err) {
		return "", false, err
	}

	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, true, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, true, nil
		}
	}

	return api.NamespaceDefault, false, nil
}

// ConfigAccess implements ClientConfig
func (c defaultingClientConfig) ConfigAccess() kclientcmd.ConfigAccess {
	return c.nested.ConfigAccess()
}

type errConfigurationMissing struct {
	err error
}

func (e errConfigurationMissing) Error() string {
	return fmt.Sprintf("%v", e.err)
}

func IsConfigurationMissing(err error) bool {
	switch err.(type) {
	case errConfigurationMissing:
		return true
	}
	return kclientcmd.IsContextNotFound(err)
}

// ClientConfig returns a complete client config
func (c defaultingClientConfig) ClientConfig() (*restclient.Config, error) {
	cfg, err := c.nested.ClientConfig()
	if err == nil {
		return cfg, nil
	}

	if !kclientcmd.IsEmptyConfig(err) {
		return nil, err
	}

	// TODO: need to expose inClusterConfig upstream and use that
	if icc, err := restclient.InClusterConfig(); err == nil {
		glog.V(4).Infof("Using in-cluster configuration")
		return icc, nil
	}

	return nil, errConfigurationMissing{fmt.Errorf(`Missing or incomplete configuration info.  Please login or point to an existing, complete config file:

  1. Via the command-line flag --config
  2. Via the KUBECONFIG environment variable
  3. In your home directory as ~/.kube/config

To view or setup config directly use the 'config' command.`)}
}

// Factory provides common options for OpenShift commands
type Factory struct {
	*cmdutil.Factory
	OpenShiftClientConfig kclientcmd.ClientConfig
	clients               *clientCache

	ImageResolutionOptions FlagBinder

	PrintResourceInfos func(*cobra.Command, []*resource.Info, io.Writer) error
}

func DefaultGenerators(cmdName string) map[string]kubectl.Generator {
	generators := map[string]map[string]kubectl.Generator{}
	generators["run"] = map[string]kubectl.Generator{
		"deploymentconfig/v1": deploycmd.BasicDeploymentConfigController{},
		"run-controller/v1":   kubectl.BasicReplicationController{}, // legacy alias for run/v1
	}
	generators["expose"] = map[string]kubectl.Generator{
		"route/v1": routegen.RouteGenerator{},
	}

	return generators[cmdName]
}

// NewFactory creates an object that holds common methods across all OpenShift commands
func NewFactory(clientConfig kclientcmd.ClientConfig) *Factory {
	restMapper := registered.RESTMapper()

	clients := &clientCache{
		clients: make(map[string]*client.Client),
		configs: make(map[string]*restclient.Config),
		loader:  clientConfig,
	}

	w := &Factory{
		Factory:                cmdutil.NewFactory(clientConfig),
		OpenShiftClientConfig:  clientConfig,
		clients:                clients,
		ImageResolutionOptions: &imageResolutionOptions{},
	}

	w.Object = func(bool) (meta.RESTMapper, runtime.ObjectTyper) {
		defaultMapper := ShortcutExpander{RESTMapper: kubectl.ShortcutExpander{RESTMapper: restMapper}}
		defaultTyper := api.Scheme

		// Output using whatever version was negotiated in the client cache. The
		// version we decode with may not be the same as what the server requires.
		cfg, err := clients.ClientConfigForVersion(nil)
		if err != nil {
			return defaultMapper, defaultTyper
		}

		cmdApiVersion := unversioned.GroupVersion{}
		if cfg.GroupVersion != nil {
			cmdApiVersion = *cfg.GroupVersion
		}

		// at this point we've negotiated and can get the client
		oclient, err := clients.ClientForVersion(nil)
		if err != nil {
			return defaultMapper, defaultTyper
		}

		cacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube"), cfg.Host)
		cachedDiscoverClient := NewCachedDiscoveryClient(client.NewDiscoveryClient(oclient.RESTClient), cacheDir, time.Duration(10*time.Minute))

		// if we can't find the server version or its too old to have Kind information in the discovery doc, skip the discovery RESTMapper
		// and use our hardcoded levels
		mapper := registered.RESTMapper()
		if serverVersion, err := cachedDiscoverClient.ServerVersion(); err == nil && useDiscoveryRESTMapper(serverVersion.GitVersion) {
			mapper = restmapper.NewDiscoveryRESTMapper(cachedDiscoverClient)
		}
		mapper = NewShortcutExpander(cachedDiscoverClient, kubectl.ShortcutExpander{RESTMapper: mapper})
		return kubectl.OutputVersionMapper{RESTMapper: mapper, OutputVersions: []unversioned.GroupVersion{cmdApiVersion}}, api.Scheme
	}

	w.UnstructuredObject = func() (meta.RESTMapper, runtime.ObjectTyper, error) {
		// load a discovery client from the default config
		cfg, err := clients.ClientConfigForVersion(nil)
		if err != nil {
			return nil, nil, err
		}
		dc, err := discovery.NewDiscoveryClientForConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		cacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube"), cfg.Host)
		cachedDiscoverClient := NewCachedDiscoveryClient(client.NewDiscoveryClient(dc.RESTClient), cacheDir, time.Duration(10*time.Minute))

		// enumerate all group resources
		groupResources, err := discovery.GetAPIGroupResources(cachedDiscoverClient)
		if err != nil {
			return nil, nil, err
		}

		// Register unknown APIs as third party for now to make
		// validation happy. TODO perhaps make a dynamic schema
		// validator to avoid this.
		for _, group := range groupResources {
			for _, version := range group.Group.Versions {
				gv := unversioned.GroupVersion{Group: group.Group.Name, Version: version.Version}
				if !registered.IsRegisteredVersion(gv) {
					registered.AddThirdPartyAPIGroupVersions(gv)
				}
			}
		}

		// construct unstructured mapper and typer
		mapper := discovery.NewRESTMapper(groupResources, meta.InterfacesForUnstructured)
		typer := discovery.NewUnstructuredObjectTyper(groupResources)
		return NewShortcutExpander(cachedDiscoverClient, kubectl.ShortcutExpander{RESTMapper: mapper}), typer, nil
	}

	kClientForMapping := w.Factory.ClientForMapping
	w.ClientForMapping = func(mapping *meta.RESTMapping) (resource.RESTClient, error) {
		if latest.OriginKind(mapping.GroupVersionKind) {
			mappingVersion := mapping.GroupVersionKind.GroupVersion()
			client, err := clients.ClientForVersion(&mappingVersion)
			if err != nil {
				return nil, err
			}
			return client.RESTClient, nil
		}
		return kClientForMapping(mapping)
	}

	kUnstructuredClientForMapping := w.Factory.UnstructuredClientForMapping
	w.UnstructuredClientForMapping = func(mapping *meta.RESTMapping) (resource.RESTClient, error) {
		if latest.OriginKind(mapping.GroupVersionKind) {
			cfg, err := clientConfig.ClientConfig()
			if err != nil {
				return nil, err
			}
			if err := client.SetOpenShiftDefaults(cfg); err != nil {
				return nil, err
			}
			cfg.APIPath = "/apis"
			if mapping.GroupVersionKind.Group == api.GroupName {
				cfg.APIPath = "/oapi"
			}
			gv := mapping.GroupVersionKind.GroupVersion()
			cfg.ContentConfig = dynamic.ContentConfig()
			cfg.GroupVersion = &gv
			return restclient.RESTClientFor(cfg)
		}
		return kUnstructuredClientForMapping(mapping)
	}

	// Save original Describer function
	kDescriberFunc := w.Factory.Describer
	w.Describer = func(mapping *meta.RESTMapping) (kubectl.Describer, error) {
		if latest.OriginKind(mapping.GroupVersionKind) {
			oClient, kClient, err := w.Clients()
			if err != nil {
				return nil, fmt.Errorf("unable to create client %s: %v", mapping.GroupVersionKind.Kind, err)
			}

			mappingVersion := mapping.GroupVersionKind.GroupVersion()
			cfg, err := clients.ClientConfigForVersion(&mappingVersion)
			if err != nil {
				return nil, fmt.Errorf("unable to load a client %s: %v", mapping.GroupVersionKind.Kind, err)
			}

			describer, ok := describe.DescriberFor(mapping.GroupVersionKind.GroupKind(), oClient, kClient, cfg.Host)
			if !ok {
				return nil, fmt.Errorf("no description has been implemented for %q", mapping.GroupVersionKind.Kind)
			}
			return describer, nil
		}
		return kDescriberFunc(mapping)
	}
	kScalerFunc := w.Factory.Scaler
	w.Scaler = func(mapping *meta.RESTMapping) (kubectl.Scaler, error) {
		if mapping.GroupVersionKind.GroupKind() == deployapi.Kind("DeploymentConfig") {
			oc, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return deploycmd.NewDeploymentConfigScaler(oc, kc), nil
		}
		return kScalerFunc(mapping)
	}
	kReaperFunc := w.Factory.Reaper
	w.Reaper = func(mapping *meta.RESTMapping) (kubectl.Reaper, error) {
		switch mapping.GroupVersionKind.GroupKind() {
		case deployapi.Kind("DeploymentConfig"):
			oc, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return deploycmd.NewDeploymentConfigReaper(oc, kc), nil
		case authorizationapi.Kind("Role"):
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return authorizationreaper.NewRoleReaper(oc, oc), nil
		case authorizationapi.Kind("ClusterRole"):
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return authorizationreaper.NewClusterRoleReaper(oc, oc, oc), nil
		case userapi.Kind("User"):
			oc, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return authenticationreaper.NewUserReaper(
				client.UsersInterface(oc),
				client.GroupsInterface(oc),
				client.ClusterRoleBindingsInterface(oc),
				client.RoleBindingsNamespacer(oc),
				kclient.SecurityContextConstraintsInterface(kc),
			), nil
		case userapi.Kind("Group"):
			oc, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return authenticationreaper.NewGroupReaper(
				client.GroupsInterface(oc),
				client.ClusterRoleBindingsInterface(oc),
				client.RoleBindingsNamespacer(oc),
				kclient.SecurityContextConstraintsInterface(kc),
			), nil
		case buildapi.Kind("BuildConfig"):
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return buildcmd.NewBuildConfigReaper(oc), nil
		}
		return kReaperFunc(mapping)
	}
	kGenerators := w.Factory.Generators
	w.Generators = func(cmdName string) map[string]kubectl.Generator {
		originGenerators := DefaultGenerators(cmdName)
		kubeGenerators := kGenerators(cmdName)

		ret := map[string]kubectl.Generator{}
		for k, v := range kubeGenerators {
			ret[k] = v
		}
		for k, v := range originGenerators {
			ret[k] = v
		}
		return ret
	}
	kMapBasedSelectorForObjectFunc := w.Factory.MapBasedSelectorForObject
	w.MapBasedSelectorForObject = func(object runtime.Object) (string, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			return kubectl.MakeLabels(t.Spec.Selector), nil
		default:
			return kMapBasedSelectorForObjectFunc(object)
		}
	}
	kPortsForObjectFunc := w.Factory.PortsForObject
	w.PortsForObject = func(object runtime.Object) ([]string, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			return getPorts(t.Spec.Template.Spec), nil
		default:
			return kPortsForObjectFunc(object)
		}
	}
	kLogsForObjectFunc := w.Factory.LogsForObject
	w.LogsForObject = func(object, options runtime.Object) (*restclient.Request, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			dopts, ok := options.(*deployapi.DeploymentLogOptions)
			if !ok {
				return nil, errors.New("provided options object is not a DeploymentLogOptions")
			}
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return oc.DeploymentLogs(t.Namespace).Get(t.Name, *dopts), nil
		case *buildapi.Build:
			bopts, ok := options.(*buildapi.BuildLogOptions)
			if !ok {
				return nil, errors.New("provided options object is not a BuildLogOptions")
			}
			if bopts.Version != nil {
				return nil, errors.New("cannot specify a version and a build")
			}
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return oc.BuildLogs(t.Namespace).Get(t.Name, *bopts), nil
		case *buildapi.BuildConfig:
			bopts, ok := options.(*buildapi.BuildLogOptions)
			if !ok {
				return nil, errors.New("provided options object is not a BuildLogOptions")
			}
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			builds, err := oc.Builds(t.Namespace).List(api.ListOptions{})
			if err != nil {
				return nil, err
			}
			builds.Items = buildapi.FilterBuilds(builds.Items, buildapi.ByBuildConfigPredicate(t.Name))
			if len(builds.Items) == 0 {
				return nil, fmt.Errorf("no builds found for %q", t.Name)
			}
			if bopts.Version != nil {
				// If a version has been specified, try to get the logs from that build.
				desired := buildutil.BuildNameForConfigVersion(t.Name, int(*bopts.Version))
				return oc.BuildLogs(t.Namespace).Get(desired, *bopts), nil
			}
			sort.Sort(sort.Reverse(buildapi.BuildSliceByCreationTimestamp(builds.Items)))
			return oc.BuildLogs(t.Namespace).Get(builds.Items[0].Name, *bopts), nil
		default:
			return kLogsForObjectFunc(object, options)
		}
	}
	// Saves current resource name (or alias if any) in PrintOptions. Once saved, it will not be overwritten by the
	// kubernetes resource alias look-up, as it will notice a non-empty value in `options.Kind`
	w.Printer = func(mapping *meta.RESTMapping, options kubectl.PrintOptions) (kubectl.ResourcePrinter, error) {
		if mapping != nil {
			options.Kind = mapping.Resource
			if alias, ok := resourceShortFormFor(mapping.Resource); ok {
				options.Kind = alias
			}
		}
		return describe.NewHumanReadablePrinter(options), nil
	}
	// PrintResourceInfos receives a list of resource infos and prints versioned objects if a generic output format was specified
	// otherwise, it iterates through info objects, printing each resource with a unique printer for its mapping
	w.PrintResourceInfos = func(cmd *cobra.Command, infos []*resource.Info, out io.Writer) error {
		printer, generic, err := cmdutil.PrinterForCommand(cmd)
		if err != nil {
			return nil
		}
		if !generic {
			for _, info := range infos {
				mapping := info.ResourceMapping()
				printer, err := w.PrinterForMapping(cmd, mapping, false)
				if err != nil {
					return err
				}
				if err := printer.PrintObj(info.Object, out); err != nil {
					return nil
				}
			}
			return nil
		}

		clientConfig, err := w.ClientConfig()
		if err != nil {
			return err
		}
		outputVersion, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
		if err != nil {
			return err
		}
		object, err := resource.AsVersionedObject(infos, len(infos) != 1, outputVersion, api.Codecs.LegacyCodec(outputVersion))
		if err != nil {
			return err
		}
		return printer.PrintObj(object, out)

	}
	kCanBeExposed := w.Factory.CanBeExposed
	w.CanBeExposed = func(kind unversioned.GroupKind) error {
		if kind == deployapi.Kind("DeploymentConfig") {
			return nil
		}
		return kCanBeExposed(kind)
	}
	kCanBeAutoscaled := w.Factory.CanBeAutoscaled
	w.CanBeAutoscaled = func(kind unversioned.GroupKind) error {
		if kind == deployapi.Kind("DeploymentConfig") {
			return nil
		}
		return kCanBeAutoscaled(kind)
	}
	kAttachablePodForObjectFunc := w.Factory.AttachablePodForObject
	w.AttachablePodForObject = func(object runtime.Object) (*api.Pod, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			_, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			selector := labels.SelectorFromSet(t.Spec.Selector)
			f := func(pods []*api.Pod) sort.Interface { return sort.Reverse(controller.ActivePods(pods)) }
			pod, _, err := cmdutil.GetFirstPod(kc, t.Namespace, selector, 1*time.Minute, f)
			return pod, err
		default:
			return kAttachablePodForObjectFunc(object)
		}
	}
	kUpdatePodSpecForObject := w.Factory.UpdatePodSpecForObject
	w.UpdatePodSpecForObject = func(obj runtime.Object, fn func(*api.PodSpec) error) (bool, error) {
		switch t := obj.(type) {
		case *deployapi.DeploymentConfig:
			template := t.Spec.Template
			if template == nil {
				t.Spec.Template = template
				template = &api.PodTemplateSpec{}
			}
			return true, fn(&template.Spec)
		default:
			return kUpdatePodSpecForObject(obj, fn)
		}
	}
	kProtocolsForObject := w.Factory.ProtocolsForObject
	w.ProtocolsForObject = func(object runtime.Object) (map[string]string, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			return getProtocols(t.Spec.Template.Spec), nil
		default:
			return kProtocolsForObject(object)
		}
	}

	kSwaggerSchemaFunc := w.Factory.SwaggerSchema
	w.Factory.SwaggerSchema = func(gvk unversioned.GroupVersionKind) (*swagger.ApiDeclaration, error) {
		if !latest.OriginKind(gvk) {
			return kSwaggerSchemaFunc(gvk)
		}
		// TODO: we need to register the OpenShift API under the Kube group, and start returning the OpenShift
		// group from the scheme.
		oc, _, err := w.Clients()
		if err != nil {
			return nil, err
		}
		return w.OriginSwaggerSchema(oc.RESTClient, gvk.GroupVersion())
	}

	w.EditorEnvs = func() []string {
		return []string{"OC_EDITOR", "EDITOR"}
	}
	w.PrintObjectSpecificMessage = func(obj runtime.Object, out io.Writer) {}
	kPauseObjectFunc := w.Factory.PauseObject
	w.Factory.PauseObject = func(object runtime.Object) (bool, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			if t.Spec.Paused {
				return true, nil
			}
			t.Spec.Paused = true
			oc, _, err := w.Clients()
			if err != nil {
				return false, err
			}
			_, err = oc.DeploymentConfigs(t.Namespace).Update(t)
			// TODO: Pause the deployer containers.
			return false, err
		default:
			return kPauseObjectFunc(object)
		}
	}
	kResumeObjectFunc := w.Factory.ResumeObject
	w.Factory.ResumeObject = func(object runtime.Object) (bool, error) {
		switch t := object.(type) {
		case *deployapi.DeploymentConfig:
			if !t.Spec.Paused {
				return true, nil
			}
			t.Spec.Paused = false
			oc, _, err := w.Clients()
			if err != nil {
				return false, err
			}
			_, err = oc.DeploymentConfigs(t.Namespace).Update(t)
			// TODO: Resume the deployer containers.
			return false, err
		default:
			return kResumeObjectFunc(object)
		}
	}
	kResolveImageFunc := w.Factory.ResolveImage
	w.Factory.ResolveImage = func(image string) (string, error) {
		options := w.ImageResolutionOptions.(*imageResolutionOptions)
		if imageutil.IsDocker(options.Source) {
			return kResolveImageFunc(image)
		}
		oc, _, err := w.Clients()
		if err != nil {
			return "", err
		}
		namespace, _, err := w.DefaultNamespace()
		if err != nil {
			return "", err
		}
		return imageutil.ResolveImagePullSpec(oc, oc, options.Source, image, namespace)
	}
	kHistoryViewerFunc := w.Factory.HistoryViewer
	w.Factory.HistoryViewer = func(mapping *meta.RESTMapping) (kubectl.HistoryViewer, error) {
		switch mapping.GroupVersionKind.GroupKind() {
		case deployapi.Kind("DeploymentConfig"):
			oc, kc, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return deploycmd.NewDeploymentConfigHistoryViewer(oc, kc), nil
		}
		return kHistoryViewerFunc(mapping)
	}
	kRollbackerFunc := w.Factory.Rollbacker
	w.Factory.Rollbacker = func(mapping *meta.RESTMapping) (kubectl.Rollbacker, error) {
		switch mapping.GroupVersionKind.GroupKind() {
		case deployapi.Kind("DeploymentConfig"):
			oc, _, err := w.Clients()
			if err != nil {
				return nil, err
			}
			return deploycmd.NewDeploymentConfigRollbacker(oc), nil
		}
		return kRollbackerFunc(mapping)
	}
	kStatusViewerFunc := w.Factory.StatusViewer
	w.Factory.StatusViewer = func(mapping *meta.RESTMapping) (kubectl.StatusViewer, error) {
		oc, _, err := w.Clients()
		if err != nil {
			return nil, err
		}

		switch mapping.GroupVersionKind.GroupKind() {
		case deployapi.Kind("DeploymentConfig"):
			return deploycmd.NewDeploymentConfigStatusViewer(oc), nil
		}
		return kStatusViewerFunc(mapping)
	}

	return w
}

// FlagBinder represents an interface that allows to bind extra flags into commands.
type FlagBinder interface {
	// Bound returns true if the flag is already bound to a command.
	Bound() bool
	// Bind allows to bind an extra flag to a command
	Bind(*pflag.FlagSet)
}

// ImageResolutionOptions provides the "--source" flag to commands that deal with images
// and need to provide extra capabilities for working with ImageStreamTags and
// ImageStreamImages.
type imageResolutionOptions struct {
	bound  bool
	Source string
}

func (o *imageResolutionOptions) Bound() bool {
	return o.bound
}

func (o *imageResolutionOptions) Bind(f *pflag.FlagSet) {
	if o.Bound() {
		return
	}
	f.StringVarP(&o.Source, "source", "", "istag", "The image source type; valid types are valid values are 'imagestreamtag', 'istag', 'imagestreamimage', 'isimage', and 'docker'")
	o.bound = true
}

// useDiscoveryRESTMapper checks the server version to see if its recent enough to have
// enough discovery information avaiable to reliably build a RESTMapper.  If not, use the
// hardcoded mapper in this client (legacy behavior)
func useDiscoveryRESTMapper(serverVersion string) bool {
	serverSemVer, err := semver.Parse(serverVersion[1:])
	if err != nil {
		return false
	}
	if serverSemVer.LT(semver.MustParse("1.3.0")) {
		return false
	}
	return true
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")

	return filepath.Join(parentDir, safeHost)
}

func getPorts(spec api.PodSpec) []string {
	result := []string{}
	for _, container := range spec.Containers {
		for _, port := range container.Ports {
			result = append(result, strconv.Itoa(int(port.ContainerPort)))
		}
	}
	return result
}

func getProtocols(spec api.PodSpec) map[string]string {
	result := make(map[string]string)
	for _, container := range spec.Containers {
		for _, port := range container.Ports {
			result[strconv.Itoa(int(port.ContainerPort))] = string(port.Protocol)
		}
	}
	return result
}

func ResourceMapper(f *Factory) *resource.Mapper {
	mapper, typer := f.Object(false)
	return &resource.Mapper{
		RESTMapper:   mapper,
		ObjectTyper:  typer,
		ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
	}
}

// UpdateObjectEnvironment update the environment variables in object specification.
func (f *Factory) UpdateObjectEnvironment(obj runtime.Object, fn func(*[]api.EnvVar) error) (bool, error) {
	switch t := obj.(type) {
	case *buildapi.BuildConfig:
		if t.Spec.Strategy.CustomStrategy != nil {
			return true, fn(&t.Spec.Strategy.CustomStrategy.Env)
		}
		if t.Spec.Strategy.SourceStrategy != nil {
			return true, fn(&t.Spec.Strategy.SourceStrategy.Env)
		}
		if t.Spec.Strategy.DockerStrategy != nil {
			return true, fn(&t.Spec.Strategy.DockerStrategy.Env)
		}
	}
	return false, fmt.Errorf("object does not contain any environment variables")
}

// ExtractFileContents returns a map of keys to contents, false if the object cannot support such an
// operation, or an error.
func (f *Factory) ExtractFileContents(obj runtime.Object) (map[string][]byte, bool, error) {
	switch t := obj.(type) {
	case *api.Secret:
		return t.Data, true, nil
	case *api.ConfigMap:
		out := make(map[string][]byte)
		for k, v := range t.Data {
			out[k] = []byte(v)
		}
		return out, true, nil
	default:
		return nil, false, nil
	}
}

// ApproximatePodTemplateForObject returns a pod template object for the provided source.
// It may return both an error and a object. It attempt to return the best possible template
// available at the current time.
func (w *Factory) ApproximatePodTemplateForObject(object runtime.Object) (*api.PodTemplateSpec, error) {
	switch t := object.(type) {
	case *imageapi.ImageStreamTag:
		// create a minimal pod spec that uses the image referenced by the istag without any introspection
		// it possible that we could someday do a better job introspecting it
		return &api.PodTemplateSpec{
			Spec: api.PodSpec{
				RestartPolicy: api.RestartPolicyNever,
				Containers: []api.Container{
					{Name: "container-00", Image: t.Image.DockerImageReference},
				},
			},
		}, nil
	case *imageapi.ImageStreamImage:
		// create a minimal pod spec that uses the image referenced by the istag without any introspection
		// it possible that we could someday do a better job introspecting it
		return &api.PodTemplateSpec{
			Spec: api.PodSpec{
				RestartPolicy: api.RestartPolicyNever,
				Containers: []api.Container{
					{Name: "container-00", Image: t.Image.DockerImageReference},
				},
			},
		}, nil
	case *deployapi.DeploymentConfig:
		fallback := t.Spec.Template

		_, kc, err := w.Clients()
		if err != nil {
			return fallback, err
		}

		latestDeploymentName := deployutil.LatestDeploymentNameForConfig(t)
		deployment, err := kc.ReplicationControllers(t.Namespace).Get(latestDeploymentName)
		if err != nil {
			return fallback, err
		}

		fallback = deployment.Spec.Template

		pods, err := kc.Pods(deployment.Namespace).List(api.ListOptions{LabelSelector: labels.SelectorFromSet(deployment.Spec.Selector)})
		if err != nil {
			return fallback, err
		}

		for i := range pods.Items {
			pod := &pods.Items[i]
			if fallback == nil || pod.CreationTimestamp.Before(fallback.CreationTimestamp) {
				fallback = &api.PodTemplateSpec{
					ObjectMeta: pod.ObjectMeta,
					Spec:       pod.Spec,
				}
			}
		}
		return fallback, nil

	default:
		pod, err := w.AttachablePodForObject(object)
		if pod != nil {
			return &api.PodTemplateSpec{
				ObjectMeta: pod.ObjectMeta,
				Spec:       pod.Spec,
			}, err
		}
		switch t := object.(type) {
		case *api.ReplicationController:
			return t.Spec.Template, err
		case *extensions.ReplicaSet:
			return &t.Spec.Template, err
		case *extensions.DaemonSet:
			return &t.Spec.Template, err
		case *batch.Job:
			return &t.Spec.Template, err
		}
		return nil, err
	}
}

func (f *Factory) PodForResource(resource string, timeout time.Duration) (string, error) {
	sortBy := func(pods []*api.Pod) sort.Interface { return sort.Reverse(controller.ActivePods(pods)) }
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return "", err
	}
	mapper, _ := f.Object(false)
	resourceType, name, err := util.ResolveResource(api.Resource("pods"), resource, mapper)
	if err != nil {
		return "", err
	}

	switch resourceType {
	case api.Resource("pods"):
		return name, nil
	case api.Resource("replicationcontrollers"):
		kc, err := f.Client()
		if err != nil {
			return "", err
		}
		rc, err := kc.ReplicationControllers(namespace).Get(name)
		if err != nil {
			return "", err
		}
		selector := labels.SelectorFromSet(rc.Spec.Selector)
		pod, _, err := cmdutil.GetFirstPod(kc, namespace, selector, timeout, sortBy)
		if err != nil {
			return "", err
		}
		return pod.Name, nil
	case deployapi.Resource("deploymentconfigs"):
		oc, kc, err := f.Clients()
		if err != nil {
			return "", err
		}
		dc, err := oc.DeploymentConfigs(namespace).Get(name)
		if err != nil {
			return "", err
		}
		selector := labels.SelectorFromSet(dc.Spec.Selector)
		pod, _, err := cmdutil.GetFirstPod(kc, namespace, selector, timeout, sortBy)
		if err != nil {
			return "", err
		}
		return pod.Name, nil
	case extensions.Resource("daemonsets"):
		kc, err := f.Client()
		if err != nil {
			return "", err
		}
		ds, err := kc.Extensions().DaemonSets(namespace).Get(name)
		if err != nil {
			return "", err
		}
		selector, err := unversioned.LabelSelectorAsSelector(ds.Spec.Selector)
		if err != nil {
			return "", err
		}
		pod, _, err := cmdutil.GetFirstPod(kc, namespace, selector, timeout, sortBy)
		if err != nil {
			return "", err
		}
		return pod.Name, nil
	case extensions.Resource("jobs"):
		kc, err := f.Client()
		if err != nil {
			return "", err
		}
		job, err := kc.Extensions().Jobs(namespace).Get(name)
		if err != nil {
			return "", err
		}
		return podNameForJob(job, kc, timeout, sortBy)
	case batch.Resource("jobs"):
		kc, err := f.Client()
		if err != nil {
			return "", err
		}
		job, err := kc.Batch().Jobs(namespace).Get(name)
		if err != nil {
			return "", err
		}
		return podNameForJob(job, kc, timeout, sortBy)
	default:
		return "", fmt.Errorf("remote shell for %s is not supported", resourceType)
	}
}

func podNameForJob(job *batch.Job, kc *kclient.Client, timeout time.Duration, sortBy func(pods []*api.Pod) sort.Interface) (string, error) {
	selector, err := unversioned.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return "", err
	}
	pod, _, err := cmdutil.GetFirstPod(kc, job.Namespace, selector, timeout, sortBy)
	if err != nil {
		return "", err
	}
	return pod.Name, nil
}

// Clients returns an OpenShift and Kubernetes client.
func (f *Factory) Clients() (*client.Client, *kclient.Client, error) {
	kClient, err := f.Client()
	if err != nil {
		return nil, nil, err
	}
	osClient, err := f.clients.ClientForVersion(nil)
	if err != nil {
		return nil, nil, err
	}
	return osClient, kClient, nil
}

// OriginSwaggerSchema returns a swagger API doc for an Origin schema under the /oapi prefix.
func (f *Factory) OriginSwaggerSchema(client *restclient.RESTClient, version unversioned.GroupVersion) (*swagger.ApiDeclaration, error) {
	if version.Empty() {
		return nil, fmt.Errorf("groupVersion cannot be empty")
	}
	body, err := client.Get().AbsPath("/").Suffix("swaggerapi", "oapi", version.Version).Do().Raw()
	if err != nil {
		return nil, err
	}
	var schema swagger.ApiDeclaration
	err = json.Unmarshal(body, &schema)
	if err != nil {
		return nil, fmt.Errorf("got '%s': %v", string(body), err)
	}
	return &schema, nil
}

// clientCache caches previously loaded clients for reuse. This is largely
// copied from upstream (because of typing) but reuses the negotiation logic.
// TODO: Consolidate this entire concept with upstream's ClientCache.
type clientCache struct {
	loader        kclientcmd.ClientConfig
	clients       map[string]*client.Client
	configs       map[string]*restclient.Config
	defaultConfig *restclient.Config
	// negotiatingClient is used only for negotiating versions with the server.
	negotiatingClient *kclient.Client
}

// ClientConfigForVersion returns the correct config for a server
func (c *clientCache) ClientConfigForVersion(version *unversioned.GroupVersion) (*restclient.Config, error) {
	if c.defaultConfig == nil {
		config, err := c.loader.ClientConfig()
		if err != nil {
			return nil, err
		}
		c.defaultConfig = config
	}
	// TODO: have a better config copy method
	cacheKey := ""
	if version != nil {
		cacheKey = version.String()
	}
	if config, ok := c.configs[cacheKey]; ok {
		return config, nil
	}
	if c.negotiatingClient == nil {
		// TODO: We want to reuse the upstream negotiation logic, which is coupled
		// to a concrete kube Client. The negotiation will ultimately try and
		// build an unversioned URL using the config prefix to ask for supported
		// server versions. If we use the default kube client config, the prefix
		// will be /api, while we need to use the OpenShift prefix to ask for the
		// OpenShift server versions. For now, set OpenShift defaults on the
		// config to ensure the right prefix gets used. The client cache and
		// negotiation logic should be refactored upstream to support downstream
		// reuse so that we don't need to do any of this cache or negotiation
		// duplication.
		negotiatingConfig := *c.defaultConfig
		client.SetOpenShiftDefaults(&negotiatingConfig)
		negotiatingClient, err := kclient.New(&negotiatingConfig)
		if err != nil {
			return nil, err
		}
		c.negotiatingClient = negotiatingClient
	}
	config := *c.defaultConfig
	negotiatedVersion, err := negotiateVersion(c.negotiatingClient, &config, version, latest.Versions)
	if err != nil {
		return nil, err
	}
	config.GroupVersion = negotiatedVersion
	client.SetOpenShiftDefaults(&config)
	c.configs[cacheKey] = &config

	// `version` does not necessarily equal `config.Version`.  However, we know that we call this method again with
	// `config.Version`, we should get the the config we've just built.
	configCopy := config
	c.configs[config.GroupVersion.String()] = &configCopy

	return &config, nil
}

// ClientForVersion initializes or reuses a client for the specified version, or returns an
// error if that is not possible
func (c *clientCache) ClientForVersion(version *unversioned.GroupVersion) (*client.Client, error) {
	cacheKey := ""
	if version != nil {
		cacheKey = version.String()
	}
	if client, ok := c.clients[cacheKey]; ok {
		return client, nil
	}
	config, err := c.ClientConfigForVersion(version)
	if err != nil {
		return nil, err
	}
	client, err := client.New(config)
	if err != nil {
		return nil, err
	}

	c.clients[config.GroupVersion.String()] = client
	return client, nil
}

// FindAllCanonicalResources returns all resource names that map directly to their kind (Kind -> Resource -> Kind)
// and are not subresources. This is the closest mapping possible from the client side to resources that can be
// listed and updated. Note that this may return some virtual resources (like imagestreamtags) that can be otherwise
// represented.
// TODO: add a field to APIResources for "virtual" (or that points to the canonical resource).
// TODO: fallback to the scheme when discovery is not possible.
func FindAllCanonicalResources(d discovery.DiscoveryInterface, m meta.RESTMapper) ([]unversioned.GroupResource, error) {
	set := make(map[unversioned.GroupResource]struct{})
	all, err := d.ServerResources()
	if err != nil {
		return nil, err
	}
	for apiVersion, v := range all {
		gv, err := unversioned.ParseGroupVersion(apiVersion)
		if err != nil {
			continue
		}
		for _, r := range v.APIResources {
			// ignore subresources
			if strings.Contains(r.Name, "/") {
				continue
			}
			// because discovery info doesn't tell us whether the object is virtual or not, perform a lookup
			// by the kind for resource (which should be the canonical resource) and then verify that the reverse
			// lookup (KindsFor) does not error.
			if mapping, err := m.RESTMapping(unversioned.GroupKind{Group: gv.Group, Kind: r.Kind}, gv.Version); err == nil {
				if _, err := m.KindsFor(mapping.GroupVersionKind.GroupVersion().WithResource(mapping.Resource)); err == nil {
					set[unversioned.GroupResource{Group: mapping.GroupVersionKind.Group, Resource: mapping.Resource}] = struct{}{}
				}
			}
		}
	}
	var groupResources []unversioned.GroupResource
	for k := range set {
		groupResources = append(groupResources, k)
	}
	sort.Sort(groupResourcesByName(groupResources))
	return groupResources, nil
}

type groupResourcesByName []unversioned.GroupResource

func (g groupResourcesByName) Len() int { return len(g) }
func (g groupResourcesByName) Less(i, j int) bool {
	if g[i].Resource < g[j].Resource {
		return true
	}
	if g[i].Resource > g[j].Resource {
		return false
	}
	return g[i].Group < g[j].Group
}
func (g groupResourcesByName) Swap(i, j int) { g[i], g[j] = g[j], g[i] }
