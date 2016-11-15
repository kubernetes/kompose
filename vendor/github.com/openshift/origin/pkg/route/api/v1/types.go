package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// +genclient=true

// A route allows developers to expose services through an HTTP(S) aware load balancing and proxy
// layer via a public DNS entry. The route may further specify TLS options and a certificate, or
// specify a public CNAME that the router should also accept for HTTP and HTTPS traffic. An
// administrator typically configures their router to be visible outside the cluster firewall, and
// may also add additional security, caching, or traffic controls on the service content. Routers
// usually talk directly to the service endpoints.
//
// Once a route is created, the `host` field may not be changed. Generally, routers use the oldest
// route with a given host when resolving conflicts.
//
// Routers are subject to additional customization and may support additional controls via the
// annotations field.
//
// Because administrators may configure multiple routers, the route status field is used to
// return information to clients about the names and states of the route under each router.
// If a client chooses a duplicate name, for instance, the route status conditions are used
// to indicate the route cannot be chosen.
type Route struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec is the desired state of the route
	Spec RouteSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// status is the current state of the route
	Status RouteStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// RouteList is a collection of Routes.
type RouteList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is a list of routes
	Items []Route `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// RouteSpec describes the hostname or path the route exposes, any security information,
// and one or more backends the route points to. Weights on each backend can define
// the balance of traffic sent to each backend - if all weights are zero the route will
// be considered to have no backends and return a standard 503 response.
//
// The `tls` field is optional and allows specific certificates or behavior for the
// route. Routers typically configure a default certificate on a wildcard domain to
// terminate routes without explicit certificates, but custom hostnames usually must
// choose passthrough (send traffic directly to the backend via the TLS Server-Name-
// Indication field) or provide a certificate.
type RouteSpec struct {
	// host is an alias/DNS that points to the service. Optional.
	// If not specified a route name will typically be automatically
	// chosen.
	// Must follow DNS952 subdomain conventions.
	Host string `json:"host" protobuf:"bytes,1,opt,name=host"`
	// Path that the router watches for, to route traffic for to the service. Optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// to is an object the route should use as the primary backend. Only the Service kind
	// is allowed, and it will be defaulted to Service. If the weight field is set to zero,
	// no traffic will be sent to this service.
	To RouteTargetReference `json:"to" protobuf:"bytes,3,opt,name=to"`

	// alternateBackends is an extension of the 'to' field. If more than one service needs to be
	// pointed to, then use this field. Use the weight field in RouteTargetReference object
	// to specify relative preference. If the weight field is zero, the backend is ignored.
	AlternateBackends []RouteTargetReference `json:"alternateBackends,omitempty" protobuf:"bytes,4,rep,name=alternateBackends"`

	// If specified, the port to be used by the router. Most routers will use all
	// endpoints exposed by the service by default - set this value to instruct routers
	// which port to use.
	Port *RoutePort `json:"port,omitempty" protobuf:"bytes,5,opt,name=port"`

	// The tls field provides the ability to configure certificates and termination for the route.
	TLS *TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`

	// Wildcard policy if any for the route.
	// Currently only 'Subdomain' or 'None' is allowed.
	WildcardPolicy WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,7,opt,name=wildcardPolicy"`
}

// RouteTargetReference specifies the target that resolve into endpoints. Only the 'Service'
// kind is allowed. Use 'weight' field to emphasize one over others.
type RouteTargetReference struct {
	// The kind of target that the route is referring to. Currently, only 'Service' is allowed
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`

	// name of the service/target that is being referred to. e.g. name of the service
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// weight as an integer between 1 and 256 that specifies the target's relative weight
	// against other target reference objects
	Weight *int32 `json:"weight" protobuf:"varint,3,opt,name=weight"`
}

// RoutePort defines a port mapping from a router to an endpoint in the service endpoints.
type RoutePort struct {
	// The target port on pods selected by the service this route points to.
	// If this is a string, it will be looked up as a named port in the target
	// endpoints port list. Required
	TargetPort intstr.IntOrString `json:"targetPort" protobuf:"bytes,1,opt,name=targetPort"`
}

// RouteStatus provides relevant info about the status of a route, including which routers
// acknowledge it.
type RouteStatus struct {
	// ingress describes the places where the route may be exposed. The list of
	// ingress points may contain duplicate Host or RouterName values. Routes
	// are considered live once they are `Ready`
	Ingress []RouteIngress `json:"ingress" protobuf:"bytes,1,rep,name=ingress"`
}

// RouteIngress holds information about the places where a route is exposed.
type RouteIngress struct {
	// Host is the host string under which the route is exposed; this value is required
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	// Name is a name chosen by the router to identify itself; this value is required
	RouterName string `json:"routerName,omitempty" protobuf:"bytes,2,opt,name=routerName"`
	// Conditions is the state of the route, may be empty.
	Conditions []RouteIngressCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// Wildcard policy is the wildcard policy that was allowed where this route is exposed.
	WildcardPolicy WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,4,opt,name=wildcardPolicy"`
}

// RouteIngressConditionType is a valid value for RouteCondition
type RouteIngressConditionType string

// These are valid conditions of pod.
const (
	// RouteAdmitted means the route is able to service requests for the provided Host
	RouteAdmitted RouteIngressConditionType = "Admitted"
	// TODO: add other route condition types
)

// RouteIngressCondition contains details for the current condition of this route on a particular
// router.
type RouteIngressCondition struct {
	// Type is the type of the condition.
	// Currently only Ready.
	Type RouteIngressConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RouteIngressConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status kapi.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/kubernetes/pkg/api/v1.ConditionStatus"`
	// (brief) reason for the condition's last transition, and is usually a machine and human
	// readable constant
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// RFC 3339 date and time when this condition last transitioned
	LastTransitionTime *unversioned.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

// RouterShard has information of a routing shard and is used to
// generate host names and routing table entries when a routing shard is
// allocated for a specific route.
// Caveat: This is WIP and will likely undergo modifications when sharding
//         support is added.
type RouterShard struct {
	// shardName uniquely identifies a router shard in the "set" of
	// routers used for routing traffic to the services.
	ShardName string `json:"shardName" protobuf:"bytes,1,opt,name=shardName"`

	// dnsSuffix for the shard ala: shard-1.v3.openshift.com
	DNSSuffix string `json:"dnsSuffix" protobuf:"bytes,2,opt,name=dnsSuffix"`
}

// TLSConfig defines config used to secure a route and provide termination
type TLSConfig struct {
	// termination indicates termination type.
	Termination TLSTerminationType `json:"termination" protobuf:"bytes,1,opt,name=termination,casttype=TLSTerminationType"`

	// certificate provides certificate contents
	Certificate string `json:"certificate,omitempty" protobuf:"bytes,2,opt,name=certificate"`

	// key provides key file contents
	Key string `json:"key,omitempty" protobuf:"bytes,3,opt,name=key"`

	// caCertificate provides the cert authority certificate contents
	CACertificate string `json:"caCertificate,omitempty" protobuf:"bytes,4,opt,name=caCertificate"`

	// destinationCACertificate provides the contents of the ca certificate of the final destination.  When using reencrypt
	// termination this file should be provided in order to have routers use it for health checks on the secure connection
	DestinationCACertificate string `json:"destinationCACertificate,omitempty" protobuf:"bytes,5,opt,name=destinationCACertificate"`

	// insecureEdgeTerminationPolicy indicates the desired behavior for insecure connections to a route. While
	// each router may make its own decisions on which ports to expose, this is normally port 80.
	//
	// * Allow - traffic is sent to the server on the insecure port (default)
	// * Disable - no traffic is allowed on the insecure port.
	// * Redirect - clients are redirected to the secure port.
	InsecureEdgeTerminationPolicy InsecureEdgeTerminationPolicyType `json:"insecureEdgeTerminationPolicy,omitempty" protobuf:"bytes,6,opt,name=insecureEdgeTerminationPolicy,casttype=InsecureEdgeTerminationPolicyType"`
}

// TLSTerminationType dictates where the secure communication will stop
// TODO: Reconsider this type in v2
type TLSTerminationType string

// InsecureEdgeTerminationPolicyType dictates the behavior of insecure
// connections to an edge-terminated route.
type InsecureEdgeTerminationPolicyType string

const (
	// TLSTerminationEdge terminate encryption at the edge router.
	TLSTerminationEdge TLSTerminationType = "edge"
	// TLSTerminationPassthrough terminate encryption at the destination, the destination is responsible for decrypting traffic
	TLSTerminationPassthrough TLSTerminationType = "passthrough"
	// TLSTerminationReencrypt terminate encryption at the edge router and re-encrypt it with a new certificate supplied by the destination
	TLSTerminationReencrypt TLSTerminationType = "reencrypt"
)

// WildcardPolicyType indicates the type of wildcard support needed by routes.
type WildcardPolicyType string

const (
	// WildcardPolicyNone indicates no wildcard support is needed.
	WildcardPolicyNone WildcardPolicyType = "None"

	// WildcardPolicySubdomain indicates the host needs wildcard support for the subdomain.
	// Example: For host = "www.acme.test", indicates that the router
	//          should support requests for *.acme.test
	//          Note that this will not match acme.test only *.acme.test
	WildcardPolicySubdomain WildcardPolicyType = "Subdomain"
)
