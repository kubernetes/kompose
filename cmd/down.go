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
	"github.com/spf13/cobra"
)

// TODO: comment
var (
	DownNamespace  string
	DownController string
	DownOpt        kobject.ConvertOptions
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Delete instantiated services/deployments from kubernetes",
	Long:  `Delete instantiated services/deployments from kubernetes. (default "kubernetes")`,
	PreRun: func(cmd *cobra.Command, args []string) {

		// Create the Convert options.
		DownOpt = kobject.ConvertOptions{
			InputFiles:      GlobalFiles,
			Provider:        GlobalProvider,
			Namespace:       DownNamespace,
			Controller:      strings.ToLower(DownController),
			IsNamespaceFlag: cmd.Flags().Lookup("namespace").Changed,
		}

		// Validate before doing anything else.
		app.ValidateComposeFile(&DownOpt)
	},
	Run: func(cmd *cobra.Command, args []string) {
		app.Down(DownOpt)
	},
}

func init() {
	downCmd.Flags().StringVar(&DownNamespace, "namespace", "default", " Specify Namespace to deploy your application")
	downCmd.Flags().StringVar(&DownController, "controller", "", `Set the output controller ("deployment"|"daemonSet"|"replicationController")`)
	RootCmd.AddCommand(downCmd)
}
