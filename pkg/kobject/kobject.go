/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package kobject

import "k8s.io/kubernetes/pkg/api"

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
}

type ConvertOptions struct {
	ToStdout               bool
	CreateD                bool
	CreateRC               bool
	CreateDS               bool
	CreateDeploymentConfig bool
	CreateChart            bool
	GenerateYaml           bool
	EmptyVols              bool
	Replicas               int
	InputFile              string
	OutFile                string
	Provider               string
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
	ContainerName string
	Image         string
	Environment   []EnvVar
	Port          []Ports
	Command       []string
	WorkingDir    string
	Args          []string
	Volumes       []string
	Network       []string
	Labels        map[string]string
	Annotations   map[string]string
	CPUSet        string
	CPUShares     int64
	CPUQuota      int64
	CapAdd        []string
	CapDrop       []string
	Expose        []string
	Privileged    bool
	Restart       string
	User          string
	VolumesFrom   []string
	ServiceType   string
	Build         string
}

// EnvVar holds the environment variable struct of a container
type EnvVar struct {
	Name  string
	Value string
}

// Ports holds the ports struct of a container
type Ports struct {
	HostPort      int32
	ContainerPort int32
	Protocol      api.Protocol
}
