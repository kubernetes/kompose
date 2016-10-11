package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// +genclient=true

// PodSecurityPolicySubjectReview checks whether a particular user/SA tuple can create the PodTemplateSpec.
type PodSecurityPolicySubjectReview struct {
	unversioned.TypeMeta

	// Spec defines specification for the PodSecurityPolicySubjectReview.
	Spec PodSecurityPolicySubjectReviewSpec

	// Status represents the current information/status for the PodSecurityPolicySubjectReview.
	Status PodSecurityPolicySubjectReviewStatus
}

// PodSecurityPolicySubjectReviewSpec defines specification for PodSecurityPolicySubjectReview
type PodSecurityPolicySubjectReviewSpec struct {
	// Template is the PodTemplateSpec to check. If PodTemplateSpec.Spec.ServiceAccountName is empty it will not be defaulted.
	// If its non-empty, it will be checked.
	Template kapi.PodTemplateSpec

	// User is the user you're testing for.
	// If you specify "User" but not "Group", then is it interpreted as "What if User were not a member of any groups.
	// If User and Groups are empty, then the check is performed using *only* the ServiceAccountName in the PodTemplateSpec.
	User string

	// Groups is the groups you're testing for.
	Groups []string
}

// PodSecurityPolicySubjectReviewStatus contains information/status for PodSecurityPolicySubjectReview.
type PodSecurityPolicySubjectReviewStatus struct {
	// AllowedBy is a reference to the rule that allows the PodTemplateSpec.
	// A rule can be a SecurityContextConstraint or a PodSecurityPolicy
	// A `nil`, indicates that it was denied.
	AllowedBy *kapi.ObjectReference

	// A machine-readable description of why this operation is in the
	// "Failure" status. If this value is empty there
	// is no information available.
	Reason string

	// Template is the PodTemplateSpec after the defaulting is applied.
	Template kapi.PodTemplateSpec
}

// PodSecurityPolicySelfSubjectReview checks whether this user/SA tuple can create the PodTemplateSpec.
type PodSecurityPolicySelfSubjectReview struct {
	unversioned.TypeMeta

	// Spec defines specification the PodSecurityPolicySelfSubjectReview.
	Spec PodSecurityPolicySelfSubjectReviewSpec

	// Status represents the current information/status for the PodSecurityPolicySelfSubjectReview.
	Status PodSecurityPolicySubjectReviewStatus
}

// PodSecurityPolicySelfSubjectReviewSpec contains specification for PodSecurityPolicySelfSubjectReview.
type PodSecurityPolicySelfSubjectReviewSpec struct {
	// Template is the PodTemplateSpec to check.
	Template kapi.PodTemplateSpec
}

// PodSecurityPolicyReview checks which service accounts (not users, since that would be cluster-wide) can create the `PodTemplateSpec` in question.
type PodSecurityPolicyReview struct {
	unversioned.TypeMeta

	// Spec is the PodSecurityPolicy to check.
	Spec PodSecurityPolicyReviewSpec

	// Status represents the current information/status for the PodSecurityPolicyReview.
	Status PodSecurityPolicyReviewStatus
}

// PodSecurityPolicyReviewSpec defines specification for PodSecurityPolicyReview
type PodSecurityPolicyReviewSpec struct {
	// Template is the PodTemplateSpec to check. The PodTemplateSpec.Spec.ServiceAccountName field is used
	// if ServiceAccountNames is empty, unless the PodTemplateSpec.Spec.ServiceAccountName is empty,
	// in which case "default" is used.
	// If ServiceAccountNames is specified, PodTemplateSpec.Spec.ServiceAccountName is ignored.
	Template kapi.PodTemplateSpec

	// ServiceAccountNames is an optional set of ServiceAccounts to run the check with.
	// If ServiceAccountNames is empty, the PodTemplateSpec.Spec.ServiceAccountName is used,
	// unless it's empty, in which case "default" is used instead.
	// If ServiceAccountNames is specified, PodTemplateSpec.Spec.ServiceAccountName is ignored.
	ServiceAccountNames []string // TODO: find a way to express 'all service accounts'
}

// PodSecurityPolicyReviewStatus represents the status of PodSecurityPolicyReview.
type PodSecurityPolicyReviewStatus struct {
	// AllowedServiceAccounts returns the list of service accounts in *this* namespace that have the power to create the PodTemplateSpec.
	AllowedServiceAccounts []ServiceAccountPodSecurityPolicyReviewStatus
}

// ServiceAccountPodSecurityPolicyReviewStatus represents ServiceAccount name and related review status
type ServiceAccountPodSecurityPolicyReviewStatus struct {
	PodSecurityPolicySubjectReviewStatus

	// Name contains the allowed and the denied ServiceAccount name
	Name string
}
