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
	libcomposeyaml "github.com/docker/libcompose/yaml"
	"io/ioutil"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"

	log "github.com/Sirupsen/logrus"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/pkg/errors"
	"os"
)

// converts os.Environ() ([]string) to map[string]string
// based on https://github.com/docker/cli/blob/5dd30732a23bbf14db1c64d084ae4a375f592cfa/cli/command/stack/deploy_composefile.go#L143
func buildEnvironment() (map[string]string, error) {
	// This is a test
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

	// Load and then parse the YAML first!
	loadedFile, err := ioutil.ReadFile(files[0])
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
		Filename: files[0],
		Config:   parsedComposeFile,
	}

	// get environment variables
	env, err := buildEnvironment()
	if err != nil {
		return kobject.KomposeObject{}, errors.Wrap(err, "cannot build environment variables")
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
	config, err := loader.Load(configDetails)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	// TODO: Check all "unsupported" keys and output details
	// Specifically, keys such as "volumes_from" are not supported in V3.

	// Finally, we convert the object from docker/cli's ServiceConfig to our appropriate one
	komposeObject, err := dockerComposeToKomposeMapping(config)
	if err != nil {
		return kobject.KomposeObject{}, err
	}

	return komposeObject, nil
}

// Convert the Docker Compose v3 volumes to []string (the old way)
// TODO: Check to see if it's a "bind" or "volume". Ignore for now.
// TODO: Refactor it similar to loadV3Ports
// See: https://docs.docker.com/compose/compose-file/#long-syntax-2
func loadV3Volumes(volumes []types.ServiceVolumeConfig) []string {
	var volArray []string
	for _, vol := range volumes {

		// There will *always* be Source when parsing
		v := normalizeServiceNames(vol.Source)

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
func loadV3Ports(ports []types.ServicePortConfig) []kobject.Ports {
	komposePorts := []kobject.Ports{}

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

	}

	return komposePorts
}

func dockerComposeToKomposeMapping(composeObject *types.Config) (kobject.KomposeObject, error) {

	// Step 1. Initialize what's going to be returned
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
		LoadedFrom:     "compose",
	}

	// Step 2. Parse through the object and conver it to kobject.KomposeObject!
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
		serviceConfig.ContainerName = composeServiceConfig.ContainerName
		serviceConfig.Command = composeServiceConfig.Entrypoint
		serviceConfig.Args = composeServiceConfig.Command

		//Handling replicas
		if composeServiceConfig.Deploy.Replicas != nil {
			serviceConfig.Replicas = int(*composeServiceConfig.Deploy.Replicas)
		}

		// This is a bit messy since we use yaml.MemStringorInt
		// TODO: Refactor yaml.MemStringorInt in kobject.go to int64
		// Since Deploy.Resources.Limits does not initialize, we must check type Resources before continuing
		if (composeServiceConfig.Deploy.Resources != types.Resources{}) {
			serviceConfig.MemLimit = libcomposeyaml.MemStringorInt(composeServiceConfig.Deploy.Resources.Limits.MemoryBytes)
		}
		//Here we handle all Docker Compose Deploy keys

		//Handling restart-policy
		if composeServiceConfig.Deploy.RestartPolicy != nil {
			serviceConfig.Restart = composeServiceConfig.Deploy.RestartPolicy.Condition
		}
		// POOF. volumes_From is gone in v3. docker/cli will error out of volumes_from is added in v3
		// serviceConfig.VolumesFrom = composeServiceConfig.VolumesFrom

		// TODO: Build is not yet supported, see:
		// https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L9
		// We will have to *manually* add this / parse.
		// serviceConfig.Build = composeServiceConfig.Build.Context
		// serviceConfig.Dockerfile = composeServiceConfig.Build.Dockerfile

		// Gather the environment values
		// DockerCompose uses map[string]*string while we use []string
		// So let's convert that using this hack
		for name, value := range composeServiceConfig.Environment {
			env := kobject.EnvVar{Name: name, Value: *value}
			serviceConfig.Environment = append(serviceConfig.Environment, env)
		}

		// Parse the ports
		// v3 uses a new format called "long syntax" starting in 3.2
		// https://docs.docker.com/compose/compose-file/#ports
		serviceConfig.Port = loadV3Ports(composeServiceConfig.Ports)

		// Parse the volumes
		// Again, in v3, we use the "long syntax" for volumes in terms of parsing
		// https://docs.docker.com/compose/compose-file/#long-syntax-2
		serviceConfig.Volumes = loadV3Volumes(composeServiceConfig.Volumes)

		// Label handler
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

		// Log if the name will been changed
		if normalizeServiceNames(name) != name {
			log.Infof("Service name in docker-compose has been changed from %q to %q", name, normalizeServiceNames(name))
		}

		// Final step, add to the array!
		komposeObject.ServiceConfigs[normalizeServiceNames(name)] = serviceConfig
	}

	return komposeObject, nil
}
