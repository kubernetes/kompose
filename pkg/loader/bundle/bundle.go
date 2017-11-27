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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	"github.com/fatih/structs"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Bundle is docker bundle file loader, implements Loader interface
type Bundle struct {
}

// Bundlefile stores the contents of a bundlefile
type Bundlefile struct {
	Version  string
	Services map[string]Service
}

// Service is a service from a bundlefile
type Service struct {
	Image      string
	Command    []string          `json:",omitempty"`
	Args       []string          `json:",omitempty"`
	Env        []string          `json:",omitempty"`
	Labels     map[string]string `json:",omitempty"`
	Ports      []Port            `json:",omitempty"`
	WorkingDir *string           `json:",omitempty"`
	User       *string           `json:",omitempty"`
	Networks   []string          `json:",omitempty"`
}

// Port is a port as defined in a bundlefile
type Port struct {
	Protocol string
	Port     uint32
}

// checkUnsupportedKey checks if dab contains
// keys that are not supported by this loader.
// list of all unsupported keys are stored in unsupportedKey variable
// returns list of unsupported JSON/YAML keys
func checkUnsupportedKey(bundleStruct *Bundlefile) []string {
	// list of all unsupported keys for this loader
	// this is map to make searching for keys easier
	// to make sure that unsupported key is not going to be reported twice
	// by keeping record if already saw this key in another service
	var unsupportedKey = map[string]bool{
		"Networks": false,
	}

	// collect all keys found in project
	var keysFound []string
	for _, service := range bundleStruct.Services {
		// this reflection is used in check for empty arrays
		val := reflect.ValueOf(service)
		s := structs.New(service)

		for _, f := range s.Fields() {
			// Check if given key is among unsupported keys, and skip it if we already saw this key
			if alreadySaw, ok := unsupportedKey[f.Name()]; ok && !alreadySaw {
				if f.IsExported() && !f.IsZero() {
					jsonTagName := strings.Split(f.Tag("json"), ",")[0]
					if jsonTagName == "" {
						jsonTagName = f.Name()
					}
					// IsZero returns false for empty array/slice ([])
					// this check if field is Slice, and then it checks its size
					if field := val.FieldByName(f.Name()); field.Kind() == reflect.Slice {
						if field.Len() == 0 {
							// array is empty it doesn't matter if it is in unsupportedKey or not
							continue
						}
					}
					keysFound = append(keysFound, jsonTagName)
					unsupportedKey[f.Name()] = true
				}
			}
		}
	}
	return keysFound
}

// load image from dab file
func loadImage(service Service) (string, error) {
	character := "@"
	if strings.Contains(service.Image, character) {
		return service.Image[0:strings.Index(service.Image, character)], nil
	}
	return "", errors.New("Invalid image format")
}

// load environment variables from dab file
func loadEnvVars(service Service) ([]kobject.EnvVar, error) {
	envs := []kobject.EnvVar{}
	for _, env := range service.Env {
		character := "="
		if strings.Contains(env, character) {
			value := env[strings.Index(env, character)+1:]
			name := env[0:strings.Index(env, character)]
			name = strings.TrimSpace(name)
			value = strings.TrimSpace(value)
			envs = append(envs, kobject.EnvVar{
				Name:  name,
				Value: value,
			})
		} else {
			character = ":"
			if strings.Contains(env, character) {
				charQuote := "'"
				value := env[strings.Index(env, character)+1:]
				name := env[0:strings.Index(env, character)]
				name = strings.TrimSpace(name)
				value = strings.TrimSpace(value)
				if strings.Contains(value, charQuote) {
					value = strings.Trim(value, "'")
				}
				envs = append(envs, kobject.EnvVar{
					Name:  name,
					Value: value,
				})
			} else {
				return envs, errors.New("Invalid container env")
			}
		}
	}
	return envs, nil
}

// load ports from dab file
func loadPorts(service Service) ([]kobject.Ports, error) {
	ports := []kobject.Ports{}
	for _, port := range service.Ports {
		var p api.Protocol
		switch port.Protocol {
		default:
			p = api.ProtocolTCP
		case "TCP":
			p = api.ProtocolTCP
		case "UDP":
			p = api.ProtocolUDP
		}
		ports = append(ports, kobject.Ports{
			HostPort:      int32(port.Port),
			ContainerPort: int32(port.Port),
			Protocol:      p,
		})
	}
	return ports, nil
}

// LoadFile loads dab file into KomposeObject
func (b *Bundle) LoadFile(files []string) (kobject.KomposeObject, error) {
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
		LoadedFrom:     "bundle",
	}
	file := files[0]
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "ioutil.ReadFile failed, Failed to read bundles file")
	}
	reader := strings.NewReader(string(buf))
	bundle, err := loadFile(reader)
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "loadFile failed, Failed to parse bundles file")
	}

	noSupKeys := checkUnsupportedKey(bundle)
	for _, keyName := range noSupKeys {
		log.Warningf("Unsupported %s key - ignoring", keyName)
	}

	for name, service := range bundle.Services {

		serviceConfig := kobject.ServiceConfig{}
		serviceConfig.Command = service.Command
		serviceConfig.Args = service.Args
		// convert bundle labels to annotations
		serviceConfig.Annotations = service.Labels

		image, err := loadImage(service)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "loadImage failed, Failed to load image from bundles file")
		}
		serviceConfig.Image = image

		envs, err := loadEnvVars(service)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "loadEnvVars failed, Failed to load envvar from bundles file")
		}
		serviceConfig.Environment = envs

		ports, err := loadPorts(service)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "loadPorts failed, Failed to load ports from bundles file")
		}
		serviceConfig.Port = ports

		if service.WorkingDir != nil {
			serviceConfig.WorkingDir = *service.WorkingDir
		}

		komposeObject.ServiceConfigs[name] = serviceConfig
	}

	return komposeObject, nil
}

// LoadFile loads a bundlefile from a path to the file
func loadFile(reader io.Reader) (*Bundlefile, error) {
	bundlefile := &Bundlefile{}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(bundlefile); err != nil {
		switch jsonErr := err.(type) {
		case *json.SyntaxError:
			return nil, fmt.Errorf(
				"JSON syntax error at byte %v: %s",
				jsonErr.Offset,
				jsonErr.Error())
		case *json.UnmarshalTypeError:
			return nil, fmt.Errorf(
				"unexpected type at byte %v. expected %s but received %s",
				jsonErr.Offset,
				jsonErr.Type,
				jsonErr.Value)
		}
		return nil, err
	}

	return bundlefile, nil
}
