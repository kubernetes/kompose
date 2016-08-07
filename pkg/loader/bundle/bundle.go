/*
Copyright 2016 Skippbox, Ltd All rights reserved.

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
	"io/ioutil"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/client/bundlefile"
	"github.com/skippbox/kompose/pkg/kobject"
)

type Bundle struct {
}

// load Image from bundles file
func loadImage(service bundlefile.Service) (string, string) {
	character := "@"
	if strings.Contains(service.Image, character) {
		return service.Image[0:strings.Index(service.Image, character)], ""
	}
	return "", "Invalid image format"
}

// load Environment Variable from bundles file
func loadEnvVarsfromBundle(service bundlefile.Service) ([]kobject.EnvVar, string) {
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
				return envs, "Invalid container env " + env
			}
		}
	}
	return envs, ""
}

// load Ports from bundles file
func loadPortsfromBundle(service bundlefile.Service) ([]kobject.Ports, string) {
	ports := []kobject.Ports{}
	for _, port := range service.Ports {
		var p kobject.Protocol
		switch port.Protocol {
		default:
			p = kobject.ProtocolTCP
		case "TCP":
			p = kobject.ProtocolTCP
		case "UDP":
			p = kobject.ProtocolUDP
		}
		ports = append(ports, kobject.Ports{
			HostPort:      int32(port.Port),
			ContainerPort: int32(port.Port),
			Protocol:      p,
		})
	}
	return ports, ""
}

// load Bundlefile into KomposeObject
func (b *Bundle) LoadFile(file string) kobject.KomposeObject {
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Fatalf("Failed to read bundles file: ", err)
	}
	reader := strings.NewReader(string(buf))
	bundle, err := bundlefile.LoadFile(reader)
	if err != nil {
		logrus.Fatalf("Failed to parse bundles file: ", err)
	}

	for name, service := range bundle.Services {
		kobject.CheckUnsupportedKey(service)

		serviceConfig := kobject.ServiceConfig{}
		serviceConfig.Command = service.Command
		serviceConfig.Args = service.Args
		// convert bundle labels to annotations
		serviceConfig.Annotations = service.Labels

		image, err := loadImage(service)
		if err != "" {
			logrus.Fatalf("Failed to load image from bundles file: " + err)
		}
		serviceConfig.Image = image

		envs, err := loadEnvVarsfromBundle(service)
		if err != "" {
			logrus.Fatalf("Failed to load envvar from bundles file: " + err)
		}
		serviceConfig.Environment = envs

		ports, err := loadPortsfromBundle(service)
		if err != "" {
			logrus.Fatalf("Failed to load ports from bundles file: " + err)
		}
		serviceConfig.Port = ports

		if service.WorkingDir != nil {
			serviceConfig.WorkingDir = *service.WorkingDir
		}

		komposeObject.ServiceConfigs[name] = serviceConfig
	}

	return komposeObject
}
