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
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	dockerlib "github.com/fsouza/go-dockerclient"
	"github.com/kubernetes/kompose/pkg/utils/archive"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Build will provide methods for interaction with API regarding building images
type Build struct {
	Client dockerlib.Client
}

/*
BuildImage builds a Docker image via the Docker API or Docker CLI.
Takes the source directory and image name and then builds the appropriate image. Tarball is utilized
in order to make building easier.

if the DOCKER_BUILDKIT is '1', then we will use the docker CLI to build the image
*/
func (c *Build) BuildImage(source string, image string, dockerfile string, buildargs []dockerlib.BuildArg) error {
	log.Infof("Building image '%s' from directory '%s'", image, path.Base(source))

	outputBuffer := bytes.NewBuffer(nil)
	var err error

	if usecli, _ := strconv.ParseBool(os.Getenv("DOCKER_BUILDKIT")); usecli {
		err = buildDockerCli(source, image, dockerfile, buildargs, outputBuffer)
	} else {
		err = c.buildDockerClient(source, image, dockerfile, buildargs, outputBuffer)
	}

	log.Debugf("Image %s build output:\n%s", image, outputBuffer)

	if err != nil {
		return errors.Wrap(err, "Unable to build image. For more output, use -v or --verbose when converting.")
	}

	log.Infof("Image '%s' from directory '%s' built successfully", image, path.Base(source))

	return nil
}

func (c *Build) buildDockerClient(source string, image string, dockerfile string, buildargs []dockerlib.BuildArg, outputBuffer *bytes.Buffer) error {
	// Create a temporary file for tarball image packaging
	tmpFile, err := os.CreateTemp(os.TempDir(), "kompose-image-build-")
	if err != nil {
		return err
	}
	log.Debugf("Created temporary file %v for Docker image tarballing", tmpFile.Name())

	// Create a tarball of the source directory in order to build the resulting image
	err = archive.CreateTarball(strings.Join([]string{source, ""}, "/"), tmpFile.Name())
	if err != nil {
		return errors.Wrap(err, "Unable to create a tarball")
	}

	// Load the file into memory
	tarballSource, err := os.Open(tmpFile.Name())
	if err != nil {
		return errors.Wrap(err, "Unable to load tarball into memory")
	}

	// Let's create all the options for the image building.
	opts := dockerlib.BuildImageOptions{
		Name:         image,
		InputStream:  tarballSource,
		OutputStream: outputBuffer,
		Dockerfile:   dockerfile,
		BuildArgs:    buildargs,
	}

	// Build it!
	return c.Client.BuildImage(opts)
}

func buildDockerCli(source string, image string, dockerfile string, buildargs []dockerlib.BuildArg, outputBuffer *bytes.Buffer) error {
	args := []string{"build", "-t", image}

	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}

	for _, buildarg := range buildargs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", buildarg.Name, buildarg.Value))
	}

	args = append(args, source)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = outputBuffer
	cmd.Stderr = outputBuffer

	log.Debugf("Image %s build calling command %v", image, cmd)

	return cmd.Run()
}
