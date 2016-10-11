package util

import (
	"fmt"
	"strconv"

	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"

	"github.com/openshift/origin/pkg/route/api"
)

// UnsecuredRoute will return a route with enough info so that it can direct traffic to
// the service provided by --service. Callers of this helper are responsible for providing
// tls configuration, path, and the hostname of the route.
func UnsecuredRoute(kc *kclient.Client, namespace, routeName, serviceName, portString string) (*api.Route, error) {
	if len(routeName) == 0 {
		routeName = serviceName
	}

	svc, err := kc.Services(namespace).Get(serviceName)
	if err != nil {
		if len(portString) == 0 {
			return nil, fmt.Errorf("you need to provide a route port via --port when exposing a non-existent service")
		}
		return &api.Route{
			ObjectMeta: kapi.ObjectMeta{
				Name: routeName,
			},
			Spec: api.RouteSpec{
				To: api.RouteTargetReference{
					Name: serviceName,
				},
				Port: resolveRoutePort(portString),
			},
		}, nil
	}

	ok, port := supportsTCP(svc)
	if !ok {
		return nil, fmt.Errorf("service %q doesn't support TCP", svc.Name)
	}

	route := &api.Route{
		ObjectMeta: kapi.ObjectMeta{
			Name:   routeName,
			Labels: svc.Labels,
		},
		Spec: api.RouteSpec{
			To: api.RouteTargetReference{
				Name: serviceName,
			},
		},
	}

	// If the service has multiple ports and the user didn't specify --port,
	// then default the route port to a service port name.
	if len(port.Name) > 0 && len(portString) == 0 {
		route.Spec.Port = resolveRoutePort(port.Name)
	}
	// --port uber alles
	if len(portString) > 0 {
		route.Spec.Port = resolveRoutePort(portString)
	}

	return route, nil
}

func resolveRoutePort(portString string) *api.RoutePort {
	if len(portString) == 0 {
		return nil
	}
	var routePort intstr.IntOrString
	integer, err := strconv.Atoi(portString)
	if err != nil {
		routePort = intstr.FromString(portString)
	} else {
		routePort = intstr.FromInt(integer)
	}
	return &api.RoutePort{
		TargetPort: routePort,
	}
}

func supportsTCP(svc *kapi.Service) (bool, kapi.ServicePort) {
	for _, port := range svc.Spec.Ports {
		if port.Protocol == kapi.ProtocolTCP {
			return true, port
		}
	}
	return false, kapi.ServicePort{}
}
