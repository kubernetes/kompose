package project

import "fmt"

// EventType defines a type of libcompose event.
type EventType int

// Definitions of libcompose events
const (
	NoEvent = EventType(iota)

	EventContainerCreated = EventType(iota)
	EventContainerStarted = EventType(iota)

	EventServiceAdd          = EventType(iota)
	EventServiceUpStart      = EventType(iota)
	EventServiceUpIgnored    = EventType(iota)
	EventServiceUp           = EventType(iota)
	EventServiceCreateStart  = EventType(iota)
	EventServiceCreate       = EventType(iota)
	EventServiceDeleteStart  = EventType(iota)
	EventServiceDelete       = EventType(iota)
	EventServiceDownStart    = EventType(iota)
	EventServiceDown         = EventType(iota)
	EventServiceRestartStart = EventType(iota)
	EventServiceRestart      = EventType(iota)
	EventServicePullStart    = EventType(iota)
	EventServicePull         = EventType(iota)
	EventServiceKillStart    = EventType(iota)
	EventServiceKill         = EventType(iota)
	EventServiceStartStart   = EventType(iota)
	EventServiceStart        = EventType(iota)
	EventServiceBuildStart   = EventType(iota)
	EventServiceBuild        = EventType(iota)
	EventServicePauseStart   = EventType(iota)
	EventServicePause        = EventType(iota)
	EventServiceUnpauseStart = EventType(iota)
	EventServiceUnpause      = EventType(iota)

	EventProjectDownStart     = EventType(iota)
	EventProjectDownDone      = EventType(iota)
	EventProjectCreateStart   = EventType(iota)
	EventProjectCreateDone    = EventType(iota)
	EventProjectUpStart       = EventType(iota)
	EventProjectUpDone        = EventType(iota)
	EventProjectDeleteStart   = EventType(iota)
	EventProjectDeleteDone    = EventType(iota)
	EventProjectRestartStart  = EventType(iota)
	EventProjectRestartDone   = EventType(iota)
	EventProjectReload        = EventType(iota)
	EventProjectReloadTrigger = EventType(iota)
	EventProjectKillStart     = EventType(iota)
	EventProjectKillDone      = EventType(iota)
	EventProjectStartStart    = EventType(iota)
	EventProjectStartDone     = EventType(iota)
	EventProjectBuildStart    = EventType(iota)
	EventProjectBuildDone     = EventType(iota)
	EventProjectPauseStart    = EventType(iota)
	EventProjectPauseDone     = EventType(iota)
	EventProjectUnpauseStart  = EventType(iota)
	EventProjectUnpauseDone   = EventType(iota)
)

func (e EventType) String() string {
	var m string
	switch e {
	case EventContainerCreated:
		m = "Created container"
	case EventContainerStarted:
		m = "Started container"

	case EventServiceAdd:
		m = "Adding"
	case EventServiceUpStart:
		m = "Starting"
	case EventServiceUpIgnored:
		m = "Ignoring"
	case EventServiceUp:
		m = "Started"
	case EventServiceCreateStart:
		m = "Creating"
	case EventServiceCreate:
		m = "Created"
	case EventServiceDeleteStart:
		m = "Deleting"
	case EventServiceDelete:
		m = "Deleted"
	case EventServiceDownStart:
		m = "Stopping"
	case EventServiceDown:
		m = "Stopped"
	case EventServiceRestartStart:
		m = "Restarting"
	case EventServiceRestart:
		m = "Restarted"
	case EventServicePullStart:
		m = "Pulling"
	case EventServicePull:
		m = "Pulled"
	case EventServiceKillStart:
		m = "Killing"
	case EventServiceKill:
		m = "Killed"
	case EventServiceStartStart:
		m = "Starting"
	case EventServiceStart:
		m = "Started"
	case EventServiceBuildStart:
		m = "Building"
	case EventServiceBuild:
		m = "Built"

	case EventProjectDownStart:
		m = "Stopping project"
	case EventProjectDownDone:
		m = "Project stopped"
	case EventProjectCreateStart:
		m = "Creating project"
	case EventProjectCreateDone:
		m = "Project created"
	case EventProjectUpStart:
		m = "Starting project"
	case EventProjectUpDone:
		m = "Project started"
	case EventProjectDeleteStart:
		m = "Deleting project"
	case EventProjectDeleteDone:
		m = "Project deleted"
	case EventProjectRestartStart:
		m = "Restarting project"
	case EventProjectRestartDone:
		m = "Project restarted"
	case EventProjectReload:
		m = "Reloading project"
	case EventProjectReloadTrigger:
		m = "Triggering project reload"
	case EventProjectKillStart:
		m = "Killing project"
	case EventProjectKillDone:
		m = "Project killed"
	case EventProjectStartStart:
		m = "Starting project"
	case EventProjectStartDone:
		m = "Project started"
	case EventProjectBuildStart:
		m = "Building project"
	case EventProjectBuildDone:
		m = "Project built"
	}

	if m == "" {
		m = fmt.Sprintf("EventType: %d", int(e))
	}

	return m
}

// InfoPart holds key/value strings.
type InfoPart struct {
	Key, Value string
}

// InfoSet holds a list of Info.
type InfoSet []Info

// Info holds a list of InfoPart.
type Info []InfoPart

// ServiceConfig holds libcompose service configuration
type ServiceConfig struct {
	Build         string            `yaml:"build,omitempty"`
	CapAdd        []string          `yaml:"cap_add,omitempty"`
	CapDrop       []string          `yaml:"cap_drop,omitempty"`
	CPUSet        string            `yaml:"cpuset,omitempty"`
	CPUShares     int64             `yaml:"cpu_shares,omitempty"`
	Command       Command           `yaml:"command,flow,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Devices       []string          `yaml:"devices,omitempty"`
	DNS           Stringorslice     `yaml:"dns,omitempty"`
	DNSSearch     Stringorslice     `yaml:"dns_search,omitempty"`
	Dockerfile    string            `yaml:"dockerfile,omitempty"`
	DomainName    string            `yaml:"domainname,omitempty"`
	Entrypoint    Command           `yaml:"entrypoint,flow,omitempty"`
	EnvFile       Stringorslice     `yaml:"env_file,omitempty"`
	Environment   MaporEqualSlice   `yaml:"environment,omitempty"`
	Hostname      string            `yaml:"hostname,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Labels        SliceorMap        `yaml:"labels,omitempty"`
	Links         MaporColonSlice   `yaml:"links,omitempty"`
	LogDriver     string            `yaml:"log_driver,omitempty"`
	MemLimit      int64             `yaml:"mem_limit,omitempty"`
	MemSwapLimit  int64             `yaml:"memswap_limit,omitempty"`
	Name          string            `yaml:"name,omitempty"`
	Net           string            `yaml:"net,omitempty"`
	Pid           string            `yaml:"pid,omitempty"`
	Uts           string            `yaml:"uts,omitempty"`
	Ipc           string            `yaml:"ipc,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Privileged    bool              `yaml:"privileged,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	ReadOnly      bool              `yaml:"read_only,omitempty"`
	StdinOpen     bool              `yaml:"stdin_open,omitempty"`
	SecurityOpt   []string          `yaml:"security_opt,omitempty"`
	Tty           bool              `yaml:"tty,omitempty"`
	User          string            `yaml:"user,omitempty"`
	VolumeDriver  string            `yaml:"volume_driver,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	VolumesFrom   []string          `yaml:"volumes_from,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	Expose        []string          `yaml:"expose,omitempty"`
	ExternalLinks []string          `yaml:"external_links,omitempty"`
	LogOpt        map[string]string `yaml:"log_opt,omitempty"`
	ExtraHosts    []string          `yaml:"extra_hosts,omitempty"`
}

// EnvironmentLookup defines methods to provides environment variable loading.
type EnvironmentLookup interface {
	Lookup(key, serviceName string, config *ServiceConfig) []string
}

// ConfigLookup defines methods to provides file loading.
type ConfigLookup interface {
	Lookup(file, relativeTo string) ([]byte, string, error)
}

// Project holds libcompose project information.
type Project struct {
	Name           string
	Configs        map[string]*ServiceConfig
	File           string
	ReloadCallback func() error
	context        *Context
	reload         []string
	upCount        int
	listeners      []chan<- Event
	hasListeners   bool
}

// Service defines what a libcompose service provides.
type Service interface {
	Info(qFlag bool) (InfoSet, error)
	Name() string
	Build() error
	Create() error
	Up() error
	Start() error
	Down() error
	Delete() error
	Restart() error
	Log() error
	Pull() error
	Kill() error
	Config() *ServiceConfig
	DependentServices() []ServiceRelationship
	Containers() ([]Container, error)
	Scale(count int) error
	Pause() error
	Unpause() error
}

// Container defines what a libcompose container provides.
type Container interface {
	ID() (string, error)
	Name() string
	Port(port string) (string, error)
}

// ServiceFactory is an interface factory to create Service object for the specified
// project, with the specified name and service configuration.
type ServiceFactory interface {
	Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error)
}

// ServiceRelationshipType defines the type of service relationship.
type ServiceRelationshipType string

// RelTypeLink means the services are linked (docker links).
const RelTypeLink = ServiceRelationshipType("")

// RelTypeNetNamespace means the services share the same network namespace.
const RelTypeNetNamespace = ServiceRelationshipType("netns")

// RelTypeIpcNamespace means the service share the same ipc namespace.
const RelTypeIpcNamespace = ServiceRelationshipType("ipc")

// RelTypeVolumesFrom means the services share some volumes.
const RelTypeVolumesFrom = ServiceRelationshipType("volumesFrom")

// ServiceRelationship holds the relationship information between two services.
type ServiceRelationship struct {
	Target, Alias string
	Type          ServiceRelationshipType
	Optional      bool
}

// NewServiceRelationship creates a new Relationship based on the specified alias
// and relationship type.
func NewServiceRelationship(nameAlias string, relType ServiceRelationshipType) ServiceRelationship {
	name, alias := NameAlias(nameAlias)
	return ServiceRelationship{
		Target: name,
		Alias:  alias,
		Type:   relType,
	}
}
