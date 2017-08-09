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

package bundle

import (
	"reflect"
	"strings"
	"testing"
)

// TestUnsupportedKeys test checkUnsupportedKey function with various
// docker-compose projects
func TestUnsupportedKeys(t *testing.T) {
	user := "user"
	workDir := "workDir"

	fullBundle := Bundlefile{
		Version: "0.1",
		Services: map[string]Service{
			"foo": Service{
				Image:      "image",
				Command:    []string{"cmd"},
				Args:       []string{"arg"},
				Env:        []string{"env"},
				Labels:     map[string]string{"key": "value"},
				Ports:      []Port{Port{Protocol: "tcp", Port: uint32(80)}},
				WorkingDir: &workDir, //there is no other way to get pointer to string
				User:       &user,
				Networks:   []string{"net"},
			},
		},
	}

	bundleWithEmptyNetworks := Bundlefile{
		Version: "0.1",
		Services: map[string]Service{
			"foo": Service{
				Image:      "image",
				Command:    []string{"cmd"},
				Args:       []string{"arg"},
				Env:        []string{"env"},
				Labels:     map[string]string{"key": "value"},
				Ports:      []Port{Port{Protocol: "tcp", Port: uint32(80)}},
				WorkingDir: &workDir, //there is no other way to get pointer to string
				User:       &user,
				Networks:   []string{},
			},
		},
	}
	// define all test cases for checkUnsupportedKey function
	testCases := map[string]struct {
		bundleFile              Bundlefile
		expectedUnsupportedKeys []string
	}{
		"Full Bundle": {
			fullBundle,
			[]string{"Networks"},
		},
		"Bundle with empty Networks": {
			bundleWithEmptyNetworks,
			[]string(nil),
		},
	}

	for name, test := range testCases {
		t.Log("Test case:", name)
		keys := checkUnsupportedKey(&test.bundleFile)
		if !reflect.DeepEqual(keys, test.expectedUnsupportedKeys) {
			t.Errorf("ERROR: Expecting unsupported keys: ['%s']. Got: ['%s']", strings.Join(test.expectedUnsupportedKeys, "', '"), strings.Join(keys, "', '"))
		}
	}

}
