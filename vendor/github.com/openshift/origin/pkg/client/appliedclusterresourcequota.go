package client

import (
	kapi "k8s.io/kubernetes/pkg/api"

	quotaapi "github.com/openshift/origin/pkg/quota/api"
)

// AppliedClusterResourceQuotasNamespacer has methods to work with AppliedClusterResourceQuota resources in a namespace
type AppliedClusterResourceQuotasNamespacer interface {
	AppliedClusterResourceQuotas(namespace string) AppliedClusterResourceQuotaInterface
}

// AppliedClusterResourceQuotaInterface exposes methods on AppliedClusterResourceQuota resources.
type AppliedClusterResourceQuotaInterface interface {
	List(opts kapi.ListOptions) (*quotaapi.AppliedClusterResourceQuotaList, error)
	Get(name string) (*quotaapi.AppliedClusterResourceQuota, error)
}

// appliedClusterResourceQuotas implements AppliedClusterResourceQuotasNamespacer interface
type appliedClusterResourceQuotas struct {
	r  *Client
	ns string
}

// newAppliedClusterResourceQuotas returns a appliedClusterResourceQuotas
func newAppliedClusterResourceQuotas(c *Client, namespace string) *appliedClusterResourceQuotas {
	return &appliedClusterResourceQuotas{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of appliedClusterResourceQuotas that match the label and field selectors.
func (c *appliedClusterResourceQuotas) List(opts kapi.ListOptions) (result *quotaapi.AppliedClusterResourceQuotaList, err error) {
	result = &quotaapi.AppliedClusterResourceQuotaList{}
	err = c.r.Get().Namespace(c.ns).Resource("appliedclusterresourcequotas").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

// Get returns information about a particular appliedClusterResourceQuota and error if one occurs.
func (c *appliedClusterResourceQuotas) Get(name string) (result *quotaapi.AppliedClusterResourceQuota, err error) {
	result = &quotaapi.AppliedClusterResourceQuota{}
	err = c.r.Get().Namespace(c.ns).Resource("appliedclusterresourcequotas").Name(name).Do().Into(result)
	return
}
