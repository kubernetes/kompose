/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

	"github.com/kubernetes-incubator/kompose/pkg/app"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ConvertSource, ConvertOut, ConvertBuildRepo, ConvertBuildBranch string
	ConvertChart, ConvertDeployment, ConvertDaemonSet               bool
	ConvertReplicationController, ConvertYaml, ConvertStdout        bool
	ConvertEmptyVols, ConvertDeploymentConfig, ConvertBuildConfig   bool
	ConvertReplicas                                                 int
	ConvertOpt                                                      kobject.ConvertOptions
)

var ConvertProvider string = GlobalProvider

var convertCmd = &cobra.Command{
	Use:   "convert [file]",
	Short: "Convert a Docker Compose file",
	PreRun: func(cmd *cobra.Command, args []string) {

		// Create the Convert Options.
		ConvertOpt = kobject.ConvertOptions{
			ToStdout:               ConvertStdout,
			CreateChart:            ConvertChart,
			GenerateYaml:           ConvertYaml,
			Replicas:               ConvertReplicas,
			InputFiles:             GlobalFiles,
			OutFile:                ConvertOut,
			Provider:               strings.ToLower(GlobalProvider),
			CreateD:                ConvertDeployment,
			CreateDS:               ConvertDaemonSet,
			CreateRC:               ConvertReplicationController,
			BuildRepo:              ConvertBuildRepo,
			BuildBranch:            ConvertBuildBranch,
			CreateDeploymentConfig: ConvertDeploymentConfig,
			EmptyVols:              ConvertEmptyVols,
		}

		// Validate before doing anything else. Use "bundle" if passed in.
		app.ValidateFlags(GlobalBundle, args, cmd, &ConvertOpt)
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
	convertCmd.Flags().BoolVar(&ConvertDaemonSet, "daemon-set", false, "Generate a Kubernetes daemonset object")
	convertCmd.Flags().BoolVarP(&ConvertDeployment, "deployment", "d", false, "Generate a Kubernetes deployment object")
	convertCmd.Flags().BoolVar(&ConvertReplicationController, "replication-controller", false, "Generate a Kubernetes replication controller object")
	convertCmd.Flags().MarkHidden("chart")
	convertCmd.Flags().MarkHidden("daemon-set")
	convertCmd.Flags().MarkHidden("replication-controller")
	convertCmd.Flags().MarkHidden("deployment")

	// OpenShift only
	convertCmd.Flags().BoolVar(&ConvertDeploymentConfig, "deployment-config", true, "Generate an OpenShift deploymentconfig object")
	convertCmd.Flags().MarkHidden("deployment-config")
	convertCmd.Flags().StringVar(&ConvertBuildRepo, "build-repo", "", "Specify source repository for buildconfig (default remote origin)")
	convertCmd.Flags().MarkHidden("build-repo")
	convertCmd.Flags().StringVar(&ConvertBuildBranch, "build-branch", "", "Specify repository branch to use for buildconfig (default master)")
	convertCmd.Flags().MarkHidden("build-branch")

	// Standard between the two
	convertCmd.Flags().BoolVarP(&ConvertYaml, "yaml", "y", false, "Generate resource files into yaml format")
	convertCmd.Flags().BoolVar(&ConvertStdout, "stdout", false, "Print converted objects to stdout")
	convertCmd.Flags().BoolVar(&ConvertEmptyVols, "emptyvols", false, "Use Empty Volumes. Do not generate PVCs")
	convertCmd.Flags().StringVarP(&ConvertOut, "out", "o", "", "Specify a file name to save objects to")
	convertCmd.Flags().IntVar(&ConvertReplicas, "replicas", 1, "Specify the number of repliaces in the generate resource spec")

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

Resource Flags:
      --build-branch             Specify repository branch to use for buildconfig (default is current branch name)
      --build-repo               Specify source repository for buildconfig (default is current branch's remote url
  -c, --chart                    Create a Helm chart for converted objects
      --daemon-set               Generate a Kubernetes daemonset object
  -d, --deployment               Generate a Kubernetes deployment object
      --deployment-config        Generate an OpenShift deployment config object
      --replication-controller   Generate a Kubernetes replication controller object

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	// Set the help template + add the command to root
	convertCmd.SetHelpTemplate(customHelp)

	RootCmd.AddCommand(convertCmd)
}
