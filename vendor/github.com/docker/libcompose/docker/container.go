package docker

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	util "github.com/docker/libcompose/utils"
	dockerclient "github.com/fsouza/go-dockerclient"
)

// DefaultTag is the name of the default tag of an image.
const DefaultTag = "latest"

// Container holds information about a docker container and the service it is tied on.
// It implements Service interface by encapsulating a EmptyService.
type Container struct {
	project.EmptyService

	name    string
	service *Service
	client  *dockerclient.Client
}

// NewContainer creates a container struct with the specified docker client, name and service.
func NewContainer(client *dockerclient.Client, name string, service *Service) *Container {
	return &Container{
		client:  client,
		name:    name,
		service: service,
	}
}

func (c *Container) findExisting() (*dockerclient.APIContainers, error) {
	return GetContainerByName(c.client, c.name)
}

func (c *Container) findInfo() (*dockerclient.Container, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	return c.client.InspectContainer(container.ID)
}

// Info returns info about the container, like name, command, state or ports.
func (c *Container) Info(qFlag bool) (project.Info, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	result := project.Info{}

	if qFlag {
		result = append(result, project.InfoPart{Key: "Id", Value: container.ID})
	} else {
		result = append(result, project.InfoPart{Key: "Name", Value: name(container.Names)})
		result = append(result, project.InfoPart{Key: "Command", Value: container.Command})
		result = append(result, project.InfoPart{Key: "State", Value: container.Status})
		result = append(result, project.InfoPart{Key: "Ports", Value: portString(container.Ports)})
	}

	return result, nil
}

func portString(ports []dockerclient.APIPort) string {
	result := []string{}

	for _, port := range ports {
		if port.PublicPort > 0 {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}

	return strings.Join(result, ", ")
}

func name(names []string) string {
	max := math.MaxInt32
	var current string

	for _, v := range names {
		if len(v) < max {
			max = len(v)
			current = v
		}
	}

	return current[1:]
}

// Recreate will not refresh the container by means of relaxation and enjoyment,
// just delete it and create a new one with the current configuration
func (c *Container) Recreate(imageName string) (*dockerclient.APIContainers, error) {
	info, err := c.findInfo()
	if err != nil {
		return nil, err
	} else if info == nil {
		return nil, fmt.Errorf("Can not find container to recreate for service: %s", c.service.Name())
	}

	hash := info.Config.Labels[HASH.Str()]
	if hash == "" {
		return nil, fmt.Errorf("Failed to find hash on old container: %s", info.Name)
	}

	name := info.Name[1:]
	newName := fmt.Sprintf("%s_%s", name, info.ID[:12])
	logrus.Debugf("Renaming %s => %s", name, newName)
	if err := c.client.RenameContainer(dockerclient.RenameContainerOptions{ID: info.ID, Name: newName}); err != nil {
		logrus.Errorf("Failed to rename old container %s", c.name)
		return nil, err
	}

	newContainer, err := c.createContainer(imageName, info.ID)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Created replacement container %s", newContainer.ID)

	if err := c.client.RemoveContainer(
		dockerclient.RemoveContainerOptions{ID: info.ID, Force: true, RemoveVolumes: false}); err != nil {

		logrus.Errorf("Failed to remove old container %s", c.name)
		return nil, err
	}
	logrus.Debugf("Removed old container %s %s", c.name, info.ID)

	return newContainer, nil
}

// Create creates the container based on the specified image name and send an event
// to notify the container has been created. If the container already exists, does
// nothing.
func (c *Container) Create(imageName string) (*dockerclient.APIContainers, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	if container == nil {
		container, err = c.createContainer(imageName, "")
		if err != nil {
			return nil, err
		}
		c.service.context.Project.Notify(project.EventContainerCreated, c.service.Name(), map[string]string{
			"name": c.Name(),
		})
	}

	return container, err
}

// Down stops the container.
func (c *Container) Down() error {
	return c.withContainer(func(container *dockerclient.APIContainers) error {
		return c.client.StopContainer(container.ID, c.service.context.Timeout)
	})
}

// Pause pauses the container. If the containers are already paused, don't fail.
func (c *Container) Pause() error {
	return c.withContainer(func(container *dockerclient.APIContainers) error {
		if !strings.Contains(container.Status, "Paused") {
			return c.client.PauseContainer(container.ID)
		}
		return nil
	})
}

// Unpause unpauses the container. If the containers are not paused, don't fail.
func (c *Container) Unpause() error {
	return c.withContainer(func(container *dockerclient.APIContainers) error {
		if strings.Contains(container.Status, "Paused") {
			return c.client.UnpauseContainer(container.ID)
		}
		return nil
	})
}

// Kill kill the container.
func (c *Container) Kill() error {
	return c.withContainer(func(container *dockerclient.APIContainers) error {
		return c.client.KillContainer(dockerclient.KillContainerOptions{ID: container.ID, Signal: dockerclient.Signal(c.service.context.Signal)})
	})
}

// Delete removes the container if existing. If the container is running, it tries
// to stop it first.
func (c *Container) Delete() error {
	container, err := c.findExisting()
	if err != nil || container == nil {
		return err
	}

	info, err := c.client.InspectContainer(container.ID)
	if err != nil {
		return err
	}

	if info.State.Running {
		err := c.client.StopContainer(container.ID, c.service.context.Timeout)
		if err != nil {
			return err
		}
	}

	return c.client.RemoveContainer(dockerclient.RemoveContainerOptions{ID: container.ID, Force: true, RemoveVolumes: c.service.context.Volume})
}

// Up creates and start the container based on the image name and send an event
// to notify the container has been created. If the container exists but is stopped
// it tries to start it.
func (c *Container) Up(imageName string) error {
	var err error

	defer func() {
		if err == nil && c.service.context.Log {
			go c.Log()
		}
	}()

	container, err := c.Create(imageName)
	if err != nil {
		return err
	}

	info, err := c.client.InspectContainer(container.ID)
	if err != nil {
		return err
	}

	if !info.State.Running {
		logrus.WithFields(logrus.Fields{"container.ID": container.ID, "c.name": c.name}).Debug("Starting container")
		if err = c.client.StartContainer(container.ID, nil); err != nil {
			logrus.WithFields(logrus.Fields{"container.ID": container.ID, "c.name": c.name}).Debug("Failed to start container")
			return err
		}

		c.service.context.Project.Notify(project.EventContainerStarted, c.service.Name(), map[string]string{
			"name": c.Name(),
		})
	}

	return nil
}

// OutOfSync checks if the container is out of sync with the service definition.
// It looks if the the service hash container label is the same as the computed one.
func (c *Container) OutOfSync(imageName string) (bool, error) {
	info, err := c.findInfo()
	if err != nil || info == nil {
		return false, err
	}

	if info.Config.Image != imageName {
		logrus.Debugf("Images for %s do not match %s!=%s", c.name, info.Config.Image, imageName)
		return true, nil
	}

	if info.Config.Labels[HASH.Str()] != c.getHash() {
		logrus.Debugf("Hashes for %s do not match %s!=%s", c.name, info.Config.Labels[HASH.Str()], c.getHash())
		return true, nil
	}

	image, err := c.client.InspectImage(info.Config.Image)
	if err != nil && (err.Error() == "Not found" || image == nil) {
		logrus.Debugf("Image %s do not exist, do not know if it's out of sync", info.Config.Image)
		return false, nil
	} else if err != nil {
		return false, err
	}

	logrus.Debugf("Checking existing image name vs id: %s == %s", image.ID, info.Image)
	return image.ID != info.Image, err
}

func (c *Container) getHash() string {
	return project.GetServiceHash(c.service.Name(), c.service.Config())
}

func volumeBinds(volumes map[string]struct{}, container *dockerclient.Container) []string {
	result := make([]string, 0, len(container.Mounts))
	for _, mount := range container.Mounts {
		if _, ok := volumes[mount.Destination]; ok {
			result = append(result, fmt.Sprint(mount.Source, ":", mount.Destination))
		}
	}
	return result
}

func (c *Container) createContainer(imageName, oldContainer string) (*dockerclient.APIContainers, error) {
	createOpts, err := ConvertToAPI(c.service.serviceConfig, c.name)
	if err != nil {
		return nil, err
	}

	createOpts.Config.Image = imageName

	if createOpts.Config.Labels == nil {
		createOpts.Config.Labels = map[string]string{}
	}

	createOpts.Config.Labels[NAME.Str()] = c.name
	createOpts.Config.Labels[SERVICE.Str()] = c.service.name
	createOpts.Config.Labels[PROJECT.Str()] = c.service.context.Project.Name
	createOpts.Config.Labels[HASH.Str()] = c.getHash()

	err = c.populateAdditionalHostConfig(createOpts.HostConfig)
	if err != nil {
		return nil, err
	}

	if oldContainer != "" {
		info, err := c.client.InspectContainer(oldContainer)
		if err != nil {
			return nil, err
		}
		createOpts.HostConfig.Binds = util.Merge(createOpts.HostConfig.Binds, volumeBinds(createOpts.Config.Volumes, info))
	}

	logrus.Debugf("Creating container %s %#v", c.name, createOpts)

	container, err := c.client.CreateContainer(*createOpts)
	if err != nil && err == dockerclient.ErrNoSuchImage {
		logrus.Debugf("Not Found, pulling image %s", createOpts.Config.Image)
		if err = c.pull(createOpts.Config.Image); err != nil {
			return nil, err
		}
		if container, err = c.client.CreateContainer(*createOpts); err != nil {
			return nil, err
		}
	}

	if err != nil {
		logrus.Debugf("Failed to create container %s: %v", c.name, err)
		return nil, err
	}

	return GetContainerByID(c.client, container.ID)
}

func (c *Container) populateAdditionalHostConfig(hostConfig *dockerclient.HostConfig) error {
	links := map[string]string{}

	for _, link := range c.service.DependentServices() {
		if _, ok := c.service.context.Project.Configs[link.Target]; !ok {
			continue
		}

		service, err := c.service.context.Project.CreateService(link.Target)
		if err != nil {
			return err
		}

		containers, err := service.Containers()
		if err != nil {
			return err
		}

		if link.Type == project.RelTypeLink {
			c.addLinks(links, service, link, containers)
		} else if link.Type == project.RelTypeIpcNamespace {
			hostConfig, err = c.addIpc(hostConfig, service, containers)
		} else if link.Type == project.RelTypeNetNamespace {
			hostConfig, err = c.addNetNs(hostConfig, service, containers)
		}

		if err != nil {
			return err
		}
	}

	hostConfig.Links = []string{}
	for k, v := range links {
		hostConfig.Links = append(hostConfig.Links, strings.Join([]string{v, k}, ":"))
	}
	for _, v := range c.service.Config().ExternalLinks {
		hostConfig.Links = append(hostConfig.Links, v)
	}

	return nil
}

func (c *Container) addLinks(links map[string]string, service project.Service, rel project.ServiceRelationship, containers []project.Container) {
	for _, container := range containers {
		if _, ok := links[rel.Alias]; !ok {
			links[rel.Alias] = container.Name()
		}

		links[container.Name()] = container.Name()
	}
}

func (c *Container) addIpc(config *dockerclient.HostConfig, service project.Service, containers []project.Container) (*dockerclient.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for IPC %v", c.service.Config().Ipc)
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, err
	}

	config.IpcMode = "container:" + id
	return config, nil
}

func (c *Container) addNetNs(config *dockerclient.HostConfig, service project.Service, containers []project.Container) (*dockerclient.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for networks ns %v", c.service.Config().Net)
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, err
	}

	config.NetworkMode = "container:" + id
	return config, nil
}

// ID returns the container Id.
func (c *Container) ID() (string, error) {
	container, err := c.findExisting()
	if container == nil {
		return "", err
	}
	return container.ID, err
}

// Name returns the container name.
func (c *Container) Name() string {
	return c.name
}

// Pull pulls the image the container is based on.
func (c *Container) Pull() error {
	return c.pull(c.service.serviceConfig.Image)
}

// Restart restarts the container if existing, does nothing otherwise.
func (c *Container) Restart() error {
	container, err := c.findExisting()
	if err != nil || container == nil {
		return err
	}

	return c.client.RestartContainer(container.ID, c.service.context.Timeout)
}

// Log forwards container logs to the project configured logger.
func (c *Container) Log() error {
	container, err := c.findExisting()
	if container == nil || err != nil {
		return err
	}

	info, err := c.client.InspectContainer(container.ID)
	if info == nil || err != nil {
		return err
	}

	l := c.service.context.LoggerFactory.Create(c.name)

	err = c.client.Logs(dockerclient.LogsOptions{
		Container:    c.name,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
		Tail:         "0",
		OutputStream: &logger.Wrapper{Logger: l},
		ErrorStream:  &logger.Wrapper{Logger: l, Err: true},
		RawTerminal:  info.Config.Tty,
	})
	logrus.WithFields(logrus.Fields{"Logger": l, "err": err}).Debug("c.client.Logs() returned error")

	return err
}

func (c *Container) pull(image string) error {
	return pullImage(c.client, c.service, image)
}

func pullImage(client *dockerclient.Client, service *Service, image string) error {
	taglessRemote, tag := parsers.ParseRepositoryTag(image)
	if tag == "" {
		image = utils.ImageReference(taglessRemote, DefaultTag)
	}

	repoInfo, err := registry.ParseRepositoryInfo(taglessRemote)
	if err != nil {
		return err
	}

	authConfig := cliconfig.AuthConfig{}
	if service.context.ConfigFile != nil && repoInfo != nil && repoInfo.Index != nil {
		authConfig = registry.ResolveAuthConfig(service.context.ConfigFile, repoInfo.Index)
	}

	err = client.PullImage(
		dockerclient.PullImageOptions{
			Repository:   image,
			OutputStream: os.Stderr, // TODO maybe get the stream from some configured place
		},
		dockerclient.AuthConfiguration{
			Username: authConfig.Username,
			Password: authConfig.Password,
			Email:    authConfig.Email,
		},
	)

	if err != nil {
		logrus.Errorf("Failed to pull image %s: %v", image, err)
	}

	return err
}

func (c *Container) withContainer(action func(*dockerclient.APIContainers) error) error {
	container, err := c.findExisting()
	if err != nil {
		return err
	}

	if container != nil {
		return action(container)
	}

	return nil
}

// Port returns the host port the specified port is mapped on.
func (c *Container) Port(port string) (string, error) {
	info, err := c.findInfo()
	if err != nil {
		return "", err
	}

	if bindings, ok := info.NetworkSettings.Ports[dockerclient.Port(port)]; ok {
		result := []string{}
		for _, binding := range bindings {
			result = append(result, binding.HostIP+":"+binding.HostPort)
		}

		return strings.Join(result, "\n"), nil
	}
	return "", nil
}
