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
	"github.com/fsouza/go-dockerclient"
	"os"
)

// Client connects to Docker client on host
func Client() (*docker.Client, error) {

	var (
		err    error
		client *docker.Client
	)

	dockerHost := os.Getenv("DOCKER_HOST")

	if len(dockerHost) > 0 {
		// Create client instance from Docker's environment variables:
		// DOCKER_HOST, DOCKER_TLS_VERIFY, DOCKER_CERT_PATH
		client, err = docker.NewClientFromEnv()
	} else {
		// Default unix socket end-point
		endpoint := "unix:///var/run/docker.sock"
		client, err = docker.NewClient(endpoint)
	}
	if err != nil {
		return client, err
	}

	return client, nil
}
