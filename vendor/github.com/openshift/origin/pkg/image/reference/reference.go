package reference

import (
	"strings"

	"github.com/docker/distribution/reference"
)

// NamedDockerImageReference points to a Docker image.
type NamedDockerImageReference struct {
	Registry  string
	Namespace string
	Name      string
	Tag       string
	ID        string
}

// ParseNamedDockerImageReference parses a Docker pull spec string into a
// NamedDockerImageReference.
func ParseNamedDockerImageReference(spec string) (NamedDockerImageReference, error) {
	var ref NamedDockerImageReference

	namedRef, err := reference.ParseNamed(spec)
	if err != nil {
		return ref, err
	}

	name := namedRef.Name()
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ":.") && name[:i] != "localhost") {
		ref.Name = name
	} else {
		ref.Registry, ref.Name = name[:i], name[i+1:]
	}

	if named, ok := namedRef.(reference.NamedTagged); ok {
		ref.Tag = named.Tag()
	}

	if named, ok := namedRef.(reference.Canonical); ok {
		ref.ID = named.Digest().String()
	}

	// It's not enough just to use the reference.ParseNamed(). We have to fill
	// ref.Namespace from ref.Name
	if i := strings.IndexRune(ref.Name, '/'); i != -1 {
		ref.Namespace, ref.Name = ref.Name[:i], ref.Name[i+1:]
	}

	return ref, nil
}
