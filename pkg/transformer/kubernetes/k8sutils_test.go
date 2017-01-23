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
		CPUSet:        "cpu_set",            // not supported
		CPUShares:     1,                    // not supported
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
	objects := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})

	// Test the creation of the service
	svc := k.CreateService("foo", service, objects)
	if svc.Spec.Ports[0].Port != 123 {
		t.Errorf("Expected port 123 upon conversion, actual %d", svc.Spec.Ports[0].Port)
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
		CPUSet:        "cpu_set",            // not supported
		CPUShares:     1,                    // not supported
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

	objects := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})

	for _, obj := range objects {
		if deploy, ok := obj.(*extensions.Deployment); ok {
			uid := *deploy.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser
			if strconv.FormatInt(uid, 10) != service.User {
				t.Errorf("User in ServiceConfig is not matching user in PodSpec")
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
	output := isDir(tempDir)
	if output != true {
		t.Errorf("directory %v exists but isDir() returned %v", tempDir, output)
	}

	// Check output if file is provided
	output = isDir(tempFile)
	if output != false {
		t.Errorf("%v is a file but isDir() returned %v", tempDir, output)
	}

	// Check output if path does not exist
	output = isDir(tempAbsentDirPath)
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

	objects := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
	if err := testutils.CheckForHeadless(objects); err != nil {
		t.Error(err)
	}

}
