package client

import (
	kapierrors "k8s.io/kubernetes/pkg/api/errors"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// ResourceAccessReviews has methods to work with ResourceAccessReview resources in the cluster scope
type ResourceAccessReviews interface {
	ResourceAccessReviews() ResourceAccessReviewInterface
}

// ResourceAccessReviewInterface exposes methods on ResourceAccessReview resources.
type ResourceAccessReviewInterface interface {
	Create(policy *authorizationapi.ResourceAccessReview) (*authorizationapi.ResourceAccessReviewResponse, error)
}

// resourceAccessReviews implements ResourceAccessReviews interface
type resourceAccessReviews struct {
	r *Client
}

// newResourceAccessReviews returns a resourceAccessReviews
func newResourceAccessReviews(c *Client) *resourceAccessReviews {
	return &resourceAccessReviews{
		r: c,
	}
}

func (c *resourceAccessReviews) Create(rar *authorizationapi.ResourceAccessReview) (result *authorizationapi.ResourceAccessReviewResponse, err error) {
	result = &authorizationapi.ResourceAccessReviewResponse{}

	// if this a cluster RAR, then no special handling
	if len(rar.Action.Namespace) == 0 {
		err = c.r.Post().Resource("resourceAccessReviews").Body(rar).Do().Into(result)
		return
	}

	err = c.r.Post().Resource("resourceAccessReviews").Body(rar).Do().Into(result)

	// if the namespace values don't match then we definitely hit an old server.  If we got a forbidden, then we might have hit an old server
	// and should try the old endpoint
	if (rar.Action.Namespace != result.Namespace) || kapierrors.IsForbidden(err) {
		deprecatedResponse := &authorizationapi.ResourceAccessReviewResponse{}
		deprecatedAttemptErr := c.r.Post().Namespace(rar.Action.Namespace).Resource("resourceAccessReviews").Body(rar).Do().Into(deprecatedResponse)

		// if we definitely hit an old server, then return the error and result you get from the older server.
		if rar.Action.Namespace != result.Namespace {
			return deprecatedResponse, deprecatedAttemptErr
		}

		// if we're not certain it was an old server, success overwrites the previous error, but failure doesn't overwrite the previous error
		if deprecatedAttemptErr == nil {
			err = nil
			result = deprecatedResponse
		}
	}

	return
}
