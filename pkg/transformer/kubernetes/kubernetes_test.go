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

package kubernetes

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	deployapi "github.com/openshift/origin/pkg/deploy/api"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"os"
	"os/exec"
)

func newServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456}},
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
		Stdin:         true,
		Tty:           true,
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
			if port.Protocol == k8sPort.Protocol && port.ContainerPort == k8sPort.ContainerPort {
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
	if len(template.Spec.Volumes) == 0 || len(template.Spec.Volumes[0].Name) == 0 || template.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim == nil {
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
		objs := k.Transform(test.komposeObject, test.opt)

		// Check results
		for _, obj := range objs {
			if ing, ok := obj.(*extensions.Ingress); ok {
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
		// objects generated are deployment, service and pvc
		"Convert to Deployments (D)":            {newKomposeObject(), kobject.ConvertOptions{CreateD: true, Replicas: replicas}, 3},
		"Convert to DaemonSets (DS)":            {newKomposeObject(), kobject.ConvertOptions{CreateDS: true}, 3},
		"Convert to ReplicationController (RC)": {newKomposeObject(), kobject.ConvertOptions{CreateRC: true, Replicas: replicas}, 3},
		// objects generated are deployment, daemonset, ReplicationController, service and pvc
		"Convert to D, DS, and RC": {newKomposeObject(), kobject.ConvertOptions{CreateD: true, CreateDS: true, CreateRC: true, Replicas: replicas}, 5},
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

		objs := k.Transform(test.svc, opt)

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

// TestUnsupportedKeys test checkUnsupportedKey function
func TestUnsupportedKeys(t *testing.T) {

	kobjectWithBuild := newKomposeObject()
	kobjectWithBuild.LoadedFrom = "compose"
	serviceConfig := kobjectWithBuild.ServiceConfigs["app"]
	serviceConfig.Build = "./asdf"
	serviceConfig.Network = []string{}
	kobjectWithBuild.ServiceConfigs = map[string]kobject.ServiceConfig{"app": serviceConfig}

	// define all test cases for checkUnsupportedKey function
	testCases := map[string]struct {
		bundleFile              kobject.KomposeObject
		expectedUnsupportedKeys []string
	}{
		"Full Bundle": {
			kobjectWithBuild,
			[]string{"build"},
		},
	}

	k := Kubernetes{}

	for name, test := range testCases {
		t.Log("Test case:", name)
		keys := k.CheckUnsupportedKey(&test.bundleFile, unsupportedKey)
		if !reflect.DeepEqual(keys, test.expectedUnsupportedKeys) {
			t.Errorf("ERROR: Expecting unsupported keys: ['%s']. Got: ['%s']", strings.Join(test.expectedUnsupportedKeys, "', '"), strings.Join(keys, "', '"))
		}
	}

}

// Here we are testing a function which results in `logus.Fatalf()` when a condition is met, which further call `os.Exit()` and exits the process.
// If we write a test in the usual way that will call the function,
// it will exit the process and the running process is actually the test process and our test would fail.
// So to test the function resulting in `os.Exit()` we need invoke go test again in a separate process through `exec.Command`,
// limiting execution to the TestRestartOnFailure test using `-test.run=TestRestartOnFailure` flag set.
// The `TestRestartOnFailure` doing is two things simultaneously,
// it is going to the be the test itself and second it will a be `subprocess` that the test runs.
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
		if os.Getenv("BE_CRASHER") == "1" {
			k.Transform(test.komposeObject, test.opt)
		}
	}

	// cmd := exec.Command(os.Args[0], "-test.run=TestRestartOnFailure") will execute the test binary
	// with the flag -test.run=TestRestartOnFailure and set the environment variable BE_CRASHER=1.
	cmd := exec.Command(os.Args[0], "-test.run=TestRestartOnFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	// err := cmd.Run() will re-execute the test binary and this time os.Getenv("BE_CRASHER") == "1"
	// will return true and we can call o.Transform(test.komposeObject, test.opt).
	// so that the test binary that calls itself and execute the code on behalf of the parent process.
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Process ran with err %v, want exit status 1", err)
}

func TestInitPodSpec(t *testing.T) {
	name := "foo"
	k := Kubernetes{}
	result := k.InitPodSpec(name, newServiceConfig().Image)
	if result.Containers[0].Name != "foo" && result.Containers[0].Image != "image" {
		t.Fatalf("Pod object not found")
	}
}
