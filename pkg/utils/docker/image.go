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
	"path"

	dockerparser "github.com/novln/docker-parser"
)

// Image contains the basic information parsed from full image name
// see github.com/novln/docker-parser Reference
type Image struct {
	Name       string // the image's name (ie: debian[:8.2])
	ShortName  string // the image's name (ie: debian)
	Tag        string // the image's tag (or digest)
	Registry   string // the image's registry. (ie: host[:port])
	Repository string // the image's repository. (ie: registry/name)
	Remote     string // the image's remote identifier. (ie: registry/name[:tag])
}

// NewImageFromParsed method returns the docker image from the docker parser reference
func NewImageFromParsed(parsed *dockerparser.Reference) Image {
	return Image{
		Name:       parsed.Name(),
		ShortName:  parsed.ShortName(),
		Tag:        parsed.Tag(),
		Registry:   parsed.Registry(),
		Repository: parsed.Repository(),
		Remote:     parsed.Remote(),
	}
}

// ParseImage Using https://github.com/novln/docker-parser in order to parse the appropriate name and registry.
// 1. Return default registry when the registry is not specified from image
// 2. Return target registry when the registry is specified from command line
func ParseImage(fullImageName string, targetRegistry string) (Image, error) {
	var image Image

	// First parse to fill default fields for image
	// See github.com/novln/docker-parser/docker/reference.go
	parsedImage, err := dockerparser.Parse(fullImageName)

	if err != nil {
		return image, err
	}

	// Registry from command argument is high priority than parsed from name of image.
	if targetRegistry != "" {
		// Parse again for validating registry
		fullImageName = path.Join(targetRegistry, parsedImage.Name())
		parsedImage, err = dockerparser.Parse(fullImageName)
		if err != nil {
			return image, err
		}
	}

	image = NewImageFromParsed(parsedImage)

	return image, nil
}
