package docker

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/utils"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
)

// DefaultDockerfileName is the default name of a Dockerfile
const DefaultDockerfileName = "Dockerfile"

// Builder defines methods to provide a docker builder. This makes libcompose
// not tied up to the docker daemon builder.
type Builder interface {
	Build(p *project.Project, service project.Service) (string, error)
}

// DaemonBuilder is the daemon "docker build" Builder implementation.
type DaemonBuilder struct {
	context *Context
}

// NewDaemonBuilder creates a DaemonBuilder based on the specified context.
func NewDaemonBuilder(context *Context) *DaemonBuilder {
	return &DaemonBuilder{
		context: context,
	}
}

// Build implements Builder. It consumes the docker build API endpoint and sends
// a tar of the specified service build context.
func (d *DaemonBuilder) Build(p *project.Project, service project.Service) (string, error) {
	if service.Config().Build == "" {
		return service.Config().Image, nil
	}

	tag := fmt.Sprintf("%s_%s", p.Name, service.Name())
	context, err := CreateTar(p, service.Name())
	if err != nil {
		return "", err
	}

	defer context.Close()

	client := d.context.ClientFactory.Create(service)

	logrus.Infof("Building %s...", tag)

	err = client.BuildImage(dockerclient.BuildImageOptions{
		InputStream:    context,
		OutputStream:   os.Stdout,
		RawJSONStream:  false,
		Name:           tag,
		RmTmpContainer: true,
		Dockerfile:     service.Config().Dockerfile,
		NoCache:        d.context.NoCache,
	})

	if err != nil {
		return "", err
	}

	return tag, nil
}

// CreateTar create a build context tar for the specified project and service name.
func CreateTar(p *project.Project, name string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go

	serviceConfig := p.Configs[name]
	root := serviceConfig.Build
	dockerfileName := filepath.Join(root, serviceConfig.Dockerfile)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfileName == "" {
		// No -f/--file was specified so use the default
		dockerfileName = DefaultDockerfileName
		filename = filepath.Join(absRoot, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absRoot, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absRoot, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(root, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		logrus.Warnf("Error while reading .dockerignore (%s) : %s", dockerIgnorePath, err.Error())
		excludes = make([]string, 0)
	} else {
		excludes, err = utils.ReadDockerIgnore(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := utils.ValidateContextDirectory(root, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(root, options)
}
