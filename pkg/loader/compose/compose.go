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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/fatih/structs"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
)

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
		"Devices":       false,
		"DependsOn":     false,
		"DNS":           false,
		"DNSSearch":     false,
		"DomainName":    false,
		"EnvFile":       false,
		"Extends":       false,
		"ExternalLinks": false,
		"ExtraHosts":    false,
		"Hostname":      false,
		"Ipc":           false,
		"Logging":       false,
		"MacAddress":    false,
		"MemLimit":      false,
		"MemSwapLimit":  false,
		"NetworkMode":   false,
		"Pid":           false,
		"SecurityOpt":   false,
		"ShmSize":       false,
		"StopSignal":    false,
		"VolumeDriver":  false,
		"Uts":           false,
		"ReadOnly":      false,
		"Ulimits":       false,
		"Dockerfile":    false,
		"Net":           false,
		"Networks":      false, // there are special checks for Network in checkUnsupportedKey function
	}

	// collect all keys found in project
	var keysFound []string

	// Root level keys are not yet supported
	// Check to see if the default network is available and length is only equal to one.
	// Else, warn the user that root level networks are not supported (yet)
	if _, ok := composeProject.NetworkConfigs["default"]; ok && len(composeProject.NetworkConfigs) == 1 {
		logrus.Debug("Default network found")
	} else if len(composeProject.NetworkConfigs) > 0 {
		keysFound = append(keysFound, "root level networks")
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
					//get yaml tag name instad of variable name
					yamlTagName := strings.Split(f.Tag("yaml"), ",")[0]
					if f.Name() == "Networks" {
						// networks always contains one default element, even it isn't declared in compose v2.
						if len(serviceConfig.Networks.Networks) == 1 && serviceConfig.Networks.Networks[0].Name == "default" {
							// this is empty Network definition, skip it
							continue
						} else {
							yamlTagName = "networks"
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

// load environment variables from compose file
func loadEnvVars(envars []string) []kobject.EnvVar {
	envs := []kobject.EnvVar{}
	for _, e := range envars {
		character := ""
		equalPos := strings.Index(e, "=")
		colonPos := strings.Index(e, ":")
		switch {
		case equalPos == -1 && colonPos == -1:
			character = ""
		case equalPos == -1 && colonPos != -1:
			character = ":"
		case equalPos != -1 && colonPos == -1:
			character = "="
		case equalPos != -1 && colonPos != -1:
			if equalPos > colonPos {
				character = ":"
			} else {
				character = "="
			}
		}

		if character == "" {
			envs = append(envs, kobject.EnvVar{
				Name:  e,
				Value: os.Getenv(e),
			})
		} else {
			values := strings.SplitN(e, character, 2)
			// try to get value from os env
			if values[1] == "" {
				values[1] = os.Getenv(values[0])
			}
			envs = append(envs, kobject.EnvVar{
				Name:  values[0],
				Value: values[1],
			})
		}
	}

	return envs
}

// Load ports from compose file
func loadPorts(composePorts []string) ([]kobject.Ports, error) {
	ports := []kobject.Ports{}
	character := ":"

	// For each port listed
	for _, port := range composePorts {

		// Get the TCP / UDP protocol. Checks to see if it splits in 2 with '/' character.
		// ex. 15000:15000/tcp
		// else, set a default protocol of using TCP
		proto := api.ProtocolTCP
		protocolCheck := strings.Split(port, "/")
		if len(protocolCheck) == 2 {
			if strings.EqualFold("tcp", protocolCheck[1]) {
				proto = api.ProtocolTCP
			} else if strings.EqualFold("udp", protocolCheck[1]) {
				proto = api.ProtocolUDP
			} else {
				return nil, fmt.Errorf("invalid protocol %q", protocolCheck[1])
			}
		}

		// Split up the ports / IP without the "/tcp" or "/udp" appended to it
		justPorts := strings.Split(protocolCheck[0], character)

		if len(justPorts) == 3 {
			// ex. 127.0.0.1:80:80

			// Get the IP address
			hostIP := justPorts[0]
			ip := net.ParseIP(hostIP)
			if ip.To4() == nil && ip.To16() == nil {
				return nil, fmt.Errorf("%q contains an invalid IPv4 or IPv6 IP address", port)
			}

			// Get the host port
			hostPortInt, err := strconv.Atoi(justPorts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid host port %q valid example: 127.0.0.1:80:80", port)
			}

			// Get the container port
			containerPortInt, err := strconv.Atoi(justPorts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid container port %q valid example: 127.0.0.1:80:80", port)
			}

			// Convert to a kobject struct with ports as well as IP
			ports = append(ports, kobject.Ports{
				HostPort:      int32(hostPortInt),
				ContainerPort: int32(containerPortInt),
				HostIP:        hostIP,
				Protocol:      proto,
			})

		} else if len(justPorts) == 2 {
			// ex. 80:80

			// Get the host port
			hostPortInt, err := strconv.Atoi(justPorts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid host port %q valid example: 80:80", port)
			}

			// Get the container port
			containerPortInt, err := strconv.Atoi(justPorts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid container port %q valid example: 80:80", port)
			}

			// Convert to a kobject struct and add to the list of ports
			ports = append(ports, kobject.Ports{
				HostPort:      int32(hostPortInt),
				ContainerPort: int32(containerPortInt),
				Protocol:      proto,
			})

		} else {
			// ex. 80

			containerPortInt, err := strconv.Atoi(justPorts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid container port %q valid example: 80", port)
			}
			ports = append(ports, kobject.Ports{
				ContainerPort: int32(containerPortInt),
				Protocol:      proto,
			})
		}

	}
	return ports, nil
}

// LoadFile loads compose file into KomposeObject
func (c *Compose) LoadFile(files []string) kobject.KomposeObject {
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
		LoadedFrom:     "compose",
	}
	context := &project.Context{}
	context.ComposeFiles = files

	if context.ResourceLookup == nil {
		context.ResourceLookup = &lookup.FileResourceLookup{}
	}

	if context.EnvironmentLookup == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return kobject.KomposeObject{}
		}
		context.EnvironmentLookup = &lookup.ComposableEnvLookup{
			Lookups: []config.EnvironmentLookup{
				&lookup.EnvfileLookup{
					Path: filepath.Join(cwd, ".env"),
				},
				&lookup.OsEnvLookup{},
			},
		}
	}

	// load compose file into composeObject
	composeObject := project.NewProject(context, nil, nil)
	err := composeObject.Parse()
	if err != nil {
		logrus.Fatalf("Failed to load compose file: %v", err)
	}

	// transform composeObject into komposeObject
	composeServiceNames := composeObject.ServiceConfigs.Keys()

	noSupKeys := checkUnsupportedKey(composeObject)
	for _, keyName := range noSupKeys {
		logrus.Warningf("Unsupported %s key - ignoring", keyName)
	}

	for _, name := range composeServiceNames {
		if composeServiceConfig, ok := composeObject.ServiceConfigs.Get(name); ok {
			serviceConfig := kobject.ServiceConfig{}
			serviceConfig.Image = composeServiceConfig.Image
			serviceConfig.Build = composeServiceConfig.Build.Context
			serviceConfig.ContainerName = composeServiceConfig.ContainerName
			serviceConfig.Command = composeServiceConfig.Entrypoint
			serviceConfig.Args = composeServiceConfig.Command
			serviceConfig.Build = composeServiceConfig.Build.Context

			envs := loadEnvVars(composeServiceConfig.Environment)
			serviceConfig.Environment = envs

			// load ports
			ports, err := loadPorts(composeServiceConfig.Ports)
			if err != nil {
				logrus.Fatalf("%q failed to load ports from compose file: %v", name, err)
			}
			serviceConfig.Port = ports

			serviceConfig.WorkingDir = composeServiceConfig.WorkingDir

			if composeServiceConfig.Volumes != nil {
				for _, volume := range composeServiceConfig.Volumes.Volumes {
					serviceConfig.Volumes = append(serviceConfig.Volumes, volume.String())
				}
			}

			// canonical "Custom Labels" handler
			// Labels used to influence conversion of kompose will be handled
			// from here for docker-compose. Each loader will have such handler.
			for key, value := range composeServiceConfig.Labels {
				switch key {
				case "kompose.service.type":
					serviceConfig.ServiceType = handleServiceType(value)
				case "kompose.service.expose":
					serviceConfig.ExposeService = strings.ToLower(value)
				}
			}

			// convert compose labels to annotations
			serviceConfig.Annotations = map[string]string(composeServiceConfig.Labels)

			serviceConfig.CPUSet = composeServiceConfig.CPUSet
			serviceConfig.CPUShares = int64(composeServiceConfig.CPUShares)
			serviceConfig.CPUQuota = int64(composeServiceConfig.CPUQuota)
			serviceConfig.CapAdd = composeServiceConfig.CapAdd
			serviceConfig.CapDrop = composeServiceConfig.CapDrop
			serviceConfig.Expose = composeServiceConfig.Expose
			serviceConfig.Privileged = composeServiceConfig.Privileged
			serviceConfig.Restart = composeServiceConfig.Restart
			serviceConfig.User = composeServiceConfig.User
			serviceConfig.VolumesFrom = composeServiceConfig.VolumesFrom
			serviceConfig.Stdin = composeServiceConfig.StdinOpen
			serviceConfig.Tty = composeServiceConfig.Tty

			komposeObject.ServiceConfigs[name] = serviceConfig
		}
	}

	return komposeObject
}

func handleServiceType(ServiceType string) string {
	switch strings.ToLower(ServiceType) {
	case "", "clusterip":
		return string(api.ServiceTypeClusterIP)
	case "nodeport":
		return string(api.ServiceTypeNodePort)
	case "loadbalancer":
		return string(api.ServiceTypeLoadBalancer)
	default:
		logrus.Fatalf("Unknown value '%s', supported values are 'NodePort, ClusterIP or LoadBalancer'", ServiceType)
		return ""
	}
}
