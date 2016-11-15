/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openshift

import (
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"k8s.io/kubernetes/pkg/api"
	"testing"
)

func newServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		ContainerName: "myfoobarname",
		Image:         "image",
		Environment:   []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456, Protocol: api.ProtocolTCP}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		Volumes:       []string{"/tmp/volume"},
		Network:       []string{"network1", "network2"}, // not supported
		Labels:        nil,
		Annotations:   map[string]string{"abc": "def"},
		CPUSet:        "cpu_set",            // not supported
		CPUShares:     1,                    // not supported
		CPUQuota:      1,                    // not supported
		CapAdd:        []string{"cap_add"},  // not supported
		CapDrop:       []string{"cap_drop"}, // not supported
		Expose:        []string{"expose"},   // not supported
		Privileged:    true,
		Restart:       "always",
		User:          "user", // not supported
	}
}

func TestInitDeploymentConfig(t *testing.T) {
	o := OpenShift{}
	spec := o.initDeploymentConfig("foobar", newServiceConfig(), 1)

	// Check that "foobar" is used correctly as a name
	if spec.Spec.Template.Spec.Containers[0].Name != "foobar" {
		t.Errorf("Expected foobar for name, actual %s", spec.Spec.Template.Spec.Containers[0].Name)
	}

	// Check that "myfoobarname" is used correctly as a ContainerName
	if spec.Spec.Triggers[1].ImageChangeParams.ContainerNames[0] != "myfoobarname" {
		t.Errorf("Expected myfoobarname for name, actual %s", spec.Spec.Triggers[1].ImageChangeParams.ContainerNames[0])
	}
}

func TestKomposeConvertRoute(t *testing.T) {

	o := OpenShift{}
	name := "app"
	sc := newServiceConfig()
	sc.ExposeService = "true"
	var port int32 = 5555
	route := o.initRoute(name, sc, port)

	if route.ObjectMeta.Name != name {
		t.Errorf("Expected %s for name, actual %s", name, route.ObjectMeta.Name)
	}
	if route.Spec.To.Name != name {
		t.Errorf("Expected %s for name, actual %s", name, route.Spec.To.Name)
	}
	if route.Spec.Port.TargetPort.IntVal != port {
		t.Errorf("Expected %d for port, actual %d", port, route.Spec.Port.TargetPort.IntVal)
	}
	if route.Spec.Host != "" {
		t.Errorf("Expected Spec.Host to not be set, got %s instead", route.Spec.Host)
	}

	sc.ExposeService = "example.com"
	route = o.initRoute(name, sc, port)

	if route.Spec.Host != sc.ExposeService {
		t.Errorf("Expected %s for Spec.Host, actual %s", sc.ExposeService, route.Spec.Host)
	}

}
