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

package compose

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
)

func durationTypesPtr(value time.Duration) *types.Duration {
	target := types.Duration(value)
	return &target
}

func TestParseHealthCheck(t *testing.T) {
	helperValue := uint64(2)
	check := types.HealthCheckConfig{
		Test:        []string{"CMD-SHELL", "echo", "foobar"},
		Timeout:     durationTypesPtr(1 * time.Second),
		Interval:    durationTypesPtr(2 * time.Second),
		Retries:     &helperValue,
		StartPeriod: durationTypesPtr(3 * time.Second),
	}

	// CMD-SHELL or SHELL is included Test within docker/cli, thus we remove the first value in Test
	expected := kobject.HealthCheck{
		Test:        []string{"echo", "foobar"},
		Timeout:     1,
		Interval:    2,
		Retries:     2,
		StartPeriod: 3,
	}
	output, err := parseHealthCheck(check, nil)
	if err != nil {
		t.Errorf("Unable to convert HealthCheckConfig: %s", err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Structs are not equal, expected: %v, output: %v", expected, output)
	}
}

func TestParseHttpHealthCheck(t *testing.T) {
	helperValue := uint64(2)
	check := types.HealthCheckConfig{
		Timeout:     durationTypesPtr(1 * time.Second),
		Interval:    durationTypesPtr(2 * time.Second),
		Retries:     &helperValue,
		StartPeriod: durationTypesPtr(3 * time.Second),
	}
	label := types.Labels{
		HealthCheckLivenessHTTPGetPath: "ping",
		HealthCheckLivenessHTTPGetPort: "80",
	}

	// CMD-SHELL or SHELL is included Test within docker/cli, thus we remove the first value in Test
	expected := kobject.HealthCheck{
		HTTPPath:    "ping",
		HTTPPort:    80,
		Timeout:     1,
		Interval:    2,
		Retries:     2,
		StartPeriod: 3,
	}
	output, err := parseHealthCheck(check, label)
	if err != nil {
		t.Errorf("Unable to convert HealthCheckConfig: %s", err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Structs are not equal, expected: %v, output: %v", expected, output)
	}
}

func TestLoadV3Volumes(t *testing.T) {
	vol := types.ServiceVolumeConfig{
		Type:     "volume",
		Source:   "/tmp/foobar",
		Target:   "/tmp/foobar",
		ReadOnly: true,
	}
	volumes := []types.ServiceVolumeConfig{vol}
	output := loadV3Volumes(volumes)
	expected := "/tmp/foobar:/tmp/foobar:ro"

	if output[0] != expected {
		t.Errorf("Expected %s, got %s", expected, output[0])
	}
}

func TestLoadV3Ports(t *testing.T) {
	for _, tt := range []struct {
		desc   string
		ports  []types.ServicePortConfig
		expose []string
		want   []kobject.Ports
	}{
		{
			desc:   "ports with expose",
			ports:  []types.ServicePortConfig{{Target: 80, Published: 80, Protocol: "TCP"}},
			expose: []string{"80", "8080"},
			want: []kobject.Ports{
				{HostPort: 80, ContainerPort: 80, Protocol: api.Protocol("TCP")},
				{HostPort: 8080, ContainerPort: 8080, Protocol: api.Protocol("TCP")},
			},
		},
		{
			desc:   "exposed port including /protocol",
			ports:  []types.ServicePortConfig{{Target: 80, Published: 80, Protocol: "TCP"}},
			expose: []string{"80/udp"},
			want: []kobject.Ports{
				{HostPort: 80, ContainerPort: 80, Protocol: api.Protocol("TCP")},
				{HostPort: 80, ContainerPort: 80, Protocol: api.Protocol("UDP")},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got := loadV3Ports(tt.ports, tt.expose)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("loadV3Ports() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

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
		result, err := handleServiceType(tt.labelValue)
		if err != nil {
			t.Error(errors.Wrap(err, "handleServiceType failed"))
		}
		if result != tt.serviceType {
			t.Errorf("Expected %q, got %q", tt.serviceType, result)
		}
	}
}

// Test loading of ports
func TestLoadPorts(t *testing.T) {
	tests := []struct {
		ports  []string
		expose []string
		want   []kobject.Ports
	}{
		{
			ports: []string{"127.0.0.1:80:80/tcp"},
			want: []kobject.Ports{
				{HostIP: "127.0.0.1", HostPort: 80, ContainerPort: 80, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"80:80/tcp"},
			want: []kobject.Ports{
				{HostPort: 80, ContainerPort: 80, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"80:80"},
			want: []kobject.Ports{
				{HostPort: 80, ContainerPort: 80, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"80"},
			want: []kobject.Ports{
				{ContainerPort: 80, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"3000-3005"},
			want: []kobject.Ports{
				{ContainerPort: 3000, Protocol: api.ProtocolTCP},
				{ContainerPort: 3001, Protocol: api.ProtocolTCP},
				{ContainerPort: 3002, Protocol: api.ProtocolTCP},
				{ContainerPort: 3003, Protocol: api.ProtocolTCP},
				{ContainerPort: 3004, Protocol: api.ProtocolTCP},
				{ContainerPort: 3005, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"3000-3005:5000-5005"},
			want: []kobject.Ports{
				{HostPort: 3000, ContainerPort: 5000, Protocol: api.ProtocolTCP},
				{HostPort: 3001, ContainerPort: 5001, Protocol: api.ProtocolTCP},
				{HostPort: 3002, ContainerPort: 5002, Protocol: api.ProtocolTCP},
				{HostPort: 3003, ContainerPort: 5003, Protocol: api.ProtocolTCP},
				{HostPort: 3004, ContainerPort: 5004, Protocol: api.ProtocolTCP},
				{HostPort: 3005, ContainerPort: 5005, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"127.0.0.1:3000-3005:5000-5005"},
			want: []kobject.Ports{
				{HostIP: "127.0.0.1", HostPort: 3000, ContainerPort: 5000, Protocol: api.ProtocolTCP},
				{HostIP: "127.0.0.1", HostPort: 3001, ContainerPort: 5001, Protocol: api.ProtocolTCP},
				{HostIP: "127.0.0.1", HostPort: 3002, ContainerPort: 5002, Protocol: api.ProtocolTCP},
				{HostIP: "127.0.0.1", HostPort: 3003, ContainerPort: 5003, Protocol: api.ProtocolTCP},
				{HostIP: "127.0.0.1", HostPort: 3004, ContainerPort: 5004, Protocol: api.ProtocolTCP},
				{HostIP: "127.0.0.1", HostPort: 3005, ContainerPort: 5005, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports: []string{"80", "3000"},
			want: []kobject.Ports{
				{HostPort: 0, ContainerPort: 80, Protocol: api.ProtocolTCP},
				{HostPort: 0, ContainerPort: 3000, Protocol: api.ProtocolTCP},
			},
		},
		{
			ports:  []string{"80", "3000"},
			expose: []string{"80", "8080"},
			want: []kobject.Ports{
				{HostPort: 0, ContainerPort: 80, Protocol: api.ProtocolTCP},
				{HostPort: 0, ContainerPort: 3000, Protocol: api.ProtocolTCP},
				{HostPort: 0, ContainerPort: 8080, Protocol: api.ProtocolTCP},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("port=%q,expose=%q", tt.ports, tt.expose), func(t *testing.T) {
			got, err := loadPorts(tt.ports, tt.expose)
			if err != nil {
				t.Fatalf("Unexpected error with loading ports %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("loadPorts() mismatch (-want +got):\n%s", diff)
			}
		})
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
				{
					Name: "net1",
				},
			},
		},
	})
	projectWithNetworks.ServiceConfigs.Add("bar", &config.ServiceConfig{
		Image: "bar/foo",
		Build: yaml.Build{
			Context: "./build",
		},
		Hostname: "localhost",
		Ports:    []string{}, // test empty array
		Networks: &yaml.Networks{
			Networks: []*yaml.Network{
				{
					Name: "net1",
				},
			},
		},
	})
	projectWithNetworks.VolumeConfigs = map[string]*config.VolumeConfig{
		"foo": {
			Driver: "storage",
		},
	}
	projectWithNetworks.NetworkConfigs = map[string]*config.NetworkConfig{
		"foo": {
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
				{
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
			//root level network and network are now supported"
			[]string{"root level volumes"},
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

func TestNormalizeServiceNames(t *testing.T) {
	testCases := []struct {
		composeServiceName    string
		normalizedServiceName string
	}{
		{"foo_bar", "foo-bar"},
		{"foo", "foo"},
		{"foo.bar", "foo.bar"},
		//{"", ""},
	}

	for _, testCase := range testCases {
		returnValue := normalizeServiceNames(testCase.composeServiceName)
		if returnValue != testCase.normalizedServiceName {
			t.Logf("Expected %q, got %q", testCase.normalizedServiceName, returnValue)
		}
	}
}

func TestNormalizeNetworkNames(t *testing.T) {
	testCases := []struct {
		composeNetworkName    string
		normalizedNetworkName string
	}{
		{"foo_bar", "foo-bar"},
		{"foo", "foo"},
		{"FOO", "foo"},
		{"foo.bar", "foo.bar"},
		//{"", ""},
	}

	for _, testCase := range testCases {
		returnValue, err := normalizeNetworkNames(testCase.composeNetworkName)
		if err != nil {
			t.Log("Unexpected error, got ", err)
		}
		if returnValue != testCase.normalizedNetworkName {
			t.Logf("Expected %q, got %q", testCase.normalizedNetworkName, returnValue)
		}
	}
}

func TestCheckLabelsPorts(t *testing.T) {
	testCases := []struct {
		name        string
		noOfPort    int
		labels      string
		svcName     string
		expectError bool
	}{
		{"ports is defined", 1, "NodePort", "foo", false},
		{"ports is not defined", 0, "NodePort", "foo", true},
	}

	var err error
	for _, testcase := range testCases {
		t.Log(testcase.name)
		err = checkLabelsPorts(testcase.noOfPort, testcase.labels, testcase.svcName)
		if testcase.expectError && err == nil {
			t.Log("Expected error, got ", err)
		}
	}
}

func TestCheckPlacementCustomLabels(t *testing.T) {
	placement := types.Placement{
		Constraints: []string{
			"node.labels.something == anything",
			"node.labels.monitor != xxx",
		},
	}
	output := loadV3Placement(placement.Constraints)

	expected := map[string]string{"something": "anything"}

	if output["something"] != expected["something"] {
		t.Errorf("Expected %s, got %s", expected, output)
	}
}
