/*
Copyright 2016 Skippbox, Ltd All rights reserved.

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

package kubernetes

import (
	"fmt"
	"testing"

	deployapi "github.com/openshift/origin/pkg/deploy/api"

	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/transformer"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func newServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		ContainerName: "name",
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

func newKomposeObject() kobject.KomposeObject {
	return kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": newServiceConfig()},
	}
}

func equalStringSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s1[i] {
			return false
		}
	}
	return true
}

func equalEnv(kEnvs []kobject.EnvVar, k8sEnvs []api.EnvVar) bool {
	if len(kEnvs) != len(k8sEnvs) {
		return false
	}
	for _, kEnv := range kEnvs {
		found := false
		for _, k8sEnv := range k8sEnvs {
			if kEnv.Name == k8sEnv.Name && kEnv.Value == k8sEnv.Value {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalPorts(kPorts []kobject.Ports, k8sPorts []api.ContainerPort) bool {
	if len(kPorts) != len(k8sPorts) {
		return false
	}
	for _, kPort := range kPorts {
		found := false
		for _, k8sPort := range k8sPorts {
			// FIXME: HostPort should be copied to container port
			//if kPort.HostPort == k8sPort.HostPort && kPort.Protocol == k8sPort.Protocol && kPort.ContainerPort == k8sPort.ContainerPort {
			if kPort.Protocol == k8sPort.Protocol && kPort.ContainerPort == k8sPort.ContainerPort {
				found = true
			}
			// Name and HostIp shouldn't be set
			if len(k8sPort.Name) != 0 || len(k8sPort.HostIP) != 0 {
				return false
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalStringMaps(map1 map[string]string, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}
	for k, v := range map1 {
		if map2[k] != v {
			return false
		}
	}
	return true
}

func checkPodTemplate(config kobject.ServiceConfig, template api.PodTemplateSpec, expectedLabels map[string]string) error {
	if len(template.Spec.Containers) == 0 {
		return fmt.Errorf("Failed to set container: %#v vs. %#v", config, template)
	}
	container := template.Spec.Containers[0]
	if config.ContainerName != container.Name {
		return fmt.Errorf("Found different container name: %v vs. %v", config.ContainerName, container.Name)
	}
	if config.Image != container.Image {
		return fmt.Errorf("Found different container image: %v vs. %v", config.Image, container.Image)
	}
	if !equalEnv(config.Environment, container.Env) {
		return fmt.Errorf("Found different container env: %#v vs. %#v", config.Environment, container.Env)
	}
	if !equalPorts(config.Port, container.Ports) {
		return fmt.Errorf("Found different container ports: %#v vs. %#v", config.Port, container.Ports)
	}
	if !equalStringSlice(config.Command, container.Command) {
		return fmt.Errorf("Found different container cmd: %#v vs. %#v", config.Command, container.Command)
	}
	if config.WorkingDir != container.WorkingDir {
		return fmt.Errorf("Found different container WorkingDir: %#v vs. %#v", config.WorkingDir, container.WorkingDir)
	}
	if !equalStringSlice(config.Args, container.Args) {
		return fmt.Errorf("Found different container args: %#v vs. %#v", config.Args, container.Args)
	}
	if len(template.Spec.Volumes) == 0 || len(template.Spec.Volumes[0].Name) == 0 || template.Spec.Volumes[0].VolumeSource.EmptyDir == nil {
		return fmt.Errorf("Found incorrect volumes: %v vs. %#v", config.Volumes, template.Spec.Volumes)
	}
	// We only set controller labels here and k8s server will take care of other defaults, such as selectors
	if !equalStringMaps(expectedLabels, template.Labels) {
		return fmt.Errorf("Found different template labels: %#v vs. %#v", expectedLabels, template.Labels)
	}
	restartPolicyMapping := map[string]api.RestartPolicy{"always": api.RestartPolicyAlways}
	if restartPolicyMapping[config.Restart] != template.Spec.RestartPolicy {
		return fmt.Errorf("Found incorrect restart policy: %v vs. %v", config.Restart, template.Spec.RestartPolicy)
	}
	if config.Privileged == privilegedNilOrFalse(template) {
		return fmt.Errorf("Found different template privileged: %#v vs. %#v", config.Privileged, template.Spec.Containers[0].SecurityContext)
	}
	return nil
}

func privilegedNilOrFalse(template api.PodTemplateSpec) bool {
	return len(template.Spec.Containers) == 0 || template.Spec.Containers[0].SecurityContext == nil ||
		template.Spec.Containers[0].SecurityContext.Privileged == nil || *template.Spec.Containers[0].SecurityContext.Privileged == false
}

func checkService(config kobject.ServiceConfig, svc *api.Service, expectedLabels map[string]string) error {
	if !equalStringMaps(expectedLabels, svc.Spec.Selector) {
		return fmt.Errorf("Found unexpected selector: %#v vs. %#v", expectedLabels, svc.Spec.Selector)
	}
	// TODO: finish this
	return nil
}

func checkMeta(config kobject.ServiceConfig, meta api.ObjectMeta, expectedName string, shouldSetLabels bool) error {
	if expectedName != meta.Name {
		return fmt.Errorf("Found unexpected name: %s vs. %s", expectedName, meta.Name)
	}
	if !equalStringMaps(config.Annotations, meta.Annotations) {
		return fmt.Errorf("Found different annotations: %#v vs. %#v", config.Annotations, meta.Annotations)
	}
	if shouldSetLabels != (len(meta.Labels) > 0) {
		return fmt.Errorf("Unexpected labels: %#v", meta.Labels)
	}
	return nil
}

func TestKomposeConvert(t *testing.T) {
	replicas := 3
	testCases := map[string]struct {
		komposeObject   kobject.KomposeObject
		opt             kobject.ConvertOptions
		expectedNumObjs int
	}{
		"Convert to Deployments (D)":            {newKomposeObject(), kobject.ConvertOptions{CreateD: true, Replicas: replicas}, 2},
		"Convert to DaemonSets (DS)":            {newKomposeObject(), kobject.ConvertOptions{CreateDS: true}, 2},
		"Convert to ReplicationController (RC)": {newKomposeObject(), kobject.ConvertOptions{CreateRC: true, Replicas: replicas}, 2},
		"Convert to D, DS, and RC":              {newKomposeObject(), kobject.ConvertOptions{CreateD: true, CreateDS: true, CreateRC: true, Replicas: replicas}, 4},
		// TODO: add more tests
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		k := Kubernetes{}
		// Run Transform
		objs := k.Transform(test.komposeObject, test.opt)
		if len(objs) != test.expectedNumObjs {
			t.Errorf("Expected %d objects returned, got %d", test.expectedNumObjs, len(objs))
		}

		var foundSVC, foundD, foundDS, foundRC, foundDC bool
		name := "app"
		labels := transformer.ConfigLabels(name)
		config := test.komposeObject.ServiceConfigs[name]
		// Check results
		for _, obj := range objs {
			if svc, ok := obj.(*api.Service); ok {
				if err := checkService(config, svc, labels); err != nil {
					t.Errorf("%v", err)
				}
				if err := checkMeta(config, svc.ObjectMeta, name, true); err != nil {
					t.Errorf("%v", err)
				}
				foundSVC = true
			}
			if test.opt.CreateD {
				if d, ok := obj.(*extensions.Deployment); ok {
					if err := checkPodTemplate(config, d.Spec.Template, labels); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, d.ObjectMeta, name, false); err != nil {
						t.Errorf("%v", err)
					}
					if (int)(d.Spec.Replicas) != replicas {
						t.Errorf("Expected %d replicas, got %d", replicas, d.Spec.Replicas)
					}
					if d.Spec.Selector != nil && len(d.Spec.Selector.MatchLabels) > 0 {
						t.Errorf("Expect selector be unset, got: %#v", d.Spec.Selector)
					}
					foundD = true
				}
			}
			if test.opt.CreateDS {
				if ds, ok := obj.(*extensions.DaemonSet); ok {
					if err := checkPodTemplate(config, ds.Spec.Template, labels); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, ds.ObjectMeta, name, false); err != nil {
						t.Errorf("%v", err)
					}
					if ds.Spec.Selector != nil && len(ds.Spec.Selector.MatchLabels) > 0 {
						t.Errorf("Expect selector be unset, got: %#v", ds.Spec.Selector)
					}
					foundDS = true
				}
			}
			if test.opt.CreateRC {
				if rc, ok := obj.(*api.ReplicationController); ok {
					if err := checkPodTemplate(config, *rc.Spec.Template, labels); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, rc.ObjectMeta, name, false); err != nil {
						t.Errorf("%v", err)
					}
					if (int)(rc.Spec.Replicas) != replicas {
						t.Errorf("Expected %d replicas, got %d", replicas, rc.Spec.Replicas)
					}
					if len(rc.Spec.Selector) > 0 {
						t.Errorf("Expect selector be unset, got: %#v", rc.Spec.Selector)
					}
					foundRC = true
				}
			}
			// TODO: k8s & openshift transformer is now separated; either separate the test or combine the transformer
			if test.opt.CreateDeploymentConfig {
				if dc, ok := obj.(*deployapi.DeploymentConfig); ok {
					if err := checkPodTemplate(config, *dc.Spec.Template, labels); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, dc.ObjectMeta, name, false); err != nil {
						t.Errorf("%v", err)
					}
					if (int)(dc.Spec.Replicas) != replicas {
						t.Errorf("Expected %d replicas, got %d", replicas, dc.Spec.Replicas)
					}
					if len(dc.Spec.Selector) > 0 {
						t.Errorf("Expect selector be unset, got: %#v", dc.Spec.Selector)
					}
					foundDC = true
				}
			}
		}
		if !foundSVC {
			t.Errorf("Unexpected Service not created")
		}
		if test.opt.CreateD != foundD {
			t.Errorf("Expected create Deployment: %v, found Deployment: %v", test.opt.CreateD, foundD)
		}
		if test.opt.CreateDS != foundDS {
			t.Errorf("Expected create Daemon Set: %v, found Daemon Set: %v", test.opt.CreateDS, foundDS)
		}
		if test.opt.CreateRC != foundRC {
			t.Errorf("Expected create Replication Controller: %v, found Replication Controller: %v", test.opt.CreateRC, foundRC)
		}
		if test.opt.CreateDeploymentConfig != foundDC {
			t.Errorf("Expected create Deployment Config: %v, found Deployment Config: %v", test.opt.CreateDeploymentConfig, foundDC)
		}
	}
}
