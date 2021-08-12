/*
Copyright 2016 The Kubernetes Authors All rights reserved

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

package docker

import (
	"bytes"

	dockerlib "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Push will provide methods for interaction with API regarding pushing images
type Push struct {
	Client dockerlib.Client
}

/*
PushImage pushes a Docker image via the Docker API. Takes the image name,
parses the URL details and then push based on environment authentication
credentials.
*/
func (c *Push) PushImage(image Image) error {
	log.Infof("Pushing image '%s' to registry '%s'", image.Name, image.Registry)

	// Let's setup the push and authentication options
	outputBuffer := bytes.NewBuffer(nil)
	options := dockerlib.PushImageOptions{
		Tag:          image.Tag,
		Name:         image.Repository,
		Registry:     image.Registry,
		OutputStream: outputBuffer,
	}

	// Retrieve the authentication configuration file
	// Files checked as per https://godoc.org/github.com/fsouza/go-dockerclient#NewAuthConfigurationsFromFile
	// $DOCKER_CONFIG/config.json, $HOME/.docker/config.json , $HOME/.dockercfg
	credentials, err := dockerlib.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		log.Warn(errors.Wrap(err, "Unable to retrieve .docker/config.json authentication details. Check that 'docker login' works successfully on the command line."))
	} else {
		handleDockerRegistry(credentials)
	}

	// Find the authentication matched to registry
	auth, ok := credentials.Configs[image.Registry]
	if !ok {
		// Fallback to unauthenticated access in case if no auth credentials are retrieved
		log.Infof("Authentication credential of registry '%s' is not found. Will try push without authentication.", image.Registry)
		// Header X-Registry-Auth is required
		// Or API error (400): Bad parameters and missing X-Registry-Auth: EOF will throw
		// Just to make not empty struct
		auth = dockerlib.AuthConfiguration{Username: "docker"}
	}

	log.Debugf("Pushing image with options %+v", options)
	err = c.Client.PushImage(options, auth)
	if err != nil {
		log.Errorf("Unable to push image '%s' to registry '%s'. Error: %s", image.Name, image.Registry, err)
		return errors.New("unable to push docker image(s). Check that `docker login` works successfully on the command line")
	}

	log.Debugf("Image '%+v' push output:\n%s", image, outputBuffer)
	log.Infof("Successfully pushed image '%s' to registry '%s'", image.Name, image.Registry)
	return nil
}

// handleDockerRegistry adapt legacy docker registry address
// After docker login to docker.io, there must be https://index.docker.io/v1/ in config.json of authentication
// Reference: https://docs.docker.com/engine/api/v1.23/
//     > However (for legacy reasons) the “official” Docker, Inc. hosted registry
//     > must be specified with both a “https://” prefix and a “/v1/” suffix
//     > even though Docker will prefer to use the v2 registry API.
func handleDockerRegistry(auth *dockerlib.AuthConfigurations) {
	const address = "docker.io"
	const legacyAddress = "https://index.docker.io/v1/"

	if legacyAddressConfig, ok := auth.Configs[legacyAddress]; ok {
		auth.Configs[address] = legacyAddressConfig
	}
}
