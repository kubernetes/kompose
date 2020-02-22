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
	dockerCliTypes "github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/yaml"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
	"path/filepath"
	"time"
)

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
	// LoadedFrom is name of the loader that created KomposeObject
	// Transformer need to know origin format in order to tell user what tag is not supported in origin format
	// as they can have different names. For example environment variables  are called environment in compose but Env in bundle.
	LoadedFrom string

	Secrets map[string]dockerCliTypes.SecretConfig
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
	PushImage                   bool
	CreateChart                 bool
	GenerateYaml                bool
	GenerateJSON                bool
	StoreManifest               bool
	EmptyVols                   bool
	Volumes                     string
	InsecureRepository          bool
	Replicas                    int
	InputFiles                  []string
	OutFile                     string
	Provider                    string
	Namespace                   string
	Controller                  string
	IsDeploymentFlag            bool
	IsDaemonSetFlag             bool
	IsReplicationControllerFlag bool
	IsReplicaSetFlag            bool
	IsDeploymentConfigFlag      bool
	IsNamespaceFlag             bool

	Server string

	YAMLIndent int
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
	ContainerName     string
	Image             string              `compose:"image"`
	Environment       []EnvVar            `compose:"environment"`
	EnvFile           []string            `compose:"env_file"`
	Port              []Ports             `compose:"ports"`
	Command           []string            `compose:"command"`
	WorkingDir        string              `compose:""`
	DomainName        string              `compose:"domainname"`
	HostName          string              `compose:"hostname"`
	Args              []string            `compose:"args"`
	VolList           []string            `compose:"volumes"`
	Network           []string            `compose:"network"`
	Labels            map[string]string   `compose:"labels"`
	Annotations       map[string]string   `compose:""`
	CPUSet            string              `compose:"cpuset"`
	CPUShares         int64               `compose:"cpu_shares"`
	CPUQuota          int64               `compose:"cpu_quota"`
	CPULimit          int64               `compose:""`
	CPUReservation    int64               `compose:""`
	CapAdd            []string            `compose:"cap_add"`
	CapDrop           []string            `compose:"cap_drop"`
	Expose            []string            `compose:"expose"`
	ImagePullPolicy   string              `compose:"kompose.image-pull-policy"`
	Pid               string              `compose:"pid"`
	Privileged        bool                `compose:"privileged"`
	Restart           string              `compose:"restart"`
	User              string              `compose:"user"`
	VolumesFrom       []string            `compose:"volumes_from"`
	ServiceType       string              `compose:"kompose.service.type"`
	NodePortPort      int32               `compose:"kompose.service.nodeport.port"`
	StopGracePeriod   string              `compose:"stop_grace_period"`
	Build             string              `compose:"build"`
	BuildArgs         map[string]*string  `compose:"build-args"`
	ExposeService     string              `compose:"kompose.service.expose"`
	ExposeServicePath string              `compose:"kompose.service.expose.path"`
	BuildLabels       map[string]string   `compose:"build-labels"`
	ExposeServiceTLS  string              `compose:"kompose.service.expose.tls-secret"`
	ImagePullSecret   string              `compose:"kompose.image-pull-secret"`
	Stdin             bool                `compose:"stdin_open"`
	Tty               bool                `compose:"tty"`
	MemLimit          yaml.MemStringorInt `compose:"mem_limit"`
	MemReservation    yaml.MemStringorInt `compose:""`
	DeployMode        string              `compose:""`
	// DeployLabels mapping to kubernetes labels
	DeployLabels       map[string]string           `compose:""`
	DeployUpdateConfig dockerCliTypes.UpdateConfig `compose:""`
	TmpFs              []string                    `compose:"tmpfs"`
	Dockerfile         string                      `compose:"dockerfile"`
	Replicas           int                         `compose:"replicas"`
	GroupAdd           []int64                     `compose:"group_add"`
	Volumes            []Volumes                   `compose:""`
	Secrets            []dockerCliTypes.ServiceSecretConfig
	HealthChecks       HealthCheck       `compose:""`
	Placement          map[string]string `compose:""`
	//This is for long LONG SYNTAX link(https://docs.docker.com/compose/compose-file/#long-syntax)
	Configs []dockerCliTypes.ServiceConfigObjConfig `compose:""`
	//This is for SHORT SYNTAX link(https://docs.docker.com/compose/compose-file/#configs)
	ConfigsMetaData map[string]dockerCliTypes.ConfigObjConfig `compose:""`
}

// HealthCheck the healthcheck configuration for a service
// "StartPeriod" is not yet added to compose, see:
// https://github.com/docker/cli/issues/116
type HealthCheck struct {
	Test        []string
	Timeout     int32
	Interval    int32
	Retries     int32
	StartPeriod int32
	Disable     bool
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
	SvcName       string // Service name to which volume is linked
	MountPath     string // Mountpath extracted from docker-compose file
	VFrom         string // denotes service name from which volume is coming
	VolumeName    string // name of volume if provided explicitly
	Host          string // host machine address
	Container     string // Mountpath
	Mode          string // access mode for volume
	PVCName       string // name of PVC
	PVCSize       string // PVC size
	SelectorValue string // Value of the label selector
}

// GetConfigMapKeyFromMeta ...
// given a source name ,find the file and extract the filename which will be act as ConfigMap key
// return "" if not found
func (s *ServiceConfig) GetConfigMapKeyFromMeta(name string) (string, error) {
	if s.ConfigsMetaData == nil {
		return "", errors.Errorf("config %s not found", name)
	}
	if _, ok := s.ConfigsMetaData[name]; !ok {
		return "", errors.Errorf("config %s not found", name)
	}

	config := s.ConfigsMetaData[name]
	if config.External.External {
		return "", errors.Errorf("config %s is external", name)
	}

	return filepath.Base(config.File), nil

}

// GetKubernetesUpdateStrategy from compose update_config
// 1. only apply to Deployment, but the check is not happened here
// 2. only support `parallelism` and `order`
// return nil if not support
func (s *ServiceConfig) GetKubernetesUpdateStrategy() *extensions.RollingUpdateDeployment {
	config := s.DeployUpdateConfig
	r := extensions.RollingUpdateDeployment{}
	if config.Order == "stop-first" {
		if config.Parallelism != nil {
			r.MaxUnavailable = intstr.FromInt(cast.ToInt(*config.Parallelism))

		}
		r.MaxSurge = intstr.FromInt(0)
		return &r
	}

	if config.Order == "start-first" {
		if config.Parallelism != nil {
			r.MaxSurge = intstr.FromInt(cast.ToInt(*config.Parallelism))
		}
		r.MaxUnavailable = intstr.FromInt(0)
		return &r
	}
	return nil

}

// GetOSUpdateStrategy ...
func (s *ServiceConfig) GetOSUpdateStrategy() *deployapi.RollingDeploymentStrategyParams {
	config := s.DeployUpdateConfig
	r := deployapi.RollingDeploymentStrategyParams{}

	delay := time.Second * 1
	if config.Delay != 0 {
		delay = config.Delay
	}

	interval := cast.ToInt64(delay.Seconds())

	if config.Order == "stop-first" {
		if config.Parallelism != nil {
			r.MaxUnavailable = intstr.FromInt(cast.ToInt(*config.Parallelism))
		}
		r.MaxSurge = intstr.FromInt(0)
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	if config.Order == "start-first" {
		if config.Parallelism != nil {
			r.MaxSurge = intstr.FromInt(cast.ToInt(*config.Parallelism))
		}
		r.MaxUnavailable = intstr.FromInt(0)
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	if cast.ToInt64(config.Delay) != 0 {
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	return nil
}
