package api

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// DockerImage is the type representing a docker image and its various properties when
// retrieved from the Docker client API.
type DockerImage struct {
	unversioned.TypeMeta `json:",inline"`

	ID              string           `json:"Id"`
	Parent          string           `json:"Parent,omitempty"`
	Comment         string           `json:"Comment,omitempty"`
	Created         unversioned.Time `json:"Created,omitempty"`
	Container       string           `json:"Container,omitempty"`
	ContainerConfig DockerConfig     `json:"ContainerConfig,omitempty"`
	DockerVersion   string           `json:"DockerVersion,omitempty"`
	Author          string           `json:"Author,omitempty"`
	Config          *DockerConfig    `json:"Config,omitempty"`
	Architecture    string           `json:"Architecture,omitempty"`
	Size            int64            `json:"Size,omitempty"`
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
	Labels          map[string]string   `json:"Labels,omitempty"`
}

// Descriptor describes targeted content. Used in conjunction with a blob
// store, a descriptor can be used to fetch, store and target any kind of
// blob. The struct also describes the wire protocol format. Fields should
// only be added but never changed.
type Descriptor struct {
	// MediaType describe the type of the content. All text based formats are
	// encoded as utf-8.
	MediaType string `json:"mediaType,omitempty"`

	// Size in bytes of content.
	Size int64 `json:"size,omitempty"`

	// Digest uniquely identifies the content. A byte stream can be verified
	// against against this digest.
	Digest string `json:"digest,omitempty"`
}

// DockerImageManifest represents the Docker v2 image format.
type DockerImageManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType,omitempty"`

	// schema1
	Name         string          `json:"name"`
	Tag          string          `json:"tag"`
	Architecture string          `json:"architecture"`
	FSLayers     []DockerFSLayer `json:"fsLayers"`
	History      []DockerHistory `json:"history"`

	// schema2
	Layers []Descriptor `json:"layers"`
	Config Descriptor   `json:"config"`
}

// DockerFSLayer is a container struct for BlobSums defined in an image manifest
type DockerFSLayer struct {
	// DockerBlobSum is the tarsum of the referenced filesystem image layer
	// TODO make this digest.Digest once docker/distribution is in Godeps
	DockerBlobSum string `json:"blobSum"`
}

// DockerHistory stores unstructured v1 compatibility information
type DockerHistory struct {
	// DockerV1Compatibility is the raw v1 compatibility information
	DockerV1Compatibility string `json:"v1Compatibility"`
}

// DockerV1CompatibilityImage represents the structured v1
// compatibility information.
type DockerV1CompatibilityImage struct {
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

// DockerV1CompatibilityImageSize represents the structured v1
// compatibility information for size
type DockerV1CompatibilityImageSize struct {
	Size int64 `json:"size,omitempty"`
}

// DockerImageConfig stores the image configuration
type DockerImageConfig struct {
	ID              string                `json:"id"`
	Parent          string                `json:"parent,omitempty"`
	Comment         string                `json:"comment,omitempty"`
	Created         unversioned.Time      `json:"created"`
	Container       string                `json:"container,omitempty"`
	ContainerConfig DockerConfig          `json:"container_config,omitempty"`
	DockerVersion   string                `json:"docker_version,omitempty"`
	Author          string                `json:"author,omitempty"`
	Config          *DockerConfig         `json:"config,omitempty"`
	Architecture    string                `json:"architecture,omitempty"`
	Size            int64                 `json:"size,omitempty"`
	RootFS          *DockerConfigRootFS   `json:"rootfs,omitempty"`
	History         []DockerConfigHistory `json:"history,omitempty"`
	OSVersion       string                `json:"os.version,omitempty"`
	OSFeatures      []string              `json:"os.features,omitempty"`
}

// DockerConfigHistory stores build commands that were used to create an image
type DockerConfigHistory struct {
	Created    unversioned.Time `json:"created"`
	Author     string           `json:"author,omitempty"`
	CreatedBy  string           `json:"created_by,omitempty"`
	Comment    string           `json:"comment,omitempty"`
	EmptyLayer bool             `json:"empty_layer,omitempty"`
}

// DockerConfigRootFS describes images root filesystem
type DockerConfigRootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids,omitempty"`
}
