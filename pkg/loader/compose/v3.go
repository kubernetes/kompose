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
	"github.com/spf13/cast"
	"strconv"
	"strings"
	"time"

	libcomposeyaml "github.com/docker/libcompose/yaml"

	"k8s.io/kubernetes/pkg/api"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"

	"os"

	"fmt"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// converts os.Environ() ([]string) to map[string]string
// based on https://github.com/docker/cli/blob/5dd30732a23bbf14db1c64d084ae4a375f592cfa/cli/command/stack/deploy_composefile.go#L143
func buildEnvironment() (map[string]string, error) {
	env := os.Environ()
	result := make(map[string]string, len(env))
	for _, s := range env {
		// if value is empty, s is like "K=", not "K".
		if !strings.Contains(s, "=") {
			return result, errors.Errorf("unexpected environment %q", s)
		}
		kv := strings.SplitN(s, "=", 2)
		result[kv[0]] = kv[1]
	}
	return result, nil
}

// The purpose of this is not to deploy, but to be able to parse
// v3 of Docker Compose into a suitable format. In this case, whatever is returned
// by docker/cli's ServiceConfig
func parseV3(files []string) (kobject.KomposeObject, error) {

	// In order to get V3 parsing to work, we have to go through some preliminary steps
	// for us to hack up github.com/docker/cli in order to correctly convert to a kobject.KomposeObject

	// Gather the working directory
	workingDir, err := getComposeFileDir(files)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	// get environment variables
	env, err := buildEnvironment()
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "cannot build environment variables")
	}

	var config *types.Config
	for _, file := range files {
		// Load and then parse the YAML first!
		loadedFile, err := ReadFile(file)
		if err != nil {
			return kobject.KomposeObject{}, err
		}

		// Parse the Compose File
		parsedComposeFile, err := loader.ParseYAML(loadedFile)
		if err != nil {
			return kobject.KomposeObject{}, err
		}

		// Config file
		configFile := types.ConfigFile{
			Filename: file,
			Config:   parsedComposeFile,
		}

		// Config details
		configDetails := types.ConfigDetails{
			WorkingDir:  workingDir,
			ConfigFiles: []types.ConfigFile{configFile},
			Environment: env,
		}

		// Actual config
		// We load it in order to retrieve the parsed output configuration!
		// This will output a github.com/docker/cli ServiceConfig
		// Which is similar to our version of ServiceConfig
		currentConfig, err := loader.Load(configDetails)
		if err != nil {
			return kobject.KomposeObject{}, err
		}
		if config == nil {
			config = currentConfig
		} else {
			config, err = mergeComposeObject(config, currentConfig)
			if err != nil {
				return kobject.KomposeObject{}, err
			}
		}
	}

	noSupKeys := checkUnsupportedKeyForV3(config)
	for _, keyName := range noSupKeys {
		log.Warningf("Unsupported %s key - ignoring", keyName)
	}

	// Finally, we convert the object from docker/cli's ServiceConfig to our appropriate one
	komposeObject, err := dockerComposeToKomposeMapping(config)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	return komposeObject, nil
}

func loadV3Placement(constraints []string) map[string]string {
	placement := make(map[string]string)
	errMsg := " constraints in placement is not supported, only 'node.hostname', 'engine.labels.operatingsystem' and 'node.labels.xxx' (ex: node.labels.something == anything) is supported as a constraint "
	for _, j := range constraints {
		p := strings.Split(j, " == ")
		if len(p) < 2 {
			log.Warn(p[0], errMsg)
			continue
		}
		if p[0] == "node.hostname" {
			placement["kubernetes.io/hostname"] = p[1]
		} else if p[0] == "engine.labels.operatingsystem" {
			placement["beta.kubernetes.io/os"] = p[1]
		} else if strings.HasPrefix(p[0], "node.labels.") {
			label := strings.TrimPrefix(p[0], "node.labels.")
			placement[label] = p[1]
		} else {
			log.Warn(p[0], errMsg)
		}
	}
	return placement
}

// Convert the Docker Compose v3 volumes to []string (the old way)
// TODO: Check to see if it's a "bind" or "volume". Ignore for now.
// TODO: Refactor it similar to loadV3Ports
// See: https://docs.docker.com/compose/compose-file/#long-syntax-3
func loadV3Volumes(volumes []types.ServiceVolumeConfig) []string {

	var volArray []string
	for _, vol := range volumes {
		// There will *always* be Source when parsing
		v := vol.Source

		if vol.Target != "" {
			v = v + ":" + vol.Target
		}

		if vol.ReadOnly {
			v = v + ":ro"
		}

		volArray = append(volArray, v)
	}
	return volArray
}

// Convert Docker Compose v3 ports to kobject.Ports
// expose ports will be treated as TCP ports
func loadV3Ports(ports []types.ServicePortConfig, expose []string) []kobject.Ports {
	komposePorts := []kobject.Ports{}

	exist := map[string]bool{}

	for _, port := range ports {

		// Convert to a kobject struct with ports
		// NOTE: V3 doesn't use IP (they utilize Swarm instead for host-networking).
		// Thus, IP is blank.
		komposePorts = append(komposePorts, kobject.Ports{
			HostPort:      int32(port.Published),
			ContainerPort: int32(port.Target),
			HostIP:        "",
			Protocol:      api.Protocol(strings.ToUpper(string(port.Protocol))),
		})

		exist[cast.ToString(port.Target)+strings.ToUpper(string(port.Protocol))] = true

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

			if exist[portValue+string(protocol)] {
				continue
			}
			komposePorts = append(komposePorts, kobject.Ports{
				HostPort:      cast.ToInt32(portValue),
				ContainerPort: cast.ToInt32(portValue),
				HostIP:        "",
				Protocol:      protocol,
			})
		}
	}

	return komposePorts
}

/* Convert the HealthCheckConfig as designed by Docker to
a Kubernetes-compatible format.
*/
func parseHealthCheck(composeHealthCheck types.HealthCheckConfig) (kobject.HealthCheck, error) {

	var timeout, interval, retries, startPeriod int32

	// Here we convert the timeout from 1h30s (example) to 36030 seconds.
	if composeHealthCheck.Timeout != nil {
		parse, err := time.ParseDuration(composeHealthCheck.Timeout.String())
		if err != nil {
			return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check timeout variable")
		}
		timeout = int32(parse.Seconds())
	}

	if composeHealthCheck.Interval != nil {
		parse, err := time.ParseDuration(composeHealthCheck.Interval.String())
		if err != nil {
			return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check interval variable")
		}
		interval = int32(parse.Seconds())
	}

	if composeHealthCheck.Retries != nil {
		retries = int32(*composeHealthCheck.Retries)
	}

	if composeHealthCheck.StartPeriod != nil {
		parse, err := time.ParseDuration(composeHealthCheck.StartPeriod.String())
		if err != nil {
			return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check startPeriod variable")
		}
		startPeriod = int32(parse.Seconds())
	}

	// Due to docker/cli adding "CMD-SHELL" to the struct, we remove the first element of composeHealthCheck.Test
	return kobject.HealthCheck{
		Test:        composeHealthCheck.Test[1:],
		Timeout:     timeout,
		Interval:    interval,
		Retries:     retries,
		StartPeriod: startPeriod,
	}, nil
}

func dockerComposeToKomposeMapping(composeObject *types.Config) (kobject.KomposeObject, error) {

	// Step 1. Initialize what's going to be returned
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
		LoadedFrom:     "compose",
		Secrets:        composeObject.Secrets,
	}

	// Step 2. Parse through the object and convert it to kobject.KomposeObject!
	// Here we "clean up" the service configuration so we return something that includes
	// all relevant information as well as avoid the unsupported keys as well.
	for _, composeServiceConfig := range composeObject.Services {

		// Standard import
		// No need to modify before importation
		name := composeServiceConfig.Name
		serviceConfig := kobject.ServiceConfig{}
		serviceConfig.Image = composeServiceConfig.Image
		serviceConfig.WorkingDir = composeServiceConfig.WorkingDir
		serviceConfig.Annotations = map[string]string(composeServiceConfig.Labels)
		serviceConfig.CapAdd = composeServiceConfig.CapAdd
		serviceConfig.CapDrop = composeServiceConfig.CapDrop
		serviceConfig.Expose = composeServiceConfig.Expose
		serviceConfig.Privileged = composeServiceConfig.Privileged
		serviceConfig.User = composeServiceConfig.User
		serviceConfig.Stdin = composeServiceConfig.StdinOpen
		serviceConfig.Tty = composeServiceConfig.Tty
		serviceConfig.TmpFs = composeServiceConfig.Tmpfs
		serviceConfig.ContainerName = normalizeContainerNames(composeServiceConfig.ContainerName)
		serviceConfig.Command = composeServiceConfig.Entrypoint
		serviceConfig.Args = composeServiceConfig.Command
		serviceConfig.Labels = composeServiceConfig.Labels
		serviceConfig.HostName = composeServiceConfig.Hostname
		serviceConfig.DomainName = composeServiceConfig.DomainName
		serviceConfig.Secrets = composeServiceConfig.Secrets

		parseV3Network(&composeServiceConfig, &serviceConfig, composeObject)

		if err := parseV3Resources(&composeServiceConfig, &serviceConfig); err != nil {
			return kobject.KomposeObject{}, err
		}

		// Deploy keys
		// mode:
		serviceConfig.DeployMode = composeServiceConfig.Deploy.Mode
		// labels
		serviceConfig.DeployLabels = composeServiceConfig.Deploy.Labels

		// HealthCheck
		if composeServiceConfig.HealthCheck != nil && !composeServiceConfig.HealthCheck.Disable {
			var err error
			serviceConfig.HealthChecks, err = parseHealthCheck(*composeServiceConfig.HealthCheck)
			if err != nil {
				return kobject.KomposeObject{}, errors.Wrap(err, "Unable to parse health check")
			}
		}

		// restart-policy: deploy.restart_policy.condition will rewrite restart option
		// see: https://docs.docker.com/compose/compose-file/#restart_policy
		serviceConfig.Restart = composeServiceConfig.Restart
		if composeServiceConfig.Deploy.RestartPolicy != nil {
			serviceConfig.Restart = composeServiceConfig.Deploy.RestartPolicy.Condition
		}
		if serviceConfig.Restart == "unless-stopped" {
			log.Warnf("Restart policy 'unless-stopped' in service %s is not supported, convert it to 'always'", name)
			serviceConfig.Restart = "always"
		}

		// replicas:
		if composeServiceConfig.Deploy.Replicas != nil {
			serviceConfig.Replicas = int(*composeServiceConfig.Deploy.Replicas)
		}

		// placement:
		serviceConfig.Placement = loadV3Placement(composeServiceConfig.Deploy.Placement.Constraints)

		if composeServiceConfig.Deploy.UpdateConfig != nil {
			serviceConfig.DeployUpdateConfig = *composeServiceConfig.Deploy.UpdateConfig
		}

		// TODO: Build is not yet supported, see:
		// https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L9
		// We will have to *manually* add this / parse.
		serviceConfig.Build = composeServiceConfig.Build.Context
		serviceConfig.Dockerfile = composeServiceConfig.Build.Dockerfile
		serviceConfig.BuildArgs = composeServiceConfig.Build.Args
		serviceConfig.BuildLabels = composeServiceConfig.Build.Labels

		// env
		parseV3Environment(&composeServiceConfig, &serviceConfig)

		// Get env_file
		serviceConfig.EnvFile = composeServiceConfig.EnvFile

		// Parse the ports
		// v3 uses a new format called "long syntax" starting in 3.2
		// https://docs.docker.com/compose/compose-file/#ports

		// here we will translate `expose` too, they basically means the same thing in kubernetes
		serviceConfig.Port = loadV3Ports(composeServiceConfig.Ports, serviceConfig.Expose)

		// Parse the volumes
		// Again, in v3, we use the "long syntax" for volumes in terms of parsing
		// https://docs.docker.com/compose/compose-file/#long-syntax-3
		serviceConfig.VolList = loadV3Volumes(composeServiceConfig.Volumes)

		if err := parseKomposeLabels(composeServiceConfig.Labels, &serviceConfig); err != nil {
			return kobject.KomposeObject{}, err
		}

		// Log if the name will been changed
		if normalizeServiceNames(name) != name {
			log.Infof("Service name in docker-compose has been changed from %q to %q", name, normalizeServiceNames(name))
		}

		serviceConfig.Configs = composeServiceConfig.Configs
		serviceConfig.ConfigsMetaData = composeObject.Configs
		if composeServiceConfig.Deploy.EndpointMode == "vip" {
			serviceConfig.ServiceType = string(api.ServiceTypeNodePort)
		}
		// Final step, add to the array!
		komposeObject.ServiceConfigs[normalizeServiceNames(name)] = serviceConfig
	}

	handleV3Volume(&komposeObject, &composeObject.Volumes)

	return komposeObject, nil
}

func parseV3Network(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig, composeObject *types.Config) {
	if len(composeServiceConfig.Networks) == 0 {
		if defaultNetwork, ok := composeObject.Networks["default"]; ok {
			serviceConfig.Network = append(serviceConfig.Network, defaultNetwork.Name)
		}
	} else {
		var alias = ""
		for key := range composeServiceConfig.Networks {
			alias = key
			netName := composeObject.Networks[alias].Name
			// if Network Name Field is empty in the docker-compose definition
			// we will use the alias name defined in service config file
			if netName == "" {
				netName = alias
			}
			serviceConfig.Network = append(serviceConfig.Network, netName)
		}
	}
}

func parseV3Resources(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig) error {
	if (composeServiceConfig.Deploy.Resources != types.Resources{}) {

		// memory:
		// TODO: Refactor yaml.MemStringorInt in kobject.go to int64
		// cpu:
		// convert to k8s format, for example: 0.5 = 500m
		// See: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
		// "The expression 0.1 is equivalent to the expression 100m, which can be read as “one hundred millicpu”."

		// Since Deploy.Resources.Limits does not initialize, we must check type Resources before continuing
		if composeServiceConfig.Deploy.Resources.Limits != nil {
			serviceConfig.MemLimit = libcomposeyaml.MemStringorInt(composeServiceConfig.Deploy.Resources.Limits.MemoryBytes)

			if composeServiceConfig.Deploy.Resources.Limits.NanoCPUs != "" {
				cpuLimit, err := strconv.ParseFloat(composeServiceConfig.Deploy.Resources.Limits.NanoCPUs, 64)
				if err != nil {
					return errors.Wrap(err, "Unable to convert cpu limits resources value")
				}
				serviceConfig.CPULimit = int64(cpuLimit * 1000)
			}
		}
		if composeServiceConfig.Deploy.Resources.Reservations != nil {
			serviceConfig.MemReservation = libcomposeyaml.MemStringorInt(composeServiceConfig.Deploy.Resources.Reservations.MemoryBytes)

			if composeServiceConfig.Deploy.Resources.Reservations.NanoCPUs != "" {
				cpuReservation, err := strconv.ParseFloat(composeServiceConfig.Deploy.Resources.Reservations.NanoCPUs, 64)
				if err != nil {
					return errors.Wrap(err, "Unable to convert cpu limits reservation value")
				}
				serviceConfig.CPUReservation = int64(cpuReservation * 1000)
			}
		}
	}
	return nil

}

func parseV3Environment(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig) {
	// Gather the environment values
	// DockerCompose uses map[string]*string while we use []string
	// So let's convert that using this hack
	// Note: unset env pick up the env value on host if exist
	for name, value := range composeServiceConfig.Environment {
		var env kobject.EnvVar
		if value != nil {
			env = kobject.EnvVar{Name: name, Value: *value}
		} else {
			result, ok := os.LookupEnv(name)
			if ok {
				env = kobject.EnvVar{Name: name, Value: result}
			} else {
				continue
			}
		}
		serviceConfig.Environment = append(serviceConfig.Environment, env)
	}
}

// parseKomposeLabels parse kompose labels, also do some validation
func parseKomposeLabels(labels map[string]string, serviceConfig *kobject.ServiceConfig) error {
	// Label handler
	// Labels used to influence conversion of kompose will be handled
	// from here for docker-compose. Each loader will have such handler.

	if serviceConfig.Labels == nil {
		serviceConfig.Labels = make(map[string]string)
	}

	for key, value := range labels {
		switch key {
		case LabelServiceType:
			serviceType, err := handleServiceType(value)
			if err != nil {
				return errors.Wrap(err, "handleServiceType failed")
			}

			serviceConfig.ServiceType = serviceType
		case LabelServiceExpose:
			serviceConfig.ExposeService = strings.Trim(strings.ToLower(value), " ,")
		case LabelNodePortPort:
			serviceConfig.NodePortPort = cast.ToInt32(value)
		case LabelServiceExposeTLSSecret:
			serviceConfig.ExposeServiceTLS = value
		case LabelImagePullSecret:
			serviceConfig.ImagePullSecret = value
		case LabelImagePullPolicy:
			serviceConfig.ImagePullPolicy = value
		default:
			serviceConfig.Labels[key] = value
		}
	}

	if serviceConfig.ExposeService == "" && serviceConfig.ExposeServiceTLS != "" {
		return errors.New("kompose.service.expose.tls-secret was specified without kompose.service.expose")
	}

	if serviceConfig.ServiceType != string(api.ServiceTypeNodePort) && serviceConfig.NodePortPort != 0 {
		return errors.New("kompose.service.type must be nodeport when assign node port value")
	}

	if len(serviceConfig.Port) > 1 && serviceConfig.NodePortPort != 0 {
		return errors.New("cannot set kompose.service.nodeport.port when service has multiple ports")
	}

	return nil
}

func handleV3Volume(komposeObject *kobject.KomposeObject, volumes *map[string]types.VolumeConfig) {
	for name := range komposeObject.ServiceConfigs {
		// retrieve volumes of service
		vols, err := retrieveVolume(name, *komposeObject)
		if err != nil {
			errors.Wrap(err, "could not retrieve vvolume")
		}
		for volName, vol := range vols {
			size, selector := getV3VolumeLabels(vol.VolumeName, volumes)
			if len(size) > 0 || len(selector) > 0 {
				// We can't assign value to struct field in map while iterating over it, so temporary variable `temp` is used here
				var temp = vols[volName]
				temp.PVCSize = size
				temp.SelectorValue = selector
				vols[volName] = temp
			}
		}
		// We can't assign value to struct field in map while iterating over it, so temporary variable `temp` is used here
		var temp = komposeObject.ServiceConfigs[name]
		temp.Volumes = vols
		komposeObject.ServiceConfigs[name] = temp
	}
}

func getV3VolumeLabels(name string, volumes *map[string]types.VolumeConfig) (string, string) {
	size, selector := "", ""

	if volume, ok := (*volumes)[name]; ok {
		for key, value := range volume.Labels {
			if key == "kompose.volume.size" {
				size = value
			} else if key == "kompose.volume.selector" {
				selector = value
			}
		}
	}

	return size, selector
}

func mergeComposeObject(oldCompose *types.Config, newCompose *types.Config) (*types.Config, error) {
	if oldCompose == nil || newCompose == nil {
		return nil, fmt.Errorf("Merge multiple compose error, compose config is nil")
	}
	oldComposeServiceNameMap := make(map[string]int, len(oldCompose.Services))
	for index, service := range oldCompose.Services {
		oldComposeServiceNameMap[service.Name] = index
	}

	for _, service := range newCompose.Services {
		index := 0
		if tmpIndex, ok := oldComposeServiceNameMap[service.Name]; !ok {
			oldCompose.Services = append(oldCompose.Services, service)
			continue
		} else {
			index = tmpIndex
		}
		tmpOldService := oldCompose.Services[index]
		if service.Build.Dockerfile != "" {
			tmpOldService.Build = service.Build
		}
		if len(service.CapAdd) != 0 {
			tmpOldService.CapAdd = service.CapAdd
		}
		if len(service.CapDrop) != 0 {
			tmpOldService.CapDrop = service.CapDrop
		}
		if service.CgroupParent != "" {
			tmpOldService.CgroupParent = service.CgroupParent
		}
		if len(service.Command) != 0 {
			tmpOldService.Command = service.Command
		}
		if len(service.Configs) != 0 {
			tmpOldService.Configs = service.Configs
		}
		if service.ContainerName != "" {
			tmpOldService.ContainerName = service.ContainerName
		}
		if service.CredentialSpec.File != "" || service.CredentialSpec.Registry != "" {
			tmpOldService.CredentialSpec = service.CredentialSpec
		}
		if len(service.DependsOn) != 0 {
			tmpOldService.DependsOn = service.DependsOn
		}
		if service.Deploy.Mode != "" {
			tmpOldService.Deploy = service.Deploy
		}
		if service.Deploy.Resources.Limits != nil {
			tmpOldService.Deploy.Resources.Limits = service.Deploy.Resources.Limits
		}

		if service.Deploy.Resources.Reservations != nil {
			tmpOldService.Deploy.Resources.Reservations = service.Deploy.Resources.Reservations
		}

		if len(service.Devices) != 0 {
			// merge the 2 sets of values
			// TODO: need to merge the sets based on target values
			// Not implemented yet as we don't convert devices to k8s anyway
			tmpOldService.Devices = service.Devices
		}
		if len(service.DNS) != 0 {
			// concat the 2 sets of values
			tmpOldService.DNS = append(tmpOldService.DNS, service.DNS...)
		}
		if len(service.DNSSearch) != 0 {
			// concat the 2 sets of values
			tmpOldService.DNSSearch = append(tmpOldService.DNSSearch, service.DNSSearch...)
		}
		if service.DomainName != "" {
			tmpOldService.DomainName = service.DomainName
		}
		if len(service.Entrypoint) != 0 {
			tmpOldService.Entrypoint = service.Entrypoint
		}
		if len(service.Environment) != 0 {
			// merge the 2 sets of values
			for k, v := range service.Environment {
				tmpOldService.Environment[k] = v
			}
		}
		if len(service.EnvFile) != 0 {
			tmpOldService.EnvFile = service.EnvFile
		}
		if len(service.Expose) != 0 {
			// concat the 2 sets of values
			tmpOldService.Expose = append(tmpOldService.Expose, service.Expose...)
		}
		if len(service.ExternalLinks) != 0 {
			// concat the 2 sets of values
			tmpOldService.ExternalLinks = append(tmpOldService.ExternalLinks, service.ExternalLinks...)
		}
		if len(service.ExtraHosts) != 0 {
			tmpOldService.ExtraHosts = service.ExtraHosts
		}
		if service.Hostname != "" {
			tmpOldService.Hostname = service.Hostname
		}
		if service.HealthCheck != nil {
			tmpOldService.HealthCheck = service.HealthCheck
		}
		if service.Image != "" {
			tmpOldService.Image = service.Image
		}
		if service.Ipc != "" {
			tmpOldService.Ipc = service.Ipc
		}
		if len(service.Labels) != 0 {
			// merge the 2 sets of values
			for k, v := range service.Labels {
				tmpOldService.Labels[k] = v
			}
		}
		if len(service.Links) != 0 {
			tmpOldService.Links = service.Links
		}
		if service.Logging != nil {
			tmpOldService.Logging = service.Logging
		}
		if service.MacAddress != "" {
			tmpOldService.MacAddress = service.MacAddress
		}
		if service.NetworkMode != "" {
			tmpOldService.NetworkMode = service.NetworkMode
		}
		if len(service.Networks) != 0 {
			tmpOldService.Networks = service.Networks
		}
		if service.Pid != "" {
			tmpOldService.Pid = service.Pid
		}
		if len(service.Ports) != 0 {
			// concat the 2 sets of values
			tmpOldService.Ports = append(tmpOldService.Ports, service.Ports...)
		}
		if service.Privileged != tmpOldService.Privileged {
			tmpOldService.Privileged = service.Privileged
		}
		if service.ReadOnly != tmpOldService.ReadOnly {
			tmpOldService.ReadOnly = service.ReadOnly
		}
		if service.Restart != "" {
			tmpOldService.Restart = service.Restart
		}
		if len(service.Secrets) != 0 {
			tmpOldService.Secrets = service.Secrets
		}
		if len(service.SecurityOpt) != 0 {
			tmpOldService.SecurityOpt = service.SecurityOpt
		}
		if service.StdinOpen != tmpOldService.StdinOpen {
			tmpOldService.StdinOpen = service.StdinOpen
		}
		if service.StopGracePeriod != nil {
			tmpOldService.StopGracePeriod = service.StopGracePeriod
		}
		if service.StopSignal != "" {
			tmpOldService.StopSignal = service.StopSignal
		}
		if len(service.Tmpfs) != 0 {
			// concat the 2 sets of values
			tmpOldService.Tmpfs = append(tmpOldService.Tmpfs, service.Tmpfs...)
		}
		if service.Tty != tmpOldService.Tty {
			tmpOldService.Tty = service.Tty
		}
		if len(service.Ulimits) != 0 {
			tmpOldService.Ulimits = service.Ulimits
		}
		if service.User != "" {
			tmpOldService.User = service.User
		}
		if len(service.Volumes) != 0 {
			// merge the 2 sets of values

			// Store volumes by Target
			volumeConfigsMap := make(map[string]types.ServiceVolumeConfig)
			// map is not iterated in the same order.
			// but why only this code cause test error?
			var keys []string

			// populate the older values
			for _, volConfig := range tmpOldService.Volumes {
				volumeConfigsMap[volConfig.Target] = volConfig
				keys = append(keys, volConfig.Target)
			}
			// add the new values, overriding as needed
			for _, volConfig := range service.Volumes {
				if _, ok := volumeConfigsMap[volConfig.Target]; !ok {
					keys = append(keys, volConfig.Target)
				}
				volumeConfigsMap[volConfig.Target] = volConfig
			}

			// get the new list of volume configs
			var volumes []types.ServiceVolumeConfig

			for _, key := range keys {
				volumes = append(volumes, volumeConfigsMap[key])
			}

			tmpOldService.Volumes = volumes
		}
		if service.WorkingDir != "" {
			tmpOldService.WorkingDir = service.WorkingDir
		}
		oldCompose.Services[index] = tmpOldService
	}

	// Merge the networks information
	for idx, network := range newCompose.Networks {
		oldCompose.Networks[idx] = network
	}

	// Merge the volumes information
	for idx, volume := range newCompose.Volumes {
		oldCompose.Volumes[idx] = volume
	}

	// Merge the secrets information
	for idx, secret := range newCompose.Secrets {
		oldCompose.Secrets[idx] = secret
	}

	// Merge the configs information
	for idx, config := range newCompose.Configs {
		oldCompose.Configs[idx] = config
	}

	return oldCompose, nil
}

func checkUnsupportedKeyForV3(composeObject *types.Config) []string {
	if composeObject == nil {
		return []string{}
	}

	var keysFound []string

	for _, service := range composeObject.Services {
		for _, tmpConfig := range service.Configs {
			if tmpConfig.GID != "" {
				keysFound = append(keysFound, "long syntax config gid")
			}
			if tmpConfig.UID != "" {
				keysFound = append(keysFound, "long syntax config uid")
			}
		}

		if service.CredentialSpec.Registry != "" || service.CredentialSpec.File != "" {
			keysFound = append(keysFound, "credential_spec")

		}
	}

	for _, config := range composeObject.Configs {
		if config.External.External {
			keysFound = append(keysFound, "external config")
		}
	}

	return keysFound
}
