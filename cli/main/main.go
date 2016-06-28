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

	"github.com/urfave/cli"
	dockerApp "github.com/skippbox/kompose2/cli/docker/app"
	"github.com/skippbox/kompose2/version"
	"github.com/skippbox/kompose2/cli/command"
	cliApp "github.com/skippbox/kompose2/cli/app"
)

func main() {
	factory := &dockerApp.ProjectFactory{}

	app := cli.NewApp()
	app.Name = "kompose"
	app.Usage = "Command line interface for Skippbox."
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"
	app.Author = "Skippbox Compose Contributors"
	app.Email = "https://github.com/skippbox/kompose"
	app.Before = cliApp.BeforeApp
	app.Flags = append(command.CommonFlags())
	app.Commands = []cli.Command{
		command.ConvertCommand(factory),
		command.UpCommand(factory),
		command.PsCommand(factory),
		command.DeleteCommand(factory),
		command.ScaleCommand(factory),
	}

	app.Run(os.Args)
}
