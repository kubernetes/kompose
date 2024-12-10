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
	"fmt"

	"github.com/kubernetes/kompose/pkg/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kompose",
	Run: func(cmd *cobra.Command, args []string) {
		// See pkg/version/version.go for more information as to why we use the git commit / hash value
		fmt.Println("Kompose version:", "v"+version.VERSION)
		fmt.Println("Git HEAD:", version.GITCOMMIT)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
