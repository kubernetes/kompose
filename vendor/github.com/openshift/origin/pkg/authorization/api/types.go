package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kruntime "k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/sets"
)

// Authorization is calculated against
// 1. all deny RoleBinding PolicyRules in the master namespace - short circuit on match
// 2. all allow RoleBinding PolicyRules in the master namespace - short circuit on match
// 3. all deny RoleBinding PolicyRules in the namespace - short circuit on match
// 4. all allow RoleBinding PolicyRules in the namespace - short circuit on match
// 5. deny by default

const (
	// PolicyName is the name of Policy
	PolicyName     = "default"
	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	NonResourceAll = "*"

	ScopesKey           = "authorization.openshift.io/scopes"
	ScopesAllNamespaces = "*"

	UserKind           = "User"
	GroupKind          = "Group"
	ServiceAccountKind = "ServiceAccount"
	SystemUserKind     = "SystemUser"
	SystemGroupKind    = "SystemGroup"

	UserResource           = "users"
	GroupResource          = "groups"
	ServiceAccountResource = "serviceaccounts"
	SystemUserResource     = "systemusers"
	SystemGroupResource    = "systemgroups"
)

// DiscoveryRule is a rule that allows a client to discover the API resources available on this server
var DiscoveryRule = PolicyRule{
	Verbs: sets.NewString("get"),
	NonResourceURLs: sets.NewString(
		// Server version checking
		"/version", "/version/*",

		// API discovery/negotiation
		"/api", "/api/*",
		"/apis", "/apis/*",
		"/oapi", "/oapi/*",
		"/osapi", "/osapi/", // these cannot be removed until we can drop support for pre 3.1 clients
		"/.well-known", "/.well-known/*",
	),
}

// PolicyRule holds information that describes a policy rule, but does not contain information
// about who the rule applies to or which namespace the rule applies to.
type PolicyRule struct {
	// Verbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.  VerbAll represents all kinds.
	Verbs sets.String
	// AttributeRestrictions will vary depending on what the Authorizer/AuthorizationAttributeBuilder pair supports.
	// If the Authorizer does not recognize how to handle the AttributeRestrictions, the Authorizer should report an error.
	AttributeRestrictions kruntime.Object
	// APIGroups is the name of the APIGroup that contains the resources.  If this field is empty, then both kubernetes and origin API groups are assumed.
	// That means that if an action is requested against one of the enumerated resources in either the kubernetes or the origin API group, the request
	// will be allowed
	APIGroups []string
	// Resources is a list of resources this rule applies to.  ResourceAll represents all resources.
	Resources sets.String
	// ResourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.
	ResourceNames sets.String
	// NonResourceURLs is a set of partial urls that a user should have access to.  *s are allowed, but only as the full, final step in the path
	// If an action is not a resource API request, then the URL is split on '/' and is checked against the NonResourceURLs to look for a match.
	NonResourceURLs sets.String
}

// IsPersonalSubjectAccessReview is a marker for PolicyRule.AttributeRestrictions that denotes that subjectaccessreviews on self should be allowed
type IsPersonalSubjectAccessReview struct {
	unversioned.TypeMeta
}

// Role is a logical grouping of PolicyRules that can be referenced as a unit by RoleBindings.
type Role struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// Rules holds all the PolicyRules for this Role
	Rules []PolicyRule
}

// RoleBinding references a Role, but not contain it.  It can reference any Role in the same namespace or in the global namespace.
// It adds who information via Users and Groups and namespace information by which namespace it exists in.  RoleBindings in a given
// namespace only have effect in that namespace (excepting the master namespace which has power in all namespaces).
type RoleBinding struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Subjects hold object references of to authorize with this rule
	Subjects []kapi.ObjectReference

	// RoleRef can only reference the current namespace and the global namespace
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	// Since Policy is a singleton, this is sufficient knowledge to locate a role
	RoleRef kapi.ObjectReference
}

type RolesByName map[string]*Role

// +genclient=true

// Policy is a object that holds all the Roles for a particular namespace.  There is at most
// one Policy document per namespace.
type Policy struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// LastModified is the last time that any part of the Policy was created, updated, or deleted
	LastModified unversioned.Time

	// Roles holds all the Roles held by this Policy, mapped by Role.Name
	Roles RolesByName
}

type RoleBindingsByName map[string]*RoleBinding

// PolicyBinding is a object that holds all the RoleBindings for a particular namespace.  There is
// one PolicyBinding document per referenced Policy namespace
type PolicyBinding struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// LastModified is the last time that any part of the PolicyBinding was created, updated, or deleted
	LastModified unversioned.Time

	// PolicyRef is a reference to the Policy that contains all the Roles that this PolicyBinding's RoleBindings may reference
	PolicyRef kapi.ObjectReference
	// RoleBindings holds all the RoleBindings held by this PolicyBinding, mapped by RoleBinding.Name
	RoleBindings RoleBindingsByName
}

// SelfSubjectRulesReview is a resource you can create to determine which actions you can perform in a namespace
type SelfSubjectRulesReview struct {
	unversioned.TypeMeta

	// Spec adds information about how to conduct the check
	Spec SelfSubjectRulesReviewSpec

	// Status is completed by the server to tell which permissions you have
	Status SubjectRulesReviewStatus
}

// SelfSubjectRulesReviewSpec adds information about how to conduct the check
type SelfSubjectRulesReviewSpec struct {
	// Scopes to use for the evaluation.  Empty means "use the unscoped (full) permissions of the user/groups".
	// Nil for a self-SubjectRulesReview, means "use the scopes on this request".
	// Nil for a regular SubjectRulesReview, means the same as empty.
	Scopes []string
}

// SubjectRulesReview is a resource you can create to determine which actions another user can perform in a namespace
type SubjectRulesReview struct {
	unversioned.TypeMeta

	// Spec adds information about how to conduct the check
	Spec SubjectRulesReviewSpec

	// Status is completed by the server to tell which permissions you have
	Status SubjectRulesReviewStatus
}

// SubjectRulesReviewSpec adds information about how to conduct the check
type SubjectRulesReviewSpec struct {
	// User is optional.  At least one of User and Groups must be specified.
	User string
	// Groups is optional.  Groups is the list of groups to which the User belongs.  At least one of User and Groups must be specified.
	Groups []string
	// Scopes to use for the evaluation.  Empty means "use the unscoped (full) permissions of the user/groups".
	Scopes []string
}

// SubjectRulesReviewStatus is contains the result of a rules check
type SubjectRulesReviewStatus struct {
	// Rules is the list of rules (no particular sort) that are allowed for the subject
	Rules []PolicyRule
	// EvaluationError can appear in combination with Rules.  It means some error happened during evaluation
	// that may have prevented additional rules from being populated.
	EvaluationError string
}

// ResourceAccessReviewResponse describes who can perform the action
type ResourceAccessReviewResponse struct {
	unversioned.TypeMeta

	// Namespace is the namespace used for the access review
	Namespace string
	// Users is the list of users who can perform the action
	// +k8s:conversion-gen=false
	Users sets.String
	// Groups is the list of groups who can perform the action
	// +k8s:conversion-gen=false
	Groups sets.String

	// EvaluationError is an indication that some error occurred during resolution, but partial results can still be returned.
	// It is entirely possible to get an error and be able to continue determine authorization status in spite of it.  This is
	// most common when a bound role is missing, but enough roles are still present and bound to reason about the request.
	EvaluationError string
}

// ResourceAccessReview is a means to request a list of which users and groups are authorized to perform the
// action specified by spec
type ResourceAccessReview struct {
	unversioned.TypeMeta

	// Action describes the action being tested
	Action
}

// SubjectAccessReviewResponse describes whether or not a user or group can perform an action
type SubjectAccessReviewResponse struct {
	unversioned.TypeMeta

	// Namespace is the namespace used for the access review
	Namespace string
	// Allowed is required.  True if the action would be allowed, false otherwise.
	Allowed bool
	// Reason is optional.  It indicates why a request was allowed or denied.
	Reason string
	// EvaluationError is an indication that some error occurred during the authorization check.
	// It is entirely possible to get an error and be able to continue determine authorization status in spite of it.  This is
	// most common when a bound role is missing, but enough roles are still present and bound to reason about the request.
	EvaluationError string
}

// SubjectAccessReview is an object for requesting information about whether a user or group can perform an action
type SubjectAccessReview struct {
	unversioned.TypeMeta

	// Action describes the action being tested
	Action
	// User is optional.  If both User and Groups are empty, the current authenticated user is used.
	User string
	// Groups is optional.  Groups is the list of groups to which the User belongs.
	// +k8s:conversion-gen=false
	Groups sets.String
	// Scopes to use for the evaluation.  Empty means "use the unscoped (full) permissions of the user/groups".
	// Nil for a self-SAR, means "use the scopes on this request".
	// Nil for a regular SAR, means the same as empty.
	Scopes []string
}

// LocalResourceAccessReview is a means to request a list of which users and groups are authorized to perform the action specified by spec in a particular namespace
type LocalResourceAccessReview struct {
	unversioned.TypeMeta

	// Action describes the action being tested
	Action
}

// LocalSubjectAccessReview is an object for requesting information about whether a user or group can perform an action in a particular namespace
type LocalSubjectAccessReview struct {
	unversioned.TypeMeta

	// Action describes the action being tested.  The Namespace element is FORCED to the current namespace.
	Action
	// User is optional.  If both User and Groups are empty, the current authenticated user is used.
	User string
	// Groups is optional.  Groups is the list of groups to which the User belongs.
	// +k8s:conversion-gen=false
	Groups sets.String
	// Scopes to use for the evaluation.  Empty means "use the unscoped (full) permissions of the user/groups".
	// Nil for a self-SAR, means "use the scopes on this request".
	// Nil for a regular SAR, means the same as empty.
	Scopes []string
}

// Action describes a request to be authorized
type Action struct {
	// Namespace is the namespace of the action being requested.  Currently, there is no distinction between no namespace and all namespaces
	Namespace string
	// Verb is one of: get, list, watch, create, update, delete
	Verb string
	// Group is the API group of the resource
	Group string
	// Version is the API version of the resource
	Version string
	// Resource is one of the existing resource types
	Resource string
	// ResourceName is the name of the resource being requested for a "get" or deleted for a "delete"
	ResourceName string
	// Content is the actual content of the request for create and update
	Content kruntime.Object
}

// PolicyList is a collection of Policies
type PolicyList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of policies
	Items []Policy
}

// PolicyBindingList is a collection of PolicyBindings
type PolicyBindingList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of policyBindings
	Items []PolicyBinding
}

// RoleBindingList is a collection of RoleBindings
type RoleBindingList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of roleBindings
	Items []RoleBinding
}

// RoleList is a collection of Roles
type RoleList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of roles
	Items []Role
}

// ClusterRole is a logical grouping of PolicyRules that can be referenced as a unit by ClusterRoleBindings.
type ClusterRole struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// Rules holds all the PolicyRules for this ClusterRole
	Rules []PolicyRule
}

// ClusterRoleBinding references a ClusterRole, but not contain it.  It can reference any ClusterRole in the same namespace or in the global namespace.
// It adds who information via Users and Groups and namespace information by which namespace it exists in.  ClusterRoleBindings in a given
// namespace only have effect in that namespace (excepting the master namespace which has power in all namespaces).
type ClusterRoleBinding struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// Subjects hold object references of to authorize with this rule
	Subjects []kapi.ObjectReference

	// RoleRef can only reference the current namespace and the global namespace
	// If the ClusterRoleRef cannot be resolved, the Authorizer must return an error.
	// Since Policy is a singleton, this is sufficient knowledge to locate a role
	RoleRef kapi.ObjectReference
}

type ClusterRolesByName map[string]*ClusterRole

// ClusterPolicy is a object that holds all the ClusterRoles for a particular namespace.  There is at most
// one ClusterPolicy document per namespace.
type ClusterPolicy struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// LastModified is the last time that any part of the ClusterPolicy was created, updated, or deleted
	LastModified unversioned.Time

	// Roles holds all the ClusterRoles held by this ClusterPolicy, mapped by Role.Name
	Roles ClusterRolesByName
}

type ClusterRoleBindingsByName map[string]*ClusterRoleBinding

// ClusterPolicyBinding is a object that holds all the ClusterRoleBindings for a particular namespace.  There is
// one ClusterPolicyBinding document per referenced ClusterPolicy namespace
type ClusterPolicyBinding struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// LastModified is the last time that any part of the ClusterPolicyBinding was created, updated, or deleted
	LastModified unversioned.Time

	// ClusterPolicyRef is a reference to the ClusterPolicy that contains all the ClusterRoles that this ClusterPolicyBinding's RoleBindings may reference
	PolicyRef kapi.ObjectReference
	// RoleBindings holds all the RoleBindings held by this ClusterPolicyBinding, mapped by RoleBinding.Name
	RoleBindings ClusterRoleBindingsByName
}

// ClusterPolicyList is a collection of ClusterPolicies
type ClusterPolicyList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of ClusterPolicies
	Items []ClusterPolicy
}

// ClusterPolicyBindingList is a collection of ClusterPolicyBindings
type ClusterPolicyBindingList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of ClusterPolicyBindings
	Items []ClusterPolicyBinding
}

// ClusterRoleBindingList is a collection of ClusterRoleBindings
type ClusterRoleBindingList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of ClusterRoleBindings
	Items []ClusterRoleBinding
}

// ClusterRoleList is a collection of ClusterRoles
type ClusterRoleList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of ClusterRoles
	Items []ClusterRole
}
