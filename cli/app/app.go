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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	// install kubernetes api
	_ "github.com/openshift/origin/pkg/deploy/api/install"

	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/loader"
	"github.com/skippbox/kompose/pkg/transformer"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"fmt"
	"strings"
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

// Convert komposeObject to K8S controllers
func komposeConvert(komposeObject KomposeObject, opt convertOptions) []runtime.Object {
	var svcnames []string

	// this will hold all the converted data
	var allobjects []runtime.Object
	for name, service := range komposeObject.ServiceConfigs {
		var objects []runtime.Object
		svcnames = append(svcnames, name)
		sc := initSC(name, service)

		if opt.createD {
			objects = append(objects, initDC(name, service, opt.replicas))
		}
		if opt.createDS {
			objects = append(objects, initDS(name, service))
		}
		if opt.createRC {
			objects = append(objects, initRC(name, service, opt.replicas))
		}
		if opt.createDeploymentConfig {
			objects = append(objects, initDeploymentConfig(name, service, opt.replicas)) // OpenShift DeploymentConfigs
		}

		// Configure the environment variables.
		envs := configEnvs(name, service)

		// Configure the container command.
		var cmds []string
		for _, cmd := range service.Command {
			cmds = append(cmds, cmd)
		}
		// Configure the container volumes.
		volumesMount, volumes := configVolumes(service)

		// Configure the container ports.
		ports := configPorts(name, service)

		// Configure the service ports.
		servicePorts := configServicePorts(name, service)
		sc.Spec.Ports = servicePorts

		// Configure labels
		labels := map[string]string{"service": name}
		sc.ObjectMeta.Labels = labels
		// Configure annotations
		annotations := map[string]string{}
		for key, value := range service.Annotations {
			annotations[key] = value
		}
		sc.ObjectMeta.Annotations = annotations

		// fillTemplate fills the pod template with the value calculated from config
		fillTemplate := func(template *api.PodTemplateSpec) {
			template.Spec.Containers[0].Env = envs
			template.Spec.Containers[0].Command = cmds
			template.Spec.Containers[0].WorkingDir = service.WorkingDir
			template.Spec.Containers[0].VolumeMounts = volumesMount
			template.Spec.Volumes = volumes
			// Configure the container privileged mode
			if service.Privileged == true {
				template.Spec.Containers[0].SecurityContext = &api.SecurityContext{
					Privileged: &service.Privileged,
				}
			}
			template.Spec.Containers[0].Ports = ports
			template.ObjectMeta.Labels = labels
			// Configure the container restart policy.
			switch service.Restart {
			case "", "always":
				template.Spec.RestartPolicy = api.RestartPolicyAlways
			case "no":
				template.Spec.RestartPolicy = api.RestartPolicyNever
			case "on-failure":
				template.Spec.RestartPolicy = api.RestartPolicyOnFailure
			default:
				logrus.Fatalf("Unknown restart policy %s for service %s", service.Restart, name)
			}
		}

		// fillObjectMeta fills the metadata with the value calculated from config
		fillObjectMeta := func(meta *api.ObjectMeta) {
			meta.Labels = labels
			meta.Annotations = annotations
		}

		// update supported controller
		for _, obj := range objects {
			updateController(obj, fillTemplate, fillObjectMeta)
		}

		// If ports not provided in configuration we will not make service
		if len(ports) == 0 {
			logrus.Warningf("[%s] Service cannot be created because of missing port.", name)
		} else {
			objects = append(objects, sc)
		}
		allobjects = append(allobjects, objects...)
	}
	return allobjects
}

// PrintList will take the data converted and decide on the commandline attributes given
func PrintList(objects []runtime.Object, opt convertOptions) error {
	f := createOutFile(opt.outFile)
	defer f.Close()

	var err error
	var files []string

	// if asked to print to stdout or to put in single file
	// we will create a list
	if opt.toStdout || f != nil {
		list := &api.List{}
		list.Items = objects

		// version each object in the list
		list.Items, err = ConvertToVersion(list.Items)
		if err != nil {
			return err
		}

		// version list itself
		listVersion := unversioned.GroupVersion{Group: "", Version: "v1"}
		convertedList, err := api.Scheme.ConvertToVersion(list, listVersion)
		if err != nil {
			return err
		}
		data, err := marshal(convertedList, opt.generateYaml)
		if err != nil {
			return fmt.Errorf("Error in marshalling the List: %v", err)
		}
		files = append(files, print("", "", data, opt.toStdout, opt.generateYaml, f))
	} else {
		var file string
		// create a separate file for each provider
		for _, v := range objects {
			data, err := marshal(v, opt.generateYaml)
			if err != nil {
				return err
			}
			switch t := v.(type) {
			case *api.ReplicationController:
				file = print(t.Name, strings.ToLower(t.Kind), data, opt.toStdout, opt.generateYaml, f)
			case *extensions.Deployment:
				file = print(t.Name, strings.ToLower(t.Kind), data, opt.toStdout, opt.generateYaml, f)
			case *extensions.DaemonSet:
				file = print(t.Name, strings.ToLower(t.Kind), data, opt.toStdout, opt.generateYaml, f)
			case *deployapi.DeploymentConfig:
				file = print(t.Name, strings.ToLower(t.Kind), data, opt.toStdout, opt.generateYaml, f)
			case *api.Service:
				file = print(t.Name, strings.ToLower(t.Kind), data, opt.toStdout, opt.generateYaml, f)
			}
			files = append(files, file)

		}
	}
	if opt.createChart {
		generateHelm(opt.inputFile, files)
	}
	return nil
}

// marshal object runtime.Object and return byte array
func marshal(obj runtime.Object, yamlFormat bool) (data []byte, err error) {
	// convert data to yaml or json
	if yamlFormat {
		data, err = yaml.Marshal(obj)
	} else {
		data, err = json.MarshalIndent(obj, "", "  ")
	}
	if err != nil {
		data = nil
	}
	return
}

// Convert all objects in objs to versioned objects
func ConvertToVersion(objs []runtime.Object) ([]runtime.Object, error) {
	ret := []runtime.Object{}

	for _, obj := range objs {

		objectVersion := obj.GetObjectKind().GroupVersionKind()
		version := unversioned.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
		convertedObject, err := api.Scheme.ConvertToVersion(obj, version)
		if err != nil {
			return nil, err
		}
		ret = append(ret, convertedObject)
	}

	return ret, nil
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
			logrus.Fatalf("Error: only one type of Kubernetes controller can be generated when --out or --stdout is specified")
		}
	}
	if len(dabFile) > 0 && len(inputFile) > 0 && inputFile != DefaultComposeFile {
		logrus.Fatalf("Error: compose file and dab file cannot be specified at the same time")
	}
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

	//komposeObject.Loader(file, inputFormat)
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

	// Convert komposeObject to K8S controllers
	mServices, mDeployments, mDaemonSets, mReplicationControllers, mDeploymentConfigs, svcnames := transformer.Transform(komposeObject, opt)

	// Print output
	transformer.PrintControllers(mServices, mDeployments, mDaemonSets, mReplicationControllers, mDeploymentConfigs, svcnames, opt, f)
}

// Up brings up deployment, svc.
func Up(c *cli.Context) {
	fmt.Println("We are going to create Kubernetes deployment and service for your dockerized application. \n" +
		"If you need more kind of controllers, use 'kompose convert' and 'kubectl create -f' instead. \n")

	factory := cmdutil.NewFactory(nil)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		logrus.Fatalf("Failed to access the Kubernetes cluster. Make sure you have a Kubernetes running: %v", err)
	}
	client := client.NewOrDie(clientConfig)

	inputFile := c.String("file")
	dabFile := c.String("bundle")

	komposeObject := KomposeObject{}
	opt := convertOptions{
		replicas: 1,
		createD:  true,
	}

	validateFlags(opt, false, dabFile, inputFile)

	if len(dabFile) > 0 {
		komposeObject = loadBundlesFile(dabFile)
	} else {
		komposeObject = loadComposeFile(inputFile)
	}

	//Convert komposeObject to K8S controllers
	objects := komposeConvert(komposeObject, opt)
	objects = sortServicesFirst(objects)

	for _, v := range objects {
		switch t := v.(type) {
		case *extensions.Deployment:
			_, err := client.Deployments(api.NamespaceDefault).Create(t)
			if err != nil {
				logrus.Fatalf("Error: '%v' while creating deployment: %s", err, t.Name)
			}
			logrus.Infof("Successfully created deployment: %s", t.Name)
		case *api.Service:
			_, err := client.Services(api.NamespaceDefault).Create(t)
			if err != nil {
				logrus.Fatalf("Error: '%v' while creating service: %s", err, t.Name)
			}
			logrus.Infof("Successfully created service: %s", t.Name)
		}
	}
	fmt.Println("\nApplication has been deployed to Kubernetes. You can run 'kubectl get deployment,svc' for details.")
}

// the objects that we get can be in any order this keeps services first
// according to best practice kubernetes services should be created first
// http://kubernetes.io/docs/user-guide/config-best-practices/
func sortServicesFirst(objs []runtime.Object) []runtime.Object {
	var svc []runtime.Object
	var others []runtime.Object
	var ret []runtime.Object

	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Service" {
			svc = append(svc, obj)
		} else {
			others = append(others, obj)
		}
	}
	ret = append(ret, svc...)
	ret = append(ret, others...)
	return ret
}
