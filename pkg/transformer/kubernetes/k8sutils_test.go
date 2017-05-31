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
	"strconv"
	"testing"

	"os"
	"path/filepath"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/testutils"

	"reflect"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

/*
	Test the creation of a service
*/
func TestCreateService(t *testing.T) {

	// An example service
	service := kobject.ServiceConfig{
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
		CPUQuota:      1,                    // not supported
		CapAdd:        []string{"cap_add"},  // not supported
		CapDrop:       []string{"cap_drop"}, // not supported
		Expose:        []string{"expose"},   // not supported
		Privileged:    true,
		Restart:       "always",
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	// Test the creation of the service
	svc := k.CreateService("foo", service, objects)

	if svc.Spec.Ports[0].Port != 123 {
		t.Errorf("Expected port 123 upon conversion, actual %d", svc.Spec.Ports[0].Port)
	}
}

/*
	Test the creation of a service with a memory limit
*/
func TestCreateServiceWithMemLimit(t *testing.T) {

	// An example service
	service := kobject.ServiceConfig{
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
		CPUQuota:      1,                    // not supported
		CapAdd:        []string{"cap_add"},  // not supported
		CapDrop:       []string{"cap_drop"}, // not supported
		Expose:        []string{"expose"},   // not supported
		Privileged:    true,
		Restart:       "always",
		MemLimit:      1337,
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	// Retrieve the deployment object and test that it matches the MemLimit value
	for _, obj := range objects {
		if deploy, ok := obj.(*extensions.Deployment); ok {
			memTest, _ := deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().AsInt64()
			if memTest != 1337 {
				t.Errorf("Expected 1337 for mem_limit check, got %v", memTest)
			}
		}
	}
}

/*
	Test the creation of a service with a specified user.
	The expected result is that Kompose will set user in PodSpec
*/
func TestCreateServiceWithServiceUser(t *testing.T) {

	// An example service
	service := kobject.ServiceConfig{
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
		Annotations:   map[string]string{"kompose.service.type": "nodeport"},
		CPUQuota:      1,                    // not supported
		CapAdd:        []string{"cap_add"},  // not supported
		CapDrop:       []string{"cap_drop"}, // not supported
		Expose:        []string{"expose"},   // not supported
		Privileged:    true,
		Restart:       "always",
		User:          "1234",
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}

	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	for _, obj := range objects {
		if deploy, ok := obj.(*extensions.Deployment); ok {
			uid := *deploy.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser
			if strconv.FormatInt(uid, 10) != service.User {
				t.Errorf("User in ServiceConfig is not matching user in PodSpec")
			}
		}
	}

}

func TestTransformWithPid(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456, Protocol: api.ProtocolTCP}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		Volumes:       []string{"/tmp/volume"},
		Network:       []string{"network1", "network2"},
		Restart:       "always",
		Pid:           "host",
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	for _, obj := range objects {
		if deploy, ok := obj.(*extensions.Deployment); ok {
			hostPid := deploy.Spec.Template.Spec.SecurityContext.HostPID
			if hostPid != true {
				t.Errorf("Pid in ServiceConfig is not matching HostPID in PodSpec")
			}
		}
	}
}

func TestTransformWithInvaildPid(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456, Protocol: api.ProtocolTCP}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		Volumes:       []string{"/tmp/volume"},
		Network:       []string{"network1", "network2"},
		Restart:       "always",
		Pid:           "badvalue",
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	for _, obj := range objects {
		if deploy, ok := obj.(*extensions.Deployment); ok {
			if deploy.Spec.Template.Spec.SecurityContext != nil {
				hostPid := deploy.Spec.Template.Spec.SecurityContext.HostPID
				if hostPid != false {
					t.Errorf("Pid in ServiceConfig is not matching HostPID in PodSpec")
				}
			}
		}
	}
}

func TestIsDir(t *testing.T) {
	tempPath := "/tmp/kompose_unit"
	tempDir := filepath.Join(tempPath, "i_am_dir")
	tempFile := filepath.Join(tempPath, "i_am_file")
	tempAbsentDirPath := filepath.Join(tempPath, "i_do_not_exist")

	// create directory
	err := os.MkdirAll(tempDir, 0744)
	if err != nil {
		t.Errorf("Unable to create directory: %v", err)
	}

	// create empty file
	f, err := os.Create(tempFile)
	if err != nil {
		t.Errorf("Unable to create empty file: %v", err)
	}
	f.Close()

	// Check output if directory exists
	output, err := isDir(tempDir)
	if err != nil {
		t.Error(errors.Wrap(err, "isDir failed"))
	}
	if output != true {
		t.Errorf("directory %v exists but isDir() returned %v", tempDir, output)
	}

	// Check output if file is provided
	output, err = isDir(tempFile)
	if err != nil {
		t.Error(errors.Wrap(err, "isDir failed"))
	}
	if output != false {
		t.Errorf("%v is a file but isDir() returned %v", tempDir, output)
	}

	// Check output if path does not exist
	output, err = isDir(tempAbsentDirPath)
	if err != nil {
		t.Error(errors.Wrap(err, "isDir failed"))
	}
	if output != false {
		t.Errorf("Directory %v does not exist, but isDir() returned %v", tempAbsentDirPath, output)
	}

	// delete temporary directory
	err = os.RemoveAll(tempPath)
	if err != nil {
		t.Errorf("Error removing the temporary directory during cleanup: %v", err)
	}
}

// TestServiceWithoutPort this tests if Headless Service is created for services without Port.
func TestServiceWithoutPort(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}

	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}
	if err := testutils.CheckForHeadless(objects); err != nil {
		t.Error(err)
	}

}

// Tests if deployment strategy is being set to Recreate when volumes are
// present
func TestRecreateStrategyWithVolumesPresent(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Volumes:       []string{"/tmp/volume"},
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}

	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}
	for _, obj := range objects {
		if deployment, ok := obj.(*extensions.Deployment); ok {
			if deployment.Spec.Strategy.Type != extensions.RecreateDeploymentStrategyType {
				t.Errorf("Expected %v as Strategy Type, got %v",
					extensions.RecreateDeploymentStrategyType,
					deployment.Spec.Strategy.Type)
			}
		}
	}
}

func TestSortedKeys(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
	}
	service1 := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
	}
	c := []string{"a", "b"}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"b": service, "a": service1},
	}
	a := SortedKeys(komposeObject)
	if !reflect.DeepEqual(a, c) {
		t.Logf("Test Fail output should be %s", c)
	}
}

//test conversion from duration string to seconds *int64
func TestDurationStrToSecondsInt(t *testing.T) {
	testCases := map[string]struct {
		in  string
		out *int64
	}{
		"5s":         {in: "5s", out: &[]int64{5}[0]},
		"1m30s":      {in: "1m30s", out: &[]int64{90}[0]},
		"empty":      {in: "", out: nil},
		"onlynumber": {in: "2", out: nil},
		"illegal":    {in: "abc", out: nil},
	}

	for name, test := range testCases {
		result, _ := DurationStrToSecondsInt(test.in)
		if test.out == nil && result != nil {
			t.Errorf("Case '%v' for TestDurationStrToSecondsInt fail, Expected 'nil' , got '%v'", name, *result)
		}
		if test.out != nil && result == nil {
			t.Errorf("Case '%v' for TestDurationStrToSecondsInt fail, Expected '%v' , got 'nil'", name, *test.out)
		}
		if test.out != nil && result != nil && *test.out != *result {
			t.Errorf("Case '%v' for TestDurationStrToSecondsInt fail, Expected '%v' , got '%v'", name, *test.out, *result)
		}
	}
}
