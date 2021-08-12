/*
Copyright 2016 The Kubernetes Authors All rights reserved

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
