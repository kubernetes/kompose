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

package command

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kubernetes-incubator/kompose/cli/app"
	"github.com/urfave/cli"
)

// Hook for erroring and exit out on warning
type errorOnWarningHook struct{}

// array consisting of our common conversion flags that will get passed along
// for the autocomplete aspect
var (
	commonConvertFlagsList = []string{"out", "replicas", "yaml", "stdout", "emptyvols"}
)

func (errorOnWarningHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

func (errorOnWarningHook) Fire(entry *logrus.Entry) error {
	logrus.Fatalln(entry.Message)
	return nil
}

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {

	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else if c.GlobalBool("suppress-warnings") {
		logrus.SetLevel(logrus.ErrorLevel)
	} else if c.GlobalBool("error-on-warning") {
		hook := errorOnWarningHook{}
		logrus.AddHook(hook)
	}

	// First command added was dummy convert command so removing it
	c.App.Commands = c.App.Commands[1:]
	provider := strings.ToLower(c.GlobalString("provider"))
	switch provider {
	case "kubernetes":
		c.App.Commands = append(c.App.Commands, ConvertKubernetesCommand())
	case "openshift":
		c.App.Commands = append(c.App.Commands, ConvertOpenShiftCommand())
	default:
		logrus.Fatalf("Unknown provider. Supported providers are kubernetes and openshift.")
	}

	return nil
}

// When user tries out `kompose -h`, the convert option should be visible
// so adding a dummy `convert` command, real convert commands depending on Providers
// mentioned are added in `BeforeApp` function
func ConvertCommandDummy() cli.Command {
	command := cli.Command{
		Name:  "convert",
		Usage: fmt.Sprintf("Convert Docker Compose file (e.g. %s) to Kubernetes/OpenShift objects", app.DefaultComposeFile),
	}
	return command
}

// Generate the Bash completion flag taking the common flags plus whatever is
// passed into the function to correspond to the primary command specific args
func generateBashCompletion(args []string) {
	commonArgs := []string{"bundle", "file", "suppress-warnings", "verbose", "error-on-warning", "provider"}
	flags := append(commonArgs, args...)

	for _, f := range flags {
		fmt.Printf("--%s\n", f)
	}
}

// ConvertKubernetesCommand defines the kompose convert subcommand for Kubernetes provider
func ConvertKubernetesCommand() cli.Command {
	command := cli.Command{
		Name:  "convert",
		Usage: fmt.Sprintf("Convert Docker Compose file (e.g. %s) to Kubernetes objects", app.DefaultComposeFile),
		Action: func(c *cli.Context) {
			app.Convert(c)
		},
		BashComplete: func(c *cli.Context) {
			flags := []string{"chart", "deployment", "daemonset", "replicationcontroller"}
			generateBashCompletion(append(flags, commonConvertFlagsList...))
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "chart,c",
				Usage: "Create a Helm chart for converted objects",
			},
			cli.BoolFlag{
				Name:  "deployment,d",
				Usage: "Generate a Kubernetes deployment object (default on)",
			},
			cli.BoolFlag{
				Name:  "daemonset,ds",
				Usage: "Generate a Kubernetes daemonset object",
			},
			cli.BoolFlag{
				Name:  "replicationcontroller,rc",
				Usage: "Generate a Kubernetes replication controller object",
			},
		},
	}
	command.Flags = append(command.Flags, commonConvertFlags()...)
	return command
}

// ConvertOpenShiftCommand defines the kompose convert subcommand for OpenShift provider
func ConvertOpenShiftCommand() cli.Command {
	command := cli.Command{
		Name:  "convert",
		Usage: fmt.Sprintf("Convert Docker Compose file (e.g. %s) to OpenShift objects", app.DefaultComposeFile),
		Action: func(c *cli.Context) {
			app.Convert(c)
		},
		BashComplete: func(c *cli.Context) {
			flags := []string{"deploymentconfig"}
			generateBashCompletion(append(flags, commonConvertFlagsList...))
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "deploymentconfig,dc",
				Usage: "Generate a OpenShift DeploymentConfig object",
			},
		},
	}
	command.Flags = append(command.Flags, commonConvertFlags()...)
	return command
}

func commonConvertFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "out,o",
			Usage:  "Specify path to a file or a directory to save generated objects into. If path is a directory, the objects are stored in that directory. If path is a file, then objects are stored in that single file. File is created if it does not exist.",
			EnvVar: "OUTPUT_FILE",
		},
		cli.IntFlag{
			Name:  "replicas",
			Value: 1,
			Usage: "Specify the number of replicas in the generated resource spec (default 1)",
		},
		cli.BoolFlag{
			Name:  "yaml, y",
			Usage: "Generate resource file in yaml format",
		},
		cli.BoolFlag{
			Name:  "stdout",
			Usage: "Print converted objects to stdout",
		},
		cli.BoolFlag{
			Name:  "emptyvols",
			Usage: "Use Empty Volumes. Don't generate PVCs",
		},
	}
}

// UpCommand defines the kompose up subcommand.
func UpCommand() cli.Command {
	return cli.Command{
		Name:  "up",
		Usage: "Deploy your Dockerized application to Kubernetes (default: creating Kubernetes deployment and service)",
		Action: func(c *cli.Context) {
			app.Up(c)
		},
		BashComplete: func(c *cli.Context) {
			flags := []string{"emptyvols"}
			generateBashCompletion(flags)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "emptyvols",
				Usage: "Use Empty Volumes. Don't generate PVCs",
			},
		},
	}
}

// DownCommand defines the kompose down subcommand.
func DownCommand() cli.Command {
	return cli.Command{
		Name:  "down",
		Usage: "Delete instantiated services/deployments from kubernetes",
		Action: func(c *cli.Context) {
			app.Down(c)
		},
		BashComplete: func(c *cli.Context) {
			flags := []string{"emptyvols"}
			generateBashCompletion(flags)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "emptyvols",
				Usage: "Use Empty Volumes. Don't generate PVCs",
			},
		},
	}
}

// CommonFlags defines the flags that are in common for all subcommands.
func CommonFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "bundle,dab",
			Usage:  "Specify a Distributed Application Bundle (DAB) file",
			EnvVar: "DAB_FILE",
		},

		cli.StringFlag{
			Name:   "file,f",
			Usage:  fmt.Sprintf("Specify an alternative compose file (default: %s)", app.DefaultComposeFile),
			Value:  app.DefaultComposeFile,
			EnvVar: "COMPOSE_FILE",
		},
		// creating a flag to suppress warnings
		cli.BoolFlag{
			Name:  "suppress-warnings",
			Usage: "Suppress all warnings",
		},
		// creating a flag to show all kinds of warnings
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show all type of logs",
		},
		// flag to treat any warning as error
		cli.BoolFlag{
			Name:  "error-on-warning",
			Usage: "Treat any warning as error",
		},
		// mention the end provider
		cli.StringFlag{
			Name:   "provider",
			Usage:  "Generate artifacts for this provider",
			Value:  app.DefaultProvider,
			EnvVar: "PROVIDER",
		},
	}
}
