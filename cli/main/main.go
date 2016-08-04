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

	cliApp "github.com/skippbox/kompose/cli/app"
	"github.com/skippbox/kompose/cli/command"
	"github.com/skippbox/kompose/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kompose"
	app.Usage = "A tool helping Docker Compose users move to Kubernetes."
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"
	app.Author = "Skippbox Kompose Contributors"
	app.Email = "https://github.com/skippbox/kompose"
	app.EnableBashCompletion = true
	app.Before = cliApp.BeforeApp
	app.Flags = append(command.CommonFlags())
	app.Commands = []cli.Command{
		command.ConvertCommand(),
		command.UpCommand(),
		// TODO: enable these commands and update docs once we fix them
		//command.PsCommand(),
		//command.DeleteCommand(),
		//command.ScaleCommand(),
	}

	app.Run(os.Args)
}
