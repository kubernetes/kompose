package docker

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/kubernetes-incubator/kompose/pkg/utils/archive"
)

// Build builds a Docker image from source and returns the image ID
func Build(source string, image string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		logrus.Warnf("Docker env client get error: %s", err)
	}

	tmpfile, err := ioutil.TempFile("/tmp", "kompose-image-build-")
	if err != nil {
		logrus.Warnf("Temp file create error: %s", err)
	}
	defer os.Remove(tmpfile.Name())

	// Create tarball for Docker source dir
	archive.CreateTarball(strings.Join([]string{source, ""}, "/"), tmpfile.Name())
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		logrus.Warnf("Error in creating tarball for Docker source: %s", err)
	}

	// Build Docker image
	out, err := cli.ImageBuild(ctx, file, types.ImageBuildOptions{Tags: []string{image}})
	if err != nil {
		logrus.Warnf("Docker image build error: %s", err)
	}
	io.Copy(os.Stdout, out.Body)
	out.Body.Close()
}
