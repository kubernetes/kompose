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

package transformer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/kubernetes/kompose/pkg/kobject"

	"path/filepath"

	"github.com/kubernetes/kompose/pkg/utils/docker"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
)

// Selector used as labels and selector
const Selector = "io.kompose.service"

// CreateOutFile creates the file to write to if --out is specified
func CreateOutFile(out string) (*os.File, error) {
	var f *os.File
	var err error
	if len(out) != 0 {
		f, err = os.Create(out)
		if err != nil {
			return nil, errors.Wrap(err, "error creating file, os.Create failed")
		}
	}
	return f, nil
}

// ParseVolume parses a given volume, which might be [name:][host:]container[:access_mode]
func ParseVolume(volume string) (name, host, container, mode string, err error) {
	separator := ":"

	// Parse based on ":"
	volumeStrings := strings.Split(volume, separator)
	if len(volumeStrings) == 0 {
		return
	}

	// Set name if existed
	if !isPath(volumeStrings[0]) {
		name = volumeStrings[0]
		volumeStrings = volumeStrings[1:]
	}

	// Check if *anything* has been passed
	if len(volumeStrings) == 0 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}

	// Get the last ":" passed which is presumingly the "access mode"
	possibleAccessMode := volumeStrings[len(volumeStrings)-1]

	// Check to see if :Z or :z exists. We do not support SELinux relabeling at the moment.
	// See https://github.com/kubernetes/kompose/issues/176
	// Otherwise, check to see if "rw" or "ro" has been passed
	if possibleAccessMode == "z" || possibleAccessMode == "Z" {
		log.Warnf("Volume mount \"%s\" will be mounted without labeling support. :z or :Z not supported", volume)
		mode = ""
		volumeStrings = volumeStrings[:len(volumeStrings)-1]
	} else if possibleAccessMode == "rw" || possibleAccessMode == "ro" {
		mode = possibleAccessMode
		volumeStrings = volumeStrings[:len(volumeStrings)-1]
	}

	// Check the volume format as well as host
	container = volumeStrings[len(volumeStrings)-1]
	volumeStrings = volumeStrings[:len(volumeStrings)-1]
	if len(volumeStrings) == 1 {
		host = volumeStrings[0]
	}
	if !isPath(container) || (len(host) > 0 && !isPath(host)) || len(volumeStrings) > 1 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}
	return
}

func isPath(substring string) bool {
	return strings.Contains(substring, "/") || substring == "."
}

// ConfigLabels configures label
func ConfigLabels(name string) map[string]string {
	return map[string]string{Selector: name}
}

// ConfigAnnotations configures annotations
func ConfigAnnotations(service kobject.ServiceConfig) map[string]string {
	annotations := map[string]string{}
	for key, value := range service.Annotations {
		annotations[key] = value
	}
	annotations["kompose.cmd"] = strings.Join(os.Args, " ")
	version := exec.Command("kompose", "version")
	out, err := version.Output()
	if err != nil {
		errors.Wrap(err, "Failed to get kompose version")

	}
	annotations["kompose.version"] = strings.Trim(string(out), " \n")
	return annotations
}

// Print either prints to stdout or to file/s
func Print(name, path string, trailing string, data []byte, toStdout, generateJSON bool, f *os.File, provider string) (string, error) {
	file := ""
	if generateJSON {
		file = fmt.Sprintf("%s-%s.json", name, trailing)
	} else {
		file = fmt.Sprintf("%s-%s.yaml", name, trailing)
	}
	if toStdout {
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
		return "", nil
	} else if f != nil {
		// Write all content to a single file f
		if _, err := f.WriteString(fmt.Sprintf("%s\n", string(data))); err != nil {
			return "", errors.Wrap(err, "f.WriteString failed, Failed to write %s to file: "+trailing)
		}
		f.Sync()
	} else {
		// Write content separately to each file
		file = filepath.Join(path, file)
		if err := ioutil.WriteFile(file, []byte(data), 0644); err != nil {
			return "", errors.Wrap(err, "Failed to write %s: "+trailing)
		}
		log.Printf("%s file %q created", formatProviderName(provider), file)
	}
	return file, nil
}

// If Openshift, change to OpenShift!
func formatProviderName(provider string) string {
	if strings.EqualFold(provider, "openshift") {
		return "OpenShift"
	} else if strings.EqualFold(provider, "kubernetes") {
		return "Kubernetes"
	}
	return provider
}

// EnvSort struct
type EnvSort []api.EnvVar

// returns the number of elements in the collection.
func (env EnvSort) Len() int {
	return len(env)
}

// returns whether the element with index i should sort before
// the element with index j.
func (env EnvSort) Less(i, j int) bool {
	return env[i].Name < env[j].Name
}

// swaps the elements with indexes i and j.
func (env EnvSort) Swap(i, j int) {
	env[i], env[j] = env[j], env[i]
}

// GetComposeFileDir returns compose file directory
func GetComposeFileDir(inputFiles []string) (string, error) {
	// Lets assume all the docker-compose files are in the same directory
	inputFile := inputFiles[0]
	if strings.Index(inputFile, "/") != 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		inputFile = filepath.Join(workDir, inputFile)
	}
	log.Debugf("Compose file dir: %s", filepath.Dir(inputFile))
	return filepath.Dir(inputFile), nil
}

//BuildDockerImage builds docker image
func BuildDockerImage(service kobject.ServiceConfig, name string, relativePath string) error {

	// First, let's figure out the relative path of the Dockerfile!
	// else, we error out.
	if _, err := os.Stat(service.Build); err != nil {
		return errors.Wrapf(err, "%s is not a valid path for building image %s. Check if this dir exists.", service.Build, name)
	}

	// Get the appropriate image source and name
	// use path.Base to get the last element of the relative build path
	imagePath := path.Join(relativePath, path.Base(service.Build))
	imageName := name
	if service.Image != "" {
		imageName = service.Image
	}

	// Connect to the Docker client
	client, err := docker.DockerClient()
	if err != nil {
		return err
	}

	// Use the build struct function to build the image
	// Build the image!
	build := docker.Build{Client: *client}
	err = build.BuildImage(imagePath, imageName)

	if err != nil {
		return err
	}

	return nil
}

// PushDockerImage pushes docker image
func PushDockerImage(service kobject.ServiceConfig, serviceName string) error {

	log.Debugf("Pushing Docker image '%s'", service.Image)

	// Don't do anything if service.Image is blank, but at least WARN about it
	// lse, let's push the image
	if service.Image == "" {
		log.Warnf("No image name has been passed for service %s, skipping pushing to repository", serviceName)
		return nil
	} else {

		// Connect to the Docker client
		client, err := docker.DockerClient()
		if err != nil {
			return err
		}

		push := docker.Push{Client: *client}
		err = push.PushImage(service.Image)

		if err != nil {
			return err
		}
	}

	return nil
}
