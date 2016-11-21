package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type OAuthAccessToken struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// ClientName references the client that created this token.
	ClientName string

	// ExpiresIn is the seconds from CreationTime before this token expires.
	ExpiresIn int64

	// Scopes is an array of the requested scopes.
	Scopes []string

	// RedirectURI is the redirection associated with the token.
	RedirectURI string

	// UserName is the user name associated with this token
	UserName string

	// UserUID is the unique UID associated with this token
	UserUID string

	// AuthorizeToken contains the token that authorized this token
	AuthorizeToken string

	// RefreshToken is the value by which this token can be renewed. Can be blank.
	RefreshToken string
}

type OAuthAuthorizeToken struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// ClientName references the client that created this token.
	ClientName string

	// ExpiresIn is the seconds from CreationTime before this token expires.
	ExpiresIn int64

	// Scopes is an array of the requested scopes.
	Scopes []string

	// RedirectURI is the redirection associated with the token.
	RedirectURI string

	// State data from request
	State string

	// UserName is the user name associated with this token
	UserName string

	// UserUID is the unique UID associated with this token. UserUID and UserName must both match
	// for this token to be valid.
	UserUID string

	// CodeChallenge is the optional code_challenge associated with this authorization code, as described in rfc7636
	CodeChallenge string

	// CodeChallengeMethod is the optional code_challenge_method associated with this authorization code, as described in rfc7636
	CodeChallengeMethod string
}

// +genclient=true

type OAuthClient struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Secret is the unique secret associated with a client
	Secret string

	// AdditionalSecrets holds other secrets that may be used to identify the client.  This is useful for rotation
	// and for service account token validation
	AdditionalSecrets []string

	// RespondWithChallenges indicates whether the client wants authentication needed responses made in the form of challenges instead of redirects
	RespondWithChallenges bool

	// RedirectURIs is the valid redirection URIs associated with a client
	RedirectURIs []string

	// GrantMethod determines how to handle grants for this client. If no method is provided, the
	// cluster default grant handling method will be used
	GrantMethod GrantHandlerType

	// ScopeRestrictions describes which scopes this client can request.  Each requested scope
	// is checked against each restriction.  If any restriction matches, then the scope is allowed.
	// If no restriction matches, then the scope is denied.
	ScopeRestrictions []ScopeRestriction
}

type GrantHandlerType string

const (
	// GrantHandlerAuto auto-approves client authorization grant requests
	GrantHandlerAuto GrantHandlerType = "auto"
	// GrantHandlerPrompt prompts the user to approve new client authorization grant requests
	GrantHandlerPrompt GrantHandlerType = "prompt"
	// GrantHandlerDeny auto-denies client authorization grant requests
	GrantHandlerDeny GrantHandlerType = "deny"
)

// ScopeRestriction describe one restriction on scopes.  Exactly one option must be non-nil.
type ScopeRestriction struct {
	// ExactValues means the scope has to match a particular set of strings exactly
	ExactValues []string

	// ClusterRole describes a set of restrictions for cluster role scoping.
	ClusterRole *ClusterRoleScopeRestriction
}

// ClusterRoleScopeRestriction describes restrictions on cluster role scopes
type ClusterRoleScopeRestriction struct {
	// RoleNames is the list of cluster roles that can referenced.  * means anything
	RoleNames []string
	// Namespaces is the list of namespaces that can be referenced.  * means any of them (including *)
	Namespaces []string
	// AllowEscalation indicates whether you can request roles and their escalating resources
	AllowEscalation bool
}

type OAuthClientAuthorization struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// ClientName references the client that created this authorization
	ClientName string

	// UserName is the user name that authorized this client
	UserName string

	// UserUID is the unique UID associated with this authorization. UserUID and UserName
	// must both match for this authorization to be valid.
	UserUID string

	// Scopes is an array of the granted scopes.
	Scopes []string
}

type OAuthAccessTokenList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []OAuthAccessToken
}

type OAuthAuthorizeTokenList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []OAuthAuthorizeToken
}

type OAuthClientList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []OAuthClient
}

type OAuthClientAuthorizationList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []OAuthClientAuthorization
}

type OAuthRedirectReference struct {
	unversioned.TypeMeta
	kapi.ObjectMeta
	Reference RedirectReference
}

type RedirectReference struct {
	Group string
	Kind  string
	Name  string
}
