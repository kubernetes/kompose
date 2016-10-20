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

package main

import (
	"os"

	"github.com/kubernetes-incubator/kompose/cli/command"
	"github.com/kubernetes-incubator/kompose/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kompose"
	app.Usage = "A tool helping Docker Compose users move to Kubernetes."
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"
	app.Author = "Kompose Contributors"
	app.Email = "https://github.com/kubernetes-incubator/kompose"
	app.EnableBashCompletion = true
	app.Before = command.BeforeApp
	app.Flags = append(command.CommonFlags())
	app.Commands = []cli.Command{
		// NOTE: Always add this in first, because this dummy command will be removed later
		// in  command.BeforeApp function and provider specific command will be added
		command.ConvertCommandDummy(),
		// command.ConvertKubernetesCommand or command.ConvertOpenShiftCommand
		// is added depending on provider mentioned.

		command.UpCommand(),
		command.DownCommand(),
		// TODO: enable these commands and update docs once we fix them
		//command.PsCommand(),
		//command.ScaleCommand(),
	}

	app.Run(os.Args)
}
