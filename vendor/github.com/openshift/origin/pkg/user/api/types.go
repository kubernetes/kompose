package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// Auth system gets identity name and provider
// POST to UserIdentityMapping, get back error or a filled out UserIdentityMapping object

// +genclient=true

type User struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	FullName string

	Identities []string

	Groups []string
}

type UserList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []User
}

type Identity struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// ProviderName is the source of identity information
	ProviderName string

	// ProviderUserName uniquely represents this identity in the scope of the provider
	ProviderUserName string

	// User is a reference to the user this identity is associated with
	// Both Name and UID must be set
	User kapi.ObjectReference

	Extra map[string]string
}

type IdentityList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []Identity
}

type UserIdentityMapping struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	Identity kapi.ObjectReference
	User     kapi.ObjectReference
}

// Group represents a referenceable set of Users
type Group struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	Users []string
}

type GroupList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []Group
}
