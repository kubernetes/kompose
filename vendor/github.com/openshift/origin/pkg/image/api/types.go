package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// ImageList is a list of Image objects.
type ImageList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	Items []Image
}

const (
	// ManagedByOpenShiftAnnotation indicates that an image is managed by OpenShift's registry.
	ManagedByOpenShiftAnnotation = "openshift.io/image.managed"

	// DockerImageRepositoryCheckAnnotation indicates that OpenShift has
	// attempted to import tag and image information from an external Docker
	// image repository.
	DockerImageRepositoryCheckAnnotation = "openshift.io/image.dockerRepositoryCheck"

	// InsecureRepositoryAnnotation may be set true on an image stream to allow insecure access to pull content.
	InsecureRepositoryAnnotation = "openshift.io/image.insecureRepository"

	// ExcludeImageSecretAnnotation indicates that a secret should not be returned by imagestream/secrets.
	ExcludeImageSecretAnnotation = "openshift.io/image.excludeSecret"

	// DefaultImageTag is used when an image tag is needed and the configuration does not specify a tag to use.
	DefaultImageTag = "latest"

	// ResourceImageStreams represents a number of image streams in a project.
	ResourceImageStreams kapi.ResourceName = "openshift.io/imagestreams"

	// ResourceImageStreamImages represents a number of unique references to images in all image stream
	// statuses of a project.
	ResourceImageStreamImages kapi.ResourceName = "openshift.io/images"

	// ResourceImageStreamTags represents a number of unique references to images in all image stream specs
	// of a project.
	ResourceImageStreamTags kapi.ResourceName = "openshift.io/image-tags"

	// Limit that applies to images. Used with a max["storage"] LimitRangeItem to set
	// the maximum size of an image.
	LimitTypeImage kapi.LimitType = "openshift.io/Image"

	// Limit that applies to image streams. Used with a max[resource] LimitRangeItem to set the maximum number
	// of resource. Where the resource is one of "openshift.io/images" and "openshift.io/image-tags".
	LimitTypeImageStream kapi.LimitType = "openshift.io/ImageStream"
)

// +genclient=true

// Image is an immutable representation of a Docker image and metadata at a point in time.
type Image struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// The string that can be used to pull this image.
	DockerImageReference string
	// Metadata about this image
	DockerImageMetadata DockerImage
	// This attribute conveys the version of docker metadata the JSON should be stored in, which if empty defaults to "1.0"
	DockerImageMetadataVersion string
	// The raw JSON of the manifest
	DockerImageManifest string
	// DockerImageLayers represents the layers in the image. May not be set if the image does not define that data.
	DockerImageLayers []ImageLayer
	// Signatures holds all signatures of the image.
	Signatures []ImageSignature
	// DockerImageSignatures provides the signatures as opaque blobs. This is a part of manifest schema v1.
	DockerImageSignatures [][]byte
	// DockerImageManifestMediaType specifies the mediaType of manifest. This is a part of manifest schema v2.
	DockerImageManifestMediaType string
	// DockerImageConfig is a JSON blob that the runtime uses to set up the container. This is a part of manifest schema v2.
	DockerImageConfig string
}

// ImageLayer represents a single layer of the image. Some images may have multiple layers. Some may have none.
type ImageLayer struct {
	// Name of the layer as defined by the underlying store.
	Name string
	// LayerSize of the layer as defined by the underlying store.
	LayerSize int64
	// MediaType of the referenced object.
	MediaType string
}

const (
	// The supported type of image signature.
	ImageSignatureTypeAtomicImageV1 string = "AtomicImageV1"
)

// ImageSignature holds a signature of an image. It allows to verify image identity and possibly other claims
// as long as the signature is trusted. Based on this information it is possible to restrict runnable images
// to those matching cluster-wide policy.
// Mandatory fields should be parsed by clients doing image verification. The others are parsed from
// signature's content by the server. They serve just an informative purpose.
type ImageSignature struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Required: Describes a type of stored blob.
	Type string
	// Required: An opaque binary string which is an image's signature.
	Content []byte
	// Conditions represent the latest available observations of a signature's current state.
	Conditions []SignatureCondition

	// Following metadata fields will be set by server if the signature content is successfully parsed and
	// the information available.

	// A human readable string representing image's identity. It could be a product name and version, or an
	// image pull spec (e.g. "registry.access.redhat.com/rhel7/rhel:7.2").
	ImageIdentity string
	// Contains claims from the signature.
	SignedClaims map[string]string
	// If specified, it is the time of signature's creation.
	Created *unversioned.Time
	// If specified, it holds information about an issuer of signing certificate or key (a person or entity
	// who signed the signing certificate or key).
	IssuedBy *SignatureIssuer
	// If specified, it holds information about a subject of signing certificate or key (a person or entity
	// who signed the image).
	IssuedTo *SignatureSubject
}

// These are valid conditions of an image signature.
const (
	// SignatureTrusted means the signing key or certificate was valid and the signature matched the image at
	// the probe time.
	SignatureTrusted = "Trusted"
	// SignatureForImage means the signature matches image object containing it.
	SignatureForImage = "ForImage"
	// SignatureExpired means the signature or its signing key or certificate had been expired at the probe
	// time.
	SignatureExpired = "Expired"
	// SignatureRevoked means the signature or its signing key or certificate has been revoked.
	SignatureRevoked = "Revoked"
)

/// SignatureConditionType is a type of image signature condition.
type SignatureConditionType string

// SignatureCondition describes an image signature condition of particular kind at particular probe time.
type SignatureCondition struct {
	// Type of signature condition, Complete or Failed.
	Type SignatureConditionType
	// Status of the condition, one of True, False, Unknown.
	Status kapi.ConditionStatus
	// Last time the condition was checked.
	LastProbeTime unversioned.Time
	// Last time the condition transit from one status to another.
	LastTransitionTime unversioned.Time
	// (brief) reason for the condition's last transition.
	Reason string
	// Human readable message indicating details about last transition.
	Message string
}

// SignatureGenericEntity holds a generic information about a person or entity who is an issuer or a subject
// of signing certificate or key.
type SignatureGenericEntity struct {
	// Organization name.
	Organization string
	// Common name (e.g. openshift-signing-service).
	CommonName string
}

// SignatureIssuer holds information about an issuer of signing certificate or key.
type SignatureIssuer struct {
	SignatureGenericEntity
}

// SignatureSubject holds information about a person or entity who created the signature.
type SignatureSubject struct {
	SignatureGenericEntity
	// If present, it is a human readable key id of public key belonging to the subject used to verify image
	// signature. It should contain at least 64 lowest bits of public key's fingerprint (e.g.
	// 0x685ebe62bf278440).
	PublicKeyID string
}

// ImageStreamList is a list of ImageStream objects.
type ImageStreamList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	Items []ImageStream
}

// ImageStream stores a mapping of tags to images, metadata overrides that are applied
// when images are tagged in a stream, and an optional reference to a Docker image
// repository on a registry.
type ImageStream struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Spec describes the desired state of this stream
	Spec ImageStreamSpec
	// Status describes the current state of this stream
	Status ImageStreamStatus
}

// ImageStreamSpec represents options for ImageStreams.
type ImageStreamSpec struct {
	// Optional, if specified this stream is backed by a Docker repository on this server
	DockerImageRepository string
	// Tags map arbitrary string values to specific image locators
	Tags map[string]TagReference
}

// TagReference specifies optional annotations for images using this tag and an optional reference to
// an ImageStreamTag, ImageStreamImage, or DockerImage this tag should track.
type TagReference struct {
	// Name of the tag
	Name string
	// Optional; if specified, annotations that are applied to images retrieved via ImageStreamTags.
	Annotations map[string]string
	// Optional; if specified, a reference to another image that this tag should point to. Valid values
	// are ImageStreamTag, ImageStreamImage, and DockerImage.
	From *kapi.ObjectReference
	// Reference states if the tag will be imported. Default value is false, which means the tag will
	// be imported.
	Reference bool
	// Generation is a counter that tracks mutations to the spec tag (user intent). When a tag reference
	// is changed the generation is set to match the current stream generation (which is incremented every
	// time spec is changed). Other processes in the system like the image importer observe that the
	// generation of spec tag is newer than the generation recorded in the status and use that as a trigger
	// to import the newest remote tag. To trigger a new import, clients may set this value to zero which
	// will reset the generation to the latest stream generation. Legacy clients will send this value as
	// nil which will be merged with the current tag generation.
	Generation *int64
	// ImportPolicy is information that controls how images may be imported by the server.
	ImportPolicy TagImportPolicy
}

type TagImportPolicy struct {
	// Insecure is true if the server may bypass certificate verification or connect directly over HTTP during image import.
	Insecure bool
	// Scheduled indicates to the server that this tag should be periodically checked to ensure it is up to date, and imported
	Scheduled bool
}

// ImageStreamStatus contains information about the state of this image stream.
type ImageStreamStatus struct {
	// DockerImageRepository represents the effective location this stream may be accessed at. May be empty until the server
	// determines where the repository is located
	DockerImageRepository string
	// A historical record of images associated with each tag. The first entry in the TagEvent array is
	// the currently tagged image.
	Tags map[string]TagEventList
}

// TagEventList contains a historical record of images associated with a tag.
type TagEventList struct {
	Items []TagEvent
	// Conditions is an array of conditions that apply to the tag event list.
	Conditions []TagEventCondition
}

// TagEvent is used by ImageRepositoryStatus to keep a historical record of images associated with a tag.
type TagEvent struct {
	// When the TagEvent was created
	Created unversioned.Time
	// The string that can be used to pull this image
	DockerImageReference string
	// The image
	Image string
	// Generation is the spec tag generation that resulted in this tag being updated
	Generation int64
}

type TagEventConditionType string

// These are valid conditions of TagEvents.
const (
	// ImportSuccess with status False means the import of the specific tag failed
	ImportSuccess TagEventConditionType = "ImportSuccess"
)

// TagEventCondition contains condition information for a tag event.
type TagEventCondition struct {
	// Type of tag event condition, currently only ImportSuccess
	Type TagEventConditionType
	// Status of the condition, one of True, False, Unknown.
	Status kapi.ConditionStatus
	// LastTransitionTIme is the time the condition transitioned from one status to another.
	LastTransitionTime unversioned.Time
	// Reason is a brief machine readable explanation for the condition's last transition.
	Reason string
	// Message is a human readable description of the details about last transition, complementing reason.
	Message string
	// Generation is the spec tag generation that this status corresponds to. If this value is
	// older than the spec tag generation, the user has requested this status tag be updated.
	// This value is set to zero for older versions of streams, which means that no generation
	// was recorded.
	Generation int64
}

// ImageStreamMapping represents a mapping from a single tag to a Docker image as
// well as the reference to the Docker image repository the image came from.
type ImageStreamMapping struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// The Docker image repository the specified image is located in
	// DEPRECATED: remove once v1beta1 support is dropped
	// +k8s:conversion-gen=false
	DockerImageRepository string
	// A Docker image.
	Image Image
	// A string value this image can be located with inside the repository.
	Tag string
}

// ImageStreamTag has a .Name in the format <stream name>:<tag>.
type ImageStreamTag struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Tag is the spec tag associated with this image stream tag, and it may be null
	// if only pushes have occurred to this image stream.
	Tag *TagReference

	// Generation is the current generation of the tagged image - if tag is provided
	// and this value is not equal to the tag generation, a user has requested an
	// import that has not completed, or Conditions will be filled out indicating any
	// error.
	Generation int64

	// Conditions is an array of conditions that apply to the image stream tag.
	Conditions []TagEventCondition

	// The Image associated with the ImageStream and tag.
	Image Image
}

// ImageStreamTagList is a list of ImageStreamTag objects.
type ImageStreamTagList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	Items []ImageStreamTag
}

// ImageStreamImage represents an Image that is retrieved by image name from an ImageStream.
type ImageStreamImage struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// The Image associated with the ImageStream and image name.
	Image Image
}

// DockerImageReference points to a Docker image.
type DockerImageReference struct {
	Registry  string
	Namespace string
	Name      string
	Tag       string
	ID        string
}

// ImageStreamImport allows a caller to request information about a set of images for possible
// import into an image stream, or actually tag the images into the image stream.
type ImageStreamImport struct {
	unversioned.TypeMeta
	// ObjectMeta must identify the name of the image stream to create or update. If resourceVersion
	// or UID are set, they must match the image stream that will be loaded from the server.
	kapi.ObjectMeta

	// Spec is the set of items desired to be imported
	Spec ImageStreamImportSpec
	// Status is the result of the import
	Status ImageStreamImportStatus
}

// ImageStreamImportSpec defines what images should be imported.
type ImageStreamImportSpec struct {
	// Import indicates whether to perform an import - if so, the specified tags are set on the spec
	// and status of the image stream defined by the type meta.
	Import bool
	// Repository is an optional import of an entire Docker image repository. A maximum limit on the
	// number of tags imported this way is imposed by the server.
	Repository *RepositoryImportSpec
	// Images are a list of individual images to import.
	Images []ImageImportSpec
}

// ImageStreamImportStatus contains information about the status of an image stream import.
type ImageStreamImportStatus struct {
	// Import is the image stream that was successfully updated or created when 'to' was set.
	Import *ImageStream
	// Repository is set if spec.repository was set to the outcome of the import
	Repository *RepositoryImportStatus
	// Images is set with the result of importing spec.images
	Images []ImageImportStatus
}

// RepositoryImport indicates to load a set of tags from a given Docker image repository
type RepositoryImportSpec struct {
	// The source of the import, only kind DockerImage is supported
	From kapi.ObjectReference

	ImportPolicy    TagImportPolicy
	IncludeManifest bool
}

// RepositoryImportStatus describes the outcome of the repository import
type RepositoryImportStatus struct {
	// Status reflects whether any failure occurred during import
	Status unversioned.Status
	// Images is the list of imported images
	Images []ImageImportStatus
	// AdditionalTags are tags that exist in the repository but were not imported because
	// a maximum limit of automatic imports was applied.
	AdditionalTags []string
}

// ImageImportSpec defines how an image is imported.
type ImageImportSpec struct {
	From kapi.ObjectReference
	To   *kapi.LocalObjectReference

	ImportPolicy    TagImportPolicy
	IncludeManifest bool
}

// ImageImportStatus describes the result of an image import.
type ImageImportStatus struct {
	Tag    string
	Status unversioned.Status
	Image  *Image
}
