package nodes

import (
	"github.com/gonum/graph"

	kapi "k8s.io/kubernetes/pkg/api"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

func EnsureImageNode(g osgraph.MutableUniqueGraph, img *imageapi.Image) graph.Node {
	return osgraph.EnsureUnique(g,
		ImageNodeName(img),
		func(node osgraph.Node) graph.Node {
			return &ImageNode{node, img}
		},
	)
}

// EnsureAllImageStreamTagNodes creates all the ImageStreamTagNodes that are guaranteed to be present based on the ImageStream.
// This is different than inferring the presence of an object, since the IST is an object derived from a join between the ImageStream
// and the Image it references.
func EnsureAllImageStreamTagNodes(g osgraph.MutableUniqueGraph, is *imageapi.ImageStream) []*ImageStreamTagNode {
	ret := []*ImageStreamTagNode{}

	for tag := range is.Status.Tags {
		ist := &imageapi.ImageStreamTag{}
		ist.Namespace = is.Namespace
		ist.Name = imageapi.JoinImageStreamTag(is.Name, tag)

		istNode := EnsureImageStreamTagNode(g, ist)
		ret = append(ret, istNode)
	}

	return ret
}

func FindImage(g osgraph.MutableUniqueGraph, imageName string) graph.Node {
	return g.Find(ImageNodeName(&imageapi.Image{ObjectMeta: kapi.ObjectMeta{Name: imageName}}))
}

// EnsureDockerRepositoryNode adds the named Docker repository tag reference to the graph if it does
// not already exist. If the reference is invalid, the Name field of the graph will be used directly.
func EnsureDockerRepositoryNode(g osgraph.MutableUniqueGraph, name, tag string) graph.Node {
	ref, err := imageapi.ParseDockerImageReference(name)
	if err == nil {
		if len(tag) != 0 {
			ref.Tag = tag
		}
		ref = ref.DockerClientDefaults()
	} else {
		ref = imageapi.DockerImageReference{Name: name}
	}

	return osgraph.EnsureUnique(g,
		DockerImageRepositoryNodeName(ref),
		func(node osgraph.Node) graph.Node {
			return &DockerImageRepositoryNode{node, ref}
		},
	)
}

// MakeImageStreamTagObjectMeta returns an ImageStreamTag that has enough information to join the graph, but it is not
// based on a full IST object.  This can be used to properly initialize the graph without having to retrieve all ISTs
func MakeImageStreamTagObjectMeta(namespace, name, tag string) *imageapi.ImageStreamTag {
	return &imageapi.ImageStreamTag{
		ObjectMeta: kapi.ObjectMeta{
			Namespace: namespace,
			Name:      imageapi.JoinImageStreamTag(name, tag),
		},
	}
}

// MakeImageStreamTagObjectMeta2 returns an ImageStreamTag that has enough information to join the graph, but it is not
// based on a full IST object.  This can be used to properly initialize the graph without having to retrieve all ISTs
func MakeImageStreamTagObjectMeta2(namespace, name string) *imageapi.ImageStreamTag {
	return &imageapi.ImageStreamTag{
		ObjectMeta: kapi.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

// EnsureImageStreamTagNode adds a graph node for the specific tag in an Image Stream if it does not already exist.
func EnsureImageStreamTagNode(g osgraph.MutableUniqueGraph, ist *imageapi.ImageStreamTag) *ImageStreamTagNode {
	return osgraph.EnsureUnique(g,
		ImageStreamTagNodeName(ist),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamTagNode{node, ist, true}
		},
	).(*ImageStreamTagNode)
}

// FindOrCreateSyntheticImageStreamTagNode returns the existing ISTNode or creates a synthetic node in its place
func FindOrCreateSyntheticImageStreamTagNode(g osgraph.MutableUniqueGraph, ist *imageapi.ImageStreamTag) *ImageStreamTagNode {
	return osgraph.EnsureUnique(g,
		ImageStreamTagNodeName(ist),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamTagNode{node, ist, false}
		},
	).(*ImageStreamTagNode)
}

// MakeImageStreamImageObjectMeta returns an ImageStreamImage that has enough information to join the graph, but it is not
// based on a full ISI object.  This can be used to properly initialize the graph without having to retrieve all ISIs
func MakeImageStreamImageObjectMeta(namespace, name string) *imageapi.ImageStreamImage {
	return &imageapi.ImageStreamImage{
		ObjectMeta: kapi.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

// EnsureImageStreamImageNode adds a graph node for the specific ImageStreamImage if it
// does not already exist.
func EnsureImageStreamImageNode(g osgraph.MutableUniqueGraph, namespace, name string) graph.Node {
	isi := &imageapi.ImageStreamImage{
		ObjectMeta: kapi.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return osgraph.EnsureUnique(g,
		ImageStreamImageNodeName(isi),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamImageNode{node, isi, true}
		},
	)
}

// FindOrCreateSyntheticImageStreamImageNode returns the existing ISINode or creates a synthetic node in its place
func FindOrCreateSyntheticImageStreamImageNode(g osgraph.MutableUniqueGraph, isi *imageapi.ImageStreamImage) *ImageStreamImageNode {
	return osgraph.EnsureUnique(g,
		ImageStreamImageNodeName(isi),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamImageNode{node, isi, false}
		},
	).(*ImageStreamImageNode)
}

// EnsureImageStreamNode adds a graph node for the Image Stream if it does not already exist.
func EnsureImageStreamNode(g osgraph.MutableUniqueGraph, is *imageapi.ImageStream) graph.Node {
	return osgraph.EnsureUnique(g,
		ImageStreamNodeName(is),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamNode{node, is, true}
		},
	)
}

// FindOrCreateSyntheticImageStreamNode returns the existing ISNode or creates a synthetic node in its place
func FindOrCreateSyntheticImageStreamNode(g osgraph.MutableUniqueGraph, is *imageapi.ImageStream) *ImageStreamNode {
	return osgraph.EnsureUnique(g,
		ImageStreamNodeName(is),
		func(node osgraph.Node) graph.Node {
			return &ImageStreamNode{node, is, false}
		},
	).(*ImageStreamNode)
}

// EnsureImageLayerNode adds a graph node for the layer if it does not already exist.
func EnsureImageLayerNode(g osgraph.MutableUniqueGraph, layer string) graph.Node {
	return osgraph.EnsureUnique(g,
		ImageLayerNodeName(layer),
		func(node osgraph.Node) graph.Node {
			return &ImageLayerNode{node, layer}
		},
	)
}
