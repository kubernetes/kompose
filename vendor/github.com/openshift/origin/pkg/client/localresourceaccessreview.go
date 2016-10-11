package client

import (
	kapierrors "k8s.io/kubernetes/pkg/api/errors"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// LocalResourceAccessReviewsNamespacer has methods to work with LocalResourceAccessReview resources in a namespace
type LocalResourceAccessReviewsNamespacer interface {
	LocalResourceAccessReviews(namespace string) LocalResourceAccessReviewInterface
}

// LocalResourceAccessReviewInterface exposes methods on LocalResourceAccessReview resources.
type LocalResourceAccessReviewInterface interface {
	Create(policy *authorizationapi.LocalResourceAccessReview) (*authorizationapi.ResourceAccessReviewResponse, error)
}

// localResourceAccessReviews implements ResourceAccessReviewsNamespacer interface
type localResourceAccessReviews struct {
	r  *Client
	ns string
}

// newLocalResourceAccessReviews returns a localLocalResourceAccessReviews
func newLocalResourceAccessReviews(c *Client, namespace string) *localResourceAccessReviews {
	return &localResourceAccessReviews{
		r:  c,
		ns: namespace,
	}
}

func (c *localResourceAccessReviews) Create(rar *authorizationapi.LocalResourceAccessReview) (result *authorizationapi.ResourceAccessReviewResponse, err error) {
	result = &authorizationapi.ResourceAccessReviewResponse{}
	err = c.r.Post().Namespace(c.ns).Resource("localResourceAccessReviews").Body(rar).Do().Into(result)

	// if we get one of these failures, we may be talking to an older openshift.  In that case, we need to try hitting ns/namespace-name/subjectaccessreview
	if kapierrors.IsForbidden(err) || kapierrors.IsNotFound(err) {
		deprecatedRAR := &authorizationapi.ResourceAccessReview{
			Action: rar.Action,
		}
		deprecatedResponse := &authorizationapi.ResourceAccessReviewResponse{}
		deprecatedAttemptErr := c.r.Post().Namespace(c.ns).Resource("resourceAccessReviews").Body(deprecatedRAR).Do().Into(deprecatedResponse)
		if deprecatedAttemptErr == nil {
			err = nil
			result = deprecatedResponse
		}
	}

	return
}
