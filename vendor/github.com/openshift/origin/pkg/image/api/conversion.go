package api

import (
	"github.com/fsouza/go-dockerclient"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/conversion"
)

func init() {
	err := kapi.Scheme.AddConversionFuncs(
		// Convert docker client object to internal object
		func(in *docker.Image, out *DockerImage, s conversion.Scope) error {
			if err := s.Convert(&in.Config, &out.Config, conversion.AllowDifferentFieldTypeNames); err != nil {
				return err
			}
			if err := s.Convert(&in.ContainerConfig, &out.ContainerConfig, conversion.AllowDifferentFieldTypeNames); err != nil {
				return err
			}
			out.ID = in.ID
			out.Parent = in.Parent
			out.Comment = in.Comment
			out.Created = unversioned.NewTime(in.Created)
			out.Container = in.Container
			out.DockerVersion = in.DockerVersion
			out.Author = in.Author
			out.Architecture = in.Architecture
			out.Size = in.Size
			return nil
		},
		func(in *DockerImage, out *docker.Image, s conversion.Scope) error {
			if err := s.Convert(&in.Config, &out.Config, conversion.AllowDifferentFieldTypeNames); err != nil {
				return err
			}
			if err := s.Convert(&in.ContainerConfig, &out.ContainerConfig, conversion.AllowDifferentFieldTypeNames); err != nil {
				return err
			}
			out.ID = in.ID
			out.Parent = in.Parent
			out.Comment = in.Comment
			out.Created = in.Created.Time
			out.Container = in.Container
			out.DockerVersion = in.DockerVersion
			out.Author = in.Author
			out.Architecture = in.Architecture
			out.Size = in.Size
			return nil
		},
	)
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}
