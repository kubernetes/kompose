package docker

import (
	b64 "encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Push pushes a Docker image to a specided Docker registry in the image name
func Push(image string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		logrus.Warnf("Docker client get error: %s", err)
	}
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	fmt.Println("Username: ", username, "Password: ", password)
	registryAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	fmt.Println("Registry auth", registryAuth)

	logrus.Infof("Pushing Docker image: %s", image)
	// FIXME: base64 encoded value of registry credentials shoule be passed in
	// RegistryAuth in types.ImagePushOptions. Challenges are retrieving this
	// value automatically. It can be retrieved from ~/.docker/config.json
	// or by explicitly asking for username/password from the user
	out, err := cli.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: registryAuth})
	if err != nil {
		logrus.Warnf("Docker image push error: %s", err)
	}
	io.Copy(os.Stdout, out)
	out.Close()
}
