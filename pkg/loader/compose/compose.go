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
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/fatih/structs"
	"github.com/google/shlex"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	api "k8s.io/api/core/v1"
)

// StdinData is data bytes read from stdin
var StdinData []byte

// Compose is docker compose file loader, implements Loader interface
type Compose struct {
}

// checkUnsupportedKey checks if compose-go project contains
// keys that are not supported by this loader.
// list of all unsupported keys are stored in unsupportedKey variable
// returns list of unsupported YAML keys from docker-compose
func checkUnsupportedKey(composeProject *types.Project) []string {
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
	if _, ok := composeProject.Networks["default"]; ok && len(composeProject.Networks) == 1 {
		log.Debug("Default network found")
	}

	// Root level volumes are not yet supported
	if len(composeProject.Volumes) > 0 {
		keysFound = append(keysFound, "root level volumes")
	}

	for _, serviceConfig := range composeProject.AllServices() {
		// this reflection is used in check for empty arrays
		val := reflect.ValueOf(serviceConfig)
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
						if len(serviceConfig.Networks) == 1 && serviceConfig.NetworksByPriority()[0] == "default" {
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
	// Gather the working directory
	workingDir, err := getComposeFileDir(files)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	projectOptions, err := cli.NewProjectOptions(files, cli.WithOsEnv, cli.WithWorkingDirectory(workingDir), cli.WithInterpolation(false))
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "Unable to create compose options")
	}

	project, err := cli.ProjectFromOptions(projectOptions)
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "Unable to load files")
	}

	komposeObject, err := dockerComposeToKomposeMapping(project)
	if err != nil {
		return kobject.KomposeObject{}, err
	}
	return komposeObject, nil
}

func loadPlacement(placement types.Placement) kobject.Placement {
	komposePlacement := kobject.Placement{
		PositiveConstraints: make(map[string]string),
		NegativeConstraints: make(map[string]string),
		Preferences:         make([]string, 0, len(placement.Preferences)),
	}

	// Convert constraints
	equal, notEqual := " == ", " != "
	for _, j := range placement.Constraints {
		operator := equal
		if strings.Contains(j, notEqual) {
			operator = notEqual
		}
		p := strings.Split(j, operator)
		if len(p) < 2 {
			log.Warnf("Failed to parse placement constraints %s, the correct format is 'label == xxx'", j)
			continue
		}

		key, err := convertDockerLabel(p[0])
		if err != nil {
			log.Warn("Ignore placement constraints: ", err.Error())
			continue
		}

		if operator == equal {
			komposePlacement.PositiveConstraints[key] = p[1]
		} else if operator == notEqual {
			komposePlacement.NegativeConstraints[key] = p[1]
		}
	}

	// Convert preferences
	for _, p := range placement.Preferences {
		// Spread is the only supported strategy currently
		label, err := convertDockerLabel(p.Spread)
		if err != nil {
			log.Warn("Ignore placement preferences: ", err.Error())
			continue
		}
		komposePlacement.Preferences = append(komposePlacement.Preferences, label)
	}
	return komposePlacement
}

// Convert docker label to k8s label
func convertDockerLabel(dockerLabel string) (string, error) {
	switch dockerLabel {
	case "node.hostname":
		return "kubernetes.io/hostname", nil
	case "engine.labels.operatingsystem":
		return "kubernetes.io/os", nil
	default:
		if strings.HasPrefix(dockerLabel, "node.labels.") {
			return strings.TrimPrefix(dockerLabel, "node.labels."), nil
		}
	}
	errMsg := fmt.Sprint(dockerLabel, " is not supported, only 'node.hostname', 'engine.labels.operatingsystem' and 'node.labels.xxx' (ex: node.labels.something == anything) is supported")
	return "", errors.New(errMsg)
}

// Convert the Docker Compose volumes to []string (the old way)
// TODO: Check to see if it's a "bind" or "volume". Ignore for now.
// TODO: Refactor it similar to loadPorts
// See: https://docs.docker.com/compose/compose-file/#long-syntax-3
func loadVolumes(volumes []types.ServiceVolumeConfig) []string {
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

// Convert Docker Compose ports to kobject.Ports
// expose ports will be treated as TCP ports
func loadPorts(ports []types.ServicePortConfig, expose []string) []kobject.Ports {
	komposePorts := []kobject.Ports{}
	exist := map[string]bool{}

	for _, port := range ports {
		// Convert to a kobject struct with ports
		komposePorts = append(komposePorts, kobject.Ports{
			HostPort:      cast.ToInt32(port.Published),
			ContainerPort: int32(port.Target),
			HostIP:        port.HostIP,
			Protocol:      strings.ToUpper(port.Protocol),
		})
		exist[cast.ToString(port.Target)+port.Protocol] = true
	}

	for _, port := range expose {
		portValue := port
		protocol := string(api.ProtocolTCP)
		if strings.Contains(portValue, "/") {
			splits := strings.Split(port, "/")
			portValue = splits[0]
			protocol = splits[1]
		}

		if exist[portValue+protocol] {
			continue
		}
		komposePorts = append(komposePorts, kobject.Ports{
			HostPort:      cast.ToInt32(portValue),
			ContainerPort: cast.ToInt32(portValue),
			HostIP:        "",
			Protocol:      strings.ToUpper(protocol),
		})
	}

	return komposePorts
}

/*
	Convert the HealthCheckConfig as designed by Docker to

a Kubernetes-compatible format.
*/
func parseHealthCheckReadiness(labels types.Labels) (kobject.HealthCheck, error) {
	var test []string
	var httpPath string
	var httpPort, tcpPort, timeout, interval, retries, startPeriod int32
	var disable bool

	for key, value := range labels {
		switch key {
		case HealthCheckReadinessDisable:
			disable = cast.ToBool(value)
		case HealthCheckReadinessTest:
			if len(value) > 0 {
				test, _ = shlex.Split(value)
			}
		case HealthCheckReadinessHTTPGetPath:
			httpPath = value
		case HealthCheckReadinessHTTPGetPort:
			httpPort = cast.ToInt32(value)
		case HealthCheckReadinessTCPPort:
			tcpPort = cast.ToInt32(value)
		case HealthCheckReadinessInterval:
			parse, err := time.ParseDuration(value)
			if err != nil {
				return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check interval variable")
			}
			interval = int32(parse.Seconds())
		case HealthCheckReadinessTimeout:
			parse, err := time.ParseDuration(value)
			if err != nil {
				return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check timeout variable")
			}
			timeout = int32(parse.Seconds())
		case HealthCheckReadinessRetries:
			retries = cast.ToInt32(value)
		case HealthCheckReadinessStartPeriod:
			parse, err := time.ParseDuration(value)
			if err != nil {
				return kobject.HealthCheck{}, errors.Wrap(err, "unable to parse health check startPeriod variable")
			}
			startPeriod = int32(parse.Seconds())
		}
	}

	if len(test) > 0 {
		if test[0] == "NONE" {
			disable = true
			test = test[1:]
		}
		// Due to docker/cli adding "CMD-SHELL" to the struct, we remove the first element of composeHealthCheck.Test
		if test[0] == "CMD" || test[0] == "CMD-SHELL" {
			test = test[1:]
		}
	}

	return kobject.HealthCheck{
		Test:        test,
		HTTPPath:    httpPath,
		HTTPPort:    httpPort,
		TCPPort:     tcpPort,
		Timeout:     timeout,
		Interval:    interval,
		Retries:     retries,
		StartPeriod: startPeriod,
		Disable:     disable,
	}, nil
}

/*
	Convert the HealthCheckConfig as designed by Docker to

a Kubernetes-compatible format.
*/
func parseHealthCheck(composeHealthCheck types.HealthCheckConfig, labels types.Labels) (kobject.HealthCheck, error) {
	var httpPort, tcpPort, timeout, interval, retries, startPeriod int32
	var test []string
	var httpPath string

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

	if composeHealthCheck.Test != nil {
		test = composeHealthCheck.Test[1:]
	}

	for key, value := range labels {
		switch key {
		case HealthCheckLivenessHTTPGetPath:
			httpPath = value
		case HealthCheckLivenessHTTPGetPort:
			httpPort = cast.ToInt32(value)
		case HealthCheckLivenessTCPPort:
			tcpPort = cast.ToInt32(value)
		}
	}

	// Due to docker/cli adding "CMD-SHELL" to the struct, we remove the first element of composeHealthCheck.Test
	return kobject.HealthCheck{
		Test:        test,
		TCPPort:     tcpPort,
		HTTPPath:    httpPath,
		HTTPPort:    httpPort,
		Timeout:     timeout,
		Interval:    interval,
		Retries:     retries,
		StartPeriod: startPeriod,
	}, nil
}

func dockerComposeToKomposeMapping(composeObject *types.Project) (kobject.KomposeObject, error) {
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
		serviceConfig.Name = name
		serviceConfig.Image = composeServiceConfig.Image
		serviceConfig.WorkingDir = composeServiceConfig.WorkingDir
		serviceConfig.Annotations = composeServiceConfig.Labels
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

		if composeServiceConfig.StopGracePeriod != nil {
			serviceConfig.StopGracePeriod = composeServiceConfig.StopGracePeriod.String()
		}

		if err := parseNetwork(&composeServiceConfig, &serviceConfig, composeObject); err != nil {
			return kobject.KomposeObject{}, err
		}

		if err := parseResources(&composeServiceConfig, &serviceConfig); err != nil {
			return kobject.KomposeObject{}, err
		}

		serviceConfig.Restart = composeServiceConfig.Restart

		if composeServiceConfig.Deploy != nil {
			// Deploy keys
			// mode:
			serviceConfig.DeployMode = composeServiceConfig.Deploy.Mode
			// labels
			serviceConfig.DeployLabels = composeServiceConfig.Deploy.Labels

			// restart-policy: deploy.restart_policy.condition will rewrite restart option
			// see: https://docs.docker.com/compose/compose-file/#restart_policy
			if composeServiceConfig.Deploy.RestartPolicy != nil {
				serviceConfig.Restart = composeServiceConfig.Deploy.RestartPolicy.Condition
			}

			// replicas:
			if composeServiceConfig.Deploy.Replicas != nil {
				serviceConfig.Replicas = int(*composeServiceConfig.Deploy.Replicas)
			}

			// placement:
			serviceConfig.Placement = loadPlacement(composeServiceConfig.Deploy.Placement)

			if composeServiceConfig.Deploy.UpdateConfig != nil {
				serviceConfig.DeployUpdateConfig = *composeServiceConfig.Deploy.UpdateConfig
			}

			if composeServiceConfig.Deploy.EndpointMode == "vip" {
				serviceConfig.ServiceType = string(api.ServiceTypeNodePort)
			}
		}

		// HealthCheck Liveness
		if composeServiceConfig.HealthCheck != nil && !composeServiceConfig.HealthCheck.Disable {
			var err error
			serviceConfig.HealthChecks.Liveness, err = parseHealthCheck(*composeServiceConfig.HealthCheck, composeServiceConfig.Labels)
			if err != nil {
				return kobject.KomposeObject{}, errors.Wrap(err, "Unable to parse health check")
			}
		}

		// HealthCheck Readiness
		var readiness, errReadiness = parseHealthCheckReadiness(composeServiceConfig.Labels)
		if !readiness.Disable {
			serviceConfig.HealthChecks.Readiness = readiness
			if errReadiness != nil {
				return kobject.KomposeObject{}, errors.Wrap(errReadiness, "Unable to parse health check")
			}
		}

		if serviceConfig.Restart == "unless-stopped" {
			log.Warnf("Restart policy 'unless-stopped' in service %s is not supported, convert it to 'always'", name)
			serviceConfig.Restart = "always"
		}

		if composeServiceConfig.Build != nil {
			serviceConfig.Build = composeServiceConfig.Build.Context
			serviceConfig.Dockerfile = composeServiceConfig.Build.Dockerfile
			serviceConfig.BuildArgs = composeServiceConfig.Build.Args
			serviceConfig.BuildLabels = composeServiceConfig.Build.Labels
		}

		// env
		parseEnvironment(&composeServiceConfig, &serviceConfig)

		// Get env_file
		serviceConfig.EnvFile = composeServiceConfig.EnvFile

		// Parse the ports
		// v3 uses a new format called "long syntax" starting in 3.2
		// https://docs.docker.com/compose/compose-file/#ports

		// here we will translate `expose` too, they basically means the same thing in kubernetes
		serviceConfig.Port = loadPorts(composeServiceConfig.Ports, serviceConfig.Expose)

		// Parse the volumes
		// Again, in v3, we use the "long syntax" for volumes in terms of parsing
		// https://docs.docker.com/compose/compose-file/#long-syntax-3
		serviceConfig.VolList = loadVolumes(composeServiceConfig.Volumes)
		if err := parseKomposeLabels(composeServiceConfig.Labels, &serviceConfig); err != nil {
			return kobject.KomposeObject{}, err
		}

		// Log if the name will been changed
		if normalizeServiceNames(name) != name {
			log.Infof("Service name in docker-compose has been changed from %q to %q", name, normalizeServiceNames(name))
		}

		serviceConfig.Configs = composeServiceConfig.Configs
		serviceConfig.ConfigsMetaData = composeObject.Configs

		// Get GroupAdd, group should be mentioned in gid format but not the group name
		groupAdd, err := getGroupAdd(composeServiceConfig.GroupAdd)
		if err != nil {
			return kobject.KomposeObject{}, errors.Wrap(err, "GroupAdd should be mentioned in gid format, not a group name")
		}
		serviceConfig.GroupAdd = groupAdd

		// Final step, add to the array!
		komposeObject.ServiceConfigs[normalizeServiceNames(name)] = serviceConfig
	}

	handleVolume(&komposeObject, &composeObject.Volumes)
	return komposeObject, nil
}

func parseNetwork(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig, composeObject *types.Project) error {
	if len(composeServiceConfig.Networks) == 0 {
		if defaultNetwork, ok := composeObject.Networks["default"]; ok {
			normalizedNetworkName, err := normalizeNetworkNames(defaultNetwork.Name)
			if err != nil {
				return errors.Wrap(err, "Unable to normalize network name")
			}
			serviceConfig.Network = append(serviceConfig.Network, normalizedNetworkName)
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

			normalizedNetworkName, err := normalizeNetworkNames(netName)
			if err != nil {
				return errors.Wrap(err, "Unable to normalize network name")
			}

			serviceConfig.Network = append(serviceConfig.Network, normalizedNetworkName)
		}
	}

	return nil
}

func parseResources(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig) error {
	serviceConfig.MemLimit = composeServiceConfig.MemLimit

	if composeServiceConfig.Deploy != nil {
		// memory:
		// See: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
		// "The expression 0.1 is equivalent to the expression 100m, which can be read as “one hundred millicpu”."

		// Since Deploy.Resources.Limits does not initialize, we must check type Resources before continuing
		if composeServiceConfig.Deploy.Resources.Limits != nil {
			serviceConfig.MemLimit = composeServiceConfig.Deploy.Resources.Limits.MemoryBytes

			if composeServiceConfig.Deploy.Resources.Limits.NanoCPUs != "" {
				cpuLimit, err := strconv.ParseFloat(composeServiceConfig.Deploy.Resources.Limits.NanoCPUs, 64)
				if err != nil {
					return errors.Wrap(err, "Unable to convert cpu limits resources value")
				}
				serviceConfig.CPULimit = int64(cpuLimit * 1000)
			}
		}
		if composeServiceConfig.Deploy.Resources.Reservations != nil {
			serviceConfig.MemReservation = composeServiceConfig.Deploy.Resources.Reservations.MemoryBytes

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

func parseEnvironment(composeServiceConfig *types.ServiceConfig, serviceConfig *kobject.ServiceConfig) {
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
		case LabelServiceExternalTrafficPolicy:
			serviceExternalTypeTrafficPolicy, err := handleServiceExternalTrafficPolicy(value)
			if err != nil {
				return errors.Wrap(err, "handleServiceExternalTrafficPolicy failed")
			}

			serviceConfig.ServiceExternalTrafficPolicy = serviceExternalTypeTrafficPolicy
		case LabelSecurityContextFsGroup:
			serviceConfig.FsGroup = cast.ToInt64(value)
		case LabelServiceExpose:
			serviceConfig.ExposeService = strings.Trim(strings.ToLower(value), " ,")
		case LabelNodePortPort:
			serviceConfig.NodePortPort = cast.ToInt32(value)
		case LabelServiceExposeTLSSecret:
			serviceConfig.ExposeServiceTLS = value
		case LabelServiceExposeIngressClassName:
			serviceConfig.ExposeServiceIngressClassName = value
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

	if serviceConfig.ExposeService == "" && serviceConfig.ExposeServiceIngressClassName != "" {
		return errors.New("kompose.service.expose.ingress-class-name was specified without kompose.service.expose")
	}

	if serviceConfig.ServiceType != string(api.ServiceTypeNodePort) && serviceConfig.NodePortPort != 0 {
		return errors.New("kompose.service.type must be nodeport when assign node port value")
	}

	if len(serviceConfig.Port) > 1 && serviceConfig.NodePortPort != 0 {
		return errors.New("cannot set kompose.service.nodeport.port when service has multiple ports")
	}

	return nil
}

func handleVolume(komposeObject *kobject.KomposeObject, volumes *types.Volumes) {
	for name := range komposeObject.ServiceConfigs {
		// retrieve volumes of service
		vols, err := retrieveVolume(name, *komposeObject)
		if err != nil {
			errors.Wrap(err, "could not retrieve vvolume")
		}
		for volName, vol := range vols {
			size, selector := getVolumeLabels(vol.VolumeName, volumes)
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

func getVolumeLabels(name string, volumes *types.Volumes) (string, string) {
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
