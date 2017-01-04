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

package app

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	// install kubernetes api
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	// install OpenShift api
	_ "github.com/openshift/origin/pkg/build/api/install"
	_ "github.com/openshift/origin/pkg/deploy/api/install"
	_ "github.com/openshift/origin/pkg/image/api/install"
	_ "github.com/openshift/origin/pkg/route/api/install"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/loader"
	"github.com/kubernetes-incubator/kompose/pkg/transformer"
	"github.com/kubernetes-incubator/kompose/pkg/transformer/kubernetes"
	"github.com/kubernetes-incubator/kompose/pkg/transformer/openshift"
)

const (
	// DefaultComposeFile name of the file that kompose will use if no file is explicitly set
	DefaultComposeFile = "docker-compose.yml"
	// DefaultProvider - provider that will be used if there is no provider was explicitly set
	DefaultProvider = "kubernetes"
)

var inputFormat = "compose"

// ValidateFlags validates all command line flags
func ValidateFlags(bundle string, args []string, cmd *cobra.Command, opt *kobject.ConvertOptions) {

	// Check to see if the "file" has changed from the default flag value
	isFileSet := cmd.Flags().Lookup("file").Changed

	if opt.OutFile == "-" {
		opt.ToStdout = true
		opt.OutFile = ""
	}

	// Get the provider
	provider := cmd.Flags().Lookup("provider").Value.String()
	logrus.Debug("Checking validation of provider %s", provider)

	// OpenShift specific flags
	deploymentConfig := cmd.Flags().Lookup("deployment-config").Changed
	buildRepo := cmd.Flags().Lookup("build-repo").Changed
	buildBranch := cmd.Flags().Lookup("build-branch").Changed

	// Kubernetes specific flags
	chart := cmd.Flags().Lookup("chart").Changed
	daemonSet := cmd.Flags().Lookup("daemon-set").Changed
	replicationController := cmd.Flags().Lookup("replication-controller").Changed
	deployment := cmd.Flags().Lookup("deployment").Changed

	// Check validations against provider flags
	switch {
	case provider == "openshift":
		if chart {
			logrus.Fatalf("--chart, -c is a Kubernetes only flag")
		}
		if daemonSet {
			logrus.Fatalf("--daemon-set is a Kubernetes only flag")
		}
		if replicationController {
			logrus.Fatalf("--replication-controller is a Kubernetes only flag")
		}
		if deployment {
			logrus.Fatalf("--deployment, -d is a Kubernetes only flag")
		}
	case provider == "kubernetes":
		if deploymentConfig {
			logrus.Fatalf("--deployment-config is an OpenShift only flag")
		}
		if buildRepo {
			logrus.Fatalf("--build-repo is an Openshift only flag")
		}
		if buildBranch {
			logrus.Fatalf("--build-branch is an Openshift only flag")
		}
	}

	// Standard checks regardless of provider
	if len(opt.OutFile) != 0 && opt.ToStdout {
		logrus.Fatalf("Error: --out and --stdout can't be set at the same time")
	}

	if opt.CreateChart && opt.ToStdout {
		logrus.Fatalf("Error: chart cannot be generated when --stdout is specified")
	}

	if opt.Replicas < 0 {
		logrus.Fatalf("Error: --replicas cannot be negative")
	}

	if len(bundle) > 0 {
		inputFormat = "bundle"
		opt.InputFiles = []string{bundle}
	}

	if len(bundle) > 0 && isFileSet {
		logrus.Fatalf("Error: 'compose' file and 'dab' file cannot be specified at the same time")
	}

	if len(args) != 0 {
		logrus.Fatal("Unknown Argument(s): ", strings.Join(args, ","))
	}

	if opt.GenerateJSON && opt.GenerateYaml {
		logrus.Fatalf("YAML and JSON format cannot be provided at the same time")
	}
}

func validateControllers(opt *kobject.ConvertOptions) {

	singleOutput := len(opt.OutFile) != 0 || opt.OutFile == "-" || opt.ToStdout

	if opt.Provider == "kubernetes" {
		// create deployment by default if no controller has been set
		if !opt.CreateD && !opt.CreateDS && !opt.CreateRC {
			opt.CreateD = true
		}
		if singleOutput {
			count := 0
			if opt.CreateD {
				count++
			}
			if opt.CreateDS {
				count++
			}
			if opt.CreateRC {
				count++
			}
			if count > 1 {
				logrus.Fatalf("Error: only one kind of Kubernetes resource can be generated when --out or --stdout is specified")
			}
		}

	} else if opt.Provider == "openshift" {
		// create deploymentconfig by default if no controller has been set
		if !opt.CreateDeploymentConfig {
			opt.CreateDeploymentConfig = true
		}
		if singleOutput {
			count := 0
			if opt.CreateDeploymentConfig {
				count++
			}
			// Add more controllers here once they are available in OpenShift
			// if opt.foo {count++}

			if count > 1 {
				logrus.Fatalf("Error: only one kind of OpenShift resource can be generated when --out or --stdout is specified")
			}
		}
	}
}

// Convert transforms docker compose or dab file to k8s objects
func Convert(opt kobject.ConvertOptions) {

	validateControllers(&opt)

	// loader parses input from file into komposeObject.
	l, err := loader.GetLoader(inputFormat)
	if err != nil {
		logrus.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject = l.LoadFile(opt.InputFiles)

	// Get a transformer that maps komposeObject to provider's primitives
	t := getTransformer(opt)

	// Do the transformation
	objects := t.Transform(komposeObject, opt)

	// Print output
	kubernetes.PrintList(objects, opt)
}

// Up brings up deployment, svc.
func Up(opt kobject.ConvertOptions) {

	validateControllers(&opt)

	// loader parses input from file into komposeObject.
	l, err := loader.GetLoader(inputFormat)
	if err != nil {
		logrus.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject = l.LoadFile(opt.InputFiles)

	// Get the transformer
	t := getTransformer(opt)

	//Submit objects to provider
	errDeploy := t.Deploy(komposeObject, opt)
	if errDeploy != nil {
		logrus.Fatalf("Error while deploying application: %s", errDeploy)
	}
}

// Down deletes all deployment, svc.
func Down(opt kobject.ConvertOptions) {

	validateControllers(&opt)

	// loader parses input from file into komposeObject.
	l, err := loader.GetLoader(inputFormat)
	if err != nil {
		logrus.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject = l.LoadFile(opt.InputFiles)

	// Get the transformer
	t := getTransformer(opt)

	//Remove deployed application
	errUndeploy := t.Undeploy(komposeObject, opt)
	if errUndeploy != nil {
		logrus.Fatalf("Error while deleting application: %s", errUndeploy)
	}

}

// Convenience method to return the appropriate Transformer based on
// what provider we are using.
func getTransformer(opt kobject.ConvertOptions) transformer.Transformer {
	var t transformer.Transformer
	if opt.Provider == "kubernetes" {
		// Create/Init new Kubernetes object with CLI opts
		t = &kubernetes.Kubernetes{Opt: opt}
	} else {
		// Create/Init new OpenShift object that is initialized with a newly
		// created Kubernetes object. Openshift inherits from Kubernetes
		t = &openshift.OpenShift{Kubernetes: kubernetes.Kubernetes{Opt: opt}}
	}
	return t
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		logrus.Fatal(err)
	}
	if response == "yes" {
		return true
	} else if response == "no" {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}
