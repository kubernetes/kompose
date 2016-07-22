/*
Copyright 2016 Skippbox, Ltd All rights reserved.

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
	"github.com/skippbox/kompose/cli/app"
	"github.com/urfave/cli"
)

// ConvertCommand defines the kompose convert subcommand.
func ConvertCommand() cli.Command {
	return cli.Command{
		Name:  "convert",
		Usage: "Convert docker-compose.yml to Kubernetes objects",
		Action: func(c *cli.Context) {
			app.Convert(c)
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "file,f",
				Usage:  "Specify an alternate compose file (default: docker-compose.yml)",
				Value:  "docker-compose.yml",
				EnvVar: "COMPOSE_FILE",
			},
			cli.StringFlag{
				Name:   "out,o",
				Usage:  "Specify file name in order to save objects into",
				EnvVar: "OUTPUT_FILE",
			},
			cli.BoolFlag{
				Name:  "deployment,d",
				Usage: "Generate a deployment resource file (default on)",
			},
			cli.BoolFlag{
				Name:  "daemonset,ds",
				Usage: "Generate a daemonset resource file",
			},
			cli.BoolFlag{
				Name:  "deploymentconfig,dc",
				Usage: "Generate a DeploymentConfig for OpenShift",
			},
			cli.BoolFlag{
				Name:  "replicationcontroller,rc",
				Usage: "Generate a replication controller resource file",
			},
			cli.BoolFlag{
				Name:  "replicaset,rs",
				Usage: "Generate a replicaset resource file",
			},
			cli.IntFlag{
				Name:  "replicas",
				Value: 1,
				Usage: "Specify the number of replicas in the generated resource spec (default 1)",
			},
			cli.BoolFlag{
				Name:  "chart,c",
				Usage: "Create a chart deployment",
			},
			cli.BoolFlag{
				Name:  "yaml, y",
				Usage: "Generate resource file in yaml format",
			},
			cli.BoolFlag{
				Name:  "stdout",
				Usage: "Print Kubernetes objects to stdout",
			},
			// FIXME: this flag should be used together with --file/-f in order to specify dab file.
			cli.BoolFlag{
				Name:  "from-bundles",
				Usage: "Getting input from docker DAB file",
			},
		},
	}
}

// UpCommand defines the kompose up subcommand.
func UpCommand() cli.Command {
	return cli.Command{
		Name:  "up",
		Usage: "Submit rc, svc objects to kubernetes API endpoint",
		Action: func(c *cli.Context) {
			app.Up(c)
		},
	}
}

// PsCommand defines the kompose ps subcommand.
func PsCommand() cli.Command {
	return cli.Command{
		Name:  "ps",
		Usage: "Get active data in the kubernetes cluster",
		Action: func(c *cli.Context) {
			app.Ps(c)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "service,svc",
				Usage: "Get active services",
			},
			cli.BoolFlag{
				Name:  "replicationcontroller,rc",
				Usage: "Get active replication controller",
			},
		},
	}
}

// DeleteCommand defines the kompose delete subcommand.
func DeleteCommand() cli.Command {
	return cli.Command{
		Name:  "delete",
		Usage: "Remove instantiated services/rc from kubernetes",
		Action: func(c *cli.Context) {
			app.Delete(c)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "replicationcontroller,rc",
				Usage: "Remove active replication controllers",
			},
			cli.BoolFlag{
				Name:  "service,svc",
				Usage: "Remove active services",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "Name of the object to remove",
			},
		},
	}
}

// ScaleCommand defines the kompose up subcommand.
func ScaleCommand() cli.Command {
	return cli.Command{
		Name:  "scale",
		Usage: "Globally scale instantiated replication controllers",
		Action: func(c *cli.Context) {
			app.Scale(c)
		},
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "scale",
				Usage: "New number of replicas",
			},
			cli.StringFlag{
				Name:  "replicationcontroller,rc",
				Usage: "A specific replication controller to scale",
			},
		},
	}
}

// CommonFlags defines the flags that are in common for all subcommands.
func CommonFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name: "verbose,debug",
		},
		cli.StringFlag{
			Name:   "file,f",
			Usage:  "Specify an alternate compose file (default: docker-compose.yml)",
			Value:  "docker-compose.yml",
			EnvVar: "COMPOSE_FILE",
		},
	}
}
