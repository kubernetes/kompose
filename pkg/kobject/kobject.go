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
	"path/filepath"
	"strconv"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	deployapi "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
	// LoadedFrom is name of the loader that created KomposeObject
	// Transformer need to know origin format in order to tell user what tag is not supported in origin format
	// as they can have different names. For example environment variables  are called environment in compose but Env in bundle.
	LoadedFrom string

	Secrets types.Secrets

	// Namespace is the namespace where all the generated objects would be assigned to
	Namespace string
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
	Profiles                    []string
	PushImage                   bool
	PushImageRegistry           string
	CreateChart                 bool
	GenerateYaml                bool
	GenerateJSON                bool
	StoreManifest               bool
	EmptyVols                   bool
	Volumes                     string
	PVCRequestSize              string
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

	BuildCommand string
	PushCommand  string

	Server string

	YAMLIndent int

	WithKomposeAnnotation bool

	MultipleContainerMode   bool
	ServiceGroupMode        string
	ServiceGroupName        string
	SecretsAsFiles          bool
	GenerateNetworkPolicies bool
}

// IsPodController indicate if the user want to use a controller
func (opt *ConvertOptions) IsPodController() bool {
	return opt.IsDeploymentFlag || opt.IsDaemonSetFlag || opt.IsReplicationControllerFlag || opt.Controller != ""
}

// ServiceConfigGroup holds an array of a ServiceConfig objects.
type ServiceConfigGroup []ServiceConfig

// ServiceConfig holds the basic struct of a container
// which should not introduce any kubernetes specific struct
type ServiceConfig struct {
	Name                          string
	ContainerName                 string
	Image                         string             `compose:"image"`
	Environment                   []EnvVar           `compose:"environment"`
	EnvFile                       []string           `compose:"env_file"`
	Port                          []Ports            `compose:"ports"`
	Command                       []string           `compose:"command"`
	WorkingDir                    string             `compose:""`
	DomainName                    string             `compose:"domainname"`
	HostName                      string             `compose:"hostname"`
	ReadOnly                      bool               `compose:"read_only"`
	Args                          []string           `compose:"args"`
	VolList                       []string           `compose:"volumes"`
	NetworkMode                   string             `compose:"network_mode"`
	Network                       []string           `compose:"network"`
	Labels                        map[string]string  `compose:"labels"`
	Annotations                   map[string]string  `compose:""`
	CPUSet                        string             `compose:"cpuset"`
	CPUShares                     int64              `compose:"cpu_shares"`
	CPUQuota                      int64              `compose:"cpu_quota"`
	CPULimit                      int64              `compose:""`
	CPUReservation                int64              `compose:""`
	CapAdd                        []string           `compose:"cap_add"`
	CapDrop                       []string           `compose:"cap_drop"`
	Expose                        []string           `compose:"expose"`
	ImagePullPolicy               string             `compose:"kompose.image-pull-policy"`
	Pid                           string             `compose:"pid"`
	Privileged                    bool               `compose:"privileged"`
	Restart                       string             `compose:"restart"`
	User                          string             `compose:"user"`
	VolumesFrom                   []string           `compose:"volumes_from"`
	ServiceType                   string             `compose:"kompose.service.type"`
	ServiceExternalTrafficPolicy  string             `compose:"kompose.service.external-traffic-policy"`
	NodePortPort                  int32              `compose:"kompose.service.nodeport.port"`
	StopGracePeriod               string             `compose:"stop_grace_period"`
	Build                         string             `compose:"build"`
	BuildArgs                     map[string]*string `compose:"build-args"`
	ExposeContainerToHost         bool               `compose:"kompose.controller.port.expose"`
	ExposeService                 string             `compose:"kompose.service.expose"`
	ExposeServicePath             string             `compose:"kompose.service.expose.path"`
	BuildLabels                   map[string]string  `compose:"build-labels"`
	BuildTarget                   string             `compose:""`
	ExposeServiceTLS              string             `compose:"kompose.service.expose.tls-secret"`
	ExposeServiceIngressClassName string             `compose:"kompose.service.expose.ingress-class-name"`
	ImagePullSecret               string             `compose:"kompose.image-pull-secret"`
	Stdin                         bool               `compose:"stdin_open"`
	Tty                           bool               `compose:"tty"`
	MemLimit                      types.UnitBytes    `compose:"mem_limit"`
	MemReservation                types.UnitBytes    `compose:""`
	DeployMode                    string             `compose:""`
	VolumeMountSubPath            string             `compose:"kompose.volume.subpath"`
	// DeployLabels mapping to kubernetes labels
	DeployLabels             map[string]string         `compose:""`
	DeployUpdateConfig       types.UpdateConfig        `compose:""`
	TmpFs                    []string                  `compose:"tmpfs"`
	Dockerfile               string                    `compose:"dockerfile"`
	Replicas                 int                       `compose:"replicas"`
	GroupAdd                 []int64                   `compose:"group_add"`
	FsGroup                  int64                     `compose:"kompose.security-context.fsgroup"`
	CronJobSchedule          string                    `compose:"kompose.cronjob.schedule"`
	CronJobConcurrencyPolicy batchv1.ConcurrencyPolicy `compose:"kompose.cronjob.concurrency_policy"`
	CronJobBackoffLimit      *int32                    `compose:"kompose.cronjob.backoff_limit"`
	Volumes                  []Volumes                 `compose:""`
	Secrets                  []types.ServiceSecretConfig
	HealthChecks             HealthChecks `compose:""`
	Placement                Placement    `compose:""`
	//This is for long LONG SYNTAX link(https://docs.docker.com/compose/compose-file/#long-syntax)
	Configs []types.ServiceConfigObjConfig `compose:""`
	//This is for SHORT SYNTAX link(https://docs.docker.com/compose/compose-file/#configs)
	ConfigsMetaData types.Configs `compose:""`

	WithKomposeAnnotation bool `compose:""`
	InGroup               bool
}

// HealthChecks used to distinguish between liveness and readiness
type HealthChecks struct {
	Liveness  HealthCheck
	Readiness HealthCheck
}

// HealthCheck the healthcheck configuration for a service
// "StartPeriod" was added to v3.4 of the compose, see:
// https://github.com/docker/cli/issues/116
type HealthCheck struct {
	Test        []string
	Timeout     int32
	Interval    int32
	Retries     int32
	StartPeriod int32
	Disable     bool
	HTTPPath    string
	HTTPPort    int32
	TCPPort     int32
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
	Protocol      string // Upper string
}

// ID returns an unique id for this port settings, to avoid conflict
func (port *Ports) ID() string {
	return strconv.Itoa(int(port.ContainerPort)) + port.Protocol
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

// Placement holds the placement struct of container
type Placement struct {
	PositiveConstraints map[string]string
	NegativeConstraints map[string]string
	Preferences         []string
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
	if config.External {
		return "", errors.Errorf("config %s is external", name)
	}

	if config.File != "" {
		return filepath.Base(config.File), nil
	} else if config.Content != "" {
		// loop through s.Configs to find the config with the same name
		for _, cfg := range s.Configs {
			if cfg.Source == name {
				if cfg.Target == "" {
					return filepath.Base(cfg.Source), nil
				} else {
					return filepath.Base(cfg.Target), nil
				}
			}
		}
	} else {
		return "", errors.Errorf("config %s is empty", name)
	}

	return "", errors.Errorf("config %s not found", name)
}

// GetKubernetesUpdateStrategy from compose update_config
// 1. only apply to Deployment, but the check is not happened here
// 2. only support `parallelism` and `order`
// return nil if not support
func (s *ServiceConfig) GetKubernetesUpdateStrategy() *v1.RollingUpdateDeployment {
	config := s.DeployUpdateConfig
	r := v1.RollingUpdateDeployment{}
	if config.Order == "stop-first" {
		if config.Parallelism != nil {
			v := intstr.FromInt(cast.ToInt(*config.Parallelism))
			r.MaxUnavailable = &v
		}

		v := intstr.FromInt(0)
		r.MaxSurge = &v
		return &r
	}

	if config.Order == "start-first" {
		if config.Parallelism != nil {
			v := intstr.FromInt(cast.ToInt(*config.Parallelism))
			r.MaxSurge = &v
		}
		v := intstr.FromInt(0)
		r.MaxUnavailable = &v
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
		delay = time.Duration(config.Delay)
	}

	interval := cast.ToInt64(delay.Seconds())

	if config.Order == "stop-first" {
		if config.Parallelism != nil {
			v := intstr.FromInt(cast.ToInt(*config.Parallelism))
			r.MaxUnavailable = &v
		}
		*r.MaxSurge = intstr.FromInt(0)
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	if config.Order == "start-first" {
		if config.Parallelism != nil {
			v := intstr.FromInt(cast.ToInt(*config.Parallelism))
			r.MaxSurge = &v
		}

		v := intstr.FromInt(0)
		r.MaxUnavailable = &v
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	if cast.ToInt64(config.Delay) != 0 {
		r.UpdatePeriodSeconds = &interval
		return &r
	}

	return nil
}
