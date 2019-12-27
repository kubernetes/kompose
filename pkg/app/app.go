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

package app

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	// install kubernetes api
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	// install OpenShift api
	_ "github.com/openshift/origin/pkg/build/api/install"
	_ "github.com/openshift/origin/pkg/deploy/api/install"
	_ "github.com/openshift/origin/pkg/image/api/install"
	_ "github.com/openshift/origin/pkg/route/api/install"

	"os"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	"github.com/kubernetes/kompose/pkg/transformer/openshift"
)

const (
	// DefaultComposeFile name of the file that kompose will use if no file is explicitly set
	DefaultComposeFile = "docker-compose.yml"
)

const (
	// ProviderKubernetes is provider kubernetes
	ProviderKubernetes = "kubernetes"
	// ProviderOpenshift is provider openshift
	ProviderOpenshift = "openshift"
	// DefaultProvider - provider that will be used if there is no provider was explicitly set
	DefaultProvider = ProviderKubernetes
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
	log.Debugf("Checking validation of provider: %s", provider)

	// OpenShift specific flags
	deploymentConfig := cmd.Flags().Lookup("deployment-config").Changed
	buildRepo := cmd.Flags().Lookup("build-repo").Changed
	buildBranch := cmd.Flags().Lookup("build-branch").Changed

	// Kubernetes specific flags
	chart := cmd.Flags().Lookup("chart").Changed
	daemonSet := cmd.Flags().Lookup("daemon-set").Changed
	replicationController := cmd.Flags().Lookup("replication-controller").Changed
	deployment := cmd.Flags().Lookup("deployment").Changed

	// Get the controller
	controller := opt.Controller
	log.Debugf("Checking validation of controller: %s", controller)

	// Check validations against provider flags
	switch {
	case provider == ProviderOpenshift:
		if chart {
			log.Fatalf("--chart, -c is a Kubernetes only flag")
		}
		if daemonSet {
			log.Fatalf("--daemon-set is a Kubernetes only flag")
		}
		if replicationController {
			log.Fatalf("--replication-controller is a Kubernetes only flag")
		}
		if deployment {
			log.Fatalf("--deployment, -d is a Kubernetes only flag")
		}
		if controller == "daemonset" || controller == "replicationcontroller" || controller == "deployment" {
			log.Fatalf("--controller= daemonset, replicationcontroller or deployment is a Kubernetes only flag")
		}
	case provider == ProviderKubernetes:
		if deploymentConfig {
			log.Fatalf("--deployment-config is an OpenShift only flag")
		}
		if buildRepo {
			log.Fatalf("--build-repo is an Openshift only flag")
		}
		if buildBranch {
			log.Fatalf("--build-branch is an Openshift only flag")
		}
		if controller == "deploymentconfig" {
			log.Fatalf("--controller=deploymentConfig is an OpenShift only flag")
		}
	}

	// Standard checks regardless of provider
	if len(opt.OutFile) != 0 && opt.ToStdout {
		log.Fatalf("Error: --out and --stdout can't be set at the same time")
	}

	if opt.CreateChart && opt.ToStdout {
		log.Fatalf("Error: chart cannot be generated when --stdout is specified")
	}

	if opt.Replicas < 0 {
		log.Fatalf("Error: --replicas cannot be negative")
	}

	if len(bundle) > 0 {
		inputFormat = "bundle"
		log.Fatalf("DAB / bundle (--bundle | -b) is no longer supported. See issue: https://github.com/kubernetes/kompose/issues/390")
		opt.InputFiles = []string{bundle}
	}

	if len(bundle) > 0 && isFileSet {
		log.Fatalf("Error: 'compose' file and 'dab' file cannot be specified at the same time")
	}

	if len(args) != 0 {
		log.Fatal("Unknown Argument(s): ", strings.Join(args, ","))
	}

	if opt.GenerateJSON && opt.GenerateYaml {
		log.Fatalf("YAML and JSON format cannot be provided at the same time")
	}

	if opt.Volumes != "persistentVolumeClaim" && opt.Volumes != "emptyDir" && opt.Volumes != "hostPath" && opt.Volumes != "configMap" {
		log.Fatal("Unknown Volume type: ", opt.Volumes, ", possible values are: persistentVolumeClaim, configMap and emptyDir")
	}
}

// ValidateComposeFile validates the compose file provided for conversion
func ValidateComposeFile(opt *kobject.ConvertOptions) {
	if len(opt.InputFiles) == 0 {
		// Here docker-compose is the input
		opt.InputFiles = []string{DefaultComposeFile}
		_, err := os.Stat(DefaultComposeFile)
		if err != nil {
			log.Debugf("'%s' not found: %v", DefaultComposeFile, err)
			opt.InputFiles = []string{"docker-compose.yaml"}
			_, err = os.Stat("docker-compose.yaml")
			if err != nil {
				log.Fatalf("No 'docker-compose' file found: %v", err)
			}
		}
	}
}

func validateControllers(opt *kobject.ConvertOptions) {

	singleOutput := len(opt.OutFile) != 0 || opt.OutFile == "-" || opt.ToStdout
	if opt.Provider == ProviderKubernetes {
		// create deployment by default if no controller has been set
		if !opt.CreateD && !opt.CreateDS && !opt.CreateRC && opt.Controller == "" {
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
				log.Fatalf("Error: only one kind of Kubernetes resource can be generated when --out or --stdout is specified")
			}
		}

	} else if opt.Provider == ProviderOpenshift {
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
				log.Fatalf("Error: only one kind of OpenShift resource can be generated when --out or --stdout is specified")
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
		log.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject, err = l.LoadFile(opt.InputFiles)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Get a transformer that maps komposeObject to provider's primitives
	t := getTransformer(opt)

	// Do the transformation
	objects, err := t.Transform(komposeObject, opt)

	if err != nil {
		log.Fatalf(err.Error())
	}

	// Print output
	err = kubernetes.PrintList(objects, opt)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// Up brings up deployment, svc.
func Up(opt kobject.ConvertOptions) {

	validateControllers(&opt)

	// loader parses input from file into komposeObject.
	l, err := loader.GetLoader(inputFormat)
	if err != nil {
		log.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject, err = l.LoadFile(opt.InputFiles)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Get the transformer
	t := getTransformer(opt)

	//Submit objects to provider
	errDeploy := t.Deploy(komposeObject, opt)
	if errDeploy != nil {
		log.Fatalf("Error while deploying application: %s", errDeploy)
	}
}

// Down deletes all deployment, svc.
func Down(opt kobject.ConvertOptions) {

	validateControllers(&opt)

	// loader parses input from file into komposeObject.
	l, err := loader.GetLoader(inputFormat)
	if err != nil {
		log.Fatal(err)
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject, err = l.LoadFile(opt.InputFiles)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Get the transformer
	t := getTransformer(opt)

	//Remove deployed application
	errUndeploy := t.Undeploy(komposeObject, opt)
	if errUndeploy != nil {
		for _, err = range errUndeploy {
			log.Fatalf("Error while deleting application: %s", err)
		}
	}

}

// Convenience method to return the appropriate Transformer based on
// what provider we are using.
func getTransformer(opt kobject.ConvertOptions) transformer.Transformer {
	var t transformer.Transformer
	if opt.Provider == DefaultProvider {
		// Create/Init new Kubernetes object with CLI opts
		t = &kubernetes.Kubernetes{Opt: opt}
	} else {
		// Create/Init new OpenShift object that is initialized with a newly
		// created Kubernetes object. Openshift inherits from Kubernetes
		t = &openshift.OpenShift{Kubernetes: kubernetes.Kubernetes{Opt: opt}}
	}
	return t
}
