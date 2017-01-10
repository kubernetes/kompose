package utils

import (
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
)

// DockerBuildImage Build and tag Docker image
func DockerBuildImage(name string, service kobject.ServiceConfig, composeFileDir string) string {
	image := name
	if service.Image != "" {
		image = service.Image
	}

	cmd := NewCommand(fmt.Sprintf("docker build -t %s %s", image, service.Build))
	cmd.Dir = composeFileDir

	out, err := Execute(cmd)

	logrus.Infof("Building image for service %s: %s", name, out)

	if err != nil {
		logrus.Errorf("Error during building image for service %s: %s", name, err)
	}

	return image
}

// DockerPushImage Push Docker image
func DockerPushImage(name string, image string) {
	cmd := NewCommand(fmt.Sprintf("docker push %s", image))
	out, err := Execute(cmd)

	logrus.Infof("Image push logs for service %s: %s", name, out)

	if err != nil {
		logrus.Errorf("Error during pushing image '%s' for service '%s'", image, name)
	}
}
