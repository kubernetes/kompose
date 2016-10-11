package client

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	sdnapi "github.com/openshift/origin/pkg/sdn/api"
)

// EgressNetworkPoliciesNamespacer has methods to work with EgressNetworkPolicy resources in a namespace
type EgressNetworkPoliciesNamespacer interface {
	EgressNetworkPolicies(namespace string) EgressNetworkPolicyInterface
}

// EgressNetworkPolicyInterface exposes methods on egressNetworkPolicy resources.
type EgressNetworkPolicyInterface interface {
	List(opts kapi.ListOptions) (*sdnapi.EgressNetworkPolicyList, error)
	Get(name string) (*sdnapi.EgressNetworkPolicy, error)
	Create(sub *sdnapi.EgressNetworkPolicy) (*sdnapi.EgressNetworkPolicy, error)
	Update(sub *sdnapi.EgressNetworkPolicy) (*sdnapi.EgressNetworkPolicy, error)
	Delete(name string) error
	Watch(opts kapi.ListOptions) (watch.Interface, error)
}

// egressNetworkPolicy implements EgressNetworkPolicyInterface interface
type egressNetworkPolicy struct {
	r  *Client
	ns string
}

// newEgressNetworkPolicy returns a egressNetworkPolicy
func newEgressNetworkPolicy(c *Client, namespace string) *egressNetworkPolicy {
	return &egressNetworkPolicy{
		r:  c,
		ns: namespace,
	}
}

// List returns a list of EgressNetworkPolicy that match the label and field selectors.
func (c *egressNetworkPolicy) List(opts kapi.ListOptions) (result *sdnapi.EgressNetworkPolicyList, err error) {
	result = &sdnapi.EgressNetworkPolicyList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("egressNetworkPolicies").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular firewall
func (c *egressNetworkPolicy) Get(name string) (result *sdnapi.EgressNetworkPolicy, err error) {
	result = &sdnapi.EgressNetworkPolicy{}
	err = c.r.Get().Namespace(c.ns).Resource("egressNetworkPolicies").Name(name).Do().Into(result)
	return
}

// Create creates a new EgressNetworkPolicy. Returns the server's representation of EgressNetworkPolicy and error if one occurs.
func (c *egressNetworkPolicy) Create(fw *sdnapi.EgressNetworkPolicy) (result *sdnapi.EgressNetworkPolicy, err error) {
	result = &sdnapi.EgressNetworkPolicy{}
	err = c.r.Post().Namespace(c.ns).Resource("egressNetworkPolicies").Body(fw).Do().Into(result)
	return
}

// Update updates the EgressNetworkPolicy on the server. Returns the server's representation of the EgressNetworkPolicy and error if one occurs.
func (c *egressNetworkPolicy) Update(fw *sdnapi.EgressNetworkPolicy) (result *sdnapi.EgressNetworkPolicy, err error) {
	result = &sdnapi.EgressNetworkPolicy{}
	err = c.r.Put().Namespace(c.ns).Resource("egressNetworkPolicies").Name(fw.Name).Body(fw).Do().Into(result)
	return
}

// Delete takes the name of the EgressNetworkPolicy, and returns an error if one occurs during deletion of the EgressNetworkPolicy
func (c *egressNetworkPolicy) Delete(name string) error {
	return c.r.Delete().Namespace(c.ns).Resource("egressNetworkPolicies").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested EgressNetworkPolicies
func (c *egressNetworkPolicy) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("egressNetworkPolicies").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}
