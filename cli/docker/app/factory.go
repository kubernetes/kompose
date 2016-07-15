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

package app

import (
	"github.com/urfave/cli"
	"github.com/skippbox/kompose/cli/command"
	"github.com/docker/libcompose/cli/logger"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
)

// ProjectFactory is a struct that hold the app.ProjectFactory implementation.
type ProjectFactory struct {
}

// Create implements ProjectFactory.Create using docker client.
func (p *ProjectFactory) Create(c *cli.Context) (*project.Project, error) {
	context := &docker.Context{}
	context.LoggerFactory = logger.NewColorLoggerFactory()
	//Populate(context, c)
	command.Populate(&context.Context, c)

	return docker.NewProject(context)
}
