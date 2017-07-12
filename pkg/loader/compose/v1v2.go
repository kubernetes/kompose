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
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
)

// Parse Docker Compose with libcompose (only supports v1 and v2). Eventually we will
// switch to using only libcompose once v3 is supported.
func parseV1V2(files []string) (kobject.KomposeObject, error) {

	// Gather the appropriate context for parsing
	context := &project.Context{}
	context.ComposeFiles = files

	if context.ResourceLookup == nil {
		context.ResourceLookup = &lookup.FileResourceLookup{}
	}

	if context.EnvironmentLookup == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return kobject.KomposeObject{}, nil
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

	// Load the context and let's start parsing
	composeObject := project.NewProject(context, nil, nil)
	err := composeObject.Parse()
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "composeObject.Parse() failed, Failed to load compose file")
	}

	noSupKeys := checkUnsupportedKey(composeObject)
	for _, keyName := range noSupKeys {
		log.Warningf("Unsupported %s key - ignoring", keyName)
	}

	// Map the parsed struct to a struct we understand (kobject)
	komposeObject, err := libComposeToKomposeMapping(composeObject)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	return komposeObject, nil
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

// Uses libcompose's APIProject type and converts it to a Kompose object for us to understand
func libComposeToKomposeMapping(composeObject *project.Project) (kobject.KomposeObject, error) {

	// Initialize what's going to be returned
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
		LoadedFrom:     "compose",
	}

	// Here we "clean up" the service configuration so we return something that includes
	// all relevant information as well as avoid the unsupported keys as well.
	for name, composeServiceConfig := range composeObject.ServiceConfigs.All() {
		serviceConfig := kobject.ServiceConfig{}
		serviceConfig.Image = composeServiceConfig.Image
		serviceConfig.Build = composeServiceConfig.Build.Context
		newName := normalizeServiceNames(composeServiceConfig.ContainerName)
		serviceConfig.ContainerName = newName
		if newName != composeServiceConfig.ContainerName {
			log.Infof("Container name in service %q has been changed from %q to %q", name, composeServiceConfig.ContainerName, newName)
		}
		serviceConfig.Command = composeServiceConfig.Entrypoint
		serviceConfig.Args = composeServiceConfig.Command
		serviceConfig.Dockerfile = composeServiceConfig.Build.Dockerfile
		serviceConfig.BuildArgs = composeServiceConfig.Build.Args

		envs := loadEnvVars(composeServiceConfig.Environment)
		serviceConfig.Environment = envs

		//Validate dockerfile path
		if filepath.IsAbs(serviceConfig.Dockerfile) {
			log.Fatalf("%q defined in service %q is an absolute path, it must be a relative path.", serviceConfig.Dockerfile, name)
		}

		// load ports
		ports, err := loadPorts(composeServiceConfig.Ports)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "loadPorts failed. "+name+" failed to load ports from compose file")
		}
		serviceConfig.Port = ports

		serviceConfig.WorkingDir = composeServiceConfig.WorkingDir

		if composeServiceConfig.Volumes != nil {
			for _, volume := range composeServiceConfig.Volumes.Volumes {
				v := normalizeServiceNames(volume.String())
				serviceConfig.Volumes = append(serviceConfig.Volumes, v)
			}
		}

		// canonical "Custom Labels" handler
		// Labels used to influence conversion of kompose will be handled
		// from here for docker-compose. Each loader will have such handler.
		for key, value := range composeServiceConfig.Labels {
			switch key {
			case "kompose.service.type":
				serviceType, err := handleServiceType(value)
				if err != nil {
					return kobject.KomposeObject{}, errors.Wrap(err, "handleServiceType failed")
				}

				serviceConfig.ServiceType = serviceType
			case "kompose.service.expose":
				serviceConfig.ExposeService = strings.ToLower(value)
			}
		}
		err = checkLabelsPorts(len(serviceConfig.Port), composeServiceConfig.Labels["kompose.service.type"], name)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "kompose.service.type can't be set if service doesn't expose any ports.")
		}

		// convert compose labels to annotations
		serviceConfig.Annotations = map[string]string(composeServiceConfig.Labels)
		serviceConfig.CPUQuota = int64(composeServiceConfig.CPUQuota)
		serviceConfig.CapAdd = composeServiceConfig.CapAdd
		serviceConfig.CapDrop = composeServiceConfig.CapDrop
		serviceConfig.Pid = composeServiceConfig.Pid
		serviceConfig.Expose = composeServiceConfig.Expose
		serviceConfig.Privileged = composeServiceConfig.Privileged
		serviceConfig.Restart = composeServiceConfig.Restart
		serviceConfig.User = composeServiceConfig.User
		serviceConfig.VolumesFrom = composeServiceConfig.VolumesFrom
		serviceConfig.Stdin = composeServiceConfig.StdinOpen
		serviceConfig.Tty = composeServiceConfig.Tty
		serviceConfig.MemLimit = composeServiceConfig.MemLimit
		serviceConfig.TmpFs = composeServiceConfig.Tmpfs
		serviceConfig.StopGracePeriod = composeServiceConfig.StopGracePeriod
		komposeObject.ServiceConfigs[normalizeServiceNames(name)] = serviceConfig
		if normalizeServiceNames(name) != name {
			log.Infof("Service name in docker-compose has been changed from %q to %q", name, normalizeServiceNames(name))
		}
	}
	return komposeObject, nil
}

func checkLabelsPorts(noOfPort int, labels string, svcName string) error {
	if noOfPort == 0 && (labels == "NodePort" || labels == "LoadBalancer") {
		return errors.Errorf("%s defined in service %s with no ports present. Issues may occur when bringing up artifacts.", labels, svcName)
	}
	return nil
}
