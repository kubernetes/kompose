package project

import (
	"strings"

	"github.com/docker/docker/runconfig"
)

// DefaultDependentServices return the dependent services (as an array of ServiceRelationship)
// for the specified project and service. It looks for : links, volumesFrom, net and ipc configuration.
func DefaultDependentServices(p *Project, s Service) []ServiceRelationship {
	config := s.Config()
	if config == nil {
		return []ServiceRelationship{}
	}

	result := []ServiceRelationship{}
	for _, link := range config.Links.Slice() {
		result = append(result, NewServiceRelationship(link, RelTypeLink))
	}

	for _, volumesFrom := range config.VolumesFrom {
		result = append(result, NewServiceRelationship(volumesFrom, RelTypeVolumesFrom))
	}

	result = appendNs(p, result, s.Config().Net, RelTypeNetNamespace)
	result = appendNs(p, result, s.Config().Ipc, RelTypeIpcNamespace)

	return result
}

func appendNs(p *Project, rels []ServiceRelationship, conf string, relType ServiceRelationshipType) []ServiceRelationship {
	service := GetContainerFromIpcLikeConfig(p, conf)
	if service != "" {
		rels = append(rels, NewServiceRelationship(service, relType))
	}
	return rels
}

// NameAlias returns the name and alias based on the specified string.
// If the name contains a colon (like name:alias) it will split it, otherwise
// it will return the specified name as name and alias.
func NameAlias(name string) (string, string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], parts[0]
}

// GetContainerFromIpcLikeConfig returns name of the service that shares the IPC
// namespace with the specified service.
func GetContainerFromIpcLikeConfig(p *Project, conf string) string {
	ipc := runconfig.IpcMode(conf)
	if !ipc.IsContainer() {
		return ""
	}

	name := ipc.Container()
	if name == "" {
		return ""
	}

	if _, ok := p.Configs[name]; ok {
		return name
	}
	return ""
}
