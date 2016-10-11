package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	quotaapi "github.com/openshift/origin/pkg/quota/api"
)

type ClusterResourceQuotasInterface interface {
	ClusterResourceQuotas() ClusterResourceQuotaInterface
}

type ClusterResourceQuotaInterface interface {
	List(opts kapi.ListOptions) (*quotaapi.ClusterResourceQuotaList, error)
	Get(name string) (*quotaapi.ClusterResourceQuota, error)
	Create(resourceQuota *quotaapi.ClusterResourceQuota) (*quotaapi.ClusterResourceQuota, error)
	Update(resourceQuota *quotaapi.ClusterResourceQuota) (*quotaapi.ClusterResourceQuota, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)

	UpdateStatus(resourceQuota *quotaapi.ClusterResourceQuota) (*quotaapi.ClusterResourceQuota, error)
}

type clusterResourceQuotas struct {
	r *Client
}

// newClusterResourceQuotas returns a clusterResourceQuotas
func newClusterResourceQuotas(c *Client) *clusterResourceQuotas {
	return &clusterResourceQuotas{
		r: c,
	}
}

func (c *clusterResourceQuotas) List(opts kapi.ListOptions) (result *quotaapi.ClusterResourceQuotaList, err error) {
	result = &quotaapi.ClusterResourceQuotaList{}
	err = c.r.Get().Resource("clusterresourcequotas").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

func (c *clusterResourceQuotas) Get(name string) (result *quotaapi.ClusterResourceQuota, err error) {
	result = &quotaapi.ClusterResourceQuota{}
	err = c.r.Get().Resource("clusterresourcequotas").Name(name).Do().Into(result)
	return
}

func (c *clusterResourceQuotas) Create(resourceQuota *quotaapi.ClusterResourceQuota) (result *quotaapi.ClusterResourceQuota, err error) {
	result = &quotaapi.ClusterResourceQuota{}
	err = c.r.Post().Resource("clusterresourcequotas").Body(resourceQuota).Do().Into(result)
	return
}

func (c *clusterResourceQuotas) Update(resourceQuota *quotaapi.ClusterResourceQuota) (result *quotaapi.ClusterResourceQuota, err error) {
	result = &quotaapi.ClusterResourceQuota{}
	err = c.r.Put().Resource("clusterresourcequotas").Name(resourceQuota.Name).Body(resourceQuota).Do().Into(result)
	return
}

func (c *clusterResourceQuotas) Delete(name string) (err error) {
	err = c.r.Delete().Resource("clusterresourcequotas").Name(name).Do().Error()
	return
}

func (c *clusterResourceQuotas) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().Prefix("watch").Resource("clusterresourcequotas").VersionedParams(&opts, kapi.ParameterCodec).Watch()
}

func (c *clusterResourceQuotas) UpdateStatus(resourceQuota *quotaapi.ClusterResourceQuota) (result *quotaapi.ClusterResourceQuota, err error) {
	result = &quotaapi.ClusterResourceQuota{}
	err = c.r.Put().Resource("clusterresourcequotas").Name(resourceQuota.Name).SubResource("status").Body(resourceQuota).Do().Into(result)
	return
}
