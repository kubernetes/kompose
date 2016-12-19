package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
)

func TestValidateController(t *testing.T) {
	testsCases := map[string]struct {
		provider    string
		controllers []string
		opt         kobject.ConvertOptions
	}{
		"Provider: kubernetes, Controller: deployment": {"kubernetes",
			[]string{"deployment"},
			kobject.ConvertOptions{CreateD: true}},
		"Provider: kubernetes, Controller: replicationcontroller, daemonset": {"kubernetes",
			[]string{"replicationcontroller", "daemonset"},
			kobject.ConvertOptions{CreateRC: true, CreateDS: true}},
		"Provider: kubernetes, Controller: deploymentconfig, should return empty opt": {"kubernetes",
			[]string{"deploymentconfig"},
			kobject.ConvertOptions{}},
		"Provider: openshift, Controller: deploymentConfig": {"openshift",
			[]string{"deploymentconfig"},
			kobject.ConvertOptions{CreateDeploymentConfig: true}},
		"Provider: openshift, controller: deployment": {"openshift",
			[]string{"deployment"},
			kobject.ConvertOptions{}},
	}

	for name, test := range testsCases {
		t.Log("Test case: ", name)

		opt, _ := validateController(test.provider, test.controllers)
		if opt != test.opt {
			t.Errorf("Expected %#v obj, got %#v.", test.opt, opt)
		}
	}
}

func TestValidatePreferenceFile(t *testing.T) {
	testCases := map[string]struct {
		filename string
		err      error
	}{
		"Invalid controller type given: 'rcs' in 'kompose-invalid-objects.yml'":           {"kompose-invalid-objects.yml", fmt.Errorf("")},
		"Invalid provider type given: 'unknownprofile' in 'kompose-invalid-provider.yml'": {"kompose-invalid-provider.yml", fmt.Errorf("")},
		"Valid preference file: 'kompose-valid1.yml'":                                     {"kompose-valid1.yml", nil},
	}

	for name, test := range testCases {
		t.Logf("Test case: %s", name)

		_, err := validatePreferenceFile(filepath.Join("testdata", test.filename))
		if test.err != nil && err == nil {
			t.Errorf("Expected error %v for invalid file: %s but received none.", test.filename, test.err)
		} else if test.err == nil && err != nil {
			t.Errorf("Expected no error for valid file: %s but received one. Error: %v", test.filename, err)
		}
	}
}

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		filename    string
		opt         kobject.ConvertOptions
		expectedErr error
	}{
		"Provider Kubernetes file: 'kompose-valid1.yml'": {"kompose-valid1.yml", kobject.ConvertOptions{CreateD: true, CreateRC: true, Provider: "kubernetes"}, nil},
		"Provider OpenShift file: 'kompose-valid2.yml'":  {"kompose-valid2.yml", kobject.ConvertOptions{CreateDeploymentConfig: true, Provider: "openshift"}, nil},
		"Non-existent file given":                        {"unknownfile.yml", kobject.ConvertOptions{}, fmt.Errorf("")},
	}

	for name, test := range testCases {
		t.Logf("Test case: %s", name)

		receivedOpt, receivedErr := Validate(filepath.Join("testdata", test.filename))
		if receivedErr == nil && test.expectedErr == nil {
			if receivedOpt != test.opt {
				t.Errorf("Expected %#v obj, got %#v.", test.opt, receivedOpt)
			}
		} else if receivedErr == nil && test.expectedErr != nil {
			t.Errorf("Expecting an error %v but received none.", test.expectedErr)
		} else if receivedErr != nil && test.expectedErr == nil {
			t.Errorf("Not expecting an error but received one: %v", receivedErr)
		}
	}
}
