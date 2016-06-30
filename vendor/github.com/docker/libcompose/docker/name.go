package docker

import (
	"fmt"
	"io"
	"time"

	dockerclient "github.com/fsouza/go-dockerclient"
)

const format = "%s_%s_%d"

// Namer defines method to provide container name.
type Namer interface {
	io.Closer
	Next() string
}

type inOrderNamer struct {
	names chan string
	done  chan bool
}

type singleNamer struct {
	name string
}

// NewSingleNamer returns a namer that only allows a single name.
func NewSingleNamer(name string) Namer {
	return &singleNamer{name}
}

// NewNamer returns a namer that returns names based on the specified project and
// service name and an inner counter, e.g. project_service_1, project_service_2…
func NewNamer(client *dockerclient.Client, project, service string) Namer {
	namer := &inOrderNamer{
		names: make(chan string),
		done:  make(chan bool),
	}

	go func() {
		for i := 1; true; i++ {
			name := fmt.Sprintf(format, project, service, i)
			c, err := GetContainerByName(client, name)
			if err != nil {
				// Sleep here to avoid crazy tight loop when things go south
				time.Sleep(time.Second * 1)
				continue
			}
			if c != nil {
				continue
			}

			select {
			case namer.names <- name:
			case <-namer.done:
				close(namer.names)
				return
			}
		}
	}()

	return namer
}

func (i *inOrderNamer) Next() string {
	return <-i.names
}

func (i *inOrderNamer) Close() error {
	close(i.done)
	return nil
}

func (s *singleNamer) Next() string {
	return s.name
}

func (s *singleNamer) Close() error {
	return nil
}
