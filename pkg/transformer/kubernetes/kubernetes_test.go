/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	dockerCliTypes "github.com/docker/cli/cli/compose/types"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader/compose"
	"github.com/kubernetes/kompose/pkg/transformer"
	deployapi "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		Name:            "app",
		ContainerName:   "name",
		Image:           "image",
		Environment:     []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:            []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456}, kobject.Ports{HostPort: 123, ContainerPort: 456, Protocol: string(api.ProtocolUDP)}},
		Command:         []string{"cmd"},
		WorkingDir:      "dir",
		Args:            []string{"arg1", "arg2"},
		VolList:         []string{"/tmp/volume"},
		Network:         []string{"network1", "network2"}, // supported
		Labels:          nil,
		Annotations:     map[string]string{"abc": "def"},
		CPUQuota:        1, // not supported
		CapAdd:          []string{"cap_add"},
		CapDrop:         []string{"cap_drop"},
		Expose:          []string{"expose"}, // not supported
		Privileged:      true,
		Restart:         "always",
		ImagePullSecret: "regcred",
		Stdin:           true,
		Tty:             true,
		TmpFs:           []string{"/tmp"},
		Replicas:        2,
		Volumes:         []kobject.Volumes{{SvcName: "app", MountPath: "/tmp/volume", PVCName: "app-claim0"}},
		GroupAdd:        []int64{1003, 1005},
		Configs:         []dockerCliTypes.ServiceConfigObjConfig{{Source: "config", Target: "/etc/world"}},
		ConfigsMetaData: map[string]dockerCliTypes.ConfigObjConfig{"config": dockerCliTypes.ConfigObjConfig{Name: "myconfig", File: "kubernetes_test.go"}},
	}
}

func newSimpleServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		Name:          "app",
		ContainerName: "name",
		Image:         "image",
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
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func equalEnv(kobjectEnvs []kobject.EnvVar, k8sEnvs []api.EnvVar) bool {
	if len(kobjectEnvs) != len(k8sEnvs) {
		return false
	}
	for _, env := range kobjectEnvs {
		found := false
		for _, k8sEnv := range k8sEnvs {
			if env.Name == k8sEnv.Name && env.Value == k8sEnv.Value {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalPorts(kobjectPorts []kobject.Ports, k8sPorts []api.ContainerPort) bool {
	if len(kobjectPorts) != len(k8sPorts) {
		return false
	}
	for _, port := range kobjectPorts {
		found := false
		for _, k8sPort := range k8sPorts {
			// FIXME: HostPort should be copied to container port
			//if port.HostPort == k8sPort.HostPort && port.Protocol == k8sPort.Protocol && port.ContainerPort == k8sPort.ContainerPort {
			if port.Protocol == string(k8sPort.Protocol) && port.ContainerPort == k8sPort.ContainerPort {
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
	if len(template.Spec.Volumes) == 0 || len(template.Spec.Volumes[0].Name) == 0 || template.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim == nil && template.Spec.Volumes[0].ConfigMap == nil {
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
	if config.Stdin != template.Spec.Containers[0].Stdin {
		return fmt.Errorf("Found different values for stdin: %#v vs. %#v", config.Stdin, template.Spec.Containers[0].Stdin)
	}
	if config.Tty != template.Spec.Containers[0].TTY {
		return fmt.Errorf("Found different values for TTY: %#v vs. %#v", config.Tty, template.Spec.Containers[0].TTY)
	}
	if config.ImagePullSecret != template.Spec.ImagePullSecrets[0].Name {
		return fmt.Errorf("Found different values for ImagePullSecrets: %#v vs. %#v", config.ImagePullSecret, template.Spec.ImagePullSecrets[0].Name)
	}
	return nil
}

func privilegedNilOrFalse(template api.PodTemplateSpec) bool {
	return len(template.Spec.Containers) == 0 || template.Spec.Containers[0].SecurityContext == nil ||
		template.Spec.Containers[0].SecurityContext.Privileged == nil || !*template.Spec.Containers[0].SecurityContext.Privileged
}

func checkService(config kobject.ServiceConfig, svc *api.Service, expectedLabels map[string]string) error {
	if !equalStringMaps(expectedLabels, svc.Spec.Selector) {
		return fmt.Errorf("Found unexpected selector: %#v vs. %#v", expectedLabels, svc.Spec.Selector)
	}
	for _, port := range svc.Spec.Ports {
		name := port.Name
		expectedName := strings.ToLower(name)
		if expectedName != name {
			return fmt.Errorf("Found unexpected port name: %#v vs. %#v", expectedName, name)
		}
	}
	// TODO: finish this
	return nil
}

func checkMeta(config kobject.ServiceConfig, meta metav1.ObjectMeta, expectedName string, shouldSetLabels bool) error {
	if expectedName != meta.Name {
		return fmt.Errorf("Found unexpected name: %s vs. %s", expectedName, meta.Name)
	}

	if shouldSetLabels != (len(meta.Labels) > 0) {
		return fmt.Errorf("Unexpected labels: %#v", meta.Labels)
	}
	return nil
}

func TestKomposeConvertIngress(t *testing.T) {
	testCases := map[string]struct {
		komposeObject kobject.KomposeObject
		opt           kobject.ConvertOptions
		labelValue    string
	}{
		"Convert to Ingress: label set to true":        {newKomposeObject(), kobject.ConvertOptions{CreateD: true}, "true"},
		"Convert to Ingress: label set to example.com": {newKomposeObject(), kobject.ConvertOptions{CreateD: true}, "example.com"},
	}

	for name, test := range testCases {
		var expectedHost string

		t.Log("Test case:", name)
		k := Kubernetes{}

		appName := "app"

		// Setting value for ExposeService in ServiceConfig
		config := test.komposeObject.ServiceConfigs[appName]
		config.ExposeService = test.labelValue
		test.komposeObject.ServiceConfigs[appName] = config

		switch test.labelValue {
		case "true":
			expectedHost = ""
		default:
			expectedHost = test.labelValue
		}

		// Run Transform
		objs, err := k.Transform(test.komposeObject, test.opt)
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}

		// Check results
		for _, obj := range objs {
			if ing, ok := obj.(*networkingv1beta1.Ingress); ok {
				if ing.ObjectMeta.Name != appName {
					t.Errorf("Expected ObjectMeta.Name to be %s, got %s instead", appName, ing.ObjectMeta.Name)
				}
				if ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName != appName {
					t.Errorf("Expected Backend.ServiceName to be %s, got %s instead", appName, ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName)
				}
				if ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal != config.Port[0].HostPort {
					t.Errorf("Expected Backend.ServicePort to be %d, got %v instead", config.Port[0].HostPort, ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal)
				}
				if ing.Spec.Rules[0].Host != expectedHost {
					t.Errorf("Expected Rules[0].Host to be %s, got %s instead", expectedHost, ing.Spec.Rules[0].Host)
				}
			}
		}
	}
}

func TestKomposeConvert(t *testing.T) {
	replicas := 3
	testCases := map[string]struct {
		komposeObject   kobject.KomposeObject
		opt             kobject.ConvertOptions
		expectedNumObjs int
	}{
		// objects generated are deployment, service nework policies (2) and pvc
		"Convert to Deployments (D)":                  {newKomposeObject(), kobject.ConvertOptions{CreateD: true, Replicas: replicas, IsReplicaSetFlag: true}, 6},
		"Convert to Deployments (D) with v3 replicas": {newKomposeObject(), kobject.ConvertOptions{CreateD: true}, 6},
		"Convert to DaemonSets (DS)":                  {newKomposeObject(), kobject.ConvertOptions{CreateDS: true}, 6},
		// objects generated are deployment, daemonset, ReplicationController, service and pvc
		"Convert to D, DS, and RC":                  {newKomposeObject(), kobject.ConvertOptions{CreateD: true, CreateDS: true, CreateRC: true, Replicas: replicas, IsReplicaSetFlag: true}, 7},
		"Convert to D, DS, and RC with v3 replicas": {newKomposeObject(), kobject.ConvertOptions{CreateD: true, CreateDS: true, CreateRC: true}, 7},
		// objects generated are statefulset
		"Convert to SS with replicas ":   {newKomposeObject(), kobject.ConvertOptions{Controller: StatefulStateController, Replicas: replicas, IsReplicaSetFlag: true}, 5},
		"Convert to SS without replicas": {newKomposeObject(), kobject.ConvertOptions{Controller: StatefulStateController}, 5},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		k := Kubernetes{}
		// Run Transform
		objs, err := k.Transform(test.komposeObject, test.opt)
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}
		if len(objs) != test.expectedNumObjs {
			t.Errorf("Expected %d objects returned, got %d", test.expectedNumObjs, len(objs))
		}

		var foundSVC, foundD, foundDS, foundDC, foundSS bool
		name := "app"
		labels := transformer.ConfigLabels(name)
		config := test.komposeObject.ServiceConfigs[name]
		labelsWithNetwork := transformer.ConfigLabelsWithNetwork(name, config.Network)
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
				if d, ok := obj.(*appsv1.Deployment); ok {
					if err := checkPodTemplate(config, d.Spec.Template, labelsWithNetwork); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, d.ObjectMeta, name, true); err != nil {
						t.Errorf("%v", err)
					}
					if test.opt.IsReplicaSetFlag {
						if (int)(*d.Spec.Replicas) != replicas {
							t.Errorf("Expected %d replicas, got %d", replicas, d.Spec.Replicas)
						}
					} else {
						if (int)(*d.Spec.Replicas) != newServiceConfig().Replicas {
							t.Errorf("Expected %d replicas, got %d", newServiceConfig().Replicas, d.Spec.Replicas)
						}
					}
					foundD = true
				}

				if u, ok := obj.(*unstructured.Unstructured); ok {
					if u.GetKind() == "Deployment" {
						u.SetGroupVersionKind(schema.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						})
						data, err := json.Marshal(u)
						if err != nil {
							t.Errorf("%v", err)
						}
						var d appsv1.Deployment
						if err := json.Unmarshal(data, &d); err == nil {
							if err := checkPodTemplate(config, d.Spec.Template, labelsWithNetwork); err != nil {
								t.Errorf("%v", err)
							}
							if err := checkMeta(config, d.ObjectMeta, name, true); err != nil {
								t.Errorf("%v", err)
							}
							if test.opt.IsReplicaSetFlag {
								if (int)(*d.Spec.Replicas) != replicas {
									t.Errorf("Expected %d replicas, got %d", replicas, d.Spec.Replicas)
								}
							} else {
								if (int)(*d.Spec.Replicas) != newServiceConfig().Replicas {
									t.Errorf("Expected %d replicas, got %d", newServiceConfig().Replicas, d.Spec.Replicas)
								}
							}
							foundD = true
						}
					}
				}
			}
			if test.opt.CreateDS {
				if ds, ok := obj.(*appsv1.DaemonSet); ok {
					if err := checkPodTemplate(config, ds.Spec.Template, labelsWithNetwork); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, ds.ObjectMeta, name, true); err != nil {
						t.Errorf("%v", err)
					}
					if ds.Spec.Selector != nil && len(ds.Spec.Selector.MatchLabels) > 0 {
						t.Errorf("Expect selector be unset, got: %#v", ds.Spec.Selector)
					}
					foundDS = true
				}

				if u, ok := obj.(*unstructured.Unstructured); ok {
					if u.GetKind() == "DaemonSet" {
						u.SetGroupVersionKind(schema.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "DaemonSet",
						})
						data, err := json.Marshal(u)
						if err != nil {
							t.Errorf("%v", err)
						}
						var ds appsv1.DaemonSet
						if err := json.Unmarshal(data, &ds); err == nil {
							if err := checkPodTemplate(config, ds.Spec.Template, labelsWithNetwork); err != nil {
								t.Errorf("%v", err)
							}
							if err := checkMeta(config, ds.ObjectMeta, name, true); err != nil {
								t.Errorf("%v", err)
							}
							foundDS = true
						}
					}
				}
			}

			if test.opt.Controller == StatefulStateController {
				if ss, ok := obj.(*appsv1.StatefulSet); ok {
					if err := checkPodTemplate(config, ss.Spec.Template, labelsWithNetwork); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, ss.ObjectMeta, name, true); err != nil {
						t.Errorf("%v", err)
					}
					if test.opt.IsReplicaSetFlag {
						if (int)(*ss.Spec.Replicas) != replicas {
							t.Errorf("Expected %d replicas, got %d", replicas, ss.Spec.Replicas)
						}
					} else {
						if (int)(*ss.Spec.Replicas) != newServiceConfig().Replicas {
							t.Errorf("Expected %d replicas, got %d", newServiceConfig().Replicas, ss.Spec.Replicas)
						}
					}
					foundSS = true
				}

				if u, ok := obj.(*unstructured.Unstructured); ok {
					if u.GetKind() == "Statefulset" {
						u.SetGroupVersionKind(schema.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Statefulset",
						})
						data, err := json.Marshal(u)
						if err != nil {
							t.Errorf("%v", err)
						}
						var d appsv1.Deployment
						if err := json.Unmarshal(data, &d); err == nil {
							if err := checkPodTemplate(config, d.Spec.Template, labelsWithNetwork); err != nil {
								t.Errorf("%v", err)
							}
							if err := checkMeta(config, d.ObjectMeta, name, true); err != nil {
								t.Errorf("%v", err)
							}
							if test.opt.IsReplicaSetFlag {
								if (int)(*d.Spec.Replicas) != replicas {
									t.Errorf("Expected %d replicas, got %d", replicas, d.Spec.Replicas)
								}
							} else {
								if (int)(*d.Spec.Replicas) != newServiceConfig().Replicas {
									t.Errorf("Expected %d replicas, got %d", newServiceConfig().Replicas, d.Spec.Replicas)
								}
							}
							foundSS = true
						}
					}
				}
			}

			// TODO: k8s & openshift transformer is now separated; either separate the test or combine the transformer
			if test.opt.CreateDeploymentConfig {
				if dc, ok := obj.(*deployapi.DeploymentConfig); ok {
					if err := checkPodTemplate(config, *dc.Spec.Template, labelsWithNetwork); err != nil {
						t.Errorf("%v", err)
					}
					if err := checkMeta(config, dc.ObjectMeta, name, true); err != nil {
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

		if test.opt.Controller == StatefulStateController && !foundSS {
			t.Errorf("Expected create StatefulStateController")
		}
		if test.opt.CreateDeploymentConfig != foundDC {
			t.Errorf("Expected create Deployment Config: %v, found Deployment Config: %v", test.opt.CreateDeploymentConfig, foundDC)
		}
	}
}

func TestConvertRestartOptions(t *testing.T) {
	var opt kobject.ConvertOptions
	var k Kubernetes

	testCases := map[string]struct {
		svc           kobject.KomposeObject
		restartPolicy api.RestartPolicy
	}{
		"'restart' is set to 'no'":         {kobject.KomposeObject{ServiceConfigs: map[string]kobject.ServiceConfig{"app": kobject.ServiceConfig{Image: "foobar", Restart: "no"}}}, api.RestartPolicyNever},
		"'restart' is set to 'on-failure'": {kobject.KomposeObject{ServiceConfigs: map[string]kobject.ServiceConfig{"app": kobject.ServiceConfig{Image: "foobar", Restart: "on-failure"}}}, api.RestartPolicyOnFailure},
	}

	for name, test := range testCases {
		t.Log("Test Case:", name)

		objs, err := k.Transform(test.svc, opt)
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}

		if len(objs) != 1 {
			t.Errorf("Expected only one pod, more elements generated.")
		}

		for _, obj := range objs {
			if pod, ok := obj.(*api.Pod); ok {
				if pod.Spec.RestartPolicy != test.restartPolicy {
					t.Errorf("Expected restartPolicy as %s, got %#v", test.restartPolicy, pod.Spec.RestartPolicy)
				}
			} else {
				t.Errorf("Expected 'pod' object not found one")
			}
		}
	}
}

func TestRestartOnFailure(t *testing.T) {
	kobjectWithRestartOnFailure := newKomposeObject()
	serviceConfig := kobjectWithRestartOnFailure.ServiceConfigs["app"]
	serviceConfig.Restart = "on-failure"
	kobjectWithRestartOnFailure.ServiceConfigs = map[string]kobject.ServiceConfig{"app": serviceConfig}

	// define all test cases for RestartOnFailure function
	replicas := 2
	testCase := map[string]struct {
		komposeObject kobject.KomposeObject
		opt           kobject.ConvertOptions
	}{
		// objects generated are deployment, service and replication controller
		"Do not Create Deployment (D) with restart:'on-failure'":             {kobjectWithRestartOnFailure, kobject.ConvertOptions{IsDeploymentFlag: true, Replicas: replicas}},
		"Do not Create DaemonSet (DS) with restart:'on-failure'":             {kobjectWithRestartOnFailure, kobject.ConvertOptions{IsDaemonSetFlag: true, Replicas: replicas}},
		"Do not Create ReplicationController (RC) with restart:'on-failure'": {kobjectWithRestartOnFailure, kobject.ConvertOptions{IsReplicationControllerFlag: true, Replicas: replicas}},
	}

	for name, test := range testCase {
		t.Log("Test case:", name)
		k := Kubernetes{}
		_, err := k.Transform(test.komposeObject, test.opt)
		if err != nil {
			t.Errorf("Expected nil error, got %v instead", err)
		}
	}
}

func TestInitPodSpec(t *testing.T) {
	name := "foo"
	k := Kubernetes{}
	result := k.InitPodSpec(name, newServiceConfig().Image, "")
	if result.Containers[0].Name != "foo" && result.Containers[0].Image != "image" {
		t.Fatalf("Pod object not found")
	}
}

func TestConfigTmpfs(t *testing.T) {
	name := "foo"
	k := Kubernetes{}
	resultVolumeMount, resultVolume := k.ConfigTmpfs(name, newServiceConfig())

	if resultVolumeMount[0].Name != "foo-tmpfs0" || resultVolume[0].EmptyDir.Medium != "Memory" {
		t.Fatalf("Tmpfs not found")
	}
}

func TestConfigCapabilities(t *testing.T) {
	testCases := map[string]struct {
		service kobject.ServiceConfig
		result  api.Capabilities
	}{
		"ConfigCapsWithAddDrop": {kobject.ServiceConfig{CapAdd: []string{"CHOWN", "SETUID"}, CapDrop: []string{"KILL"}}, api.Capabilities{Add: []api.Capability{api.Capability("CHOWN"), api.Capability("SETUID")}, Drop: []api.Capability{api.Capability("KILL")}}},
		"ConfigCapsNoAddDrop":   {kobject.ServiceConfig{CapAdd: nil, CapDrop: nil}, api.Capabilities{Add: []api.Capability{}, Drop: []api.Capability{}}},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		result := ConfigCapabilities(test.service)
		if !reflect.DeepEqual(result.Add, test.result.Add) || !reflect.DeepEqual(result.Drop, test.result.Drop) {
			t.Errorf("Not expected result for ConfigCapabilities")
		}
	}
}

func TestConfigAffinity(t *testing.T) {
	testCases := map[string]struct {
		service kobject.ServiceConfig
		result  *api.Affinity
	}{
		"ConfigAffinity": {
			service: kobject.ServiceConfig{
				Placement: kobject.Placement{
					PositiveConstraints: map[string]string{
						"foo": "bar",
					},
					NegativeConstraints: map[string]string{
						"baz": "qux",
					},
				},
			},
			result: &api.Affinity{
				NodeAffinity: &api.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &api.NodeSelector{
						NodeSelectorTerms: []api.NodeSelectorTerm{
							{
								MatchExpressions: []api.NodeSelectorRequirement{
									{Key: "foo", Operator: api.NodeSelectorOpIn, Values: []string{"bar"}},
									{Key: "baz", Operator: api.NodeSelectorOpNotIn, Values: []string{"qux"}},
								},
							},
						},
					},
				},
			},
		},
		"ConfigAffinity (nil)": {
			kobject.ServiceConfig{},
			nil,
		},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		result := ConfigAffinity(test.service)
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("Not expected result for ConfigAffinity")
		}
	}
}

func TestConfigTopologySpreadConstraints(t *testing.T) {
	serviceName := "app"
	testCases := map[string]struct {
		service kobject.ServiceConfig
		result  []api.TopologySpreadConstraint
	}{
		"ConfigTopologySpreadConstraint": {
			service: kobject.ServiceConfig{
				Name: serviceName,
				Placement: kobject.Placement{
					Preferences: []string{
						"zone", "ssd",
					},
				},
			},
			result: []api.TopologySpreadConstraint{
				{
					MaxSkew:           2,
					TopologyKey:       "zone",
					WhenUnsatisfiable: api.ScheduleAnyway,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: transformer.ConfigLabels(serviceName),
					},
				},
				{
					MaxSkew:           1,
					TopologyKey:       "ssd",
					WhenUnsatisfiable: api.ScheduleAnyway,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: transformer.ConfigLabels(serviceName),
					},
				},
			},
		},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		result := ConfigTopologySpreadConstraints(test.service)
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("Not expected result for ConfigTopologySpreadConstraints")
		}
	}
}

func TestMultipleContainersInPod(t *testing.T) {
	groupName := "pod_group"

	createConfig := func(name string, containerName string) kobject.ServiceConfig {
		config := newSimpleServiceConfig()
		config.Labels = map[string]string{compose.LabelServiceGroup: groupName}
		config.Name = name
		if containerName != "" {
			config.ContainerName = containerName
		}
		config.Volumes = []kobject.Volumes{
			{
				VolumeName: "mountVolume",
				MountPath:  "/data",
			},
		}
		return config
	}

	testCases := map[string]struct {
		komposeObject   kobject.KomposeObject
		opt             kobject.ConvertOptions
		expectedNumObjs int
		expectedNames   []string
	}{
		"Converted multiple containers to Deployments (D)": {
			kobject.KomposeObject{
				ServiceConfigs: map[string]kobject.ServiceConfig{
					"app1": createConfig("app1", "app1"),
					"app2": createConfig("app2", "app2"),
				},
			}, kobject.ConvertOptions{ServiceGroupMode: "label", CreateD: true}, 2, []string{"app1", "app2"}},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		k := Kubernetes{}
		// Run Transform
		objs, err := k.Transform(test.komposeObject, test.opt)
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}
		if len(objs) != test.expectedNumObjs {
			t.Errorf("Expected %d objects returned, got %d", test.expectedNumObjs, len(objs))
		}

		// Check results
		for _, obj := range objs {
			if svc, ok := obj.(*api.Service); ok {
				if svc.Name != groupName {
					t.Errorf("Expected %v returned, got %v", groupName, svc.Name)
				}
			}
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				if deployment.Name != groupName {
					t.Errorf("Expected %v returned, got %v", groupName, deployment.Name)
				}
				if len(deployment.Spec.Template.Spec.Containers) != 2 {
					t.Errorf("Expected %d returned, got %d", 2, len(deployment.Spec.Template.Spec.Containers))
				}
				nameSet := make(map[string]api.Container)
				for _, container := range deployment.Spec.Template.Spec.Containers {
					nameSet[container.Name] = container
				}
				if container, ok := nameSet[test.expectedNames[0]]; !ok {
					t.Errorf("Expected %v returned, got %v", test.expectedNames[0], container.Name)
				} else if len(container.VolumeMounts) != 1 {
					t.Errorf("Expected %v returned, got %v", 1, len(container.VolumeMounts))
				}
				if container, ok := nameSet[test.expectedNames[1]]; !ok {
					t.Errorf("Expected %v returned, got %v", test.expectedNames[1], container.Name)
				} else if len(container.VolumeMounts) != 1 {
					t.Errorf("Expected %v returned, got %v", 1, len(container.VolumeMounts))
				}
			}
		}
	}
}

func TestServiceAccountNameOnMultipleContainers(t *testing.T) {
	groupName := "pod_group"
	serviceAccountName := "my-service"

	createConfigs := func(labels map[string]string) map[string]kobject.ServiceConfig {
		createConfig := func(name string) kobject.ServiceConfig {
			config := newSimpleServiceConfig()
			config.Labels = map[string]string{compose.LabelServiceGroup: groupName}
			for k, v := range labels {
				config.Labels[k] = v
			}
			config.Name = name
			config.ContainerName = ""
			config.Volumes = []kobject.Volumes{
				{
					VolumeName: "mountVolume",
					MountPath:  "/data",
				},
			}
			return config
		}
		return map[string]kobject.ServiceConfig{"app1": createConfig("app1"), "app2": createConfig("app2")}
	}

	testCases := map[string]struct {
		komposeObject      kobject.KomposeObject
		expectedLabelNames []string
	}{
		"Converted multiple containers with ServiceAccountName": {
			kobject.KomposeObject{
				ServiceConfigs: createConfigs(map[string]string{compose.LabelServiceAccountName: serviceAccountName}),
			}, []string{serviceAccountName}},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		k := Kubernetes{}
		// Run Transform
		objs, err := k.Transform(test.komposeObject, kobject.ConvertOptions{ServiceGroupMode: "label", CreateD: true})
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}

		// Check results
		for _, obj := range objs {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				if deployment.Name != groupName {
					t.Errorf("Expected %v returned, got %v", groupName, deployment.Name)
				}
				if deployment.Spec.Template.Spec.ServiceAccountName != serviceAccountName {
					t.Errorf("Expected %v returned, got %v", serviceAccountName, deployment.Spec.Template.Spec.ServiceAccountName)
				}
			}
		}
	}
}

func TestHealthCheckOnMultipleContainers(t *testing.T) {
	groupName := "pod_group"

	createHealthCheck := func(TCPPort int32) kobject.HealthCheck {
		return kobject.HealthCheck{
			TCPPort: TCPPort,
		}
	}

	createConfig := func(name string, livenessTCPPort, readinessTCPPort int32) kobject.ServiceConfig {
		config := newSimpleServiceConfig()
		config.Labels = map[string]string{compose.LabelServiceGroup: groupName}
		config.Name = name
		config.ContainerName = name
		config.HealthChecks.Liveness = createHealthCheck(livenessTCPPort)
		config.HealthChecks.Readiness = createHealthCheck(readinessTCPPort)
		return config
	}

	testCases := map[string]struct {
		komposeObject      kobject.KomposeObject
		opt                kobject.ConvertOptions
		expectedContainers map[string]api.Container
	}{
		"Converted multiple containers to Deployments": {
			kobject.KomposeObject{
				ServiceConfigs: map[string]kobject.ServiceConfig{
					"app1": createConfig("app1", 8081, 9091),
					"app2": createConfig("app2", 8082, 9092),
				},
			},
			kobject.ConvertOptions{ServiceGroupMode: "label", CreateD: true},
			map[string]api.Container{
				"app1": {
					LivenessProbe:  configProbe(createHealthCheck(8081)),
					ReadinessProbe: configProbe(createHealthCheck(9091)),
				},
				"app2": {
					LivenessProbe:  configProbe(createHealthCheck(8082)),
					ReadinessProbe: configProbe(createHealthCheck(9092)),
				},
			},
		},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		k := Kubernetes{}
		// Run Transform
		objs, err := k.Transform(test.komposeObject, test.opt)
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}

		// Check results
		for _, obj := range objs {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				if len(deployment.Spec.Template.Spec.Containers) != len(test.expectedContainers) {
					t.Errorf("Containers len is not equal, expected %d, got %d",
						len(deployment.Spec.Template.Spec.Containers), len(test.expectedContainers))
				}
				for _, result := range deployment.Spec.Template.Spec.Containers {
					expected, ok := test.expectedContainers[result.Name]
					if !ok {
						t.Errorf("Container %s doesn't expected", result.Name)
					}
					if !reflect.DeepEqual(result.LivenessProbe, expected.LivenessProbe) {
						t.Errorf("Container %s: LivenessProbe expected %v returned, got %v", result.Name, expected.LivenessProbe, result.LivenessProbe)
					}
					if !reflect.DeepEqual(result.ReadinessProbe, expected.ReadinessProbe) {
						t.Errorf("Container %s: ReadinessProbe expected %v returned, got %v", result.Name, expected.ReadinessProbe, result.ReadinessProbe)
					}
				}
			}
		}
	}
}

func TestCreatePVC(t *testing.T) {
	storageClassName := "custom-storage-class-name"
	k := Kubernetes{}
	result, err := k.CreatePVC("", "", PVCRequestSize, "", storageClassName)
	if err != nil {
		t.Error(errors.Wrap(err, "k.CreatePVC failed"))
	}
	if *result.Spec.StorageClassName != storageClassName {
		t.Errorf("Expected %s returned, got %s", storageClassName, *result.Spec.StorageClassName)
	}
}
