package docker

import (
	dockerlib "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Tag will provide methods for interaction with API regarding tagging images
type Tag struct {
	Client dockerlib.Client
}

func (c *Tag) TagImage(image Image) error {
	options := dockerlib.TagImageOptions{
		Tag:  image.Tag,
		Repo: image.Repository,
	}

	log.Infof("Tagging image '%s' into repository '%s'", image.Name, image.Repository)
	err := c.Client.TagImage(image.ShortName, options)
	if err != nil {
		log.Errorf("Unable to tag image '%s' into repository '%s'. Error: %s", image.Name, image.Registry, err)
		return errors.New("unable to tag docker image(s)")
	}

	log.Infof("Successfully tagged image '%s'", image.Remote)
	return nil
}
