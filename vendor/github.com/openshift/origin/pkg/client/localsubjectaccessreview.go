package client

import (
	kapierrors "k8s.io/kubernetes/pkg/api/errors"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

type LocalSubjectAccessReviewsImpersonator interface {
	ImpersonateLocalSubjectAccessReviews(namespace, token string) LocalSubjectAccessReviewInterface
}

// LocalSubjectAccessReviewsNamespacer has methods to work with LocalSubjectAccessReview resources in a namespace
type LocalSubjectAccessReviewsNamespacer interface {
	LocalSubjectAccessReviews(namespace string) LocalSubjectAccessReviewInterface
}

// LocalSubjectAccessReviewInterface exposes methods on LocalSubjectAccessReview resources.
type LocalSubjectAccessReviewInterface interface {
	Create(policy *authorizationapi.LocalSubjectAccessReview) (*authorizationapi.SubjectAccessReviewResponse, error)
}

// localSubjectAccessReviews implements LocalSubjectAccessReviewsNamespacer interface
type localSubjectAccessReviews struct {
	r     *Client
	ns    string
	token *string
}

// newImpersonatingLocalSubjectAccessReviews returns a subjectAccessReviews
func newImpersonatingLocalSubjectAccessReviews(c *Client, namespace, token string) *localSubjectAccessReviews {
	return &localSubjectAccessReviews{
		r:     c,
		ns:    namespace,
		token: &token,
	}
}

// newLocalSubjectAccessReviews returns a localSubjectAccessReviews
func newLocalSubjectAccessReviews(c *Client, namespace string) *localSubjectAccessReviews {
	return &localSubjectAccessReviews{
		r:  c,
		ns: namespace,
	}
}

func (c *localSubjectAccessReviews) Create(sar *authorizationapi.LocalSubjectAccessReview) (*authorizationapi.SubjectAccessReviewResponse, error) {
	result := &authorizationapi.SubjectAccessReviewResponse{}

	req, err := overrideAuth(c.token, c.r.Post().Namespace(c.ns).Resource("localSubjectAccessReviews"))
	if err != nil {
		return &authorizationapi.SubjectAccessReviewResponse{}, err
	}

	err = req.Body(sar).Do().Into(result)

	// if we get one of these failures, we may be talking to an older openshift.  In that case, we need to try hitting ns/namespace-name/subjectaccessreview
	if kapierrors.IsForbidden(err) || kapierrors.IsNotFound(err) {
		deprecatedSAR := &authorizationapi.SubjectAccessReview{
			Action: sar.Action,
			User:   sar.User,
			Groups: sar.Groups,
		}
		deprecatedResponse := &authorizationapi.SubjectAccessReviewResponse{}

		deprecatedReq, deprecatedAttemptErr := overrideAuth(c.token, c.r.Post().Namespace(c.ns).Resource("subjectAccessReviews"))
		if deprecatedAttemptErr != nil {
			return &authorizationapi.SubjectAccessReviewResponse{}, deprecatedAttemptErr
		}
		deprecatedAttemptErr = deprecatedReq.Body(deprecatedSAR).Do().Into(deprecatedResponse)
		if deprecatedAttemptErr == nil {
			err = nil
			result = deprecatedResponse
		}
	}

	return result, err
}
