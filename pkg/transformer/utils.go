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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	dockerlib "github.com/fsouza/go-dockerclient"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/utils/docker"
	"github.com/kubernetes/kompose/pkg/version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
)

// Selector used as labels and selector
const Selector = "io.kompose.service"

// Exists returns true if a file path exists.
// Otherwise, returns false.
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// CreateOutFile creates the file to write to if --out is specified
func CreateOutFile(out string) (*os.File, error) {
	if len(out) == 0 {
		return nil, nil
	}
	// Creates directories if "out" contains nonexistent directories.
	if dir := filepath.Dir(out); !Exists(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, errors.Wrap(err, "failed to create directories")
		}
	}
	f, err := os.Create(out)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file, os.Create failed")
	}
	return f, nil
}

// ParseVolume parses a given volume, which might be [name:][host:]container[:access_mode]
func ParseVolume(volume string) (name, host, container, mode string, err error) {
	if containWindowsPath(volume) {
		return parseWindowsVolume(volume)
	}
	return parseVolume(volume)
}

func parseVolume(volume string) (name, host, container, mode string, err error) {
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

// parseVolume parses window volume.
// example: windows host mount to windows container
// volume = dataVolumeName:C:\Users\Data:D:\config:rw
// it can be parsed:
// name=dataVolumeName, host=C:\Users\Data, container=D:\config, mode=rw
// example: windows host mount to linux container
// volume = dataVolumeName:C:\Users\Data:/etc/config:rw
// it can be parsed:
// name=dataVolumeName, host=C:\Users\Data, container=/etc/config, mode=rw
func parseWindowsVolume(volume string) (name, host, container, mode string, err error) {
	var (
		buffer, volumePaths []string
		volumeStrings       = strings.Split(volume, ":")
	)

	// extract path and leave order
	for _, fragment := range volumeStrings {
		switch {
		case containWindowsPath(fragment):
			if len(buffer) == 0 {
				err = fmt.Errorf("invalid windows volume %s", volume)
				return
			}

			driveLetter := buffer[len(buffer)-1]
			if len(driveLetter) != 1 {
				err = fmt.Errorf("invalid windows volume %s", volume)
				return
			}
			volumePaths = append(volumePaths, driveLetter+":"+fragment)
			buffer = buffer[:len(buffer)-1]

		case isPath(fragment):
			volumePaths = append(volumePaths, fragment)
		default:
			buffer = append(buffer, fragment)
		}
	}

	// set name and mode if exist
	if len(buffer) == 1 {
		if volumeStrings[0] == buffer[0] {
			name = buffer[0]
		} else if volumeStrings[len(volumeStrings)-1] == buffer[0] {
			mode = buffer[0]
		}
	} else if len(buffer) == 2 {
		name = buffer[0]
		mode = buffer[1]
	} else if len(buffer) > 2 {
		err = fmt.Errorf("invalid windows volume %s", volume)
		return
	}

	// Support in pass time
	// Check to see if :Z or :z exists. We do not support SELinux relabeling at the moment.
	// See https://github.com/kubernetes/kompose/issues/176
	// Otherwise, check to see if "rw" or "ro" has been passed
	if mode == "z" || mode == "Z" {
		log.Warnf("Volume mount \"%s\" will be mounted without labeling support. :z or :Z not supported", volume)
		mode = ""
	}

	// Set host and container if exist
	if len(volumePaths) == 1 {
		container = volumePaths[0]
	} else if len(volumePaths) == 2 {
		host = volumePaths[0]
		container = volumePaths[1]
	} else {
		err = fmt.Errorf("invalid windows volume %s", volume)
		return
	}
	return
}

// containWindowsPath check whether it contains windows path.
// windows path's separator is "\"
func containWindowsPath(substring string) bool {
	return strings.Contains(substring, "\\")
}

// ParseIngressPath parse path for ingress.
// eg. example.com/org -> example.com org
func ParseIngressPath(url string) (string, string) {
	if strings.Contains(url, "/") {
		splits := strings.Split(url, "/")
		return splits[0], "/" + strings.Join(splits[1:], "/")
	}
	return url, ""
}

func isPath(substring string) bool {
	return strings.Contains(substring, "/") || substring == "."
}

// ConfigLabels configures label name alone
func ConfigLabels(name string) map[string]string {
	return map[string]string{Selector: name}
}

// ConfigLabelsWithNetwork configures label and add Network Information in labels
func ConfigLabelsWithNetwork(name string, net []string) map[string]string {
	labels := map[string]string{}
	labels[Selector] = name

	for _, n := range net {
		labels["io.kompose.network/"+n] = "true"
	}
	return labels
	//return map[string]string{Selector: name, "Network": net}
}

// ConfigAllLabels creates labels with service nam and deploy labels
func ConfigAllLabels(name string, service *kobject.ServiceConfig) map[string]string {
	base := ConfigLabels(name)
	if service.DeployLabels != nil {
		for k, v := range service.DeployLabels {
			base[k] = v
		}
	}
	return base
}

// ConfigAnnotations configures annotations
func ConfigAnnotations(service kobject.ServiceConfig) map[string]string {
	annotations := map[string]string{}
	for key, value := range service.Annotations {
		annotations[key] = value
	}

	if !service.WithKomposeAnnotation {
		return annotations
	}

	annotations["kompose.cmd"] = strings.Join(os.Args, " ")
	versionCmd := exec.Command("kompose", "version")
	out, err := versionCmd.Output()
	if err != nil {
		errors.Wrap(err, "Failed to get kompose version")
	}
	annotations["kompose.version"] = strings.Trim(string(out), " \n")

	// If the version is blank (couldn't retrieve the kompose version for whatever reason)
	if annotations["kompose.version"] == "" {
		annotations["kompose.version"] = version.VERSION + " (" + version.GITCOMMIT + ")"
	}

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
		if err := os.WriteFile(file, data, 0644); err != nil {
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

// Len returns the number of elements in the collection.
func (env EnvSort) Len() int {
	return len(env)
}

// Less returns whether the element with index i should sort before
// the element with index j.
func (env EnvSort) Less(i, j int) bool {
	return env[i].Name < env[j].Name
}

// Swap swaps the elements with indexes i and j.
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

// BuildDockerImage builds docker image
func BuildDockerImage(service kobject.ServiceConfig, name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Debug("Build image working dir is: ", wd)

	log.Debug("Build image service build is: ", service.Build)
	// Get the appropriate image source and name
	imagePath := service.Build
	if !path.IsAbs(service.Build) {
		imagePath = path.Join(wd, service.Build)
	}
	log.Debugf("Build image context is: %s", imagePath)

	if _, err := os.Stat(imagePath); err != nil {
		return errors.Wrapf(err, "%s is not a valid path for building image %s. Check if this dir exists.", service.Build, name)
	}

	imageName := name
	if service.Image != "" {
		imageName = service.Image
	}

	buildargs := []dockerlib.BuildArg{}
	for envName, envValue := range service.BuildArgs {
		var value string
		if envValue == nil {
			value = os.Getenv(envName)
		} else {
			value = *envValue
		}
		buildargs = append(buildargs, dockerlib.BuildArg{Name: envName, Value: value})
	}

	// Connect to the Docker client
	client, err := docker.Client()
	if err != nil {
		return err
	}

	// Use the build struct function to build the image
	// Build the image!
	build := docker.Build{Client: *client}
	err = build.BuildImage(imagePath, imageName, service.Dockerfile, buildargs)

	if err != nil {
		return err
	}

	return nil
}

// PushDockerImageWithOpt pushes docker image
func PushDockerImageWithOpt(service kobject.ServiceConfig, serviceName string, opt kobject.ConvertOptions) error {
	if !opt.PushImage {
		// Don't do anything if registry is specified but push is disabled, just WARN about it
		if opt.PushImageRegistry != "" {
			log.Warnf("Push image registry '%s' is specified but push image is disabled, skipping pushing to repository", opt.PushImageRegistry)
		}
		return nil
	}

	log.Infof("Push image is enabled. Attempting to push image '%s'", service.Image)

	// Don't do anything if service.Image is blank, but at least WARN about it
	// else, let's push the image
	if service.Image == "" {
		log.Warnf("No image name has been passed for service %s, skipping pushing to repository", serviceName)
		return nil
	}

	image, err := docker.ParseImage(service.Image, opt.PushImageRegistry)
	if err != nil {
		return err
	}

	client, err := docker.Client()
	if err != nil {
		return err
	}

	if opt.PushImageRegistry != "" {
		log.Info("Push image registry is specified. Tag the image into registry firstly.")
		tag := docker.Tag{Client: *client}
		err = tag.TagImage(image)

		if err != nil {
			return err
		}
	}

	push := docker.Push{Client: *client}
	err = push.PushImage(image)
	if err != nil {
		return err
	}

	return nil
}
