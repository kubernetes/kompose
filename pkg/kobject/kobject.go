/*
Copyright 2016 Skippbox, Ltd All rights reserved.

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
	"github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"k8s.io/kubernetes/pkg/api"
)

var unsupportedKey = map[string]int{
	"Build":         0,
	"CapAdd":        0,
	"CapDrop":       0,
	"CPUSet":        0,
	"CPUShares":     0,
	"CPUQuota":      0,
	"CgroupParent":  0,
	"Devices":       0,
	"DependsOn":     0,
	"DNS":           0,
	"DNSSearch":     0,
	"DomainName":    0,
	"EnvFile":       0,
	"Expose":        0,
	"Extends":       0,
	"ExternalLinks": 0,
	"ExtraHosts":    0,
	"Hostname":      0,
	"Ipc":           0,
	"Logging":       0,
	"MacAddress":    0,
	"MemLimit":      0,
	"MemSwapLimit":  0,
	"NetworkMode":   0,
	"Networks":      0,
	"Pid":           0,
	"SecurityOpt":   0,
	"ShmSize":       0,
	"StopSignal":    0,
	"VolumeDriver":  0,
	"VolumesFrom":   0,
	"Uts":           0,
	"ReadOnly":      0,
	"StdinOpen":     0,
	"Tty":           0,
	"User":          0,
	"Ulimits":       0,
	"Dockerfile":    0,
	"Net":           0,
	"Args":          0,
}

var composeOptions = map[string]string{
	"Build":         "build",
	"CapAdd":        "cap_add",
	"CapDrop":       "cap_drop",
	"CPUSet":        "cpuset",
	"CPUShares":     "cpu_shares",
	"CPUQuota":      "cpu_quota",
	"CgroupParent":  "cgroup_parent",
	"Devices":       "devices",
	"DependsOn":     "depends_on",
	"DNS":           "dns",
	"DNSSearch":     "dns_search",
	"DomainName":    "domainname",
	"Entrypoint":    "entrypoint",
	"EnvFile":       "env_file",
	"Expose":        "expose",
	"Extends":       "extends",
	"ExternalLinks": "external_links",
	"ExtraHosts":    "extra_hosts",
	"Hostname":      "hostname",
	"Ipc":           "ipc",
	"Logging":       "logging",
	"MacAddress":    "mac_address",
	"MemLimit":      "mem_limit",
	"MemSwapLimit":  "memswap_limit",
	"NetworkMode":   "network_mode",
	"Networks":      "networks",
	"Pid":           "pid",
	"SecurityOpt":   "security_opt",
	"ShmSize":       "shm_size",
	"StopSignal":    "stop_signal",
	"VolumeDriver":  "volume_driver",
	"VolumesFrom":   "volumes_from",
	"Uts":           "uts",
	"ReadOnly":      "read_only",
	"StdinOpen":     "stdin_open",
	"Tty":           "tty",
	"User":          "user",
	"Ulimits":       "ulimits",
	"Dockerfile":    "dockerfile",
	"Net":           "net",
	"Args":          "args",
}

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
	Replicas               int
	InputFile              string
	OutFile                string
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

func CheckUnsupportedKey(service interface{}) {
	s := structs.New(service)
	for _, f := range s.Fields() {
		if f.IsExported() && !f.IsZero() && f.Name() != "Networks" {
			if count, ok := unsupportedKey[f.Name()]; ok && count == 0 {
				logrus.Warningf("Unsupported key %s - ignoring", composeOptions[f.Name()])
				unsupportedKey[f.Name()]++
			}
		}
	}
}
