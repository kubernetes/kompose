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

package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/transformer"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

/**
 * Generate Helm Chart configuration
 */
func generateHelm(filename string, outFiles []string) error {
	type ChartDetails struct {
		Name string
	}

	extension := filepath.Ext(filename)
	dirName := filename[0 : len(filename)-len(extension)]
	details := ChartDetails{dirName}
	manifestDir := dirName + string(os.PathSeparator) + "templates"
	dir, err := os.Open(dirName)

	/* Setup the initial directories/files */
	if err == nil {
		_ = dir.Close()
	}

	if err != nil {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			return err
		}

		err = os.Mkdir(manifestDir, 0755)
		if err != nil {
			return err
		}

		/* Create the readme file */
		readme := "This chart was created by Kompose\n"
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"README.md", []byte(readme), 0644)
		if err != nil {
			return err
		}

		/* Create the Chart.yaml file */
		chart := `name: {{.Name}}
description: A generated Helm Chart for {{.Name}} from Skippbox Kompose
version: 0.0.1
keywords:
  - {{.Name}}
sources:
home:
`

		t, err := template.New("ChartTmpl").Parse(chart)
		if err != nil {
			logrus.Fatalf("Failed to generate Chart.yaml template: %s\n", err)
		}
		var chartData bytes.Buffer
		_ = t.Execute(&chartData, details)

		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"Chart.yaml", chartData.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	/* Copy all related json/yaml files into the newly created manifests directory */
	for _, filename := range outFiles {
		if err = cpFileToChart(manifestDir, filename); err != nil {
			logrus.Warningln(err)
		}
		if err = os.Remove(filename); err != nil {
			logrus.Warningln(err)
		}
	}
	logrus.Infof("chart created in %q\n", "."+string(os.PathSeparator)+dirName+string(os.PathSeparator))
	return nil
}

func cpFileToChart(manifestDir, filename string) error {
	infile, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Warningf("Error reading %s: %s\n", filename, err)
		return err
	}

	return ioutil.WriteFile(manifestDir+string(os.PathSeparator)+filename, infile, 0644)
}

// PrintList will take the data converted and decide on the commandline attributes given
func PrintList(objects []runtime.Object, opt kobject.ConvertOptions) error {
	f := transformer.CreateOutFile(opt.OutFile)
	defer f.Close()

	var err error
	var files []string

	// if asked to print to stdout or to put in single file
	// we will create a list
	if opt.ToStdout || f != nil {
		list := &api.List{}
		list.Items = objects

		// version each object in the list
		list.Items, err = convertToVersion(list.Items)
		if err != nil {
			return err
		}

		// version list itself
		listVersion := unversioned.GroupVersion{Group: "", Version: "v1"}
		convertedList, err := api.Scheme.ConvertToVersion(list, listVersion)
		if err != nil {
			return err
		}
		data, err := marshal(convertedList, opt.GenerateYaml)
		if err != nil {
			return fmt.Errorf("Error in marshalling the List: %v", err)
		}
		files = append(files, transformer.Print("", "", data, opt.ToStdout, opt.GenerateYaml, f))
	} else {
		var file string
		// create a separate file for each provider
		for _, v := range objects {
			data, err := marshal(v, opt.GenerateYaml)
			if err != nil {
				return err
			}
			switch t := v.(type) {
			case *api.ReplicationController:
				file = transformer.Print(t.Name, strings.ToLower(t.Kind), data, opt.ToStdout, opt.GenerateYaml, f)
			case *extensions.Deployment:
				file = transformer.Print(t.Name, strings.ToLower(t.Kind), data, opt.ToStdout, opt.GenerateYaml, f)
			case *extensions.DaemonSet:
				file = transformer.Print(t.Name, strings.ToLower(t.Kind), data, opt.ToStdout, opt.GenerateYaml, f)
			case *deployapi.DeploymentConfig:
				file = transformer.Print(t.Name, strings.ToLower(t.Kind), data, opt.ToStdout, opt.GenerateYaml, f)
			case *api.Service:
				file = transformer.Print(t.Name, strings.ToLower(t.Kind), data, opt.ToStdout, opt.GenerateYaml, f)
			}
			files = append(files, file)
		}
	}
	if opt.CreateChart {
		generateHelm(opt.InputFile, files)
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
func convertToVersion(objs []runtime.Object) ([]runtime.Object, error) {
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

func PortsExist(name string, service kobject.ServiceConfig) bool {
	if len(service.Port) == 0 {
		logrus.Warningf("[%s] Service cannot be created because of missing port.", name)
		return false
	} else {
		return true
	}
}

// create a k8s service
func CreateService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := InitSvc(name, service)

	// Configure the service ports.
	servicePorts := ConfigServicePorts(name, service)
	svc.Spec.Ports = servicePorts

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// load configurations to k8s objects
func UpdateKubernetesObjects(name string, service kobject.ServiceConfig, objects []runtime.Object) {
	// Configure the environment variables.
	envs := ConfigEnvs(name, service)

	// Configure the container volumes.
	volumesMount, volumes := ConfigVolumes(service)

	// Configure the container ports.
	ports := ConfigPorts(name, service)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)

	// fillTemplate fills the pod template with the value calculated from config
	fillTemplate := func(template *api.PodTemplateSpec) {
		if len(service.ContainerName) > 0 {
			template.Spec.Containers[0].Name = service.ContainerName
		}
		template.Spec.Containers[0].Env = envs
		template.Spec.Containers[0].Command = service.Entrypoint
		template.Spec.Containers[0].Args = service.Command
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
		template.ObjectMeta.Labels = transformer.ConfigLabels(name)
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
		meta.Annotations = annotations
	}

	// update supported controller
	for _, obj := range objects {
		UpdateController(obj, fillTemplate, fillObjectMeta)
	}
}
