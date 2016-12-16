package client

import (
	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

type SubjectRulesReviewsNamespacer interface {
	SubjectRulesReviews(namespace string) SubjectRulesReviewInterface
}

type SubjectRulesReviewInterface interface {
	Create(*authorizationapi.SubjectRulesReview) (*authorizationapi.SubjectRulesReview, error)
}

type subjectRulesReviews struct {
	r  *Client
	ns string
}

func newSubjectRulesReviews(c *Client, namespace string) *subjectRulesReviews {
	return &subjectRulesReviews{
		r:  c,
		ns: namespace,
	}
}

func (c *subjectRulesReviews) Create(subjectRulesReview *authorizationapi.SubjectRulesReview) (result *authorizationapi.SubjectRulesReview, err error) {
	result = &authorizationapi.SubjectRulesReview{}
	err = c.r.Post().Namespace(c.ns).Resource("subjectRulesReviews").Body(subjectRulesReview).Do().Into(result)

	return
}
