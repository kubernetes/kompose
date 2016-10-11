package client

import (
	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

type SelfSubjectRulesReviewsNamespacer interface {
	SelfSubjectRulesReviews(namespace string) SelfSubjectRulesReviewInterface
}

type SelfSubjectRulesReviewInterface interface {
	Create(*authorizationapi.SelfSubjectRulesReview) (*authorizationapi.SelfSubjectRulesReview, error)
}

type selfSubjectRulesReviews struct {
	r  *Client
	ns string
}

func newSelfSubjectRulesReviews(c *Client, namespace string) *selfSubjectRulesReviews {
	return &selfSubjectRulesReviews{
		r:  c,
		ns: namespace,
	}
}

func (c *selfSubjectRulesReviews) Create(selfSubjectRulesReview *authorizationapi.SelfSubjectRulesReview) (result *authorizationapi.SelfSubjectRulesReview, err error) {
	result = &authorizationapi.SelfSubjectRulesReview{}
	err = c.r.Post().Namespace(c.ns).Resource("selfSubjectRulesReviews").Body(selfSubjectRulesReview).Do().Into(result)

	return
}
