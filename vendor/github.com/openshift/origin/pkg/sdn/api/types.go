package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	ClusterNetworkDefault = "default"
)

// +genclient=true

type ClusterNetwork struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	Network          string
	HostSubnetLength uint32
	ServiceNetwork   string
	PluginName       string
}

type ClusterNetworkList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []ClusterNetwork
}

// HostSubnet encapsulates the inputs needed to define the container subnet network on a node
type HostSubnet struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// host may just be an IP address, resolvable hostname or a complete DNS
	Host   string
	HostIP string
	Subnet string
}

// HostSubnetList is a collection of HostSubnets
type HostSubnetList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []HostSubnet
}

// NetNamespace holds the network id against its name
type NetNamespace struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	NetName string
	NetID   uint32
}

// NetNamespaceList is a collection of NetNamespaces
type NetNamespaceList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []NetNamespace
}
