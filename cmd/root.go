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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Logrus hooks

// Hook for error and exit out on warning
type errorOnWarningHook struct{}

func (errorOnWarningHook) Levels() []log.Level {
	return []log.Level{log.WarnLevel}
}

func (errorOnWarningHook) Fire(entry *log.Entry) error {
	log.Fatalln(entry.Message)
	return nil
}

// TODO: comment
var (
	GlobalProvider              string
	GlobalVerbose               bool
	GlobalSuppressWarnings      bool
	GlobalErrorOnWarning        bool
	GlobalFiles                 []string
	GlobalDefaultLimitsRequests bool
)

// RootCmd root level flags and commands
var RootCmd = &cobra.Command{
	Use:   "kompose",
	Short: "A tool helping Docker Compose users move to Kubernetes",
	Long:  `Kompose is a tool to help users who are familiar with docker-compose move to Kubernetes.`,
	Example: `  kompose --file compose.yaml convert
  kompose -f first.yaml -f second.yaml convert
  kompose --provider openshift --file compose.yaml convert
  kompose completion bash`,
	SilenceErrors: true,
	// PersistentPreRun will be "inherited" by all children and ran before *every* command unless
	// the child has overridden the functionality. This functionality was implemented to check / modify
	// all global flag calls regardless of app call.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Add extra logging when verbosity is passed
		if GlobalVerbose {
			log.SetLevel(log.DebugLevel)
		}

		// Disable the timestamp (Kompose is too fast!)
		formatter := new(log.TextFormatter)
		formatter.DisableTimestamp = true
		formatter.ForceColors = true
		log.SetFormatter(formatter)

		// Set the appropriate suppress warnings and error on warning flags
		if GlobalSuppressWarnings {
			log.SetLevel(log.ErrorLevel)
		} else if GlobalErrorOnWarning {
			hook := errorOnWarningHook{}
			log.AddHook(hook)
		}

		// Error out of the user has not chosen Kubernetes or OpenShift
		provider := strings.ToLower(GlobalProvider)
		if provider != "kubernetes" && provider != "openshift" {
			log.Fatalf("%s is an unsupported provider. Supported providers are: 'kubernetes', 'openshift'.", GlobalProvider)
		}

		v := viper.New()
		v.BindEnv("file", "COMPOSE_FILE")

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			configName := f.Name
			if configName == "file" && !f.Changed && v.IsSet(configName) {
				GlobalFiles = v.GetStringSlice(configName)
			}
		})
	},
}

// Execute executes the root level command.
// It returns an erorr if any.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&GlobalVerbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolVar(&GlobalSuppressWarnings, "suppress-warnings", false, "Suppress all warnings")
	RootCmd.PersistentFlags().BoolVar(&GlobalErrorOnWarning, "error-on-warning", false, "Treat any warning as an error")
	RootCmd.PersistentFlags().StringSliceVarP(&GlobalFiles, "file", "f", []string{}, "Specify an alternative compose file")
	RootCmd.PersistentFlags().StringVar(&GlobalProvider, "provider", "kubernetes", "Specify a provider. Kubernetes or OpenShift.")
	RootCmd.PersistentFlags().BoolVarP(&GlobalDefaultLimitsRequests, "limitsrequests", "l", false, "Output default Limits and Requests")
}
