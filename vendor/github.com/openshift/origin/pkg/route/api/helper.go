package api

import (
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
)

// IngressConditionStatus returns the first status and condition matching the provided ingress condition type. Conditions
// prefer the first matching entry and clients are allowed to ignore later conditions of the same type.
func IngressConditionStatus(ingress *RouteIngress, t RouteIngressConditionType) (kapi.ConditionStatus, RouteIngressCondition) {
	for _, condition := range ingress.Conditions {
		if t != condition.Type {
			continue
		}
		return condition.Status, condition
	}
	return kapi.ConditionUnknown, RouteIngressCondition{}
}

func RouteLessThan(route1, route2 *Route) bool {
	if route1.CreationTimestamp.Before(route2.CreationTimestamp) {
		return true
	}

	if route2.CreationTimestamp.Before(route1.CreationTimestamp) {
		return false
	}

	return route1.UID < route2.UID
}

// GetDomainForHost returns the domain for the specified host.
// Note for top level domains, this will return an empty string.
func GetDomainForHost(host string) string {
	if idx := strings.IndexRune(host, '.'); idx > -1 {
		return host[idx+1:]
	}

	return ""
}
