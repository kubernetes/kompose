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
	log "github.com/Sirupsen/logrus"
	dockerlib "github.com/fsouza/go-dockerclient"
	"github.com/novln/docker-parser"
	"github.com/pkg/errors"
)

type Push struct {
	Client dockerlib.Client
}

/*
Push a Docker image via the Docker API. Takes the image name,
parses the URL details and then push based on environment authentication
credentials.
*/
func (c *Push) PushImage(fullImageName string) error {
	outputBuffer := bytes.NewBuffer(nil)

	// Using https://github.com/novln/docker-parser in order to parse the appropriate
	// name and registry.
	parsedImage, err := dockerparser.Parse(fullImageName)
	if err != nil {
		return err
	}
	image, registry := parsedImage.Name(), parsedImage.Registry()

	log.Infof("Pushing image '%s' to registry '%s'", image, registry)

	// Let's setup the push and authentication options
	options := dockerlib.PushImageOptions{
		Name:         parsedImage.Name(),
		Registry:     parsedImage.Registry(),
		OutputStream: outputBuffer,
	}

	// Retrieve the authentication configuration file
	// Files checked as per https://godoc.org/github.com/fsouza/go-dockerclient#NewAuthConfigurationsFromFile
	// $DOCKER_CONFIG/config.json, $HOME/.docker/config.json , $HOME/.dockercfg
	credentials, err := dockerlib.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve .docker/config.json authentication details. Check that 'docker login' works successfully on the command line.")
	}

	// Push the image to the repository (based on the URL)
	// We will iterate through all available authentication configurations until we find one that pushes successfully
	// and then return nil.
	if len(credentials.Configs) > 1 {
		log.Info("Multiple authentication credentials detected. Will try each configuration.")
	}

	for k, v := range credentials.Configs {
		log.Infof("Attempting authentication credentials '%s", k)
		err = c.Client.PushImage(options, v)
		if err != nil {
			log.Errorf("Unable to push image '%s' to registry '%s'. Error: %s", image, registry, err)
		} else {
			log.Debugf("Image '%s' push output:\n%s", image, outputBuffer)
			log.Infof("Successfully pushed image '%s' to registry '%s'", image, registry)
			return nil
		}
	}

	return errors.New("Unable to push docker image(s). Check that `docker login` works successfully on the command line.")
}
