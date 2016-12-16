package client

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/typed/discovery"

	"github.com/openshift/origin/pkg/api/latest"
	"github.com/openshift/origin/pkg/version"
)

// Interface exposes methods on OpenShift resources.
type Interface interface {
	BuildsNamespacer
	BuildConfigsNamespacer
	BuildLogsNamespacer
	ImagesInterfacer
	ImageSignaturesInterfacer
	ImageStreamsNamespacer
	ImageStreamMappingsNamespacer
	ImageStreamTagsNamespacer
	ImageStreamImagesNamespacer
	ImageStreamSecretsNamespacer
	DeploymentConfigsNamespacer
	DeploymentLogsNamespacer
	RoutesNamespacer
	HostSubnetsInterface
	NetNamespacesInterface
	ClusterNetworkingInterface
	EgressNetworkPoliciesNamespacer
	IdentitiesInterface
	UsersInterface
	GroupsInterface
	UserIdentityMappingsInterface
	ProjectsInterface
	ProjectRequestsInterface
	LocalSubjectAccessReviewsImpersonator
	SubjectAccessReviewsImpersonator
	LocalResourceAccessReviewsNamespacer
	ResourceAccessReviews
	SubjectAccessReviews
	LocalSubjectAccessReviewsNamespacer
	SelfSubjectRulesReviewsNamespacer
	SubjectRulesReviewsNamespacer
	TemplatesNamespacer
	TemplateConfigsNamespacer
	OAuthClientsInterface
	OAuthClientAuthorizationsInterface
	OAuthAccessTokensInterface
	OAuthAuthorizeTokensInterface
	PoliciesNamespacer
	PolicyBindingsNamespacer
	RolesNamespacer
	RoleBindingsNamespacer
	ClusterPoliciesInterface
	ClusterPolicyBindingsInterface
	ClusterRolesInterface
	ClusterRoleBindingsInterface
	ClusterResourceQuotasInterface
	AppliedClusterResourceQuotasNamespacer
}

// Builds provides a REST client for Builds
func (c *Client) Builds(namespace string) BuildInterface {
	return newBuilds(c, namespace)
}

// BuildConfigs provides a REST client for BuildConfigs
func (c *Client) BuildConfigs(namespace string) BuildConfigInterface {
	return newBuildConfigs(c, namespace)
}

// BuildLogs provides a REST client for BuildLogs
func (c *Client) BuildLogs(namespace string) BuildLogsInterface {
	return newBuildLogs(c, namespace)
}

// Images provides a REST client for Images
func (c *Client) Images() ImageInterface {
	return newImages(c)
}

// ImageSignatures provides a REST client for ImageSignatures
func (c *Client) ImageSignatures() ImageSignatureInterface {
	return newImageSignatures(c)
}

// ImageStreamImages provides a REST client for retrieving image secrets in a namespace
func (c *Client) ImageStreamSecrets(namespace string) ImageStreamSecretInterface {
	return newImageStreamSecrets(c, namespace)
}

// ImageStreams provides a REST client for ImageStream
func (c *Client) ImageStreams(namespace string) ImageStreamInterface {
	return newImageStreams(c, namespace)
}

// ImageStreamMappings provides a REST client for ImageStreamMapping
func (c *Client) ImageStreamMappings(namespace string) ImageStreamMappingInterface {
	return newImageStreamMappings(c, namespace)
}

// ImageStreamTags provides a REST client for ImageStreamTag
func (c *Client) ImageStreamTags(namespace string) ImageStreamTagInterface {
	return newImageStreamTags(c, namespace)
}

// ImageStreamImages provides a REST client for ImageStreamImage
func (c *Client) ImageStreamImages(namespace string) ImageStreamImageInterface {
	return newImageStreamImages(c, namespace)
}

// DeploymentConfigs provides a REST client for DeploymentConfig
func (c *Client) DeploymentConfigs(namespace string) DeploymentConfigInterface {
	return newDeploymentConfigs(c, namespace)
}

// DeploymentLogs provides a REST client for DeploymentLog
func (c *Client) DeploymentLogs(namespace string) DeploymentLogInterface {
	return newDeploymentLogs(c, namespace)
}

// Routes provides a REST client for Route
func (c *Client) Routes(namespace string) RouteInterface {
	return newRoutes(c, namespace)
}

// HostSubnets provides a REST client for HostSubnet
func (c *Client) HostSubnets() HostSubnetInterface {
	return newHostSubnet(c)
}

// NetNamespaces provides a REST client for NetNamespace
func (c *Client) NetNamespaces() NetNamespaceInterface {
	return newNetNamespace(c)
}

// ClusterNetwork provides a REST client for ClusterNetworking
func (c *Client) ClusterNetwork() ClusterNetworkInterface {
	return newClusterNetwork(c)
}

// EgressNetworkPolicies provides a REST client for EgressNetworkPolicy
func (c *Client) EgressNetworkPolicies(namespace string) EgressNetworkPolicyInterface {
	return newEgressNetworkPolicy(c, namespace)
}

// Users provides a REST client for User
func (c *Client) Users() UserInterface {
	return newUsers(c)
}

// Identities provides a REST client for Identity
func (c *Client) Identities() IdentityInterface {
	return newIdentities(c)
}

// UserIdentityMappings provides a REST client for UserIdentityMapping
func (c *Client) UserIdentityMappings() UserIdentityMappingInterface {
	return newUserIdentityMappings(c)
}

// Groups provides a REST client for Groups
func (c *Client) Groups() GroupInterface {
	return newGroups(c)
}

// Projects provides a REST client for Projects
func (c *Client) Projects() ProjectInterface {
	return newProjects(c)
}

// ProjectRequests provides a REST client for Projects
func (c *Client) ProjectRequests() ProjectRequestInterface {
	return newProjectRequests(c)
}

// TemplateConfigs provides a REST client for TemplateConfig
func (c *Client) TemplateConfigs(namespace string) TemplateConfigInterface {
	return newTemplateConfigs(c, namespace)
}

// Templates provides a REST client for Templates
func (c *Client) Templates(namespace string) TemplateInterface {
	return newTemplates(c, namespace)
}

// Policies provides a REST client for Policies
func (c *Client) Policies(namespace string) PolicyInterface {
	return newPolicies(c, namespace)
}

// PolicyBindings provides a REST client for PolicyBindings
func (c *Client) PolicyBindings(namespace string) PolicyBindingInterface {
	return newPolicyBindings(c, namespace)
}

// Roles provides a REST client for Roles
func (c *Client) Roles(namespace string) RoleInterface {
	return newRoles(c, namespace)
}

// RoleBindings provides a REST client for RoleBindings
func (c *Client) RoleBindings(namespace string) RoleBindingInterface {
	return newRoleBindings(c, namespace)
}

// LocalResourceAccessReviews provides a REST client for LocalResourceAccessReviews
func (c *Client) LocalResourceAccessReviews(namespace string) LocalResourceAccessReviewInterface {
	return newLocalResourceAccessReviews(c, namespace)
}

// ClusterResourceAccessReviews provides a REST client for ClusterResourceAccessReviews
func (c *Client) ResourceAccessReviews() ResourceAccessReviewInterface {
	return newResourceAccessReviews(c)
}

// ImpersonateSubjectAccessReviews provides a REST client for SubjectAccessReviews
func (c *Client) ImpersonateSubjectAccessReviews(token string) SubjectAccessReviewInterface {
	return newImpersonatingSubjectAccessReviews(c, token)
}

// ImpersonateLocalSubjectAccessReviews provides a REST client for SubjectAccessReviews
func (c *Client) ImpersonateLocalSubjectAccessReviews(namespace, token string) LocalSubjectAccessReviewInterface {
	return newImpersonatingLocalSubjectAccessReviews(c, namespace, token)
}

// LocalSubjectAccessReviews provides a REST client for LocalSubjectAccessReviews
func (c *Client) LocalSubjectAccessReviews(namespace string) LocalSubjectAccessReviewInterface {
	return newLocalSubjectAccessReviews(c, namespace)
}

// SubjectAccessReviews provides a REST client for SubjectAccessReviews
func (c *Client) SubjectAccessReviews() SubjectAccessReviewInterface {
	return newSubjectAccessReviews(c)
}

func (c *Client) SelfSubjectRulesReviews(namespace string) SelfSubjectRulesReviewInterface {
	return newSelfSubjectRulesReviews(c, namespace)
}

func (c *Client) SubjectRulesReviews(namespace string) SubjectRulesReviewInterface {
	return newSubjectRulesReviews(c, namespace)
}

func (c *Client) OAuthClients() OAuthClientInterface {
	return newOAuthClients(c)
}

func (c *Client) OAuthClientAuthorizations() OAuthClientAuthorizationInterface {
	return newOAuthClientAuthorizations(c)
}

func (c *Client) OAuthAccessTokens() OAuthAccessTokenInterface {
	return newOAuthAccessTokens(c)
}

func (c *Client) OAuthAuthorizeTokens() OAuthAuthorizeTokenInterface {
	return newOAuthAuthorizeTokens(c)
}

func (c *Client) ClusterPolicies() ClusterPolicyInterface {
	return newClusterPolicies(c)
}

func (c *Client) ClusterPolicyBindings() ClusterPolicyBindingInterface {
	return newClusterPolicyBindings(c)
}

func (c *Client) ClusterRoles() ClusterRoleInterface {
	return newClusterRoles(c)
}

func (c *Client) ClusterRoleBindings() ClusterRoleBindingInterface {
	return newClusterRoleBindings(c)
}

func (c *Client) ClusterResourceQuotas() ClusterResourceQuotaInterface {
	return newClusterResourceQuotas(c)
}

func (c *Client) AppliedClusterResourceQuotas(namespace string) AppliedClusterResourceQuotaInterface {
	return newAppliedClusterResourceQuotas(c, namespace)
}

// Client is an OpenShift client object
type Client struct {
	*restclient.RESTClient
}

// New creates an OpenShift client for the given config. This client works with builds, deployments,
// templates, routes, and images. It allows operations such as list, get, update and delete on these
// objects. An error is returned if the provided configuration is not valid.
func New(c *restclient.Config) (*Client, error) {
	config := *c
	if err := SetOpenShiftDefaults(&config); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

// DiscoveryClient returns a discovery client.
func (c *Client) Discovery() discovery.DiscoveryInterface {
	d := NewDiscoveryClient(c.RESTClient)
	return d
}

// SetOpenShiftDefaults sets the default settings on the passed
// client configuration
func SetOpenShiftDefaults(config *restclient.Config) error {
	if len(config.UserAgent) == 0 {
		config.UserAgent = DefaultOpenShiftUserAgent()
	}
	if config.GroupVersion == nil {
		// Clients default to the preferred code API version
		groupVersionCopy := latest.Version
		config.GroupVersion = &groupVersionCopy
	}
	if config.APIPath == "" {
		config.APIPath = "/oapi"
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = kapi.Codecs
	}
	return nil
}

// NewOrDie creates an OpenShift client and panics if the provided API version is not recognized.
func NewOrDie(c *restclient.Config) *Client {
	client, err := New(c)
	if err != nil {
		panic(err)
	}
	return client
}

// DefaultOpenShiftUserAgent returns the default user agent that clients can use.
func DefaultOpenShiftUserAgent() string {
	commit := version.Get().GitCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if len(commit) == 0 {
		commit = "unknown"
	}
	version := version.Get().GitVersion
	seg := strings.SplitN(version, "-", 2)
	version = seg[0]
	return fmt.Sprintf("%s/%s (%s/%s) openshift/%s", path.Base(os.Args[0]), version, runtime.GOOS, runtime.GOARCH, commit)
}

// IsStatusErrorKind returns true if this error describes the provided kind.
func IsStatusErrorKind(err error, kind string) bool {
	if s, ok := err.(errors.APIStatus); ok {
		if details := s.Status().Details; details != nil {
			return kind == details.Kind
		}
	}
	return false
}
