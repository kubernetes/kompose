package util

import (
	"fmt"
	"strings"

	"github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

// ResolveImagePullSpec resolves the provided source which can be "docker", "istag" or
// "isimage" and returns the full Docker pull spec.
func ResolveImagePullSpec(images client.ImageStreamImagesNamespacer, tags client.ImageStreamTagsNamespacer, source, name, defaultNamespace string) (string, error) {
	// for Docker source, just passtrough the image name
	if IsDocker(source) {
		return name, nil
	}
	// parse the namespace from the provided image
	namespace, image := splitNamespaceAndImage(name)
	if len(namespace) == 0 {
		namespace = defaultNamespace
	}

	dockerImageReference := ""

	if IsImageStreamTag(source) {
		name, tag, ok := imageapi.SplitImageStreamTag(image)
		if !ok {
			return "", fmt.Errorf("invalid image stream tag %q, must be of the form [NAMESPACE/]NAME:TAG", name)
		}
		if resolved, err := tags.ImageStreamTags(namespace).Get(name, tag); err != nil {
			return "", fmt.Errorf("failed to get image stream tag %q: %v", name, err)
		} else {
			dockerImageReference = resolved.Image.DockerImageReference
		}
	}

	if IsImageStreamImage(source) {
		name, digest, ok := imageapi.SplitImageStreamImage(image)
		if !ok {
			return "", fmt.Errorf("invalid image stream image %q, must be of the form [NAMESPACE/]NAME@DIGEST", name)
		}
		if resolved, err := images.ImageStreamImages(namespace).Get(name, digest); err != nil {
			return "", fmt.Errorf("failed to get image stream image %q: %v", name, err)
		} else {
			dockerImageReference = resolved.Image.DockerImageReference
		}
	}

	if len(dockerImageReference) == 0 {
		return "", fmt.Errorf("unable to resolve %s %q", source, name)
	}

	reference, err := imageapi.ParseDockerImageReference(dockerImageReference)
	if err != nil {
		return "", err
	}
	return reference.String(), nil
}

func IsDocker(source string) bool {
	return source == "docker"
}

func IsImageStreamTag(source string) bool {
	return source == "istag" || source == "imagestreamtag"
}

func IsImageStreamImage(source string) bool {
	return source == "isimage" || source == "imagestreamimage"
}

func splitNamespaceAndImage(name string) (string, string) {
	namespace := ""
	imageName := ""
	if parts := strings.Split(name, "/"); len(parts) == 2 {
		namespace, imageName = parts[0], parts[1]
	} else if len(parts) == 1 {
		imageName = parts[0]
	}
	return namespace, imageName
}
