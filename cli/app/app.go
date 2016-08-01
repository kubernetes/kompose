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
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	// install kubernetes api
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	// install kubernetes api
	_ "github.com/openshift/origin/pkg/deploy/api/install"

	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/loader"
	"github.com/skippbox/kompose/pkg/transformer"
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

// Ps lists all rc, svc.
func Ps(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//if c.BoolT("svc") {
	//	fmt.Printf("%-20s%-20s%-20s%-20s\n", "Name", "Cluster IP", "Ports", "Selectors")
	//	for name := range p.Configs {
	//		var ports string
	//		var selectors string
	//		services, err := client.Services(api.NamespaceDefault).Get(name)
	//
	//		if err != nil {
	//			logrus.Debugf("Cannot find service for: ", name)
	//		} else {
	//
	//			for i := range services.Spec.Ports {
	//				p := strconv.Itoa(int(services.Spec.Ports[i].Port))
	//				ports += ports + string(services.Spec.Ports[i].Protocol) + "(" + p + "),"
	//			}
	//
	//			for k, v := range services.ObjectMeta.Labels {
	//				selectors += selectors + k + "=" + v + ","
	//			}
	//
	//			ports = strings.TrimSuffix(ports, ",")
	//			selectors = strings.TrimSuffix(selectors, ",")
	//
	//			fmt.Printf("%-20s%-20s%-20s%-20s\n", services.ObjectMeta.Name,
	//				services.Spec.ClusterIP, ports, selectors)
	//		}
	//
	//	}
	//}
	//
	//if c.BoolT("rc") {
	//	fmt.Printf("%-15s%-15s%-30s%-10s%-20s\n", "Name", "Containers", "Images",
	//		"Replicas", "Selectors")
	//	for name := range p.Configs {
	//		var selectors string
	//		var containers string
	//		var images string
	//		rc, err := client.ReplicationControllers(api.NamespaceDefault).Get(name)
	//
	//		/* Should grab controller, container, image, selector, replicas */
	//
	//		if err != nil {
	//			logrus.Debugf("Cannot find rc for: ", string(name))
	//		} else {
	//
	//			for k, v := range rc.Spec.Selector {
	//				selectors += selectors + k + "=" + v + ","
	//			}
	//
	//			for i := range rc.Spec.Template.Spec.Containers {
	//				c := rc.Spec.Template.Spec.Containers[i]
	//				containers += containers + c.Name + ","
	//				images += images + c.Image + ","
	//			}
	//			selectors = strings.TrimSuffix(selectors, ",")
	//			containers = strings.TrimSuffix(containers, ",")
	//			images = strings.TrimSuffix(images, ",")
	//
	//			fmt.Printf("%-15s%-15s%-30s%-10d%-20s\n", rc.ObjectMeta.Name, containers,
	//				images, rc.Spec.Replicas, selectors)
	//		}
	//	}
	//}

}

// Delete deletes all rc, svc.
func Delete(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//for name := range p.Configs {
	//	if len(c.String("name")) > 0 && name != c.String("name") {
	//		continue
	//	}
	//
	//	if c.BoolT("svc") {
	//		err := client.Services(api.NamespaceDefault).Delete(name)
	//		if err != nil {
	//			logrus.Fatalf("Unable to delete service %s: %s\n", name, err)
	//		}
	//	} else if c.BoolT("rc") {
	//		err := client.ReplicationControllers(api.NamespaceDefault).Delete(name)
	//		if err != nil {
	//			logrus.Fatalf("Unable to delete replication controller %s: %s\n", name, err)
	//		}
	//	}
	//}
}

// Scale scales rc.
func Scale(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//if c.Int("scale") <= 0 {
	//	logrus.Fatalf("Scale must be defined and a positive number")
	//}
	//
	//for name := range p.Configs {
	//	if len(c.String("rc")) == 0 || c.String("rc") == name {
	//		s, err := client.ExtensionsClient.Scales(api.NamespaceDefault).Get("ReplicationController", name)
	//		if err != nil {
	//			logrus.Fatalf("Error retrieving scaling data: %s\n", err)
	//		}
	//
	//		s.Spec.Replicas = int32(c.Int("scale"))
	//
	//		s, err = client.ExtensionsClient.Scales(api.NamespaceDefault).Update("ReplicationController", s)
	//		if err != nil {
	//			logrus.Fatalf("Error updating scaling data: %s\n", err)
	//		}
	//
	//		fmt.Printf("Scaling %s to: %d\n", name, s.Spec.Replicas)
	//	}
	//}
}

// Convert tranforms docker compose or dab file to k8s objects
func Convert(c *cli.Context) {
	inputFile := c.String("file")
	dabFile := c.String("bundle")
	outFile := c.String("out")
	generateYaml := c.BoolT("yaml")
	toStdout := c.BoolT("stdout")
	createD := c.BoolT("deployment")
	createDS := c.BoolT("daemonset")
	createRS := c.BoolT("replicaset")
	createRC := c.BoolT("replicationcontroller")
	createChart := c.BoolT("chart")
	replicas := c.Int("replicas")
	singleOutput := len(outFile) != 0 || toStdout
	createDeploymentConfig := c.BoolT("deploymentconfig")

	// Create Deployment by default if no controller has be set
	if !createD && !createDS && !createRS && !createRC && !createDeploymentConfig {
		createD = true
	}

	// Validate the flags
	if len(outFile) != 0 && toStdout {
		logrus.Fatalf("Error: --out and --stdout can't be set at the same time")
	}
	if createChart && toStdout {
		logrus.Fatalf("Error: chart cannot be generated when --stdout is specified")
	}
	if replicas < 0 {
		logrus.Fatalf("Error: --replicas cannot be negative")
	}
	if singleOutput {
		count := 0
		if createD {
			count++
		}
		if createDS {
			count++
		}
		if createRS {
			count++
		}
		if createRC {
			count++
		}
		if createDeploymentConfig {
			count++
		}
		if count > 1 {
			logrus.Fatalf("Error: only one type of Kubernetes controller can be generated when --out or --stdout is specified")
		}
	}
	if len(dabFile) > 0 && len(inputFile) > 0 && inputFile != DefaultComposeFile {
		logrus.Fatalf("Error: compose file and dab file cannot be specified at the same time")
	}

	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	file := inputFile

	if len(dabFile) > 0 {
		inputFormat = "bundle"
		file = dabFile
	}

	// loader parses input from file into komposeObject.
	switch inputFormat {
	case "bundle":
		komposeObject = loader.LoadBundle(file)
	case "compose":
		komposeObject = loader.LoadCompose(file)
	default:
		logrus.Fatalf("Input file format is not supported")

	}

	opt := kobject.ConvertOptions{
		ToStdout:               toStdout,
		CreateD:                createD,
		CreateRS:               createRS,
		CreateRC:               createRC,
		CreateDS:               createDS,
		CreateDeploymentConfig: createDeploymentConfig,
		CreateChart:            createChart,
		GenerateYaml:           generateYaml,
		Replicas:               replicas,
		InputFile:              file,
		OutFile:                outFile,
	}

	// transformer maps komposeObject to provider(K8S, OpenShift) primitives
	transformer.Transform(komposeObject, opt)
}

// Up brings up rc, svc.
func Up(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//files, err := ioutil.ReadDir(".")
	//if err != nil {
	//	logrus.Fatalf("Failed to load rc, svc manifest files: %s\n", err)
	//}
	//
	//// submit svc first
	//sc := &api.Service{}
	//for _, file := range files {
	//	if strings.Contains(file.Name(), "svc") {
	//		datasvc, err := ioutil.ReadFile(file.Name())
	//
	//		if err != nil {
	//			logrus.Fatalf("Failed to load %s: %s\n", file.Name(), err)
	//		}
	//
	//		if strings.Contains(file.Name(), "json") {
	//			err := json.Unmarshal(datasvc, &sc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to svc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		if strings.Contains(file.Name(), "yaml") {
	//			err := yaml.Unmarshal(datasvc, &sc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to svc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		// submit sc to k8s
	//		scCreated, err := client.Services(api.NamespaceDefault).Create(sc)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		logrus.Debugf("%s\n", scCreated)
	//	}
	//}
	//
	//// then submit rc
	//rc := &api.ReplicationController{}
	//for _, file := range files {
	//	if strings.Contains(file.Name(), "rc") {
	//		datarc, err := ioutil.ReadFile(file.Name())
	//
	//		if err != nil {
	//			logrus.Fatalf("Failed to load %s: %s\n", file.Name(), err)
	//		}
	//
	//		if strings.Contains(file.Name(), "json") {
	//			err := json.Unmarshal(datarc, &rc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to rc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		if strings.Contains(file.Name(), "yaml") {
	//			err := yaml.Unmarshal(datarc, &rc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to rc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		// submit rc to k8s
	//		rcCreated, err := client.ReplicationControllers(api.NamespaceDefault).Create(rc)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		logrus.Debugf("%s\n", rcCreated)
	//	}
	//}

}
