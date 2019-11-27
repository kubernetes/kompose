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
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/docker/libcompose/project"
	"github.com/fatih/structs"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//StdinData is data bytes read from stdin
var StdinData []byte

// Compose is docker compose file loader, implements Loader interface
type Compose struct {
}

// checkUnsupportedKey checks if libcompose project contains
// keys that are not supported by this loader.
// list of all unsupported keys are stored in unsupportedKey variable
// returns list of unsupported YAML keys from docker-compose
func checkUnsupportedKey(composeProject *project.Project) []string {

	// list of all unsupported keys for this loader
	// this is map to make searching for keys easier
	// to make sure that unsupported key is not going to be reported twice
	// by keeping record if already saw this key in another service
	var unsupportedKey = map[string]bool{
		"CgroupParent":  false,
		"CPUSet":        false,
		"CPUShares":     false,
		"Devices":       false,
		"DependsOn":     false,
		"DNS":           false,
		"DNSSearch":     false,
		"EnvFile":       false,
		"ExternalLinks": false,
		"ExtraHosts":    false,
		"Ipc":           false,
		"Logging":       false,
		"MacAddress":    false,
		"MemSwapLimit":  false,
		"NetworkMode":   false,
		"SecurityOpt":   false,
		"ShmSize":       false,
		"StopSignal":    false,
		"VolumeDriver":  false,
		"Uts":           false,
		"ReadOnly":      false,
		"Ulimits":       false,
		"Net":           false,
		"Sysctls":       false,
		//"Networks":    false, // We shall be spporting network now. There are special checks for Network in checkUnsupportedKey function
		"Links": false,
	}

	var keysFound []string

	// Root level keys are not yet supported except Network
	// Check to see if the default network is available and length is only equal to one.
	if _, ok := composeProject.NetworkConfigs["default"]; ok && len(composeProject.NetworkConfigs) == 1 {
		log.Debug("Default network found")
	}

	// Root level volumes are not yet supported
	if len(composeProject.VolumeConfigs) > 0 {
		keysFound = append(keysFound, "root level volumes")
	}

	for _, serviceConfig := range composeProject.ServiceConfigs.All() {
		// this reflection is used in check for empty arrays
		val := reflect.ValueOf(serviceConfig).Elem()
		s := structs.New(serviceConfig)

		for _, f := range s.Fields() {
			// Check if given key is among unsupported keys, and skip it if we already saw this key
			if alreadySaw, ok := unsupportedKey[f.Name()]; ok && !alreadySaw {
				if f.IsExported() && !f.IsZero() {
					// IsZero returns false for empty array/slice ([])
					// this check if field is Slice, and then it checks its size
					if field := val.FieldByName(f.Name()); field.Kind() == reflect.Slice {
						if field.Len() == 0 {
							// array is empty it doesn't matter if it is in unsupportedKey or not
							continue
						}
					}
					//get yaml tag name instead of variable name
					yamlTagName := strings.Split(f.Tag("yaml"), ",")[0]
					if f.Name() == "Networks" {
						// networks always contains one default element, even it isn't declared in compose v2.
						if len(serviceConfig.Networks.Networks) == 1 && serviceConfig.Networks.Networks[0].Name == "default" {
							// this is empty Network definition, skip it
							continue
						}
					}

					if linksArray := val.FieldByName(f.Name()); f.Name() == "Links" && linksArray.Kind() == reflect.Slice {
						//Links has "SERVICE:ALIAS" style, we don't support SERVICE != ALIAS
						findUnsupportedLinksFlag := false
						for i := 0; i < linksArray.Len(); i++ {
							if tmpLink := linksArray.Index(i); tmpLink.Kind() == reflect.String {
								tmpLinkStr := tmpLink.String()
								tmpLinkStrSplit := strings.Split(tmpLinkStr, ":")
								if len(tmpLinkStrSplit) == 2 && tmpLinkStrSplit[0] != tmpLinkStrSplit[1] {
									findUnsupportedLinksFlag = true
									break
								}
							}
						}
						if !findUnsupportedLinksFlag {
							continue
						}

					}

					keysFound = append(keysFound, yamlTagName)
					unsupportedKey[f.Name()] = true
				}
			}
		}
	}
	return keysFound
}

// LoadFile loads a compose file into KomposeObject
func (c *Compose) LoadFile(files []string) (kobject.KomposeObject, error) {

	// Load the json / yaml file in order to get the version value
	var version string

	for _, file := range files {
		composeVersion, err := getVersionFromFile(file)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "Unable to load yaml/json file for version parsing")
		}

		// Check that the previous file loaded matches.
		if len(files) > 0 && version != "" && version != composeVersion {
			return kobject.KomposeObject{}, errors.New("All Docker Compose files must be of the same version")
		}
		version = composeVersion
	}

	log.Debugf("Docker Compose version: %s", version)

	// Convert based on version
	switch version {
	// Use libcompose for 1 or 2
	// If blank, it's assumed it's 1 or 2
	case "", "1", "1.0", "2", "2.0":
		komposeObject, err := parseV1V2(files)
		if err != nil {
			return kobject.KomposeObject{}, err
		}
		return komposeObject, nil
		// Use docker/cli for 3
	case "3", "3.0", "3.1", "3.2", "3.3", "3.4", "3.5", "3.6", "3.7":
		komposeObject, err := parseV3(files)
		if err != nil {
			return kobject.KomposeObject{}, err
		}
		return komposeObject, nil
	default:
		return kobject.KomposeObject{}, fmt.Errorf("Version %s of Docker Compose is not supported. Please use version 1, 2 or 3", version)
	}

}

func getVersionFromFile(file string) (string, error) {
	type ComposeVersion struct {
		Version string `json:"version"` // This affects YAML as well
	}
	var version ComposeVersion
	loadedFile, err := ReadFile(file)

	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(loadedFile, &version)
	if err != nil {
		return "", err
	}

	return version.Version, nil
}
