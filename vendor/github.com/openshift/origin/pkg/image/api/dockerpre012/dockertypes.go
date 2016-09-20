package dockerpre012

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// DockerImage is for earlier versions of the Docker API (pre-012 to be specific). It is also the
// version of metadata that the Docker registry uses to persist metadata.
type DockerImage struct {
	unversioned.TypeMeta `json:",inline"`

	ID              string           `json:"id"`
	Parent          string           `json:"parent,omitempty"`
	Comment         string           `json:"comment,omitempty"`
	Created         unversioned.Time `json:"created"`
	Container       string           `json:"container,omitempty"`
	ContainerConfig DockerConfig     `json:"container_config,omitempty"`
	DockerVersion   string           `json:"docker_version,omitempty"`
	Author          string           `json:"author,omitempty"`
	Config          *DockerConfig    `json:"config,omitempty"`
	Architecture    string           `json:"architecture,omitempty"`
	Size            int64            `json:"size,omitempty"`
}

// DockerConfig is the list of configuration options used when creating a container.
type DockerConfig struct {
	Hostname        string              `json:"Hostname,omitempty"`
	Domainname      string              `json:"Domainname,omitempty"`
	User            string              `json:"User,omitempty"`
	Memory          int64               `json:"Memory,omitempty"`
	MemorySwap      int64               `json:"MemorySwap,omitempty"`
	CPUShares       int64               `json:"CpuShares,omitempty"`
	CPUSet          string              `json:"Cpuset,omitempty"`
	AttachStdin     bool                `json:"AttachStdin,omitempty"`
	AttachStdout    bool                `json:"AttachStdout,omitempty"`
	AttachStderr    bool                `json:"AttachStderr,omitempty"`
	PortSpecs       []string            `json:"PortSpecs,omitempty"`
	ExposedPorts    map[string]struct{} `json:"ExposedPorts,omitempty"`
	Tty             bool                `json:"Tty,omitempty"`
	OpenStdin       bool                `json:"OpenStdin,omitempty"`
	StdinOnce       bool                `json:"StdinOnce,omitempty"`
	Env             []string            `json:"Env,omitempty"`
	Cmd             []string            `json:"Cmd,omitempty"`
	DNS             []string            `json:"Dns,omitempty"` // For Docker API v1.9 and below only
	Image           string              `json:"Image,omitempty"`
	Volumes         map[string]struct{} `json:"Volumes,omitempty"`
	VolumesFrom     string              `json:"VolumesFrom,omitempty"`
	WorkingDir      string              `json:"WorkingDir,omitempty"`
	Entrypoint      []string            `json:"Entrypoint,omitempty"`
	NetworkDisabled bool                `json:"NetworkDisabled,omitempty"`
	SecurityOpts    []string            `json:"SecurityOpts,omitempty"`
	OnBuild         []string            `json:"OnBuild,omitempty"`
	// This field is not supported in pre012 and will always be empty.
	Labels map[string]string `json:"Labels,omitempty"`
}
