package api

import (
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
	if route1.CreationTimestamp == route2.CreationTimestamp && route1.UID < route2.UID {
		return true
	}
	if route1.Namespace < route2.Namespace {
		return true
	}
	return route1.Name < route2.Name
}
