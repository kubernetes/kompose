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
	ConvertOpt                   kobject.ConvertOptions
	ConvertYAMLIndent            int
)

var convertCmd = &cobra.Command{
	Use:   "convert [file]",
	Short: "Convert a Docker Compose file",
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
			CreateDeploymentConfig:      ConvertDeploymentConfig,
			EmptyVols:                   ConvertEmptyVols,
			Volumes:                     ConvertVolumes,
			InsecureRepository:          ConvertInsecureRepo,
			IsDeploymentFlag:            cmd.Flags().Lookup("deployment").Changed,
			IsDaemonSetFlag:             cmd.Flags().Lookup("daemon-set").Changed,
			IsReplicationControllerFlag: cmd.Flags().Lookup("replication-controller").Changed,
			Controller:                  strings.ToLower(ConvertController),
			IsReplicaSetFlag:            cmd.Flags().Lookup("replicas").Changed,
			IsDeploymentConfigFlag:      cmd.Flags().Lookup("deployment-config").Changed,
			YAMLIndent:                  ConvertYAMLIndent,
		}

		// Validate before doing anything else. Use "bundle" if passed in.
		app.ValidateFlags(GlobalBundle, args, cmd, &ConvertOpt)
		app.ValidateComposeFile(&ConvertOpt)
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
	convertCmd.Flags().BoolVarP(&ConvertYaml, "yaml", "y", false, "Generate resource files into YAML format")
	convertCmd.Flags().MarkDeprecated("yaml", "YAML is the default format now.")
	convertCmd.Flags().MarkShorthandDeprecated("y", "YAML is the default format now.")
	convertCmd.Flags().BoolVarP(&ConvertJSON, "json", "j", false, "Generate resource files into JSON format")
	convertCmd.Flags().BoolVar(&ConvertStdout, "stdout", false, "Print converted objects to stdout")
	convertCmd.Flags().StringVarP(&ConvertOut, "out", "o", "", "Specify a file name or directory to save objects to (if path does not exist, a file will be created)")
	convertCmd.Flags().IntVar(&ConvertReplicas, "replicas", 1, "Specify the number of replicas in the generated resource spec")
	convertCmd.Flags().StringVar(&ConvertVolumes, "volumes", "persistentVolumeClaim", `Volumes to be generated ("persistentVolumeClaim"|"emptyDir"|"hostPath" | "configMap")`)

	// Deprecated commands
	convertCmd.Flags().BoolVar(&ConvertEmptyVols, "emptyvols", false, "Use Empty Volumes. Do not generate PVCs")
	convertCmd.Flags().MarkDeprecated("emptyvols", "emptyvols has been marked as deprecated. Use --volumes empty")

	convertCmd.Flags().IntVar(&ConvertYAMLIndent, "indent", 2, "Spaces length to indent generated yaml files")

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
      --daemon-set               Generate a Kubernetes daemonset object (deprecated, use --controller instead)
  -d, --deployment               Generate a Kubernetes deployment object (deprecated, use --controller instead)
  -c, --chart                    Create a Helm chart for converted objects
      --replication-controller   Generate a Kubernetes replication controller object (deprecated, use --controller instead)

OpenShift Flags:
      --build-branch             Specify repository branch to use for buildconfig (default is current branch name)
      --build-repo               Specify source repository for buildconfig (default is current branch's remote url)
      --deployment-config        Generate an OpenShift deployment config object
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
