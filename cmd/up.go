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
)

// TODO: comment
var (
	UpReplicas  int
	UpEmptyVols bool
	UpOpt       kobject.ConvertOptions
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy your Dockerized application to a container orchestrator.",
	Long:  `Deploy your Dockerized application to a container orchestrator. (default "kubernetes")`,
	PreRun: func(cmd *cobra.Command, args []string) {

		// Create the Convert options.
		UpOpt = kobject.ConvertOptions{
			Replicas:   UpReplicas,
			InputFiles: GlobalFiles,
			Provider:   strings.ToLower(GlobalProvider),
			EmptyVols:  UpEmptyVols,
		}

		// Validate before doing anything else.
		app.ValidateComposeFile(cmd, &UpOpt)
	},
	Run: func(cmd *cobra.Command, args []string) {
		app.Up(UpOpt)
	},
}

func init() {
	upCmd.Flags().BoolVar(&UpEmptyVols, "emptyvols", false, "Use empty volumes. Do not generate PersistentVolumeClaim")
	upCmd.Flags().IntVar(&UpReplicas, "replicas", 1, "Specify the number of replicas generated")
	RootCmd.AddCommand(upCmd)
}
