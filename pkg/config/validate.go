package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"

	yamlTojson "github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/xeipuuv/gojsonschema"
)

// Validate the kompose preference file given is valid by comparing it
// with schema defined.
func validatePreferenceFile(prefFile string) ([]byte, error) {
	// Get the preference file name
	prefFileContents, err := ioutil.ReadFile(prefFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading file - %s: %v", prefFile, err)
	}

	// To match with the schema we need the contents to be JSON
	// converting the preference file from YAML to JSON
	prefFileJSON, err := yamlTojson.YAMLToJSON(prefFileContents)
	if err != nil {
		return nil, fmt.Errorf("Error converting YAML to JSON: %v", err)
	}
	// Created an object gojsonschema can match with schema
	documentLoader := gojsonschema.NewStringLoader(string(prefFileJSON))

	// Load the preference file schema
	schemaLoader := gojsonschema.NewStringLoader(preferenceFileSchema)

	// Actual validation of code
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("Error while validating preference file with schema: %v", err)
	}

	if !result.Valid() {
		var errs []string
		errs = append(errs, fmt.Sprintf("Preference %s is invalid, following errors found:", prefFile))
		for _, desc := range result.Errors() {
			errs = append(errs, fmt.Sprintf("- %s", desc))
		}
		return nil, fmt.Errorf(strings.Join(errs, "\n"))
	}
	return prefFileContents, nil
}

// Check if the controllers given in a file are valid and supported
// by that provider and accordingly set kobject.ConvertOptions
func validateController(provider string, controllers []string) (kobject.ConvertOptions, error) {

	var opt kobject.ConvertOptions
	// Validate the controllers coming from the preference file
	for _, controller := range controllers {
		switch provider {
		case "kubernetes":
			switch controller {
			case "deployment":
				opt.CreateD = true
			case "replicationcontroller":
				opt.CreateRC = true
			case "daemonset":
				opt.CreateDS = true
			default:
				return kobject.ConvertOptions{}, fmt.Errorf("Unsupported controller in preference file for provider %q: %q", provider, controller)
			}
		case "openshift":
			switch controller {
			case "deploymentconfig":
				opt.CreateDeploymentConfig = true
			default:
				return kobject.ConvertOptions{}, fmt.Errorf("Unsupported controller in preference file for provider %q: %q", provider, controller)
			}
		}
	}
	return opt, nil
}

// Validate function takes kompose preference file and returns 'opt'
// which has appropirate flags enabled according to file.
func Validate(prefFile string) (kobject.ConvertOptions, error) {

	prefFileContents, err := validatePreferenceFile(prefFile)
	if err != nil {
		return kobject.ConvertOptions{}, err
	}
	// read into internal datastructures
	var config Config
	if err = yaml.Unmarshal(prefFileContents, &config); err != nil {
		return kobject.ConvertOptions{}, fmt.Errorf("Error Unmarshalling file - %s: %v", prefFile, err)
	}

	// read the current profile data
	val, ok := config.Profiles[config.CurrentProfile]
	if !ok {
		return kobject.ConvertOptions{}, fmt.Errorf("Error no such profile in preference file: %q", config.CurrentProfile)
	}

	opt, err := validateController(val.Provider, val.Objects)
	if err != nil {
		return opt, err
	}
	opt.Provider = val.Provider
	return opt, nil
}
