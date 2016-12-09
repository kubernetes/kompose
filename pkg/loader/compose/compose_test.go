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

package compose

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
)

// Test if service types are parsed properly on user input
// give a service type and expect correct input
func TestHandleServiceType(t *testing.T) {
	tests := []struct {
		labelValue  string
		serviceType string
	}{
		{"NodePort", "NodePort"},
		{"nodeport", "NodePort"},
		{"LoadBalancer", "LoadBalancer"},
		{"loadbalancer", "LoadBalancer"},
		{"ClusterIP", "ClusterIP"},
		{"clusterip", "ClusterIP"},
		{"", "ClusterIP"},
	}

	for _, tt := range tests {
		result := handleServiceType(tt.labelValue)
		if result != tt.serviceType {
			t.Errorf("Expected %q, got %q", tt.serviceType, result)
		}
	}
}

func TestLoadEnvVar(t *testing.T) {
	ev1 := []string{"foo=bar"}
	rs1 := kobject.EnvVar{
		Name:  "foo",
		Value: "bar",
	}
	ev2 := []string{"foo:bar"}
	rs2 := kobject.EnvVar{
		Name:  "foo",
		Value: "bar",
	}
	ev3 := []string{"foo"}
	rs3 := kobject.EnvVar{
		Name:  "foo",
		Value: "",
	}
	ev4 := []string{"osfoo"}
	rs4 := kobject.EnvVar{
		Name:  "osfoo",
		Value: "osbar",
	}
	ev5 := []string{"foo:bar=foobar"}
	rs5 := kobject.EnvVar{
		Name:  "foo",
		Value: "bar=foobar",
	}
	ev6 := []string{"foo=foo:bar"}
	rs6 := kobject.EnvVar{
		Name:  "foo",
		Value: "foo:bar",
	}
	ev7 := []string{"foo:"}
	rs7 := kobject.EnvVar{
		Name:  "foo",
		Value: "",
	}
	ev8 := []string{"foo="}
	rs8 := kobject.EnvVar{
		Name:  "foo",
		Value: "",
	}

	tests := []struct {
		envvars []string
		results kobject.EnvVar
	}{
		{ev1, rs1},
		{ev2, rs2},
		{ev3, rs3},
		{ev4, rs4},
		{ev5, rs5},
		{ev6, rs6},
		{ev7, rs7},
		{ev8, rs8},
	}

	os.Setenv("osfoo", "osbar")

	for _, tt := range tests {
		result := loadEnvVars(tt.envvars)
		if result[0] != tt.results {
			t.Errorf("Expected %q, got %q", tt.results, result[0])
		}
	}
}

// TestUnsupportedKeys test checkUnsupportedKey function with various
// docker-compose projects
func TestUnsupportedKeys(t *testing.T) {
	// create project that will be used in test cases
	projectWithNetworks := project.NewProject(&project.Context{}, nil, nil)
	projectWithNetworks.ServiceConfigs = config.NewServiceConfigs()
	projectWithNetworks.ServiceConfigs.Add("foo", &config.ServiceConfig{
		Image: "foo/bar",
		Build: yaml.Build{
			Context: "./build",
		},
		Hostname: "localhost",
		Ports:    []string{}, // test empty array
		Networks: &yaml.Networks{
			Networks: []*yaml.Network{
				&yaml.Network{
					Name: "net1",
				},
			},
		},
	})
	projectWithNetworks.VolumeConfigs = map[string]*config.VolumeConfig{
		"foo": &config.VolumeConfig{
			Driver: "storage",
		},
	}
	projectWithNetworks.NetworkConfigs = map[string]*config.NetworkConfig{
		"foo": &config.NetworkConfig{
			Driver: "bridge",
		},
	}

	projectWithEmptyNetwork := project.NewProject(&project.Context{}, nil, nil)
	projectWithEmptyNetwork.ServiceConfigs = config.NewServiceConfigs()
	projectWithEmptyNetwork.ServiceConfigs.Add("foo", &config.ServiceConfig{
		Networks: &yaml.Networks{},
	})

	projectWithDefaultNetwork := project.NewProject(&project.Context{}, nil, nil)
	projectWithDefaultNetwork.ServiceConfigs = config.NewServiceConfigs()

	projectWithDefaultNetwork.ServiceConfigs.Add("foo", &config.ServiceConfig{
		Networks: &yaml.Networks{
			Networks: []*yaml.Network{
				&yaml.Network{
					Name: "default",
				},
			},
		},
	})

	// define all test cases for checkUnsupportedKey function
	testCases := map[string]struct {
		composeProject          *project.Project
		expectedUnsupportedKeys []string
	}{
		"With Networks (service and root level)": {
			projectWithNetworks,
			[]string{"root level networks", "root level volumes", "hostname", "networks"},
		},
		"Empty Networks on Service level": {
			projectWithEmptyNetwork,
			[]string{"networks"},
		},
		"Default root level Network": {
			projectWithDefaultNetwork,
			[]string(nil),
		},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		keys := checkUnsupportedKey(test.composeProject)
		if !reflect.DeepEqual(keys, test.expectedUnsupportedKeys) {
			t.Errorf("ERROR: Expecting unsupported keys: ['%s']. Got: ['%s']", strings.Join(test.expectedUnsupportedKeys, "', '"), strings.Join(keys, "', '"))
		}
	}

}
