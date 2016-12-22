package v1

import (
	"fmt"
	"time"

	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

// +genclient=true

// Build encapsulates the inputs needed to produce a new deployable image, as well as
// the status of the execution and a reference to the Pod which executed the build.
type Build struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec is all the inputs used to execute the build.
	Spec BuildSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// status is the current status of the build.
	Status BuildStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// BuildSpec has the information to represent a build and also additional
// information about a build
type BuildSpec struct {
	// CommonSpec is the information that represents a build
	CommonSpec `json:",inline" protobuf:"bytes,1,opt,name=commonSpec"`

	// triggeredBy describes which triggers started the most recent update to the
	// build configuration and contains information about those triggers.
	TriggeredBy []BuildTriggerCause `json:"triggeredBy" protobuf:"bytes,2,rep,name=triggeredBy"`
}

// OptionalNodeSelector is a map that may also be left nil to distinguish between set and unset.
// +protobuf.nullable=true
// +protobuf.options.(gogoproto.goproto_stringer)=false
type OptionalNodeSelector map[string]string

func (t OptionalNodeSelector) String() string {
	return fmt.Sprintf("%v", map[string]string(t))
}

// CommonSpec encapsulates all the inputs necessary to represent a build.
type CommonSpec struct {
	// serviceAccount is the name of the ServiceAccount to use to run the pod
	// created by this build.
	// The pod will be allowed to use secrets referenced by the ServiceAccount
	ServiceAccount string `json:"serviceAccount,omitempty" protobuf:"bytes,1,opt,name=serviceAccount"`

	// source describes the SCM in use.
	Source BuildSource `json:"source,omitempty" protobuf:"bytes,2,opt,name=source"`

	// revision is the information from the source for a specific repo snapshot.
	// This is optional.
	Revision *SourceRevision `json:"revision,omitempty" protobuf:"bytes,3,opt,name=revision"`

	// strategy defines how to perform a build.
	Strategy BuildStrategy `json:"strategy" protobuf:"bytes,4,opt,name=strategy"`

	// output describes the Docker image the Strategy should produce.
	Output BuildOutput `json:"output,omitempty" protobuf:"bytes,5,opt,name=output"`

	// resources computes resource requirements to execute the build.
	Resources kapi.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,6,opt,name=resources"`

	// postCommit is a build hook executed after the build output image is
	// committed, before it is pushed to a registry.
	PostCommit BuildPostCommitSpec `json:"postCommit,omitempty" protobuf:"bytes,7,opt,name=postCommit"`

	// completionDeadlineSeconds is an optional duration in seconds, counted from
	// the time when a build pod gets scheduled in the system, that the build may
	// be active on a node before the system actively tries to terminate the
	// build; value must be positive integer
	CompletionDeadlineSeconds *int64 `json:"completionDeadlineSeconds,omitempty" protobuf:"varint,8,opt,name=completionDeadlineSeconds"`

	// nodeSelector is a selector which must be true for the build pod to fit on a node
	// If nil, it can be overridden by default build nodeselector values for the cluster.
	// If set to an empty map or a map with any values, default build nodeselector values
	// are ignored.
	NodeSelector OptionalNodeSelector `json:"nodeSelector" protobuf:"bytes,9,name=nodeSelector"`
}

// BuildTriggerCause holds information about a triggered build. It is used for
// displaying build trigger data for each build and build configuration in oc
// describe. It is also used to describe which triggers led to the most recent
// update in the build configuration.
type BuildTriggerCause struct {
	// message is used to store a human readable message for why the build was
	// triggered. E.g.: "Manually triggered by user", "Configuration change",etc.
	Message string `json:"message,omitempty" protobuf:"bytes,1,opt,name=message"`

	// genericWebHook holds data about a builds generic webhook trigger.
	GenericWebHook *GenericWebHookCause `json:"genericWebHook,omitempty" protobuf:"bytes,2,opt,name=genericWebHook"`

	// gitHubWebHook represents data for a GitHub webhook that fired a
	//specific build.
	GitHubWebHook *GitHubWebHookCause `json:"githubWebHook,omitempty" protobuf:"bytes,3,opt,name=githubWebHook"`

	// imageChangeBuild stores information about an imagechange event
	// that triggered a new build.
	ImageChangeBuild *ImageChangeCause `json:"imageChangeBuild,omitempty" protobuf:"bytes,4,opt,name=imageChangeBuild"`
}

// GenericWebHookCause holds information about a generic WebHook that
// triggered a build.
type GenericWebHookCause struct {
	// revision is an optional field that stores the git source revision
	// information of the generic webhook trigger when it is available.
	Revision *SourceRevision `json:"revision,omitempty" protobuf:"bytes,1,opt,name=revision"`

	// secret is the obfuscated webhook secret that triggered a build.
	Secret string `json:"secret,omitempty" protobuf:"bytes,2,opt,name=secret"`
}

// GitHubWebHookCause has information about a GitHub webhook that triggered a
// build.
type GitHubWebHookCause struct {
	// revision is the git revision information of the trigger.
	Revision *SourceRevision `json:"revision,omitempty" protobuf:"bytes,1,opt,name=revision"`

	// secret is the obfuscated webhook secret that triggered a build.
	Secret string `json:"secret,omitempty" protobuf:"bytes,2,opt,name=secret"`
}

// ImageChangeCause contains information about the image that triggered a
// build
type ImageChangeCause struct {
	// imageID is the ID of the image that triggered a a new build.
	ImageID string `json:"imageID,omitempty" protobuf:"bytes,1,opt,name=imageID"`

	// fromRef contains detailed information about an image that triggered a
	// build.
	FromRef *kapi.ObjectReference `json:"fromRef,omitempty" protobuf:"bytes,2,opt,name=fromRef"`
}

// BuildStatus contains the status of a build
type BuildStatus struct {
	// phase is the point in the build lifecycle.
	Phase BuildPhase `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=BuildPhase"`

	// cancelled describes if a cancel event was triggered for the build.
	Cancelled bool `json:"cancelled,omitempty" protobuf:"varint,2,opt,name=cancelled"`

	// reason is a brief CamelCase string that describes any failure and is meant for machine parsing and tidy display in the CLI.
	Reason StatusReason `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason,casttype=StatusReason"`

	// message is a human-readable message indicating details about why the build has this status.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`

	// startTimestamp is a timestamp representing the server time when this Build started
	// running in a Pod.
	// It is represented in RFC3339 form and is in UTC.
	StartTimestamp *unversioned.Time `json:"startTimestamp,omitempty" protobuf:"bytes,5,opt,name=startTimestamp"`

	// completionTimestamp is a timestamp representing the server time when this Build was
	// finished, whether that build failed or succeeded.  It reflects the time at which
	// the Pod running the Build terminated.
	// It is represented in RFC3339 form and is in UTC.
	CompletionTimestamp *unversioned.Time `json:"completionTimestamp,omitempty" protobuf:"bytes,6,opt,name=completionTimestamp"`

	// duration contains time.Duration object describing build time.
	Duration time.Duration `json:"duration,omitempty" protobuf:"varint,7,opt,name=duration,casttype=time.Duration"`

	// outputDockerImageReference contains a reference to the Docker image that
	// will be built by this build. Its value is computed from
	// Build.Spec.Output.To, and should include the registry address, so that
	// it can be used to push and pull the image.
	OutputDockerImageReference string `json:"outputDockerImageReference,omitempty" protobuf:"bytes,8,opt,name=outputDockerImageReference"`

	// config is an ObjectReference to the BuildConfig this Build is based on.
	Config *kapi.ObjectReference `json:"config,omitempty" protobuf:"bytes,9,opt,name=config"`
}

// BuildPhase represents the status of a build at a point in time.
type BuildPhase string

// Valid values for BuildPhase.
const (
	// BuildPhaseNew is automatically assigned to a newly created build.
	BuildPhaseNew BuildPhase = "New"

	// BuildPhasePending indicates that a pod name has been assigned and a build is
	// about to start running.
	BuildPhasePending BuildPhase = "Pending"

	// BuildPhaseRunning indicates that a pod has been created and a build is running.
	BuildPhaseRunning BuildPhase = "Running"

	// BuildPhaseComplete indicates that a build has been successful.
	BuildPhaseComplete BuildPhase = "Complete"

	// BuildPhaseFailed indicates that a build has executed and failed.
	BuildPhaseFailed BuildPhase = "Failed"

	// BuildPhaseError indicates that an error prevented the build from executing.
	BuildPhaseError BuildPhase = "Error"

	// BuildPhaseCancelled indicates that a running/pending build was stopped from executing.
	BuildPhaseCancelled BuildPhase = "Cancelled"
)

// StatusReason is a brief CamelCase string that describes a temporary or
// permanent build error condition, meant for machine parsing and tidy display
// in the CLI.
type StatusReason string

// BuildSourceType is the type of SCM used.
type BuildSourceType string

// Valid values for BuildSourceType.
const (
	//BuildSourceGit instructs a build to use a Git source control repository as the build input.
	BuildSourceGit BuildSourceType = "Git"
	// BuildSourceDockerfile uses a Dockerfile as the start of a build
	BuildSourceDockerfile BuildSourceType = "Dockerfile"
	// BuildSourceBinary indicates the build will accept a Binary file as input.
	BuildSourceBinary BuildSourceType = "Binary"
	// BuildSourceImage indicates the build will accept an image as input
	BuildSourceImage BuildSourceType = "Image"
	// BuildSourceNone indicates the build has no predefined input (only valid for Source and Custom Strategies)
	BuildSourceNone BuildSourceType = "None"
)

// BuildSource is the SCM used for the build.
type BuildSource struct {
	// type of build input to accept
	// +k8s:conversion-gen=false
	Type BuildSourceType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=BuildSourceType"`

	// binary builds accept a binary as their input. The binary is generally assumed to be a tar,
	// gzipped tar, or zip file depending on the strategy. For Docker builds, this is the build
	// context and an optional Dockerfile may be specified to override any Dockerfile in the
	// build context. For Source builds, this is assumed to be an archive as described above. For
	// Source and Docker builds, if binary.asFile is set the build will receive a directory with
	// a single file. contextDir may be used when an archive is provided. Custom builds will
	// receive this binary as input on STDIN.
	Binary *BinaryBuildSource `json:"binary,omitempty" protobuf:"bytes,2,opt,name=binary"`

	// dockerfile is the raw contents of a Dockerfile which should be built. When this option is
	// specified, the FROM may be modified based on your strategy base image and additional ENV
	// stanzas from your strategy environment will be added after the FROM, but before the rest
	// of your Dockerfile stanzas. The Dockerfile source type may be used with other options like
	// git - in those cases the Git repo will have any innate Dockerfile replaced in the context
	// dir.
	Dockerfile *string `json:"dockerfile,omitempty" protobuf:"bytes,3,opt,name=dockerfile"`

	// git contains optional information about git build source
	Git *GitBuildSource `json:"git,omitempty" protobuf:"bytes,4,opt,name=git"`

	// images describes a set of images to be used to provide source for the build
	Images []ImageSource `json:"images,omitempty" protobuf:"bytes,5,rep,name=images"`

	// contextDir specifies the sub-directory where the source code for the application exists.
	// This allows to have buildable sources in directory other than root of
	// repository.
	ContextDir string `json:"contextDir,omitempty" protobuf:"bytes,6,opt,name=contextDir"`

	// sourceSecret is the name of a Secret that would be used for setting
	// up the authentication for cloning private repository.
	// The secret contains valid credentials for remote repository, where the
	// data's key represent the authentication method to be used and value is
	// the base64 encoded credentials. Supported auth methods are: ssh-privatekey.
	SourceSecret *kapi.LocalObjectReference `json:"sourceSecret,omitempty" protobuf:"bytes,7,opt,name=sourceSecret"`

	// secrets represents a list of secrets and their destinations that will
	// be used only for the build.
	Secrets []SecretBuildSource `json:"secrets,omitempty" protobuf:"bytes,8,rep,name=secrets"`
}

// ImageSource is used to describe build source that will be extracted from an image. A reference of
// type ImageStreamTag, ImageStreamImage or DockerImage may be used. A pull secret can be specified
// to pull the image from an external registry or override the default service account secret if pulling
// from the internal registry. A list of paths to copy from the image and their respective destination
// within the build directory must be specified in the paths array.
type ImageSource struct {
	// from is a reference to an ImageStreamTag, ImageStreamImage, or DockerImage to
	// copy source from.
	From kapi.ObjectReference `json:"from" protobuf:"bytes,1,opt,name=from"`

	// paths is a list of source and destination paths to copy from the image.
	Paths []ImageSourcePath `json:"paths" protobuf:"bytes,2,rep,name=paths"`

	// pullSecret is a reference to a secret to be used to pull the image from a registry
	// If the image is pulled from the OpenShift registry, this field does not need to be set.
	PullSecret *kapi.LocalObjectReference `json:"pullSecret,omitempty" protobuf:"bytes,3,opt,name=pullSecret"`
}

// ImageSourcePath describes a path to be copied from a source image and its destination within the build directory.
type ImageSourcePath struct {
	// sourcePath is the absolute path of the file or directory inside the image to
	// copy to the build directory.
	SourcePath string `json:"sourcePath" protobuf:"bytes,1,opt,name=sourcePath"`

	// destinationDir is the relative directory within the build directory
	// where files copied from the image are placed.
	DestinationDir string `json:"destinationDir" protobuf:"bytes,2,opt,name=destinationDir"`
}

// SecretBuildSource describes a secret and its destination directory that will be
// used only at the build time. The content of the secret referenced here will
// be copied into the destination directory instead of mounting.
type SecretBuildSource struct {
	// secret is a reference to an existing secret that you want to use in your
	// build.
	Secret kapi.LocalObjectReference `json:"secret" protobuf:"bytes,1,opt,name=secret"`

	// destinationDir is the directory where the files from the secret should be
	// available for the build time.
	// For the Source build strategy, these will be injected into a container
	// where the assemble script runs. Later, when the script finishes, all files
	// injected will be truncated to zero length.
	// For the Docker build strategy, these will be copied into the build
	// directory, where the Dockerfile is located, so users can ADD or COPY them
	// during docker build.
	DestinationDir string `json:"destinationDir,omitempty" protobuf:"bytes,2,opt,name=destinationDir"`
}

// BinaryBuildSource describes a binary file to be used for the Docker and Source build strategies,
// where the file will be extracted and used as the build source.
type BinaryBuildSource struct {
	// asFile indicates that the provided binary input should be considered a single file
	// within the build input. For example, specifying "webapp.war" would place the provided
	// binary as `/webapp.war` for the builder. If left empty, the Docker and Source build
	// strategies assume this file is a zip, tar, or tar.gz file and extract it as the source.
	// The custom strategy receives this binary as standard input. This filename may not
	// contain slashes or be '..' or '.'.
	AsFile string `json:"asFile,omitempty" protobuf:"bytes,1,opt,name=asFile"`
}

// SourceRevision is the revision or commit information from the source for the build
type SourceRevision struct {
	// type of the build source, may be one of 'Source', 'Dockerfile', 'Binary', or 'Images'
	// +k8s:conversion-gen=false
	Type BuildSourceType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=BuildSourceType"`

	// Git contains information about git-based build source
	Git *GitSourceRevision `json:"git,omitempty" protobuf:"bytes,2,opt,name=git"`
}

// GitSourceRevision is the commit information from a git source for a build
type GitSourceRevision struct {
	// commit is the commit hash identifying a specific commit
	Commit string `json:"commit,omitempty" protobuf:"bytes,1,opt,name=commit"`

	// author is the author of a specific commit
	Author SourceControlUser `json:"author,omitempty" protobuf:"bytes,2,opt,name=author"`

	// committer is the committer of a specific commit
	Committer SourceControlUser `json:"committer,omitempty" protobuf:"bytes,3,opt,name=committer"`

	// message is the description of a specific commit
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}

// ProxyConfig defines what proxies to use for an operation
type ProxyConfig struct {
	// httpProxy is a proxy used to reach the git repository over http
	HTTPProxy *string `json:"httpProxy,omitempty" protobuf:"bytes,3,opt,name=httpProxy"`

	// httpsProxy is a proxy used to reach the git repository over https
	HTTPSProxy *string `json:"httpsProxy,omitempty" protobuf:"bytes,4,opt,name=httpsProxy"`

	// noProxy is the list of domains for which the proxy should not be used
	NoProxy *string `json:"noProxy,omitempty" protobuf:"bytes,5,opt,name=noProxy"`
}

// GitBuildSource defines the parameters of a Git SCM
type GitBuildSource struct {
	// uri points to the source that will be built. The structure of the source
	// will depend on the type of build to run
	URI string `json:"uri" protobuf:"bytes,1,opt,name=uri"`

	// ref is the branch/tag/ref to build.
	Ref string `json:"ref,omitempty" protobuf:"bytes,2,opt,name=ref"`

	// proxyConfig defines the proxies to use for the git clone operation
	ProxyConfig `json:",inline" protobuf:"bytes,3,opt,name=proxyConfig"`
}

// SourceControlUser defines the identity of a user of source control
type SourceControlUser struct {
	// name of the source control user
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// email of the source control user
	Email string `json:"email,omitempty" protobuf:"bytes,2,opt,name=email"`
}

// BuildStrategy contains the details of how to perform a build.
type BuildStrategy struct {
	// type is the kind of build strategy.
	// +k8s:conversion-gen=false
	Type BuildStrategyType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=BuildStrategyType"`

	// dockerStrategy holds the parameters to the Docker build strategy.
	DockerStrategy *DockerBuildStrategy `json:"dockerStrategy,omitempty" protobuf:"bytes,2,opt,name=dockerStrategy"`

	// sourceStrategy holds the parameters to the Source build strategy.
	SourceStrategy *SourceBuildStrategy `json:"sourceStrategy,omitempty" protobuf:"bytes,3,opt,name=sourceStrategy"`

	// customStrategy holds the parameters to the Custom build strategy
	CustomStrategy *CustomBuildStrategy `json:"customStrategy,omitempty" protobuf:"bytes,4,opt,name=customStrategy"`

	// JenkinsPipelineStrategy holds the parameters to the Jenkins Pipeline build strategy.
	// This strategy is in tech preview.
	JenkinsPipelineStrategy *JenkinsPipelineBuildStrategy `json:"jenkinsPipelineStrategy,omitempty" protobuf:"bytes,5,opt,name=jenkinsPipelineStrategy"`
}

// BuildStrategyType describes a particular way of performing a build.
type BuildStrategyType string

// Valid values for BuildStrategyType.
const (
	// DockerBuildStrategyType performs builds using a Dockerfile.
	DockerBuildStrategyType BuildStrategyType = "Docker"

	// SourceBuildStrategyType performs builds build using Source To Images with a Git repository
	// and a builder image.
	SourceBuildStrategyType BuildStrategyType = "Source"

	// CustomBuildStrategyType performs builds using custom builder Docker image.
	CustomBuildStrategyType BuildStrategyType = "Custom"

	// JenkinsPipelineBuildStrategyType indicates the build will run via Jenkine Pipeline.
	JenkinsPipelineBuildStrategyType BuildStrategyType = "JenkinsPipeline"
)

// CustomBuildStrategy defines input parameters specific to Custom build.
type CustomBuildStrategy struct {
	// from is reference to an DockerImage, ImageStreamTag, or ImageStreamImage from which
	// the docker image should be pulled
	From kapi.ObjectReference `json:"from" protobuf:"bytes,1,opt,name=from"`

	// pullSecret is the name of a Secret that would be used for setting up
	// the authentication for pulling the Docker images from the private Docker
	// registries
	PullSecret *kapi.LocalObjectReference `json:"pullSecret,omitempty" protobuf:"bytes,2,opt,name=pullSecret"`

	// env contains additional environment variables you want to pass into a builder container
	Env []kapi.EnvVar `json:"env,omitempty" protobuf:"bytes,3,rep,name=env"`

	// exposeDockerSocket will allow running Docker commands (and build Docker images) from
	// inside the Docker container.
	// TODO: Allow admins to enforce 'false' for this option
	ExposeDockerSocket bool `json:"exposeDockerSocket,omitempty" protobuf:"varint,4,opt,name=exposeDockerSocket"`

	// forcePull describes if the controller should configure the build pod to always pull the images
	// for the builder or only pull if it is not present locally
	ForcePull bool `json:"forcePull,omitempty" protobuf:"varint,5,opt,name=forcePull"`

	// secrets is a list of additional secrets that will be included in the build pod
	Secrets []SecretSpec `json:"secrets,omitempty" protobuf:"bytes,6,rep,name=secrets"`

	// buildAPIVersion is the requested API version for the Build object serialized and passed to the custom builder
	BuildAPIVersion string `json:"buildAPIVersion,omitempty" protobuf:"bytes,7,opt,name=buildAPIVersion"`
}

// DockerBuildStrategy defines input parameters specific to Docker build.
type DockerBuildStrategy struct {
	// from is reference to an DockerImage, ImageStreamTag, or ImageStreamImage from which
	// the docker image should be pulled
	// the resulting image will be used in the FROM line of the Dockerfile for this build.
	From *kapi.ObjectReference `json:"from,omitempty" protobuf:"bytes,1,opt,name=from"`

	// pullSecret is the name of a Secret that would be used for setting up
	// the authentication for pulling the Docker images from the private Docker
	// registries
	PullSecret *kapi.LocalObjectReference `json:"pullSecret,omitempty" protobuf:"bytes,2,opt,name=pullSecret"`

	// noCache if set to true indicates that the docker build must be executed with the
	// --no-cache=true flag
	NoCache bool `json:"noCache,omitempty" protobuf:"varint,3,opt,name=noCache"`

	// env contains additional environment variables you want to pass into a builder container
	Env []kapi.EnvVar `json:"env,omitempty" protobuf:"bytes,4,rep,name=env"`

	// forcePull describes if the builder should pull the images from registry prior to building.
	ForcePull bool `json:"forcePull,omitempty" protobuf:"varint,5,opt,name=forcePull"`

	// dockerfilePath is the path of the Dockerfile that will be used to build the Docker image,
	// relative to the root of the context (contextDir).
	DockerfilePath string `json:"dockerfilePath,omitempty" protobuf:"bytes,6,opt,name=dockerfilePath"`
}

// SourceBuildStrategy defines input parameters specific to an Source build.
type SourceBuildStrategy struct {
	// from is reference to an DockerImage, ImageStreamTag, or ImageStreamImage from which
	// the docker image should be pulled
	From kapi.ObjectReference `json:"from" protobuf:"bytes,1,opt,name=from"`

	// pullSecret is the name of a Secret that would be used for setting up
	// the authentication for pulling the Docker images from the private Docker
	// registries
	PullSecret *kapi.LocalObjectReference `json:"pullSecret,omitempty" protobuf:"bytes,2,opt,name=pullSecret"`

	// env contains additional environment variables you want to pass into a builder container
	Env []kapi.EnvVar `json:"env,omitempty" protobuf:"bytes,3,rep,name=env"`

	// scripts is the location of Source scripts
	Scripts string `json:"scripts,omitempty" protobuf:"bytes,4,opt,name=scripts"`

	// incremental flag forces the Source build to do incremental builds if true.
	Incremental *bool `json:"incremental,omitempty" protobuf:"varint,5,opt,name=incremental"`

	// forcePull describes if the builder should pull the images from registry prior to building.
	ForcePull bool `json:"forcePull,omitempty" protobuf:"varint,6,opt,name=forcePull"`

	// runtimeImage is an optional image that is used to run an application
	// without unneeded dependencies installed. The building of the application
	// is still done in the builder image but, post build, you can copy the
	// needed artifacts in the runtime image for use.
	// This field and the feature it enables are in tech preview.
	RuntimeImage *kapi.ObjectReference `json:"runtimeImage,omitempty" protobuf:"bytes,7,opt,name=runtimeImage"`

	// runtimeArtifacts specifies a list of source/destination pairs that will be
	// copied from the builder to the runtime image. sourcePath can be a file or
	// directory. destinationDir must be a directory. destinationDir can also be
	// empty or equal to ".", in this case it just refers to the root of WORKDIR.
	// This field and the feature it enables are in tech preview.
	RuntimeArtifacts []ImageSourcePath `json:"runtimeArtifacts,omitempty" protobuf:"bytes,8,rep,name=runtimeArtifacts"`
}

// JenkinsPipelineBuildStrategy holds parameters specific to a Jenkins Pipeline build.
// This strategy is in tech preview.
type JenkinsPipelineBuildStrategy struct {
	// JenkinsfilePath is the optional path of the Jenkinsfile that will be used to configure the pipeline
	// relative to the root of the context (contextDir). If both JenkinsfilePath & Jenkinsfile are
	// both not specified, this defaults to Jenkinsfile in the root of the specified contextDir.
	JenkinsfilePath string `json:"jenkinsfilePath,omitempty" protobuf:"bytes,1,opt,name=jenkinsfilePath"`

	// Jenkinsfile defines the optional raw contents of a Jenkinsfile which defines a Jenkins pipeline build.
	Jenkinsfile string `json:"jenkinsfile,omitempty" protobuf:"bytes,2,opt,name=jenkinsfile"`
}

// A BuildPostCommitSpec holds a build post commit hook specification. The hook
// executes a command in a temporary container running the build output image,
// immediately after the last layer of the image is committed and before the
// image is pushed to a registry. The command is executed with the current
// working directory ($PWD) set to the image's WORKDIR.
//
// The build will be marked as failed if the hook execution fails. It will fail
// if the script or command return a non-zero exit code, or if there is any
// other error related to starting the temporary container.
//
// There are five different ways to configure the hook. As an example, all forms
// below are equivalent and will execute `rake test --verbose`.
//
// 1. Shell script:
//
//        "postCommit": {
//          "script": "rake test --verbose",
//        }
//
//     The above is a convenient form which is equivalent to:
//
//        "postCommit": {
//          "command": ["/bin/sh", "-ic"],
//          "args":    ["rake test --verbose"]
//        }
//
// 2. A command as the image entrypoint:
//
//        "postCommit": {
//          "commit": ["rake", "test", "--verbose"]
//        }
//
//     Command overrides the image entrypoint in the exec form, as documented in
//     Docker: https://docs.docker.com/engine/reference/builder/#entrypoint.
//
// 3. Pass arguments to the default entrypoint:
//
//        "postCommit": {
// 		      "args": ["rake", "test", "--verbose"]
// 	      }
//
//     This form is only useful if the image entrypoint can handle arguments.
//
// 4. Shell script with arguments:
//
//        "postCommit": {
//          "script": "rake test $1",
//          "args":   ["--verbose"]
//        }
//
//     This form is useful if you need to pass arguments that would otherwise be
//     hard to quote properly in the shell script. In the script, $0 will be
//     "/bin/sh" and $1, $2, etc, are the positional arguments from Args.
//
// 5. Command with arguments:
//
//        "postCommit": {
//          "command": ["rake", "test"],
//          "args":    ["--verbose"]
//        }
//
//     This form is equivalent to appending the arguments to the Command slice.
//
// It is invalid to provide both Script and Command simultaneously. If none of
// the fields are specified, the hook is not executed.
type BuildPostCommitSpec struct {
	// command is the command to run. It may not be specified with Script.
	// This might be needed if the image doesn't have `/bin/sh`, or if you
	// do not want to use a shell. In all other cases, using Script might be
	// more convenient.
	Command []string `json:"command,omitempty" protobuf:"bytes,1,rep,name=command"`
	// args is a list of arguments that are provided to either Command,
	// Script or the Docker image's default entrypoint. The arguments are
	// placed immediately after the command to be run.
	Args []string `json:"args,omitempty" protobuf:"bytes,2,rep,name=args"`
	// script is a shell script to be run with `/bin/sh -ic`. It may not be
	// specified with Command. Use Script when a shell script is appropriate
	// to execute the post build hook, for example for running unit tests
	// with `rake test`. If you need control over the image entrypoint, or
	// if the image does not have `/bin/sh`, use Command and/or Args.
	// The `-i` flag is needed to support CentOS and RHEL images that use
	// Software Collections (SCL), in order to have the appropriate
	// collections enabled in the shell. E.g., in the Ruby image, this is
	// necessary to make `ruby`, `bundle` and other binaries available in
	// the PATH.
	Script string `json:"script,omitempty" protobuf:"bytes,3,opt,name=script"`
}

// BuildOutput is input to a build strategy and describes the Docker image that the strategy
// should produce.
type BuildOutput struct {
	// to defines an optional location to push the output of this build to.
	// Kind must be one of 'ImageStreamTag' or 'DockerImage'.
	// This value will be used to look up a Docker image repository to push to.
	// In the case of an ImageStreamTag, the ImageStreamTag will be looked for in the namespace of
	// the build unless Namespace is specified.
	To *kapi.ObjectReference `json:"to,omitempty" protobuf:"bytes,1,opt,name=to"`

	// PushSecret is the name of a Secret that would be used for setting
	// up the authentication for executing the Docker push to authentication
	// enabled Docker Registry (or Docker Hub).
	PushSecret *kapi.LocalObjectReference `json:"pushSecret,omitempty" protobuf:"bytes,2,opt,name=pushSecret"`

	// imageLabels define a list of labels that are applied to the resulting image. If there
	// are multiple labels with the same name then the last one in the list is used.
	ImageLabels []ImageLabel `json:"imageLabels,omitempty" protobuf:"bytes,3,rep,name=imageLabels"`
}

// ImageLabel represents a label applied to the resulting image.
type ImageLabel struct {
	// name defines the name of the label. It must have non-zero length.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// value defines the literal value of the label.
	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

// Build configurations define a build process for new Docker images. There are three types of builds possible - a Docker build using a Dockerfile, a Source-to-Image build that uses a specially prepared base image that accepts source code that it can make runnable, and a custom build that can run // arbitrary Docker images as a base and accept the build parameters. Builds run on the cluster and on completion are pushed to the Docker registry specified in the "output" section. A build can be triggered via a webhook, when the base image changes, or when a user manually requests a new build be // created.
//
// Each build created by a build configuration is numbered and refers back to its parent configuration. Multiple builds can be triggered at once. Builds that do not have "output" set can be used to test code or run a verification build.
type BuildConfig struct {
	unversioned.TypeMeta `json:",inline"`
	// metadata for BuildConfig.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec holds all the input necessary to produce a new build, and the conditions when
	// to trigger them.
	Spec BuildConfigSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// status holds any relevant information about a build config
	Status BuildConfigStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// BuildConfigSpec describes when and how builds are created
type BuildConfigSpec struct {

	//triggers determine how new Builds can be launched from a BuildConfig. If
	//no triggers are defined, a new build can only occur as a result of an
	//explicit client build creation.
	Triggers []BuildTriggerPolicy `json:"triggers" protobuf:"bytes,1,rep,name=triggers"`

	// RunPolicy describes how the new build created from this build
	// configuration will be scheduled for execution.
	// This is optional, if not specified we default to "Serial".
	RunPolicy BuildRunPolicy `json:"runPolicy,omitempty" protobuf:"bytes,2,opt,name=runPolicy,casttype=BuildRunPolicy"`

	// CommonSpec is the desired build specification
	CommonSpec `json:",inline" protobuf:"bytes,3,opt,name=commonSpec"`
}

// BuildRunPolicy defines the behaviour of how the new builds are executed
// from the existing build configuration.
type BuildRunPolicy string

const (
	// BuildRunPolicyParallel schedules new builds immediately after they are
	// created. Builds will be executed in parallel.
	BuildRunPolicyParallel BuildRunPolicy = "Parallel"

	// BuildRunPolicySerial schedules new builds to execute in a sequence as
	// they are created. Every build gets queued up and will execute when the
	// previous build completes. This is the default policy.
	BuildRunPolicySerial BuildRunPolicy = "Serial"

	// BuildRunPolicySerialLatestOnly schedules only the latest build to execute,
	// cancelling all the previously queued build.
	BuildRunPolicySerialLatestOnly BuildRunPolicy = "SerialLatestOnly"
)

// BuildConfigStatus contains current state of the build config object.
type BuildConfigStatus struct {
	// lastVersion is used to inform about number of last triggered build.
	LastVersion int64 `json:"lastVersion" protobuf:"varint,1,opt,name=lastVersion"`
}

// WebHookTrigger is a trigger that gets invoked using a webhook type of post
type WebHookTrigger struct {
	// secret used to validate requests.
	Secret string `json:"secret,omitempty" protobuf:"bytes,1,opt,name=secret"`

	// allowEnv determines whether the webhook can set environment variables; can only
	// be set to true for GenericWebHook.
	AllowEnv bool `json:"allowEnv,omitempty" protobuf:"varint,2,opt,name=allowEnv"`
}

// ImageChangeTrigger allows builds to be triggered when an ImageStream changes
type ImageChangeTrigger struct {
	// lastTriggeredImageID is used internally by the ImageChangeController to save last
	// used image ID for build
	LastTriggeredImageID string `json:"lastTriggeredImageID,omitempty" protobuf:"bytes,1,opt,name=lastTriggeredImageID"`

	// from is a reference to an ImageStreamTag that will trigger a build when updated
	// It is optional. If no From is specified, the From image from the build strategy
	// will be used. Only one ImageChangeTrigger with an empty From reference is allowed in
	// a build configuration.
	From *kapi.ObjectReference `json:"from,omitempty" protobuf:"bytes,2,opt,name=from"`
}

// BuildTriggerPolicy describes a policy for a single trigger that results in a new Build.
type BuildTriggerPolicy struct {
	// type is the type of build trigger
	Type BuildTriggerType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=BuildTriggerType"`

	// github contains the parameters for a GitHub webhook type of trigger
	GitHubWebHook *WebHookTrigger `json:"github,omitempty" protobuf:"bytes,2,opt,name=github"`

	// generic contains the parameters for a Generic webhook type of trigger
	GenericWebHook *WebHookTrigger `json:"generic,omitempty" protobuf:"bytes,3,opt,name=generic"`

	// imageChange contains parameters for an ImageChange type of trigger
	ImageChange *ImageChangeTrigger `json:"imageChange,omitempty" protobuf:"bytes,4,opt,name=imageChange"`
}

// BuildTriggerType refers to a specific BuildTriggerPolicy implementation.
type BuildTriggerType string

const (
	// GitHubWebHookBuildTriggerType represents a trigger that launches builds on
	// GitHub webhook invocations
	GitHubWebHookBuildTriggerType           BuildTriggerType = "GitHub"
	GitHubWebHookBuildTriggerTypeDeprecated BuildTriggerType = "github"

	// GenericWebHookBuildTriggerType represents a trigger that launches builds on
	// generic webhook invocations
	GenericWebHookBuildTriggerType           BuildTriggerType = "Generic"
	GenericWebHookBuildTriggerTypeDeprecated BuildTriggerType = "generic"

	// ImageChangeBuildTriggerType represents a trigger that launches builds on
	// availability of a new version of an image
	ImageChangeBuildTriggerType           BuildTriggerType = "ImageChange"
	ImageChangeBuildTriggerTypeDeprecated BuildTriggerType = "imageChange"

	// ConfigChangeBuildTriggerType will trigger a build on an initial build config creation
	// WARNING: In the future the behavior will change to trigger a build on any config change
	ConfigChangeBuildTriggerType BuildTriggerType = "ConfigChange"
)

// BuildList is a collection of Builds.
type BuildList struct {
	unversioned.TypeMeta `json:",inline"`
	// metadata for BuildList.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is a list of builds
	Items []Build `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// BuildConfigList is a collection of BuildConfigs.
type BuildConfigList struct {
	unversioned.TypeMeta `json:",inline"`
	// metadata for BuildConfigList.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is a list of build configs
	Items []BuildConfig `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GenericWebHookEvent is the payload expected for a generic webhook post
type GenericWebHookEvent struct {
	// type is the type of source repository
	// +k8s:conversion-gen=false
	Type BuildSourceType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=BuildSourceType"`

	// git is the git information if the Type is BuildSourceGit
	Git *GitInfo `json:"git,omitempty" protobuf:"bytes,2,opt,name=git"`

	// env contains additional environment variables you want to pass into a builder container
	Env []kapi.EnvVar `json:"env,omitempty" protobuf:"bytes,3,rep,name=env"`
}

// GitInfo is the aggregated git information for a generic webhook post
type GitInfo struct {
	GitBuildSource    `json:",inline" protobuf:"bytes,1,opt,name=gitBuildSource"`
	GitSourceRevision `json:",inline" protobuf:"bytes,2,opt,name=gitSourceRevision"`
}

// BuildLog is the (unused) resource associated with the build log redirector
type BuildLog struct {
	unversioned.TypeMeta `json:",inline"`
}

// BuildRequest is the resource used to pass parameters to build generator
type BuildRequest struct {
	unversioned.TypeMeta `json:",inline"`
	// metadata for BuildRequest.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// revision is the information from the source for a specific repo snapshot.
	Revision *SourceRevision `json:"revision,omitempty" protobuf:"bytes,2,opt,name=revision"`

	// triggeredByImage is the Image that triggered this build.
	TriggeredByImage *kapi.ObjectReference `json:"triggeredByImage,omitempty" protobuf:"bytes,3,opt,name=triggeredByImage"`

	// from is the reference to the ImageStreamTag that triggered the build.
	From *kapi.ObjectReference `json:"from,omitempty" protobuf:"bytes,4,opt,name=from"`

	// binary indicates a request to build from a binary provided to the builder
	Binary *BinaryBuildSource `json:"binary,omitempty" protobuf:"bytes,5,opt,name=binary"`

	// lastVersion (optional) is the LastVersion of the BuildConfig that was used
	// to generate the build. If the BuildConfig in the generator doesn't match, a build will
	// not be generated.
	LastVersion *int64 `json:"lastVersion,omitempty" protobuf:"varint,6,opt,name=lastVersion"`

	// env contains additional environment variables you want to pass into a builder container
	Env []kapi.EnvVar `json:"env,omitempty" protobuf:"bytes,7,rep,name=env"`

	// triggeredBy describes which triggers started the most recent update to the
	// build configuration and contains information about those triggers.
	TriggeredBy []BuildTriggerCause `json:"triggeredBy" protobuf:"bytes,8,rep,name=triggeredBy"`
}

// BinaryBuildRequestOptions are the options required to fully speficy a binary build request
type BinaryBuildRequestOptions struct {
	unversioned.TypeMeta `json:",inline"`
	// metadata for BinaryBuildRequestOptions.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// asFile determines if the binary should be created as a file within the source rather than extracted as an archive
	AsFile string `json:"asFile,omitempty" protobuf:"bytes,2,opt,name=asFile"`

	// TODO: Improve map[string][]string conversion so we can handled nested objects

	// revision.commit is the value identifying a specific commit
	Commit string `json:"revision.commit,omitempty" protobuf:"bytes,3,opt,name=revisionCommit"`

	// revision.message is the description of a specific commit
	Message string `json:"revision.message,omitempty" protobuf:"bytes,4,opt,name=revisionMessage"`

	// revision.authorName of the source control user
	AuthorName string `json:"revision.authorName,omitempty" protobuf:"bytes,5,opt,name=revisionAuthorName"`

	// revision.authorEmail of the source control user
	AuthorEmail string `json:"revision.authorEmail,omitempty" protobuf:"bytes,6,opt,name=revisionAuthorEmail"`

	// revision.committerName of the source control user
	CommitterName string `json:"revision.committerName,omitempty" protobuf:"bytes,7,opt,name=revisionCommitterName"`

	// revision.committerEmail of the source control user
	CommitterEmail string `json:"revision.committerEmail,omitempty" protobuf:"bytes,8,opt,name=revisionCommitterEmail"`
}

// BuildLogOptions is the REST options for a build log
type BuildLogOptions struct {
	unversioned.TypeMeta `json:",inline"`

	// cointainer for which to stream logs. Defaults to only container if there is one container in the pod.
	Container string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	// follow if true indicates that the build log should be streamed until
	// the build terminates.
	Follow bool `json:"follow,omitempty" protobuf:"varint,2,opt,name=follow"`
	// previous returns previous build logs. Defaults to false.
	Previous bool `json:"previous,omitempty" protobuf:"varint,3,opt,name=previous"`
	// sinceSeconds is a relative time in seconds before the current time from which to show logs. If this value
	// precedes the time a pod was started, only logs since the pod start will be returned.
	// If this value is in the future, no logs will be returned.
	// Only one of sinceSeconds or sinceTime may be specified.
	SinceSeconds *int64 `json:"sinceSeconds,omitempty" protobuf:"varint,4,opt,name=sinceSeconds"`
	// sinceTime is an RFC3339 timestamp from which to show logs. If this value
	// precedes the time a pod was started, only logs since the pod start will be returned.
	// If this value is in the future, no logs will be returned.
	// Only one of sinceSeconds or sinceTime may be specified.
	SinceTime *unversioned.Time `json:"sinceTime,omitempty" protobuf:"bytes,5,opt,name=sinceTime"`
	// timestamps, If true, add an RFC3339 or RFC3339Nano timestamp at the beginning of every line
	// of log output. Defaults to false.
	Timestamps bool `json:"timestamps,omitempty" protobuf:"varint,6,opt,name=timestamps"`
	// tailLines, If set, is the number of lines from the end of the logs to show. If not specified,
	// logs are shown from the creation of the container or sinceSeconds or sinceTime
	TailLines *int64 `json:"tailLines,omitempty" protobuf:"varint,7,opt,name=tailLines"`
	// limitBytes, If set, is the number of bytes to read from the server before terminating the
	// log output. This may not display a complete final line of logging, and may return
	// slightly more or slightly less than the specified limit.
	LimitBytes *int64 `json:"limitBytes,omitempty" protobuf:"varint,8,opt,name=limitBytes"`

	// noWait if true causes the call to return immediately even if the build
	// is not available yet. Otherwise the server will wait until the build has started.
	// TODO: Fix the tag to 'noWait' in v2
	NoWait bool `json:"nowait,omitempty" protobuf:"varint,9,opt,name=nowait"`

	// version of the build for which to view logs.
	Version *int64 `json:"version,omitempty" protobuf:"varint,10,opt,name=version"`
}

// SecretSpec specifies a secret to be included in a build pod and its corresponding mount point
type SecretSpec struct {
	// secretSource is a reference to the secret
	SecretSource kapi.LocalObjectReference `json:"secretSource" protobuf:"bytes,1,opt,name=secretSource"`

	// mountPath is the path at which to mount the secret
	MountPath string `json:"mountPath" protobuf:"bytes,2,opt,name=mountPath"`
}
