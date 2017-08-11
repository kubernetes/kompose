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
	Build                       string
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
	IsReplicaSetFlag            bool
	IsDeploymentConfigFlag      bool
	IsNamespaceFlag             bool
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
	// use tags to mark from what element this value comes
	ContainerName string
	Image         string   `compose:"image"`
	Environment   []EnvVar `compose:"environment"`
	Port          []Ports  `compose:"ports"`
	Command       []string `compose:"command"`
	WorkingDir    string   `compose:""`
	Args          []string `compose:"args"`
	// VolList is list of volumes extracted from docker-compose file
	VolList         []string            `compose:"volumes"`
	Network         []string            `compose:"network"`
	Labels          map[string]string   `compose:"labels"`
	Annotations     map[string]string   `compose:""`
	CPUSet          string              `compose:"cpuset"`
	CPUShares       int64               `compose:"cpu_shares"`
	CPUQuota        int64               `compose:"cpu_quota"`
	CPULimit        int64               `compose:""`
	CPUReservation  int64               `compose:""`
	CapAdd          []string            `compose:"cap_add"`
	CapDrop         []string            `compose:"cap_drop"`
	Expose          []string            `compose:"expose"`
	Pid             string              `compose:"pid"`
	Privileged      bool                `compose:"privileged"`
	Restart         string              `compose:"restart"`
	User            string              `compose:"user"`
	VolumesFrom     []string            `compose:"volumes_from"`
	ServiceType     string              `compose:"kompose.service.type"`
	StopGracePeriod string              `compose:"stop_grace_period"`
	Build           string              `compose:"build"`
	BuildArgs       map[string]*string  `compose:"build-args"`
	ExposeService   string              `compose:"kompose.service.expose"`
	Stdin           bool                `compose:"stdin_open"`
	Tty             bool                `compose:"tty"`
	MemLimit        yaml.MemStringorInt `compose:"mem_limit"`
	MemReservation  yaml.MemStringorInt `compose:""`
	TmpFs           []string            `compose:"tmpfs"`
	Dockerfile      string              `compose:"dockerfile"`
	Replicas        int                 `compose:"replicas"`
	GroupAdd        []int64             `compose:"group_add"`
	// Volumes is a struct which contains all information about each volume
	Volumes []Volumes `compose:""`
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

// Volumes holds the volume struct of container
type Volumes struct {
	SvcName    string // Service name to which volume is linked
	MountPath  string // Mountpath extracted from docker-compose file
	VFrom      string // denotes service name from which volume is coming
	VolumeName string // name of volume if provided explicitly
	Host       string // host machine address
	Container  string // Mountpath
	Mode       string // access mode for volume
	PVCName    string // name of PVC
}
