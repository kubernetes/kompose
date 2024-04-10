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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader/compose"
	"github.com/kubernetes/kompose/pkg/testutils"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

/*
Test the creation of a service
*/
func TestCreateService(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		VolList:       []string{"/tmp/volume"},
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
	_, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	// Test the creation of the service
	svc := k.CreateService("foo", service)

	if svc.Spec.Ports[0].Port != 123 {
		t.Errorf("Expected port 123 upon conversion, actual %d", svc.Spec.Ports[0].Port)
	}
}

/*
Test the creation of a service with a memory limit and reservation
*/
func TestCreateServiceWithMemLimit(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName:  "name",
		Image:          "image",
		Environment:    []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:           []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:        []string{"cmd"},
		WorkingDir:     "dir",
		Args:           []string{"arg1", "arg2"},
		VolList:        []string{"/tmp/volume"},
		Network:        []string{"network1", "network2"}, // not supported
		Labels:         nil,
		Annotations:    map[string]string{"abc": "def"},
		CPUQuota:       1,                    // not supported
		CapAdd:         []string{"cap_add"},  // not supported
		CapDrop:        []string{"cap_drop"}, // not supported
		Expose:         []string{"expose"},   // not supported
		Privileged:     true,
		Restart:        "always",
		MemLimit:       1337,
		MemReservation: 1338,
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

	// Retrieve the deployment object and test that it matches the mem value
	for _, obj := range objects {
		if deploy, ok := obj.(*appsv1.Deployment); ok {
			memLimit, _ := deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().AsInt64()
			if memLimit != 1337 {
				t.Errorf("Expected 1337 for memory limit check, got %v", memLimit)
			}
			memReservation, _ := deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsInt64()
			if memReservation != 1338 {
				t.Errorf("Expected 1338 for memory reservation check, got %v", memReservation)
			}
		}
	}
}

/*
Test the creation of a service with a cpu limit and reservation
*/
func TestCreateServiceWithCPULimit(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName:  "name",
		Image:          "image",
		Environment:    []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:           []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:        []string{"cmd"},
		WorkingDir:     "dir",
		Args:           []string{"arg1", "arg2"},
		VolList:        []string{"/tmp/volume"},
		Network:        []string{"network1", "network2"}, // not supported
		Labels:         nil,
		Annotations:    map[string]string{"abc": "def"},
		CPUQuota:       1,                    // not supported
		CapAdd:         []string{"cap_add"},  // not supported
		CapDrop:        []string{"cap_drop"}, // not supported
		Expose:         []string{"expose"},   // not supported
		Privileged:     true,
		Restart:        "always",
		CPULimit:       10,
		CPUReservation: 1,
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

	// Retrieve the deployment object and test that it matches the cpu value
	for _, obj := range objects {
		if deploy, ok := obj.(*appsv1.Deployment); ok {
			cpuLimit := deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()
			if cpuLimit != 10 {
				t.Errorf("Expected 10 for cpu limit check, got %v", cpuLimit)
			}
			cpuReservation := deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()
			if cpuReservation != 1 {
				t.Errorf("Expected 1 for cpu reservation check, got %v", cpuReservation)
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
		Environment:   []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		VolList:       []string{"/tmp/volume"},
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
		if deploy, ok := obj.(*appsv1.Deployment); ok {
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
		Environment:   []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		VolList:       []string{"/tmp/volume"},
		Network:       []string{"network1", "network2"},
		Restart:       "always",
		Pid:           "host",
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	_, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	//for _, obj := range objects {
	//	if deploy, ok := obj.(*appsv1.Deployment); ok {
	//		hostPid := deploy.Spec.Template.Spec.SecurityContext.HostPID
	//		if !hostPid {
	//			t.Errorf("Pid in ServiceConfig is not matching HostPID in PodSpec")
	//		}
	//	}
	//}
}

func TestTransformWithInvalidPid(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:       []string{"cmd"},
		WorkingDir:    "dir",
		Args:          []string{"arg1", "arg2"},
		VolList:       []string{"/tmp/volume"},
		Network:       []string{"network1", "network2"},
		Restart:       "always",
		Pid:           "badvalue",
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	_, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	//for _, obj := range objects {
	//	if deploy, ok := obj.(*appsv1.Deployment); ok {
	//		if deploy.Spec.Template.Spec.SecurityContext != nil {
	//			hostPid := deploy.Spec.Template.Spec.SecurityContext.HostPID
	//			if hostPid {
	//				t.Errorf("Pid in ServiceConfig is not matching HostPID in PodSpec")
	//			}
	//		}
	//	}
	//}
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
	if !output {
		t.Errorf("directory %v exists but isDir() returned %v", tempDir, output)
	}

	// Check output if file is provided
	output, err = isDir(tempFile)
	if err != nil {
		t.Error(errors.Wrap(err, "isDir failed"))
	}
	if output {
		t.Errorf("%v is a file but isDir() returned %v", tempDir, output)
	}

	// Check output if path does not exist
	output, err = isDir(tempAbsentDirPath)
	if err != nil {
		t.Error(errors.Wrap(err, "isDir failed"))
	}
	if output {
		t.Errorf("Directory %v does not exist, but isDir() returned %v", tempAbsentDirPath, output)
	}

	// delete temporary directory
	err = os.RemoveAll(tempPath)
	if err != nil {
		t.Errorf("Error removing the temporary directory during cleanup: %v", err)
	}
}

// TestServiceWithHealthCheck this tests if Headless Service is created for services with HealthCheck.
func TestServiceWithHealthCheck(t *testing.T) {
	testCases := map[string]struct {
		service kobject.ServiceConfig
	}{
		"Exec": {
			service: kobject.ServiceConfig{
				ContainerName: "name",
				Image:         "image",
				ServiceType:   "Headless",
				HealthChecks: kobject.HealthChecks{
					Readiness: kobject.HealthCheck{
						Test:        []string{"arg1", "arg2"},
						Timeout:     10,
						Interval:    5,
						Retries:     3,
						StartPeriod: 60,
					},
					Liveness: kobject.HealthCheck{
						Test:        []string{"arg1", "arg2"},
						Timeout:     11,
						Interval:    6,
						Retries:     4,
						StartPeriod: 61,
					},
				},
			},
		},
		"HTTPGet": {
			service: kobject.ServiceConfig{
				ContainerName: "name",
				Image:         "image",
				ServiceType:   "Headless",
				HealthChecks: kobject.HealthChecks{
					Readiness: kobject.HealthCheck{
						HTTPPath:    "/health",
						HTTPPort:    8080,
						Timeout:     10,
						Interval:    5,
						Retries:     3,
						StartPeriod: 60,
					},
					Liveness: kobject.HealthCheck{
						HTTPPath:    "/ready",
						HTTPPort:    8080,
						Timeout:     11,
						Interval:    6,
						Retries:     4,
						StartPeriod: 61,
					},
				},
			},
		},
		"TCPSocket": {
			service: kobject.ServiceConfig{
				ContainerName: "name",
				Image:         "image",
				ServiceType:   "Headless",
				HealthChecks: kobject.HealthChecks{
					Readiness: kobject.HealthCheck{
						TCPPort:     8080,
						Timeout:     10,
						Interval:    5,
						Retries:     3,
						StartPeriod: 60,
					},
					Liveness: kobject.HealthCheck{
						TCPPort:     8080,
						Timeout:     11,
						Interval:    6,
						Retries:     4,
						StartPeriod: 61,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		k := Kubernetes{}
		komposeObject := kobject.KomposeObject{
			ServiceConfigs: map[string]kobject.ServiceConfig{"app": testCase.service},
		}
		objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
		if err != nil {
			t.Error(errors.Wrap(err, "k.Transform failed"))
		}
		if err := testutils.CheckForHealthCheckLivenessAndReadiness(objects); err != nil {
			t.Error(err)
		}
	}
}

// TestServiceWithoutPort this tests if Headless Service is created for services without Port.
func TestServiceWithoutPort(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		ServiceType:   "Headless",
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
		VolList:       []string{"/tmp/volume"},
		Volumes:       []kobject.Volumes{{SvcName: "app", MountPath: "/tmp/volume", PVCName: "app-claim0"}},
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
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
				t.Errorf("Expected %v as Strategy Type, got %v",
					appsv1.RecreateDeploymentStrategyType,
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
	a := SortedKeys(komposeObject.ServiceConfigs)
	if !reflect.DeepEqual(a, c) {
		t.Logf("Test Fail output should be %s", c)
	}
}

// test conversion from duration string to seconds *int64
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

func TestServiceWithServiceAccount(t *testing.T) {
	assertServiceAccountName := "my-service"

	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Port:          []kobject.Ports{{HostPort: 55555}},
		Labels:        map[string]string{compose.LabelServiceAccountName: assertServiceAccountName},
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}

	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}
	for _, obj := range objects {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.Spec.Template.Spec.ServiceAccountName != assertServiceAccountName {
				t.Errorf("Expected %v returned, got %v", assertServiceAccountName, deployment.Spec.Template.Spec.ServiceAccountName)
			}
		}
	}
}

func TestCreateServiceWithSpecialName(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "front_end",
		Image:         "nginx",
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
	expectedContainerName := "front-end"
	for _, obj := range objects {
		if deploy, ok := obj.(*appsv1.Deployment); ok {
			containerName := deploy.Spec.Template.Spec.Containers[0].Name
			if containerName != "front-end" {
				t.Errorf("Error while transforming container name. Expected %s Got %s", expectedContainerName, containerName)
			}
		}
	}
}

func TestArgsInterpolation(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		Environment:   []kobject.EnvVar{{Name: "PROTOCOL", Value: "https"}, {Name: "DOMAIN", Value: "google.com"}},
		Port:          []kobject.Ports{{HostPort: 123, ContainerPort: 456, Protocol: string(corev1.ProtocolTCP)}},
		Command:       []string{"curl"},
		Args:          []string{"$PROTOCOL://$DOMAIN/"},
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

	expectedArgs := []string{"$(PROTOCOL)://$(DOMAIN)/"}
	for _, obj := range objects {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			args := deployment.Spec.Template.Spec.Containers[0].Args[0]
			if args != expectedArgs[0] {
				t.Errorf("Expected args %v upon conversion, actual %v", expectedArgs, args)
			}
		}
	}
}

func TestReadOnlyRootFS(t *testing.T) {
	// An example service
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		ReadOnly:      true,
	}

	// An example object generated via k8s runtime.Objects()
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	k := Kubernetes{}
	objects, err := k.Transform(komposeObject, kobject.ConvertOptions{CreateD: true})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	for _, obj := range objects {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			readOnlyFS := deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem
			if *readOnlyFS != true {
				t.Errorf("Expected ReadOnlyRootFileSystem %v upon conversion, actual %v", true, readOnlyFS)
			}
		}
	}
}

func TestFormatEnvName(t *testing.T) {
	type args struct {
		name        string
		serviceName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "check dot conversion",
			args: args{
				name: "random.test",
			},
			want: "random-test",
		},
		{
			name: "check that path is shortened",
			args: args{
				name: "random/test/v1",
			},
			want: "v1",
		},
		{
			name: "check that ./ is removed",
			args: args{
				name: "./random",
			},
			want: "random",
		},
		{
			name: "check that ./ is removed",
			args: args{
				name: "abcdefghijklnmopqrstuvxyzabcdefghijklmnopqrstuvwxyzabcdejghijkl$Hereisadditional",
			},
			want: "abcdefghijklnmopqrstuvxyzabcdefghijklmnopqrstuvwxyzabcdejghijkl",
		},
		{
			name: "check that not begins with -",
			args: args{
				name:        "src/app/.env",
				serviceName: "app",
			},
			want: "app-env",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatEnvName(tt.args.name, tt.args.serviceName); got != tt.want {
				t.Errorf("FormatEnvName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test empty interfaces removal
func TestRemoveEmptyInterfaces(t *testing.T) {
	type Obj = map[string]interface{}
	var testCases = []struct {
		input  interface{}
		output interface{}
	}{
		{Obj{"useless": Obj{}}, Obj{}},
		{Obj{"usefull": Obj{"usefull": "usefull"}}, Obj{"usefull": Obj{"usefull": "usefull"}}},
		{Obj{"usefull": Obj{"usefull": "usefull", "uselessdeep": Obj{}, "uselessnil": nil}}, Obj{"usefull": Obj{"usefull": "usefull"}}},
		{Obj{"uselessdeep": Obj{"uselessdeep": Obj{}, "uselessnil": nil}}, Obj{}},
		{Obj{"uselessempty": []interface{}{nil}}, Obj{}},
		{"test", "test"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Test removeEmptyInterfaces(%s)", tc.input), func(t *testing.T) {
			result := removeEmptyInterfaces(tc.input)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("Expected %v, got %v", tc.output, result)
			}
		})
	}
}

func Test_parseContainerCommandsFromStr(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "line command without spaces in between",
			line: `[ "bundle", "exec", "thin", "-p", "3000" ]`,
			want: []string{
				"bundle", "exec", "thin", "-p", "3000",
			},
		},
		{
			name: `line command spaces inside ""`,
			line: `[ " bundle ",   " exec ", " thin ", " -p ", "3000" ]`,
			want: []string{
				"bundle", "exec", "thin", "-p", "3000",
			},
		},
		{
			name: `more use cases for line command spaces inside ""`,
			line: `[  " bundle ",   "exec ",   " thin ", " -p ", "3000  " ]`,
			want: []string{
				"bundle", "exec", "thin", "-p", "3000",
			},
		},
		{
			name: `line command without [] and ""`,
			line: `bundle exec thin -p 3000`,
			want: []string{
				"bundle exec thin -p 3000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseContainerCommandsFromStr(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseContainerCommandsFromStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fillInitContainers(t *testing.T) {
	type args struct {
		template *api.PodTemplateSpec
		service  kobject.ServiceConfig
	}
	tests := []struct {
		name string
		args args
		want []corev1.Container
	}{
		{
			name: "Testing init container are generated from labels with ,",
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "name",
						compose.LabelInitContainerImage:   "image",
						compose.LabelInitContainerCommand: `[ "bundle", "exec", "thin", "-p", "3000" ]`,
					},
				},
			},
			want: []corev1.Container{
				{
					Name:  "name",
					Image: "image",
					Command: []string{
						"bundle", "exec", "thin", "-p", "3000",
					},
				},
			},
		},
		{
			name: "Testing init container are generated from labels without ,",
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "name",
						compose.LabelInitContainerImage:   "image",
						compose.LabelInitContainerCommand: `bundle exec thin -p 3000`,
					},
				},
			},
			want: []corev1.Container{
				{
					Name:  "name",
					Image: "image",
					Command: []string{
						`bundle exec thin -p 3000`,
					},
				},
			},
		},
		{
			name: `Testing init container with long command with vars inside and ''`,
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "init-myservice",
						compose.LabelInitContainerImage:   "busybox:1.28",
						compose.LabelInitContainerCommand: `['sh', '-c', "until nslookup myservice.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done"]`,
					},
				},
			},
			want: []corev1.Container{
				{
					Name:  "init-myservice",
					Image: "busybox:1.28",
					Command: []string{
						"sh", "-c", `until nslookup myservice.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done`,
					},
				},
			},
		},
		{
			name: `without image`,
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "init-myservice",
						compose.LabelInitContainerImage:   "",
						compose.LabelInitContainerCommand: `['sh', '-c', "until nslookup myservice.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done"]`,
					},
				},
			},
			want: nil,
		},
		{
			name: `Testing init container without name`,
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "",
						compose.LabelInitContainerImage:   "busybox:1.28",
						compose.LabelInitContainerCommand: `['sh', '-c', "until nslookup myservice.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done"]`,
					},
				},
			},
			want: []corev1.Container{
				{
					Name:  "init-service",
					Image: "busybox:1.28",
					Command: []string{
						"sh", "-c", `until nslookup myservice.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done`,
					},
				},
			},
		},
		{
			name: `Testing init container without command`,
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:    "init-service",
						compose.LabelInitContainerImage:   "busybox:1.28",
						compose.LabelInitContainerCommand: ``,
					},
				},
			},
			want: []corev1.Container{
				{
					Name:    "init-service",
					Image:   "busybox:1.28",
					Command: []string{},
				},
			},
		},
		{
			name: `Testing init container without command`,
			args: args{
				template: &api.PodTemplateSpec{},
				service: kobject.ServiceConfig{
					Labels: map[string]string{
						compose.LabelInitContainerName:  "init-service",
						compose.LabelInitContainerImage: "busybox:1.28",
					},
				},
			},
			want: []corev1.Container{
				{
					Name:    "init-service",
					Image:   "busybox:1.28",
					Command: []string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillInitContainers(tt.args.template, tt.args.service)
			if !reflect.DeepEqual(tt.args.template.Spec.InitContainers, tt.want) {
				t.Errorf("Test_fillInitContainers Fail got %v, want %v", tt.args.template.Spec.InitContainers, tt.want)
			}
		})
	}
}

func Test_removeFromSlice(t *testing.T) {
	type args struct {
		objects        []runtime.Object
		objectToRemove runtime.Object
	}
	tests := []struct {
		name string
		args args
		want []runtime.Object
	}{
		{
			name: "remove object in the middle",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove objects in the middle and last object",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
		},
		{
			name: "remove 1 object",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove object at the beginning",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove object at the end",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove object that doesn't exist",
			args: args{
				objects: []runtime.Object{
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "missing"}},
			},
			want: []runtime.Object{
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove object from empty slice",
			args: args{
				objects:        []runtime.Object{},
				objectToRemove: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
			},
			want: []runtime.Object{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeFromSlice(tt.args.objects, tt.args.objectToRemove); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeFromSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeTargetDeployment(t *testing.T) {
	type args struct {
		objects              *[]runtime.Object
		targetDeploymentName string
	}
	tests := []struct {
		name string
		args args
		want *[]runtime.Object
	}{
		{
			name: "remove middle object",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove 2 objects from slice",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove 2 object from slice, only persist last one",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove target deployment from an empty slice",
			args: args{
				objects:              &[]runtime.Object{},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{},
		},
		{
			name: "remove target deployment that does not exist",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				targetDeploymentName: "missing",
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
		{
			name: "remove target deployment from a slice with only the target deployment",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
				},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{},
		},
		{
			name: "remove all targets",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "remove"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
				targetDeploymentName: "remove",
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeTargetDeployment(tt.args.objects, tt.args.targetDeploymentName)
			if !reflect.DeepEqual(tt.args.objects, tt.want) {
				t.Errorf("removeFromSlice() = %v, want %v", tt.args.objects, tt.want)
			}
		})
	}
}

func Test_removeDeploymentTransfered(t *testing.T) {
	type args struct {
		deploymentMappings []DeploymentMapping
		objects            *[]runtime.Object
	}
	tests := []struct {
		name string
		args args
		want *[]runtime.Object
	}{
		{
			name: "remove deployment already transferred",
			args: args{
				deploymentMappings: []DeploymentMapping{
					{
						TargetDeploymentName: "app",
						SourceDeploymentName: "db",
					},
				},
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
			},
		},
		{
			name: "remove multiple deployments already transferred",
			args: args{
				deploymentMappings: []DeploymentMapping{
					{
						TargetDeploymentName: "app",
						SourceDeploymentName: "db",
					},
					{
						TargetDeploymentName: "web",
						SourceDeploymentName: "cache",
					},
				},
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "web"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "cache"}},
				},
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "web"}},
			},
		},
		{
			name: "remove deployment transferred with multiple containers",
			args: args{
				deploymentMappings: []DeploymentMapping{
					{
						TargetDeploymentName: "app",
						SourceDeploymentName: "db",
					},
				},
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
			},
		},
		{
			name: "remove deployment transferred with different container names",
			args: args{
				deploymentMappings: []DeploymentMapping{
					{
						TargetDeploymentName: "app",
						SourceDeploymentName: "db",
					},
				},
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "db"}},
				},
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
			},
		},
		{
			name: "remove deployment transferred when no matching source deployment",
			args: args{
				deploymentMappings: []DeploymentMapping{
					{
						TargetDeploymentName: "app",
						SourceDeploymentName: "db",
					},
				},
				objects: &[]runtime.Object{
					&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
				},
			},
			want: &[]runtime.Object{
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeDeploymentTransfered(tt.args.deploymentMappings, tt.args.objects)
			if !reflect.DeepEqual(tt.args.objects, tt.want) {
				t.Errorf("removeFromSlice() = %v, want %v", tt.args.objects, tt.want)
			}
		})
	}
}

func Test_searchNetworkModeToService(t *testing.T) {
	tests := []struct {
		name     string
		services map[string]kobject.ServiceConfig
		want     []DeploymentMapping
	}{
		{
			name: "search network mode to service",
			services: map[string]kobject.ServiceConfig{
				"app": {
					Name: "app",
				},
				"db": {
					Name:        "db",
					NetworkMode: "service:app",
				},
			},
			want: []DeploymentMapping{
				{
					SourceDeploymentName: "db",
					TargetDeploymentName: "app",
				},
			},
		},
		{
			name: "error and not set service:app",
			services: map[string]kobject.ServiceConfig{
				"app": {
					Name: "app",
				},
				"db": {
					Name: "db",
				},
			},
			want: []DeploymentMapping{},
		},
		{
			name: "search network mode to service with multiple source deployments",
			services: map[string]kobject.ServiceConfig{
				"app": {
					Name: "app",
				},
				"db1": {
					Name:        "db1",
					NetworkMode: "service:app",
				},
				"db2": {
					Name:        "db2",
					NetworkMode: "service:app",
				},
			},
			want: []DeploymentMapping{
				{
					SourceDeploymentName: "db1",
					TargetDeploymentName: "app",
				},
				{
					SourceDeploymentName: "db2",
					TargetDeploymentName: "app",
				},
			},
		},
		{
			name: "search network mode to service with multiple target deployments",
			services: map[string]kobject.ServiceConfig{
				"app1": {
					Name: "app1",
				},
				"app2": {
					Name: "app2",
				},
				"db": {
					Name:        "db",
					NetworkMode: "service:app1",
				},
			},
			want: []DeploymentMapping{
				{
					SourceDeploymentName: "db",
					TargetDeploymentName: "app1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDeploymentMappings := searchNetworkModeToService(tt.services); !reflect.DeepEqual(gotDeploymentMappings, tt.want) {
				t.Errorf("searchNetworkModeToService() = %v, want %v", gotDeploymentMappings, tt.want)
			}
		})
	}
}

func Test_addContainersToTargetDeployment(t *testing.T) {
	k := Kubernetes{}

	appK, err := k.Transform(
		kobject.KomposeObject{
			ServiceConfigs: map[string]kobject.ServiceConfig{"app": {
				ContainerName: "app",
				Image:         "image",
			},
			},
		}, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	dbK, err := k.Transform(
		kobject.KomposeObject{
			ServiceConfigs: map[string]kobject.ServiceConfig{"db": {
				ContainerName: "db",
				Image:         "image",
				NetworkMode:   "service:app",
			},
			},
		}, kobject.ConvertOptions{CreateD: true, Replicas: 3})
	if err != nil {
		t.Error(errors.Wrap(err, "k.Transform failed"))
	}

	type args struct {
		objects                  []runtime.Object
		containersToAppend       []api.Container
		nameDeploymentToTransfer string
	}
	const (
		FirstObject  = 0
		SecondObject = 1
	)
	tests := []struct {
		name         string
		args         args
		wantAfter    int
		wantBefore   int
		targetObject int
	}{

		{
			name: "no containers to add, appk target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0]},
				containersToAppend:       []api.Container{},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    1,
			targetObject: FirstObject,
		},
		{
			name: "no match in the deployment names, no target found",
			args: args{
				objects:                  []runtime.Object{dbK[0]},
				containersToAppend:       []api.Container{},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    1,
			targetObject: FirstObject,
		},
		{
			name: "1 container more, appk target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0]},
				containersToAppend:       []api.Container{{Name: "new-1"}},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    2,
			targetObject: FirstObject,
		},
		{
			name: "1 containers and 2 more, appk target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0], dbK[0]},
				containersToAppend:       []api.Container{{Name: "new-1"}, {Name: "new-2"}},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    3,
			targetObject: FirstObject,
		},
		{
			name: "one containers to transfer, appk target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0]},
				containersToAppend:       []api.Container{{Name: "new-1"}},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    2,
			targetObject: FirstObject,
		},
		{
			name: "1 container in appk and add 2 container in dbK, db target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0], dbK[0]},
				containersToAppend:       []api.Container{{Name: "new-1"}, {Name: "new-2"}},
				nameDeploymentToTransfer: "db",
			},
			wantBefore:   1,
			wantAfter:    3,
			targetObject: SecondObject,
		},
		{
			name: "1 container in appk, cannot add 2 container in dbK, app target transfer",
			args: args{
				objects:                  []runtime.Object{appK[0], dbK[0]},
				containersToAppend:       []api.Container{{Name: "new-1"}, {Name: "new-2"}},
				nameDeploymentToTransfer: "app",
			},
			wantBefore:   1,
			wantAfter:    1,
			targetObject: SecondObject,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeContainers := (tt.args.objects)[tt.targetObject].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(beforeContainers) != tt.wantBefore {
				t.Errorf("Before Expected %d containers, got %d", tt.wantBefore, len(beforeContainers))
			}

			addContainersToTargetDeployment(&tt.args.objects, tt.args.containersToAppend, tt.args.nameDeploymentToTransfer)
			afterContainers := (tt.args.objects)[tt.targetObject].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(afterContainers) != tt.wantAfter {
				t.Errorf("After Expected %d containers, got %d", tt.wantAfter, len(afterContainers))
			}
			// reset containers
			newContainer := appK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers[0]
			appK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers = nil
			appK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers = append(appK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers, newContainer)

			newContainer = dbK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers[0]
			dbK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers = nil
			dbK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers = append(dbK[0].(*appsv1.Deployment).Spec.Template.Spec.Containers, newContainer)
		})
	}
}

func Test_addContainersFromSourceToTargetDeployment(t *testing.T) {
	type args struct {
		objects              *[]runtime.Object
		currentDeploymentMap DeploymentMapping
	}
	const (
		FirstObject  = 0
		SecondObject = 1
	)
	tests := []struct {
		name             string
		args             args
		wantBefore       int
		wantAfter        int
		targetDeployment int
	}{
		{
			name: "add one container more to target deployment",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				currentDeploymentMap: DeploymentMapping{
					SourceDeploymentName: "db",
					TargetDeploymentName: "app",
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: FirstObject,
		},
		{
			name: "no containers to transfer",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "new-1",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				currentDeploymentMap: DeploymentMapping{},
			},
			wantBefore:       1,
			wantAfter:        1,
			targetDeployment: FirstObject,
		},
		{
			name: "no containers to transfer",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "new-1",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				currentDeploymentMap: DeploymentMapping{
					SourceDeploymentName: "app",
					TargetDeploymentName: "db",
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: SecondObject,
		},
		{
			name: "target deployment not found",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "new-1",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				currentDeploymentMap: DeploymentMapping{
					SourceDeploymentName: "no-exist-deployment",
					TargetDeploymentName: "target-no-exist",
				},
			},
			wantBefore:       1,
			wantAfter:        1,
			targetDeployment: FirstObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(beforeContainers) != tt.wantBefore {
				t.Errorf("Before expected %d containers, got %d", tt.wantBefore, len(beforeContainers))
			}

			addContainersFromSourceToTargetDeployment(tt.args.objects, tt.args.currentDeploymentMap)
			afterContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(afterContainers) != tt.wantAfter {
				t.Errorf("After expected %d containers, got %d", tt.wantAfter, len(afterContainers))
			}
		})
	}
}

func Test_mergeContainersIntoDestinationDeployment(t *testing.T) {
	type args struct {
		deploymentMappings []DeploymentMapping
		objects            *[]runtime.Object
	}
	const (
		FirstObject  = 0
		SecondObject = 1
	)
	tests := []struct {
		name             string
		args             args
		wantBefore       int
		wantAfter        int
		targetDeployment int
	}{
		{
			name: "merge containers into db",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{},
								},
							},
						},
					},
				},
				deploymentMappings: []DeploymentMapping{
					{
						SourceDeploymentName: "app",
						TargetDeploymentName: "db",
					},
				},
			},
			wantBefore:       0,
			wantAfter:        1,
			targetDeployment: SecondObject,
		},
		{
			name: "merge containers into destination deployment",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db-1",
											Image: "image",
										},
										{
											Name:  "db-2",
											Image: "image",
										},
										{
											Name:  "db-3",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				deploymentMappings: []DeploymentMapping{
					{
						SourceDeploymentName: "db",
						TargetDeploymentName: "app",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        4,
			targetDeployment: FirstObject,
		},
		{
			name: "no containers to transfer",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				deploymentMappings: []DeploymentMapping{},
			},
			wantBefore:       1,
			wantAfter:        1,
			targetDeployment: FirstObject,
		},
		{
			name: "merge 2 containers db deployment into app deployment",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
										{
											Name:  "db-2",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				deploymentMappings: []DeploymentMapping{
					{
						SourceDeploymentName: "db",
						TargetDeploymentName: "app",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        3,
			targetDeployment: FirstObject,
		},
		{
			name: "merge containers in app deployment into db deployment",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				deploymentMappings: []DeploymentMapping{
					{
						SourceDeploymentName: "app",
						TargetDeploymentName: "db",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: SecondObject,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(beforeContainers) != tt.wantBefore {
				t.Errorf("Before expected %d containers, got %d", tt.wantBefore, len(beforeContainers))
			}
			mergeContainersIntoDestinationDeployment(tt.args.deploymentMappings, tt.args.objects)
			afterContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(afterContainers) != tt.wantAfter {
				t.Errorf("After expected %d containers, got %d", tt.wantAfter, len(afterContainers))
			}
		})
	}
}

func TestKubernetes_fixNetworkModeToService(t *testing.T) {
	type args struct {
		objects  *[]runtime.Object
		services map[string]kobject.ServiceConfig
	}
	const (
		FirstObject  = 0
		SecondObject = 1
	)
	tests := []struct {
		name             string
		args             args
		wantBefore       int
		wantAfter        int
		targetDeployment int
		wantDeployments  int
	}{
		{
			name: "fix network mode to service with transfer",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "nginx",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				services: map[string]kobject.ServiceConfig{
					"nginx": {
						Name:  "nginx",
						Image: "image",
					},
					"app": {
						Name:        "app",
						Image:       "image",
						NetworkMode: "service:db",
					},
					"db": {
						Name:  "db",
						Image: "image",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: FirstObject,
			wantDeployments:  2,
		},
		{
			name: "not transfer because wronge service name",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				services: map[string]kobject.ServiceConfig{
					"app": {
						Name:  "app",
						Image: "image",
					},
					"db": {
						Name:        "db",
						Image:       "image",
						NetworkMode: "service:wrong",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        1,
			targetDeployment: FirstObject,
			wantDeployments:  1,
		},
		{
			name: "deployment db removed",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				services: map[string]kobject.ServiceConfig{
					"app": {
						Name:        "app",
						Image:       "image",
						NetworkMode: "service:db",
					},
					"db": {
						Name:  "db",
						Image: "image",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: FirstObject, // db deployment removed
			wantDeployments:  1,
		},
		{
			name: "deployment app removed and added 1 container",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},

				services: map[string]kobject.ServiceConfig{
					"app": {
						Name:        "app",
						Image:       "image",
						NetworkMode: "service:db",
					},
					"db": {
						Name:  "db",
						Image: "image",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        2,
			targetDeployment: FirstObject,
			wantDeployments:  1,
		},
		{
			name: "deployment db removed and added 3 containers",
			args: args{
				objects: &[]runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "app"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "app",
											Image: "image",
										},
									},
								},
							},
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "db"},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "db-1",
											Image: "image",
										},
										{
											Name:  "db-2",
											Image: "image",
										},
										{
											Name:  "db-3",
											Image: "image",
										},
									},
								},
							},
						},
					},
				},
				services: map[string]kobject.ServiceConfig{
					"app": {
						Name:        "app",
						Image:       "image",
						NetworkMode: "service:db",
					},
					"db": {
						Name:  "db",
						Image: "image",
					},
				},
			},
			wantBefore:       1,
			wantAfter:        4,
			targetDeployment: FirstObject, // db deployment removed
			wantDeployments:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{}
			beforeContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(beforeContainers) != tt.wantBefore {
				t.Errorf("Expected %d containers, got %d", tt.wantBefore, len(beforeContainers))
			}

			k.fixNetworkModeToService(tt.args.objects, tt.args.services)
			afterContainers := (*tt.args.objects)[tt.targetDeployment].(*appsv1.Deployment).Spec.Template.Spec.Containers
			if len(afterContainers) != tt.wantAfter {
				t.Errorf("Expected %d containers, got %d", tt.wantAfter, len(afterContainers))
			}
			if len(*tt.args.objects) != tt.wantDeployments {
				t.Errorf("Expected %d deployments, got %d", tt.wantDeployments, len(*tt.args.objects))
			}
		})
	}
}
