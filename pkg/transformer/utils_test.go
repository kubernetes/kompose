/*
Copyright 2016 The Kubernetes Authors All rights reserved

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

package transformer

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatProviderName(t *testing.T) {
	if formatProviderName("openshift") != "OpenShift" {
		t.Errorf("Got %s, expected OpenShift", formatProviderName("openshift"))
	}
	if formatProviderName("kubernetes") != "Kubernetes" {
		t.Errorf("Got %s, expected Kubernetes", formatProviderName("kubernetes"))
	}
}

// When passing "z" or "Z" we expect "" back.
func TestZParseVolumeLabeling(t *testing.T) {
	testCase := "/foobar:/foobar:Z"
	windowVolumeTestCase := "C:\\foobar:/foobar:Z"
	_, _, _, mode, err := ParseVolume(testCase)
	if err != nil {
		t.Errorf("In test case %q, returned unexpected error %v", testCase, err)
	}
	if mode != "" {
		t.Errorf("In test case %q, returned mode %s, expected \"\"", testCase, mode)
	}

	_, _, _, mode, err = ParseVolume(windowVolumeTestCase)
	if err != nil {
		t.Errorf("In test case %q, returned unexpected error %v", windowVolumeTestCase, err)
	}
	if mode != "" {
		t.Errorf("In test case %q, returned mode %s, expected \"\"", windowVolumeTestCase, mode)
	}
}

func TestParseWindowsVolumeMountLinuxContainer(t *testing.T) {
	name := "datavolume"
	windowsHosts := "C:\\Users"
	linuxContainer := "/etc/configs/"
	mode := "rw"

	tests := []struct {
		test, volume, name, host, container, mode string
	}{
		{
			"name:host:container:mode",
			fmt.Sprintf("%s:%s:%s:%s", name, windowsHosts, linuxContainer, mode),
			name,
			windowsHosts,
			linuxContainer,
			mode,
		},
		{
			"host:container:mode",
			fmt.Sprintf("%s:%s:%s", windowsHosts, linuxContainer, mode),
			"",
			windowsHosts,
			linuxContainer,
			mode,
		},
		{
			"name:container:mode",
			fmt.Sprintf("%s:%s:%s", name, linuxContainer, mode),
			name,
			"",
			linuxContainer,
			mode,
		},
		{
			"name:host:container",
			fmt.Sprintf("%s:%s:%s", name, windowsHosts, linuxContainer),
			name,
			windowsHosts,
			linuxContainer,
			"",
		},
		{
			"host:container",
			fmt.Sprintf("%s:%s", windowsHosts, linuxContainer),
			"",
			windowsHosts,
			linuxContainer,
			"",
		},
		{
			"container:mode",
			fmt.Sprintf("%s:%s", linuxContainer, mode),
			"",
			"",
			linuxContainer,
			mode,
		},
		{
			"name:container",
			fmt.Sprintf("%s:%s", name, linuxContainer),
			name,
			"",
			linuxContainer,
			"",
		},
		{
			"container",
			linuxContainer,
			"",
			"",
			linuxContainer,
			"",
		},
	}

	for _, test := range tests {
		name, host, container, mode, err := ParseVolume(test.volume)
		if err != nil {
			t.Errorf("In test case %q, returned unexpected error %v", test.test, err)
		}
		if name != test.name {
			t.Errorf("In test case %q, returned volume name %s, expected %s", test.test, name, test.name)
		}
		if host != test.host {
			t.Errorf("In test case %q, returned host path %s, expected %s", test.test, host, test.host)
		}
		if container != test.container {
			t.Errorf("In test case %q, returned container path %s, expected %s", test.test, container, test.container)
		}
		if mode != test.mode {
			t.Errorf("In test case %q, returned access mode %s, expected %s", test.test, mode, test.mode)
		}
	}
}

func TestParseWindowsVolumeMountWindowsContainer(t *testing.T) {
	name := "datavolume"
	windowsHosts := "C:\\Users"
	windowsContainer := "D:\\Users"
	mode := "rw"

	tests := []struct {
		test, volume, name, host, container, mode string
	}{
		{
			"name:host:container:mode",
			fmt.Sprintf("%s:%s:%s:%s", name, windowsHosts, windowsContainer, mode),
			name,
			windowsHosts,
			windowsContainer,
			mode,
		},
		{
			"host:container:mode",
			fmt.Sprintf("%s:%s:%s", windowsHosts, windowsContainer, mode),
			"",
			windowsHosts,
			windowsContainer,
			mode,
		},
		{
			"name:container:mode",
			fmt.Sprintf("%s:%s:%s", name, windowsContainer, mode),
			name,
			"",
			windowsContainer,
			mode,
		},
		{
			"name:host:container",
			fmt.Sprintf("%s:%s:%s", name, windowsHosts, windowsContainer),
			name,
			windowsHosts,
			windowsContainer,
			"",
		},
		{
			"host:container",
			fmt.Sprintf("%s:%s", windowsHosts, windowsContainer),
			"",
			windowsHosts,
			windowsContainer,
			"",
		},
		{
			"container:mode",
			fmt.Sprintf("%s:%s", windowsContainer, mode),
			"",
			"",
			windowsContainer,
			mode,
		},
		{
			"name:container",
			fmt.Sprintf("%s:%s", name, windowsContainer),
			name,
			"",
			windowsContainer,
			"",
		},
		{
			"container",
			windowsContainer,
			"",
			"",
			windowsContainer,
			"",
		},
	}

	for _, test := range tests {
		name, host, container, mode, err := ParseVolume(test.volume)
		if err != nil {
			t.Errorf("In test case %q, returned unexpected error %v", test.test, err)
		}
		if name != test.name {
			t.Errorf("In test case %q, returned volume name %s, expected %s", test.test, name, test.name)
		}
		if host != test.host {
			t.Errorf("In test case %q, returned host path %s, expected %s", test.test, host, test.host)
		}
		if container != test.container {
			t.Errorf("In test case %q, returned container path %s, expected %s", test.test, container, test.container)
		}
		if mode != test.mode {
			t.Errorf("In test case %q, returned access mode %s, expected %s", test.test, mode, test.mode)
		}
	}
}

func TestParseVolume(t *testing.T) {
	name1 := "datavolume"
	host1 := "./cache"
	host2 := "~/configs"
	container1 := "/tmp/cache"
	container2 := "/etc/configs/"
	mode := "rw"

	tests := []struct {
		test, volume, name, host, container, mode string
	}{
		{
			"name:host:container:mode",
			fmt.Sprintf("%s:%s:%s:%s", name1, host1, container1, mode),
			name1,
			host1,
			container1,
			mode,
		},
		{
			"host:container:mode",
			fmt.Sprintf("%s:%s:%s", host2, container2, mode),
			"",
			host2,
			container2,
			mode,
		},
		{
			"name:container:mode",
			fmt.Sprintf("%s:%s:%s", name1, container1, mode),
			name1,
			"",
			container1,
			mode,
		},
		{
			"name:host:container",
			fmt.Sprintf("%s:%s:%s", name1, host1, container1),
			name1,
			host1,
			container1,
			"",
		},
		{
			"host:container",
			fmt.Sprintf("%s:%s", host1, container1),
			"",
			host1,
			container1,
			"",
		},
		{
			"container:mode",
			fmt.Sprintf("%s:%s", container2, mode),
			"",
			"",
			container2,
			mode,
		},
		{
			"name:container",
			fmt.Sprintf("%s:%s", name1, container1),
			name1,
			"",
			container1,
			"",
		},
		{
			"container",
			container2,
			"",
			"",
			container2,
			"",
		},
	}

	for _, test := range tests {
		name, host, container, mode, err := ParseVolume(test.volume)
		if err != nil {
			t.Errorf("In test case %q, returned unexpected error %v", test.test, err)
		}
		if name != test.name {
			t.Errorf("In test case %q, returned volume name %s, expected %s", test.test, name, test.name)
		}
		if host != test.host {
			t.Errorf("In test case %q, returned host path %s, expected %s", test.test, host, test.host)
		}
		if container != test.container {
			t.Errorf("In test case %q, returned container path %s, expected %s", test.test, container, test.container)
		}
		if mode != test.mode {
			t.Errorf("In test case %q, returned access mode %s, expected %s", test.test, mode, test.mode)
		}
	}
}

func TestGetComposeFileDir(t *testing.T) {
	output, err := GetComposeFileDir([]string{"foobar/docker-compose.yaml"})
	if err != nil {
		t.Errorf("Error with GetComposeFileDir %v", err)
	}
	if !strings.Contains(output, "foobar") {
		t.Errorf("Expected $PWD/foobar, got %v", output)
	}
}
