/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"strings"

	"github.com/kubernetes/kompose/pkg/app"
	"github.com/kubernetes/kompose/pkg/kobject"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: comment
var (
	ConvertOut                   string
	ConvertBuildRepo             string
	ConvertBuildBranch           string
	ConvertBuild                 string
	ConvertVolumes               string
	ConvertPVCRequestSize        string
	ConvertChart                 bool
	ConvertDeployment            bool
	ConvertDaemonSet             bool
	ConvertReplicationController bool
	ConvertYaml                  bool
	ConvertJSON                  bool
	ConvertStdout                bool
	ConvertEmptyVols             bool
	ConvertInsecureRepo          bool
	ConvertDeploymentConfig      bool
	ConvertReplicas              int
	ConvertController            string
	ConvertProfiles              []string
	ConvertPushImage             bool
	ConvertNamespace             string
	ConvertPushImageRegistry     string
	ConvertOpt                   kobject.ConvertOptions
	ConvertYAMLIndent            int
	GenerateNetworkPolicies      bool

	UpBuild string

	BuildCommand string
	PushCommand  string
	// WithKomposeAnnotation decides if we will add metadata about this convert to resource's annotation.
	// default is true.
	WithKomposeAnnotation bool

	// MultipleContainerMode which enables creating multi containers in a single pod is a developing function.
	// default is false
	MultipleContainerMode bool

	ServiceGroupMode string
	ServiceGroupName string

	// SecretsAsFiles forces secrets to result in files inside a container instead of symlinked directories containing
	// files of the same name. This reproduces the behavior of file-based secrets in docker-compose and should probably
	// be the default for kompose, but we must keep compatibility with the previous behavior.
	// See https://github.com/kubernetes/kompose/issues/1280 for more details.
	SecretsAsFiles bool
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a Compose file",
	Example: `  kompose --file compose.yaml convert
  kompose -f first.yaml -f second.yaml convert
  kompose --provider openshift --file compose.yaml convert`,
	PreRun: func(cmd *cobra.Command, args []string) {

		// Check that build-config wasn't passed in with --provider=kubernetes
		if GlobalProvider == "kubernetes" && UpBuild == "build-config" {
			log.Fatalf("build-config is not a valid --build parameter with provider Kubernetes")
		}

		// Create the Convert Options.
		ConvertOpt = kobject.ConvertOptions{
			ToStdout:                    ConvertStdout,
			CreateChart:                 ConvertChart,
			GenerateYaml:                ConvertYaml,
			GenerateJSON:                ConvertJSON,
			Replicas:                    ConvertReplicas,
			InputFiles:                  GlobalFiles,
			OutFile:                     ConvertOut,
			Provider:                    GlobalProvider,
			CreateD:                     ConvertDeployment,
			CreateDS:                    ConvertDaemonSet,
			CreateRC:                    ConvertReplicationController,
			Build:                       ConvertBuild,
			BuildRepo:                   ConvertBuildRepo,
			BuildBranch:                 ConvertBuildBranch,
			PushImage:                   ConvertPushImage,
			PushImageRegistry:           ConvertPushImageRegistry,
			CreateDeploymentConfig:      ConvertDeploymentConfig,
			EmptyVols:                   ConvertEmptyVols,
			Volumes:                     ConvertVolumes,
			PVCRequestSize:              ConvertPVCRequestSize,
			InsecureRepository:          ConvertInsecureRepo,
			IsDeploymentFlag:            cmd.Flags().Lookup("deployment").Changed,
			IsDaemonSetFlag:             cmd.Flags().Lookup("daemon-set").Changed,
			IsReplicationControllerFlag: cmd.Flags().Lookup("replication-controller").Changed,
			Controller:                  strings.ToLower(ConvertController),
			IsReplicaSetFlag:            cmd.Flags().Lookup("replicas").Changed,
			IsDeploymentConfigFlag:      cmd.Flags().Lookup("deployment-config").Changed,
			YAMLIndent:                  ConvertYAMLIndent,
			Profiles:                    ConvertProfiles,
			WithKomposeAnnotation:       WithKomposeAnnotation,
			MultipleContainerMode:       MultipleContainerMode,
			ServiceGroupMode:            ServiceGroupMode,
			ServiceGroupName:            ServiceGroupName,
			SecretsAsFiles:              SecretsAsFiles,
			GenerateNetworkPolicies:     GenerateNetworkPolicies,
			BuildCommand:                BuildCommand,
			PushCommand:                 PushCommand,
			Namespace:                   ConvertNamespace,
		}

		if ServiceGroupMode == "" && MultipleContainerMode {
			ConvertOpt.ServiceGroupMode = "label"
		}

		app.ValidateFlags(args, cmd, &ConvertOpt)

		// Since ValidateComposeFiles returns an error, let's validate it and output the error appropriately if the validation fails
		err := app.ValidateComposeFile(&ConvertOpt)
		if err != nil {
			log.Fatalf("Error validating compose file: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		app.Convert(ConvertOpt)
	},
}

func init() {
	// Automatically grab environment variables
	viper.AutomaticEnv()

	// Kubernetes only
	convertCmd.Flags().BoolVarP(&ConvertChart, "chart", "c", false, "Create a Helm chart for converted objects")
	convertCmd.Flags().BoolVar(&ConvertDaemonSet, "daemon-set", false, "Generate a Kubernetes daemonset object (deprecated, use --controller instead)")
	convertCmd.Flags().BoolVarP(&ConvertDeployment, "deployment", "d", false, "Generate a Kubernetes deployment object (deprecated, use --controller instead)")
	convertCmd.Flags().BoolVar(&ConvertReplicationController, "replication-controller", false, "Generate a Kubernetes replication controller object (deprecated, use --controller instead)")
	convertCmd.Flags().StringVar(&ConvertController, "controller", "", `Set the output controller ("deployment"|"daemonSet"|"replicationController")`)
	convertCmd.Flags().MarkDeprecated("daemon-set", "use --controller")
	convertCmd.Flags().MarkDeprecated("deployment", "use --controller")
	convertCmd.Flags().MarkDeprecated("replication-controller", "use --controller")
	convertCmd.Flags().MarkHidden("chart")
	convertCmd.Flags().MarkHidden("daemon-set")
	convertCmd.Flags().MarkHidden("replication-controller")
	convertCmd.Flags().MarkHidden("deployment")
	convertCmd.Flags().BoolVar(&MultipleContainerMode, "multiple-container-mode", false, "Create multiple containers grouped by 'kompose.service.group' label")
	convertCmd.Flags().StringVar(&ServiceGroupMode, "service-group-mode", "", "Group multiple service to create single workload by `label`(`kompose.service.group`) or `volume`(shared volumes)")
	convertCmd.Flags().StringVar(&ServiceGroupName, "service-group-name", "", "Using with --service-group-mode=volume to specific a final service name for the group")
	convertCmd.Flags().MarkDeprecated("multiple-container-mode", "use --service-group-mode=label")
	convertCmd.Flags().BoolVar(&SecretsAsFiles, "secrets-as-files", false, "Always convert docker-compose secrets into files instead of symlinked directories")

	// OpenShift only
	convertCmd.Flags().BoolVar(&ConvertDeploymentConfig, "deployment-config", true, "Generate an OpenShift deploymentconfig object")
	convertCmd.Flags().BoolVar(&ConvertInsecureRepo, "insecure-repository", false, "Use an insecure Docker repository for OpenShift ImageStream")
	convertCmd.Flags().StringVar(&ConvertBuildRepo, "build-repo", "", "Specify source repository for buildconfig (default remote origin)")
	convertCmd.Flags().StringVar(&ConvertBuildBranch, "build-branch", "", "Specify repository branch to use for buildconfig (default master)")
	convertCmd.Flags().MarkDeprecated("deployment-config", "use --controller")
	convertCmd.Flags().MarkHidden("deployment-config")
	convertCmd.Flags().MarkHidden("insecure-repository")
	convertCmd.Flags().MarkHidden("build-repo")
	convertCmd.Flags().MarkHidden("build-branch")

	// Standard between the two
	convertCmd.Flags().StringVar(&ConvertBuild, "build", "none", `Set the type of build ("local"|"build-config"(OpenShift only)|"none")`)
	convertCmd.Flags().BoolVar(&ConvertPushImage, "push-image", false, "If we should push the docker image we built")
	convertCmd.Flags().StringVar(&BuildCommand, "build-command", "", `Set the command used to build the container image, which will override the docker build command. Should be used in conjuction with --push-command flag.`)
	convertCmd.Flags().StringVar(&PushCommand, "push-command", "", `Set the command used to push the container image. override the docker push command. Should be used in conjuction with --build-command flag.`)
	convertCmd.Flags().StringVar(&ConvertPushImageRegistry, "push-image-registry", "", "Specify registry for pushing image, which will override registry from image name")
	convertCmd.Flags().BoolVarP(&ConvertYaml, "yaml", "y", false, "Generate resource files into YAML format")
	convertCmd.Flags().MarkDeprecated("yaml", "YAML is the default format now")
	convertCmd.Flags().MarkShorthandDeprecated("y", "YAML is the default format now")
	convertCmd.Flags().BoolVarP(&ConvertJSON, "json", "j", false, "Generate resource files into JSON format")
	convertCmd.Flags().BoolVar(&ConvertStdout, "stdout", false, "Print converted objects to stdout")
	convertCmd.Flags().StringVarP(&ConvertOut, "out", "o", "", "Specify a file name or directory to save objects to (if path does not exist, a file will be created)")
	convertCmd.Flags().IntVar(&ConvertReplicas, "replicas", 1, "Specify the number of replicas in the generated resource spec")
	convertCmd.Flags().StringVar(&ConvertVolumes, "volumes", "persistentVolumeClaim", `Volumes to be generated ("persistentVolumeClaim"|"emptyDir"|"hostPath" | "configMap")`)
	convertCmd.Flags().StringVar(&ConvertPVCRequestSize, "pvc-request-size", "", `Specify the size of pvc storage requests in the generated resource spec`)
	convertCmd.Flags().StringVarP(&ConvertNamespace, "namespace", "n", "", `Specify the namespace of the generated resources`)
	convertCmd.Flags().BoolVar(&GenerateNetworkPolicies, "generate-network-policies", false, "Specify whether to generate network policies or not")

	convertCmd.Flags().BoolVar(&WithKomposeAnnotation, "with-kompose-annotation", true, "Add kompose annotations to generated resource")

	// Deprecated commands
	convertCmd.Flags().BoolVar(&ConvertEmptyVols, "emptyvols", false, "Use Empty Volumes. Do not generate PVCs")
	convertCmd.Flags().MarkDeprecated("emptyvols", "emptyvols has been marked as deprecated. Use --volumes emptyDir")

	convertCmd.Flags().IntVar(&ConvertYAMLIndent, "indent", 2, "Spaces length to indent generated yaml files")

	convertCmd.Flags().StringArrayVar(&ConvertProfiles, "profile", []string{}, `Specify the profile to use, can use multiple profiles`)

	// In order to 'separate' both OpenShift and Kubernetes only flags. A custom help page is created
	customHelp := `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Kubernetes Flags:
  -c, --chart                    Create a Helm chart for converted objects
      --controller               Set the output controller ("deployment"|"daemonSet"|"replicationController")
      --service-group-mode       Group multiple service to create single workload by "label"("kompose.service.group") or "volume"(shared volumes)
      --service-group-name       Using with --service-group-mode=volume to specific a final service name for the group

OpenShift Flags:
      --build-branch             Specify repository branch to use for buildconfig (default is current branch name)
      --build-repo               Specify source repository for buildconfig (default is current branch's remote url)
      --insecure-repository      Specify to use insecure docker repository while generating Openshift image stream object

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	// Set the help template + add the command to root
	convertCmd.SetUsageTemplate(customHelp)

	RootCmd.AddCommand(convertCmd)
}
