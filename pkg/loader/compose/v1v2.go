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
	"github.com/spf13/cast"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
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
// also load `expose` here
func loadPorts(composePorts []string, expose []string) ([]kobject.Ports, error) {
	ports := []kobject.Ports{}
	character := ":"
	exist := map[string]bool{}

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

	// load remain expose ports
	for _, port := range ports {
		// must use cast...
		exist[cast.ToString(port.ContainerPort)+string(port.Protocol)] = true
	}

	if expose != nil {
		for _, port := range expose {
			portValue := port
			protocol := api.ProtocolTCP
			if strings.Contains(portValue, "/") {
				splits := strings.Split(port, "/")
				portValue = splits[0]
				protocol = api.Protocol(strings.ToUpper(splits[1]))
			}

			if !exist[portValue+string(protocol)] {
				ports = append(ports, kobject.Ports{
					ContainerPort: cast.ToInt32(portValue),
					Protocol:      protocol,
				})
			}
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
		newName := normalizeContainerNames(composeServiceConfig.ContainerName)
		serviceConfig.ContainerName = newName
		if newName != composeServiceConfig.ContainerName {
			log.Infof("Container name in service %q has been changed from %q to %q", name, composeServiceConfig.ContainerName, newName)
		}
		serviceConfig.Command = composeServiceConfig.Entrypoint
		serviceConfig.HostName = composeServiceConfig.Hostname
		serviceConfig.DomainName = composeServiceConfig.DomainName
		serviceConfig.Args = composeServiceConfig.Command
		serviceConfig.Dockerfile = composeServiceConfig.Build.Dockerfile
		serviceConfig.BuildArgs = composeServiceConfig.Build.Args
		serviceConfig.Expose = composeServiceConfig.Expose

		envs := loadEnvVars(composeServiceConfig.Environment)
		serviceConfig.Environment = envs

		// Validate dockerfile path
		if filepath.IsAbs(serviceConfig.Dockerfile) {
			log.Fatalf("%q defined in service %q is an absolute path, it must be a relative path.", serviceConfig.Dockerfile, name)
		}

		// load ports, same as v3, we also load `expose`
		ports, err := loadPorts(composeServiceConfig.Ports, serviceConfig.Expose)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "loadPorts failed. "+name+" failed to load ports from compose file")
		}
		serviceConfig.Port = ports

		serviceConfig.WorkingDir = composeServiceConfig.WorkingDir

		if composeServiceConfig.Volumes != nil {
			for _, volume := range composeServiceConfig.Volumes.Volumes {
				v := volume.String()
				serviceConfig.VolList = append(serviceConfig.VolList, v)
			}
		}

		// canonical "Custom Labels" handler
		// Labels used to influence conversion of kompose will be handled
		// from here for docker-compose. Each loader will have such handler.
		if err := parseKomposeLabels(composeServiceConfig.Labels, &serviceConfig); err != nil {
			return kobject.KomposeObject{}, err
		}

		err = checkLabelsPorts(len(serviceConfig.Port), composeServiceConfig.Labels[LabelServiceType], name)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "kompose.service.type can't be set if service doesn't expose any ports.")
		}

		// convert compose labels to annotations
		serviceConfig.Annotations = map[string]string(composeServiceConfig.Labels)
		serviceConfig.CPUQuota = int64(composeServiceConfig.CPUQuota)
		serviceConfig.CapAdd = composeServiceConfig.CapAdd
		serviceConfig.CapDrop = composeServiceConfig.CapDrop
		serviceConfig.Pid = composeServiceConfig.Pid

		serviceConfig.Privileged = composeServiceConfig.Privileged
		serviceConfig.User = composeServiceConfig.User
		serviceConfig.VolumesFrom = composeServiceConfig.VolumesFrom
		serviceConfig.Stdin = composeServiceConfig.StdinOpen
		serviceConfig.Tty = composeServiceConfig.Tty
		serviceConfig.MemLimit = composeServiceConfig.MemLimit
		serviceConfig.TmpFs = composeServiceConfig.Tmpfs
		serviceConfig.StopGracePeriod = composeServiceConfig.StopGracePeriod

		// pretty much same as v3
		serviceConfig.Restart = composeServiceConfig.Restart
		if serviceConfig.Restart == "unless-stopped" {
			log.Warnf("Restart policy 'unless-stopped' in service %s is not supported, convert it to 'always'", name)
			serviceConfig.Restart = "always"
		}

		if composeServiceConfig.Networks != nil {
			if len(composeServiceConfig.Networks.Networks) > 0 {
				for _, value := range composeServiceConfig.Networks.Networks {
					if value.Name != "default" {
						serviceConfig.Network = append(serviceConfig.Network, value.RealName)
					}
				}
			}
		}
		// Get GroupAdd, group should be mentioned in gid format but not the group name
		groupAdd, err := getGroupAdd(composeServiceConfig.GroupAdd)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "GroupAdd should be mentioned in gid format, not a group name")
		}
		serviceConfig.GroupAdd = groupAdd

		komposeObject.ServiceConfigs[normalizeServiceNames(name)] = serviceConfig
		if normalizeServiceNames(name) != name {
			log.Infof("Service name in docker-compose has been changed from %q to %q", name, normalizeServiceNames(name))
		}
	}

	// This will handle volume at earlier stage itself, it will resolves problems occurred due to `volumes_from` key
	handleVolume(&komposeObject)

	return komposeObject, nil
}

// This function will retrieve volumes for each service, as well as it will parse volume information and store it in Volumes struct
func handleVolume(komposeObject *kobject.KomposeObject) {
	for name := range komposeObject.ServiceConfigs {
		// retrieve volumes of service
		vols, err := retrieveVolume(name, *komposeObject)
		if err != nil {
			errors.Wrap(err, "could not retrieve volume")
		}
		// We can't assign value to struct field in map while iterating over it, so temporary variable `temp` is used here
		var temp = komposeObject.ServiceConfigs[name]
		temp.Volumes = vols
		komposeObject.ServiceConfigs[name] = temp
	}
}

func checkLabelsPorts(noOfPort int, labels string, svcName string) error {
	if noOfPort == 0 && (labels == "NodePort" || labels == "LoadBalancer") {
		return errors.Errorf("%s defined in service %s with no ports present. Issues may occur when bringing up artifacts.", labels, svcName)
	}
	return nil
}

// returns all volumes associated with service, if `volumes_from` key is used, we have to retrieve volumes from the services which are mentioned there. Hence, recursive function is used here.
func retrieveVolume(svcName string, komposeObject kobject.KomposeObject) (volume []kobject.Volumes, err error) {
	// if volumes-from key is present
	if komposeObject.ServiceConfigs[svcName].VolumesFrom != nil {
		// iterating over services from `volumes-from`
		for _, depSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
			// recursive call for retrieving volumes of services from `volumes-from`
			dVols, err := retrieveVolume(depSvc, komposeObject)
			if err != nil {
				return nil, errors.Wrapf(err, "could not retrieve the volume")
			}
			var cVols []kobject.Volumes
			cVols, err = ParseVols(komposeObject.ServiceConfigs[svcName].VolList, svcName)
			if err != nil {
				return nil, errors.Wrapf(err, "error generating current volumes")
			}

			for _, cv := range cVols {
				// check whether volumes of current service is same or not as that of dependent volumes coming from `volumes-from`
				ok, dv := getVol(cv, dVols)
				if ok {
					// change current volumes service name to dependent service name
					if dv.VFrom == "" {
						cv.VFrom = dv.SvcName
						cv.SvcName = dv.SvcName
					} else {
						cv.VFrom = dv.VFrom
						cv.SvcName = dv.SvcName
					}
					cv.PVCName = dv.PVCName
				}
				volume = append(volume, cv)

			}
			// iterating over dependent volumes
			for _, dv := range dVols {
				// check whether dependent volume is already present or not
				if checkVolDependent(dv, volume) {
					// if found, add service name to `VFrom`
					dv.VFrom = dv.SvcName
					volume = append(volume, dv)
				}
			}
		}
	} else {
		// if `volumes-from` is not present
		volume, err = ParseVols(komposeObject.ServiceConfigs[svcName].VolList, svcName)
		if err != nil {
			return nil, errors.Wrapf(err, "error generating current volumes")
		}
	}
	return
}

// checkVolDependent returns false if dependent volume is present
func checkVolDependent(dv kobject.Volumes, volume []kobject.Volumes) bool {
	for _, vol := range volume {
		if vol.PVCName == dv.PVCName {
			return false
		}
	}
	return true

}

// ParseVols parse volumes
func ParseVols(volNames []string, svcName string) ([]kobject.Volumes, error) {
	var volumes []kobject.Volumes
	var err error

	for i, vn := range volNames {
		var v kobject.Volumes
		v.VolumeName, v.Host, v.Container, v.Mode, err = transformer.ParseVolume(vn)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse volume %q: %v", vn, err)
		}
		v.VolumeName = normalizeVolumes(v.VolumeName)
		v.SvcName = svcName
		v.MountPath = fmt.Sprintf("%s:%s", v.Host, v.Container)
		v.PVCName = fmt.Sprintf("%s-claim%d", v.SvcName, i)
		volumes = append(volumes, v)
	}

	return volumes, nil
}

// for dependent volumes, returns true and the respective volume if mountpath are same
func getVol(toFind kobject.Volumes, Vols []kobject.Volumes) (bool, kobject.Volumes) {
	for _, dv := range Vols {
		if toFind.MountPath == dv.MountPath {
			return true, dv
		}
	}
	return false, kobject.Volumes{}
}

// getGroupAdd will return group in int64 format
func getGroupAdd(group []string) ([]int64, error) {
	var groupAdd []int64
	for _, i := range group {
		j, err := strconv.Atoi(i)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get group_add key")
		}
		groupAdd = append(groupAdd, int64(j))

	}
	return groupAdd, nil
}
