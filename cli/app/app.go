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
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	// install kubernetes api
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"

	// install kubernetes api
	_ "github.com/openshift/origin/pkg/deploy/api/install"
	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/loader"
	"github.com/skippbox/kompose/pkg/loader/bundle"
	"github.com/skippbox/kompose/pkg/loader/compose"
	"github.com/skippbox/kompose/pkg/transformer"
	"github.com/skippbox/kompose/pkg/transformer/kubernetes"
	"github.com/skippbox/kompose/pkg/transformer/openshift"
)

const (
	DefaultComposeFile = "docker-compose.yml"
)

var inputFormat = "compose"

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func validateFlags(opt kobject.ConvertOptions, singleOutput bool, dabFile, inputFile string) {
	if len(opt.OutFile) != 0 && opt.ToStdout {
		logrus.Fatalf("Error: --out and --stdout can't be set at the same time")
	}
	if opt.CreateChart && opt.ToStdout {
		logrus.Fatalf("Error: chart cannot be generated when --stdout is specified")
	}
	if opt.Replicas < 0 {
		logrus.Fatalf("Error: --replicas cannot be negative")
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
		if opt.CreateDeploymentConfig {
			count++
		}
		if count > 1 {
			logrus.Fatalf("Error: only one kind of Kubernetes resource can be generated when --out or --stdout is specified")
		}
	}
	if len(dabFile) > 0 && len(inputFile) > 0 && inputFile != DefaultComposeFile {
		logrus.Fatalf("Error: compose file and dab file cannot be specified at the same time")
	}
}

// Convert transforms docker compose or dab file to k8s objects
func Convert(c *cli.Context) {
	inputFile := c.String("file")
	dabFile := c.String("bundle")
	outFile := c.String("out")
	generateYaml := c.BoolT("yaml")
	toStdout := c.BoolT("stdout")
	createD := c.BoolT("deployment")
	createDS := c.BoolT("daemonset")
	createRC := c.BoolT("replicationcontroller")
	createChart := c.BoolT("chart")
	replicas := c.Int("replicas")
	singleOutput := len(outFile) != 0 || toStdout
	createDeploymentConfig := c.BoolT("deploymentconfig")

	// Create Deployment by default if no controller has be set
	if !createD && !createDS && !createRC && !createDeploymentConfig {
		createD = true
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}

	file := inputFile
	if len(dabFile) > 0 {
		inputFormat = "bundle"
		file = dabFile
	}

	opt := kobject.ConvertOptions{
		ToStdout:               toStdout,
		CreateD:                createD,
		CreateRC:               createRC,
		CreateDS:               createDS,
		CreateDeploymentConfig: createDeploymentConfig,
		CreateChart:            createChart,
		GenerateYaml:           generateYaml,
		Replicas:               replicas,
		InputFile:              file,
		OutFile:                outFile,
	}

	validateFlags(opt, singleOutput, dabFile, inputFile)

	// loader parses input from file into komposeObject.
	var l loader.Loader
	switch inputFormat {
	case "bundle":
		l = new(bundle.Bundle)
	case "compose":
		l = new(compose.Compose)
	default:
		logrus.Fatalf("Input file format is not supported")
	}

	komposeObject = l.LoadFile(file)

	// transformer maps komposeObject to provider's primitives
	var t transformer.Transformer
	if !createDeploymentConfig {
		t = new(kubernetes.Kubernetes)
	} else {
		t = new(openshift.OpenShift)
	}

	objects := t.Transform(komposeObject, opt)

	// Print output
	kubernetes.PrintList(objects, opt)
}

// Up brings up deployment, svc.
func Up(c *cli.Context) {
	fmt.Println("We are going to create Kubernetes deployments and services for your Dockerized application. \n" +
		"If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. \n")

	factory := cmdutil.NewFactory(nil)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		logrus.Fatalf("Failed to access the Kubernetes cluster. Make sure you have a Kubernetes cluster running: %v", err)
	}
	client := client.NewOrDie(clientConfig)

	inputFile := c.String("file")
	dabFile := c.String("bundle")

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}

	file := inputFile
	if len(dabFile) > 0 {
		inputFormat = "bundle"
		file = dabFile
	}

	opt := kobject.ConvertOptions{
		Replicas: 1,
		CreateD:  true,
	}

	validateFlags(opt, false, dabFile, inputFile)

	// loader parses input from file into komposeObject.
	var l loader.Loader
	switch inputFormat {
	case "bundle":
		l = new(bundle.Bundle)
	case "compose":
		l = new(compose.Compose)
	default:
		logrus.Fatalf("Input file format is not supported")
	}
	komposeObject = l.LoadFile(file)

	t := new(kubernetes.Kubernetes)

	//Convert komposeObject to K8S controllers
	objects := t.Transform(komposeObject, opt)
	sortServicesFirst(&objects)

	//Submit objects to K8s endpoint
	kubernetes.CreateObjects(client, objects)
}

// the objects that we get can be in any order this keeps services first
// according to best practice kubernetes services should be created first
// http://kubernetes.io/docs/user-guide/config-best-practices/
func sortServicesFirst(objs *[]runtime.Object) {
	var svc, others, ret []runtime.Object

	for _, obj := range *objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Service" {
			svc = append(svc, obj)
		} else {
			others = append(others, obj)
		}
	}
	ret = append(ret, svc...)
	ret = append(ret, others...)
	*objs = ret
}
