package api

import (
	"k8s.io/kubernetes/pkg/util/sets"
)

// NEVER TOUCH ANYTHING IN THIS FILE!

const (
	// resourceGroupPrefix is the prefix for indicating that a resource entry is actually a group of resources.  The groups are defined in code and indicate resources that are commonly permissioned together
	resourceGroupPrefix = "resourcegroup:"
	buildGroupName      = resourceGroupPrefix + "builds"
	deploymentGroupName = resourceGroupPrefix + "deployments"
	imageGroupName      = resourceGroupPrefix + "images"
	oauthGroupName      = resourceGroupPrefix + "oauth"
	userGroupName       = resourceGroupPrefix + "users"
	templateGroupName   = resourceGroupPrefix + "templates"
	sdnGroupName        = resourceGroupPrefix + "sdn"
	// policyOwnerGroupName includes the physical resources behind the permissionGrantingGroupName.  Unless these physical objects are created first, users with privileges to permissionGrantingGroupName will
	// only be able to bind to global roles
	policyOwnerGroupName = resourceGroupPrefix + "policy"
	// permissionGrantingGroupName includes resources that are necessary to maintain authorization roles and bindings.  By itself, this group is insufficient to create anything except for bindings
	// to master roles.  If a local Policy already exists, then privileges to this group will allow for modification of local roles.
	permissionGrantingGroupName = resourceGroupPrefix + "granter"
	// openshiftExposedGroupName includes resources that are commonly viewed and modified by end users of the system.  It does not include any sensitive resources that control authentication or authorization
	openshiftExposedGroupName = resourceGroupPrefix + "exposedopenshift"
	openshiftAllGroupName     = resourceGroupPrefix + "allopenshift"
	openshiftStatusGroupName  = resourceGroupPrefix + "allopenshift-status"

	quotaGroupName = resourceGroupPrefix + "quota"
	// kubeInternalsGroupName includes those resources that should reasonably be viewable to end users, but that most users should probably not modify.  Kubernetes herself will maintain these resources
	kubeInternalsGroupName = resourceGroupPrefix + "privatekube"
	// kubeExposedGroupName includes resources that are commonly viewed and modified by end users of the system.
	kubeExposedGroupName = resourceGroupPrefix + "exposedkube"
	kubeAllGroupName     = resourceGroupPrefix + "allkube"
	kubeStatusGroupName  = resourceGroupPrefix + "allkube-status"

	// nonescalatingResourcesGroupName contains all resources that can be viewed without exposing the risk of using view rights to locate a secret to escalate privileges.  For example, view
	// rights on secrets could be used locate a secret that happened to be  serviceaccount token that has more privileges
	nonescalatingResourcesGroupName         = resourceGroupPrefix + "non-escalating"
	kubeNonEscalatingViewableGroupName      = resourceGroupPrefix + "kube-non-escalating"
	openshiftNonEscalatingViewableGroupName = resourceGroupPrefix + "openshift-non-escalating"

	// escalatingResourcesGroupName contains all resources that can be used to escalate privileges when simply viewed
	escalatingResourcesGroupName         = resourceGroupPrefix + "escalating"
	kubeEscalatingViewableGroupName      = resourceGroupPrefix + "kube-escalating"
	openshiftEscalatingViewableGroupName = resourceGroupPrefix + "openshift-escalating"
)

var (
	groupsToResources = map[string][]string{
		buildGroupName:       {"builds", "buildconfigs", "buildlogs", "buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/log", "builds/clone", "buildconfigs/webhooks"},
		imageGroupName:       {"imagestreams", "imagestreammappings", "imagestreamtags", "imagestreamimages", "imagestreamimports"},
		deploymentGroupName:  {"deploymentconfigs", "generatedeploymentconfigs", "deploymentconfigrollbacks", "deploymentconfigs/log", "deploymentconfigs/scale"},
		sdnGroupName:         {"clusternetworks", "hostsubnets", "netnamespaces"},
		templateGroupName:    {"templates", "templateconfigs", "processedtemplates"},
		userGroupName:        {"identities", "users", "useridentitymappings", "groups"},
		oauthGroupName:       {"oauthauthorizetokens", "oauthaccesstokens", "oauthclients", "oauthclientauthorizations"},
		policyOwnerGroupName: {"policies", "policybindings"},

		// RAR and SAR are in this list to support backwards compatibility with clients that expect access to those resource in a namespace scope and a cluster scope.
		// TODO remove once we have eliminated the namespace scoped resource.
		permissionGrantingGroupName: {"roles", "rolebindings", "resourceaccessreviews" /* cluster scoped*/, "subjectaccessreviews" /* cluster scoped*/, "localresourceaccessreviews", "localsubjectaccessreviews"},
		openshiftExposedGroupName:   {buildGroupName, imageGroupName, deploymentGroupName, templateGroupName, "routes"},
		openshiftAllGroupName: {openshiftExposedGroupName, userGroupName, oauthGroupName, policyOwnerGroupName, sdnGroupName, permissionGrantingGroupName, openshiftStatusGroupName, "projects",
			"clusterroles", "clusterrolebindings", "clusterpolicies", "clusterpolicybindings", "images" /* cluster scoped*/, "projectrequests", "builds/details", "imagestreams/secrets",
			"selfsubjectrulesreviews"},
		openshiftStatusGroupName: {"imagestreams/status", "routes/status", "deploymentconfigs/status"},

		quotaGroupName:         {"limitranges", "resourcequotas", "resourcequotausages"},
		kubeExposedGroupName:   {"pods", "replicationcontrollers", "serviceaccounts", "services", "endpoints", "persistentvolumeclaims", "pods/log", "configmaps"},
		kubeInternalsGroupName: {"minions", "nodes", "bindings", "events", "namespaces", "persistentvolumes", "securitycontextconstraints"},
		kubeAllGroupName:       {kubeInternalsGroupName, kubeExposedGroupName, quotaGroupName},
		kubeStatusGroupName:    {"pods/status", "resourcequotas/status", "namespaces/status", "replicationcontrollers/status"},

		openshiftEscalatingViewableGroupName: {"oauthauthorizetokens", "oauthaccesstokens", "imagestreams/secrets"},
		kubeEscalatingViewableGroupName:      {"secrets"},
		escalatingResourcesGroupName:         {openshiftEscalatingViewableGroupName, kubeEscalatingViewableGroupName},

		nonescalatingResourcesGroupName: {openshiftNonEscalatingViewableGroupName, kubeNonEscalatingViewableGroupName},
	}
)

func init() {
	// set the non-escalating groups
	groupsToResources[openshiftNonEscalatingViewableGroupName] = NormalizeResources(sets.NewString(groupsToResources[openshiftAllGroupName]...)).
		Difference(NormalizeResources(sets.NewString(groupsToResources[openshiftEscalatingViewableGroupName]...))).List()

	groupsToResources[kubeNonEscalatingViewableGroupName] = NormalizeResources(sets.NewString(groupsToResources[kubeAllGroupName]...)).
		Difference(NormalizeResources(sets.NewString(groupsToResources[kubeEscalatingViewableGroupName]...))).List()
}
