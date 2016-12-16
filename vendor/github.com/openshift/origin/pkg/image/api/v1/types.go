package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

// ImageList is a list of Image objects.
type ImageList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of images
	Items []Image `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient=true

// Image is an immutable representation of a Docker image and metadata at a point in time.
type Image struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// DockerImageReference is the string that can be used to pull this image.
	DockerImageReference string `json:"dockerImageReference,omitempty" protobuf:"bytes,2,opt,name=dockerImageReference"`
	// DockerImageMetadata contains metadata about this image
	DockerImageMetadata runtime.RawExtension `json:"dockerImageMetadata,omitempty" protobuf:"bytes,3,opt,name=dockerImageMetadata"`
	// DockerImageMetadataVersion conveys the version of the object, which if empty defaults to "1.0"
	DockerImageMetadataVersion string `json:"dockerImageMetadataVersion,omitempty" protobuf:"bytes,4,opt,name=dockerImageMetadataVersion"`
	// DockerImageManifest is the raw JSON of the manifest
	DockerImageManifest string `json:"dockerImageManifest,omitempty" protobuf:"bytes,5,opt,name=dockerImageManifest"`
	// DockerImageLayers represents the layers in the image. May not be set if the image does not define that data.
	DockerImageLayers []ImageLayer `json:"dockerImageLayers" protobuf:"bytes,6,rep,name=dockerImageLayers"`
	// Signatures holds all signatures of the image.
	Signatures []ImageSignature `json:"signatures,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=signatures"`
	// DockerImageSignatures provides the signatures as opaque blobs. This is a part of manifest schema v1.
	DockerImageSignatures [][]byte `json:"dockerImageSignatures,omitempty" protobuf:"bytes,8,rep,name=dockerImageSignatures"`
	// DockerImageManifestMediaType specifies the mediaType of manifest. This is a part of manifest schema v2.
	DockerImageManifestMediaType string `json:"dockerImageManifestMediaType,omitempty" protobuf:"bytes,9,opt,name=dockerImageManifestMediaType"`
	// DockerImageConfig is a JSON blob that the runtime uses to set up the container. This is a part of manifest schema v2.
	DockerImageConfig string `json:"dockerImageConfig,omitempty" protobuf:"bytes,10,opt,name=dockerImageConfig"`
}

// ImageLayer represents a single layer of the image. Some images may have multiple layers. Some may have none.
type ImageLayer struct {
	// Name of the layer as defined by the underlying store.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Size of the layer in bytes as defined by the underlying store.
	LayerSize int64 `json:"size" protobuf:"varint,2,opt,name=size"`
	// MediaType of the referenced object.
	MediaType string `json:"mediaType" protobuf:"bytes,3,opt,name=mediaType"`
}

// ImageSignature holds a signature of an image. It allows to verify image identity and possibly other claims
// as long as the signature is trusted. Based on this information it is possible to restrict runnable images
// to those matching cluster-wide policy.
// Mandatory fields should be parsed by clients doing image verification. The others are parsed from
// signature's content by the server. They serve just an informative purpose.
type ImageSignature struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Required: Describes a type of stored blob.
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Required: An opaque binary string which is an image's signature.
	Content []byte `json:"content" protobuf:"bytes,3,opt,name=content"`
	// Conditions represent the latest available observations of a signature's current state.
	Conditions []SignatureCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`

	// Following metadata fields will be set by server if the signature content is successfully parsed and
	// the information available.

	// A human readable string representing image's identity. It could be a product name and version, or an
	// image pull spec (e.g. "registry.access.redhat.com/rhel7/rhel:7.2").
	ImageIdentity string `json:"imageIdentity,omitempty" protobuf:"bytes,5,opt,name=imageIdentity"`
	// Contains claims from the signature.
	SignedClaims map[string]string `json:"signedClaims,omitempty" protobuf:"bytes,6,rep,name=signedClaims"`
	// If specified, it is the time of signature's creation.
	Created *unversioned.Time `json:"created,omitempty" protobuf:"bytes,7,opt,name=created"`
	// If specified, it holds information about an issuer of signing certificate or key (a person or entity
	// who signed the signing certificate or key).
	IssuedBy *SignatureIssuer `json:"issuedBy,omitempty" protobuf:"bytes,8,opt,name=issuedBy"`
	// If specified, it holds information about a subject of signing certificate or key (a person or entity
	// who signed the image).
	IssuedTo *SignatureSubject `json:"issuedTo,omitempty" protobuf:"bytes,9,opt,name=issuedTo"`
}

/// SignatureConditionType is a type of image signature condition.
type SignatureConditionType string

// SignatureCondition describes an image signature condition of particular kind at particular probe time.
type SignatureCondition struct {
	// Type of signature condition, Complete or Failed.
	Type SignatureConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=SignatureConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status kapi.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/kubernetes/pkg/api/v1.ConditionStatus"`
	// Last time the condition was checked.
	LastProbeTime unversioned.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transit from one status to another.
	LastTransitionTime unversioned.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// SignatureGenericEntity holds a generic information about a person or entity who is an issuer or a subject
// of signing certificate or key.
type SignatureGenericEntity struct {
	// Organization name.
	Organization string `json:"organization,omitempty" protobuf:"bytes,1,opt,name=organization"`
	// Common name (e.g. openshift-signing-service).
	CommonName string `json:"commonName,omitempty" protobuf:"bytes,2,opt,name=commonName"`
}

// SignatureIssuer holds information about an issuer of signing certificate or key.
type SignatureIssuer struct {
	SignatureGenericEntity `json:",inline" protobuf:"bytes,1,opt,name=signatureGenericEntity"`
}

// SignatureSubject holds information about a person or entity who created the signature.
type SignatureSubject struct {
	SignatureGenericEntity `json:",inline" protobuf:"bytes,1,opt,name=signatureGenericEntity"`
	// If present, it is a human readable key id of public key belonging to the subject used to verify image
	// signature. It should contain at least 64 lowest bits of public key's fingerprint (e.g.
	// 0x685ebe62bf278440).
	PublicKeyID string `json:"publicKeyID" protobuf:"bytes,2,opt,name=publicKeyID"`
}

// ImageStreamList is a list of ImageStream objects.
type ImageStreamList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of imageStreams
	Items []ImageStream `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// ImageStream stores a mapping of tags to images, metadata overrides that are applied
// when images are tagged in a stream, and an optional reference to a Docker image
// repository on a registry.
type ImageStream struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec describes the desired state of this stream
	Spec ImageStreamSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// Status describes the current state of this stream
	Status ImageStreamStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ImageStreamSpec represents options for ImageStreams.
type ImageStreamSpec struct {
	// DockerImageRepository is optional, if specified this stream is backed by a Docker repository on this server
	DockerImageRepository string `json:"dockerImageRepository,omitempty" protobuf:"bytes,1,opt,name=dockerImageRepository"`
	// Tags map arbitrary string values to specific image locators
	Tags []TagReference `json:"tags,omitempty" protobuf:"bytes,2,rep,name=tags"`
}

// TagReference specifies optional annotations for images using this tag and an optional reference to an ImageStreamTag, ImageStreamImage, or DockerImage this tag should track.
type TagReference struct {
	// Name of the tag
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Annotations associated with images using this tag
	Annotations map[string]string `json:"annotations" protobuf:"bytes,2,rep,name=annotations"`
	// From is a reference to an image stream tag or image stream this tag should track
	From *kapi.ObjectReference `json:"from,omitempty" protobuf:"bytes,3,opt,name=from"`
	// Reference states if the tag will be imported. Default value is false, which means the tag will be imported.
	Reference bool `json:"reference,omitempty" protobuf:"varint,4,opt,name=reference"`
	// Generation is the image stream generation that updated this tag - setting it to 0 is an indication that the generation must be updated.
	// Legacy clients will send this as nil, which means the client doesn't know or care.
	Generation *int64 `json:"generation" protobuf:"varint,5,opt,name=generation"`
	// Import is information that controls how images may be imported by the server.
	ImportPolicy TagImportPolicy `json:"importPolicy,omitempty" protobuf:"bytes,6,opt,name=importPolicy"`
}

// TagImportPolicy describes the tag import policy
type TagImportPolicy struct {
	// Insecure is true if the server may bypass certificate verification or connect directly over HTTP during image import.
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,1,opt,name=insecure"`
	// Scheduled indicates to the server that this tag should be periodically checked to ensure it is up to date, and imported
	Scheduled bool `json:"scheduled,omitempty" protobuf:"varint,2,opt,name=scheduled"`
}

// ImageStreamStatus contains information about the state of this image stream.
type ImageStreamStatus struct {
	// DockerImageRepository represents the effective location this stream may be accessed at.
	// May be empty until the server determines where the repository is located
	DockerImageRepository string `json:"dockerImageRepository" protobuf:"bytes,1,opt,name=dockerImageRepository"`
	// Tags are a historical record of images associated with each tag. The first entry in the
	// TagEvent array is the currently tagged image.
	Tags []NamedTagEventList `json:"tags,omitempty" protobuf:"bytes,2,rep,name=tags"`
}

// NamedTagEventList relates a tag to its image history.
type NamedTagEventList struct {
	// Tag is the tag for which the history is recorded
	Tag string `json:"tag" protobuf:"bytes,1,opt,name=tag"`
	// Standard object's metadata.
	Items []TagEvent `json:"items" protobuf:"bytes,2,rep,name=items"`
	// Conditions is an array of conditions that apply to the tag event list.
	Conditions []TagEventCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// TagEvent is used by ImageStreamStatus to keep a historical record of images associated with a tag.
type TagEvent struct {
	// Created holds the time the TagEvent was created
	Created unversioned.Time `json:"created" protobuf:"bytes,1,opt,name=created"`
	// DockerImageReference is the string that can be used to pull this image
	DockerImageReference string `json:"dockerImageReference" protobuf:"bytes,2,opt,name=dockerImageReference"`
	// Image is the image
	Image string `json:"image" protobuf:"bytes,3,opt,name=image"`
	// Generation is the spec tag generation that resulted in this tag being updated
	Generation int64 `json:"generation" protobuf:"varint,4,opt,name=generation"`
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
	Type TagEventConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=TagEventConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status kapi.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/kubernetes/pkg/api/v1.ConditionStatus"`
	// LastTransitionTIme is the time the condition transitioned from one status to another.
	LastTransitionTime unversioned.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// Reason is a brief machine readable explanation for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// Message is a human readable description of the details about last transition, complementing reason.
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
	// Generation is the spec tag generation that this status corresponds to
	Generation int64 `json:"generation" protobuf:"varint,6,opt,name=generation"`
}

// ImageStreamMapping represents a mapping from a single tag to a Docker image as
// well as the reference to the Docker image stream the image came from.
type ImageStreamMapping struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Image is a Docker image.
	Image Image `json:"image" protobuf:"bytes,2,opt,name=image"`
	// Tag is a string value this image can be located with inside the stream.
	Tag string `json:"tag" protobuf:"bytes,3,opt,name=tag"`
}

// ImageStreamTag represents an Image that is retrieved by tag name from an ImageStream.
type ImageStreamTag struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Tag is the spec tag associated with this image stream tag, and it may be null
	// if only pushes have occurred to this image stream.
	Tag *TagReference `json:"tag" protobuf:"bytes,2,opt,name=tag"`

	// Generation is the current generation of the tagged image - if tag is provided
	// and this value is not equal to the tag generation, a user has requested an
	// import that has not completed, or Conditions will be filled out indicating any
	// error.
	Generation int64 `json:"generation" protobuf:"varint,3,opt,name=generation"`

	// Conditions is an array of conditions that apply to the image stream tag.
	Conditions []TagEventCondition `json:"conditions,omitempty" protobuf:"bytes,4,rep,name=conditions"`

	// Image associated with the ImageStream and tag.
	Image Image `json:"image" protobuf:"bytes,5,opt,name=image"`
}

// ImageStreamTagList is a list of ImageStreamTag objects.
type ImageStreamTagList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of image stream tags
	Items []ImageStreamTag `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// ImageStreamImage represents an Image that is retrieved by image name from an ImageStream.
type ImageStreamImage struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Image associated with the ImageStream and image name.
	Image Image `json:"image" protobuf:"bytes,2,opt,name=image"`
}

// DockerImageReference points to a Docker image.
type DockerImageReference struct {
	// Registry is the registry that contains the Docker image
	Registry string `protobuf:"bytes,1,opt,name=registry"`
	// Namespace is the namespace that contains the Docker image
	Namespace string `protobuf:"bytes,2,opt,name=namespace"`
	// Name is the name of the Docker image
	Name string `protobuf:"bytes,3,opt,name=name"`
	// Tag is which tag of the Docker image is being referenced
	Tag string `protobuf:"bytes,4,opt,name=tag"`
	// ID is the identifier for the Docker image
	ID string `protobuf:"bytes,5,opt,name=iD"`
}

// The image stream import resource provides an easy way for a user to find and import Docker images
// from other Docker registries into the server. Individual images or an entire image repository may
// be imported, and users may choose to see the results of the import prior to tagging the resulting
// images into the specified image stream.
//
// This API is intended for end-user tools that need to see the metadata of the image prior to import
// (for instance, to generate an application from it). Clients that know the desired image can continue
// to create spec.tags directly into their image streams.
type ImageStreamImport struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec is a description of the images that the user wishes to import
	Spec ImageStreamImportSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// Status is the the result of importing the image
	Status ImageStreamImportStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// ImageStreamImportSpec defines what images should be imported.
type ImageStreamImportSpec struct {
	// Import indicates whether to perform an import - if so, the specified tags are set on the spec
	// and status of the image stream defined by the type meta.
	Import bool `json:"import" protobuf:"varint,1,opt,name=import"`
	// Repository is an optional import of an entire Docker image repository. A maximum limit on the
	// number of tags imported this way is imposed by the server.
	Repository *RepositoryImportSpec `json:"repository,omitempty" protobuf:"bytes,2,opt,name=repository"`
	// Images are a list of individual images to import.
	Images []ImageImportSpec `json:"images,omitempty" protobuf:"bytes,3,rep,name=images"`
}

// ImageStreamImportStatus contains information about the status of an image stream import.
type ImageStreamImportStatus struct {
	// Import is the image stream that was successfully updated or created when 'to' was set.
	Import *ImageStream `json:"import,omitempty" protobuf:"bytes,1,opt,name=import"`
	// Repository is set if spec.repository was set to the outcome of the import
	Repository *RepositoryImportStatus `json:"repository,omitempty" protobuf:"bytes,2,opt,name=repository"`
	// Images is set with the result of importing spec.images
	Images []ImageImportStatus `json:"images,omitempty" protobuf:"bytes,3,rep,name=images"`
}

// RepositoryImportSpec describes a request to import images from a Docker image repository.
type RepositoryImportSpec struct {
	// From is the source for the image repository to import; only kind DockerImage and a name of a Docker image repository is allowed
	From kapi.ObjectReference `json:"from" protobuf:"bytes,1,opt,name=from"`

	// ImportPolicy is the policy controlling how the image is imported
	ImportPolicy TagImportPolicy `json:"importPolicy,omitempty" protobuf:"bytes,2,opt,name=importPolicy"`
	// IncludeManifest determines if the manifest for each image is returned in the response
	IncludeManifest bool `json:"includeManifest,omitempty" protobuf:"varint,3,opt,name=includeManifest"`
}

// RepositoryImportStatus describes the result of an image repository import
type RepositoryImportStatus struct {
	// Status reflects whether any failure occurred during import
	Status unversioned.Status `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	// Images is a list of images successfully retrieved by the import of the repository.
	Images []ImageImportStatus `json:"images,omitempty" protobuf:"bytes,2,rep,name=images"`
	// AdditionalTags are tags that exist in the repository but were not imported because
	// a maximum limit of automatic imports was applied.
	AdditionalTags []string `json:"additionalTags,omitempty" protobuf:"bytes,3,rep,name=additionalTags"`
}

// ImageImportSpec describes a request to import a specific image.
type ImageImportSpec struct {
	// From is the source of an image to import; only kind DockerImage is allowed
	From kapi.ObjectReference `json:"from" protobuf:"bytes,1,opt,name=from"`
	// To is a tag in the current image stream to assign the imported image to, if name is not specified the default tag from from.name will be used
	To *kapi.LocalObjectReference `json:"to,omitempty" protobuf:"bytes,2,opt,name=to"`

	// ImportPolicy is the policy controlling how the image is imported
	ImportPolicy TagImportPolicy `json:"importPolicy,omitempty" protobuf:"bytes,3,opt,name=importPolicy"`
	// IncludeManifest determines if the manifest for each image is returned in the response
	IncludeManifest bool `json:"includeManifest,omitempty" protobuf:"varint,4,opt,name=includeManifest"`
}

// ImageImportStatus describes the result of an image import.
type ImageImportStatus struct {
	// Status is the status of the image import, including errors encountered while retrieving the image
	Status unversioned.Status `json:"status" protobuf:"bytes,1,opt,name=status"`
	// Image is the metadata of that image, if the image was located
	Image *Image `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// Tag is the tag this image was located under, if any
	Tag string `json:"tag,omitempty" protobuf:"bytes,3,opt,name=tag"`
}
