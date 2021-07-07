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
	"sort"
	"strconv"
	"strings"

	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	api "k8s.io/api/core/v1"
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
	kp := []kobject.Ports{}
	exist := map[string]bool{}

	for _, cp := range composePorts {
		var hostIP string

		if parts := strings.Split(cp, ":"); len(parts) == 3 {
			if ip := net.ParseIP(parts[0]); ip.To4() == nil && ip.To16() == nil {
				return nil, fmt.Errorf("%q contains an invalid IPv4 or IPv6 IP address", parts[0])
			}
			hostIP = parts[0]
		}

		np, pbs, err := nat.ParsePortSpecs([]string{cp})
		if err != nil {
			return nil, fmt.Errorf("invalid port, error = %v", err)
		}
		// Force HostIP value to avoid warning raised by github.com/docker/cli/opts
		// The opts package will warn if the bindings contains host IP except
		// 0.0.0.0. However, the message is not useful in this case since the value
		// should be handled by kompose properly.
		for _, pb := range pbs {
			for i, p := range pb {
				p.HostIP = ""
				pb[i] = p
			}
		}

		var ports []string
		for p := range np {
			ports = append(ports, string(p))
		}
		sort.Strings(ports)

		for _, p := range ports {
			pc, err := opts.ConvertPortToPortConfig(nat.Port(p), pbs)
			if err != nil {
				return nil, fmt.Errorf("invalid port, error = %v", err)
			}
			for _, cfg := range pc {
				kp = append(kp, kobject.Ports{
					HostPort:      int32(cfg.PublishedPort),
					ContainerPort: int32(cfg.TargetPort),
					HostIP:        hostIP,
					Protocol:      api.Protocol(strings.ToUpper(string(cfg.Protocol))),
				})
			}
		}
	}

	// load remain expose ports
	for _, p := range kp {
		// must use cast...
		exist[cast.ToString(p.ContainerPort)+string(p.Protocol)] = true
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
				kp = append(kp, kobject.Ports{
					ContainerPort: cast.ToInt32(portValue),
					Protocol:      protocol,
				})
			}
		}
	}

	return kp, nil
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
						nomalizedNetworkName, err := normalizeNetworkNames(value.RealName)
						if err != nil {
							return kobject.KomposeObject{}, errors.Wrap(err, "Error trying to normalize network names")
						}
						if nomalizedNetworkName != value.RealName {
							log.Warnf("Network name in docker-compose has been changed from %q to %q", value.RealName, nomalizedNetworkName)
						}
						serviceConfig.Network = append(serviceConfig.Network, nomalizedNetworkName)
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
