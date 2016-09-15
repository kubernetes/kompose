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

package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

/**
 * Generate Helm Chart configuration
 */
func generateHelm(filenames []string, outFiles []string) error {
	type ChartDetails struct {
		Name string
	}
	// Let assume all the docker-compose files are in the same directory
	filename := filenames[0]
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

// Check if given path is a directory
func isDir(name string) bool {

	// Open file to get stat later
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	// Get file attributes and information
	fileStat, err := f.Stat()
	if err != nil {
		logrus.Fatalf("error retrieving file information: %v", err)
	}

	// Check if given path is a directory
	if fileStat.IsDir() {
		return true
	}
	return false
}

// PrintList will take the data converted and decide on the commandline attributes given
func PrintList(objects []runtime.Object, opt kobject.ConvertOptions) error {

	var f *os.File
	var dirName string

	// Check if output file is a directory
	if isDir(opt.OutFile) {
		dirName = opt.OutFile
	} else {
		f = transformer.CreateOutFile(opt.OutFile)
		defer f.Close()
	}

	var files []string

	// if asked to print to stdout or to put in single file
	// we will create a list
	if opt.ToStdout || f != nil {
		list := &api.List{}
		// convert objects to versioned and add them to list
		for _, object := range objects {
			versionedObject, err := convertToVersion(object, unversioned.GroupVersion{})
			if err != nil {
				return err
			}

			list.Items = append(list.Items, versionedObject)

		}
		// version list itself
		listVersion := unversioned.GroupVersion{Group: "", Version: "v1"}
		convertedList, err := convertToVersion(list, listVersion)
		if err != nil {
			return err
		}
		data, err := marshal(convertedList, opt.GenerateJSON)
		if err != nil {
			return fmt.Errorf("Error in marshalling the List: %v", err)
		}
		files = append(files, transformer.Print("", dirName, "", data, opt.ToStdout, opt.GenerateJSON, f))
	} else {
		var file string
		// create a separate file for each provider
		for _, v := range objects {
			versionedObject, err := convertToVersion(v, unversioned.GroupVersion{})
			if err != nil {
				return err
			}
			data, err := marshal(versionedObject, opt.GenerateJSON)
			if err != nil {
				return err
			}

			val := reflect.ValueOf(v).Elem()
			// Use reflect to access TypeMeta struct inside runtime.Object.
			// cast it to correct type - unversioned.TypeMeta
			typeMeta := val.FieldByName("TypeMeta").Interface().(unversioned.TypeMeta)

			// Use reflect to access ObjectMeta struct inside runtime.Object.
			// cast it to correct type - api.ObjectMeta
			objectMeta := val.FieldByName("ObjectMeta").Interface().(api.ObjectMeta)

			file = transformer.Print(objectMeta.Name, dirName, strings.ToLower(typeMeta.Kind), data, opt.ToStdout, opt.GenerateJSON, f)

			files = append(files, file)
		}
	}
	if opt.CreateChart {
		generateHelm(opt.InputFiles, files)
	}
	return nil
}

// marshal object runtime.Object and return byte array
func marshal(obj runtime.Object, jsonFormat bool) (data []byte, err error) {
	// convert data to yaml or json
	if jsonFormat {
		data, err = json.MarshalIndent(obj, "", "  ")
	} else {
		data, err = yaml.Marshal(obj)
	}
	if err != nil {
		data = nil
	}
	return
}

// Convert object to versioned object
// if groupVersion is  empty (unversioned.GroupVersion{}), use version from original object (obj)
func convertToVersion(obj runtime.Object, groupVersion unversioned.GroupVersion) (runtime.Object, error) {

	var version unversioned.GroupVersion

	if groupVersion.Empty() {
		objectVersion := obj.GetObjectKind().GroupVersionKind()
		version = unversioned.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
	} else {
		version = groupVersion
	}
	convertedObject, err := api.Scheme.ConvertToVersion(obj, version)
	if err != nil {
		return nil, err
	}
	return convertedObject, nil
}

// PortsExist checks if service has ports defined
func (k *Kubernetes) PortsExist(name string, service kobject.ServiceConfig) bool {
	if len(service.Port) == 0 {
		logrus.Warningf("[%s] No ports defined, we will create a Headless service.", name)
		return false
	}
	return true

}

// CreateService creates a k8s service
func (k *Kubernetes) CreateService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := k.InitSvc(name, service)

	// Configure the service ports.
	servicePorts := k.ConfigServicePorts(name, service)
	svc.Spec.Ports = servicePorts

	svc.Spec.Type = api.ServiceType(service.ServiceType)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// CreateHeadlessService creates a k8s headless service
func (k *Kubernetes) CreateHeadlessService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := k.InitSvc(name, service)

	servicePorts := []api.ServicePort{}
	// Configure a dummy port: https://github.com/kubernetes/kubernetes/issues/32766.
	servicePorts = append(servicePorts, api.ServicePort{
		Name: "headless",
		Port: 55555,
	})

	svc.Spec.Ports = servicePorts
	svc.Spec.ClusterIP = "None"

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// UpdateKubernetesObjects loads configurations to k8s objects
func (k *Kubernetes) UpdateKubernetesObjects(name string, service kobject.ServiceConfig, objects *[]runtime.Object) {
	// Configure the environment variables.
	envs := k.ConfigEnvs(name, service)

	// Configure the container volumes.
	volumesMount, volumes, pvc := k.ConfigVolumes(name, service)
	if pvc != nil {
		// Looping on the slice pvc instead of `*objects = append(*objects, pvc...)`
		// because the type of objects and pvc is different, but when doing append
		// one element at a time it gets converted to runtime.Object for objects slice
		for _, p := range pvc {
			*objects = append(*objects, p)
		}
	}

	// Configure the container ports.
	ports := k.ConfigPorts(name, service)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)

	// fillTemplate fills the pod template with the value calculated from config
	fillTemplate := func(template *api.PodTemplateSpec) {
		if len(service.ContainerName) > 0 {
			template.Spec.Containers[0].Name = service.ContainerName
		}
		template.Spec.Containers[0].Env = envs
		template.Spec.Containers[0].Command = service.Command
		template.Spec.Containers[0].Args = service.Args
		template.Spec.Containers[0].WorkingDir = service.WorkingDir
		template.Spec.Containers[0].VolumeMounts = volumesMount
		template.Spec.Containers[0].Stdin = service.Stdin
		template.Spec.Containers[0].TTY = service.Tty
		template.Spec.Volumes = volumes

		securityContext := &api.SecurityContext{}
		if service.Privileged == true {
			securityContext.Privileged = &service.Privileged
		}
		if service.User != "" {
			uid, err := strconv.ParseInt(service.User, 10, 64)
			if err != nil {
				logrus.Warn("Ignoring user directive. User to be specified as a UID (numeric).")
			} else {
				securityContext.RunAsUser = &uid
			}

		}
		// update template only if securityContext is not empty
		if *securityContext != (api.SecurityContext{}) {
			template.Spec.Containers[0].SecurityContext = securityContext
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
	for _, obj := range *objects {
		k.UpdateController(obj, fillTemplate, fillObjectMeta)
	}
}

// SortServicesFirst - the objects that we get can be in any order this keeps services first
// according to best practice kubernetes services should be created first
// http://kubernetes.io/docs/user-guide/config-best-practices/
func (k *Kubernetes) SortServicesFirst(objs *[]runtime.Object) {
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

func (k *Kubernetes) findDependentVolumes(svcname string, komposeObject kobject.KomposeObject) (volumes []api.Volume, volumeMounts []api.VolumeMount) {
	// Get all the volumes and volumemounts this particular service is dependent on
	for _, dependentSvc := range komposeObject.ServiceConfigs[svcname].VolumesFrom {
		vols, volMounts := k.findDependentVolumes(dependentSvc, komposeObject)
		volumes = append(volumes, vols...)
		volumeMounts = append(volumeMounts, volMounts...)
	}
	// add the volumes info of this service
	volMounts, vols, _ := k.ConfigVolumes(svcname, komposeObject.ServiceConfigs[svcname])
	volumes = append(volumes, vols...)
	volumeMounts = append(volumeMounts, volMounts...)
	return
}

// VolumesFrom creates volums and volumeMounts for volumes_from
func (k *Kubernetes) VolumesFrom(objects *[]runtime.Object, komposeObject kobject.KomposeObject) {
	for _, obj := range *objects {
		switch t := obj.(type) {
		case *api.ReplicationController:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts := k.findDependentVolumes(dependentSvc, komposeObject)
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *extensions.Deployment:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts := k.findDependentVolumes(dependentSvc, komposeObject)
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *extensions.DaemonSet:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts := k.findDependentVolumes(dependentSvc, komposeObject)
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *deployapi.DeploymentConfig:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts := k.findDependentVolumes(dependentSvc, komposeObject)
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		}
	}
}
