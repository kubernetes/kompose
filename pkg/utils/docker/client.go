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
)

func DockerClient() (*docker.Client, error) {

	// Default end-point, HTTP + TLS support to be added in the future
	// Eventually functionality to specify end-point added to command-line
	endpoint := "unix:///var/run/docker.sock"

	// Use the unix socker end-point. No support for TLS (yet)
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return client, err
	}

	return client, nil
}
