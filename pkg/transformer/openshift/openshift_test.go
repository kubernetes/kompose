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

package openshift

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/runtime"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	deployapi "github.com/openshift/origin/pkg/deploy/api"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/testutils"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	"github.com/pkg/errors"
)

func newServiceConfig() kobject.ServiceConfig {
	return kobject.ServiceConfig{
		ContainerName: "myfoobarname",
		Image:         "image",
		Environment:   []kobject.EnvVar{kobject.EnvVar{Name: "env", Value: "value"}},
		Port:          []kobject.Ports{kobject.Ports{HostPort: 123, ContainerPort: 456, Protocol: kapi.ProtocolTCP}},
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
		Stdin:         true,
		Tty:           true,
	}
}

func TestOpenShiftUpdateKubernetesObjects(t *testing.T) {
	t.Log("Test case: Testing o.UpdateKubernetesObjects()")
	var object []runtime.Object
	o := OpenShift{}
	serviceConfig := newServiceConfig()
	opt := kobject.ConvertOptions{}

	object = append(object, o.initDeploymentConfig("foobar", serviceConfig, 3))
	o.UpdateKubernetesObjects("foobar", serviceConfig, opt, &object)

	for _, obj := range object {
		switch tobj := obj.(type) {
		case *deployapi.DeploymentConfig:
			t.Log("> Testing if stdin is set correctly")
			if tobj.Spec.Template.Spec.Containers[0].Stdin != serviceConfig.Stdin {
				t.Errorf("Expected stdin to be %v, got %v instead", serviceConfig.Stdin, tobj.Spec.Template.Spec.Containers[0].Stdin)
			}
			t.Log("> Testing if TTY is set correctly")
			if tobj.Spec.Template.Spec.Containers[0].TTY != serviceConfig.Tty {
				t.Errorf("Expected TTY to be %v, got %v instead", serviceConfig.Tty, tobj.Spec.Template.Spec.Containers[0].TTY)
			}
		}
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

//Test getting git remote url for a directory
func TestGetGitRemote(t *testing.T) {
	var output string
	var err error

	gitDir := testutils.CreateLocalGitDirectory(t)
	testutils.SetGitRemote(t, gitDir, "newremote", "https://git.test.com/somerepo")
	testutils.CreateGitRemoteBranch(t, gitDir, "newbranch", "newremote")
	dir := testutils.CreateLocalDirectory(t)
	defer os.RemoveAll(gitDir)
	defer os.RemoveAll(dir)

	testCases := map[string]struct {
		expectError bool
		dir         string
		branch      string
		output      string
	}{
		"Get git remote for branch success":   {false, gitDir, "newbranch", "https://git.test.com/somerepo.git"},
		"Get git remote error in non git dir": {true, dir, "", ""},
	}

	for name, test := range testCases {
		t.Log("Test case: ", name)
		output, err = GetGitCurrentRemoteURL(test.dir)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error, got success instead!")
			}
		} else {
			if err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if output != test.output {
				t.Errorf("Expected: %#v, got: %#v", test.output, output)
			}
		}
	}
}

// Test getting current git branch in a directory
func TestGitGetCurrentBranch(t *testing.T) {
	var output string
	var err error

	gitDir := testutils.CreateLocalGitDirectory(t)
	testutils.SetGitRemote(t, gitDir, "newremote", "https://git.test.com/somerepo")
	testutils.CreateGitRemoteBranch(t, gitDir, "newbranch", "newremote")
	dir := testutils.CreateLocalDirectory(t)
	defer os.RemoveAll(gitDir)
	defer os.RemoveAll(dir)

	testCases := map[string]struct {
		expectError bool
		dir         string
		output      string
	}{
		"Get git current branch success": {false, gitDir, "newbranch"},
		"Get git current branch error":   {true, dir, ""},
	}

	for name, test := range testCases {
		t.Log("Test case: ", name)
		output, err = GetGitCurrentBranch(test.dir)

		if test.expectError {
			if err == nil {
				t.Error("Expected error, got success instead!")
			}
		} else {
			if err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if output != test.output {
				t.Errorf("Expected: %#v, got: %#v", test.output, output)
			}
		}
	}
}

// Test getting compose file directory path: relative to project dir or absolute path
func TestGetComposeFileDir(t *testing.T) {
	var output string
	var err error
	wd, _ := os.Getwd()

	testCases := map[string]struct {
		inputFiles []string
		output     string
	}{
		"Get compose file dir for relative input file path": {[]string{"foo/bar.yaml"}, filepath.Join(wd, "foo")},
		"Get compose file dir for abs input file path":      {[]string{"/abs/path/to/compose.yaml"}, "/abs/path/to"},
	}

	for name, test := range testCases {
		t.Log("Test case: ", name)

		output, err = transformer.GetComposeFileDir(test.inputFiles)

		if err != nil {
			t.Errorf("Expected success, got error: %#v", err)
		}

		if output != test.output {
			t.Errorf("Expected output: %#v, got: %#v", test.output, output)
		}
	}
}

// Test getting build context relative to project's root dir
func TestGetAbsBuildContext(t *testing.T) {
	var output string
	var err error

	gitDir := testutils.CreateLocalGitDirectory(t)
	testutils.SetGitRemote(t, gitDir, "newremote", "https://git.test.com/somerepo")
	testutils.CreateGitRemoteBranch(t, gitDir, "newbranch", "newremote")
	testutils.CreateSubdir(t, gitDir, "a/b/build")
	testutils.CreateSubdir(t, gitDir, "build")
	dir := testutils.CreateLocalDirectory(t)
	defer os.RemoveAll(gitDir)
	defer os.RemoveAll(dir)

	testCases := map[string]struct {
		expectError bool
		context     string
		output      string
	}{
		"Get abs build context success case-1": {false, filepath.Join(gitDir, "a/b/build"), "a/b/build/"},
		"Get abs build context success case-2": {false, filepath.Join(gitDir, "build"), "build/"},
		"Get abs build context error case-1":   {true, "example/build", "example/build/"},
		"Get abs build context error case-2":   {true, "/tmp", ""},
	}

	for name, test := range testCases {
		t.Log("Test case: ", name)
		output, err = GetAbsBuildContext(test.context)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error, got success instead!")
			}
		} else {
			if err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if output != test.output {
				t.Errorf("Expected: %#v, got: %#v", test.output, output)
			}
		}
	}
}

// Test initializing buildconfig for a service
func TestInitBuildConfig(t *testing.T) {
	serviceName := "serviceA"
	repo := "https://git.test.com/org/repo1"
	branch := "somebranch"
	buildArgs := []kapi.EnvVar{{Name: "name", Value: "value"}}
	value := "value"
	testDir := "a/build"

	dir := testutils.CreateLocalGitDirectory(t)
	testutils.CreateSubdir(t, dir, testDir)
	defer os.RemoveAll(dir)

	testCases := []struct {
		Name          string
		ServiceConfig kobject.ServiceConfig
	}{
		{
			Name: "Service config without image key",
			ServiceConfig: kobject.ServiceConfig{
				Build:      filepath.Join(dir, testDir),
				Dockerfile: "Dockerfile-alternate",
				BuildArgs:  map[string]*string{"name": &value},
			},
		},
		{
			Name: "Service config with image key",
			ServiceConfig: kobject.ServiceConfig{
				Build:      filepath.Join(dir, testDir),
				Dockerfile: "Dockerfile-alternate",
				BuildArgs:  map[string]*string{"name": &value},
				Image:      "foo:bar",
			},
		},
	}

	for _, test := range testCases {

		bc, err := initBuildConfig(serviceName, test.ServiceConfig, repo, branch)
		if err != nil {
			t.Error(errors.Wrap(err, "initBuildConfig failed"))
		}

		assertions := map[string]struct {
			field string
			value string
		}{
			"Assert buildconfig source git URI":     {bc.Spec.CommonSpec.Source.Git.URI, repo},
			"Assert buildconfig source git Ref":     {bc.Spec.CommonSpec.Source.Git.Ref, branch},
			"Assert buildconfig source context dir": {bc.Spec.CommonSpec.Source.ContextDir, testDir + "/"},
			// BuildConfig output image is named after service name. If image key is set than tag from that is used.
			"Assert buildconfig output name":    {bc.Spec.CommonSpec.Output.To.Name, serviceName + ":" + GetImageTag(test.ServiceConfig.Image)},
			"Assert buildconfig dockerfilepath": {bc.Spec.CommonSpec.Strategy.DockerStrategy.DockerfilePath, test.ServiceConfig.Dockerfile},
		}

		for name, assertionTest := range assertions {
			if assertionTest.field != assertionTest.value {
				t.Errorf("%s Expected: %#v, got: %#v", name, assertionTest.value, assertionTest.field)
			}
		}
		if !reflect.DeepEqual(bc.Spec.CommonSpec.Strategy.DockerStrategy.Env, buildArgs) {
			t.Errorf("Expected: %#v, got: %#v", bc.Spec.CommonSpec.Strategy.DockerStrategy.Env, buildArgs)
		}
	}
}

// TestServiceWithoutPort this tests if Headless Service is created for services without Port (with label)
func TestServiceWithoutPort(t *testing.T) {
	service := kobject.ServiceConfig{
		ContainerName: "name",
		Image:         "image",
		ServiceType:   "Headless",
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}
	o := OpenShift{Kubernetes: kubernetes.Kubernetes{}}

	objects, err := o.Transform(komposeObject, kobject.ConvertOptions{CreateD: true, Replicas: 1})
	if err != nil {
		t.Error(errors.Wrap(err, "o.Transform failed"))
	}
	if err := testutils.CheckForHeadless(objects); err != nil {
		t.Error(err)
	}

}

func TestRestartOnFailure(t *testing.T) {

	service := kobject.ServiceConfig{
		Restart: "on-failure",
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: map[string]kobject.ServiceConfig{"app": service},
	}

	// define all test cases for RestartOnFailure function
	replicas := 2
	testCase := map[string]struct {
		komposeObject kobject.KomposeObject
		opt           kobject.ConvertOptions
	}{
		// objects generated are deployment, service and replication controller
		"Do not Create DeploymentConfig (DC) with restart:'on-failure'": {komposeObject, kobject.ConvertOptions{IsDeploymentConfigFlag: true, Replicas: replicas}},
	}

	for name, test := range testCase {
		t.Log("Test case:", name)
		o := OpenShift{}

		_, err := o.Transform(test.komposeObject, test.opt)
		if err == nil {
			t.Errorf("Expected an error, got %v instead", err)
		}
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

	o := OpenShift{Kubernetes: kubernetes.Kubernetes{}}

	objects, err := o.Transform(komposeObject, kobject.ConvertOptions{CreateDeploymentConfig: true, Replicas: 1})
	if err != nil {
		t.Error(errors.Wrap(err, "o.Transform failed"))
	}
	for _, obj := range objects {
		if deploymentConfig, ok := obj.(*deployapi.DeploymentConfig); ok {
			if deploymentConfig.Spec.Strategy.Type != deployapi.DeploymentStrategyTypeRecreate {
				t.Errorf("Expected %v as Strategy Type, got %v",
					deployapi.DeploymentStrategyTypeRecreate,
					deploymentConfig.Spec.Strategy.Type)
			}
		}
	}
}
