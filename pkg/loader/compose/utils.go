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
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
)

const (
	// LabelServiceType defines the type of service to be created
	LabelServiceType = "kompose.service.type"
	// LabelServiceGroup defines the group of services in a single pod
	LabelServiceGroup = "kompose.service.group"
	// LabelNodePortPort defines the port value for NodePort service
	LabelNodePortPort = "kompose.service.nodeport.port"
	// LabelServiceExpose defines if the service needs to be made accessible from outside the cluster or not
	LabelServiceExpose = "kompose.service.expose"
	// LabelServiceExposeTLSSecret provides the name of the TLS secret to use with the Kubernetes ingress controller
	LabelServiceExposeTLSSecret = "kompose.service.expose.tls-secret"
	// LabelServiceExposeIngressClassName provides the name of ingress class to use with the Kubernetes ingress controller
	LabelServiceExposeIngressClassName = "kompose.service.expose.ingress-class-name"
	// LabelServiceAccountName defines the service account name to provide the credential info of the pod.
	LabelServiceAccountName = "kompose.serviceaccount-name"
	// LabelControllerType defines the type of controller to be created
	LabelControllerType = "kompose.controller.type"
	// LabelImagePullSecret defines a secret name for kubernetes ImagePullSecrets
	LabelImagePullSecret = "kompose.image-pull-secret"
	// LabelImagePullPolicy defines Kubernetes PodSpec imagePullPolicy.
	LabelImagePullPolicy = "kompose.image-pull-policy"
	// HealthCheckReadinessDisable defines readiness health check disable
	HealthCheckReadinessDisable = "kompose.service.healthcheck.readiness.disable"
	// HealthCheckReadinessTest defines readiness health check test
	HealthCheckReadinessTest = "kompose.service.healthcheck.readiness.test"
	// HealthCheckReadinessInterval defines readiness health check interval
	HealthCheckReadinessInterval = "kompose.service.healthcheck.readiness.interval"
	// HealthCheckReadinessTimeout defines readiness health check timeout
	HealthCheckReadinessTimeout = "kompose.service.healthcheck.readiness.timeout"
	// HealthCheckReadinessRetries defines readiness health check retries
	HealthCheckReadinessRetries = "kompose.service.healthcheck.readiness.retries"
	// HealthCheckReadinessStartPeriod defines readiness health check start period
	HealthCheckReadinessStartPeriod = "kompose.service.healthcheck.readiness.start_period"
	// HealthCheckReadinessHTTPGetPath defines readiness health check HttpGet path
	HealthCheckReadinessHTTPGetPath = "kompose.service.healthcheck.readiness.http_get_path"
	// HealthCheckReadinessHTTPGetPort defines readiness health check HttpGet port
	HealthCheckReadinessHTTPGetPort = "kompose.service.healthcheck.readiness.http_get_port"
	// HealthCheckReadinessTCPPort defines readiness health check tcp port
	HealthCheckReadinessTCPPort = "kompose.service.healthcheck.readiness.tcp_port"
	// HealthCheckLivenessHTTPGetPath defines liveness health check HttpGet path
	HealthCheckLivenessHTTPGetPath = "kompose.service.healthcheck.liveness.http_get_path"
	// HealthCheckLivenessHTTPGetPort defines liveness health check HttpGet port
	HealthCheckLivenessHTTPGetPort = "kompose.service.healthcheck.liveness.http_get_port"
	// HealthCheckLivenessTCPPort defines liveness health check tcp port
	HealthCheckLivenessTCPPort = "kompose.service.healthcheck.liveness.tcp_port"

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
	netval := strings.ToLower(strings.Replace(netName, "_", "-", -1))
	regString := "[^A-Za-z0-9.-]+"
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
			data, err := io.ReadAll(os.Stdin)
			StdinData = data
			return data, err
		}
		return StdinData, nil
	}
	return os.ReadFile(fileName)
}
