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

import (
	"github.com/docker/libcompose/yaml"
	"k8s.io/kubernetes/pkg/api"
)

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
	// LoadedFrom is name of the loader that created KomposeObject
	// Transformer need to know origin format in order to tell user what tag is not supported in origin format
	// as they can have different names. For example environment variables  are called environment in compose but Env in bundle.
	LoadedFrom string
}

// ConvertOptions holds all options that controls transformation process
type ConvertOptions struct {
	ToStdout                    bool
	CreateD                     bool
	CreateRC                    bool
	CreateDS                    bool
	CreateDeploymentConfig      bool
	BuildRepo                   string
	BuildBranch                 string
	CreateChart                 bool
	GenerateYaml                bool
	GenerateJSON                bool
	EmptyVols                   bool
	InsecureRepository          bool
	Replicas                    int
	InputFiles                  []string
	OutFile                     string
	Provider                    string
	Namespace                   string
	IsDeploymentFlag            bool
	IsDaemonSetFlag             bool
	IsReplicationControllerFlag bool
	IsDeploymentConfigFlag      bool
	IsNamespaceFlag             bool
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
	// use tags to mark from what element this value comes
	ContainerName   string
	Image           string              `compose:"image" bundle:"Image"`
	Environment     []EnvVar            `compose:"environment" bundle:"Env"`
	Port            []Ports             `compose:"ports" bundle:"Ports"`
	Command         []string            `compose:"command" bundle:"Command"`
	WorkingDir      string              `compose:"" bundle:"WorkingDir"`
	Args            []string            `compose:"args" bundle:"Args"`
	Volumes         []string            `compose:"volumes" bundle:"Volumes"`
	Network         []string            `compose:"network" bundle:"Networks"`
	Labels          map[string]string   `compose:"labels" bundle:"Labels"`
	Annotations     map[string]string   `compose:"" bundle:""`
	CPUSet          string              `compose:"cpuset" bundle:""`
	CPUShares       int64               `compose:"cpu_shares" bundle:""`
	CPUQuota        int64               `compose:"cpu_quota" bundle:""`
	CapAdd          []string            `compose:"cap_add" bundle:""`
	CapDrop         []string            `compose:"cap_drop" bundle:""`
	Expose          []string            `compose:"expose" bundle:""`
	Pid             string              `compose:"pid" bundle:""`
	Privileged      bool                `compose:"privileged" bundle:""`
	Restart         string              `compose:"restart" bundle:""`
	User            string              `compose:"user" bundle:"User"`
	VolumesFrom     []string            `compose:"volumes_from" bundle:""`
	ServiceType     string              `compose:"kompose.service.type" bundle:""`
	StopGracePeriod string              `compose:"stop_grace_period" bundle:""`
	Build           string              `compose:"build" bundle:""`
	BuildArgs       map[string]*string  `compose:"build-args" bundle:""`
	ExposeService   string              `compose:"kompose.service.expose" bundle:""`
	Stdin           bool                `compose:"stdin_open" bundle:""`
	Tty             bool                `compose:"tty" bundle:""`
	MemLimit        yaml.MemStringorInt `compose:"mem_limit" bundle:""`
	TmpFs           []string            `compose:"tmpfs" bundle:""`
	Dockerfile      string              `compose:"dockerfile" bundle:""`
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
	HostIP        string
	Protocol      api.Protocol
}
