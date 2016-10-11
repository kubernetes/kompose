package client

import (
	sdnapi "github.com/openshift/origin/pkg/sdn/api"
)

// ClusterNetworkingInterface has methods to work with ClusterNetwork resources
type ClusterNetworkingInterface interface {
	ClusterNetwork() ClusterNetworkInterface
}

// ClusterNetworkInterface exposes methods on clusterNetwork resources.
type ClusterNetworkInterface interface {
	Get(name string) (*sdnapi.ClusterNetwork, error)
	Create(sub *sdnapi.ClusterNetwork) (*sdnapi.ClusterNetwork, error)
	Update(sub *sdnapi.ClusterNetwork) (*sdnapi.ClusterNetwork, error)
}

// clusterNetwork implements ClusterNetworkInterface interface
type clusterNetwork struct {
	r *Client
}

// newClusterNetwork returns a clusterNetwork
func newClusterNetwork(c *Client) *clusterNetwork {
	return &clusterNetwork{
		r: c,
	}
}

// Get returns information about a particular network
func (c *clusterNetwork) Get(networkName string) (result *sdnapi.ClusterNetwork, err error) {
	result = &sdnapi.ClusterNetwork{}
	err = c.r.Get().Resource("clusterNetworks").Name(networkName).Do().Into(result)
	return
}

// Create creates a new ClusterNetwork. Returns the server's representation of ClusterNetwork and error if one occurs.
func (c *clusterNetwork) Create(cn *sdnapi.ClusterNetwork) (result *sdnapi.ClusterNetwork, err error) {
	result = &sdnapi.ClusterNetwork{}
	err = c.r.Post().Resource("clusterNetworks").Body(cn).Do().Into(result)
	return
}

// Update updates the ClusterNetwork on the server. Returns the server's representation of the ClusterNetwork and error if one occurs.
func (c *clusterNetwork) Update(cn *sdnapi.ClusterNetwork) (result *sdnapi.ClusterNetwork, err error) {
	result = &sdnapi.ClusterNetwork{}
	err = c.r.Put().Resource("clusterNetworks").Name(cn.Name).Body(cn).Do().Into(result)
	return
}
