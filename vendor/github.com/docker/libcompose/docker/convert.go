package docker

import (
	"strings"

	"github.com/docker/docker/pkg/nat"
	"github.com/docker/docker/runconfig"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"

	dockerclient "github.com/fsouza/go-dockerclient"
)

// Filter filters the specified string slice with the specified function.
func Filter(vs []string, f func(string) bool) []string {
	r := make([]string, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

// ConvertToAPI converts a service configuration to a docker API container configuration.
func ConvertToAPI(c *project.ServiceConfig, name string) (*dockerclient.CreateContainerOptions, error) {
	config, hostConfig, err := Convert(c)
	if err != nil {
		return nil, err
	}

	result := dockerclient.CreateContainerOptions{
		Name:       name,
		Config:     config,
		HostConfig: hostConfig,
	}
	return &result, nil
}

func volumes(c *project.ServiceConfig) map[string]struct{} {
	vs := Filter(c.Volumes, isVolume)

	volumes := make(map[string]struct{}, len(vs))
	for _, v := range vs {
		volumes[v] = struct{}{}
	}
	return volumes
}

func restartPolicy(c *project.ServiceConfig) (*dockerclient.RestartPolicy, error) {
	restart, err := runconfig.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, err
	}
	return &dockerclient.RestartPolicy{Name: restart.Name, MaximumRetryCount: restart.MaximumRetryCount}, nil
}

func ports(c *project.ServiceConfig) (map[dockerclient.Port]struct{}, map[dockerclient.Port][]dockerclient.PortBinding, error) {
	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}

	exPorts, _, err := nat.ParsePortSpecs(c.Expose)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range exPorts {
		ports[k] = v
	}

	exposedPorts := map[dockerclient.Port]struct{}{}
	for k, v := range ports {
		exposedPorts[dockerclient.Port(k)] = v
	}

	portBindings := map[dockerclient.Port][]dockerclient.PortBinding{}
	for k, bv := range binding {
		dcbs := make([]dockerclient.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = dockerclient.PortBinding{HostIP: v.HostIP, HostPort: v.HostPort}
		}
		portBindings[dockerclient.Port(k)] = dcbs
	}
	return exposedPorts, portBindings, nil
}

// Convert converts a service configuration to an docker API structures (Config and HostConfig)
func Convert(c *project.ServiceConfig) (*dockerclient.Config, *dockerclient.HostConfig, error) {
	restartPolicy, err := restartPolicy(c)
	if err != nil {
		return nil, nil, err
	}

	exposedPorts, portBindings, err := ports(c)
	if err != nil {
		return nil, nil, err
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return nil, nil, err
	}

	config := &dockerclient.Config{
		Entrypoint:   utils.CopySlice(c.Entrypoint.Slice()),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          utils.CopySlice(c.Environment.Slice()),
		Cmd:          utils.CopySlice(c.Command.Slice()),
		Image:        c.Image,
		Labels:       utils.CopyMap(c.Labels.MapParts()),
		ExposedPorts: exposedPorts,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		VolumeDriver: c.VolumeDriver,
		Volumes:      volumes(c),
	}
	hostConfig := &dockerclient.HostConfig{
		VolumesFrom: utils.CopySlice(c.VolumesFrom),
		CapAdd:      utils.CopySlice(c.CapAdd),
		CapDrop:     utils.CopySlice(c.CapDrop),
		CPUShares:   c.CPUShares,
		CPUSetCPUs:  c.CPUSet,
		ExtraHosts:  utils.CopySlice(c.ExtraHosts),
		Privileged:  c.Privileged,
		Binds:       Filter(c.Volumes, isBind),
		Devices:     deviceMappings,
		DNS:         utils.CopySlice(c.DNS.Slice()),
		DNSSearch:   utils.CopySlice(c.DNSSearch.Slice()),
		LogConfig: dockerclient.LogConfig{
			Type:   c.LogDriver,
			Config: utils.CopyMap(c.LogOpt),
		},
		Memory:         c.MemLimit,
		MemorySwap:     c.MemSwapLimit,
		NetworkMode:    c.Net,
		ReadonlyRootfs: c.ReadOnly,
		PidMode:        c.Pid,
		UTSMode:        c.Uts,
		IpcMode:        c.Ipc,
		PortBindings:   portBindings,
		RestartPolicy:  *restartPolicy,
		SecurityOpt:    utils.CopySlice(c.SecurityOpt),
	}

	return config, hostConfig, nil
}

func parseDevices(devices []string) ([]dockerclient.Device, error) {
	// parse device mappings
	deviceMappings := []dockerclient.Device{}
	for _, device := range devices {
		v, err := runconfig.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, dockerclient.Device{
			PathOnHost:        v.PathOnHost,
			PathInContainer:   v.PathInContainer,
			CgroupPermissions: v.CgroupPermissions,
		})
	}

	return deviceMappings, nil
}
