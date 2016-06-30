package docker

import (
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
)

// ClientFactory is a factory to create docker clients.
type ClientFactory interface {
	// Create constructs a Docker client for the given service. The passed in
	// config may be nil in which case a generic client for the project should
	// be returned.
	Create(service project.Service) *dockerclient.Client
}

type defaultClientFactory struct {
	client *dockerclient.Client
}

// NewDefaultClientFactory creates and returns the default client factory that uses
// github.com/samalba/dockerclient.
func NewDefaultClientFactory(opts ClientOpts) (ClientFactory, error) {
	client, err := CreateClient(opts)
	if err != nil {
		return nil, err
	}

	return &defaultClientFactory{
		client: client,
	}, nil
}

func (s *defaultClientFactory) Create(service project.Service) *dockerclient.Client {
	return s.client
}
