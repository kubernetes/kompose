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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
)

const (
	// LabelServiceType defines the type of service to be created
	LabelServiceType = "kompose.service.type"
	// LabelNodePortPort defines the port value for NodePort service
	LabelNodePortPort = "kompose.service.nodeport.port"
	// LabelServiceExpose defines if the service needs to be made accessible from outside the cluster or not
	LabelServiceExpose = "kompose.service.expose"
	// LabelServiceExposeTLSSecret  provides the name of the TLS secret to use with the Kubernetes ingress controller
	LabelServiceExposeTLSSecret = "kompose.service.expose.tls-secret"
	// LabelControllerType defines the type of controller to be created
	LabelControllerType = "kompose.controller.type"
	// LabelImagePullSecret defines a secret name for kubernetes ImagePullSecrets
	LabelImagePullSecret = "kompose.image-pull-secret"
	// LabelImagePullPolicy defines Kubernetes PodSpec imagePullPolicy.
	LabelImagePullPolicy = "kompose.image-pull-policy"

	// ServiceTypeHeadless ...
	ServiceTypeHeadless = "Headless"
)

// load environment variables from compose file
func loadEnvVars(envars []string) []kobject.EnvVar {
	envs := []kobject.EnvVar{}
	for _, e := range envars {
		character := ""
		equalPos := strings.Index(e, "=")
		colonPos := strings.Index(e, ":")
		switch {
		case equalPos == -1 && colonPos == -1:
			character = ""
		case equalPos == -1 && colonPos != -1:
			character = ":"
		case equalPos != -1 && colonPos == -1:
			character = "="
		case equalPos != -1 && colonPos != -1:
			if equalPos > colonPos {
				character = ":"
			} else {
				character = "="
			}
		}

		if character == "" {
			envs = append(envs, kobject.EnvVar{
				Name:  e,
				Value: os.Getenv(e),
			})
		} else {
			values := strings.SplitN(e, character, 2)
			// try to get value from os env
			if values[1] == "" {
				values[1] = os.Getenv(values[0])
			}
			envs = append(envs, kobject.EnvVar{
				Name:  values[0],
				Value: values[1],
			})
		}
	}

	return envs
}

// getComposeFileDir returns compose file directory
// Assume all the docker-compose files are in the same directory
func getComposeFileDir(inputFiles []string) (string, error) {
	inputFile := inputFiles[0]
	if strings.Index(inputFile, "/") != 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "Unable to retrieve compose file directory")
		}
		inputFile = filepath.Join(workDir, inputFile)
	}
	return filepath.Dir(inputFile), nil
}

func handleServiceType(ServiceType string) (string, error) {
	switch strings.ToLower(ServiceType) {
	case "", "clusterip":
		return string(api.ServiceTypeClusterIP), nil
	case "nodeport":
		return string(api.ServiceTypeNodePort), nil
	case "loadbalancer":
		return string(api.ServiceTypeLoadBalancer), nil
	case "headless":
		return ServiceTypeHeadless, nil
	default:
		return "", errors.New("Unknown value " + ServiceType + " , supported values are 'nodeport, clusterip, headless or loadbalancer'")
	}
}

func normalizeContainerNames(svcName string) string {
	return strings.ToLower(svcName)
}

func normalizeServiceNames(svcName string) string {
	re := regexp.MustCompile("[._]")
	return strings.ToLower(re.ReplaceAllString(svcName, "-"))
}

func normalizeVolumes(svcName string) string {
	return strings.Replace(svcName, "_", "-", -1)
}

func normalizeNetworkNames(netName string) (string, error) {
	netval := strings.ToLower(netName)
	regString := ("[^A-Za-z0-9.-]+")
	reg, err := regexp.Compile(regString)
	if err != nil {
		return "", err
	}
	netval = reg.ReplaceAllString(netval, "")
	return netval, nil
}

// ReadFile read data from file or stdin
func ReadFile(fileName string) ([]byte, error) {
	if fileName == "-" {
		if StdinData == nil {
			data, err := ioutil.ReadAll(os.Stdin)
			StdinData = data
			return data, err
		}
		return StdinData, nil
	}
	return ioutil.ReadFile(fileName)

}
