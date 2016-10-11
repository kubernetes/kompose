package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/gonum/graph"
	"github.com/gonum/graph/topo"

	"k8s.io/kubernetes/pkg/api/unversioned"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildedges "github.com/openshift/origin/pkg/build/graph"
	buildgraph "github.com/openshift/origin/pkg/build/graph/nodes"
	imageapi "github.com/openshift/origin/pkg/image/api"
	imageedges "github.com/openshift/origin/pkg/image/graph"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
)

const (
	TagNotAvailableWarning         = "ImageStreamTagNotAvailable"
	LatestBuildFailedErr           = "LatestBuildFailed"
	MissingRequiredRegistryErr     = "MissingRequiredRegistry"
	MissingOutputImageStreamErr    = "MissingOutputImageStream"
	CyclicBuildConfigWarning       = "CyclicBuildConfig"
	MissingImageStreamTagWarning   = "MissingImageStreamTag"
	MissingImageStreamImageWarning = "MissingImageStreamImage"
)

// FindUnpushableBuildConfigs checks all build configs that will output to an IST backed by an ImageStream and checks to make sure their builds can push.
func FindUnpushableBuildConfigs(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	// note, unlike with Inputs, ImageStreamImage is not a valid type for build output

bc:
	for _, bcNode := range g.NodesByKind(buildgraph.BuildConfigNodeKind) {
		for _, istNode := range g.SuccessorNodesByEdgeKind(bcNode, buildedges.BuildOutputEdgeKind) {
			for _, uncastImageStreamNode := range g.SuccessorNodesByEdgeKind(istNode, imageedges.ReferencedImageStreamGraphEdgeKind) {
				imageStreamNode := uncastImageStreamNode.(*imagegraph.ImageStreamNode)

				if !imageStreamNode.IsFound {
					markers = append(markers, osgraph.Marker{
						Node:         bcNode,
						RelatedNodes: []graph.Node{istNode},

						Severity: osgraph.ErrorSeverity,
						Key:      MissingOutputImageStreamErr,
						Message: fmt.Sprintf("%s is pushing to %s, but the image stream for that tag does not exist.",
							f.ResourceName(bcNode), f.ResourceName(istNode)),
					})

					continue
				}

				if len(imageStreamNode.Status.DockerImageRepository) == 0 {
					markers = append(markers, osgraph.Marker{
						Node:         bcNode,
						RelatedNodes: []graph.Node{istNode},

						Severity: osgraph.ErrorSeverity,
						Key:      MissingRequiredRegistryErr,
						Message: fmt.Sprintf("%s is pushing to %s, but the administrator has not configured the integrated Docker registry.",
							f.ResourceName(bcNode), f.ResourceName(istNode)),
						Suggestion: osgraph.Suggestion("oc adm registry -h"),
					})

					continue bc
				}
			}
		}
	}

	return markers
}

// FindMissingInputImageStreams checks all build configs and confirms that their From element exists
//
// Precedence of failures:
// 1. A build config's input points to an image stream that does not exist
// 2. A build config's input uses an image stream tag reference in an existing image stream, but no images within the image stream have that tag assigned
// 3. A build config's input uses an image stream image reference in an exisiting image stream, but no images within the image stream have the supplied image hexadecimal ID
func FindMissingInputImageStreams(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, bcNode := range g.NodesByKind(buildgraph.BuildConfigNodeKind) {
		for _, bcInputNode := range g.PredecessorNodesByEdgeKind(bcNode, buildedges.BuildInputImageEdgeKind) {
			switch bcInputNode.(type) {
			case *imagegraph.ImageStreamTagNode:

				for _, uncastImageStreamNode := range g.SuccessorNodesByEdgeKind(bcInputNode, imageedges.ReferencedImageStreamGraphEdgeKind) {
					imageStreamNode := uncastImageStreamNode.(*imagegraph.ImageStreamNode)

					// note, BuildConfig.Spec.BuildSpec.Strategy.[Docker|Source|Custom]Stragegy.From Input of ImageStream has been converted to ImageStreamTag on the vX to api conversion
					// prior to our reaching this point in the code; so there is not need to check for that type vs. ImageStreamTag or ImageStreamImage;

					tagNode, _ := bcInputNode.(*imagegraph.ImageStreamTagNode)
					imageStream := imageStreamNode.Object().(*imageapi.ImageStream)
					if _, ok := imageStream.Status.Tags[tagNode.ImageTag()]; !ok {

						markers = append(markers, getImageStreamTagMarker(g, f, bcInputNode, imageStreamNode, tagNode, bcNode))

					}

				}

			case *imagegraph.ImageStreamImageNode:

				for _, uncastImageStreamNode := range g.SuccessorNodesByEdgeKind(bcInputNode, imageedges.ReferencedImageStreamImageGraphEdgeKind) {
					imageStreamNode := uncastImageStreamNode.(*imagegraph.ImageStreamNode)

					imageNode, _ := bcInputNode.(*imagegraph.ImageStreamImageNode)
					imageStream := imageStreamNode.Object().(*imageapi.ImageStream)
					found, imageID := validImageStreamImage(imageNode, imageStream)
					if !found {

						markers = append(markers, getImageStreamImageMarker(g, f, bcNode, bcInputNode, imageStreamNode, imageNode, imageStream, imageID))

					}

				}

			}

		}
	}
	return markers
}

// FindCircularBuilds checks all build configs for cycles
func FindCircularBuilds(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	// Filter out all but ImageStreamTag and BuildConfig nodes
	nodeFn := osgraph.NodesOfKind(imagegraph.ImageStreamTagNodeKind, buildgraph.BuildConfigNodeKind)
	// Filter out all but BuildInputImage and BuildOutput edges
	edgeFn := osgraph.EdgesOfKind(buildedges.BuildInputImageEdgeKind, buildedges.BuildOutputEdgeKind)

	// Create desired subgraph
	sub := g.Subgraph(nodeFn, edgeFn)

	markers := []osgraph.Marker{}

	// Check for cycles
	for _, cycle := range topo.CyclesIn(sub) {
		nodeNames := []string{}
		for _, node := range cycle {
			nodeNames = append(nodeNames, f.ResourceName(node))
		}

		markers = append(markers, osgraph.Marker{
			Node:         cycle[0],
			RelatedNodes: cycle,

			Severity: osgraph.WarningSeverity,
			Key:      CyclicBuildConfigWarning,
			Message:  fmt.Sprintf("Cycle detected in build configurations: %s", strings.Join(nodeNames, " -> ")),
		})

	}

	return markers
}

// multiBCStartBuildSuggestion builds the `oc start-build` suggestion string with multiple build configs
func multiBCStartBuildSuggestion(bcNodes []*buildgraph.BuildConfigNode) string {
	var ret string
	if len(bcNodes) > 1 {
		ret = "Run one of the following commands: "
	}
	for i, bcNode := range bcNodes {
		// use of f.ResourceName(bcNode) will produce a string like  oc start-build BuildConfig|example/ruby-hello-world
		ret = ret + fmt.Sprintf("oc start-build %s", bcNode.BuildConfig.GetName())
		if i < (len(bcNodes) - 1) {
			ret = ret + ", "
		}
	}
	return ret
}

// bcNodesToRelatedNodes takes an array of BuildConfigNode's and returns an array of graph.Node's for the Marker.RelatedNodes field
func bcNodesToRelatedNodes(bcNodes []*buildgraph.BuildConfigNode) []graph.Node {
	relatedNodes := []graph.Node{}
	for _, bcNode := range bcNodes {
		relatedNodes = append(relatedNodes, graph.Node(bcNode))
	}
	return relatedNodes
}

// findPendingTagMarkers is the guts behind FindPendingTags .... break out some of the content and reduce some indentation
func findPendingTagMarkers(istNode *imagegraph.ImageStreamTagNode, g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	buildFound := false
	bcNodes := buildedges.BuildConfigsForTag(g, graph.Node(istNode))
	for _, bcNode := range bcNodes {
		latestBuild := buildedges.GetLatestBuild(g, bcNode)

		// A build config points to the non existent tag but no current build exists.
		if latestBuild == nil {
			continue
		}
		buildFound = true

		// A build config points to the non existent tag but something is going on with
		// the latest build.
		// TODO: Handle other build phases.
		switch latestBuild.Build.Status.Phase {
		case buildapi.BuildPhaseCancelled:
			// TODO: Add a warning here.
		case buildapi.BuildPhaseError:
			// TODO: Add a warning here.
		case buildapi.BuildPhaseComplete:
			// We should never hit this. The output of our build is missing but the build is complete.
			// Most probably the user has messed up?
		case buildapi.BuildPhaseFailed:
			// Since the tag hasn't been populated yet, we assume there hasn't been a successful
			// build so far.
			markers = append(markers, osgraph.Marker{
				Node:         graph.Node(latestBuild),
				RelatedNodes: []graph.Node{graph.Node(istNode), graph.Node(bcNode)},

				Severity:   osgraph.ErrorSeverity,
				Key:        LatestBuildFailedErr,
				Message:    fmt.Sprintf("%s has failed.", f.ResourceName(latestBuild)),
				Suggestion: osgraph.Suggestion(fmt.Sprintf("Inspect the build failure with 'oc logs -f bc/%s'", bcNode.BuildConfig.GetName())),
			})
		default:
			// Do nothing when latest build is new, pending, or running.
		}

	}

	// if no current builds exist for any of the build configs, append marker for that
	// but ignore ISTs which have no build configs
	if !buildFound && len(bcNodes) > 0 {
		markers = append(markers, osgraph.Marker{
			Node:         graph.Node(istNode),
			RelatedNodes: bcNodesToRelatedNodes(bcNodes),

			Severity:   osgraph.WarningSeverity,
			Key:        TagNotAvailableWarning,
			Message:    fmt.Sprintf("%s needs to be imported or created by a build.", f.ResourceName(istNode)),
			Suggestion: osgraph.Suggestion(multiBCStartBuildSuggestion(bcNodes)),
		})
	}
	return markers
}

// FindPendingTags inspects all imageStreamTags that serve as outputs to builds.
//
// Precedence of failures:
// 1. A build config points to the non existent tag but no current build exists.
// 2. A build config points to the non existent tag but the latest build has failed.
func FindPendingTags(g osgraph.Graph, f osgraph.Namer) []osgraph.Marker {
	markers := []osgraph.Marker{}

	for _, uncastIstNode := range g.NodesByKind(imagegraph.ImageStreamTagNodeKind) {
		istNode := uncastIstNode.(*imagegraph.ImageStreamTagNode)
		if !istNode.Found() {
			markers = append(markers, findPendingTagMarkers(istNode, g, f)...)
		}
	}

	return markers
}

// getImageStreamTagMarker will return the appropriate marker for when a BuildConfig is missing its input ImageStreamTag
func getImageStreamTagMarker(g osgraph.Graph, f osgraph.Namer, bcInputNode graph.Node, imageStreamNode graph.Node, tagNode *imagegraph.ImageStreamTagNode, bcNode graph.Node) osgraph.Marker {
	return osgraph.Marker{
		Node: bcNode,
		RelatedNodes: []graph.Node{bcInputNode,
			imageStreamNode},
		Severity:   osgraph.WarningSeverity,
		Key:        MissingImageStreamImageWarning,
		Message:    fmt.Sprintf("%s builds from %s, but the image stream tag does not exist.", f.ResourceName(bcNode), f.ResourceName(bcInputNode)),
		Suggestion: getImageStreamTagSuggestion(g, f, tagNode),
	}
}

// getImageStreamTagSuggestion will return the appropriate marker Suggestion for when a BuildConfig is missing its input ImageStreamTag;  in particular,
// it will determine whether or not another BuildConfig can produce the aforementioned ImageStreamTag
func getImageStreamTagSuggestion(g osgraph.Graph, f osgraph.Namer, tagNode *imagegraph.ImageStreamTagNode) osgraph.Suggestion {
	bcs := []string{}
	for _, bcNode := range g.PredecessorNodesByEdgeKind(tagNode, buildedges.BuildOutputEdgeKind) {
		bcs = append(bcs, f.ResourceName(bcNode))
	}
	if len(bcs) == 1 {
		return osgraph.Suggestion(fmt.Sprintf("oc start-build %s", bcs[0]))
	}
	if len(bcs) > 0 {
		return osgraph.Suggestion(fmt.Sprintf("`oc start-build` with one of these: %s.", strings.Join(bcs[:], ",")))
	}
	return osgraph.Suggestion(fmt.Sprintf("%s needs to be imported.", f.ResourceName(tagNode)))
}

// getImageStreamImageMarker will return the appropriate marker for when a BuildConfig is missing its input ImageStreamImage
func getImageStreamImageMarker(g osgraph.Graph, f osgraph.Namer, bcNode graph.Node, bcInputNode graph.Node, imageStreamNode graph.Node, imageNode *imagegraph.ImageStreamImageNode, imageStream *imageapi.ImageStream, imageID string) osgraph.Marker {
	return osgraph.Marker{
		Node: bcNode,
		RelatedNodes: []graph.Node{bcInputNode,
			imageStreamNode},
		Severity:   osgraph.WarningSeverity,
		Key:        MissingImageStreamImageWarning,
		Message:    fmt.Sprintf("%s builds from %s, but the image stream image does not exist.", f.ResourceName(bcNode), f.ResourceName(bcInputNode)),
		Suggestion: getImageStreamImageSuggestion(imageID, imageStream),
	}
}

// getImageStreamImageSuggestion will return the appropriate marker Suggestion for when a BuildConfig is missing its input ImageStreamImage
func getImageStreamImageSuggestion(imageID string, imageStream *imageapi.ImageStream) osgraph.Suggestion {
	// check the images stream to see if any import images are in flight or have failed
	annotation, ok := imageStream.Annotations[imageapi.DockerImageRepositoryCheckAnnotation]
	if !ok {
		return osgraph.Suggestion(fmt.Sprintf("`oc import-image %s --from=` where `--from` specifies an image with hexadecimal ID %s", imageStream.GetName(), imageID))
	}

	if checkTime, err := time.Parse(time.RFC3339, annotation); err == nil {
		// this time based annotation is set by pkg/image/controller/controller.go whenever import/tag operations are performed; unless
		// in the midst of an import/tag operation, it stays set and serves as a timestamp for when the last operation occurred;
		// so we will check if the image stream has been updated "recently";
		// in case it is a slow link to the remote repo, see if if the check annotation occurred within the last 5 minutes; if so, consider that as potentially "in progress"
		compareTime := checkTime.Add(5 * time.Minute)
		currentTime, _ := time.Parse(time.RFC3339, unversioned.Now().UTC().Format(time.RFC3339))
		if compareTime.Before(currentTime) {
			return osgraph.Suggestion(fmt.Sprintf("`oc import-image %s --from=` where `--from` specifies an image with hexadecimal ID %s", imageStream.GetName(), imageID))
		}

		return osgraph.Suggestion(fmt.Sprintf("`oc import-image %s --from=` with hexadecimal ID %s possibly in progress", imageStream.GetName(), imageID))

	}
	return osgraph.Suggestion(fmt.Sprintf("Possible error occurred with `oc import-image %s --from=` with hexadecimal ID %s; inspect images stream annotations", imageStream.GetName(), imageID))
}

// validImageStreamImage will cycle through the imageStream.Status.Tags.[]TagEvent.DockerImageReference and  determine whether an image with the hexadecimal image id
// associated with an ImageStreamImage reference in fact exists in a given ImageStream; on return, this method returns a true if does exist, and as well as the hexadecimal image
// id from the ImageStreamImage
func validImageStreamImage(imageNode *imagegraph.ImageStreamImageNode, imageStream *imageapi.ImageStream) (bool, string) {
	dockerImageReference, err := imageapi.ParseDockerImageReference(imageNode.Name)
	if err == nil {
		for _, tagEventList := range imageStream.Status.Tags {
			for _, tagEvent := range tagEventList.Items {
				if strings.Contains(tagEvent.DockerImageReference, dockerImageReference.ID) {
					return true, dockerImageReference.ID
				}
			}
		}
		return false, dockerImageReference.ID
	}
	return false, ""
}
